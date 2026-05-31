#!/bin/bash
# 05-verify-per-tile-ergonomics.sh
# Verify per-tile invalidation ergonomics (tile.draw, tile.invalidate).
# Run from the loupedeck repo root.

set -euo pipefail

echo "=== Building ==="
go build ./...

echo "=== Running ui tests (includes Draw, Invalidate, Wrap) ==="
go test ./runtime/ui/ -count=1 -v

echo "=== Running full suite ==="
go test ./... -count=1 -timeout 60s

echo "=== All checks passed ==="
