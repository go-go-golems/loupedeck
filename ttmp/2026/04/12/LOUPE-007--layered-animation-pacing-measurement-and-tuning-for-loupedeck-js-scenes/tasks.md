# Tasks

## Analysis and design

- [x] Create the LOUPE-007 ticket workspace
- [x] Write a detailed intern-facing design and implementation guide for layered-scene pacing measurement
- [x] Write an operational runbook for future layered-density sweeps
- [x] Write an implementation diary entry for continuity
- [x] Validate the ticket with `docmgr doctor`
- [x] Upload the design bundle to reMarkable
- [x] Verify the uploaded reMarkable files

## Future implementation phases

### Phase A: live-runner instrumentation

- [ ] Add periodic renderer statistics logging to `cmd/loupe-js-live/main.go`
- [ ] Add writer statistics logging to `cmd/loupe-js-live/main.go`
- [ ] Add scene mode / workload labels to live-runner logs
- [ ] Keep stats aggregation lightweight enough that logging does not dominate the measured workload

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
