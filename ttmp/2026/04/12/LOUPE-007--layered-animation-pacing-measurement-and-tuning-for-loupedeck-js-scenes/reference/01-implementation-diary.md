---
Title: Implementation diary
Ticket: LOUPE-007
Status: active
Topics:
    - loupedeck
    - benchmarking
    - performance
    - rendering
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/design/01-textbook-measuring-layered-animation-density-pacing-and-tuning-for-loupedeck-js-scenes.md
      Note: Main intern-facing design and implementation guide for the pacing-analysis ticket
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/playbooks/01-layered-density-measurement-runbook.md
      Note: Operational runbook for future density-sweep measurements
ExternalSources: []
Summary: Chronological diary for creating the layered-scene pacing analysis ticket and its design package.
LastUpdated: 2026-04-12T07:00:00-04:00
WhatFor: Preserve why this ticket exists, what it is intended to measure, and how the design package was assembled.
WhenToUse: Use when resuming LOUPE-007 later or when reviewing why the pacing-analysis work was split into its own ticket.
---

# Diary

## Goal

Create a separate ticket that captures the analysis, design, and future implementation plan for measuring whether layered retained JS scenes affect pacing, responsiveness, and transport pressure on real Loupedeck hardware.

## Step 1: Create the separate pacing-analysis ticket and write the design package

The user wanted the pacing-analysis ideas captured in a dedicated ticket so that they would not get lost while the cyb-ito implementation work continued. That was the right call. The ongoing `LOUPE-006` scene work is implementation-heavy, while this new work is primarily about measurement strategy, interpretation, and later instrumentation. Splitting them keeps the future performance work coherent and searchable.

The core design task here was to turn an informal conversation about possible slowdown causes into a very explicit intern-facing guide. The guide needed to explain that "slow" can mean several different things in this repository and that the measurement plan must separate scene-update cadence, Go-side composition cadence, writer queue behavior, and actual device-visible responsiveness.

### Prompt Context

**User prompt (verbatim):** "Ok, create a separate ticket with a detailed analysis / implementatino plan explaining and capturing all the ideas above, so that we can work on it later. 

reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new ticket for layered-scene pacing analysis, write a detailed intern-facing design and implementation guide plus supporting ticket docs, and upload the resulting bundle to reMarkable.

**Inferred user intent:** Preserve the pacing-measurement strategy as a durable future work package without derailing the current cyb-ito implementation stream.

### What I did
- Created the new ticket workspace:

```bash
docmgr ticket create-ticket \
  --ticket LOUPE-007 \
  --title "Layered animation pacing measurement and tuning for Loupedeck JS scenes" \
  --topics loupedeck,goja,javascript,animation,rendering,benchmarking,performance
```

- Wrote the main design doc:

```text
design/01-textbook-measuring-layered-animation-density-pacing-and-tuning-for-loupedeck-js-scenes.md
```

- Wrote the operational runbook:

```text
playbooks/01-layered-density-measurement-runbook.md
```

- Wrote this diary entry for continuity.

### Why
- The pacing-analysis work deserves its own ticket because it is a future measurement/tuning stream rather than direct scene implementation.
- A dedicated design package makes it easier to resume later without reconstructing the reasoning from chat history.
- The current runtime is now mature enough that instrumentation work should be thoughtful and layered, not improvised.

### What worked
- The new ticket cleanly separates measurement design from active scene-authoring work.
- The design doc now captures the main conceptual split between scene cadence, renderer cadence, writer cadence, and device-visible cadence.
- The runbook turns the design into a concrete future execution sequence.

### What didn't work
- No technical tooling failed in this step.
- The work is intentionally documentation-first; it does not yet add the actual stats instrumentation.

### What I learned
- The conversation had already matured enough that it naturally mapped to a real ticket rather than another note inside LOUPE-006.
- The most important teaching point for a future intern is not any one metric; it is the need to avoid conflating several different clocks in the system.

### What was tricky to build
- The main tricky part was preserving the architecture boundaries clearly. It would have been easy to write a vague performance guide that treats the runtime as one black box, but that would be misleading in this codebase because scene work, retained composition, writer behavior, and hardware transport are deliberately separated.
- Another tricky part was making the future runbook concrete enough to be executable later without pretending the instrumentation already exists today.

