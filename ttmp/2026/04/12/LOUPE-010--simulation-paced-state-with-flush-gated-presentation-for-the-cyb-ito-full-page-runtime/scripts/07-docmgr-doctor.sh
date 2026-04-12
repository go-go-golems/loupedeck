#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

docmgr doctor --ticket LOUPE-010 --stale-after 30
