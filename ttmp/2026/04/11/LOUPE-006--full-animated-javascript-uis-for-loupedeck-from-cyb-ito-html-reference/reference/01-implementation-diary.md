---
Title: Implementation diary
Ticket: LOUPE-006
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - animation
    - rendering
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/design/01-textbook-full-animated-javascript-ui-runtime-from-cyb-ito-reference.md
      Note: Main intern-facing design and implementation guide for this ticket
    - Path: ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html
      Note: Imported reference artifact analyzed in this ticket
ExternalSources:
    - local:cyb-ito.html
Summary: Chronological diary for the cyb-ito-inspired full animated JS UI ticket, recording ticket creation, source import, source analysis, and the first detailed design package.
LastUpdated: 2026-04-11T23:31:00-04:00
WhatFor: Preserve the exact reasoning and workflow behind turning the imported cyb-ito HTML canvas reference into a detailed runtime-expansion plan.
WhenToUse: Use when resuming the ticket or when reviewing how the imported HTML reference was interpreted and translated into a Loupedeck-native implementation plan.
---

# Diary

## Goal

Create a new ticket around the imported `cyb-ito.html` reference, analyze what it actually is, and turn that into a very detailed intern-facing design and implementation guide for building full animated JavaScript UIs on top of the current Loupedeck runtime.

## Step 1: Create the ticket, import the source artifact, and analyze the HTML carefully

The first important action in this ticket was not code. It was documentation and source control discipline. The user specifically asked to create a new ticket, import `~/Downloads/cyb-ito.html` through `docmgr import file`, read it carefully, and then explain how it should be implemented with the current JS runtime.

That mattered because the source artifact should not live as an untracked local file in `Downloads`. It had to become a tracked ticket source so that future work has a stable reference and future readers can see exactly what was analyzed.

### Prompt Context

**User prompt (verbatim):** "ow create a new ticket to implement full animated JS UIs. We're going to use ~/Downloads/cyb-ito.html and import it using docmgr import file --file ~/Downloads/cyb-ito.html --ticket XXXX and read it     
carefully and then tell me how you will implement it with our JS API."

**Assistant interpretation:** Create a new `LOUPE-006` ticket, import the HTML file into ticket sources, analyze the source deeply enough to understand its rendering and interaction model, and produce a concrete architectural implementation direction.

**Inferred user intent:** Use a real imported visual reference as the forcing function for the next expansion of the JS runtime, rather than designing further in the abstract.

**Commit:** `3279292` — `Create LOUPE-006 animated JS UI ticket and import cyb-ito`

### What I did
- Determined that the next ticket ID should be `LOUPE-006`.
- Ran:

```bash
docmgr ticket create-ticket --ticket LOUPE-006 --title "Full animated JavaScript UIs for Loupedeck from cyb-ito HTML reference" --topics loupedeck,goja,javascript,animation,rendering
```

- Imported the reference artifact exactly as requested:

```bash
docmgr import file --file ~/Downloads/cyb-ito.html --ticket LOUPE-006
```

- Confirmed the imported tracked copy now lives at:

```text
ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html
```

- Read the file in multiple passes and extracted the main facts:
  - one self-contained `<canvas>` app
  - one `requestAnimationFrame(render)` loop
  - raw `ImageData` grayscale framebuffer writes
  - `4×3` grid of `90×90` animated tiles
  - left and right animated strips
  - ripple overlays
  - touch-driven tile activation and scan sweeps
- Confirmed that it is **not** a DOM UI, **not** an SVG animation, and **not** something we should literally “run as browser HTML” inside the Loupedeck runtime.

### Why
- Importing the file through `docmgr` makes the source durable and reviewable.
- Careful source reading was necessary because the right implementation strategy depends entirely on what the file actually is. If it had been a DOM/CSS design, the answer would have been very different.
- The correct next-runtime design depends on interpreting the artifact at the right level: not as HTML, but as a procedural animated scene specification.

### What worked
- The imported source is now tracked in the ticket workspace.
- The structure of the file was clear enough to identify its true nature quickly.
- The analysis made the next key decision obvious: the current JS API is not enough by itself, but the current runtime architecture is still the correct base.

### What didn't work
- No tooling failed in this step.
- The only deliberate negative result was architectural: a direct or literal HTML/canvas execution strategy is the wrong path for this repo.

