---
Title: Trace capture runbook for excessive full-page rebuild investigation
Ticket: LOUPE-009
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - rendering
    - performance
DocType: playbook
Intent: operational
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go
      Note: Live-runner entrypoint where trace output flags should be added
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js
      Note: Primary workload for the trace investigation
ExternalSources: []
Summary: Operational runbook for capturing JS and Go trace evidence that explains where repeated full-page rebuilds are coming from and how they relate to downstream flushes.
LastUpdated: 2026-04-12T16:22:00-04:00
WhatFor: Provide a repeatable measurement recipe once trace instrumentation is added.
WhenToUse: Use when the trace APIs and live-runner flags exist and a hardware capture needs to be performed or repeated.
---

# Trace capture runbook for excessive full-page rebuild investigation

## Goal

Capture an ordered trace showing the relationship between:

- `anim.loop` ticks,
- `renderAll(...)` begin/end events,
- input-triggered rebuilds if any,
- renderer flush begin/end events,
- and outbound writer activity.

The trace should make it possible to answer how many rebuilds occur between meaningful hardware flushes and whether those rebuilds are almost entirely loop-driven.

## Preconditions

Before using this runbook, the implementation should provide:

- JS trace support in `loupedeck/metrics` or `loupedeck/scene-metrics`
- a bounded trace buffer in the collector
- live-runner trace output flags
- scene instrumentation in `examples/js/10-cyb-ito-full-page-all12.js`

Suggested flags once implemented:

- `--log-js-trace`
- `--log-go-trace`
- `--trace-limit 500`
- optional `--trace-dump-on-exit`

## Primary no-input run

Use the no-input run first because it isolates the animation path.

```bash
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test
log="/tmp/loupe-cyb-ito-full10-trace-$(date +%s).log"
(timeout 30s go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration 20s \
  --log-render-stats \
  --log-writer-stats \
  --log-js-stats \
  --log-js-trace \
  --log-go-trace \
  --stats-interval 1s \
  --trace-limit 500) 2>&1 | tee "$log"
echo "$log"
```

## Optional input-assisted run

Only run this after the no-input trace is understood.

Purpose:

- verify how input-triggered `setActive(...)` rebuilds interleave with loop-triggered rebuilds
- check whether user input materially changes the rebuild/flush ratio

Suggested workflow:

- run in tmux
- press one or two touches/buttons deliberately
- stop after the capture window

## What to look for immediately

### JS-side trace events

Expected first events:

- `renderAll.begin reason=initial`
- `renderAll.end reason=initial`
- `loop.tick`
- `renderAll.begin reason=loop`
- `renderAll.end reason=loop`

### Go-side trace events

Expected first events:

- `flush.tick dirtyDisplays=1 dirtyTiles=0`
- `flush.begin ...`
- `flush.end ops=1 elapsedMs=...`

## Immediate checks

### Check 1: Is the trace dominated by loop-driven rebuilds?

Expected answer for the no-input run:

- yes

If no:

- inspect whether hidden `other` reasons or unexpected input events exist

### Check 2: How many rebuilds happen between non-empty flushes?

Expected answer:

- more than one,
- likely many more than one,
- potentially tens or more depending on the run

### Check 3: Do loop ticks continue while flushes are slow?

Expected answer:

- yes, or at least rebuilds should keep being requested at a much higher rate than flushes complete

### Check 4: Does the writer queue remain calm even during a rebuild flood?

Expected answer for the current full-page scene:

- yes

## Derived summary to compute after each run

For each non-empty flush window compute:

- `rebuilds_since_last_flush`
- `loop_ticks_since_last_flush`
- `input_rebuilds_since_last_flush`
- `writer_sends_since_last_flush`

For the full run compute:

- `total_rebuilds / total_non_empty_flushes`
- `total_loop_ticks / total_non_empty_flushes`
- `total_rebuilds / total_sent_commands`

## Interpretation guide

### If loop ticks and rebuilds vastly outnumber flushes

Interpretation:

- the scene is over-producing retained frames relative to the full-page flush path
- start with cadence limiting

### If rebuilds are not mostly loop-driven

Interpretation:

- there is another invalidation source to investigate before tuning cadence

### If flushes are frequent but writer queue backs up

Interpretation:

- revisit writer pacing or command grouping

### If flushes are rare and writer queue is calm

Interpretation:

- the bottleneck is upstream of queue growth, likely around scene cadence, retained snapshot/flush timing, or their interaction

## Evidence management

Store the resulting log path in the ticket diary and changelog.

Recommended naming:

- `/tmp/loupe-cyb-ito-full10-trace-<timestamp>.log`

## Follow-up after the first trace run

Once the first no-input trace is captured:

1. summarize the dominant source of rebuilds
2. record rebuilds-per-flush ratios
3. decide whether to implement cadence limiting immediately
4. if needed, add one deeper trace layer inside the renderer or writer
