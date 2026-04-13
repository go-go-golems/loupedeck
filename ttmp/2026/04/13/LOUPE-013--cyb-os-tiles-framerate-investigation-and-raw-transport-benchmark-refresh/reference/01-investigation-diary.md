---
Title: Investigation Diary
Ticket: LOUPE-013
Status: active
Topics:
    - loupedeck
    - benchmarking
    - performance
    - serial
    - rendering
    - animation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupe-fps-bench/main.go
      Note: Raw writer/display benchmark command scoped as the control baseline for this ticket
    - Path: cmd/loupe-js-live/run.go
      Note: Hardware live-runner path used for smoke runs and stats-enabled scene measurement
    - Path: cmd/loupe-js-live/stats.go
      Note: Render and writer stats used as the primary pacing evidence
    - Path: examples/js/11-cyb-os-tiles.js
      Note: Scene under investigation; redraws left
    - Path: pkg/device/renderer.go
      Note: Default retained flush interval explains the live-runner's ~25 FPS ceiling
    - Path: ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/01-js-path-probe-main-only.js
      Note: Main-only live-runner probe used to compare against raw full-screen throughput
    - Path: ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/02-js-path-probe-three-display.js
      Note: Three-display live-runner probe used to compare command count effects
    - Path: ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/03-js-path-probe-main-fast-sides-slow.js
      Note: Decimated-side probe used to test lower side-display cadence
    - Path: ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/04-js-path-probe-main-only-present-driven.js
      Note: Self-invalidating main-only probe used to separate presenter cadence from anim.loop cadence
    - Path: ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/05-js-path-probe-three-display-present-driven.js
      Note: Self-invalidating three-display probe used to confirm the same scheduler cap under presenter-driven invalidation
ExternalSources: []
Summary: Backfilled diary for the first cyb-os-tiles hardware smoke run, framerate investigation, main-only A/B probe, and raw benchmark ticket setup.
LastUpdated: 2026-04-13T15:32:47.878089581-04:00
WhatFor: Preserve the first evidence-backed investigation into cyb-os-tiles pacing, especially the difference between three-display redraws and main-only redraws.
WhenToUse: Use when resuming the cyb-os-tiles framerate investigation or when comparing retained-scene pacing against raw writer benchmark ceilings.
---



# Investigation Diary

## Goal

Capture the first hardware-backed investigation into why `examples/js/11-cyb-os-tiles.js` feels slower than expected on the Loupedeck Live, and preserve the evidence that separating main-display redraws from side-display redraws materially improves observed flush cadence.

## Step 1: Run `cyb-os-tiles` on hardware as a first smoke test

The first question was simply whether the `cyb-os-tiles` scene ran at all on the current tree. That mattered because we had never actually run it on hardware in this cleanup/refactor phase, so a failure would not necessarily indicate a regression. The right first move was to run the existing scene on the real device with the same aggressive writer setting we had used for other live-scene checks.

The smoke test succeeded cleanly. That was important because it established a baseline: the scene connects, draws to all expected displays, and exits on timeout. Once that was known, later complaints about visible slowness could be treated as a pacing/performance question rather than a binary “scene is broken” question.

### Prompt Context

**User prompt (verbatim):** "run cyb-os-tiles (we never ran it, so it failing prob isn't a refactor issue)"

**Assistant interpretation:** Run the `cyb-os-tiles` scene on actual hardware and verify whether it works at all before treating any result as a refactor regression.

**Inferred user intent:** Establish a clean hardware baseline for a scene that had not yet been exercised in this refactor window.

**Commit (code):** N/A

### What I did
- Located the scene at:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js`
- Ran the live runner on hardware with the scene and aggressive writer pacing:

```bash
timeout 30s go run ./cmd/loupe-js-live \
  --script ./examples/js/11-cyb-os-tiles.js \
  --duration 5s \
  --send-interval 0ms