### What I learned
- `cyb-ito.html` is much closer to a retained animated scene system than to the current small retained text/icon tile API.
- The source geometry is already strongly aligned with the Loupedeck main tile grid (`90×90`), but the side strips need adaptation to the real hardware (`60×270` rather than `36×270`).
- The future runtime needs display regions, retained surfaces, layers, and Go-owned raster primitives if it wants to express this class of visual design correctly.

### What should be done in the future
- Write a proper intern-facing design and implementation guide.
- Expand the ticket tasks and changelog.
- Upload the resulting guide to reMarkable.

## Step 2: Write the first detailed textbook-style guide for a new intern

After analyzing the imported source, the next user request narrowed the deliverable very clearly: write a detailed analysis/design/implementation guide for a new intern, store it in the ticket, and upload it to reMarkable.

The right output for that request was not a short plan or a few bullets. It had to be a real textbook-style design package that explained:

- what the imported HTML is,
- how the current runtime works,
- why the current runtime is not yet sufficient,
- what new concepts are needed,
- what the new APIs should probably look like,
- what implementation order is safest,
- and what mistakes to avoid.

### Prompt Context

**User prompt (verbatim):** "reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a substantial textbook-style design doc inside the ticket, update the ticket bookkeeping coherently, and then upload the result to the reMarkable.

**Inferred user intent:** Preserve the design at a depth where a new engineer can start implementation work safely without rediscovering the architectural rules.

### What I did
- Created the main design doc:

```text
design/01-textbook-full-animated-javascript-ui-runtime-from-cyb-ito-reference.md
```

- Wrote it as an intern-facing textbook rather than a terse proposal.
- Included:
  - executive summary
  - problem statement
  - explanation of the current baseline architecture
  - careful explanation of what `cyb-ito.html` actually is
  - analysis of why the current JS API is not enough yet
  - target architecture and new concepts
  - proposed API direction (`display` regions, `gfx`, layers)
  - mapping table from source concepts to runtime concepts
  - phased implementation plan
  - pseudocode and Mermaid diagrams
  - working rules and anti-patterns
- Created this diary file for continuity.

### Why
- The ticket needed to move from “artifact imported” to “artifact interpreted into an actionable runtime plan.”
- The user explicitly asked for something suitable for a new intern, which means the document had to explain not only *what* to build, but *why* each architectural rule exists.
- The ticket docs should preserve architecture decisions, not just immediate coding notes.

### What worked
- The design doc now gives a full architectural path from the imported HTML to a Loupedeck-native retained animated scene runtime.
- The doc clearly distinguishes between what must stay in Go and what should be exposed to JavaScript.
- The proposed next slice (`display` regions + `gfx` module + layered surfaces) is concrete enough to guide implementation without forcing every future API detail prematurely.

### What didn't work
- The repository still does not yet implement the new `gfx` or display-region model; this step is planning and specification, not implementation.
- The design necessarily proposes API sketches rather than final stabilized signatures.

### What I learned
- The imported artifact is an excellent forcing function because it reveals exactly where the current runtime stops being expressive enough.
- The existing `state` and `anim` APIs are still the right state/time layer. What really needs to grow is the retained visual scene layer.
- The most important architectural risk remains the same as before: do not let JavaScript own low-level rendering or transport.

### What should be done in the future
- Update ticket index/tasks/changelog to reflect the new design package.
- Upload the design bundle to reMarkable.
- Begin the first implementation slice with display regions rather than trying to jump straight to a full cyb-ito port.

## Step 3: Validate the ticket and upload the design bundle to reMarkable

Once the design package existed in the ticket, the final operational work was to validate the docs and deliver them to the reMarkable in the same way the earlier Loupedeck ticket bundles were delivered.

### What I did
- Ran:

```bash
docmgr doctor --ticket LOUPE-006 --stale-after 30
```

- Result: all checks passed.
- Checked reMarkable tooling with:

```bash
remarquee status
```

- Performed a dry-run bundled upload using the ticket index, main design guide, and diary.
- Uploaded the real bundle to:

```text
/ai/2026/04/11/LOUPE-006
```

with document name:

```text
LOUPE-006 Full animated JavaScript UIs from cyb-ito reference
```

- Verified the remote folder with:

```bash
remarquee cloud ls "/ai/2026/04/11/LOUPE-006" --long --non-interactive
```

### What worked
- `docmgr doctor` passed cleanly.
- The bundle upload succeeded.
- The uploaded document was verified remotely under the expected LOUPE-006 folder.

