---
Title: Implementation Plan - cmd/loupe-js-live Decomposition
Ticket: LOUPE-008
Status: active
Topics:
    - architecture
    - refactoring
    - analysis
    - code-quality
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupe-js-live/main.go
      Note: Current monolithic command implementation to be split
    - Path: cmd/loupe-js-live/
      Note: Target directory for the decomposed command-local files
ExternalSources: []
Summary: >
    Concrete phased implementation plan for decomposing cmd/loupe-js-live into a tiny entrypoint,
    a runner/orchestration file, and helper files for options, stats, logging, and cleanup while
    preserving current behavior.
LastUpdated: 2026-04-13T15:10:00-04:00
WhatFor: Execute the live-runner split in small, reviewable phases
WhenToUse: When implementing or reviewing the cmd/loupe-js-live decomposition work
---

# Implementation Plan - cmd/loupe-js-live Decomposition

## Executive Summary

This document turns the live-runner decomposition idea into a concrete sequence of steps.
The work is intentionally small-scope and behavior-preserving. The goal is to make
`cmd/loupe-js-live` easier to understand without introducing a reusable runner framework yet.

## Problem Statement

`cmd/loupe-js-live/main.go` is still responsible for too many unrelated concerns.
That makes review harder and obscures the actual application flow. Now that device setup
and input naming are cleaned up, this is a good next target.

## Proposed Solution

Split the command into a handful of files under `cmd/loupe-js-live/`, each with a clear role,
while preserving the current flags, output shape, and control flow.

### Target file map

```text
cmd/loupe-js-live/
  main.go
  options.go
  run.go
  stats.go
  logging.go
  cleanup.go
```

## Design Decisions

### Decision 1: No shared `pkg/runner` yet
Keep everything command-local until a second consumer proves what should be shared.

### Decision 2: Behavior-preserving refactor
No semantic changes in this slice. A reviewer should be able to compare before/after and see that
only structure changed.

### Decision 3: Options first
Introduce `options.go` first so later file splits have a clean input boundary.

## Alternatives Considered

### Alternative A: Full shared runner package now
Rejected as premature.

### Alternative B: Do nothing
Rejected because the command is one of the last local complexity hotspots.

### Alternative C: Only move flags
Rejected because it would not materially improve readability.

## Implementation Plan

### Phase A — carve out options parsing
1. Add `options.go`
2. Introduce an `options` struct containing all current flag values
3. Move flag setup and validation into `parseOptions() (options, error)`
4. Make `main.go` use `parseOptions()`

**Validation:**
```bash
go test ./cmd/loupe-js-live ./...
```

### Phase B — move orchestration into `run.go`
1. Add `run.go`
2. Move script loading, device connection, runtime setup, renderer setup, and the main select loop into `run(opts)`
3. Shrink `main.go` to parse + call + exit

**Validation:**
```bash
go test ./...
```

### Phase C — extract stats helpers
1. Add `stats.go`
2. Move:
   - `renderStatsWindow`
   - writer stat diff helpers
   - JS counter/timing formatting
   - trace filtering/formatting helpers
3. Keep function names the same where possible to reduce churn

**Validation:**
```bash
go test ./...
```

### Phase D — extract logging and cleanup helpers
1. Add `logging.go`
2. Move `registerEventLogging`
3. Add `cleanup.go`
4. Move `clearDisplays`

**Validation:**
```bash
go test ./...
```

### Phase E — polish and review pass
1. Ensure `main.go` is small and obvious
2. Ensure `run.go` reads top-to-bottom as app flow
3. Run `gofmt -w cmd/loupe-js-live/*.go`
4. Run `go test ./...`
5. Update the LOUPE-008 diary and changelog

## Open Questions

1. Should `run.go` also own small helper structs for runtime state?
2. Is `cleanup.go` worth a separate file if it ends up containing only one helper?
3. Do we want a follow-up commit to simplify `cmd/loupe-fps-bench` similarly?

## References

- `design-doc/05-design-decompose-cmd-loupe-js-live-into-focused-command-files.md`
- `design-doc/04-implementation-plan-clean-architecture-reorganization.md`
- `cmd/loupe-js-live/main.go`

