#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

docmgr ticket create-ticket \
  --ticket LOUPE-012 \
  --title "OpenType font API for cyb-ito kanji and sidebar text rendering" \
  --topics javascript,rendering,fonts,unicode