### Delivery result
- reMarkable remote dir: `/ai/2026/04/11/LOUPE-006`
- uploaded document: `LOUPE-006 Full animated JavaScript UIs from cyb-ito reference`

## Step 4: Implement the first retained display-region slice

With the design package in place, I started the implementation at the exact point the ticket recommended: not with `gfx`, not with pixel drawing, and not with a full cyb-ito port, but with retained display-region groundwork.

The goal of this slice was to teach the runtime that a page can own more than the main `4×3` tile grid. Specifically, it needed to gain first-class retained regions for:

- `left`
- `main`
- `right`

without breaking the already-working `page.tile(...)` model used by the existing examples and tests.

### Prompt Context

**User prompt (verbatim):** "add detailed tasks to the ticket if you haven't yet, then go through them one by one, comitting at appropriate intervals, and keep a diary"

**Assistant interpretation:** Expand the ticket into implementation-sized subtasks, then begin executing them in focused slices with commits and diary updates instead of leaving the work at the design-only stage.

**Inferred user intent:** Turn LOUPE-006 from a planning ticket into an active implementation track with clear milestone boundaries and continuity.

**Commit (planning):** `3ab0609` — `Expand LOUPE-006 implementation tasks`

**Commit (code):** `051a62d` — `Add retained JS display regions`

### What I did
- Expanded `tasks.md` into detailed Phase B/C/D/E/F/G/H implementation checkpoints.
- Added a new retained display type:

```text
runtime/ui/display.go
```

with:
- `DisplayLeft`, `DisplayMain`, `DisplayRight`
- retained text/icon/visible state
- reactive bindings for those properties
- main-display tile ownership via `display.AddTile(...)`
- display dirty tracking
- Updated:

```text
runtime/ui/page.go
```

so pages now own named displays rather than only a flat tile map.
- Preserved compatibility by keeping:

```text
page.AddTile(...)
page.Tile(...)
page.Tiles(...)
```

but making them delegate to the retained `main` display.
- Updated:

```text
runtime/ui/ui.go
```

so the UI runtime now tracks both:
- dirty tiles
- dirty displays

and filters both against the active page.
- Updated:

```text
runtime/ui/tile.go
```

so tiles now hang off a retained display instead of a page-level tile map.
- Extended the JS UI module in:

```text
runtime/js/module_ui/module.go
```

with:
- `page.display(name, fn)`
- display-level `text(...)`, `icon(...)`, `visible(...)`
- main-display `display.tile(...)`
- Extended the retained renderer bridge in:

```text
runtime/render/visual_runtime.go
```

so it can flush side-display placeholders as full retained display images while still flushing main-grid tiles independently.
- Updated the live runner in:

```text
cmd/loupe-js-live/main.go
```

so it now:
- acquires `left`, `main`, and `right` hardware displays
- flushes all retained display targets
- clears all three on exit rather than only `main`
- Added tests in:

```text
runtime/ui/ui_test.go
runtime/render/render_test.go
runtime/js/runtime_test.go
```

covering:
- retained display dirty filtering
- side-display rendering
- JS `page.display("left", ...)` integration
- Ran:

```bash
gofmt -w runtime/ui/*.go runtime/render/*.go runtime/js/module_ui/module.go runtime/js/runtime_test.go cmd/loupe-js-live/main.go
go test ./...
```

and all tests passed.

### Why
- The design guide said display regions must come before `gfx`, and that turned out to be correct in practice.
- Without this slice, the future `loupedeck/gfx` module would still have nowhere structurally appropriate to attach its retained surfaces.
- Preserving `page.tile(...)` compatibility was important because the current JS runtime examples are still valuable and should not be broken just to add scene generality.

### What worked
- The retained model now has a correct structural place for left/right display content.
- The existing tile API still works.
- The renderer can now flush side displays through the same retained-to-Go-owned-render path.
- The live runner is no longer hardcoded to only the main display.
- The new tests proved the slice without requiring hardware yet.

### What didn't work
- This slice still uses placeholder display rendering for side displays. That is expected: the point here was structural groundwork, not final cyb-ito visuals.
- There is still no `loupedeck/gfx` module yet, so JS cannot draw procedural graphics into those display regions yet.

### What I learned
- The smallest correct next step really was structural. The retained scene model needed to understand multiple displays before any serious visual API work could make sense.
- Preserving the existing main-tile API as a delegation layer is a clean compatibility strategy.
- The live runner was easier to generalize than expected once the retained renderer accepted a target map instead of a single main-display target.

