---
Title: Textbook trace-driven investigation of excessive full-page rebuild calls
Ticket: LOUPE-009
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - animation
    - rendering
    - benchmarking
    - performance
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js
      Note: Current full-page scene under investigation; contains the `anim.loop(... renderAll("loop"))` path whose rebuild frequency needs explicit trace attribution
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/anim/runtime.go
      Note: Defines the current global animation loop cadence via `FrameInterval` and `Host.SetInterval`
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_anim/module.go
      Note: JS-facing `loupedeck/anim` bindings where loop callbacks cross into the owner-thread runtime
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface.go
      Note: Retained surface batching and snapshot semantics relevant to whether repeated rebuilds are overlapping coherent frames
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go
      Note: Go-side retained renderer flush path that must be correlated with JS rebuild activity
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go
      Note: Current live-runner stats loop and the best initial place to emit trace summaries or dump trace buffers
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics.go
      Note: Current counter/timing collector that should be extended with bounded ordered trace events
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/jsmetrics/jsmetrics.go
      Note: Current reusable JS metrics binding layer where `trace(...)` APIs can be added without making the design Loupedeck-specific
    - Path: /tmp/loupe-cyb-ito-full10-stats-1776020694.log
      Note: First combined render/writer/JS evidence log before rebuild-reason counters existed
    - Path: /tmp/loupe-cyb-ito-full10-reasons-1776023397.log
      Note: Follow-up hardware evidence log showing that rebuild reasons are dominated by `loop`
ExternalSources: []
Summary: Intern-facing design guide for a new ticket dedicated to tracing excessive full-page rebuild activity, identifying exactly where rebuild calls originate, and correlating JS-side rebuilds with renderer flushes and writer/device output.
LastUpdated: 2026-04-12T16:20:00-04:00
WhatFor: Turn the current suspicion about rebuild frequency into a concrete trace-driven investigation with explicit instrumentation, expected evidence, and interpretation rules.
WhenToUse: Use when implementing or reviewing trace instrumentation for the full-page all-12 cyb-ito scene, or when deciding whether the next optimization should target loop cadence, dirty-state propagation, renderer scheduling, or writer policy.
---

# Textbook trace-driven investigation of excessive full-page rebuild calls

## Executive summary

The current full-page all-12 cyb-ito scene is visibly slow on hardware, and recent measurements have already established two important facts. First, the rebuild flood is dominated by the animation loop rather than by user input. Second, the writer queue is not backing up, which means the visible slowness is not well explained by simple queue pressure. That is good progress, but it still leaves a practical engineering question unanswered in the form a performance engineer actually needs:

> In exact event order, where are the rebuild calls coming from, how many happen between hardware-visible flushes, and what is the relationship between JS rebuild activity and Go-side flush/write activity?

Counters and timing summaries are enough to prove dominance by category. They are not enough to reconstruct sequence. This ticket exists to fill that gap. The goal is to add a lightweight, bounded trace facility that can answer questions like:

- Which call site triggered the rebuild?
- Did the rebuild begin because of the animation loop, initial boot, or input?
- How many rebuilds happened before the next non-empty flush?
- Are rebuilds being coalesced before flush, or is every rebuild forcing downstream work?
- Are long flushes delaying stats emission so much that counters look worse than they are?
- Is the scene suffering from one over-eager producer, or several coupled paths?

This guide explains what to instrument, why it matters, what results are expected, and how to interpret them without fooling ourselves.

## Why this ticket exists separately from LOUPE-007

`LOUPE-007` already established the broad pacing/tuning problem and produced the first evidence logs. That ticket now contains the larger performance narrative: tile mode versus full-page mode, batching, metrics, and the likely need for cadence control.

This new ticket exists because the next debugging question is narrower and deserves a sharper tool. We are no longer asking only, "is the scene too slow?" We are now asking, "what is the precise event sequence that generates the rebuild flood, and how does that sequence line up against actual renderer and writer activity?"

That is trace-analysis work, not just counter analysis.

## Current state of knowledge

The investigation begins from these already-supported facts:

