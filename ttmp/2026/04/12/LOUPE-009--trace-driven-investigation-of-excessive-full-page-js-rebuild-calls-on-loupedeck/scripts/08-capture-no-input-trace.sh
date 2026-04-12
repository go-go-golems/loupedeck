#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

DURATION=${1:-12s}
TRACE_LIMIT=${2:-500}
LOG="/tmp/loupe-cyb-ito-full10-trace-$(date +%s).log"
echo "$LOG"
(timeout 30s go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration "$DURATION" \
  --log-render-stats \
  --log-writer-stats \
  --log-js-stats \
  --log-js-trace \
  --log-go-trace \
  --stats-interval 1s \
  --trace-limit "$TRACE_LIMIT") 2>&1 | tee "$LOG"