### What should be done in the future
- Update the ticket docs and mark the Phase B slice tasks complete.
- Start Phase C: add the pure-Go retained graphics/surface package.
- Keep Phase D (`loupedeck/gfx`) separate from Phase C so the pure-Go semantics can be tested before the goja binding layer is added.

## Step 5: Add the pure-Go retained graphics package

After the retained display-region slice was in place, the next correct step was the one the design guide had recommended from the start: build the graphics semantics in pure Go before exposing them to goja.

The purpose of this slice was not to finish the cyb-ito port. It was to establish the first real graphics substrate that future JS scene code can target without dropping down to pixels or transport writes.

### Commit

**Commit (code):** `33a278b` — `Add retained graphics surface package`

### What I did
- Added a new package:

```text
runtime/gfx/
```

- Added:

```text
runtime/gfx/surface.go
runtime/gfx/text.go
runtime/gfx/surface_test.go
```

- Implemented a retained grayscale/additive `Surface` with:
  - `NewSurface(width, height)`
  - `Clear(v)`
  - `Set(x, y, v)`
  - `Add(x, y, v)`
  - `FillRect(...)`
  - `Line(...)`
  - `Crosshatch(...)`
  - `CompositeAdd(...)`
  - `ToRGBA(...)`
- Added a `TextOptions` model and a `Surface.Text(...)` helper that rasterizes text in Go using `basicfont.Face7x13`.
- Added focused tests covering:
  - surface clear and saturating add
  - line endpoint drawing
  - crosshatch marking
  - additive composition
  - text drawing producing visible pixels
- Ran:

```bash
gofmt -w runtime/gfx/*.go
go test ./...
```

### Why
- This slice establishes graphics semantics in a testable pure-Go package before goja bindings enter the picture.
- It proves that the runtime can have Go-owned retained surfaces without yet deciding the final JS syntax for every future scene primitive.
- It keeps the project honest about ownership: even though the future animated scene is script-facing, the actual raster substrate remains Go-owned.

### What worked
- The new package is small but already meaningful.
- The surface operations line up with the first useful subset of the imported cyb-ito reference: fill, lines, crosshatching, text, additive composition.
- The tests are focused and green, which means the surface semantics can now evolve independently of JS integration.

### What didn't work
- This slice intentionally does not yet expose graphics to JS.
- It also does not yet implement every possible future primitive from the imported HTML, such as spirals or drip effects.

### What I learned
- The graphics package needed to exist before the JS module. That sequencing still feels correct in practice.
- Even a very small retained surface model already clarifies the next runtime shape substantially.

## Step 6: Add the first JS-facing `loupedeck/gfx` module

With the pure-Go graphics semantics in place, the next slice was the JS binding layer. This step intentionally exposed the new surfaces to JS in a **coarse, retained, surface-oriented** form, not as a raw immediate-mode pixel API.

### Commit

**Commit (code):** `b4f343e` — `Add JS graphics surface module`

### What I did
- Added:

```text
runtime/js/module_gfx/module.go
```

- Registered the module from:

```text
runtime/js/runtime.go
```

under:

```text
require("loupedeck/gfx")
```

- Exposed a first JS-facing surface object with methods such as:
  - `width()`
  - `height()`
  - `clear(...)`
  - `fillRect(...)`
  - `line(...)`
  - `crosshatch(...)`
  - `text(...)`
  - `compositeAdd(...)`
  - `at(...)`
- Added JS integration coverage in:

```text
runtime/js/runtime_test.go
```

that:
- constructs a surface from JS
- draws lines, crosshatching, text, and an overlay surface
- composites the overlay
- samples brightness via `at(...)`
- verifies width reporting
- Ran:

```bash
gofmt -w runtime/js/module_gfx/*.go runtime/js/runtime.go runtime/js/runtime_test.go
go test ./...
```

### Why
- The purpose of this slice was to give JS a real graphics surface to target before layer composition or a full scene demo exists.
- Exposing surfaces in a retained way preserves the architecture: JS can describe graphics work, but it is still interacting with Go-owned surfaces rather than transport or framebuffer protocol messages.

### What worked
- The first `loupedeck/gfx` module is now real and tested.
- The API is coarse and surface-oriented rather than pixel-loop-oriented, which matches the design intent.
- The module registration dropped into the owned runtime cleanly without new owner-thread complexity because the operations are synchronous surface mutations.

