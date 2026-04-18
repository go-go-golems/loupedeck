---
Title: Investigation diary
Ticket: LOUPE-014
Status: active
Topics:
    - loupedeck
    - web
    - simulation
    - javascript
    - jsverbs
    - ui
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: loupedeck/README.md
      Note: README evidence for current hardware-first release surface
    - Path: loupedeck/examples/cmd/loupe-js-demo/main.go
      Note: Evidence that non-hardware DrawTarget rendering already exists
    - Path: loupedeck/pkg/scriptmeta/scriptmeta.go
      Note: Verb scanning and target resolution used to frame the simulator ticket
    - Path: loupedeck/runtime/host/runtime.go
      Note: EventSource boundary that lets a browser feed simulated input into JS callbacks
    - Path: loupedeck/runtime/js/runtime_test.go
      Note: Evidence that semantic input simulation already works without hardware
    - Path: loupedeck/runtime/present/runtime.go
      Note: Confirmed render-before-flush ordering during the second review pass
    - Path: loupedeck/runtime/render/visual_runtime.go
      Note: DrawTarget abstraction that makes a virtual framebuffer practical
    - Path: loupedeck/runtime/ui/ui.go
      Note: Retained UI model and dirty-state behavior that the simulator will reuse
ExternalSources: []
Summary: Chronological investigation record for the virtual Loupedeck browser simulator ticket, including architecture mapping, evidence collection, and design decisions.
LastUpdated: 2026-04-18T13:59:03-04:00
WhatFor: Give a new engineer the exact reading trail and decision context behind the virtual Loupedeck simulator design.
WhenToUse: Read when resuming this ticket, validating why the design prefers a semantic simulator over serial emulation, or onboarding someone to the current architecture.
---



# Investigation diary

## Goal

Capture the investigation that led to the new virtual Loupedeck simulator ticket. The main objective was to understand how much of the current runtime stack can be reused for a browser-driven simulator and where the remaining hardware-only coupling lives.

This diary is meant to help a new intern orient themselves quickly: what the repo already does well, where the serial dependency still exists, why the browser UI should sit on top of the runtime instead of replacing it, and what the next implementation steps should focus on.

## Step 1: Map the current hardware/runtime boundary and frame the simulator ticket

I started by reading the top-level repository README, the CLI entrypoint, the device package, the retained UI/runtime packages, the script metadata helpers, and the annotated JS verb examples. That gave me a concrete view of the system and confirmed that the simulator can reuse most of the runtime stack without touching the JS authoring model.

The key conclusion from the investigation is that the current codebase already has the right semantic seams for a virtual device: `runtime/host.EventSource` for input, `runtime/render.DrawTarget` for output, `runtime/ui.UI` for retained state, and `runtime/js` for the JS API surface. The missing piece is a device backend that is not tied to serial transport. That is why the new ticket is framed as a **virtual Loupedeck browser simulator** instead of a “web frontend for the serial stack.”

**Commit (code):** N/A — research and ticket framing only

### Prompt Context

**User prompt (verbatim):** "Analyze @loupedeck/ and create a new ticket where I want to have a virtual loupedeck instead of relying on the serial interface, and then use that virtual loupedeck to provide a web UI to simulate the loupedeck and thus make it easier to prototype and demo JS verbs.

reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file

references.
It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Build a new research ticket for a browser-based virtual Loupedeck simulator, write a detailed intern-friendly architecture/design guide, and deliver it inside the ticket workspace and reMarkable bundle.

**Inferred user intent:** Make JS verb prototyping and demos independent of the physical Loupedeck hardware by adding a faithful virtual simulator with a web UI, while documenting the current architecture and the implementation plan in a way that a new engineer can follow.

### What I did
- Ran `docmgr status --summary-only` to confirm the ticket workspace root and the current workspace state.
- Read `loupedeck/README.md` to confirm the current release-facing surface, hardware focus, and current CLI assumptions.
- Read `cmd/loupedeck/main.go`, `cmd/loupedeck/cmds/run/command.go`, and `cmd/loupedeck/cmds/verbs/command.go` to map the current command tree.
- Read `pkg/device/connect.go`, `pkg/device/dialer.go`, `pkg/device/display.go`, `pkg/device/listen.go`, `pkg/device/listeners.go`, `pkg/device/profile.go`, and `pkg/device/message.go` to understand the serial/websocket hardware path.
- Read `runtime/js/env/env.go`, `runtime/js/registrar.go`, `runtime/js/module_ui/module.go`, `runtime/js/module_state/module.go`, `runtime/js/module_anim/module.go`, `runtime/js/module_present/module.go`, `runtime/ui/ui.go`, `runtime/ui/page.go`, `runtime/ui/display.go`, `runtime/ui/tile.go`, `runtime/host/runtime.go`, `runtime/host/pages.go`, and `runtime/render/visual_runtime.go` to identify the reusable runtime seams.
- Read `pkg/scriptmeta/scriptmeta.go` and `examples/js/12-documented-scene.js` to confirm how annotated verbs and verb values are currently discovered and executed.
- Created the new ticket workspace with `docmgr ticket create-ticket --ticket LOUPE-014 --title "Virtual Loupedeck browser simulator for JS verbs prototyping and demos" --topics loupedeck,web,simulation,javascript,jsverbs,ui`.
- Created the design doc and diary reference doc with `docmgr doc add`.

