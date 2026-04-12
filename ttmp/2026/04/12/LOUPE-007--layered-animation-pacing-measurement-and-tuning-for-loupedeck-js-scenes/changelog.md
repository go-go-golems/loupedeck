# Changelog

## 2026-04-12

Created the `LOUPE-007` ticket as a separate documentation and planning track for layered-scene pacing analysis. The ticket captures how to measure whether layered retained JS scenes affect pacing, renderer cost, writer queue behavior, and user-visible responsiveness on real Loupedeck hardware without confusing those different concerns.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/index.md — Ticket overview and status entrypoint
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/tasks.md — Detailed future task breakdown for instrumentation, density sweeps, and interpretation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/design/01-textbook-measuring-layered-animation-density-pacing-and-tuning-for-loupedeck-js-scenes.md — Main intern-facing design and implementation guide
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/playbooks/01-layered-density-measurement-runbook.md — Operational benchmark/runbook for future hardware sweeps
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/reference/01-implementation-diary.md — Chronological continuity log for the ticket

## 2026-04-12

Validated the new pacing-analysis ticket with `docmgr doctor` and uploaded the ticket bundle to reMarkable. The uploaded bundle includes the ticket index, the main design/implementation guide, the operational runbook, and the implementation diary under the remote folder `/ai/2026/04/12/LOUPE-007`.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/index.md — Included in the uploaded bundle as the ticket overview
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/design/01-textbook-measuring-layered-animation-density-pacing-and-tuning-for-loupedeck-js-scenes.md — Included in the uploaded bundle as the main design guide
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/playbooks/01-layered-density-measurement-runbook.md — Included in the uploaded bundle as the operational runbook
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/reference/01-implementation-diary.md — Included in the uploaded bundle as the continuity log

## 2026-04-12

Added the first real implementation slice under `LOUPE-007`, even before stats logging: frame-atomic retained surface batching. This came directly from hardware observations while comparing tile-sized blits against the new full-page `360×270` redraw mode. The visible symptom was that later tiles in the full-page scene only appeared on some frames, which initially looked like generic slowness. The actual issue was more specific: the renderer could snapshot the shared retained `main` surface while JavaScript was still inside `renderAll()`, so the device sometimes received partially painted full-page frames. The fix was to add batching to `runtime/gfx/surface.go`, make reads wait for an in-flight batch to complete, expose `surface.batch(() => ...)` to JavaScript, and update `examples/js/10-cyb-ito-full-page-all12.js` to build a coherent frame inside one batch before the display is marked dirty.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface.go — Added retained surface batching, stable read/wait behavior, and coalesced notifications
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface_test.go — Added batching tests for notification coalescing and stable read behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go — Exposed JS `surface.batch(() => ...)`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Exercised the new JS batch API in the graphics module tests
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js — Updated the full-page scene to build each frame in one atomic retained-surface batch

## 2026-04-12

Implemented the first real instrumentation slice from `LOUPE-007`. `cmd/loupe-js-live` can now log periodic renderer stats, writer stats, and JS-scene metrics, and the JS runtime now exposes a `loupedeck/metrics` module so scenes can measure work from inside JavaScript itself. This was applied immediately to the new full-page all-12 scene so it can record `renderAll()` timing, loop ticks, activations, and per-tile draw timings rather than forcing us to guess where the time is going.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Added periodic render/writer/JS stats logging flags and aggregation helpers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics.go — Added a reusable in-process metrics collector for counters and timing windows
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics_test.go — Added collector coverage for snapshots and reset behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/env/env.go — Extended the JS environment with a metrics collector
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_metrics/module.go — Added the JS `loupedeck/metrics` module (`inc`, `observeMillis`, `time`, `now`)
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Registered the metrics module in the owned JS runtime
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go — Added coverage proving the JS metrics module records counters and timings
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js — Added JS-side timing/counter instrumentation for `renderAll()` and per-tile draw work
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/writer.go — Exposed current queue depth in `WriterStats` so periodic logs can show queue snapshots as well as cumulative totals

## 2026-04-12

Captured the first real evidence log with the new renderer/writer/JS instrumentation enabled on the batched full-page all-12 scene. The result was immediately useful because it falsified the simplest guess. JavaScript scene construction is expensive but not the main reason for the *extremely* slow visible updates. The JS-side metrics showed `scene.renderAll` averaging roughly `18–22 ms` per call with the `SPIRAL` tile as the hottest single tile at roughly `5–6 ms`, while the Go-side render window showed only one non-empty full-page flush per stats window and reported flush durations around `1.1–1.5 s`. At the same time, writer stats showed only one command sent in each window and a queue depth of zero. That means the writer queue is not backing up; instead, the full-page flush path is effectively stalling upstream while the scene keeps rebuilding frames. This is the first concrete data point showing that the next optimization should target frame availability / scene cadence rather than blindly blaming raw transport throughput.

