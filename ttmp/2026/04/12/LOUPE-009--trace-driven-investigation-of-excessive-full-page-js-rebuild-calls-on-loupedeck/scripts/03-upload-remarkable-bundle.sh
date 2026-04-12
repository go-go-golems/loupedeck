#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../../../../../.." && pwd)
cd "$REPO_ROOT"

remarquee upload bundle \
  ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/index.md \
  ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/design/01-textbook-trace-driven-investigation-of-excessive-full-page-rebuild-calls.md \
  ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/playbooks/01-trace-capture-runbook.md \
  ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/reference/01-implementation-diary.md \
  --name "LOUPE-009 Trace-driven investigation of excessive full-page rebuild calls" \
  --remote-dir "/ai/2026/04/12/LOUPE-009" \
  --toc-depth 2

remarquee cloud ls "/ai/2026/04/12/LOUPE-009" --long --non-interactive
