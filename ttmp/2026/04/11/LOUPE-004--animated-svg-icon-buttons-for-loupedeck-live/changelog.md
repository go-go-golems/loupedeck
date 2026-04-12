# Changelog

## 2026-04-11

Created the LOUPE-004 ticket, added the initial design/diary documents, and imported the full HTML icon library into the ticket workspace for later extraction and rendering.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/design-doc/01-animated-svg-icon-button-rendering-plan.md — Initial design plan for extraction, rasterization, scaling, and animation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/reference/01-implementation-diary.md — Chronological implementation record for the new ticket
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/sources/local/macos1-icon-library.html — Imported HTML icon library used as the SVG source set

## 2026-04-11

Implemented the SVG icon extraction/normalization/rasterization pipeline and added a root command that animates a 12-button icon bank on the Loupedeck Live.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/svg_icons.go — Loader, normalizer, rasterizer, and scaling helpers for the imported icon library
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/svg_icons_test.go — Tests for extraction, normalization, and rasterization behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-svg-buttons/main.go — Root hardware demo command for animated SVG-backed buttons
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/go.mod — Added Go SVG rasterization dependencies

## 2026-04-11

Ran the animated SVG button demo on actual hardware, then reduced default log noise so the demo remains usable while still exposing lifecycle warnings that matter.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-svg-buttons/main.go — Lower default log level and concise startup logging for the hardware demo
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/reference/01-implementation-diary.md — Captures the exact hardware-validation commands and warning output

## 2026-04-11

Extended the SVG demo with icon-bank paging, curated icon selection, starting offsets, automatic page cycling, and live bank switching through both physical buttons and touch controls.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-svg-buttons/main.go — Adds `--offset`, `--icons`, `--page-every`, bank state management, and button/touch control bindings
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-svg-buttons/main_test.go — Tests curated ordering, offset rotation, and bank padding behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/reference/01-implementation-diary.md — Records the failed busy-port attempt and the successful banked hardware run