### Why
- The simulator should be built on top of the semantic runtime, not on top of raw serial bytes.
- The existing runtime layers already expose the exact concepts the browser needs: pages, tiles, events, and retained rendering.
- The ticket had to be created first so the analysis could be stored in the correct docmgr workspace instead of a loose note.

### What worked
- The repository structure was easy to map once I focused on the device/runtime boundary.
- `runtime/host.EventSource` and `runtime/render.DrawTarget` turned out to be the cleanest abstraction seams.
- `pkg/scriptmeta` already provides the script root and verb discovery helpers that a simulator command can reuse.
- `docmgr ticket create-ticket` and `docmgr doc add` succeeded on the first try and created the expected workspace skeleton.

### What didn't work
- I searched for an existing virtual device or browser simulator implementation and did not find one in `loupedeck/`.
- `rg -n "virtual loupedeck|simulator|web ui|serial interface|serial-WebSocket|Loupedeck Live|demo JS verbs|JS verbs|verbs" loupedeck -S` mostly turned up the existing jsverbs design notes and hardware-specific docs, not an implementation to extend.
- That absence is not a blocker, but it confirmed that the simulator needs to be designed from the current runtime seams rather than adapted from a preexisting web backend.

### What I learned
- The repository already separates *meaningful* state from transport state more cleanly than it first appears.
- The real hardware path is still concentrated in a few files, while the retained UI and JS modules are reusable across both hardware and simulation.
- The browser simulator should act like a new device backend, not like a replacement JS runtime.

### What was tricky to build
- The tricky part was deciding where the simulator should sit. It would have been easy to aim at the serial protocol, but that would reproduce the least useful part of the stack for prototyping.
- The correct seam is higher up: use retained state and event-source abstractions so the browser can interact with the same runtime model as the real device.

### What warrants a second pair of eyes
- Whether the first simulator API should expose only Live, or whether the ticket should start with a profile selector for Live / Live S / CT.
- Whether full-frame PNG transfer over websocket is the right first transport, or whether a tiled diff protocol should be planned immediately.
- Whether the simulator should be named `simulate`, `web`, or `virtual` in the final CLI.

### What should be done in the future
- Add a simulator backend package that implements the semantic device boundary.
- Add a browser UI and websocket protocol.
- Wire the simulator into the existing script/verb loading flow.
- Add integration tests that boot a representative scene without a physical device.

### Code review instructions
- Start by reading `loupedeck/README.md` and `cmd/loupedeck/main.go` to understand the current product shape.
- Then read `pkg/device/connect.go`, `pkg/device/display.go`, `runtime/host/runtime.go`, `runtime/ui/ui.go`, `runtime/render/visual_runtime.go`, and `runtime/js/module_ui/module.go` to see the semantic seams the simulator should reuse.
- Use `pkg/scriptmeta/scriptmeta.go` and `examples/js/12-documented-scene.js` as the loading/verb reference for the browser flow.

### Technical details
- Key device values come from `pkg/device/profile.go`, which already defines the Live and CT display layouts.
- `runtime/js/env.Ensure` already wires `UI` dirty events into `Present.Invalidate("ui-dirty")`, so a simulator can wake the renderer automatically when JS changes state.
- `runtime/render.Renderer` already expects a `DrawTarget`, which is the natural hook for a virtual framebuffer or browser mirror.

## Step 2: Review the first design pass and write a revised implementation plan

After the first design doc existed, I switched roles from “initial mapper” to “independent reviewer.” The goal of this step was not to re-summarize the same material, but to stress-test the first proposal against the actual code and determine whether it was the shortest path to a working simulator.

This second pass changed the recommended implementation shape in an important way. I kept the big idea — simulate semantics, not the serial wire — but rejected the first design's tendency to create a broad virtual backend abstraction too early. The revised plan is more incremental: reuse the current script bootstrap, add a dedicated `simulate` command, and implement a simulator out of smaller pieces (`InputHub`, `FrameCompositor`, `BrowserHub`) instead of a fake `pkg/device.Loupedeck` clone.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Produce a second, independent design review and revised implementation plan rather than extending the first design doc as if it were already accepted.

**Inferred user intent:** Get a real second opinion on the architecture so the ticket does not converge prematurely on the first plausible implementation shape.

**Commit (code):** N/A — review and documentation only

