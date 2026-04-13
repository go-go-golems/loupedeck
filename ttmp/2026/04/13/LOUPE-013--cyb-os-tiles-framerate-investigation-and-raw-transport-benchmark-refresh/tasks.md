# Tasks

## TODO

- [x] Run `go run ./cmd/loupe-fps-bench` on the current tree and capture fresh raw-writer baseline numbers for full-screen, single-button, and mixed-button workloads.
- [x] Compare the fresh `loupe-fps-bench` results against the observed `cyb-os-tiles` full three-display flush cadence.
- [x] Add minimal JS path probes for main-only, three-display, and main-fast/side-slow live-runner paths under this ticket.
- [x] Identify the default live-runner render-scheduler cap on the current tree.
- [ ] Create a controlled `cyb-os-tiles` variant where `main` redraws every frame but `left` and `right` redraw at a lower cadence.
- [ ] Determine how much of the remaining `cyb-os-tiles` slowdown comes from scene-generation work versus display-update shape after accounting for the default 40ms render-scheduler cap.
- [x] Decide whether `cmd/loupe-js-live` should expose render flush interval / scheduler options for measurement and tuning.
- [ ] Measure `cyb-os-tiles` itself with reduced `--flush-interval` values and compare the benefit against the simpler JS probes.
- [ ] Summarize the findings in a concise design/report doc with explicit guidance for future scene authors.
