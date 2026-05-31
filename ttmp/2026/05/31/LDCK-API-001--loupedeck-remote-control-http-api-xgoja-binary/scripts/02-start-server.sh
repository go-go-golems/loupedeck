#!/usr/bin/env bash
# 02-start-server.sh — Start loupedeck-server in a tmux session
#
# Usage: ./02-start-server.sh [--with-hardware]
#
# Starts the server in a tmux session named "loupedeck-api".
# Use --with-hardware to connect to the real Loupedeck device.

set -euo pipefail

SESSION="loupedeck-api"
BINARY="/home/manuel/workspaces/2026-05-27/better-loupedeck-tiles/dist/loupedeck-server"
SCRIPT="/home/manuel/workspaces/2026-05-27/better-loupedeck-tiles/loupedeck-server/server.js"
DB_DIR="/home/manuel/workspaces/2026-05-27/better-loupedeck-tiles/loupedeck-server"

DECK_FLAG="--deck-enabled=false"
if [ "${1:-}" = "--with-hardware" ]; then
  DECK_FLAG="--deck-enabled"
fi

# Kill existing session if any
tmux kill-session -t "$SESSION" 2>/dev/null || true

# Clean old database
rm -f "$DB_DIR/loupedeck_events.db"

# Create new session
tmux new-session -d -s "$SESSION" -x 120 -y 40 \
  "cd $DB_DIR && $BINARY serve $SCRIPT $DECK_FLAG --http-listen :9876"

echo "Server started in tmux session: $SESSION"
echo "  URL: http://localhost:9876"
echo "  Attach: tmux attach -t $SESSION"
echo "  Stop: tmux kill-session -t $SESSION"
echo ""

# Wait for server to be ready
for i in $(seq 1 10); do
  if curl -s http://localhost:9876/api/v1/info > /dev/null 2>&1; then
    echo "Server is ready!"
    exit 0
  fi
  sleep 1
done

echo "WARNING: Server did not respond within 10 seconds"
echo "Check the tmux session: tmux attach -t $SESSION"
exit 1