### What I did
- Re-read the first design doc closely and compared it against the repository rather than accepting its proposed abstractions at face value.
- Read the deeper `runSceneSession()` implementation in `cmd/loupedeck/cmds/run/command.go` to identify the true extraction point between hardware-only and reusable runtime logic.
- Re-read `runtime/render/visual_runtime.go` and noticed that `Flush()` emits both full-display draws and partial tile patch draws, which means a simulator must composite into retained display buffers rather than store one PNG blob per display revision.
- Re-read `runtime/present/runtime.go` and confirmed that `RenderFunc` runs before `FlushFunc`, which means browser notifications should be published after flush, not during render.
- Re-read `runtime/js/runtime_test.go`, `runtime/js/jsverbs_integration_test.go`, and `examples/cmd/loupe-js-demo/main.go` as evidence that input simulation and non-hardware rendering already exist separately in the codebase.
- Created a second design doc: `design-doc/02-independent-review-and-revised-implementation-plan-for-the-virtual-loupedeck-simulator.md`.

### Why
- The first design was strong on direction but still too broad on implementation shape.
- I wanted the second proposal to minimize package-boundary churn, avoid premature abstractions, and preserve the meaning of `pkg/device` as the hardware/protocol package.
- A review-only step is valuable because it can shrink scope before code exists.

### What worked
- The current code gave much stronger evidence for a composition-based simulator than for a monolithic virtual backend.
- `runtime/js/runtime_test.go` and `examples/cmd/loupe-js-demo/main.go` were especially useful because together they prove the simulator can be assembled from existing semantic seams.
- The deeper read of `runSceneSession()` made the actual refactor point obvious.

### What didn't work
- The first design's draft `Backend` abstraction did not line up cleanly with how the runtime currently separates inputs (`host.EventSource`) from outputs (`render.DrawTarget`).
- The earlier pseudocode's idea of browser-side notification during the render callback did not survive inspection of `present.Runtime.loop()`, because render happens before flush.
- No repository code suggested that `pkg/device` should own browser simulator concerns; that package still reads as hardware/protocol-first.

### What I learned
- The hardest part of review work is distinguishing “good direction” from “good first implementation.” The first design had the former, but not yet the latter.
- `render.Renderer.Flush()` patch behavior is not a cosmetic detail; it drives the entire frame-store design for the simulator.
- Browser-side forms and command discovery are useful, but they should be phase-2 ergonomics rather than phase-1 architecture drivers.

### What was tricky to build
- The tricky part in this step was avoiding a lazy review that only tweaks wording. A useful second opinion needed to either validate the first design concretely or replace parts of it with a lower-risk plan.
- The concrete technical hinge was realizing that the frame-store problem is not “save the last PNG,” but “compose patches into retained full-display buffers.” Once that clicked, the smaller simulator architecture became much clearer.

### What warrants a second pair of eyes
- Whether `runtime/sim` is the best package home, or whether the simulator should stay CLI-internal until a second consumer exists.
- Whether the hybrid HTTP + websocket transport is preferable to pushing all frame data over websocket in the first pass.
- Whether the simulator command should reuse enough of `run` to justify an extracted shared session helper, or whether shared bootstrap helpers are sufficient for phase 1.

### What should be done in the future
- Implement the simulator in the revised order from the second design doc: shared bootstrap helpers first, then input hub, then frame compositor, then browser bridge, then command wiring.
- Validate the compositor carefully against the renderer's patch behavior before spending time on browser polish.
- Add a dedicated integration test that boots a real scene into the simulator path before expanding browser UX.

### Code review instructions
- Start with the second design doc and compare it to the first design doc side-by-side.
- Then read these repository files in this order:
  1. `cmd/loupedeck/cmds/run/command.go`
  2. `runtime/render/visual_runtime.go`
  3. `runtime/present/runtime.go`
  4. `runtime/js/runtime_test.go`
  5. `examples/cmd/loupe-js-demo/main.go`
- Validate whether the revised proposal's smaller seams really match current code shape better than the earlier generic backend idea.

### Technical details
- `runSceneSession()` currently combines hardware acquisition, event attachment, runtime startup, renderer construction, and presenter management in one hardware-shaped function.
- `present.Runtime.loop()` runs render before flush, so any browser “frame ready” signal must be published after the compositor is updated.
- `render.Renderer.Flush()` sends both full-display and partial tile draws, so the simulator's display store must support compositing.

## What I'd do differently next time

I would start by sketching the message protocol and snapshot shape before reading more implementation files. That would make it easier to reason about the browser UI and the backend seam in a single pass instead of discovering each piece separately.

## Related files

- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/README.md: repository status and hardware focus`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/main.go: root CLI wiring`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/pkg/device/connect.go: real device connection path`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/pkg/device/display.go: protocol-level draw path`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/runtime/host/runtime.go: semantic input-source boundary`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/runtime/ui/ui.go: retained page/tile state`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/runtime/render/visual_runtime.go: retained renderer and DrawTarget`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/runtime/js/module_ui/module.go: JS UI API surface`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/pkg/scriptmeta/scriptmeta.go: script and verb loading helpers`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/examples/js/12-documented-scene.js: annotated verb example`
