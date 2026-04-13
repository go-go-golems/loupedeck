---
Title: Design - Decompose cmd/loupe-js-live into Focused Command Files
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
      Note: Current all-in-one live runner that mixes CLI, bootstrapping, orchestration, stats, and cleanup
    - Path: cmd/loupe-js-live/
      Note: Target directory for the split command files
    - Path: pkg/device/connect.go
      Note: Device connection is now stable enough that runner decomposition can treat it as a clean dependency
    - Path: runtime/js/module_ui/module.go
      Note: Recently simplified name parsing means the runner no longer needs local input-name helper maps
ExternalSources: []
Summary: >
    Design rationale for splitting cmd/loupe-js-live from one large main.go into a small
    set of focused command-local files. Keeps behavior the same while separating CLI parsing,
    app bootstrap, event loop orchestration, stats formatting, and cleanup helpers.
LastUpdated: 2026-04-13T15:10:00-04:00
WhatFor: Define the target shape and boundaries of the live runner refactor before implementation
WhenToUse: When reviewing or implementing the cmd/loupe-js-live decomposition work
---

# Design - Decompose cmd/loupe-js-live into Focused Command Files

## Executive Summary

`cmd/loupe-js-live/main.go` is now one of the last obvious concentration points of local
complexity in the repo. The device driver has moved to `pkg/device`, connect-time display
profiling is fixed, and input naming duplication is gone. That makes this a good moment to
clean up the live runner itself.

The proposed refactor is intentionally conservative:
- keep the behavior and flags the same,
- keep the code inside `cmd/loupe-js-live/`,
- do **not** introduce a reusable `pkg/runner` abstraction yet,
- split the current file into focused command-local files.

The target outcome is a tiny `main.go`, a clear `run.go` containing the orchestration flow,
and helper files for options, stats/logging, and cleanup.

## Problem Statement

### The command has too many responsibilities in one file

`cmd/loupe-js-live/main.go` currently performs all of the following:
- CLI flag declaration and validation
- script loading from disk
- device connection and deferred cleanup
- display lookup and validation
- listen-loop startup
- JS environment and runtime construction
- retained renderer construction
- presenter/flush callback wiring
- event logging registration
- signal handling
- duration/timeout handling
- periodic stats collection and formatting
- trace filtering/formatting
- display clearing on exit

This is not a package-boundary problem anymore. It is a **local command-complexity** problem.
The code works, but it is harder to read and review because unrelated concerns are interleaved.

### The file is longer than the actual application flow warrants

The real live-runner control flow is short:
1. parse options
2. load script
3. connect device
4. build runtime and renderer
5. wire callbacks
6. wait on listen/signal/timeout/ticker events
7. clean up

That core flow is currently buried inside a long file full of helper types and formatting logic.

### Reuse pressure is not yet high enough for a shared package

There is only one primary JS live runner. Extracting to a reusable `pkg/runner` package today
would likely create a premature abstraction. The right move is smaller: first make the command
internally clean and only extract shared code later if a second consumer clearly needs it.

## Proposed Solution

Keep the implementation entirely inside `cmd/loupe-js-live/`, but split it into focused files.

### Target layout

```text
cmd/loupe-js-live/
  main.go        # tiny entrypoint
  options.go     # flags -> options struct
  run.go         # main application orchestration
  stats.go       # stats windows and formatting helpers
  logging.go     # optional event logging hookup
  cleanup.go     # display clearing and small shutdown helpers
```

### Responsibilities by file

#### `main.go`
Only:
- parse CLI options
- print user-facing errors
- call `run(opts)`
- exit with the correct code

#### `options.go`
Contains:
- the `options` struct
- flag registration
- validation of required fields like `--script`
- normalization of positional-arg errors

#### `run.go`
Contains the actual program flow:
- read script file
- connect device
- collect displays
- start `Listen()` goroutine
- create JS env/runtime
- create retained renderer
- wire present flush callbacks
- run the main `select` loop for signals, timeouts, stats ticks, and listen errors

#### `stats.go`
Contains pure helper logic:
- `renderStatsWindow`
- writer stat diffs
- JS counter/timing formatting
- trace filtering and formatting

These functions do not define the app; they support observability.

#### `logging.go`
Contains optional high-level event logging registration.
This keeps the noisy button/touch/knob observer wiring out of the app bootstrap path.

#### `cleanup.go`
Contains small cleanup helpers like display clearing on exit.
This keeps shutdown logic isolated from startup logic.

## Design Decisions

### Decision 1: Keep the refactor command-local
This is the most important boundary decision. We are **not** extracting a generic runner package yet.
A command-local split is enough to make the code readable without inventing a shared abstraction too early.

### Decision 2: Preserve behavior first, improve internals second
The first implementation slice should keep:
- the same flags,
- the same exit behavior,
- the same stats/tracing output,
- the same renderer/presenter flow.

This is a decomposition refactor, not a behavior redesign.

### Decision 3: Introduce an explicit `options` struct
An `options` struct gives the runner a single configuration object rather than a dozen loose locals.
That makes the orchestration function easier to read and easier to test in the future.

### Decision 4: Keep `stats.go` as helper code, not a new package
Stats formatting belongs near the command that emits it. It is not a reusable library yet.
If another command later wants the exact same output model, we can reconsider extraction.

## Alternatives Considered

### Alternative A: Leave `main.go` alone
Rejected because this is now one of the clearest remaining local complexity hotspots.
The repo architecture improved elsewhere; the command file should catch up.

### Alternative B: Create `pkg/runner` immediately
Rejected because there is only one true consumer right now. This would likely create a generic
abstraction before we understand what should actually be shared.

### Alternative C: Split only `options.go` and leave everything else in `main.go`
Rejected because the value of this refactor comes from surfacing the real app flow in `run.go`.
Only moving flags would not materially improve readability.

## Implementation Plan

High-level sequence:
1. introduce `options.go`
2. move orchestration into `run.go`
3. move stats helpers into `stats.go`
4. move event logging into `logging.go`
5. move cleanup helpers into `cleanup.go`
6. keep behavior identical and validate with `go test ./...`

## Open Questions

1. Should `run.go` expose a small `app` struct or just use free functions?
   - Default answer: start with free functions unless state sharing becomes awkward.
2. Should trace/stat formatting stay in the command long-term?
   - Default answer: yes, until another command needs the same abstraction.
3. Should script loading be its own helper file?
   - Probably not yet; that feels too small to justify another file.

## References

- `design-doc/04-implementation-plan-clean-architecture-reorganization.md` — current repo-wide cleanup plan
- `cmd/loupe-js-live/main.go` — current all-in-one implementation that motivates this split
- `pkg/device/` — now-clean driver boundary the runner depends on