### Related Files

- /tmp/loupe-cyb-ito-full10-stats-1776020694.log — First hardware evidence log with render/writer/JS stats enabled on the full-page all-12 scene
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Produced the first periodic stats output
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js — Produced the first scene-side timing breakdown (`renderAll`, per-tile timings, loop ticks)

## 2026-04-12

Refactored the underlying JS metrics implementation to be reusable beyond the current Loupedeck runtime. The low-level collector binding and native module registration now live in a generic package, `pkg/jsmetrics`, rather than hanging off the Loupedeck-specific JS environment. The current runtime now binds the collector into `runtimebridge` under a generic key and registers the modules with the `loupedeck` prefix (`loupedeck/metrics`, `loupedeck/scene-metrics`) as just one concrete flavor. This makes the instrumentation substrate much easier to move into `go-go-goja` later and reuse across unrelated JS runtimes while preserving the current module API for this repo.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/jsmetrics/jsmetrics.go — New reusable metrics binding and module-registration package
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime.go — Now binds the metrics collector generically via `runtimebridge` and registers JS metrics modules through `pkg/jsmetrics`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_metrics/module.go — Reduced to a thin compatibility wrapper around the generic package
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_scene_metrics/module.go — Reduced to a thin compatibility wrapper around the generic package
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/docs/help/topics/01-loupedeck-js-api-reference.md — Updated docs for the higher-level scene metrics helper package

## 2026-04-12

Added a full project technical report for the 12-tile cyb-ito performance investigation. This new document is intended as the durable onboarding/reference artifact for a future intern. It pulls together the architecture baseline, the tile-mode versus full-page-mode experiments, the frame-atomic batching fix, the combined render/writer/JS instrumentation results, the reasoning behind the current hypotheses, and the recommended next experiments. In other words, it turns the investigation from a sequence of diary entries into a coherent technical narrative about *why* the current runtime behaves the way it does and what would be most rational to try next.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/design/02-project-technical-report-performing-the-12-tile-javascript-canvas-cyb-ito-port.md — New long-form project technical report for the 12-tile cyb-ito performance problem
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/index.md — Updated key links and status summary to surface the new report
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/tasks.md — Updated analysis/design checklist and reset validation/upload tasks for the expanded bundle

## 2026-04-12

Reran ticket validation and uploaded a refreshed reMarkable bundle for `LOUPE-007` that now includes the new long-form technical report. The updated remote folder was verified successfully and now contains both the original pacing-analysis bundle and the new expanded technical-report bundle.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/tasks.md — Marked validation/upload/verification complete again after the new report addition
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/index.md — Updated status summary after successful validation and upload

## 2026-04-12

Reran the full-page all-12 scene on real hardware after adding the rebuild-reason metrics. This fresh evidence answered one of the main open questions directly: in a no-input run, the rebuild flood is being driven by the animation loop, not by hidden button/touch handlers or some mysterious extra invalidation source. The first stats window recorded one `initial` rebuild and then `295` loop-driven rebuilds; subsequent windows showed only `scene.renderAll.reason.loop` counts (`76`, `265`, `92`, `7`, `72`, `222`) with no touch/button categories at all. At the same time, the writer queue again stayed flat at zero while full-page render windows remained highly variable and often very large (`99 ms`, `825 ms`, `1.26 s`, `1.31 s`, `1.98 s`, `4.10 s`, `4.74 s`). This strengthens the current conclusion that the next optimization should target full-page rebuild cadence and/or frame-availability strategy, not queue tuning first.

One additional observation from this rerun is that the stats windows are not strict one-second wall-clock samples under heavy load. `cmd/loupe-js-live` services `statsTick` from the same `select` loop that performs `renderer.Flush()`, so a long flush delays stats emission. That means the raw counter values should be interpreted as "per delayed stats window" rather than naively as exact per-second rates. This does not weaken the main conclusion about *which* reason is dominant, but it does matter when interpreting absolute frequencies.

### Related Files

- /tmp/loupe-cyb-ito-full10-reasons-1776023397.log — Fresh hardware evidence log with rebuild-reason metrics enabled on the full-page all-12 scene
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js — Scene under test; its `anim.loop(... renderAll("loop"))` path is now confirmed as the dominant rebuild source in the no-input measurement
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Stats emission path whose shared `select` loop explains why heavy flushes widen the effective sampling window
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/tasks.md — Updated to mark the rebuild-reason follow-up measurement complete
