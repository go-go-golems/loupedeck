# Changelog

## 2026-04-11

Created the LOUPE-005 ticket and wrote the initial goja/JavaScript API brainstorming package, including a deep design document, a multi-approach example-script reference, and a continuity diary.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md — Main design brainstorm for the future JS runtime and API shapes
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md — Multi-scenario example scripts spanning several design styles
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/02-implementation-diary.md — Chronological record for this ticket

## 2026-04-11

Added an intern-oriented textbook for the preferred reactive runtime and a detailed implementation plan that breaks the work into pure-Go runtime phases, retained-UI phases, goja adapter phases, tests, acceptance criteria, and PR-sized milestones.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Conceptual textbook explaining signals, mutation semantics, retained UI, animation, and host/runtime responsibilities for a new intern
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Detailed phased build plan and milestone map for implementing the preferred reactive runtime

## 2026-04-11

Validated the expanded LOUPE-005 ticket docs with `docmgr doctor`, committed the new textbook and implementation-plan package to git, and uploaded the full intern-oriented bundle to the reMarkable under the existing LOUPE-005 folder.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/tasks.md — Updated to mark validation, commit, and reMarkable upload complete
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/02-implementation-diary.md — Chronological continuity record for the expanded documentation and delivery work
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md — Included in the uploaded reMarkable bundle
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Included in the uploaded reMarkable bundle
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Included in the uploaded reMarkable bundle
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md — Included in the uploaded reMarkable bundle

## 2026-04-11

Implemented milestone A of the reactive runtime as a pure-Go `runtime/reactive` package with signals, computed values, batching, eager watch/effect support, dependency tracking, cycle/reentrancy protection, and a focused unit-test suite. The implementation intentionally stayed goja-free so the semantic core could be validated in isolation before any JS bindings are added.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/runtime.go — Runtime coordination for batching, collector scoping, pending effect queues, and flush behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/graph.go — Dependency graph primitives, dependent/source tracking, and default equality helpers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/signal.go — Generic signal implementation with `Get`, `Set`, and `Update`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/computed.go — Generic computed implementation with lazy reevaluation and dirty propagation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/effect.go — Eager watch/effect implementation plus stop/unsubscribe support
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/runtime_test.go — Unit tests covering equality no-ops, invalidation chains, diamond graphs, batching, stop behavior, and panic guards

## 2026-04-11

Implemented milestone B as a pure-Go retained UI layer on top of `runtime/reactive`, including page registration, active-page switching, `4x3` main-display tile nodes, static and reactive text/icon/visible bindings, and dirty-tile tracking suitable for a later renderer bridge.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui.go — Top-level retained UI runtime, active-page selection, and dirty-tile collection/filtering
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/page.go — Page model and `4x3` tile coordinate validation/lookup
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/tile.go — Tile state, static setters, reactive bindings, and dirty marking
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui_test.go — Unit tests for page activation, hidden-page filtering, static properties, and reactive tile property updates

## 2026-04-11

Implemented milestone C as a retained-tile visual bridge in `runtime/render`, including `90x90` tile-to-main-display coordinate mapping, minimal icon/text tile rendering, and a flush path that can target any `Draw(image, x, y)` implementation — including the existing `*loupedeck.Display` output path that already flows through `display.go` and the current render scheduler.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go — Retained visual renderer, tile rectangle mapping, placeholder tile composition, and flush logic
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/render_test.go — Tests for tile-coordinate mapping, flush behavior, and preservation of hidden-page dirty tiles
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui.go — Added `ClearDirtyTiles(...)` so active-page flushes do not erase hidden-page dirty state
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/display.go — Existing `Display.Draw(image, x, y)` path that the new render layer is designed to plug into without bypassing transport ownership

## 2026-04-11

Implemented milestone D as a host runtime shell in `runtime/host`, covering attachable routing for the current `OnButton` / `OnTouch` / `OnKnob` listener APIs, page-show lifecycle hooks, and host-owned timeout/interval timers. This creates the runtime services that future goja modules can call without embedding lifecycle policy directly in the JS layer.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/runtime.go — Host runtime state, event-source attachment, timer bookkeeping, and shutdown behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/events.go — Event registration, subscription cleanup, and bridging to the current Loupedeck listener APIs
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/pages.go — Page-show hooks and `Show(...)` lifecycle routing
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/timers.go — Host-owned timeout and interval timers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/runtime_test.go — Unit tests for callback delivery, page-show hooks, and timer behavior

## 2026-04-11

Implemented milestone E as the first goja adapter slice, adding native `require("loupedeck/state")` and `require("loupedeck/ui")` modules on top of the new pure-Go runtime layers, plus a small JS demo command that renders a script-defined page into PNG tiles. This is the first point where a JS script can create reactive state, define a retained page, register UI callbacks, and drive the retained Go runtime end to end.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/go.mod — Added `goja` / `goja_nodejs` runtime dependencies needed for native-module loading via `require(...)`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/go.sum — Updated dependency lockfile after adding the first goja slice and running `go mod tidy`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/env/env.go — Shared environment bootstrap joining reactive, UI, and host runtime services
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — goja runtime construction and native-module registration
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_state/module.go — `loupedeck/state` native module exposing signals, computed values, batching, and watchers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — `loupedeck/ui` native module exposing pages, tiles, show, and input callback registration
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Integration tests proving `require(...)` module loading and JS-driven reactive page updates
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-demo/main.go — First end-to-end example command that runs a JS page script and renders dirty tiles to PNG files
