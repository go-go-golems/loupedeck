# Changelog

## 2026-04-11

Created the `LOUPE-006` ticket workspace for full animated JavaScript UIs, imported `~/Downloads/cyb-ito.html` as a tracked ticket source artifact, and wrote the first detailed intern-facing design package. The new guide explains what the imported HTML actually is, why the current retained tile API is not sufficient by itself, why JavaScript still must not own raw rendering or transport, and how to extend the runtime with display regions, Go-owned graphics surfaces, and layered retained scene composition.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html — Imported procedural canvas reference artifact that motivated the ticket
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/design/01-textbook-full-animated-javascript-ui-runtime-from-cyb-ito-reference.md — Main analysis/design/implementation guide for a new intern
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/reference/01-implementation-diary.md — Chronological continuity record for the ticket
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — Current JS UI API that the design guide identifies as too limited for cyb-ito-style scenes
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go — Current retained tile renderer that informs the next retained-scene step

## 2026-04-11

Validated the ticket with `docmgr doctor` and uploaded the first LOUPE-006 design bundle to reMarkable. The uploaded bundle contains the ticket index, the main intern-facing design guide, and the implementation diary, bundled into one PDF under the stable remote folder `/ai/2026/04/11/LOUPE-006`.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/index.md — Included in the uploaded bundle as the ticket overview
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/design/01-textbook-full-animated-javascript-ui-runtime-from-cyb-ito-reference.md — Included in the uploaded bundle as the main design guide
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/reference/01-implementation-diary.md — Included in the uploaded bundle as the continuity log

## 2026-04-11

Implemented the first Phase B runtime slice: retained JS-facing display regions for `left`, `main`, and `right`. The retained UI model now has explicit named displays, `page.tile(...)` delegates to the retained `main` display, the renderer can flush side-display placeholders in addition to main-grid tiles, and `cmd/loupe-js-live` now clears and flushes all three hardware display regions. This does **not** yet add a graphics/surface module, but it establishes the multi-region retained scene structure required before `loupedeck/gfx` can exist.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/display.go — New retained display-region type with text/icon/visible bindings and main-display tile ownership
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/page.go — Page model now owns named displays and delegates `page.tile(...)` to `main`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui.go — Added dirty-display tracking, active-page display filtering, and display-aware invalidation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/tile.go — Tile bindings now hang off the retained main display rather than a page-level tile map
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — Added `page.display(name, fn)` plus JS-facing display text/icon/visible bindings and main-display `display.tile(...)`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go — Renderer now supports flushing retained side displays in addition to main tiles
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Live runner now manages `left`, `main`, and `right` display targets instead of only `main`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui_test.go — Added retained display-region dirty/filtering compatibility tests
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/render_test.go — Added retained side-display render tests
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added JS integration coverage proving `page.display("left", ...)` works through the runtime

