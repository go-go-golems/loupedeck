# Tasks

## TODO

### Ticket setup and research framing
- [x] Create LOUPE-005 ticket workspace
- [x] Add brainstorm design doc, examples doc, and diary
- [x] Capture current renderer/writer constraints that shape the future JS API

### API brainstorming
- [x] Compare multiple JavaScript API styles (imperative, declarative, reactive, timeline-centric, hybrid)
- [x] Brainstorm animation and easing models
- [x] Brainstorm scene/page/state/runtime lifecycle models
- [x] Include tradeoffs and likely recommendation direction

### Example scenarios
- [x] Write multiple example scripts across different approaches
- [x] Include examples for easing curves, timelines, and interactive controls
- [x] Include dynamic paging, overlays, and reactive scenarios

### Future implementation planning
- [x] Choose a preferred JS module contract
- [x] Decide low-level vs high-level module boundaries (`loupedeck`, `ui`, `anim`, `easing`, etc.)
- [x] Decide host runtime semantics (`setTimeout`, `requestAnimationFrame`, fixed-timestep scheduler, etc.)
- [x] Decide state ownership model (JS-owned, Go-owned retained scene, or hybrid)
- [x] Decide what asset pipeline to expose to JS
- [ ] Decide how/if scripts persist across reconnects
- [x] Write a detailed intern-oriented reactive textbook
- [x] Write a detailed phased implementation plan

### Documentation and continuity
- [x] Start the diary
- [x] Update index/changelog/tasks coherently
- [x] Run `docmgr doctor --ticket LOUPE-005 --stale-after 30`
- [x] Commit ticket docs and bookkeeping
- [x] Upload the expanded document bundle to reMarkable

### Implementation milestone A: pure-Go reactive core
- [x] Add `runtime/reactive` package scaffold
- [x] Implement `Runtime` with batching and flush scheduling
- [x] Implement generic `Signal[T]` with `Get`, `Set`, `Update`
- [x] Implement generic `Computed[T]` with dependency tracking
- [x] Implement `Watch`/effect support for eager observers
- [x] Add cycle/reentrancy protection for computed/effect evaluation
- [x] Add unit tests for equality no-op behavior
- [x] Add unit tests for computed invalidation chains
- [x] Add unit tests for diamond dependency graphs
- [x] Add unit tests for batching semantics
- [x] Run `go test ./...`
- [x] Commit milestone A
- [x] Record milestone A in diary/changelog

### Implementation milestone B: retained UI model in pure Go
- [x] Add `runtime/ui` package scaffold
- [x] Implement page registry and active-page selection
- [x] Implement tile node model for `4x3` main touchscreen grid
- [x] Implement static property bindings for tile text/icon/visible
- [x] Implement reactive property bindings backed by `runtime/reactive`
- [x] Implement dirty-node tracking for retained UI changes
- [x] Add unit tests for page activation and visibility
- [x] Add unit tests for reactive tile property updates
- [x] Run `go test ./...`
- [x] Commit milestone B
- [x] Record milestone B in diary/changelog

### Implementation milestone C: retained visuals to current renderer bridge
- [x] Add `runtime/render` package scaffold
- [x] Implement tile-to-main-display coordinate mapping (`90x90` tiles on `360x270`)
- [x] Implement minimal retained tile rendering for icon/text output
- [x] Bridge dirty tiles into current `Display.Draw()` path
- [x] Ensure invalidation flows through current `renderer.go` scheduler
- [x] Add tests for tile-region invalidation behavior
- [x] Run `go test ./...`
- [x] Commit milestone C
- [x] Record milestone C in diary/changelog

### Implementation milestone D: host runtime shell
- [x] Add `runtime/host` package scaffold
- [x] Implement event routing from `OnButton`/`OnTouch`/`OnKnob`
- [x] Implement timer services for host-owned scheduling
- [x] Implement page-show lifecycle hooks
- [x] Add unit tests for callback registration and event delivery
- [x] Run `go test ./...`
- [x] Commit milestone D
- [x] Record milestone D in diary/changelog

### Implementation milestone E: first goja adapters
- [x] Add `goja` dependency and any module/runtime support packages needed
- [x] Add `runtime/js/module_state` with thin adapters over `runtime/reactive`
- [x] Add `runtime/js/module_ui` with thin adapters over retained UI/runtime services
- [x] Add initial module-loading integration test using `require(...)`
- [x] Add first end-to-end example command that runs a JS page script
- [x] Run `go test ./...`
- [x] Commit milestone E
- [x] Record milestone E in diary/changelog

### Implementation milestone F: animation and easing
- [x] Add `runtime/anim` package scaffold
- [x] Add `runtime/easing` package scaffold
- [x] Implement basic tweens/timelines and easing curves
- [x] Expose `loupedeck/anim` and `loupedeck/easing` modules
- [x] Add integration tests for touch feedback / looping animation
- [x] Run `go test ./...`
- [x] Commit milestone F
- [x] Record milestone F in diary/changelog

### Implementation milestone G: reconnect-safe retained replay
- [x] Decide retained-runtime reconnect semantics
- [x] Implement retained UI/state replay after reconnect
- [x] Decide animation resume vs restart behavior after reconnect
- [x] Add tests for retained replay logic where practical
- [x] Run `go test ./...`
- [x] Commit milestone G
- [x] Record milestone G in diary/changelog

### Implementation convergence phase H: adopt go-go-goja runtime ownership
- [x] Analyze `go-go-goja` runtime-owner / runtime-bridge / factory patterns for reuse
- [x] Write a dedicated design doc for converging the Loupedeck JS runtime onto go-go-goja runtime ownership
- [x] Decide dependency strategy: direct dependency on `go-go-goja` vs local port of `runtimeowner`/`runtimebridge`
- [x] Add an owner-runner layer to the local JS runtime bootstrap
- [ ] Add runtime-scoped bindings for owner/context/loop and Loupedeck services
- [ ] Refit all JS callback boundaries (`onButton`, `onTouch`, `onKnob`, timers, animation, reactive JS closures) to owner-thread scheduling
- [ ] Add tests for owner-thread callback serialization and shutdown behavior
- [ ] Refactor module wiring toward runtime-scoped registration rather than ad hoc env lookups where practical
- [ ] Add a hardware-backed JS live runner command
- [ ] Add multiple JS example scripts for live runtime validation
- [ ] Validate selected JS examples on actual Loupedeck Live hardware
- [x] Run `go test ./...`
- [x] Commit convergence phase H implementation work in focused steps
- [ ] Record convergence phase H progress in diary/changelog after each focused step