- The current scene under test is:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`
- The full-page scene currently calls:
  - `renderAll("initial")` at startup
  - `renderAll(why)` from `setActive(...)` on input
  - `renderAll("loop")` from `anim.loop(1400, ...)`
- The animation loop implementation in `runtime/anim/runtime.go` uses:
  - `FrameInterval = 16ms`
  - `Host.SetInterval(r.FrameInterval, step)`
- Therefore the loop is a scene-global timer with a target cadence around `60 Hz`, not a per-tile timer.
- The fresh hardware evidence log `/tmp/loupe-cyb-ito-full10-reasons-1776023397.log` showed that in a no-input run the rebuild reasons are dominated by:
  - `scene.renderAll.reason.loop`
- The same log also showed:
  - calm writer queue depth (`0`)
  - highly variable full-page render windows, often very large

That is enough to justify the current leading hypothesis:

> The full-page scene is an over-eager rebuild producer. It asks for new full-page retained frames much more frequently than the renderer/flush path can turn those requests into visible hardware updates.

However, this is still a hypothesis about sequence, not a complete event timeline.

## The exact question this ticket should answer

The narrow question for this ticket is:

> For the full-page all-12 scene, what is the precise runtime event sequence between `anim.loop` callbacks, `renderAll(...)` entry/exit, retained-surface dirtying, renderer flush attempts, and writer/device-visible output?

That breaks into a few more specific sub-questions:

1. Are rebuild calls coming only from the expected high-level call sites, or are there hidden/secondary paths we are not seeing in current counters?
2. How many loop ticks happen before the next non-empty flush?
3. How many `renderAll(...)` begin/end pairs happen before one full-page op is flushed?
4. Is the rebuild stream continuous even while a flush is in progress?
5. Does the current stats loop exaggerate apparent rates because long `renderer.Flush()` calls delay stats emission?
6. Are there opportunities for coalescing or cadence control that become obvious once we can see sequence rather than summary?

## Core mental model

The system we are tracing is not one function. It is a chain.

```mermaid
flowchart TD
    A[anim.loop tick] --> B[JS callback on owner thread]
    B --> C[phase.set(t)]
    C --> D[renderAll begin]
    D --> E[main.batch clear + redraw]
    E --> F[renderAll end]
    F --> G[display dirty]
    G --> H[renderer flush tick]
    H --> I[non-empty flush]
    I --> J[Display.Draw]
    J --> K[writer command send]
    K --> L[device-visible frame]

    style A fill:#214d2f,stroke:#5fbf7a
    style D fill:#1a3a5c,stroke:#4aa3ff
    style H fill:#5c3a1a,stroke:#ffad4a
    style K fill:#5c1a3a,stroke:#ff5ca3
```

Counters collapse this chain into totals. Trace events let us see it as an ordered sequence.

## Why counters and timings are not enough

The current metrics collector records:

- counters
- timing summaries

That lets us answer:

- how many loop rebuilds happened,
- how many renderAll calls happened,
- and how long renderAll took on average.

But it cannot answer:

- whether two rebuilds occurred before or after a flush began,
- whether a non-empty flush happened immediately after a renderAll end or much later,
- whether rebuilds were coalesced downstream,
- or how many rebuilds occurred between visible device updates.

This is why trace events are the right next tool.

## Proposed instrumentation design

### Design goals

The trace facility should be:

- lightweight enough to run on real hardware without destroying the workload,
- bounded so it cannot grow without limit,
- generic enough to remain reusable outside the Loupedeck runtime,
- and simple enough that the resulting logs are readable by a human.

### Design anti-goals

The trace facility should **not** initially be:

- a full profiler,
- a massive structured logging framework,
- or an always-on stack-capture system.

This ticket is about causal breadcrumbs, not maximal telemetry.

## Proposed Go-side trace collector additions

Relevant file:

- `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics.go`

Add a bounded ordered event buffer. A sketch:

```go
type TraceEvent struct {
    Seq       uint64
    TimeUnixNanos int64
    Name      string
    Fields    map[string]string
}

