#!/usr/bin/env bash
# 03-event-poll-demo.sh — Demonstrate event polling workflow
#
# Prerequisites: Server running (use 02-start-server.sh)
#
# This script demonstrates the polling workflow:
# 1. Create a page with tiles
# 2. Poll for events (initially empty)
# 3. Show the cursor-based polling pattern
# 4. Acknowledge events

set -euo pipefail
BASE_URL="${1:-http://localhost:9876}"

echo "=== Event Polling Demo ==="
echo ""

# Create a dashboard page
echo "--- Creating dashboard page ---"
curl -s -X POST "$BASE_URL/api/v1/pages" \
  -H 'Content-Type: application/json' \
  -d '{"name":"dashboard","tiles":[
    {"col":0,"row":0,"text":"STATUS","icon":"check"},
    {"col":1,"row":0,"text":"EVENTS","icon":"bell"},
    {"col":2,"row":0,"text":"0","icon":"hash"}
  ]}' | python3 -m json.tool 2>/dev/null || echo "(page may already exist)"

echo ""
echo "--- Showing dashboard page ---"
curl -s -X POST "$BASE_URL/api/v1/pages/show" \
  -H 'Content-Type: application/json' \
  -d '{"name":"dashboard"}' | python3 -m json.tool 2>/dev/null

echo ""
echo "--- Initial event poll ---"
CURSOR=0
RESP=$(curl -s "$BASE_URL/api/v1/events?since=$CURSOR&limit=10")
echo "$RESP" | python3 -m json.tool 2>/dev/null || echo "$RESP"

# Extract cursor
NEW_CURSOR=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('cursor',0))" 2>/dev/null || echo "0")
echo ""
echo "Cursor: $CURSOR → $NEW_CURSOR"

echo ""
echo "=== Polling Pattern ==="
echo ""
echo "To poll for events, a client repeats:"
echo "  1. GET /api/v1/events?since=\$CURSOR"
echo "  2. Process received events"
echo "  3. Update CURSOR to the response's cursor value"
echo "  4. DELETE /api/v1/events with {until: CURSOR} to acknowledge"
echo "  5. Wait, then go to step 1"
echo ""
echo "Example polling loop (bash):"
echo '  cursor=0'
echo '  while true; do'
echo '    resp=$(curl -s "$BASE/api/v1/events?since=$cursor")'
echo '    cursor=$(echo "$resp" | jq .cursor)'
echo '    events=$(echo "$resp" | jq .events)'
echo '    if [ "$events" != "null" ] && [ "$events" != "[]" ]; then'
echo '      echo "New events: $events"'
echo '      curl -s -X DELETE "$BASE/api/v1/events" -d "{\"until\":$cursor}"'
echo '    fi'
echo '    sleep 1'
echo '  done'