```

- Captured the log at:
  - `/tmp/loupe-cyb-os-tiles-1776108485.log`

### Why
- A first-run smoke test avoids misclassifying “never validated before” as “newly broken.”
- It also provides a known-good hardware reference before deeper pacing analysis starts.

### What worked
- The scene connected and drew repeatedly.
- The scene visibly exercised all three retained displays:
  - `left`
  - `main`
  - `right`
- The run exited cleanly on timeout with status `0`.
- The log contained a clean timeout exit line:

```text
Loupedeck JS live runner exiting reason=timeout script=./examples/js/11-cyb-os-tiles.js duration=5s
```

### What didn't work
- N/A for the smoke-test goal. The scene did run.

### What I learned
- `cyb-os-tiles` is operational on the current tree, so later concerns are about smoothness and throughput, not a straightforward regression in scene loading.
- The scene uses all three retained displays, which immediately makes it a different throughput profile than main-only scenes.

### What was tricky to build
- The tricky part was interpretive rather than technical: because the scene had not been previously validated in this branch, a failure would have had ambiguous meaning. Running it first as a simple smoke test removed that ambiguity.

### What warrants a second pair of eyes
- Review whether the same scene should also get a tiny dedicated hardware smoke script under the ticket so later runs are consistent and easy to repeat.

### What should be done in the future
- Follow with stats-enabled runs to understand why the scene feels slow despite functioning correctly.

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js`
- Validate with:

```bash
timeout 30s go run ./cmd/loupe-js-live \
  --script ./examples/js/11-cyb-os-tiles.js \
  --duration 5s \
  --send-interval 0ms
```

### Technical details
- Evidence log:
  - `/tmp/loupe-cyb-os-tiles-1776108485.log`

---

## Step 2: Measure the scene with render/writer stats and explain the slow feel

Once the smoke run proved the scene was healthy, the next question shifted from correctness to pacing: why did it look slow? Rather than guessing, I reran the scene with the live runner’s render and writer stats enabled. That made it possible to compare visible smoothness against actual flush cadence and render time.

The stats showed that the scene was effectively running around the low-teens in steady-state flush cadence while redrawing three full displays every non-empty flush. That result strongly suggested the main bottleneck was not pure JavaScript compute alone. The scene does substantial software work, but the transport/display path is paying for three device-facing draws per frame.

### Prompt Context

**User prompt (verbatim):** "why do you think the framerate seems pretty low?"

**Assistant interpretation:** Use evidence rather than intuition to explain which layer is likely limiting `cyb-os-tiles` on hardware.

**Inferred user intent:** Distinguish between JS/render cost and device transport cost before making optimization decisions.

**Commit (code):** N/A

### What I did
- Read the scene source to confirm its structure and display usage:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js`
- Reran the scene with stats enabled:

```bash
timeout 20s go run ./cmd/loupe-js-live \
  --script ./examples/js/11-cyb-os-tiles.js \
  --duration 8s \
  --send-interval 0ms \
  --log-render-stats \
  --log-writer-stats \
  --stats-interval 2s
```

- Interpreted the reported windows, especially:

```text
render stats ... flush_ticks=31 non_empty_flushes=31 displays=93 tiles=0 ops=93 avg_render_ms=4.08 max_render_ms=7.28
writer stats ... QueuedCommands:96 SentCommands:96 SentMessages:192 FailedCommands:0

render stats ... flush_ticks=21 non_empty_flushes=21 displays=63 tiles=0 ops=63 avg_render_ms=5.98 max_render_ms=8.98
writer stats ... QueuedCommands:63 SentCommands:63 SentMessages:126 FailedCommands:0

