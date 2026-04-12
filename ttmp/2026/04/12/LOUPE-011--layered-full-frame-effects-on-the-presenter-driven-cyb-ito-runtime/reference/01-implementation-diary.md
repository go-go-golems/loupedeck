---
Title: Implementation diary
Ticket: LOUPE-011
Status: active
Topics:
    - javascript
    - rendering
    - animation
    - performance
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../../tmp/loupe-cyb-ito-layered-011-success-1776029307.log
      Note: Hardware evidence log for the first successful layered full-page smoke run
    - Path: examples/js/10-cyb-ito-full-page-all12.js
      Note: Primary code file changed in Step 2
    - Path: ttmp/2026/04/12/LOUPE-011--layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime/scripts/03-run-layered-scene-fluid.sh
      Note: Reproducible hardware run command used for the fluid layered validation
ExternalSources: []
Summary: Chronological diary for layered full-frame effects on the presenter-driven cyb-ito runtime.
LastUpdated: 2026-04-12T17:20:07.233054427-04:00
WhatFor: Use this diary to understand what was changed, why, what worked, what failed, and how to review the layered full-frame scene work.
WhenToUse: Use when continuing, reviewing, or validating the LOUPE-011 layered full-frame effects implementation.
---


# Diary

## Goal

Capture the step-by-step implementation of layered full-frame effects for the presenter-driven cyb-ito runtime, including planning, code changes, validation commands, and review guidance.

## Step 1: Create the ticket, write the implementation plan, and define the first implementation slice

After confirming that the presenter-driven full-page runtime was smooth enough on hardware, the next architectural question was whether we could reintroduce richer animated layers without bringing back the old rebuild storm. The answer was yes, but only if the layers stayed inside software composition and still collapsed to one final full-page hardware flush.

This first step creates a new ticket specifically for that work instead of overloading `LOUPE-010`. It also defines the implementation plan and the phased task list before changing any code so the next steps stay narrow and reviewable.

### Prompt Context

**User prompt (verbatim):** "alright, let's ad d a new ticket and an implementation plan and a detailed list of tasks, and then work on the task, commit at appropriate intervals, and keep a detailed frequent diary."

**Assistant interpretation:** Create a dedicated ticket for the layered full-frame effects work, write the design/implementation plan and a real task breakdown, then start executing the first implementation slice with disciplined commits and diary maintenance.

**Inferred user intent:** The user wants the smooth presenter-driven full-page runtime turned into the new stable foundation for richer layered effects, with the same disciplined ticket-and-diary workflow used in the other Loupedeck investigations.

### What I did
- Created ticket `LOUPE-011` with `docmgr ticket create-ticket`.
- Added a dedicated design document for the layered full-frame effects architecture.
- Added this diary document.
- Planned the first implementation slice as: internal software layers + one final presenter-driven full-page flush.

### Why
- The architectural premise is different enough from `LOUPE-010` that it deserves its own ticket.
- The work needs a forward-only implementation plan so we do not drift back into the old “many rebuilds per visible frame” model.
- The first code slice should be small and hardware-visible rather than prematurely over-generalized.

### What worked
- `docmgr` ticket creation succeeded cleanly.
- The first-pass architecture is simple to explain: multiple internal `gfx.surface(...)` layers, one final composed frame, one hardware flush.

### What didn't work
- N/A

### What I learned
- The biggest mental trap here is to think “layers again” means “multiple output writes again.” It does not. The correct version is many software layers, one final output frame.

### What was tricky to build
- The main tricky part at the planning stage was preserving the right constraint language. The smooth run in `LOUPE-010` makes it tempting to say “we can do anything now,” but the real invariant is narrower: we can do richer per-frame composition now because the presenter owns frame production.

### What warrants a second pair of eyes
- The plan should be reviewed specifically for whether any later task accidentally reintroduces per-effect output flushes instead of software-only layer composition.

### What should be done in the future
- Implement the first code slice: refactor the full-page scene into internal layers and add the first software-composited effect pass.

### Code review instructions
- Start with the design doc in `design/01-implementation-plan-for-layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime.md`.
- Confirm that the plan preserves the presenter-driven full-frame model from `LOUPE-010`.

### Technical details
- Ticket path: `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-011--layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime/`
- Source scene to be modified next: `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`

## Step 2: Implement the first layered full-page compositor slice and validate it on hardware

The first code slice took the existing presenter-driven full-page scene and turned it into the layered software compositor described in the plan. The key constraint stayed intact: the device still sees one full-page frame flush, but that frame is now built from multiple logical internal surfaces.

This step also includes the first real hardware smoke validation for the layered version. I wanted actual evidence that the richer layered frame composition did not immediately destroy the smoothness we had just recovered in `LOUPE-010`.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Execute the first implementation slice after writing the plan and task list, then document it with disciplined commits and diary updates.

