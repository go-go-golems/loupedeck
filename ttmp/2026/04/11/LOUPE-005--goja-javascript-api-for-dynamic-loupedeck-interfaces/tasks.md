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
- [ ] Add `runtime/render` package scaffold
- [ ] Implement tile-to-main-display coordinate mapping (`90x90` tiles on `360x270`)
- [ ] Implement minimal retained tile rendering for icon/text output
- [ ] Bridge dirty tiles into current `Display.Draw()` path
- [ ] Ensure invalidation flows through current `renderer.go` scheduler
- [ ] Add tests for tile-region invalidation behavior
- [ ] Run `go test ./...`
- [ ] Commit milestone C
- [ ] Record milestone C in diary/changelog

### Implementation milestone D: host runtime shell
- [ ] Add `runtime/host` package scaffold
- [ ] Implement event routing from `OnButton`/`OnTouch`/`OnKnob`
- [ ] Implement timer services for host-owned scheduling
- [ ] Implement page-show lifecycle hooks
- [ ] Add unit tests for callback registration and event delivery
- [ ] Run `go test ./...`
- [ ] Commit milestone D
- [ ] Record milestone D in diary/changelog

### Implementation milestone E: first goja adapters
- [ ] Add `goja` dependency and any module/runtime support packages needed
- [ ] Add `runtime/js/module_state` with thin adapters over `runtime/reactive`
- [ ] Add `runtime/js/module_ui` with thin adapters over retained UI/runtime services
- [ ] Add initial module-loading integration test using `require(...)`
- [ ] Add first end-to-end example command that runs a JS page script
- [ ] Run `go test ./...`
- [ ] Commit milestone E
- [ ] Record milestone E in diary/changelog

### Implementation milestone F: animation and easing
- [ ] Add `runtime/anim` package scaffold
- [ ] Add `runtime/easing` package scaffold
- [ ] Implement basic tweens/timelines and easing curves
- [ ] Expose `loupedeck/anim` and `loupedeck/easing` modules
- [ ] Add integration tests for touch feedback / looping animation
- [ ] Run `go test ./...`
- [ ] Commit milestone F
- [ ] Record milestone F in diary/changelog

### Implementation milestone G: reconnect-safe retained replay
- [ ] Decide retained-runtime reconnect semantics
- [ ] Implement retained UI/state replay after reconnect
- [ ] Decide animation resume vs restart behavior after reconnect
- [ ] Add tests for retained replay logic where practical
- [ ] Run `go test ./...`
- [ ] Commit milestone G
- [ ] Record milestone G in diary/changelog