render stats ... flush_ticks=22 non_empty_flushes=22 displays=66 tiles=0 ops=66 avg_render_ms=6.31 max_render_ms=9.18
writer stats ... QueuedCommands:64 SentCommands:64 SentMessages:128 FailedCommands:0
```

### Why
- If the render times were huge, the likely culprit would have been JS/scene generation.
- If render times were moderate but flush cadence stayed low while issuing multiple draws per flush, that would point toward transport/display overhead.

### What worked
- The stats were coherent enough to support a real explanation.
- The scene was clearly redrawing all three displays on each non-empty flush.
- The stats windows imply roughly:
  - `31 / 2s ≈ 15.5 FPS`
  - `21 / 2s ≈ 10.5 FPS`
  - `22 / 2s ≈ 11 FPS`
- The render path itself was not catastrophically slow; `avg_render_ms` stayed around `4–6 ms`.

### What didn't work
- This did not by itself isolate whether the main pain was per-command overhead or pure pixel payload size. It only showed that the three-display full redraw pattern correlated strongly with the observed slowness.

### What I learned
- `cyb-os-tiles` is using `tiles=0` and `displays=...`, i.e. full retained-display redraws rather than fine-grained tile updates in this mode.
- The visible framerate is in the low-teens even with `--send-interval 0ms`.
- The scene is expensive mainly because it redraws `left`, `main`, and `right` every frame, not just because the tile art is procedurally dense.

### What was tricky to build
- The tricky part was interpreting stats windows honestly. The scene continues drawing while the live runner periodically emits windows, so the clearest signal came from comparing flush counts, display op counts, and render timing together rather than reading any one counter in isolation.

### What warrants a second pair of eyes
- Review whether the stats windows should be normalized further when used in human-facing reports, especially when comparing scenes with different flush scheduling behavior.

### What should be done in the future
- Run a controlled A/B where the exact same scene draws only the main display, to see whether the draw count alone materially changes flush cadence.
- Run the raw benchmark command as a control baseline.

### Code review instructions
- Read:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/stats.go`
- Reproduce with:

```bash
timeout 20s go run ./cmd/loupe-js-live \
  --script ./examples/js/11-cyb-os-tiles.js \
  --duration 8s \
  --send-interval 0ms \
  --log-render-stats \
  --log-writer-stats \
  --stats-interval 2s
```

### Technical details
- The key finding here was approximately `10–15 FPS` with three full-display draws per non-empty flush.

---

## Step 3: Run a main-only A/B probe to test whether draw count is the problem

The next question was a very good one: if the scene only drew `main`, would it speed up enough to implicate the draw count directly? That was worth testing before changing the scene permanently, so I created a temporary probe variant that removed `left` and `right` from the page and rendered only the main surface.

The A/B result was strong. Main-only redraws roughly doubled the observed flush cadence into the low-20s. That does not prove the bottleneck is *only* command count, but it does show that the three-display pattern is a major contributor. It also suggests that per-draw protocol overhead matters, not just total pixel area.

### Prompt Context

**User prompt (verbatim):** "can we draw only the main display to see if it's the count of draws?"

**Assistant interpretation:** Keep the scene otherwise similar, remove side-display draws, and compare the resulting throughput to the full three-display version.

**Inferred user intent:** Use a controlled experiment to determine whether the slowdown is dominated by the number of device-facing draw operations.

**Commit (code):** N/A

### What I did
- Inspected the scene and confirmed it currently does:
  - `renderMain()`
  - `renderLeft()`
  - `renderRight()`
  - and registers `left`, `main`, `right` displays on the page.
- First attempted a temporary probe script in `/tmp`, but the file was not visible to the later run command and failed with:

```text
read script: open /tmp/cyb-os-tiles-main-only.js: no such file or directory
exit status 1
```

- Recreated the probe variant inside the repo at:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/cyb-os-tiles-main-only.js`
- The temporary variant changed only the page/display wiring and `renderAll()` so it rendered `main` only.
- Ran the A/B probe with the same stats flags:

```bash
timeout 20s go run ./cmd/loupe-js-live \
  --script ./ttmp/cyb-os-tiles-main-only.js \
  --duration 8s \
  --send-interval 0ms \
  --log-render-stats \
  --log-writer-stats \
  --stats-interval 2s