### What warrants a second pair of eyes
- The proposed stats placement in `cmd/loupe-js-live/main.go` versus deeper reusable APIs may deserve review once implementation starts.
- The proposed scene modes for density sweeps should be checked against the evolving cyb-ito prototype so the benchmark modes stay meaningfully comparable.

### What should be done in the future
- Finish the ticket bookkeeping and validate the ticket with `docmgr doctor`.
- Upload the bundle to reMarkable.
- Then resume the active cyb-ito implementation work in LOUPE-006.

## Step 2: Validate the ticket and upload the bundle to reMarkable

Once the docs existed, the remaining work was operational rather than conceptual: validate the ticket metadata, add any missing vocabulary, upload the bundle to reMarkable, and verify the remote result.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the documentation workflow properly by validating the new ticket and delivering the bundle to reMarkable.

**Inferred user intent:** Make the new pacing-analysis ticket durable and reviewable now so it can be resumed later without depending on chat history.

### What I did
- Ran:

```bash
docmgr doctor --ticket LOUPE-007 --stale-after 30
```

- `docmgr doctor` initially warned that the new topics `benchmarking` and `performance` were not yet in vocabulary.
- Added them with:

```bash
docmgr vocab add --category topics --slug benchmarking --description "Benchmarking and measurement workflows"
docmgr vocab add --category topics --slug performance --description "Performance analysis and tuning work"
```

- Re-ran `docmgr doctor` and got a clean pass.
- Confirmed reMarkable tooling and uploaded the bundle with:

```bash
remarquee upload bundle <index + design + playbook + diary> --name "LOUPE-007 Layered animation pacing measurement and tuning" --remote-dir "/ai/2026/04/12/LOUPE-007" --toc-depth 2
```

- Verified the remote folder with:

```bash
remarquee cloud ls "/ai/2026/04/12/LOUPE-007" --long --non-interactive
```

### Why
- The ticket should be complete and discoverable now, not half-finished in the repo.
- Adding the vocabulary keeps future docmgr validation noise low and makes the topic tags reusable.

### What worked
- `docmgr doctor` passed after the vocabulary update.
- The bundle uploaded successfully.
- The remote folder now contains the expected document.

### What didn't work
- The only issue was the expected vocabulary warning for the new topics, which was resolved immediately.

### What I learned
- The new pacing-analysis topic names are useful enough to justify keeping them in the shared vocabulary.
- The LOUPE-007 bundle is now on the same footing as the earlier LOUPE ticket bundles and can be handed to a future reader independently.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- The naming of the ticket and bundle is probably fine, but future readers should confirm it remains clearly distinct from the active cyb-ito implementation ticket.

### What should be done in the future
- Resume the active cyb-ito implementation stream in `LOUPE-006`.

### Code review instructions
- Start with the main design doc in `design/`.
- Then read the operational runbook in `playbooks/`.
- Confirm the ticket framing in `index.md`, `tasks.md`, and `changelog.md`.
- Validate with:

```bash
docmgr doctor --ticket LOUPE-007 --stale-after 30
```

### Technical details
- The design doc references these current code anchors:
  - `cmd/loupe-js-live/main.go`
  - `runtime/render/visual_runtime.go`
  - `runtime/ui/display.go`
  - `writer.go`
  - `cmd/loupe-fps-bench/main.go`
  - `examples/js/07-cyb-ito-prototype.js`

## Step 3: Add the first concrete implementation slice before stats instrumentation — frame-atomic surface batching

Later, during hands-on comparison between the tile-subimage path and a new full-page `360×270` redraw path, the user reported an important symptom: in the full-page mode, tiles later in the draw order only appeared correctly on some frames. That was a very useful observation because it sounded like generic slowness at first, but it actually pointed at a more specific precondition for any meaningful pacing analysis: the runtime was allowing the renderer to snapshot a retained surface while JavaScript was still mutating it.

This meant `LOUPE-007` needed one small implementation slice before the broader stats work. Measuring renderer timing and writer pacing would be much less informative if the full-page scene could not even produce coherent frames yet.

### Prompt Context

**User prompt (verbatim):** "ok, add implementation plan and then implement"

**Assistant interpretation:** Record a concrete implementation slice for the newly diagnosed full-page frame-atomicity issue inside the pacing ticket, then implement the necessary runtime changes.

**Inferred user intent:** Do not just describe the problem abstractly; capture it as part of the actual performance/tuning plan and then fix the immediate runtime issue.

