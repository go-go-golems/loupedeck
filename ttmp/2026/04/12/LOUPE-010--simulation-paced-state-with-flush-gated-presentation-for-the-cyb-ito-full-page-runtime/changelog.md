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

