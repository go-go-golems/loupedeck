# Tasks

## Analysis and design

- [x] Create the LOUPE-007 ticket workspace
- [x] Write a detailed intern-facing design and implementation guide for layered-scene pacing measurement
- [x] Write an operational runbook for future layered-density sweeps
- [x] Write an implementation diary entry for continuity
- [x] Write a full project technical report on the 12-tile cyb-ito performance investigation, approaches tried, and current hypotheses
- [x] Validate the ticket with `docmgr doctor`
- [x] Upload the design bundle to reMarkable
- [x] Verify the uploaded reMarkable files

## Future implementation phases

### Phase 0: frame-atomic retained surface groundwork

- [x] Diagnose the full-page redraw artifact where later tiles visibly appear only on some frames
- [x] Identify the real problem as mid-frame renderer snapshots of a shared retained surface, not just generic slowness
- [x] Add retained `gfx.Surface` batching so many mutations can be coalesced into one change notification
- [x] Ensure renderer reads wait for an in-flight surface batch to complete before snapshotting
- [x] Expose JS-facing `surface.batch(() => ...)` in `loupedeck/gfx`
- [x] Apply batching to the new full-page all-12 scene so `renderAll()` builds one coherent frame before marking the display dirty
- [x] Compare the full-page all-12 scene before/after batching on real hardware and record the visual result

### Phase A: live-runner instrumentation

- [x] Add periodic renderer statistics logging to `cmd/loupe-js-live/main.go`
- [x] Add writer statistics logging to `cmd/loupe-js-live/main.go`
- [x] Add scene mode / workload labels to live-runner logs
- [x] Add JS-side metrics collection so scenes can record their own timing from inside JavaScript
- [ ] Keep stats aggregation lightweight enough that logging does not dominate the measured workload
- [x] Validate the new stats path on real hardware and capture one first evidence log
- [x] Rerun the full-page all-12 scene after adding rebuild-reason metrics and confirm which path is driving rebuild frequency on hardware

### Phase B: density-sweep scene modes

- [ ] Add controlled scene-density modes to `examples/js/07-cyb-ito-prototype.js`
- [ ] Support at least `base`, `hud`, `scan`, `ripple`, `main-full`, and `full`
- [ ] Keep the scene comparable across modes so results stay interpretable

### Phase C: controlled hardware sweep

- [ ] Run the raw hardware benchmark as the control baseline
- [ ] Run layered prototype sweeps across scene-density modes on actual hardware
- [ ] Sweep flush interval independently of writer pacing first
- [ ] Sweep writer pacing independently after the first density results exist
- [ ] Capture quantitative logs plus qualitative observer notes

### Phase D: analysis and tuning

- [ ] Summarize renderer-side findings
- [ ] Summarize writer/transport-side findings
- [ ] Separate effect-design timing issues from true pacing issues
- [ ] Decide whether any runtime or scene tuning changes are warranted
- [ ] If tuning is needed, commit measurement and tuning changes separately