### What I did
- Updated `LOUPE-007` tasks to add a new Phase 0 for frame-atomic retained surface groundwork.
- Updated the ticket index/changelog to reflect that a first runtime precondition slice now exists even though the live-runner stats flags are still future work.
- Added batching support to `runtime/gfx/surface.go` so a surface can accumulate many mutations and emit one change notification at the end.
- Made stable-read paths such as `ToRGBA(...)` wait for an in-flight batch to complete before snapshotting the surface.
- Exposed the batching API to JavaScript as:

```javascript
surface.batch(() => {
  // many mutations
});
```

- Updated the new full-page all-12 example to wrap `renderAll()` in one batch so the shared `main` surface becomes coherent before the display is marked dirty.
- Added tests for:
  - coalesced change notification count,
  - stable read behavior during a batch,
  - JS module access to the new batch API.

### Why
- Without frame-atomic scene construction, the full-page mode can look worse for the wrong reason: the hardware is not merely receiving slower frames, it is receiving partially painted frames.
- Fixing that first makes later pacing measurements and renderer statistics much more meaningful.
- The batch API is also a good long-term scene-authoring tool, not just a one-off fix for this example.

### What worked
- The retained graphics layer now has a concrete concept of batched mutation.
- The JS API can express that concept directly without exposing transport details.
- The full-page scene can now be updated as one coherent retained-surface build rather than as a long stream of observable intermediate states.

### What didn't work
- This slice does not yet add the broader measurement flags from the original LOUPE-007 plan (`--log-render-stats`, `--log-writer-stats`, `--stats-interval`, etc.).
- Hardware comparison still needs to be rerun and recorded after the batching change.

### What I learned
- Before measuring pacing, we also need to think about *frame atomicity*. A retained scene can be logically correct and still look wrong if the renderer can observe its intermediate construction states.
- The line between `LOUPE-006` and `LOUPE-007` is still healthy here: the scene symptom was discovered during active cyb-ito implementation, but the architectural explanation and the general fix belong naturally in the pacing ticket.

### What was tricky to build
- The subtle part was recognizing that the problem was not simply "JavaScript is too slow". The more specific issue was that the renderer and the JS scene authoring path were not synchronized around the lifetime of one coherent retained surface frame.
- The Go-side batching implementation also had to preserve the existing coarse graphics API while suppressing intermediate change notifications safely.

### What warrants a second pair of eyes
- The current batching semantics are a strong first step, but later measurement work may still benefit from explicit render stats showing how often a flush had to wait for a batch to finish.
- It may also be worth reviewing where batching should be used by default in future scene examples versus left as an explicit authoring tool.

### What should be done in the future
- Rerun the full-page all-12 example on hardware and record whether the missing-later-tiles symptom is gone or materially reduced.
- Then continue with the original LOUPE-007 instrumentation plan for renderer and writer stats.

## Step 4: Implement the first end-to-end instrumentation slice, including scene-side JS timing

Once the atomic-frame batching fix was in place, the next user request was the right one: do not keep speculating about whether the full-page all-12 slowdown is caused by JavaScript, renderer work, or writer pacing. Instrument it, including from inside the JS scene itself, so later optimization work can target the actual hot path.

### Prompt Context

**User prompt (verbatim):** "instrument, including inside the JS, then you will be able to do 3 more effectively"

**Assistant interpretation:** Implement the first practical measurement slice from `LOUPE-007`, and make sure it includes scene-side timing from within the JavaScript workload itself rather than only Go-side observations.

**Inferred user intent:** Replace guesswork with evidence before attempting deeper scene optimization.

### What I did
- Added a reusable in-process metrics collector in `runtime/metrics/metrics.go`.
- Extended the JS environment so each owned runtime has a metrics collector available.
- Added a JS-facing `loupedeck/metrics` module with:
  - `metrics.inc(name, delta)`
  - `metrics.observeMillis(name, value)`
  - `metrics.time(name, fn)`
  - `metrics.now()`
- Added runtime tests proving that JS can record counters and timing measurements into the collector.
- Extended `cmd/loupe-js-live/main.go` with new flags:
  - `--log-render-stats`
  - `--log-writer-stats`
  - `--log-js-stats`
  - `--stats-interval`