```

- Removed the temporary probe file afterward so the repo stayed clean.

### Why
- This was the simplest controlled experiment available without editing the real scene permanently.
- It directly tested whether the number of display draws per flush was materially affecting the effective framerate.

### What worked
- The main-only probe clearly ran faster than the three-display version.
- Representative stats windows were:

```text
render stats ... flush_ticks=48 non_empty_flushes=48 displays=48 tiles=0 ops=48 avg_render_ms=2.91 max_render_ms=5.86
render stats ... flush_ticks=48 non_empty_flushes=48 displays=48 tiles=0 ops=48 avg_render_ms=2.64 max_render_ms=4.15
render stats ... flush_ticks=44 non_empty_flushes=44 displays=44 tiles=0 ops=44 avg_render_ms=2.70 max_render_ms=4.45
```

- Those windows imply roughly:
  - `48 / 2s ≈ 24 FPS`
  - `44 / 2s ≈ 22 FPS`
- Compared to the earlier `10–15 FPS`, the main-only version is substantially faster.
- Render time also dropped from roughly `4–6 ms` to roughly `2.6–2.9 ms` in the strongest windows.

### What didn't work
- The first `/tmp` probe attempt failed because the later execution context could not see the temp file.
- The main-only test does not yet distinguish perfectly between:
  - per-command overhead
  - extra pixel payload from side displays
  - side-display scene-generation cost
  It only proves that removing the two extra display draws helps a lot.

### What I learned
- The number of display draws per frame is a major factor in the perceived slowness.
- The gain is large enough that per-draw command/ack overhead is likely significant, not just raw pixel count.
- A very plausible next optimization is to keep `main` at a high cadence while redrawing `left` and `right` less frequently.

### What was tricky to build
- The main sharp edge was making the experiment temporary and low-risk. The easiest safe method was to generate a throwaway scene variant, run it, and delete it immediately after. The failed `/tmp` attempt also reminded me that ephemeral paths are not always shared the way I expect across separate tool invocations.

### What warrants a second pair of eyes
- Review whether a side-display decimation strategy (for example every 3rd or 4th frame) would be enough to recover most of the main-only improvement without visibly harming the overall scene.

### What should be done in the future
- Run a second controlled variant where `main` updates every frame but `left`/`right` update at a lower cadence.
- Compare that result against the raw benchmark command to separate scene-level overhead from writer/display ceilings.

### Code review instructions
- Read:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js`
- Recreate the probe by removing side-display registration and making `renderAll()` call `renderMain()` only.
- Validate with:

```bash
timeout 20s go run ./cmd/loupe-js-live \
  --script ./ttmp/cyb-os-tiles-main-only.js \
  --duration 8s \
  --send-interval 0ms \
  --log-render-stats \
  --log-writer-stats \
  --stats-interval 2s
```

### Technical details
- Approximate comparison from the first A/B:
  - full three-display version: `~10–15 FPS`
  - main-only version: `~22–24 FPS`

---

## Step 4: Scope a raw benchmark follow-up and create a dedicated investigation ticket

After the A/B result, it was clear that this line of work deserved its own ticket instead of being left as loose chat context. The next step was to inspect the existing raw benchmark harness, verify what it actually measures, and create a dedicated workspace to continue the investigation in a structured way.

Reading `cmd/loupe-fps-bench/main.go` confirmed that it is the right control baseline for this question, with one important caveat: it is not “pure serial only” in the abstract. It measures the raw writer/display path with the render scheduler disabled and writer interval forced to `0`, which is much closer to transport/display ceiling than the retained JS scene, but it still includes the real device command/ack path and display work.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket for all that btw, and backfill the diary."

**Assistant interpretation:** Open a dedicated ticket for the framerate and transport investigation and record the already completed measurements so the work is not trapped in transient conversation state.

**Inferred user intent:** Turn the exploratory measurements into a durable, searchable, reviewable workstream with preserved evidence and next steps.

**Commit (code):** N/A

### What I did
- Inspected the existing benchmark command:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go`
- Confirmed it prints this mode summary:

```text
Mode: raw writer benchmark (render scheduler disabled, writer interval = 0)
```

- Created a new ticket:

```bash
docmgr ticket create-ticket \
  --ticket LOUPE-013 \
  --title "cyb-os-tiles framerate investigation and raw transport benchmark refresh" \
  --topics loupedeck,benchmarking,performance,serial,rendering,animation