**Inferred user intent:** Move from architecture talk to a real layered scene implementation while keeping the same evidence-backed, ticketed workflow.

**Commit (code):** `4b44402` — `Add layered full-frame compositor to cyb-ito scene`

### What I did
- Reworked `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js` so it no longer draws everything directly into one surface.
- Added internal software layers:
  - `baseLayer`
  - `chromeLayer`
  - `sceneLayer`
  - `fxLayer`
  - `hudLayer`
  - final `frame`
- Split the scene build into per-layer render functions:
  - `rebuildBaseLayer()`
  - `renderChromeLayer(...)`
  - `renderSceneLayer(...)`
  - `renderFXLayer(...)`
  - `renderHUDLayer(...)`
  - `composeFrame()`
- Added the first FX pass:
  - scanlines
  - sparse grain/noise
  - active-tile sweep
  - active-tile ripple
- Kept the current presenter-driven frame model intact:
  - `present.onFrame(...)` still owns `renderAll(...)`
  - the UI still attaches one final full-page surface: `display.surface(frame)`
- Ran `go test ./...`
- Ran hardware validation with aggressive pacing and captured a log:
  - `/tmp/loupe-cyb-ito-layered-011-success-1776029307.log`
- Archived the commands in ticket-local scripts under `scripts/`.

### Why
- The user explicitly wanted to bring back layered frame effects now that the presenter-driven runtime had proven smooth enough on hardware.
- The cleanest first step was to add software layers inside the scene script before touching the Go runtime again.
- This keeps the architecture honest: richer frame composition without regressing into multi-flush or multi-rebuild behavior.

### What worked
- The JS example booted and passed the full Go test suite after the refactor.
- The new layered version still ran on hardware with `--send-interval 0ms`.
- The first hardware log showed the scene running with render windows still around the low-single-digit millisecond range despite the extra compositing work.
- The user confirmed from device observation that the layered version worked.

### What didn't work
- The first attempt to start the layered scene in tmux failed with the familiar reconnect/handshake problems:
  - `connect: malformed HTTP response "\x82\x05\x05\x01\x00\x01\xff..."`
  - `connect: Port has been closed`
- A later short evidence-capture run also hit the familiar first-attempt reconnect issue before succeeding on retry:
  - `WARN dial failed err="malformed HTTP response \"\\x82\\t\\tM...\""`
- These failures look like the same known device lifecycle fragility, not a new layered-scene bug.

### What I learned
- The layered model is viable without additional Go runtime work for the first slice.
- A moderate amount of extra full-frame composition is still cheap enough compared with the current fluid hardware run configuration.
- The right first layering move was inside the scene script, not another large renderer refactor.

### What was tricky to build
- The main tricky part was refactoring the tile renderers so they no longer implicitly drew into one global surface. I solved that by threading the destination surface through the tile-art and helper functions instead of trying to invent a second scene architecture in parallel.
- The second tricky part was avoiding overengineering the first FX pass. I kept it intentionally simple and cheap: sparse noise, lightweight scanlines, and one active-tile overlay instead of a full heavy post-processing stack.
- The third tricky part was separating layered-scene behavior from the known serial-websocket reconnect instability. The failed first hardware attempts looked scary in the logs, but the subsequent successful run showed the scene architecture itself was fine.

### What warrants a second pair of eyes
- The JS scene now does more per-frame work, so someone reviewing should check whether any of the full-frame loops should be cached or thinned further before we stack on more effects.
- The bottom-of-run clear path still triggers extra left/right display draws; that is not wrong, but it is worth noticing during review.
- The reconnect noise remains a system-level hazard around validation runs.

### What should be done in the future
- Implement the next requested visual slice on top of this checkpoint instead of mixing it into the same commit.
- Consider whether selected-tile coloring requires a runtime/rendering extension rather than just a script change.
- Add larger full-screen ripple choreography in the FX layer if the user still wants more dramatic touch feedback.

### Code review instructions
- Start in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`.
- Review the new layer flow in this order:
  - layer declarations
  - tile-art functions now parameterized by destination surface
  - `renderChromeLayer(...)`
  - `renderSceneLayer(...)`
  - `renderFXLayer(...)`
  - `composeFrame()`
  - `renderAll(...)`
- Validate with:
  - `go test ./...`
  - `go run ./cmd/loupe-js-live --script ./examples/js/10-cyb-ito-full-page-all12.js --duration 4s --send-interval 0ms --stats-interval 2s --log-render-stats --log-writer-stats`

### Technical details
- Evidence log: `/tmp/loupe-cyb-ito-layered-011-success-1776029307.log`
- Ticket-local scripts:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-011--layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime/scripts/01-create-ticket.sh`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-011--layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime/scripts/02-go-test-layered-scene.sh`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-011--layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime/scripts/03-run-layered-scene-fluid.sh`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-011--layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime/scripts/04-docmgr-doctor.sh`