- Added periodic logging for:
  - renderer-side dirty-display/dirty-tile/flush timing windows,
  - writer deltas plus current queue depth snapshots,
  - JS-side counters and timing summaries.
- Applied the new JS instrumentation immediately to `examples/js/10-cyb-ito-full-page-all12.js` so it records:
  - `scene.renderAll.calls`
  - `scene.renderAll`
  - `scene.loopTicks`
  - `scene.activations`
  - `scene.tile.<name>` for each tile draw timing

### Why
- The current full-page scene is now coherent but still extremely slow. Without measurement, it is too easy to keep blaming the wrong layer.
- Go-side renderer/writer stats alone would still leave a major blind spot: the actual scene-construction time inside the JS workload.
- Adding instrumentation now is the best preparation for the next optimization step, because it should tell us whether the dominant cost is mostly JS scene generation, JS-to-Go graphics calls, retained renderer work, or queue/send pressure.

### What worked
- The instrumentation slice fits naturally into the existing architecture: Go still owns runtime and logging policy, while JS records scene-local work through a narrow metrics API.
- The full test suite remained green after adding the collector, module, and live-runner flags.
- The resulting data model is already good enough for the immediate question of whether `renderAll()` itself is dominating the time budget.

### What didn't work
- This step does not yet include a completed hardware evidence log from the new stats flags; that still needs to be captured in a follow-up run.
- The current implementation logs human-readable summaries rather than a more structured export format. That is fine for first-pass diagnostics but may need refinement later.

### What I learned
- The right first JS-facing API for performance work is not a big profiler framework. A narrow counter/timing API is enough to make scene authorship observable while keeping transport ownership in Go.
- Instrumenting from inside the scene should make the next optimization step much more honest, especially for the current full-page all-12 workload where the expensive part is likely repeated JS-to-Go raster generation.

### What was tricky to build
- The main subtlety was keeping the measurement API narrow and aligned with the existing ownership model. It would have been easy to introduce a sloppy console-like escape hatch, but the real need here was counters and duration measurements, not free-form logging from JS.
- Another subtlety was keeping the live-runner stats window simple enough for first-pass use while still surfacing renderer, writer, and JS information together.

### What warrants a second pair of eyes
- The chosen field names for JS timings/counters in `examples/js/10-cyb-ito-full-page-all12.js` should be reviewed once we have the first real hardware logs, because that log output will reveal whether the names are sufficiently readable and whether any key breakdown is still missing.
- The render-window aggregation in `cmd/loupe-js-live/main.go` is intentionally lightweight, but future work may want more structured output or reusable stats types.

### What should be done in the future
- Run the instrumented full-page all-12 scene on hardware with the new stats flags enabled.
- Capture the first evidence log and summarize whether the dominant cost is scene build time, renderer flush time, writer pacing, or a combination.
- Then use that evidence to drive the next optimization step instead of changing the scene blindly.

## Step 5: Capture the first hardware evidence log and interpret it

After the instrumentation landed, the next immediate task was to use it rather than just admire the new flags. We reran the batched full-page all-12 scene with renderer stats, writer stats, and JS-side metrics all enabled at once.

### What I ran

```bash
go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration 20s \
  --log-events \
  --log-render-stats \
  --log-writer-stats \
  --log-js-stats \
  --stats-interval 1s
```

Evidence log:

```text
/tmp/loupe-cyb-ito-full10-stats-1776020694.log
```

### What the first evidence showed
- JS-side stats showed the scene was rebuilding very frequently:
  - `scene.loopTicks=72..77` per second window
  - `scene.renderAll.calls=72..78` per second window
- JS-side timing showed `renderAll()` averaging about `18–22 ms`, with the `SPIRAL` tile the hottest individual tile at about `5–6 ms` average.
- Go-side render stats showed only **one** non-empty full-page flush in a window and reported flush durations around `1.1–1.5 s`.
- Writer stats showed only one command in the same window and queue depth staying at zero.

### Why this matters
This was the first moment where the measurement plan really paid off. Without the combined instrumentation, it would still be easy to say "the scene is slow, so maybe JavaScript is slow" and stop there. The actual picture is more specific:

- JavaScript is indeed busy, but not busy enough by itself to explain *multi-second* visible update spacing.
- The writer queue is not backing up, so this is not a simple queue-pressure story.
- The full-page flush path is effectively stalling while the scene keeps rebuilding frames.

