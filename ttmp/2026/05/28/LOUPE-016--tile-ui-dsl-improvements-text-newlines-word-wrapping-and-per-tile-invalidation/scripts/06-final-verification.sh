#!/bin/bash
# 06-final-verification.sh
# Complete verification of all LOUPE-016 features.
# Run from the loupedeck repo root.

set -euo pipefail

echo "=== Building ==="
go build ./...

echo "=== Running full test suite ==="
go test ./... -count=1 -timeout 60s

echo ""
echo "=== LOUPE-016 Feature Verification ==="
echo ""

echo "--- Feature 1: Text Newlines ---"
echo "  Surface.Text() splits on \\n and renders multiple lines"
echo "  drawCenteredLabel() splits on \\n in retained renderer"
echo "  lineGap property in JS bridge"
go test ./runtime/gfx/ -count=1 -run "TestSurfaceTextNewline|TestSurfaceTextThreeLines|TestSurfaceTextLineGapSpacing|TestSurfaceTextEmptyLinePreserved|TestSurfaceTextTrailingNewline|TestSplitLines" -v

echo ""
echo "--- Feature 2: Word Wrapping ---"
echo "  wrapText() greedy word-wrapping algorithm"
echo "  WrapWidth field in TextOptions"
echo "  drawWrappedLabel() in retained renderer"
echo "  tile.text({ wrap: true }) in JS bridge"
echo "  wrapWidth property in JS bridge"
go test ./runtime/gfx/ -count=1 -run "TestWrapText|TestSurfaceTextWrap|TestExpandTextLines" -v

echo ""
echo "--- Feature 3: Per-Tile Invalidation Ergonomics ---"
echo "  tile.Draw(fn) auto-creates surface"
echo "  tile.Invalidate() marks dirty"
echo "  tile.draw(fn) JS bridge"
echo "  tile.invalidate() JS bridge"
echo "  SurfaceObject() exported from module_gfx"
go test ./runtime/ui/ -count=1 -run "TestTileDraw|TestTileInvalidate|TestTileWrap" -v

echo ""
echo "=== All LOUPE-016 features verified ==="
echo ""
echo "Commits implementing LOUPE-016:"
git log --oneline 95a46a1^..HEAD
