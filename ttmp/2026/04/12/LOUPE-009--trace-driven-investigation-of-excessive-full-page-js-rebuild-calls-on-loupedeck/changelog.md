# Changelog

## 2026-04-12

Created the `LOUPE-009` ticket as a dedicated follow-on investigation into the many full-page rebuild calls observed in the cyb-ito all-12 scene. The core purpose of this ticket is narrower than `LOUPE-007`: not just to say the scene is slow, but to explain exactly where rebuild calls come from in ordered runtime sequence and how they correlate with downstream flushes and writer/device output.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/index.md — Ticket overview and status entrypoint
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Detailed future task breakdown for the trace collector, JS bindings, scene breadcrumbs, Go correlation, and hardware evidence
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/design/01-textbook-trace-driven-investigation-of-excessive-full-page-rebuild-calls.md — Main intern-facing design and implementation guide for the trace-analysis work
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/playbooks/01-trace-capture-runbook.md — Operational runbook for future no-input and input-assisted trace captures
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/reference/01-implementation-diary.md — Chronological continuity log for why this ticket exists and how its design package was assembled

## 2026-04-12

Validated the new trace-analysis ticket with `docmgr doctor` and uploaded the initial design bundle to reMarkable. The uploaded bundle includes the ticket index, the main trace-analysis guide, the operational runbook, and the implementation diary under the remote folder `/ai/2026/04/12/LOUPE-009`.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Marked validation/upload/verification complete
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/index.md — Updated status summary after successful validation and upload

## 2026-04-12

Implemented Phase A of `LOUPE-009`: the generic trace collector substrate. The reusable `runtime/metrics` collector now supports bounded ordered trace events in addition to counters and timing summaries. This is the minimal storage layer needed before any JS- or Go-side trace breadcrumbs can be added. The implementation introduces a `TraceEvent` type, a default trace limit, a constructor that can override the trace capacity, collector APIs for `Trace(...)`, and snapshot/reset behavior that now includes the trace buffer. Tests were added for event ordering, sequence numbering, bounded-buffer eviction behavior, defensive field copying, and reset semantics. The full repository test suite passed after this change.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics.go — Added bounded trace event support to the generic collector
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics_test.go — Added coverage for trace ordering, bounds, copying, and reset behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Marked Phase A complete

## 2026-04-12

Implemented Phase B of `LOUPE-009`: reusable JS trace bindings on top of the new bounded collector substrate. The generic `pkg/jsmetrics` package now exposes low-level `metrics.trace(name, fields)` and higher-level `sceneMetrics.trace(name, fields)` APIs. Low-level trace calls write literal event names into the collector, while scene-helper trace calls automatically namespace event names with the helper prefix, making later dumps easier to interpret (`demo.renderAll.begin`, `scene.loop.tick`, etc.). Runtime tests were extended to prove that both the low-level and scene-helper APIs record ordered trace events with field values available through the collector snapshot. The full repository test suite passed after the change.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/jsmetrics/jsmetrics.go — Added low-level and scene-helper JS trace bindings plus JS object field decoding
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added coverage proving JS trace events reach the collector through both module styles
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Marked Phase B complete

## 2026-04-12

Implemented Phase C of `LOUPE-009`: scene-level breadcrumb instrumentation in the full-page all-12 workload itself. The current `examples/js/10-cyb-ito-full-page-all12.js` scene now emits trace breadcrumbs for `loop.tick`, `renderAll.begin`, `renderAll.end`, and `setActive` while preserving the existing counter/timing metrics (`recordLoopTick`, `recordRebuild`, `timeTile`, etc.). This keeps the first trace pass at the right boundary: scene events, not per-tile internals. The repository test suite passed after the script update, which confirms the instrumented example still boots under the owned JS runtime smoke coverage.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js — Added scene-level trace breadcrumbs while preserving the existing timing/counter metrics
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/examples_test.go — Indirectly validates the updated script by booting repo example scripts under test
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Marked Phase C complete

