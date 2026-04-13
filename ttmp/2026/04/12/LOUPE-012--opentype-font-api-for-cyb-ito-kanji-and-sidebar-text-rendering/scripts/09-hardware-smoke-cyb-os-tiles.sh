#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

LOG=${1:-/tmp/loupe-cyb-os-tiles-$(date +%s).log}
timeout 30s go run ./cmd/loupe-js-live --script ./examples/js/11-cyb-os-tiles.js --duration 5s --send-interval 0ms >"$LOG" 2>&1
printf '%s\n' "$LOG"
