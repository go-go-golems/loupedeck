#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration 90s \
  --send-interval 0ms \
  --exit-on-circle=false
