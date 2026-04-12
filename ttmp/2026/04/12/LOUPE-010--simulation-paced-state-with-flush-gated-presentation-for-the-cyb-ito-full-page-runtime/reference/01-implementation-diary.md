---
Title: Implementation diary
Ticket: LOUPE-010
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - animation
    - rendering
    - performance
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-010--simulation-paced-state-with-flush-gated-presentation-for-the-cyb-ito-full-page-runtime/design/01-implementation-plan-simulation-paced-state-with-flush-gated-presentation.md
      Note: Main design guide for the presenter refactor
ExternalSources: []
Summary: Chronological diary for the forward-only presenter refactor that separates simulation state updates from flush-gated frame production.
LastUpdated: 2026-04-12T17:07:00-04:00
WhatFor: Preserve what changed, why, and how the new simulation/presentation architecture was implemented.
WhenToUse: Use when resuming or reviewing the LOUPE-010 refactor work.
---

# Diary

## Goal

Replace the current loop-driven full-page redraw model with the correct architecture for this hardware path: simulation-paced state updates plus flush-gated one-frame-in-flight presentation.

## Step 1: Create the dedicated presenter-refactor ticket and write the implementation plan

After the trace work in `LOUPE-009`, the next question was no longer observational. It was architectural. The trace evidence had shown enough to justify a real design change: the scene was rebuilding itself far too often while the flush path was still busy, so the runtime needed a better control model rather than another small local tweak. The user explicitly said not to worry about backwards compatibility, because we are still in the stage of getting the single cyb-ito full-page path right.

That made it reasonable to create a new ticket specifically for the presenter refactor. The main output of this first step is a concrete implementation plan describing the correct simulation/presentation model, what should change, what should not, and how to phase the implementation safely.

### Prompt Context

**User prompt (verbatim):** "Create an implementation plan for the correct simulation, no backwards compatiblity since we are still trying to get the single ito thing working.

Then create detailed tasks for the refactor / change, and then implement it task by task, committing at appropriate intervals, and write a detailed diary."

**Assistant interpretation:** Create a dedicated ticket and implementation plan for the forward-only presenter architecture, break it into tasks, then execute the tasks incrementally with diary updates and focused commits.

**Inferred user intent:** Stop patching symptoms. Replace the wrong architecture with the correct one and document the work like the earlier Loupedeck ticket streams.

### What I did
- Created the new ticket workspace for `LOUPE-010`.
- Wrote the main implementation-plan document describing:
  - the current wrong control model,
  - the correct simulation-versus-presentation separation,
  - the proposed pure-Go presenter runtime,
  - the proposed JS `loupedeck/present` module,
  - the live-runner refactor,
  - the scene migration,
  - and the expected trace-level outcomes.
- Prepared to add task breakdown and ticket bookkeeping next.

### Why
- The architecture change deserves its own focused ticket and diary.
- The implementation plan creates a stable reference point before code starts moving.
- A forward-only ticket avoids mixing “correct new model” work with the backwards-looking trace investigation.

### What worked
- The new ticket cleanly frames the work as a presenter refactor rather than another measurement tweak.
- The implementation plan now records the intended target model before any code changes muddy the picture.

### What didn’t work
- At this moment, the ticket still needed detailed tasks, bookkeeping, and the actual code changes.

### What should be done in the future
- Add detailed tasks and ticket bookkeeping.
- Archive scripts under the ticket’s `scripts/` directory as the work proceeds.
- Implement the presenter runtime first, then JS bindings, then live-runner wiring, then scene migration, then trace validation.

## Step 2: Add the pure-Go one-frame-in-flight presenter runtime

With the implementation plan and task list in place, the first actual code slice was the pure-Go presenter runtime. This had to come first because everything else depends on it. The JS module, the live-runner refactor, and the scene migration all need a runtime that already knows how to do the basic job: coalesce repeated invalidations into one future render/flush of the latest state.

### What I did
- Added a new package:
  - `runtime/present/`
- Implemented:
  - `runtime/present/runtime.go`
  - `runtime/present/runtime_test.go`
- The runtime now supports:
  - render callback registration
  - flush callback registration
  - `Invalidate(reason)`
  - latest-reason-wins dirty coalescing
  - explicit `Start(ctx)`
  - explicit `Close()`
- The presenter loop is deliberately simple:
  - wait until dirty and callbacks exist
  - render one frame
  - flush one frame
  - if more invalidations happened while busy, immediately loop again on the latest pending reason
- Added tests covering:
  - coalescing while the first flush is blocked
  - invalidation that happens before callbacks are installed
  - strict serial render/flush behavior
  - shutdown behavior
- Archived the first ticket-local scripts:
  - `scripts/01-create-ticket.sh`
  - `scripts/02-go-test-phase-a-present-runtime.sh`
- Ran:

```bash
gofmt -w runtime/present/*.go
go test ./runtime/present/...
go test ./...
```

### Why
- This is the structural replacement for the old “render every animation tick” model.
- Putting it in a pure-Go package first keeps the behavior testable and independent of the JS bridge.
- The coalescing semantics need to be proven before they are trusted as the basis of the refactor.

### What worked
- The presenter runtime expresses the intended one-frame-in-flight model directly.
- The full repository test suite passed after adding the new package.
- The tests give us confidence that repeated invalidations while a flush is in progress will collapse correctly.

### What didn’t work
- This slice alone does not yet change the runtime behavior of the full-page scene, because nothing is wired to the presenter yet.
- Error handling in the callbacks is intentionally minimal at this stage; the live runner will need to decide how to surface presenter callback failures cleanly.

### What I learned
- The one-frame-in-flight model is small enough to capture cleanly as a pure-Go loop.
- It is helpful to treat missing callbacks as “not ready yet” while still preserving dirty state, so early invalidations are not lost.

### What was tricky to build
- The main subtle point was deciding how `Invalidate(reason)` should behave while a frame is already being processed. The chosen rule is: keep only the latest reason and ensure at least one future present occurs once the current one finishes.
- Another subtle point was making `Close()` deterministic enough for tests while keeping the runtime small.

### What warrants a second pair of eyes
- The current runtime drops callback errors silently after the callback returns them. That may be acceptable for the substrate, but the live runner should likely add explicit logging around its presenter callbacks.
- If later we need richer reasoning than "latest reason wins", we may want to carry a small reason set or category summary instead of a single string. For the current single-scene goal, that would be premature.

### What should be done in the future
- Implement Phase B next: environment ownership plus the JS `loupedeck/present` module.
- Then wire the live runner and migrate the full-page example onto the new presenter path.