### What didn't work
- This slice still does not connect surfaces to display regions. That is the next layer: retained surface/layer composition.
- The current module exposes an `at(...)` helper primarily for testing and debugging. That is fine for now, but it should not become the center of the public authoring model.

### What I learned
- Once `runtime/gfx` existed, the JS binding layer was straightforward.
- The runtime now has the first real substrate for cyb-ito-style scene authoring: JS can build retained surfaces, but the system still does not give JS any transport escape hatch.

### What should be done in the future
- Update the ticket docs and mark the Phase C and Phase D slice tasks complete.
- Start Phase E: retained surface/layer composition so `gfx` surfaces can become actual display content.

## Step 7: Bridge retained `gfx` surfaces into retained displays

After the pure-Go graphics package and the first JS-facing `loupedeck/gfx` module existed, the next missing piece was obvious: a `gfx` surface could be created and mutated from JS, but it still could not become actual display content. The retained display model needed to be able to own one of those surfaces and become dirty whenever the surface changed.

This slice is the first real composition bridge between the new graphics substrate and the retained UI model.

### Commit

**Commit (code):** `0eff01c` — `Add retained display-owned graphics surfaces`

### What I did
- Added mutation notifications to `runtime/gfx/surface.go` via surface subscriptions so a retained surface can notify an owner when coarse drawing operations change the surface.
- Updated `runtime/gfx/text.go` so text drawing triggers the same retained-surface change notification path.
- Extended `runtime/ui/display.go` so a retained display can now:
  - own a `*gfx.Surface`
  - subscribe to that surface’s change notifications
  - mark itself dirty whenever the surface changes
- Updated `runtime/render/visual_runtime.go` so retained display rendering now prefers the attached `gfx` surface when one exists.
- Exported a surface unwrapping helper from:

```text
runtime/js/module_gfx/module.go
```

so other JS modules can safely consume `loupedeck/gfx` surface objects.
- Added:

```text
display.surface(surface)
```

to:

```text
runtime/js/module_ui/module.go
```

so JS can attach a retained graphics surface directly to a display region.
- Added tests in:

```text
runtime/ui/ui_test.go
runtime/render/render_test.go
runtime/js/runtime_test.go
```

covering:
- display dirty propagation after surface mutation
- retained display-surface rendering
- JS integration for attaching a `loupedeck/gfx` surface to a display
- Ran:

```bash
gofmt -w runtime/gfx/*.go runtime/ui/*.go runtime/render/*.go runtime/js/module_gfx/*.go runtime/js/module_ui/module.go runtime/js/runtime_test.go
go test ./...
```

### Why
- Without this bridge, `loupedeck/gfx` would remain a disconnected sandbox rather than part of the retained UI system.
- This is the first point where JS-authored graphics content becomes actual retained display output while still preserving Go-owned rendering and invalidation behavior.

### What worked
- Displays now become dirty automatically when attached surfaces mutate.
- The renderer now has a direct path from retained graphics surfaces to actual display images.
- The JS API now has a meaningful bridge point: `display.surface(surface)`.

### What didn't work
- This is still only the first composition step. It does not yet provide true multi-layer overlays or a stable ordering model for several surfaces per display.
- The current bridge is one-surface-per-display rather than a full scene-layer stack.

### What I learned
- The retained display model was already a good place to hang graphics content once surface change notifications existed.
- The API shape remains consistent with the original design rule: JS owns retained scene state, and Go owns realization and transport.

### What should be done in the future
- Update the ticket docs and mark the first retained-surface composition slice tasks complete.
- Continue Phase E toward true multi-layer display composition.
- Then move on to the first cyb-ito-inspired demo scene once the composition model is rich enough.

## Step 8: Add the first cyb-ito-inspired multi-display JS prototype scene

Once displays could own retained surfaces and JS could mutate those surfaces through `loupedeck/gfx`, the next natural milestone was a real prototype scene in `examples/js/` rather than more abstract infrastructure work.

The goal here was not a perfect final art port. The goal was to prove that the runtime could already express a recognizably cyb-ito-inspired multi-display scene using the retained-region and retained-surface model now in place.

### Commit

**Commit (code):** `605e89d` — `Add cyb-ito JS prototype scene`

### What I did
- Added:

```text
examples/js/07-cyb-ito-prototype.js
```

