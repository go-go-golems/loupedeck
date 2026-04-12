---
Title: Layered density measurement runbook
Ticket: LOUPE-007
Status: active
Topics:
    - loupedeck
    - benchmarking
    - performance
    - rendering
DocType: playbook
Intent: operational
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go
      Note: Main runtime entrypoint for future stats logging and controlled scene-mode runs
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js
      Note: Primary scene workload that should be parameterized for density sweeps
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go
      Note: Raw hardware transport benchmark used as the control/baseline
ExternalSources: []
Summary: Practical runbook for future layered-density pacing measurements on real Loupedeck hardware.
LastUpdated: 2026-04-12T06:59:00-04:00
WhatFor: Give future work a repeatable measurement sequence for raw-ceiling checks, layered-scene density sweeps, and interpretation.
WhenToUse: Use when running the future pacing/instrumentation work from LOUPE-007 on real hardware.
---

# Layered density measurement runbook

## Goal

Run a repeatable sequence of raw-ceiling checks and layered-scene hardware trials so that pacing conclusions are evidence-based rather than subjective.

## Preconditions

- Loupedeck Live is physically connected and free.
- `go test ./...` is green.
- future stats flags have been implemented in `cmd/loupe-js-live/main.go`.
- the current prototype scene supports density modes.
- tmux is available.

## Phase 1: capture the raw transport/display baseline

Use the raw benchmark first so retained-scene findings are not overinterpreted.

Example commands:

```bash
go run ./cmd/loupe-fps-bench --display main --width 360 --height 270 --duration 8s
go run ./cmd/loupe-fps-bench --display main --width 90 --height 90 --duration 8s
go run ./cmd/loupe-fps-bench --display left --width 60 --height 270 --duration 8s
```

Record:

- stable FPS
- peak FPS
- device availability errors
- reconnect noise

## Phase 2: run the layered scene in controlled modes

Recommended order:

1. `base`
2. `hud`
3. `scan`
4. `ripple`
5. `main-full`
6. `full`

Example tmux run pattern:

```bash
tmux new-session -d -s loupe-density \
  "cd /home/manuel/code/wesen/2026-04-11--loupedeck-test && \
   go run ./cmd/loupe-js-live \
     --script ./examples/js/07-cyb-ito-prototype.js \
     --duration 30s \
     --flush-interval 16ms \
     --send-interval 35ms \
     --log-render-stats \
     --log-writer-stats \
     --stats-interval 1s"
```

## Phase 3: vary only one pacing dimension at a time

### Vary flush interval

```bash
--flush-interval 16ms
--flush-interval 33ms
--flush-interval 50ms
```

### Vary writer pacing

```bash
--send-interval 35ms
--send-interval 20ms
--send-interval 10ms
```

Do **not** vary both together in the first pass.

## What to capture per run

### Quantitative

- scene mode
- flush interval
- send interval
- flushes/sec
- avg render ms
- max render ms
- displays flushed/sec
- tiles flushed/sec
- writer max queue depth
- failed commands

### Qualitative

- did touch/button feedback feel immediate?
- did scan/ripple effects look stable?
- did the scene stutter, burst, or pause?
- were there reconnect warnings or malformed responses?

## Interpretation cheat sheet

### If render time rises but queue depth stays calm

Likely renderer/composition issue.

### If queue depth rises but render time stays calm

Likely transport pacing issue.

### If both rise together

Likely full-scene density issue on both sides.

### If neither rises much but effect duration feels inconsistent

Likely effect-design issue, not transport.

## Minimal reporting template

```markdown
### Mode: full
- flush interval: 16ms
- send interval: 35ms
- flushes/sec: ...
- avg render ms: ...
- max render ms: ...
- writer max queue depth: ...
- failed commands: ...
- human notes: ...
- interpretation: ...
```

## Stop/cleanup sequence

```bash
tmux send-keys -t loupe-density:0 C-c
sleep 2
tmux capture-pane -pt loupe-density:0 -S -300 > /tmp/loupe-density-final.log
tmux kill-session -t loupe-density
```
