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

## 2026-04-11

Implemented milestone F as the first animation/easing slice, adding pure-Go easing functions, a host-backed numeric tween/loop/timeline runtime, and `require("loupedeck/anim")` / `require("loupedeck/easing")` native modules. This extends the first JS slice from “static reactive pages” to “reactive pages with host-managed animation primitives”.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/easing/easing.go — Core easing curves and `steps(n)` factory used by the new animation layer
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/easing/easing_test.go — Unit tests for easing endpoints and stepped easing behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/anim/runtime.go — Host-backed tweens, loops, and sequential timelines for numeric targets
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/anim/runtime_test.go — Unit tests for tween completion, loop progress, and sequential timeline execution
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_anim/module.go — `loupedeck/anim` native module exposing `to(...)`, `loop(...)`, and `timeline()`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_easing/module.go — `loupedeck/easing` native module exposing easing functions to JS
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — JS integration tests covering button-triggered tweens and loop-driven reactive updates

## 2026-04-11

Implemented milestone G as the first reconnect-safe retained replay slice. The chosen behavior is intentionally conservative: when the host decides a reconnect requires visual restoration, it can explicitly re-invalidate the currently active retained page so the renderer redraws it, without rerunning page-show hooks and without attempting to reconstruct or restart animation timelines automatically.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui.go — Added explicit active-page invalidation support for retained visual replay
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/pages.go — Added `ReplayActivePage()` reconnect/replay entry point with non-hook-replaying semantics
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/host/runtime_test.go — Added coverage proving replay marks tiles dirty again without rerunning page-show hooks

## 2026-04-11

Analyzed the `go-go-goja` runtime-owner architecture and added a dedicated convergence plan for migrating the current Loupedeck JS runtime onto that owner-thread / runtime-bridge / factory-style execution model before building serious hardware-backed JS demos. Added the corresponding convergence-phase tasks to the ticket so the next work is tracked explicitly.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/04-implementation-plan-converge-the-loupedeck-js-runtime-onto-go-go-goja-runtime-ownership.md — Detailed next-phase plan for adopting go-go-goja runtime ownership patterns
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/tasks.md — Added convergence-phase H tasks for owner-runner integration, runtime bindings, live runner work, and hardware-backed examples
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/index.md — Updated the ticket landing page to include the new convergence plan and next-phase status
- /home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/runtimeowner/runner.go — Source of truth for the owner-thread runner pattern we intend to adopt
- /home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/runtimebridge/runtimebridge.go — Source of truth for runtime-scoped owner/context bindings
- /home/manuel/code/wesen/corporate-headquarters/go-go-goja/engine/factory.go — Source of truth for owned-runtime composition and runtime-scoped module registration

## 2026-04-11

Started convergence phase H in code by choosing a **local port** of `go-go-goja`'s `runtimeowner` package rather than a direct repository dependency for now, then refactored the local JS bootstrap into an owned runtime with an event loop, owner runner, `RunString(...)`, and explicit `Close(...)` lifecycle. This is the first concrete step toward making hardware-backed JS execution owner-thread safe.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/errors.go — Local port of the runtimeowner error contract
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/types.go — Local port of scheduler/runner type definitions
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/runner.go — Local port of the owner-thread runner implementation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/runner_test.go — Local tests for owner-thread scheduling, cancellation, panic recovery, and leaked-owner-context behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/runner_race_test.go — Local stress test for concurrent runner calls
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Refactored JS bootstrap into an owned runtime with `VM`, `Loop`, `Owner`, `Env`, `RunString`, and `Close`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Updated JS integration tests to run through the new owner-backed runtime API
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-demo/main.go — Updated the demo command to use the owned JS runtime lifecycle

## 2026-04-11

Completed H2 by adding a local `runtimebridge` package and wiring it into the owned JS runtime. The runtime now stores owner/context/loop bindings plus the Loupedeck runtime environment against the VM, and cleans them up on close. This establishes the runtime-scoped binding layer needed for the next callback-refit step.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimebridge/runtimebridge.go — Local port of runtime-scoped VM bindings for owner/context/loop/value lookup
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimebridge/runtimebridge_test.go — Tests for store/lookup/delete behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Owned JS runtime now installs runtime bridge bindings and exposes runtime context cleanup semantics
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Tests proving bindings are installed and removed correctly

## 2026-04-11