- The script now:
  - creates retained `gfx` surfaces for:
    - `main` (`360×270`)
    - `left` (`60×270`)
    - `right` (`60×270`)
  - attaches them via:

```javascript
page.display("left", display => display.surface(left));
page.display("main", display => display.surface(main));
page.display("right", display => display.surface(right));
```

  - renders a `12`-tile main scene using retained surfaces rather than `page.tile(...)`
  - renders a prototype left-strip animated bar scene
  - renders a prototype right-strip scrolling-kanji scene
  - uses `anim.loop(...)` to update the scene continuously
  - supports touch selection and button-based active-tile stepping in prototype form
- Found that the prototype initially crashed during example boot tests.
- Traced the panic to missing-field handling in:

```text
runtime/js/module_gfx/module.go
```

where omitted option fields in `gfx.text(...)` could make `boolProp(...)` dereference a nil `goja.Value`.
- Fixed that nil-handling bug in `module_gfx`.
- Found and fixed an additional text-layout robustness issue in:

```text
runtime/gfx/text.go
```

by clamping the computed text baseline against the face descent for small text boxes.
- Added a regression test in:

```text
runtime/gfx/surface_test.go
```

for small-height text rendering.
- Re-ran:

```bash
go test ./...
```

until the full example pack, including `07-cyb-ito-prototype.js`, booted cleanly.

### Why
- The runtime needed a real scene artifact to validate the direction of the new display-region and graphics work.
- A prototype scene is a much better proving ground than more internal abstractions because it forces the API to be used the way future scripts will actually use it.
- The prototype also flushed out real bugs in the new graphics binding and text layout paths, which is exactly what a good example should do.

### What worked
- The prototype proved that the runtime can now express a multi-display scene entirely in JS using retained surfaces.
- The example boot test suite now covers the new scene script automatically.
- The prototype exercise revealed and fixed real robustness bugs in the `gfx` path.

### What didn't work
- The prototype is not yet a faithful port of the imported HTML. It is structurally inspired by it, but not yet visually complete.
- True ripple overlays and multi-layer composition are still not implemented.
- The prototype has only basic cross-display coordination right now.

### What I learned
- The runtime is now at the point where example scenes are a better guide to next work than more speculative API discussion.
- The first prototype scene already justified the retained-surface work because it immediately used all three display regions.
- Small helper bugs in option decoding and text layout only became obvious once a denser scene script existed.

### What should be done in the future
- Update the ticket docs and mark the prototype scene tasks that are honestly complete.
- Continue Phase E toward true multi-layer display composition.
- Add a better touch-reactive overlay model next so the prototype can gain ripple and scan effects more naturally.

## Step 9: Validate the prototype on real hardware and tighten the UX based on what was actually visible

At this point the prototype already booted in tests, but the next milestone had to be a real hardware run. The user had asked to keep working through the ticket tasks one by one, so the next honest task was the Phase H hardware validation slice using the tmux workflow we had already standardized for the earlier JS runtime work.

### Commits

**Commit (code):** `7ac32d3` — `Improve cyb-ito prototype interaction feedback`

### What I did
- Started the prototype scene on actual hardware with:

```bash
go run ./cmd/loupe-js-live --script ./examples/js/07-cyb-ito-prototype.js --duration 120s --log-events
```

inside a fresh tmux session.
- The first real-hardware observation from the user was effectively:
  - the main display showed white buttons
  - the side displays were animating
  - the right strip showed unreadable fallback `?` glyphs
- That first observation was already valuable because it confirmed that the retained multi-display scene was rendering on all three physical regions.
- The next interaction check showed another real usability issue: the user pressed `Button1`, `Button2`, and touched the screen, but visually it seemed like nothing changed.
- I checked the captured hardware log and confirmed that the events were in fact arriving. The log recorded many lines such as:

```text
button event button=Button1 status=down
button event button=Button2 status=down
touch event touch=Touch3 status=down x=...
```

which meant the problem was not dead input wiring. The problem was that the prototype’s visual feedback was too subtle.
- Based on that hardware evidence, I updated:

```text
examples/js/07-cyb-ito-prototype.js
```

so the prototype now:
  - starts with a visible default active tile instead of `-1`
  - renders much stronger active-tile highlighting
  - shows an obvious central status box with the selected tile and last event
  - shows `BOOT`, `B1`, `B2`, or `TouchN` status text so input feedback is unmistakable
  - replaces the right-strip Kanji with ASCII fallback words so the side display is readable on current hardware demos
