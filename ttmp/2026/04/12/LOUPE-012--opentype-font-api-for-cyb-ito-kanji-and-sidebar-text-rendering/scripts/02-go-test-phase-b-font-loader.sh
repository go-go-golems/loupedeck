#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

go test ./runtime/gfx/... ./...
