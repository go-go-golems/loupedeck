#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

docmgr doctor --ticket LOUPE-012 --stale-after 30
