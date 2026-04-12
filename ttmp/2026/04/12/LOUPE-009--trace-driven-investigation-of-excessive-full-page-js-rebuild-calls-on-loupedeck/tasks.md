# Tasks

## Analysis and design

- [x] Create the LOUPE-009 ticket workspace
- [x] Write a detailed intern-facing design guide for rebuild-call trace analysis
- [x] Write an operational runbook for the future trace capture workflow
- [x] Write an implementation diary entry for continuity
- [x] Validate the ticket with `docmgr doctor`
- [x] Upload the design bundle to reMarkable
- [x] Verify the uploaded reMarkable files

## Future implementation phases

### Phase A: generic trace collector substrate

- [x] Extend `runtime/metrics/metrics.go` with a bounded ordered trace buffer
- [x] Add collector APIs for trace append, snapshot, and snapshot+reset
- [x] Add tests for event ordering, buffer bounds, and reset behavior

### Phase B: reusable JS trace bindings

- [ ] Add low-level JS `trace(...)` support through `pkg/jsmetrics`
- [ ] Add scene-helper `sceneMetrics.trace(...)`
- [ ] Add JS runtime tests proving trace events reach the collector correctly

### Phase C: scene-level breadcrumb instrumentation

- [ ] Instrument `examples/js/10-cyb-ito-full-page-all12.js` with `loop.tick`, `renderAll.begin`, `renderAll.end`, and `setActive` trace events
- [ ] Keep initial scene tracing at the scene boundary; do not yet instrument every tile renderer
- [ ] Preserve current timing/counter metrics alongside the new breadcrumbs

### Phase D: Go-side flush correlation

- [ ] Add minimal Go-side trace events around flush tick/begin/end in `cmd/loupe-js-live/main.go`
- [ ] Add optional trace output flags and a bounded dump path in the live runner
- [ ] If needed after the first trace run, add deeper trace points inside the renderer or writer

### Phase E: hardware evidence and interpretation

- [ ] Capture a first no-input hardware trace log for the full-page all-12 scene
- [ ] Compute derived ratios such as rebuilds-per-non-empty-flush and loop-ticks-per-non-empty-flush
- [ ] Summarize the trace findings in the ticket
- [ ] Decide whether cadence limiting should be the immediate next optimization