type Collector struct {
    mu       sync.Mutex
    counters map[string]int64
    timings  map[string]TimingStats
    trace    []TraceEvent
    nextSeq  uint64
    maxTrace int
}
```

Proposed operations:

- `Trace(name string, fields map[string]string)`
- `TraceSnapshot() []TraceEvent`
- `TraceSnapshotAndReset() []TraceEvent`
- optional constructor/config for buffer size (`maxTrace`, perhaps default `500` or `1000`)

### Why a ring buffer

We want:

- last N important events,
- stable memory usage,
- and the ability to dump the most recent sequence after a hardware run.

A ring buffer gives that without turning trace collection into another performance problem.

## Proposed reusable JS binding additions

Relevant file:

- `/home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/jsmetrics/jsmetrics.go`

Add reusable JS-facing trace methods.

### Low-level API

```javascript
const metrics = require("loupedeck/metrics");
metrics.trace("renderAll.begin", { reason: "loop" });
```

### Scene helper API

```javascript
const sceneMetrics = require("loupedeck/scene-metrics").create("scene");
sceneMetrics.trace("renderAll.begin", { reason: "loop", active: String(active.get()) });
```

Suggested behavior:

- string event name
- small flat field object
- values coerced to strings for cheap serialization

This keeps the API generic and still exportable to `go-go-goja` later.

## Proposed trace events for the current scene

Relevant scene file:

- `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`

Only add the high-value breadcrumbs first.

### JS-side events

1. `loop.tick`
   - fields:
     - `phase`
2. `renderAll.begin`
   - fields:
     - `reason`
     - `active`
3. `renderAll.end`
   - fields:
     - `reason`
     - `active`
4. `setActive`
   - fields:
     - `idx`
     - `why`
5. `renderAll.skip` (future if cadence-throttling is introduced)
   - fields:
     - `reason`
     - `whySkipped`

### Go-side events

Likely initial locations:

- `cmd/loupe-js-live/main.go`
- optionally later `runtime/render/visual_runtime.go`
- optionally later `writer.go`

Important first breadcrumbs:

1. `flush.tick`
   - fields:
     - `dirtyDisplays`
     - `dirtyTiles`
2. `flush.begin`
   - fields:
     - `dirtyDisplays`
     - `dirtyTiles`
3. `flush.end`
   - fields:
     - `ops`
     - `elapsedMs`
4. `writer.sent`
   - fields:
     - `queuedCommandsDelta`
     - `sentCommandsDelta`
     - `currentQueueDepth`

The initial version does not need every internal renderer function instrumented. We mainly need enough breadcrumbs to align JS rebuilds with flush outcomes.

## Recommended initial live-runner interface

Relevant file:

- `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go`

Add flags such as:

- `--log-js-trace`
- `--log-go-trace`
- `--trace-limit 500`
- optionally `--trace-dump-on-exit`

Suggested first output shape:

```text
js trace:
  #101 loop.tick phase=0.184
  #102 renderAll.begin reason=loop active=0
  #103 renderAll.end reason=loop active=0
  #104 loop.tick phase=0.197
  #105 renderAll.begin reason=loop active=0
  #106 renderAll.end reason=loop active=0

go trace:
  #201 flush.tick dirtyDisplays=1 dirtyTiles=0
  #202 flush.begin dirtyDisplays=1 dirtyTiles=0
  #203 flush.end ops=1 elapsedMs=1265.17
```

This is enough to answer sequence questions directly.

## Derived metrics to compute from the trace

The trace is the raw source. We should also compute a few derived metrics from it because these are the ratios the user keeps asking about.

### Per-flush derived values

For each non-empty flush:

- `rebuilds_since_last_non_empty_flush`
- `loop_ticks_since_last_non_empty_flush`
- `renderAll_begins_since_last_non_empty_flush`
- `renderAll_ends_since_last_non_empty_flush`
- `writer_sends_for_this_flush`

### Cumulative derived values

Across the whole run:

- `total_renderAll_begins / total_non_empty_flushes`
- `total_loop_ticks / total_non_empty_flushes`
- `total_renderAll_begins / total_writer_sent_commands`

### Why these matter

These turn the vague idea of “many rebuilds per frame” into an explicit measured quantity.

## Expected results and how to interpret them

### Expected result A: loop is the dominant rebuild initiator

Most likely trace pattern:

```text
loop.tick
renderAll.begin reason=loop
renderAll.end reason=loop
loop.tick
renderAll.begin reason=loop
renderAll.end reason=loop
... many times ...
flush.begin
flush.end ops=1
```

Interpretation:

- The current producer is the animation loop.
- Cadence control is the first rational optimization.

### Expected result B: rebuilds cluster heavily between non-empty flushes

Most likely derived metric:

- many `renderAll.begin/end` pairs per non-empty flush

Interpretation:

- The scene is outproducing the flush path.
- It is reasonable to throttle or coalesce rebuilds.

### Expected result C: some rebuilds may never lead to downstream output

Possible trace pattern:

```text
renderAll.begin
renderAll.end
renderAll.begin
renderAll.end
renderAll.begin
renderAll.end
flush.begin
flush.end ops=1
```

Interpretation:

- Dirty-state updates are effectively being overwritten/coalesced before they become visible.
- This is good evidence for reducing rebuild frequency instead of trying to "push harder" downstream.

### Expected result D: stats-window distortion is real

Possible trace pattern:

- long gap between `flush.begin` and `flush.end`
- multiple JS events occur before the next stats summary appears

Interpretation:

- Current stats summaries are useful for attribution but poor for exact wall-clock rates.
- We should be careful when reading counts as “per second.”

### Less likely but important result E: hidden rebuild sources exist

Possible trace pattern:

- `renderAll.begin reason=other`
- or rebuilds following unexpected event types

Interpretation:

- Revisit the scene code or runtime bindings for accidental extra invalidation paths.
- This would change optimization priorities.

## Implementation order

### Phase 0: prepare the analysis path

- Add trace event support to `runtime/metrics/metrics.go`
- Keep it bounded and cheap
- Add tests for ordering, capacity, snapshot, and reset behavior

### Phase 1: expose trace APIs to JS

- Add `metrics.trace(...)`
- Add `sceneMetrics.trace(...)`
- Add tests that JS can emit trace events into the collector

### Phase 2: instrument the current full-page scene

- Add the breadcrumb events listed above to `examples/js/10-cyb-ito-full-page-all12.js`
- Do **not** trace inside every tile renderer yet; stay at scene-level first

### Phase 3: instrument the Go-side flush boundary

- Add a minimal Go trace path in `cmd/loupe-js-live/main.go`
- Optionally extend later into `runtime/render/visual_runtime.go`

### Phase 4: run controlled capture on hardware

- reuse the full-page all-12 scene
- run without user input first
- then optionally run one input-assisted session
- save trace log and summary

### Phase 5: interpret and decide next optimization

- If traces confirm loop domination and many rebuilds per flush, implement cadence limiting next
- If traces reveal hidden invalidation, fix the invalidation source first
- If traces reveal flush overlap issues, consider stronger buffering/snapshot semantics

## Example pseudocode for the scene-side trace

```javascript
const sceneMetrics = require("loupedeck/scene-metrics").create("scene");