That means the next optimization step should not start by blaming raw transport ceilings. It should start by asking why the renderer/flush path cannot obtain frames efficiently once the scene is rebuilding almost continuously.

### Best current interpretation
The most plausible current explanation is that the scene is rebuilding the shared full-page retained surface so frequently that the flush path only occasionally gets a stable frame to snapshot and send. In other words, frame atomicity is now correct, but frame *availability* is still poor under the current cadence and workload shape.

### What should be done in the future
- Consider reducing the scene rebuild cadence or making it explicit rather than rebuilding on every animation loop tick.
- Consider scene-side strategies such as active-tile-only updates or staggered tile refresh.
- Consider stronger full-page snapshot/swap semantics if the current shared-surface model still starves the flush path.
- Keep using the new stats path when evaluating each of those changes so the next decision is evidence-based.

## Step 6: Extract the underlying JS metrics implementation so it can move to go-go-goja later

At this point the user clarified an important architectural goal: the recent JS instrumentation work should not just be convenient inside this repo; the *underlying implementation* should be reusable so it can eventually move into `go-go-goja` and serve many JS runtimes, not just the Loupedeck one.

That clarification changed the right shape of the work. The first instrumentation pass had been functionally correct, but the native modules were still conceptually owned by the Loupedeck runtime because they looked up the collector through the Loupedeck JS environment. That would have made later extraction annoying and would have mixed a generic concern (JS metrics collection) with an app-specific environment.

### What I did
- Added a new generic package:
  - `pkg/jsmetrics/jsmetrics.go`
- Moved the real reusable implementation there:
  - generic collector lookup through `runtimebridge`
  - generic binding key (`metricsCollector`)
  - generic module registration helpers
  - configurable module prefix registration (`RegisterModules(registry, prefix)`) so the current runtime can expose `loupedeck/metrics` and `loupedeck/scene-metrics` as one concrete naming choice
- Updated `runtime/js/runtime.go` so it now:
  - binds the collector into `runtimebridge.Values` under the generic metrics key
  - registers metrics modules through `pkg/jsmetrics`
- Reduced the old runtime-local module packages to thin compatibility wrappers.
- Kept the current JS API stable for the repo while moving the implementation substrate into a package that is much easier to transplant into `go-go-goja` later.

### Why
- The metrics collector and scene helper logic are not inherently Loupedeck-specific.
- Future `go-go-goja` reuse should not require dragging along the Loupedeck environment model just to get JS counters/timings.
- The right extraction boundary is below the current app-specific runtime but above the raw collector type: a reusable bridge + module-registration package.

### What worked
- The full test suite stayed green after the refactor.
- The current runtime still exposes the same `loupedeck/metrics` and `loupedeck/scene-metrics` API.
- The underlying implementation is now materially closer to something that could be ported or upstreamed into `go-go-goja`.

### What warrants a second pair of eyes
- If/when this moves into `go-go-goja`, we should decide whether the module names should remain prefix-configurable or whether a different naming convention is preferable there.
- We should also consider whether the collector binding key and module registration helpers belong in one package or two once the code is extracted from this repo.

## Step 7: Write the full project technical report and prepare a fresh reMarkable bundle

Once the ticket contained the design guide, the diary, the instrumentation slices, the first evidence log, and the reusable metrics extraction, the user asked for something bigger: not another incremental note, but a real project report that explains the whole state of the 12-tile cyb-ito performance investigation for a future intern. That was the right time to do it. The ticket had accumulated enough evidence and branching experiments that a new reader would now benefit more from one coherent narrative than from reading the changelog and diary fragments in isolation.

The report needed to do more than restate the latest conclusion. It needed to explain the system, the competing approaches, why each approach seemed promising at the time, what it actually taught us on hardware, and what the current hypotheses are for making the JS canvas tile port performant. In short, it needed to turn a chronology into a technical model.

### Prompt Context

**User prompt (verbatim):** "ok let's write a full project report document in the ticket, about the many different approaches we took and research and hypothesis we ahve over how to make the JS canvas 12 tile ito port work in a perofrmant manner."

**Assistant interpretation:** Create a substantial technical report inside the pacing ticket that synthesizes the approaches tried, the evidence collected, and the current performance hypotheses around the cyb-ito 12-tile JavaScript canvas port.