- Re-ran the scene on hardware in tmux and asked the user to retry the interaction test.
- The user then confirmed: **"yes, that works"**.
- Stopped the session cleanly with `C-c` and captured the final hardware evidence log at:

```text
/tmp/loupe-cyb-ito-prototype-1775989933.log
```

### Why
- A prototype scene that only works in automated boot tests is not enough for this repo; the whole point is real Loupedeck hardware behavior.
- The first hardware pass exposed a classic issue: technically correct event handling can still fail as a demo if visual feedback is too subtle.
- Tightening prototype UX based on the first human run is exactly the kind of iteration this ticket should capture.

### What worked
- The prototype rendered on all three physical display regions.
- The live runner delivered real touch and button events during the prototype session.
- The improved prototype made interaction state obvious enough that the user confirmed it was working on hardware.

### What didn't work
- The very first hardware presentation was too ambiguous to count as a good validation demo, even though the event path itself was alive.
- The original right-strip Kanji text was not a good fit for the current font/fallback path and produced unreadable output.
- The run still showed the familiar reconnect fragility on startup, including an initial malformed-response warning before the successful run continued.

### Concrete evidence
- Successful interactive hardware log:

```text
/tmp/loupe-cyb-ito-prototype-1775989933.log
```

- Example evidence lines from that log:

```text
2026/04/12 06:32:20 INFO touch event touch=Touch2 status=down x=188 y=51
2026/04/12 06:32:31 INFO button event button=Button1 status=down
2026/04/12 06:32:32 INFO button event button=Button2 status=down
2026/04/12 06:32:57 INFO button event button=Circle status=down
```

- First-run limitation evidence from the same hardware-validation cycle:

```text
2026/04/12 06:32:15 WARN dial failed err="malformed HTTP response ..."
```

### What I learned
- The retained multi-display architecture is now far enough along that a real JS scene can be interacted with on hardware.
- Human-visible demo quality matters separately from functional correctness. The event path was already alive before the user could confidently see it.
- ASCII fallback content is a better current demo choice than unsupported glyph sets when the goal is validation clarity rather than final art fidelity.

### What should be done in the future
- Update the ticket docs and mark the honest hardware-validation tasks complete.
- Continue to true multi-layer composition so the scene can gain ripple overlays instead of only active-tile highlighting and status text.
- After overlays exist, re-run hardware validation and then measure whether the denser scene needs renderer/writer pacing adjustments.

## Step 10: Add true retained display layers and move the first ripple effect into an overlay surface

After the first hardware-validation slice, the next architectural gap was obvious again: the prototype was still drawing everything into a single retained surface per display. That was good enough to prove the `gfx` bridge, but it was not yet the right structure for cyb-ito-style overlays. The ticket had explicitly called for layered composition, so the next real implementation step was to give displays true retained overlay layers rather than continuing to overload one surface.

### Commit

**Commit (code):** `e4d43a5` — `Add retained display layer composition`

### What I did
- Extended:

```text
runtime/ui/display.go
```

so a display can now own:
  - one base surface via `display.SetSurface(...)`
  - multiple named layers via `display.SetLayer(name, surface)`
- Added a retained ordering model:
  - layers are named
  - insertion order is preserved
  - replacing an existing layer keeps its position
  - removing a layer deletes it from the stable order list
- Kept the dirty model correct by subscribing each layer surface to display invalidation, just like the base surface.
- Extended:

```text
runtime/render/visual_runtime.go
```

so the renderer now composites:
  1. background
  2. base display surface
  3. each named layer in stable order
- Extended:

```text
runtime/js/module_ui/module.go
```

with:

```javascript
display.layer(name, surface)
display.layer(name, null)
```

so JS can attach or remove named layers directly.
- Added tests in:

```text
runtime/ui/ui_test.go
runtime/render/render_test.go
runtime/js/runtime_test.go
```

covering:
  - layer mutation dirty propagation
  - stable layer ordering
  - base-plus-layer renderer composition
  - JS integration for named layers
- Updated:

```text
examples/js/07-cyb-ito-prototype.js
```

so the prototype now has a dedicated overlay surface:

```javascript
const mainFX = gfx.surface(MAIN_W, MAIN_H);
display.layer("fx", mainFX);
```

- Moved the first ripple-like interaction effect into that overlay surface instead of forcing it into the main/base surface.
- Re-ran:

```bash
go test ./...
```

and the full suite passed.

