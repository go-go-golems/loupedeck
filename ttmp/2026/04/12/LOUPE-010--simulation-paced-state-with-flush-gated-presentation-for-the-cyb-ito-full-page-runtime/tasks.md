# Tasks

## Analysis and design

- [x] Create the LOUPE-010 ticket workspace
- [x] Write the forward-only implementation plan for the presenter refactor
- [x] Write an implementation diary entry for continuity
- [x] Archive reproduction scripts in `scripts/` with numeric `XX-...` prefixes as the work proceeds

## Phase A: pure-Go presenter runtime

- [x] Add `runtime/present/` with a one-frame-in-flight presenter runtime
- [x] Support render callback registration
- [x] Support flush callback registration
- [x] Support `Invalidate(reason)` with coalesced dirty state and latest-reason wins
- [x] Add runtime tests covering coalescing, serial presentation, and shutdown behavior

## Phase B: JS environment and module wiring

- [x] Add presenter ownership to `runtime/js/env/env.go`
- [x] Register a new JS module for presentation control in `runtime/js/runtime.go`
- [x] Add `runtime/js/module_present/module.go`
- [x] Expose `present.onFrame(fn)`
- [x] Expose `present.invalidate(reason)`
- [x] Add JS runtime tests proving JS frame callbacks and invalidation semantics work correctly

## Phase C: live-runner refactor

- [x] Refactor `cmd/loupe-js-live/main.go` to use the presenter as the primary frame-production path
- [x] Remove the current full-page periodic flush ticker as the intended presentation model
- [x] Wire presenter render callback settlement onto the JS owner thread
- [x] Wire presenter flush callback to `renderer.Flush()`
- [ ] Add presenter-focused trace/metrics breadcrumbs if needed

## Phase D: scene migration

- [x] Migrate `examples/js/10-cyb-ito-full-page-all12.js` to `loupedeck/present`
- [x] Remove direct `renderAll("loop")` calls from the animation loop
- [x] Make simulation update state and invalidate presentation only
- [x] Make input paths invalidate presentation rather than forcing immediate full-page redraws

## Phase E: validation and interpretation

- [ ] Store concrete validation commands in `scripts/`
- [ ] Run the full test suite
- [ ] Capture a new no-input hardware trace log
- [ ] Compare rebuilds-per-flush against the old trace baseline
- [ ] Decide whether any deeper renderer/writer tracing is still necessary

