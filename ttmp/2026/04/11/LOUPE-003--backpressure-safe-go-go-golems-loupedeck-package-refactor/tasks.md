# Tasks

## TODO

### Phase 0 — Root module and baseline package port
- [x] Create top-level `go.mod` with module path `github.com/go-go-golems/loupedeck`
- [x] Add top-level `README.md` describing package scope and current device support
- [x] Port the current upstream baseline files into the repo root package
- [x] Ensure the root package builds with `go test ./...`
- [x] Commit Phase 0 baseline port

### Phase 1 — Safe lifecycle and composable input listeners
- [x] Replace single-slot button listener maps with multi-listener registration
- [x] Replace single-slot knob listener maps with multi-listener registration
- [x] Replace single-slot touch listener maps with multi-listener registration
- [x] Add subscription/unsubscribe support for listener cleanup
- [x] Change the read loop to return/report errors instead of panicking
- [x] Fix serial connection close so the underlying port is actually closed
- [x] Add tests for multi-listener dispatch and subscription cleanup
- [x] Commit Phase 1 lifecycle/listener improvements

### Phase 2 — B-lite outbound writer and pacing
- [x] Introduce an outbound command abstraction for transport writes
- [x] Add a single writer goroutine that owns all websocket writes
- [x] Route display draw messages through the writer queue
- [x] Route button color messages through the writer queue
- [x] Add configurable pacing interval and queue size options
- [x] Add writer stats/logging (queued, sent, failed, queue depth)
- [x] Add tests for send ordering and pacing behavior
- [x] Commit Phase 2 B-lite writer implementation

### Phase 3 — Port the feature tester onto the new package
- [x] Add a root-level feature tester command/example using the new package
- [x] Remove app-level sleep-based pacing from the tester
- [x] Validate the tester still builds cleanly against the new package
- [x] Commit the migrated feature tester

### Phase 4 — Full B groundwork: render invalidation and coalescing
- [x] Add a render scheduler with keyed invalidation
- [x] Route `Display.Draw` through the render scheduler when enabled
- [x] Coalesce repeated region invalidations so latest state wins per region
- [x] Add tests for coalescing and flushed command counts
- [x] Commit Phase 4 render scheduler groundwork

### Phase 5 — Documentation and continuity
- [x] Update the LOUPE-003 design doc with implementation deltas if the code deviates from plan
- [x] Append diary entries after each major implementation step/commit
- [x] Keep changelog and related-file bookkeeping current
- [x] Re-run `docmgr doctor --ticket LOUPE-003 --stale-after 30`

### Decision gate — assess whether C is needed
- [x] Run the hardware tester with B-lite only and capture behavior
- [x] Compare bounded send-rate behavior against the old sleep-based implementation
- [ ] Decide whether strict in-flight/ack-gated flow control (C) is still warranted
- [x] Re-run the hardware tester after a clean-exit cycle to separate transport stability from reconnect/reset issues

## Done

- [x] Write detailed architecture and implementation guide for B-lite then B
- [x] Relate evidence files and ticket docs
- [x] Run docmgr doctor and fix vocabulary/issues
- [x] Upload ticket bundle to reMarkable
