# Changelog

## 2026-04-12

- Created ticket `LOUPE-012` for adding a JS-facing OpenType font API.
- Added the main implementation plan for Go-side font loading, JS font handles, and cyb-ito kanji/sidebar integration.
- Added the implementation diary.
- Added the phased task breakdown.
- Implemented the Go-side font loader/cache in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/font.go` with tests in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/font_test.go` (`07d8f6d` — `Add gfx OpenType font loader and cache`).
- Added support for loading `.ttf/.otf` fonts and `.ttc` collections with face index selection.
- Implemented JS font handles in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go` and JS runtime coverage in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go` (`8091029` — `Add JS gfx font handles and text font option`).
- Added first cyb-ito integration in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js`, using `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc` to render actual kanji tile labels and a kanji sidebar scroller (`0aa9fc2` — `Render cyb-ito prototype kanji with loaded font`).
- Archived ticket-local reproducibility scripts under `scripts/`.