Completed H3 by refitting the live JS callback boundaries onto the owner-runner model. Hardware event callbacks now post back onto the owner thread, animation loops settle JS callbacks through the owner, numeric animation targets/easing functions use owner-aware call paths, and the reactive JS closures used by `state.computed`, `state.watch`, `signal.update`, and tile text/icon/visible bindings now run through the same owner-aware discipline. Added tests for async callback application, concurrent event serialization, and post-close callback suppression.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — Refit UI event callbacks and reactive tile-binding closures to owner-aware execution paths
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_anim/module.go — Refit animation callbacks, numeric target access, and easing callbacks to owner-aware execution paths
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_state/module.go — Refit computed/watch/update reactive JS closures to owner-aware execution paths
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added tests covering asynchronous button callback application, concurrent event serialization, and no-op callback behavior after runtime close
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/runner.go — Added public owner-context helper used by module code at known owner-thread entry points

## 2026-04-11

Completed H4 by cleaning up module wiring so the JS modules no longer depend on ad hoc environment threading from `runtime/js/runtime.go`. Instead, the environment is now resolved from runtime-scoped bindings, which brings the local module bootstrap closer to the runtime-scoped registration discipline used in `go-go-goja`.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/env/env.go — Added typed environment lookup from runtime-scoped bindings and a binding-key constant
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Simplified module registration to stop hand-threading env into each module registrar
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_state/module.go — State module now resolves the environment from runtime bindings
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — UI module now resolves the environment from runtime bindings
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_anim/module.go — Anim module now resolves the environment from runtime bindings

## 2026-04-11

Added a real hardware-backed JS live runner plus an initial example pack, then validated selected scripts on the Loupedeck Live. The new `cmd/loupe-js-live` command connects to the device, attaches the host runtime to live events, runs a JS script through the owned runtime, and continuously flushes retained UI to the main display. Added six example scripts and a non-hardware smoke test that boots each script. Hardware validation succeeded for a static hello page and for an auto-running pulse animation after one reconnect retry; the first failed pulse attempt reproduced the known serial/websocket reconnect fragility rather than a JS runtime error.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — New hardware-backed JS live runner command
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/01-hello.js — Static hello-page example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/02-counter-button.js — Button-driven counter example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/03-knob-meter.js — Knob-driven numeric example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/04-touch-feedback.js — Touch-driven feedback example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/05-pulse-animation.js — Auto-running pulse animation example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/06-page-switcher.js — Multi-page switching example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/examples_test.go — Smoke test that boots every example script under the owned runtime

## 2026-04-11

Extended the live-runner validation from the initial static/animation smoke checks to the interactive example set on actual Loupedeck Live hardware. Human-in-the-loop validation confirmed that `02-counter-button.js` updates from Circle presses, `03-knob-meter.js` responds to `Knob1`, `04-touch-feedback.js` responds to `Touch1` / `Touch6` / `Touch12`, and `06-page-switcher.js` switches pages from `Button1` / `Button2`. During this pass, `04-touch-feedback.js` was corrected so its visible labels match the actual touched regions, and the live runner gained optional high-level event logging to make hardware validation easier to verify from logs and tmux pane capture.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Added optional button/touch/knob event logging for live hardware validation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/04-touch-feedback.js — Fixed the visible tile layout so `Touch1`, `Touch6`, and `Touch12` labels line up with the actual touched regions
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/02-counter-button.js — Human-validated Circle-button counter example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/03-knob-meter.js — Human-validated `Knob1` reactive state example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/06-page-switcher.js — Human-validated retained page switching example
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md — Updated to distinguish exploratory API sketches from the concrete live example pack that now exists in the repo

## 2026-04-11

Added real end-user documentation for the implemented JS runtime in two places: Glazed-formatted help pages inside the repo and a durable Obsidian technical deep dive. The repo now contains a step-by-step user guide/tutorial for the live runner plus a detailed API reference for the currently implemented modules (`loupedeck/state`, `loupedeck/ui`, `loupedeck/anim`, `loupedeck/easing`). The vault now contains a deep implementation article that explains the owner-thread runtime, reactive core, retained UI, animation layer, and live hardware path as one coherent architecture.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/docs/help/tutorials/01-build-your-first-live-loupedeck-js-script.md — Glazed-formatted step-by-step user guide for writing and running a live Loupedeck JS script
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/docs/help/topics/01-loupedeck-js-api-reference.md — Glazed-formatted API reference for the current JS runtime modules and live runner
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/README.md — Repository discoverability update linking the new help pages
- /home/manuel/code/wesen/obsidian-vault/Projects/2026/04/11/ARTICLE - Loupedeck - Goja JavaScript Runtime and API Deep Dive.md — Durable technical article covering the implemented JS runtime architecture and API internals
