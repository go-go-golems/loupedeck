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

## 2026-04-12

Implemented the first half of Phase D of `LOUPE-009`: Go-side flush correlation and live-runner trace dump controls. `cmd/loupe-js-live` now supports trace-specific flags (`--log-js-trace`, `--log-go-trace`, `--trace-limit`, `--trace-dump-on-exit`) and can emit ordered trace events for `go.flush.tick`, `go.flush.begin`, and `go.flush.end` using the same generic metrics collector as the JS trace events. The live runner now also uses a configurable trace-capacity collector so the trace buffer can be tuned per run, and it dumps trace events through the existing stats window/reset path rather than inventing a parallel output channel. The full repository test suite passed after this change.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Added trace flags, trace-limit wiring, Go-side flush trace events, and trace dump formatting/output
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Marked the first two Phase D tasks complete

## 2026-04-12

Archived the concrete reproduction scripts used during `LOUPE-009` into the ticket’s `scripts/` directory using numeric `XX-...` prefixes, as requested. This now captures the ticket creation command, validation/upload commands, phase-specific test commands, the no-input trace capture command, and the trace analysis script used to compute rebuilds-per-flush directly from the saved log. This makes the ticket materially more reproducible and turns previously ad hoc terminal commands into tracked artifacts.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/scripts/01-create-ticket.sh — Reproduces the ticket creation step
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/scripts/08-capture-no-input-trace.sh — Reproduces the no-input hardware trace capture
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/scripts/09-analyze-trace-log.py — Recomputes rebuild/loop counts per non-empty flush from a saved trace log

## 2026-04-12

Captured the first real no-input hardware trace log for `LOUPE-009` and analyzed it with a ticket-local analysis script. The resulting evidence is much stronger than the earlier counter-only picture because it shows the event sequence directly. The scene produced `672` `scene.renderAll.begin` events and `671` `scene.loop.tick` events against `20` non-empty full-page flushes during the traced run. The derived ratio is about `33.6` rebuilds per non-empty flush on average, with a median of `27`, a minimum of `2`, and a maximum of `119`. The trace also shows that rebuilds continue to happen *while* long flushes are in progress: for example, one `go.flush.begin`/`go.flush.end ops=1` interval with `elapsedMs=2091.36` contained `118` loop ticks and `118` `renderAll.begin/end` pairs between those two Go-side events.

This is the clearest evidence so far that the problem is not just that loop rebuilds happen often in aggregate. The problem is that the full-page scene continues to generate rebuild work during long-running full-page flushes, so the scene is effectively outproducing the flush path. Based on this trace, cadence limiting should be the immediate next optimization, and deeper renderer/writer trace points are not the first thing to do next unless cadence limiting fails to move the rebuilds-per-flush ratio materially.

### Related Files

- /tmp/loupe-cyb-ito-full10-trace-1776025944.log — First no-input hardware trace log for the full-page all-12 scene
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/scripts/08-capture-no-input-trace.sh — Capture command used for the trace run
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/scripts/09-analyze-trace-log.py — Analysis script used to compute per-flush rebuild ratios
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-009--trace-driven-investigation-of-excessive-full-page-js-rebuild-calls-on-loupedeck/tasks.md — Marked Phase E complete and recorded that cadence limiting is now the recommended next optimization

