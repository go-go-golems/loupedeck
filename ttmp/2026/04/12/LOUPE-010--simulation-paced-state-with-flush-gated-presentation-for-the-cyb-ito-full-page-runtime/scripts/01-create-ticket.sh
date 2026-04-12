#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

docmgr ticket create-ticket \
  --ticket LOUPE-010 \
  --title "Simulation-paced state with flush-gated presentation for the cyb-ito full-page runtime" \
  --topics loupedeck,goja,javascript,animation,rendering,performance
