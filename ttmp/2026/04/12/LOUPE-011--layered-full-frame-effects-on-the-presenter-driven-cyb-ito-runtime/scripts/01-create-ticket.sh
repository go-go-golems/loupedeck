#!/usr/bin/env bash
set -euo pipefail
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test

docmgr ticket create-ticket \
  --ticket LOUPE-011 \
  --title "Layered full-frame effects on the presenter-driven cyb-ito runtime" \
  --topics javascript,rendering,animation,performance
