#!/usr/bin/env bash
# 04-stop-server.sh — Stop the loupedeck-server tmux session

set -euo pipefail
SESSION="loupedeck-api"

if tmux has-session -t "$SESSION" 2>/dev/null; then
  tmux kill-session -t "$SESSION"
  echo "Stopped loupedeck-server (tmux session: $SESSION)"
else
  echo "No tmux session '$SESSION' found"
fi