function renderAll(reason) {
  sceneMetrics.trace("renderAll.begin", {
    reason,
    active: String(active.get()),
  });

  const result = sceneMetrics.recordRebuild(reason, () => {
    main.batch(() => {
      main.clear(0);
      const t = phase.get() * Math.PI * 2;
      const a = active.get();
      for (let i = 0; i < 12; i++) {
        const { x, y } = tileRect(i);
        drawTile(i, x, y, t, i === a);
      }
      drawText(main, lastEvent.get(), 315, 248, 120, 42, 14);
    });
  });

  sceneMetrics.trace("renderAll.end", {
    reason,
    active: String(active.get()),
  });
  return result;
}

anim.loop(1400, t => {
  sceneMetrics.trace("loop.tick", { phase: String(t) });
  sceneMetrics.recordLoopTick();
  phase.set(t);
  renderAll("loop");
});
```

## Example pseudocode for derived analysis

```text
for each event in orderedTrace:
  if event.name == "renderAll.begin":
    rebuildsSinceLastFlush += 1
  if event.name == "loop.tick":
    loopTicksSinceLastFlush += 1
  if event.name == "flush.end" and event.fields["ops"] > 0:
    emit summary(rebuildsSinceLastFlush, loopTicksSinceLastFlush)
    reset counters
```

## Working rules for this ticket

- Do not begin with full stack traces on every loop tick.
- Do not instrument every tile renderer unless the scene-level trace is inconclusive.
- Keep the trace generic enough that `pkg/jsmetrics` remains reusable.
- Prefer bounded trace buffers over unbounded logs.
- Interpret counters and traces together; neither alone is enough.
- Separate “who asked for rebuild?” from “who actually flushed output?”

## Expected outcome of the ticket

If this ticket succeeds, we should leave with:

1. a concrete trace facility in the reusable metrics substrate,
2. a scene-level and flush-level trace for the current full-page all-12 workload,
3. direct evidence showing how many rebuilds happen between non-empty flushes,
4. and a much more defensible choice about the next optimization step.

The most likely next optimization still appears to be cadence limiting, but this ticket exists to make that decision trace-backed rather than merely counter-backed.

## See Also

- `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-007--layered-animation-pacing-measurement-and-tuning-for-loupedeck-js-scenes/design/02-project-technical-report-performing-the-12-tile-javascript-canvas-cyb-ito-port.md` — Broader performance report that motivated this narrower trace-analysis ticket
- `/tmp/loupe-cyb-ito-full10-stats-1776020694.log` — First combined render/writer/JS metrics evidence
- `/tmp/loupe-cyb-ito-full10-reasons-1776023397.log` — Follow-up evidence proving loop-dominant rebuild reasons in the no-input run
