# Changelog

## 2026-04-12

Created the `LOUPE-010` ticket as the forward-only presenter-refactor track. This ticket exists because the current full-page cyb-ito runtime is now known to use the wrong control model: the animation loop directly triggers whole-frame redraws while the flush path is still busy. The new ticket records the intended architecture instead: simulation-paced state changes plus flush-gated one-frame-in-flight presentation.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/index.md — Ticket overview and status entrypoint
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/design/01-implementation-plan-simulation-paced-state-with-flush-gated-presentation.md — Main implementation plan for the refactor
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/reference/01-implementation-diary.md — Continuity log for the refactor
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/tasks.md — Detailed phased task list for the implementation

## 2026-04-12

Implemented Phase A of `LOUPE-010`: the pure-Go one-frame-in-flight presenter runtime. The new `runtime/present` package supports render callback registration, flush callback registration, `Invalidate(reason)` with latest-reason-wins dirty coalescing, explicit start/shutdown lifecycle, and strictly serial render/flush processing. Unit tests were added for coalescing while flush is busy, deferred presentation when invalidation happens before callbacks are installed, serial presentation ordering, and shutdown behavior. The full repository test suite passed after the new package landed.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/present/runtime.go — New presenter runtime implementing one-frame-in-flight coalesced presentation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/present/runtime_test.go — Tests for coalescing, serial presentation, and shutdown behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/tasks.md — Marked Phase A complete and recorded the initial scripts archive

## 2026-04-12

Implemented Phase B of `LOUPE-010`: JS environment ownership and the new `loupedeck/present` module. The environment now owns a presenter runtime, the owned JS runtime registers a new native module, and JS can now register a frame callback with `present.onFrame(fn)` and request presentation with `present.invalidate(reason)`. Runtime tests were added to prove that JS frame callbacks are invoked with the correct reason and that repeated invalidations coalesce to the latest reason across a blocked flush boundary. The full repository test suite passed after this change.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/env/env.go — Added presenter ownership to the runtime environment
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Registered the new `loupedeck/present` module
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_present/module.go — New JS presentation module exposing `onFrame` and `invalidate`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added tests proving frame-callback execution and latest-reason invalidation semantics
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/tasks.md — Marked Phase B complete

## 2026-04-12

Implemented the main runtime behavior change for `LOUPE-010`: the live runner now treats the presenter as the primary frame-production path, and the full-page all-12 cyb-ito scene now uses `loupedeck/present` instead of calling `renderAll("loop")` directly from the animation loop. `cmd/loupe-js-live/main.go` now installs the presenter flush callback, starts the presenter, removes the old periodic full-page flush ticker from the intended presentation path, and accumulates render statistics from actual presenter-driven flushes. The scene script now registers `present.onFrame(reason => renderAll(reason))`, invalidates on initial boot, invalidates on every loop tick after updating state, and invalidates from input paths instead of forcing immediate full-page redraws. The full repository test suite passed after these changes.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Refactored to use presenter-driven frame production instead of the old periodic full-page flush ticker
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js — Migrated from loop-driven redraws to `loupedeck/present`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/scripts/04-go-test-phase-cd-live-runner-and-scene.sh — Repro script for the live-runner and scene migration validation step
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/tasks.md — Marked the main Phase C and Phase D tasks complete

## 2026-04-12

Completed the first validation pass of the presenter-driven architecture and recorded the result in ticket-local scripts and ticket docs. The new validation scripts capture and analyze a no-input presenter trace (`05-capture-presenter-trace.sh` and `06-analyze-presenter-trace.py`). The resulting hardware trace `/tmp/loupe-cyb-ito-full10-presenter-trace-1776027682.log` shows a decisive architectural improvement: the old trace baseline averaged about `33.6` rebuilds per non-empty flush, while the new presenter-driven trace shows exactly `1` rebuild per non-empty flush across `389` non-empty flushes. Loop ticks still occur more frequently than frames are presented, which is correct for the new model, but repeated invalidations are now being coalesced instead of triggering many redundant full-page rebuilds. Based on this evidence, deeper renderer/writer trace work is not the immediate next priority.

### Related Files

- /tmp/loupe-cyb-ito-full10-presenter-trace-1776027682.log — First hardware trace for the new presenter-driven architecture
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/scripts/05-capture-presenter-trace.sh — Reproducible no-input capture command for the new architecture
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/scripts/06-analyze-presenter-trace.py — Analysis script that computes rebuilds-per-flush for the new architecture
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/tasks.md — Marked Phase E complete and recorded that deeper tracing is not currently the top priority