**Inferred user intent:** Produce a durable intern-facing report that captures the state of the investigation so future work can resume from understanding instead of from transcript archaeology.

### What I did
- Added a new long-form design doc:
  - `design/02-project-technical-report-performing-the-12-tile-javascript-canvas-cyb-ito-port.md`
- Structured the report around:
  - the source artifact and hardware target,
  - the layered architecture baseline,
  - the distinct clocks in the system,
  - the sequence of approaches tried,
  - raw transport baselines,
  - tile-mode versus full-page-mode tradeoffs,
  - the frame-atomic batching fix,
  - the combined renderer/writer/JS metrics evidence,
  - and the current prioritized hypotheses for future work.
- Included prose paragraphs, bullets, Mermaid diagrams, pseudocode, API references, concrete file references, and the first evidence log path.
- Updated the ticket index to surface the new report as a first-class key link.
- Updated tasks/changelog for the expanded ticket bundle.
- Prepared the ticket for a fresh validation and reMarkable upload pass.

### Why
- The investigation is now rich enough that a new intern should not have to reconstruct it from the diary alone.
- The report gives us one place to explain how the architecture, experiments, and measurements fit together.
- A fresh reMarkable upload makes the updated ticket package reviewable away from the terminal.

### What worked
- The report can now act as the primary onboarding document for the 12-tile cyb-ito performance problem.
- The ticket now has a clearer separation of roles:
  - design doc 01 for measurement/tuning architecture,
  - design doc 02 for project-level technical synthesis,
  - diary for chronology,
  - changelog for deltas.

### What didn't work
- At the moment of writing the report, the new bundle still needed a fresh `docmgr doctor` run and reMarkable upload to fully finish the workflow.

### What I learned
- Once a performance investigation crosses several architectural branches, a dedicated synthesis document becomes more valuable than yet another local note.
- The right unit of explanation for this project is not a single optimization tip; it is the relationship between command count, frame coherence, frame availability, JS scene cost, and writer policy.

### What was tricky to build
- The hardest part was deciding what to elevate from diary detail into report-level narrative. The report needed enough chronology to explain why each branch existed, but it also needed to stay organized around concepts and evidence instead of reading like a raw transcript.
- Another tricky part was preserving the distinction between supported facts, strong hypotheses, and future candidate strategies. Blurring those would make the report look more certain than the evidence justifies.

### What warrants a second pair of eyes
- The report’s recommended next-step order—cadence reduction, selective redraw, then coarser native helpers and possibly buffer swap semantics—should be reviewed again after the next hardware evidence pass.
- If the report is later reused outside the ticket, someone should ensure its file references and evidence-log references still point to stable locations.

### What should be done in the future
- Run `docmgr doctor --ticket LOUPE-007 --stale-after 30` again now that the new report exists.
- Upload the refreshed ticket bundle, including the new report, to reMarkable.
- Continue performance experiments using the new report as the main onboarding/reference document.

## Step 8: Rerun the full-page scene after adding rebuild-reason metrics and decide the real next optimization

After the first combined instrumentation pass, we were already strongly suspicious that the full-page scene was rebuilding mostly because of `anim.loop(... renderAll("loop"))`, not because of hidden input callbacks or some accidental dirty-state feedback loop. But suspicion is not the same thing as measurement, and the whole point of `LOUPE-007` is to stop optimizing from vibes. So the next step was to rerun the full-page all-12 hardware measurement now that the scene records explicit rebuild reasons.

### Prompt Context

**User prompt (verbatim):** "alright, run it then, update your diary, device on next steps, and explain what you found and think is next to me in detail."

**Assistant interpretation:** Perform the fresh hardware run with rebuild-reason metrics enabled, record the outcome in the ticket continuity docs, and use the result to decide the next optimization step explicitly.

**Inferred user intent:** Stop talking about probable causes abstractly and produce a concrete answer about what is actually driving the rebuild frequency.

### What I ran

```bash
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test
(timeout 30s go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration 20s \
  --log-render-stats \
  --log-writer-stats \
  --log-js-stats \
  --stats-interval 1s) 2>&1 | tee /tmp/loupe-cyb-ito-full10-reasons-1776023397.log
```

### Evidence log

- `/tmp/loupe-cyb-ito-full10-reasons-1776023397.log`

### What I found
- The fresh no-input run confirmed the main open question directly: after startup, the rebuild stream is dominated by the loop path.
- First logged counters window:
  - `scene.renderAll.reason.initial=1`
  - `scene.renderAll.reason.loop=295`