```

- Added this diary document:

```bash
docmgr doc add --ticket LOUPE-013 --doc-type reference --title "Investigation Diary"
```

- Backfilled the diary with the earlier hardware smoke run, stats-based analysis, and main-only A/B results.

### Why
- The measurements already collected were strong enough to deserve persistence.
- A dedicated ticket makes it easier to add follow-up benchmark runs, scripts, design notes, and decisions without mixing them into unrelated tickets.

### What worked
- The new ticket workspace was created at:
  - `ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh`
- The backfilled diary now preserves the evidence trail that led to the current hypothesis: three-display redraw count is a major factor in the slow feel.
- The benchmark command is clearly positioned as the next control baseline.

### What didn't work
- The raw benchmark itself has not yet been rerun as part of this new ticket. At this point the ticket captures the setup and the prior scene-level evidence, but not the fresh benchmark numbers.

### What I learned
- `cmd/loupe-fps-bench` is a much better control than guessing from scene logs, but it should still be described carefully: it measures the real writer/display path with the renderer disabled, not an abstract serial bandwidth number divorced from protocol overhead.

### What was tricky to build
- The main challenge was choosing the right scope for the ticket. The issue is not only “cyb-os-tiles is slow,” and it is not only “serial is slow.” The evidence points to an interaction between scene structure, number of display updates, and writer/display overhead, so the ticket title and diary needed to cover all three without becoming vague.

### What warrants a second pair of eyes
- Review whether the ticket should stay narrowly focused on `cyb-os-tiles` pacing or whether it should grow into a broader display-update cadence / raw-writer comparison ticket.

### What should be done in the future
- Run fresh `loupe-fps-bench` control sweeps on the current tree.
- Add a lower-cadence side-display experiment for `cyb-os-tiles`.
- Compare raw benchmark ceilings against retained-scene flush rates in one concise report.

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js`
  - this diary
- Inspect the ticket workspace at:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh`

### Technical details
- Benchmark command orientation note from source:

```text
Loupedeck Live FPS benchmark
Mode: raw writer benchmark (render scheduler disabled, writer interval = 0)
Display measured as full touchscreen main display = 360x270 for product 0004
```

---

## Step 5: Rerun the raw hardware benchmark successfully on the current tree

With the ticket in place, the first real follow-up measurement was to rerun the dedicated raw benchmark harness on the current codebase. This mattered because the earlier `cyb-os-tiles` observations were scene-level and live-runner-level observations; they were not a clean control for what the underlying device path can still do right now.

The benchmark rerun completed successfully on the second attempt. The first streamed run aborted mid-output while the mixed sweep was still in progress, so I reran it quietly to a log file and then extracted the summary lines. The successful rerun confirmed that the raw writer/display path is still substantially stronger than the `cyb-os-tiles` live-scene results would suggest.

### Prompt Context

**User prompt (verbatim):** "ok, go ahead. You can also write your own JS based fps measurement scripts, to test the different paths"

**Assistant interpretation:** Proceed with the fresh benchmark work and, if useful, add dedicated measurement probes rather than relying only on the existing scene.

**Inferred user intent:** Turn the ticket from a hypothesis into evidence by collecting both raw baseline numbers and targeted live-path measurements.

**Commit (code):** N/A

### What I did
- Tried an initial streamed benchmark run:

```bash
go run ./cmd/loupe-fps-bench | tee /tmp/loupe-fps-bench-1776109234.log
```

- That run aborted mid mixed-sweep before producing a final complete summary.
- Reran the benchmark quietly to a file and extracted the summary lines:

```bash
timeout 360s go run ./cmd/loupe-fps-bench >/tmp/loupe-fps-bench-rerun-1776109486.log 2>&1
rg -n "^== |^summary:|^target=|^scale=" /tmp/loupe-fps-bench-rerun-1776109486.log | tail -120
```

- Recorded the successful raw-writer results:
  - full-screen main display stable to `36 FPS`, peak achieved `37.38 FPS`
  - single `90x90` button region stable to `320 FPS`, peak achieved `313.66 FPS`
  - mixed 12-button animation stable to total target `288 FPS`

### Why
- The ticket needed a current-tree control baseline, not only prior historical numbers.
- The raw benchmark disables the retained render scheduler and uses writer interval `0`, so it is the right baseline for what the device path can sustain absent the live runner’s normal coalescing path.

### What worked
- The benchmark completed successfully on the rerun with exit status `0`.
- The successful summary from `/tmp/loupe-fps-bench-rerun-1776109486.log` showed:

```text
summary: max stable target fps=36.0, peak achieved fps=37.38
summary: max stable target fps=320.0, peak achieved fps=313.66
summary: best stable mixed-framerate total target fps=288.00 total actual fps=288.00
```

- This is materially faster than the earlier `cyb-os-tiles` live-scene cadence, especially for full-screen main-display updates.

### What didn't work
- The first streamed benchmark run aborted before completion while the mixed sweep was still printing results.
- That did not invalidate the benchmark harness, but it did reinforce that long interactive runs are safer when redirected to a log file and summarized after the fact.

### What I learned
- The raw writer/display path on the current tree still supports roughly:
  - `36 FPS` full-screen main-display updates
  - `320 FPS` for a single `90x90` region
  - `288 total FPS` for the best stable mixed 12-button workload
- That means `cyb-os-tiles` feeling slow is not simply “the hardware can only do ~10 FPS.” The retained live-runner path and the scene structure matter a lot.

### What was tricky to build
- The main sharp edge was the interactive output path. The benchmark itself is long-running and verbose enough that streaming it directly was less robust than logging to a file and post-processing the summary lines.

### What warrants a second pair of eyes
- Review whether `cmd/loupe-fps-bench` should grow flags for narrower reruns so future comparisons do not require the full end-to-end sweep every time.

### What should be done in the future
- Compare these raw baseline numbers directly against purpose-built JS path probes under `cmd/loupe-js-live`.
- Keep distinguishing raw writer/display ceilings from retained-scene ceilings.

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go`
- Reproduce with:

```bash
timeout 360s go run ./cmd/loupe-fps-bench >/tmp/loupe-fps-bench-rerun.log 2>&1
rg -n "^== |^summary:|^target=|^scale=" /tmp/loupe-fps-bench-rerun.log | tail -120
```

### Technical details
- Successful evidence log:
  - `/tmp/loupe-fps-bench-rerun-1776109486.log`
- Key stable ceilings from that run:
  - full-screen main display: `36 FPS`
  - single `90x90` region: `320 FPS`
  - mixed 12-button bank: `288 total FPS`

---

## Step 6: Add dedicated JS path probes and uncover the live runner's default 40ms cap

After confirming the raw baseline, I added a set of dedicated JS probe scenes to test the live-runner path more directly. The goal was to separate questions that the big `cyb-os-tiles` scene confounds together: one main display versus three displays, full-speed sides versus decimated side redraws, and animation-loop pacing versus presenter-driven self-invalidation.

These probes turned out to reveal the most important ticket insight so far. The live-runner path is not operating at the raw benchmark ceiling by default. `cmd/loupe-js-live` goes through the retained render scheduler, and the scheduler’s default flush interval is `40ms`, which implies a built-in cap of about `25 FPS`. The JS probes matched that exactly at the writer layer: the main-only probe sent about `50` commands per `2s` window, and the three-display probe sent about `150` commands per `2s` window, which is still only about `25` frame-equivalents per second.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Use custom JS probes if they help isolate which part of the live path is limiting throughput.

**Inferred user intent:** Test different display-update shapes directly in the JS runtime rather than inferring everything from one complex demo scene.

**Commit (code):** `bc408ec` — "Add LOUPE-013 JS path probe scripts"

### What I did
- Added five probe scripts under:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/`
- The probe set includes:
  - `01-js-path-probe-main-only.js`
  - `02-js-path-probe-three-display.js`
  - `03-js-path-probe-main-fast-sides-slow.js`
  - `04-js-path-probe-main-only-present-driven.js`
  - `05-js-path-probe-three-display-present-driven.js`
- Ran the animation-loop-driven probes with:

```bash
timeout 25s go run ./cmd/loupe-js-live --script <probe.js> --duration 8s --send-interval 0ms --log-render-stats --log-writer-stats --stats-interval 2s
```

- Ran the presenter-driven probes with:

```bash
timeout 20s go run ./cmd/loupe-js-live --script <probe.js> --duration 6s --send-interval 0ms --log-render-stats --log-writer-stats --stats-interval 2s
```

- Inspected the live-runner render scheduler defaults in:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/renderer.go`
- Confirmed the default retained flush interval is:

```go
var DefaultRenderOptions = RenderOptions{
    FlushInterval: 40 * time.Millisecond,
    QueueSize:     256,
}
```

### Why
- The big scene alone could not distinguish “device is slow” from “live runner is coalescing at a fixed cadence” from “scene-generation work is heavy.”
- Purpose-built probes made those layers much easier to separate.

### What worked
- The scripts ran successfully on hardware and produced stable per-window writer statistics.
- The key writer-level results were strikingly consistent:
  - main-only probe: about `50` commands per `2s` window
  - three-display probe: about `150` commands per `2s` window
  - decimated-side probe: about `112–113` commands per `2s` window
- Interpreted as frame-equivalents, that is approximately:
  - main-only: `25 FPS`
  - three-display: `150 / 3 / 2s = 25 FPS`
  - main-fast/sides-slow: about `28 commands/s`, or about `25` main frames/s with side updates layered on top
- The presenter-driven self-invalidating probes produced the same writer-layer outcome as the animation-loop-driven probes, which strongly suggests the default retained scheduler cadence, not the JS invalidation source, is the main cap.

### What didn't work
- The render stats themselves can look much larger than the writer command rates because they count retained flush activity/coalescing behavior rather than literal device commands. That makes them easy to misread if viewed without the writer stats.
- The probe scripts do not yet answer how much of `cyb-os-tiles` remaining slowdown below `25 FPS` comes from scene-generation work versus the exact shape of its redraws.

### What I learned
- The raw hardware path and the live-runner default path are different ceilings.
- `cmd/loupe-js-live` currently uses the retained scheduler default of `40ms`, which is about `25 FPS` maximum flush cadence.
- The earlier `cyb-os-tiles` main-only A/B result now makes more sense: the main-only version improved to roughly `22–24 FPS`, which is very close to the live-runner default cap.
- The full three-display `cyb-os-tiles` run being down around `10–15 FPS` means there is additional scene-level cost on top of the scheduler cap.

### What was tricky to build
- The tricky part was interpreting the mixed evidence correctly. At first glance, the large `flush_ticks` values in render stats could make it look like the live runner was rendering far faster than the writer was sending. Reading `pkg/device/renderer.go` clarified the architectural reason: the renderer is a coalescing scheduler with a fixed default flush interval, and writer stats are the more honest source for effective device-facing throughput.

### What warrants a second pair of eyes
- Review whether `cmd/loupe-js-live` should expose render flush interval as a CLI flag for measurement and tuning.
- Review whether future scene-performance discussions should always report both raw benchmark numbers and live-runner scheduler settings together.

### What should be done in the future
- Add a real `cyb-os-tiles` variant with reduced side-display cadence to see how close it can get to the `25 FPS` live-runner ceiling.
- Decide whether to expose scheduler tuning in the live runner or keep the current `40ms` default as a user-experience tradeoff.

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/renderer.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/01-js-path-probe-main-only.js`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/02-js-path-probe-three-display.js`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/03-js-path-probe-main-fast-sides-slow.js`
- Reproduce the main finding with either the animation-loop or presenter-driven probe pair and compare writer stats.

### Technical details
- Example writer windows from the probes:

```text
main-only probe: delta={QueuedCommands:50 SentCommands:50 SentMessages:100 FailedCommands:0 ...}
three-display probe: delta={QueuedCommands:150 SentCommands:150 SentMessages:300 FailedCommands:0 ...}
main-fast/sides-slow probe: delta={QueuedCommands:113 SentCommands:113 SentMessages:226 FailedCommands:0 ...}
```

- These are consistent with a retained flush cadence near `25 FPS` when `FlushInterval` is `40ms`.