### Why
- The imported HTML reference is fundamentally layered: base tile content, overlays, scan-like effects, side elements, and transient highlights all want distinct visual strata.
- A one-surface-per-display model is enough for a prototype, but it is the wrong abstraction boundary for the next wave of effects.
- Adding true layers now creates the correct path for ripple overlays, scan passes, and other transient effects without entangling them with the base scene raster.

### What worked
- The retained UI model can now express a base scene plus overlays instead of a single merged display surface.
- The renderer owns composition order, which keeps the transport boundary in Go exactly as intended.
- The prototype now actually exercises the new layer path instead of leaving it unused.

### What didn't work
- This is still an intentionally small first layer system. It does not yet include blend-mode variety, per-layer visibility flags, or layer-local transforms.
- The ripple effect is still only a first overlay sketch, not yet a faithful cyb-ito ripple implementation.
- I briefly made a trivial tooling mistake by trying to run `gofmt` on the JS prototype file, which of course failed with:

```text
examples/js/07-cyb-ito-prototype.js:1:1: expected 'package', found 'const'
```

That was immediately corrected by re-running `gofmt` only on the Go files.

### What I learned
- The next runtime frontier after base surfaces really is layered composition, not more one-surface helpers.
- The combination of named layers plus stable insertion order is enough to begin exercising overlay semantics without overdesigning a full scene graph too early.
- The prototype is now a better forcing function again because it can use an overlay surface for transient interaction feedback.

### What should be done in the future
- Validate the new overlay/ripple path on real hardware.
- Consider whether the prototype should split scanlines, ripple, and status overlays into separate layers rather than one shared `fx` layer.
- Continue toward denser cyb-ito-inspired composition and then reassess pacing under the heavier layered workload.

## Step 11: Re-run the prototype on hardware after the new layer model landed

Once the real retained layer model existed and the prototype had been updated to put its first ripple-like effect into a dedicated `fx` overlay surface, the very next honest step was another hardware run. This was important because it is easy for a layered-render path to be correct in unit tests while still being visually wrong, too subtle, or too expensive on the actual device.

### What I did
- Restarted the live runner in tmux with the layered prototype script:

```bash
go run ./cmd/loupe-js-live --script ./examples/js/07-cyb-ito-prototype.js --duration 120s --log-events
```

- Confirmed from the log that the session was alive and drawing all three displays again.
- Asked the user to verify two things on-device:
  1. that the clearer selection/status behavior from the previous UX pass still worked
  2. that the new overlay/ripple effect was actually visible
- The user responded: **"yes"**.
- Stopped the run cleanly and captured the hardware evidence log:

```text
/tmp/loupe-cyb-ito-layers-1775990488.log
```

### Why
- The layer system is only meaningful if it survives contact with actual hardware.
- This rerun closes the loop between architecture, implementation, and visible device behavior.
- It also reduces risk before moving on to denser scene layering, because we now know the first overlay path is not purely theoretical.

### What worked
- The layered prototype rendered on hardware.
- Touch activity continued to arrive during the layered run.
- The user confirmed both the preserved interaction feedback and the visible overlay/ripple effect.

### What didn't work
- The now-familiar reconnect fragility still showed up as an initial malformed-response warning before the run proceeded successfully.
- This run validated visibility and interaction, but it did not yet measure whether the layered workload changes pacing/throughput characteristics.

### Concrete evidence
- Hardware log:

```text
/tmp/loupe-cyb-ito-layers-1775990488.log
```

- Example lines:

```text
2026/04/12 06:41:32 INFO touch event touch=Touch8 status=down x=374 y=140
2026/04/12 06:41:34 INFO touch event touch=Touch4 status=down x=385 y=24
2026/04/12 06:41:48 INFO button event button=Circle status=down
2026/04/12 06:41:49 WARN Read error, exiting error="Port has been closed"
```

### What I learned
- The layer model is now real enough to be treated as part of the active runtime rather than a speculative abstraction.
- The prototype can now use overlays on hardware, not just in tests or screenshots.
- The next real question is no longer "can layers work?" but rather "how should the scene be split across several layers, and does that change pacing requirements?"

### What should be done in the future
- Split the current shared `fx` overlay into more semantically focused layers if needed, such as ripple vs scan/status.
- Continue making the prototype more like the imported cyb-ito scene while preserving Go-owned rendering/composition.
- Add a focused measurement pass for layered animation workload once the scene becomes denser.