- Later windows recorded only loop-driven rebuilds:
  - `76`
  - `265`
  - `92`
  - `7`
  - `72`
  - `222`
- No touch- or button-driven rebuild reasons were recorded in this run.
- JS-side `scene.renderAll` timing stayed in roughly the same range as the earlier evidence:
  - about `17.6–18.4 ms` average per measured window
  - `SPIRAL` remained the hottest tile at roughly `4.9–5.4 ms`
- Writer stats again showed a calm queue:
  - `CurrentQueueDepth:0`
  - `MaxQueueDepth:0`
  - only `1–2` commands sent per delayed stats window
- Render windows remained highly variable and frequently very large:
  - `99 ms`
  - `825 ms`
  - `1.26 s`
  - `1.31 s`
  - `1.98 s`
  - `4.10 s`
  - `4.74 s`

### Why this matters
This rerun gives us the piece that was still missing after the first evidence log: a direct answer to the question "who is calling rebuild so often?" In this scene, the answer is now concrete: the animation loop is doing it. There is no evidence from this run that some hidden input callback or extra host event path is causing the rebuild flood.

That does **not** mean the writer is innocent in every possible scene. It means that for this particular full-page all-12 workload, the first thing to fix is not a mystery invalidation source and not queue growth. The scene itself is eagerly requesting rebuilds from the loop path, while the renderer/flush side is taking a long and uneven amount of time to turn dirty state into visible hardware updates.

### Important caveat about the counters
One subtlety showed up while interpreting the rerun: the logged counter windows are not guaranteed to be strict one-second wall-clock windows under heavy load. In `cmd/loupe-js-live/main.go`, the `statsTick` case is serviced by the same `select` loop that also performs `renderer.Flush()`. If a flush call blocks for a long time, the next stats snapshot is delayed too. That means the counter values should be read as "counts per delayed stats window" rather than exact per-second rates.

This caveat matters for absolute throughput interpretation, but it does **not** weaken the main reason-source conclusion. Even with delayed windows, the measurements still show that the rebuild reasons are overwhelmingly `loop` and not `touch` / `button`.

### What worked
- The reason metrics answered the main attribution question cleanly.
- The new evidence is consistent with the earlier render/writer/JS timing picture instead of contradicting it.
- We now have enough evidence to justify the next optimization pass without guessing.

### What didn't work
- The run still ended with the familiar close/read lifecycle noise (`Read error, exiting error="Port has been closed"`).
- The stats windowing in the live runner is still somewhat distorted by long `renderer.Flush()` calls, so the logs are excellent for *relative* attribution but not ideal for precise wall-clock rate measurement.

### What I learned
- The frequent rebuild caller really is the animation loop.
- The next optimization should therefore begin with explicit full-page cadence control rather than with input-path debugging or queue tuning.
- The current stats path is good enough to pick winners between strategies, but if we later need exact per-second rates we may want the stats ticker to emit from a less coupled path.

### What was tricky to build
- The tricky part was not the run itself; it was interpreting the counters honestly. At first glance, some windows look too large to be read as literal one-second throughput samples. Reading `cmd/loupe-js-live/main.go` clarified why: stats emission is delayed when flushes take a long time.
- That means the right reading is about *dominant causes* and *relative pressure*, not blind division by one second.

### What warrants a second pair of eyes
- Someone should review whether `statsTick` should remain in the same loop as `renderer.Flush()` or whether future measurement precision would benefit from decoupling those paths.
- After the next optimization pass, we should confirm that the ratio of loop-driven rebuilds to non-empty full-page flushes materially improves rather than just changing average timings slightly.

### What should be done in the future
- Implement a cadence-limited full-page variant first. The simplest version is to keep `phase` updates flowing but only call `renderAll("loop")` when a target frame interval has elapsed.
- Re-run the same hardware measurement on that cadence-limited variant and compare it against both evidence logs:
  - `/tmp/loupe-cyb-ito-full10-stats-1776020694.log`
  - `/tmp/loupe-cyb-ito-full10-reasons-1776023397.log`
- If cadence limiting helps materially, follow with a staggered redraw experiment for active versus background tiles.
- If cadence limiting does not help enough, investigate stronger full-page snapshot/swap semantics next.
