#!/usr/bin/env bash
# 01-smoke-test.sh — Basic API smoke test for loupedeck-server
#
# Usage: ./01-smoke-test.sh [BASE_URL]
# Default BASE_URL: http://localhost:9876

set -euo pipefail

BASE_URL="${1:-http://localhost:9876}"
PASS=0
FAIL=0

ok() { PASS=$((PASS+1)); echo "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $1"; }

check() {
  local desc="$1" method="$2" path="$3" expect_status="${4:-200}" body="${5:-}"
  local cmd="curl -s -o /tmp/smoke-res -w '%{http_code}' -X $method"
  if [ -n "$body" ]; then
    cmd="$cmd -H 'Content-Type: application/json' -d '$body'"
  fi
  cmd="$cmd $BASE_URL$path"
  status=$(eval "$cmd" 2>/dev/null)
  if [ "$status" = "$expect_status" ]; then
    ok "$desc (HTTP $status)"
  else
    fail "$desc (expected $expect_status, got $status)"
  fi
}

check_json_field() {
  local desc="$1" path="$2" field="$3" expected="$4"
  local val
  val=$(curl -s "$BASE_URL$path" | python3 -c "import sys,json; print(json.load(sys.stdin).get('$field','MISSING'))" 2>/dev/null || echo "PARSE_ERROR")
  if [ "$val" = "$expected" ]; then
    ok "$desc ($field=$expected)"
  else
    fail "$desc ($field: expected '$expected', got '$val')"
  fi
}

echo "=== Loupedeck Server Smoke Test ==="
echo "Base URL: $BASE_URL"
echo ""

# ── Device Info ──
echo "--- Device Info ---"
check "GET /api/v1/info" GET /api/v1/info
check_json_field "info has model" /api/v1/info model "Loupedeck Live"
check_json_field "info has connected" /api/v1/info connected True

# ── Brightness ──
echo "--- Brightness ---"
check "GET /api/v1/brightness" GET /api/v1/brightness
check_json_field "default brightness" /api/v1/brightness value "9"
check "PUT /api/v1/brightness" PUT /api/v1/brightness 200 '{"value":5}'
check_json_field "updated brightness" /api/v1/brightness value "5"
# Restore
check "PUT /api/v1/brightness restore" PUT /api/v1/brightness 200 '{"value":9}'

# ── Buttons ──
echo "--- Buttons ---"
check "GET /api/v1/buttons" GET /api/v1/buttons
check "GET /api/v1/buttons/Circle" GET /api/v1/buttons/Circle
check "PUT /api/v1/buttons/Circle/color" PUT /api/v1/buttons/Circle/color 200 '{"r":255,"g":0,"b":0}'
check "GET unknown button" GET /api/v1/buttons/UNKNOWN 404

# ── Displays ──
echo "--- Displays ---"
check "GET /api/v1/displays" GET /api/v1/displays
check "POST /api/v1/displays/main/draw (grid)" POST /api/v1/displays/main/draw 200 '{"layout":"grid","tiles":[{"col":0,"row":0,"text":"SMOKE"}]}'

# ── Pages ──
echo "--- Pages ---"
check "POST /api/v1/pages" POST /api/v1/pages 200 '{"name":"smoke-test","tiles":[{"col":0,"row":0,"text":"OK"}]}'
check "GET /api/v1/pages" GET /api/v1/pages
check "GET /api/v1/pages/smoke-test" GET /api/v1/pages/smoke-test
check "POST /api/v1/pages/show" POST /api/v1/pages/show 200 '{"name":"smoke-test"}'
check "DELETE /api/v1/pages/smoke-test" DELETE /api/v1/pages/smoke-test
check "GET deleted page 404" GET /api/v1/pages/smoke-test 404

# ── Events ──
echo "--- Events ---"
check "GET /api/v1/events" GET /api/v1/events
check "GET /api/v1/events?since=0&limit=10" GET "/api/v1/events?since=0&limit=10"
check "GET /api/v1/events?type=button" GET "/api/v1/events?type=button"
check "DELETE /api/v1/events" DELETE /api/v1/events 200 '{"until":999999}'

# ── Error cases ──
echo "--- Error Cases ---"
check "PUT brightness without body" PUT /api/v1/brightness 400
check "POST page without name" POST /api/v1/pages 400 '{}'
check "POST pages/show missing page" POST /api/v1/pages/show 404 '{"name":"nonexistent"}'
check "PUT button color missing rgb" PUT /api/v1/buttons/Circle/color 400

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
