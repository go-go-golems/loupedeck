#!/bin/bash
# 04-verify-word-wrapping.sh
# Verify that the word wrapping implementation compiles and passes tests.
# Run from the loupedeck repo root.

set -euo pipefail

echo "=== Building ==="
go build ./...

echo "=== Running gfx text wrapping tests ==="
go test ./runtime/gfx/ -count=1 -v -run "TestWrapText|TestSurfaceTextWrapWidth|TestSurfaceTextWrapWith|TestExpandTextLines"

echo "=== Running render tests ==="
go test ./runtime/render/ -count=1 -v

echo "=== Running ui tests ==="
go test ./runtime/ui/ -count=1 -v

echo "=== Running full suite ==="
go test ./... -count=1 -timeout 60s

echo "=== All checks passed ==="
