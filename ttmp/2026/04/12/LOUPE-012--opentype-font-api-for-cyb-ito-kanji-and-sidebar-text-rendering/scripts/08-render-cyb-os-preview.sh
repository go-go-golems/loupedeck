#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

go run ./ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/scripts/06-render-scene-preview.go \
  --script ./examples/js/11-cyb-os-tiles.js \
  --out /tmp/loupe-cyb-os-tiles-preview.png \
  --wait 500ms
