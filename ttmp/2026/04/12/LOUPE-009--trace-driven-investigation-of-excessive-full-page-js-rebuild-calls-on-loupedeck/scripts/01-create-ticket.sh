#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

docmgr ticket create-ticket \
  --ticket LOUPE-009 \
  --title "Trace-driven investigation of excessive full-page JS rebuild calls on Loupedeck" \
  --topics loupedeck,goja,javascript,animation,rendering,benchmarking,performance
