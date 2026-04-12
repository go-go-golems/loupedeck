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

## Step 3: Add presenter ownership to the JS environment and expose `loupedeck/present`

With the pure-Go presenter runtime in place, the next step was to make it reachable from JavaScript. That required two things: the environment needed to own a presenter instance, and the owned JS runtime needed a new native module that lets scripts register a frame callback and invalidate presentation without directly calling `renderAll(...)` from the animation loop.

### What I did
- Updated:
  - `runtime/js/env/env.go`
  - `runtime/js/runtime.go`
- Added a new module:
  - `runtime/js/module_present/module.go`
- The environment now owns:
  - `Present *present.Runtime`
- The JS runtime now registers:
  - `loupedeck/present`
- The new JS module exposes:
  - `present.onFrame(fn)`
  - `present.invalidate(reason)`
- The `onFrame` callback is wired through the owner-thread runtime model, so frame callbacks still execute safely on the JS owner thread.
- Added tests in `runtime/js/runtime_test.go` proving:
  - JS can register a frame callback and receive the correct reason string
  - repeated invalidations coalesce to the latest reason across a blocked flush
- Archived a new ticket-local script:
  - `scripts/03-go-test-phase-b-present-module.sh`
- Ran:

```bash
gofmt -w runtime/js/env/env.go runtime/js/runtime.go runtime/js/runtime_test.go runtime/js/module_present/module.go
go test ./runtime/js/...
go test ./...
```

### Why
- The full-page scene cannot migrate to the correct architecture until JavaScript has a way to separate simulation updates from frame production.
- Wiring the presenter through the environment keeps the structure consistent with the rest of the runtime.
- Owner-thread callback settlement is still mandatory; changing the presentation model does not relax the goja safety rules.

### What worked
- The new JS-facing API now exists and works under test.
- Invalidations can be requested from JavaScript without forcing immediate rendering.
- The coalescing rule already holds across the JS bridge, not just inside the pure-Go presenter tests.

### What didn’t work
- This slice still does not change the live-runner behavior yet, because the presenter is not the main presentation driver until the live runner is refactored.
- The existing example scripts are not yet using the new module.

### What I learned
- The new presentation API fits the existing runtime ownership model naturally.
- Testing coalescing through the JS bridge was worthwhile because it proves the semantics survive owner-thread settlement rather than only existing in the pure-Go layer.

### What was tricky to build
- The most delicate part was making sure `present.onFrame` stores a callback that settles onto the owner thread correctly and gracefully no-ops if the runtime is already closed.
- Another small but real issue was the test shape: because multiple snippets run in the same VM, the second snippet could not safely redeclare `present` with `const`.

### What warrants a second pair of eyes
- Once the full-page scene is migrated, someone should review whether `present.invalidate(reason)` wants stronger reason semantics than a plain string.
- If later we need richer presentation APIs like `invalidateLatest()` or explicit present stats, those should still be layered on top of this minimal model rather than replacing it.

### What should be done in the future
- Refactor the live runner next so the presenter becomes the primary frame-production path.
- Then migrate the full-page cyb-ito scene to `loupedeck/present` and rerun the trace comparison.

## Step 4: Refactor the live runner around the presenter and migrate the full-page scene

Once the pure-Go presenter runtime and the JS `loupedeck/present` module were both in place, the next step was the real architectural switch: make the live runner use the presenter as the intended frame-production path, and migrate the full-page all-12 scene so simulation updates state and invalidates presentation rather than directly calling `renderAll(...)` from the animation loop.

### What I did
- Updated:
  - `cmd/loupe-js-live/main.go`
  - `examples/js/10-cyb-ito-full-page-all12.js`
- In the live runner:
  - removed the old periodic full-page flush ticker from the intended presentation path
  - installed the presenter’s flush callback around `renderer.Flush()`
  - started the presenter runtime after the script load
  - accumulated render statistics from presenter-driven flushes using a small mutex around the render-stats window
  - preserved the stats/trace dump path
- In the full-page scene:
  - added `const present = require("loupedeck/present")`
  - registered:

```javascript
present.onFrame(reason => {
  renderAll(reason || "present");
});
```

  - changed startup from direct `renderAll("initial")` to:

```javascript
ui.show("full-page-all12");
present.invalidate("initial");
```

  - changed the animation loop from direct redraw to state-update-plus-invalidate:

```javascript
anim.loop(1400, t => {
  sceneMetrics.recordLoopTick();
  phase.set(t);
  present.invalidate("loop");
});
```

  - changed input paths to invalidate presentation rather than directly calling `renderAll(...)`
- Archived a new ticket-local script:
  - `scripts/04-go-test-phase-cd-live-runner-and-scene.sh`
- Ran:

```bash
gofmt -w cmd/loupe-js-live/main.go
go test ./...
```

### Why
- This is the moment the architecture actually changes. Before this slice, the new presenter existed but the full-page runtime still behaved the old way.
- The scene must stop treating the simulation clock as the presentation driver.
- The live runner must stop treating a periodic flush ticker as the intended whole-frame present policy.

### What worked
- The full repository test suite still passed after the refactor.
- The scene is now structurally aligned with the intended architecture: simulation updates state, presenter owns frame production.
- The live runner now has a concrete one-frame-in-flight presentation path ready for hardware validation.

### What didn’t work
- We have not yet rerun the hardware trace after the architectural switch, so we do not yet have the new rebuilds-per-flush ratio.
- The live runner still uses the older trace/metrics field names in places, so the next trace run should be inspected to see whether presenter-specific breadcrumbs are worth adding.

### What I learned
- The actual code change is not large once the architecture is clear. Most of the difficulty was in proving the old model was wrong before touching it.
- Moving `ui.show(...)` before the initial `present.invalidate(...)` matters, because the presenter should not try to produce the first frame before the page is active.

### What was tricky to build
- The main subtlety in the live runner was concurrent stats access: presenter-driven flushes now happen from a presenter goroutine, while stats dumping still happens on the main select loop.
- Another subtle point was not reintroducing the old architecture by accident. The animation loop had to update state and invalidate presentation only; calling `renderAll(...)` directly there would have defeated the whole change.

### What warrants a second pair of eyes
- After the first hardware run, we should review whether additional presenter-specific Go trace events are useful or whether the current flush trace is already enough.
- We should also sanity-check whether the presenter should log dropped/coalesced invalidations explicitly later, though that is not needed for the first validation pass.

### What should be done in the future
- Run the hardware validation and compare the new trace against the old baseline.
- Store those commands under `scripts/` and summarize the before/after rebuilds-per-flush delta in the ticket.
- If the ratio collapses the way we expect, the architecture change can be considered validated and we can decide whether any deeper tuning is still necessary.
