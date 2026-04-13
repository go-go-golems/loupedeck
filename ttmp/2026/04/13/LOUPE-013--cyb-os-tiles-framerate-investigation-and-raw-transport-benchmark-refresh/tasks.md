# Tasks

## TODO

- [ ] Run `go run ./cmd/loupe-fps-bench` on the current tree and capture fresh raw-writer baseline numbers for full-screen, single-button, and mixed-button workloads.
- [ ] Compare the fresh `loupe-fps-bench` results against the observed `cyb-os-tiles` full three-display flush cadence.
- [ ] Create a controlled `cyb-os-tiles` variant where `main` redraws every frame but `left` and `right` redraw at a lower cadence.
- [ ] Determine whether the main bottleneck is dominated by per-draw protocol overhead, pixel payload size, or side-display scene-generation cost.
- [ ] Summarize the findings in a concise design/report doc with explicit guidance for future scene authors.
