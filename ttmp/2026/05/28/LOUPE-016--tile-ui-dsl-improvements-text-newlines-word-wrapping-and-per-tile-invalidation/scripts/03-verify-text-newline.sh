#!/bin/bash
# 03-verify-text-newline.sh
# Verify that the text newline implementation compiles and passes tests.
# Run from the loupedeck repo root.

set -euo pipefail

echo "=== Building ==="
go build ./...

echo "=== Running gfx text tests ==="
go test ./runtime/gfx/ -count=1 -v -run "TestSurfaceText|TestSplitLines"

echo "=== Running full suite ==="
go test ./... -count=1 -timeout 60s

echo "=== All checks passed ==="
