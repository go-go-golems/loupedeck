#!/usr/bin/env bash
# 05-hardware-operator-test.sh — Exercise real Loupedeck hardware endpoints.
#
# Usage:
#   ./05-hardware-operator-test.sh [BASE_URL] [--http-only]
#
# Run after:
#   ./02-start-server.sh --with-hardware
#
# The default mode performs HTTP checks and asks the operator to confirm visible
# brightness, button LED, display, and event behavior. Use --http-only for CI-ish
# validation that the hardware-backed API calls return success without prompts.

set -euo pipefail

BASE_URL="http://localhost:9876"
HTTP_ONLY=0

for arg in "$@"; do
  case "$arg" in
    --http-only) HTTP_ONLY=1 ;;
    http://*|https://*) BASE_URL="$arg" ;;
    *) echo "Unknown argument: $arg" >&2; exit 2 ;;
  esac
done

PASS=0
FAIL=0
TMP_BODY="${TMPDIR:-/tmp}/loupedeck-hw-test-body.$$"
trap 'rm -f "$TMP_BODY"' EXIT

ok() { PASS=$((PASS+1)); echo "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $1"; }

request() {
  local desc="$1" method="$2" path="$3" expect_status="$4" body="${5:-}"
  local status
  if [ -n "$body" ]; then
    status=$(curl -sS -o "$TMP_BODY" -w '%{http_code}' \
      -H 'Content-Type: application/json' \
      -X "$method" "$BASE_URL$path" -d "$body")
  else
    status=$(curl -sS -o "$TMP_BODY" -w '%{http_code}' \
      -X "$method" "$BASE_URL$path")
  fi
  if [ "$status" = "$expect_status" ]; then
    ok "$desc (HTTP $status)"
  else
    fail "$desc (expected HTTP $expect_status, got $status; body=$(cat "$TMP_BODY"))"
  fi
}

json_field() {
  local field="$1"
  python3 -c "import json,sys; print(json.load(sys.stdin).get('$field', 'MISSING'))" < "$TMP_BODY"
}

confirm() {
  local prompt="$1"
  if [ "$HTTP_ONLY" -eq 1 ]; then
    echo "  • skipped visual confirmation: $prompt"
    return 0
  fi
  local answer
  read -r -p "Confirm: $prompt [y/N] " answer
  case "$answer" in
    y|Y|yes|YES) ok "$prompt" ;;
    *) fail "$prompt" ;;
  esac
}

count_events() {
  python3 - "$TMP_BODY" <<'PY'
import json, sys
with open(sys.argv[1], "r", encoding="utf-8") as f:
    payload = json.load(f)
items = payload.get("events", payload if isinstance(payload, list) else [])
print(len(items) if isinstance(items, list) else 0)
PY
}

echo "=== Loupedeck Hardware Operator Test ==="
echo "Base URL: $BASE_URL"
echo "Mode: $([ "$HTTP_ONLY" -eq 1 ] && echo http-only || echo operator-confirmed)"
echo ""

echo "--- Connection ---"
request "GET /api/v1/info" GET /api/v1/info 200
connected=$(json_field connected || echo MISSING)
if [ "$connected" = "True" ] || [ "$connected" = "true" ]; then
  ok "info.connected is true"
else
  fail "info.connected expected true, got $connected"
fi

echo "--- Brightness ---"
request "PUT brightness to dim value 2" PUT /api/v1/brightness 200 '{"value":2}'
sleep 0.5
request "PUT brightness back to value 9" PUT /api/v1/brightness 200 '{"value":9}'
confirm "the Loupedeck display dimmed, then returned to normal brightness"

echo "--- Button LED ---"
request "Circle LED red" PUT /api/v1/buttons/Circle/color 200 '{"r":255,"g":0,"b":0}'
sleep 0.5
request "Circle LED green" PUT /api/v1/buttons/Circle/color 200 '{"r":0,"g":255,"b":0}'
sleep 0.5
request "Circle LED blue" PUT /api/v1/buttons/Circle/color 200 '{"r":0,"g":0,"b":255}'
sleep 0.5
request "Circle LED off" PUT /api/v1/buttons/Circle/color 200 '{"r":0,"g":0,"b":0}'
confirm "the Circle LED changed red, green, blue, then off"

echo "--- Display ---"
request "Draw HW OK grid on main display" POST /api/v1/displays/main/draw 200 '{"layout":"grid","tiles":[{"col":0,"row":0,"text":"HW OK"},{"col":1,"row":0,"text":"LIVE"},{"col":0,"row":1,"text":"API"},{"col":1,"row":1,"text":"TEST"}]}'
confirm "the main display showed HW OK / LIVE / API / TEST tiles"

echo "--- Events ---"
request "Clear events" DELETE /api/v1/events 200 '{"until":999999999999}'
if [ "$HTTP_ONLY" -eq 0 ]; then
  echo "Please press the Circle button, rotate one knob, and/or touch the screen now."
  read -r -p "Press Enter after interacting with the hardware... " _
fi
request "Poll events" GET '/api/v1/events?since=0&limit=20' 200
count=$(count_events || echo 0)
if [ "$HTTP_ONLY" -eq 1 ]; then
  echo "  • observed $count event(s); not failing in http-only mode"
elif [ "$count" -gt 0 ]; then
  ok "hardware interactions appeared in event polling ($count event(s))"
else
  fail "expected at least one hardware event after operator interaction"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
