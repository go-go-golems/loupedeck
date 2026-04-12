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

## 2026-04-11

Implemented Phase C as a pure-Go retained graphics package and Phase D as the first JS-facing graphics module. The new `runtime/gfx` package introduces Go-owned grayscale/additive retained surfaces with clear, fill, line, crosshatch, text, and additive compositing operations. The new `loupedeck/gfx` module exposes those surfaces to JS in a coarse, surface-oriented way rather than as a raw per-pixel transport API. This gives the runtime the first real graphics substrate needed for cyb-ito-style procedural scenes while preserving the rule that JavaScript describes scene work and Go still owns rendering semantics.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface.go — New retained grayscale/additive surface model with clear, fill, line, crosshatch, composite, and RGBA export helpers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/text.go — Go-owned text rasterization onto retained surfaces using a basic font face
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface_test.go — Focused unit tests for surface clearing, saturating add, line drawing, crosshatch, compositing, and text drawing
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go — First JS-facing `loupedeck/gfx` module exposing retained surfaces and coarse draw operations
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Registered the new `loupedeck/gfx` module in the owned runtime bootstrap
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added JS integration coverage that constructs, draws into, and composites `loupedeck/gfx` surfaces

## 2026-04-11

Implemented the first retained-surface composition slice on top of the new graphics package. Displays can now own `gfx` surfaces directly, surfaces notify their owning displays when they mutate, the retained renderer can render those attached surfaces through the existing Go-owned invalidation/writer stack, and the JS UI module can attach a `loupedeck/gfx` surface to a display via `display.surface(...)`. This is still not full multi-layer overlay composition yet, but it is the first real bridge from JS-authored graphics surfaces into retained display output.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface.go — Added retained surface change notifications so owning displays can become dirty when surface content changes
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/text.go — Text drawing now triggers retained surface change notifications like the other coarse drawing ops
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/display.go — Displays can now own a retained `gfx.Surface` and subscribe to its mutation notifications
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go — Retained display rendering now prefers attached surface output over placeholder text/icon rendering
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go — Exported surface unwrapping helper for cross-module use
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — Added `display.surface(surface)` so JS can attach a retained graphics surface to a display
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui_test.go — Added dirty-propagation coverage for display-owned surfaces
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/render_test.go — Added retained display-surface rendering coverage
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added JS integration coverage proving displays can own graphics surfaces

## 2026-04-11

Added the first cyb-ito-inspired multi-display JS prototype scene under `examples/js/07-cyb-ito-prototype.js`. The new script uses the retained display-region and `loupedeck/gfx` work from the earlier slices to render a `12`-tile main scene plus animated left/right strip content entirely through retained surfaces. This is still a prototype rather than a faithful final art port: it does not yet include ripple overlays or full multi-layer composition, but it proves that the runtime can now express a multi-display animated scene in JavaScript without bypassing Go-owned rendering.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js — First cyb-ito-inspired animated JS scene spanning left, main, and right display regions
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go — Fixed JS options decoding so omitted fields in `gfx.text(...)` no longer panic
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/text.go — Hardened text baseline clamping for small text boxes used by the prototype scene
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface_test.go — Added regression coverage for small-height text drawing used by the prototype

## 2026-04-12

Performed the first real hardware validation pass for the cyb-ito-inspired prototype scene on an actual Loupedeck Live via the tmux-based live-runner workflow. The first on-device observation confirmed that all three regions were rendering, but the interaction feedback was too ambiguous and the right-strip Kanji text degraded into unreadable fallback glyphs. After tightening the prototype to make active selection and last-event status much more obvious and replacing the side-strip text with readable ASCII fallback, a second hardware run was human-verified as working: button stepping and touchscreen activation were confirmed on the device, and the logs recorded matching `Button1`, `Button2`, and `Touch*` events while the scene continued to animate.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js — Improved prototype scene UX with stronger active-tile highlighting, central status readout, and ASCII side-strip fallback text
- /tmp/loupe-cyb-ito-prototype-1775989933.log — Hardware validation log showing the successful interactive run, including touch and button events during the live prototype session
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Live hardware validation entrypoint used for the tmux-based prototype run

## 2026-04-12

Implemented the core retained layer-composition slice for displays. A display can now own both a base `gfx` surface and an ordered set of named overlay layers, each with its own mutation subscription and stable draw order. The retained renderer composites the base surface first and then each named layer above it, the JS UI module now exposes `display.layer(name, surface)` for attaching or removing named layers, and the cyb-ito prototype was updated to use a dedicated `fx` overlay layer for the first touch-driven ripple effect rather than forcing all visuals into one surface.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/display.go — Added named retained display layers with stable order, dirty propagation, and attach/remove behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go — Added base-plus-layers display compositing in stable order
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — Added JS-facing `display.layer(name, surface)` attach/remove API
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui_test.go — Added retained layer dirty-propagation and stable-order coverage
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/render_test.go — Added retained display layer compositing coverage
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added JS integration coverage for named display layers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js — Added a dedicated `fx` overlay layer and the first touch-driven ripple overlay path

## 2026-04-12

Validated the new layered prototype on actual Loupedeck Live hardware after the overlay-composition slice landed. The rerun confirmed that the previously improved selection/status UX still worked under the new layer model and, importantly, that the first overlay-driven ripple/flash effect was visible on-device rather than only in tests. The live runner log captured matching touch events during the session, and the user explicitly confirmed the hardware result.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js — Hardware-validated layered prototype using the new `fx` overlay surface
- /tmp/loupe-cyb-ito-layers-1775990488.log — Hardware validation log for the layered prototype run after overlay support landed

