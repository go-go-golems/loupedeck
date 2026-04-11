# Tasks

## TODO

### Phase 0 — Root module and baseline package port
- [ ] Create top-level `go.mod` with module path `github.com/go-go-golems/loupedeck`
- [ ] Add top-level `README.md` describing package scope and current device support
- [ ] Port the current upstream baseline files into the repo root package
- [ ] Ensure the root package builds with `go test ./...`
- [ ] Commit Phase 0 baseline port

### Phase 1 — Safe lifecycle and composable input listeners
- [ ] Replace single-slot button listener maps with multi-listener registration
- [ ] Replace single-slot knob listener maps with multi-listener registration
- [ ] Replace single-slot touch listener maps with multi-listener registration
- [ ] Add subscription/unsubscribe support for listener cleanup
- [ ] Change the read loop to return/report errors instead of panicking
- [ ] Fix serial connection close so the underlying port is actually closed
- [ ] Add tests for multi-listener dispatch and subscription cleanup
- [ ] Commit Phase 1 lifecycle/listener improvements

### Phase 2 — B-lite outbound writer and pacing
- [ ] Introduce an outbound command abstraction for transport writes
- [ ] Add a single writer goroutine that owns all websocket writes
- [ ] Route display draw messages through the writer queue
- [ ] Route button color messages through the writer queue
- [ ] Add configurable pacing interval and queue size options
- [ ] Add writer stats/logging (queued, sent, failed, queue depth)
- [ ] Add tests for send ordering and pacing behavior
- [ ] Commit Phase 2 B-lite writer implementation

### Phase 3 — Port the feature tester onto the new package
- [ ] Add a root-level feature tester command/example using the new package
- [ ] Remove app-level sleep-based pacing from the tester
- [ ] Validate the tester still builds cleanly against the new package
- [ ] Commit the migrated feature tester

### Phase 4 — Documentation and continuity
- [ ] Update the LOUPE-003 design doc with implementation deltas if the code deviates from plan
- [ ] Append diary entries after each major implementation step/commit
- [ ] Keep changelog and related-file bookkeeping current
- [ ] Re-run `docmgr doctor --ticket LOUPE-003 --stale-after 30`

### Decision gate — assess whether C is needed
- [ ] Run the hardware tester with B-lite only and capture behavior
- [ ] Compare bounded send-rate behavior against the old sleep-based implementation
- [ ] Decide whether strict in-flight/ack-gated flow control (C) is still warranted

## Done

- [x] Write detailed architecture and implementation guide for B-lite then B
- [x] Relate evidence files and ticket docs
- [x] Run docmgr doctor and fix vocabulary/issues
- [x] Upload ticket bundle to reMarkable
