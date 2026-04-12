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
- Inspected the side-strip implementation directly in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html` and extracted the specific left-strip dripping-bar and right-strip kanji-scroller behavior.
- Updated `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js` so the presenter-driven full-page scene now renders:
  - kanji tile chrome using the new font API,
  - a source-derived left dripping-bar strip on the `left` display,
  - a source-derived right horror-kanji strip on the `right` display,
  - source-inspired active-row pips on both side strips (`ea429f8` — `Add source-derived side strips to full-page cyb-ito scene`).
- Ran `go test ./...` successfully after the strip port.
- Ran a non-interactive hardware smoke command:
  - `timeout 30s go run ./cmd/loupe-js-live --script ./examples/js/10-cyb-ito-full-page-all12.js --duration 5s --send-interval 0ms --stats-interval 2s --log-render-stats`
- Recorded successful three-display evidence in `/tmp/loupe-cyb-ito-font-strips-1776033165.log`, including `Draw called Display=left`, `Draw called Display=main`, and `Draw called Display=right` with the expected `60x270`, `360x270`, and `60x270` dimensions.
