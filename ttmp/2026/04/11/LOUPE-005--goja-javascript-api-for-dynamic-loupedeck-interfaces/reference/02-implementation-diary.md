---
Title: Implementation diary
Ticket: LOUPE-005
Status: active
Topics:
    - loupedeck
    - go
    - goja
    - javascript
    - animation
    - rendering
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md
      Note: Main design brainstorm doc for the JS API direction
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
      Note: Intern-oriented conceptual deep dive for the preferred reactive design
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
      Note: Intern-oriented phased implementation plan for building the preferred runtime
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md
      Note: Example-script reference pack for multiple API styles and scenarios
    - Path: renderer.go
      Note: Current render scheduler constraints discussed in the brainstorm
    - Path: writer.go
      Note: Current transport ownership layer that the JS API must not bypass
ExternalSources: []
Summary: Chronological diary for the goja JavaScript API design ticket, including ticket creation, design-space mapping, example-script brainstorming, and intern-oriented textbook/implementation-plan authoring.
LastUpdated: 2026-04-11T20:40:45-04:00
WhatFor: Preserve the exact design framing, commands, and rationale behind the early planning passes for a future scriptable Loupedeck runtime.
WhenToUse: Use when resuming the JS API design work or when reviewing why a hybrid retained/timeline/state model was favored and how the reactive implementation should be phased.
---

# Diary

## Goal

Capture the first design pass for a future goja JavaScript API on top of the current Loupedeck renderer/writer stack, including the ticket creation, design-space exploration, and example scripts.

## Step 1: Create the JS API ticket and write the first deep brainstorm package

This step created a dedicated ticket for the goja/JavaScript runtime idea rather than burying the design work inside the renderer or SVG tickets. That separation matters because the question here is not merely “how do we add scripting?” but “what is the correct abstraction boundary above the existing Go-side rendering and transport layers?”

I treated this first pass as a design package rather than an implementation task. The deliverable is a well-structured brainstorm that compares multiple API styles, spells out the hard constraints from the current system, and provides many example scripts so future implementation work can be judged against actual ergonomics instead of abstract preferences.

### Prompt Context

**User prompt (verbatim):** "yes continue, but make that a new ticket with docmgr about adding a goja JS API. Then make a brainstorm document with multiple approaches and then multiple examples. try different ideas and all kinds of scenarios. Especially I want to do animations with easing curves and the like."

**Assistant interpretation:** Create a fresh docmgr ticket specifically for the goja/JS API direction, then write a deep brainstorm document and example pack exploring several API shapes and many scenarios, with special attention to animation and easing.

**Inferred user intent:** Avoid prematurely locking into one scripting model; instead, explore the design space broadly and concretely so the later implementation direction is chosen intentionally.

**Commit (code):** N/A — documentation and ticket-design step.

### What I did
- Created the new ticket:

```bash
docmgr ticket create-ticket --ticket LOUPE-005 --title "Goja JavaScript API for dynamic Loupedeck interfaces" --topics loupedeck,go,goja,javascript,animation,rendering
```

- Added the core docs:

```bash
docmgr doc add --ticket LOUPE-005 --doc-type design-doc --title "Brainstorm: goja JavaScript API approaches for dynamic Loupedeck interfaces"
docmgr doc add --ticket LOUPE-005 --doc-type reference --title "JavaScript API example scripts"
docmgr doc add --ticket LOUPE-005 --doc-type reference --title "Implementation diary"
```

- Replaced the ticket templates with:
  - a real `index.md`
  - a structured `tasks.md`
  - a ticket `changelog.md`
  - a deep brainstorm design doc
  - an example-script reference pack
  - this diary entry
- Grounded the brainstorm in the current renderer/writer realities by referencing:
  - `display.go`
  - `renderer.go`
  - `writer.go`
  - `cmd/loupe-svg-buttons/main.go`
  - `svg_icons.go`
- Included multiple approaches in the brainstorm:
  - low-level imperative
  - retained declarative pages
  - reactive signals/stores
  - timeline-centric animation
  - hybrid retained + state + timeline model
- Included multiple example scenarios with easing curves and timelines.

### Why
- The current project now has enough rendering and animation groundwork that a scripting API is plausible, but not enough that the right top-level API is obvious.
- A single “best guess” API would be too premature; comparing multiple styles makes later implementation choices more defensible.
- Example scripts are necessary because elegant scripting surfaces often sound good in prose and feel bad in practice.

### What worked
- The ticket scaffolding and docs were created cleanly.
- The current renderer/writer architecture gives the brainstorm a strong constraint set, which makes the design work concrete rather than speculative hand-waving.
- The hybrid retained/timeline/state model emerged as a clearly stronger direction than either a raw imperative API or an oversized framework-first design.

### What didn't work
- No technical tool failures occurred in this step.
- The main limitation is intentional: this pass does not yet pick one exact implementation contract. It is a design-space mapping exercise.

### What I learned
- The current Go-side renderer/writer split already suggests the right scripting boundary: JavaScript should target UI/state/animation abstractions, not transport operations.
- Easing curves are not a minor feature request; they materially shape the module design because they imply a first-class animation subsystem rather than ad hoc timers.
- A small declarative page model plus explicit state and a timeline module looks much more elegant than exposing everything as direct tile mutation.

### What was tricky to build
- The hardest part was balancing breadth and concreteness. A brainstorm can easily become too vague to be useful, while an overcommitted design doc can accidentally turn into an implementation decision before alternatives have been explored. The solution here was to compare multiple styles directly and attach many realistic scripts to them.
- Another subtle point was keeping the design honest about the current hardware and renderer constraints. It is easy to imagine a luxurious JS UI runtime that ignores transport ownership, dirty-region planning, and reconnect behavior. This brainstorm deliberately keeps those lower-level realities in view.

### What warrants a second pair of eyes
- The hybrid recommendation is strong, but a reviewer should still challenge whether the first implementation slice should start even smaller—for example with only pages + events before adding signals and timelines.
- The examples currently assume several module boundaries (`ui`, `state`, `anim`, `easing`). A future implementation pass should re-evaluate whether that split is the right one for discoverability.

### What should be done in the future
- Validate the ticket with `docmgr doctor`.
- Commit the new ticket docs.
- Narrow the brainstorm into a smaller RFC for an initial implementable slice.

### Code review instructions
- Read in this order:
  - `ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md`
  - `ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md`
  - `ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/tasks.md`
- Cross-check the current renderer assumptions against:
  - `renderer.go`
  - `writer.go`
  - `display.go`

### Technical details
- The most important emerging recommendation from this step is:

```text
retained page/tile model + explicit state helpers + first-class animation/timeline/easing module + narrow imperative escape hatch
```

- This keeps transport and low-level rendering policy in Go while letting scripts describe dynamic behavior elegantly.

## Step 2: Write the reactive textbook and the detailed execution plan for a new intern

After the first brainstorm pass, the next request was not for more alternatives but for something much more operational: a teaching document that explains the preferred reactive design in textbook form, and then an execution plan that a new intern could actually follow to build it in the right order.

This changed the character of the ticket. The first documents were “what are the possible API shapes?” documents. The new documents are “what exactly do we mean by the reactive model, what does mutating a signal mean, and what is the implementation order that preserves transport safety?” documents. That is a very different deliverable, and it is much more directly useful to the person who will eventually build the system.

### Prompt Context

**User prompt (verbatim):** "write a detailed textbook course document about the implementation of the reactive version, what mutate signals means, etc... 

Create a detailed textbook that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Add a deep intern-oriented textbook for the preferred reactive runtime, add a concrete implementation plan for building it, update ticket bookkeeping, and upload the expanded design package to the reMarkable.

**Inferred user intent:** Move the ticket from broad ideation to execution-ready onboarding material, so a new engineer can understand the reactive model and then implement it in the correct order without rediscovering the architecture.

**Commit (code):** N/A — documentation and delivery step.

### What I did
- Added a new textbook-style design doc:

```text
design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
```

- Wrote it as a course-style explanation covering:
  - why reactive state is preferred here
  - what a signal is
  - what “mutate a signal” means semantically
  - how computed values and dependency tracking work
  - how retained UI nodes participate in the reactive graph
  - how animations should mutate state or retained properties rather than transport
  - how the runtime layers stack on top of the existing renderer/writer
- Added a new execution-plan design doc:

```text
design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
```

- Broke the implementation into explicit phases:
  1. narrow first stable JS slice
  2. pure-Go reactive core
  3. retained UI model
  4. retained visual realization
  5. bridge into current renderer/writer
  6. host event runtime shell
  7. pure-Go animation/easing engine
  8. goja bindings
  9. first end-to-end example
  10. reconnect replay later
- Added concrete package/file proposals, tests, acceptance criteria, and PR-sized milestone boundaries.
- Updated the ticket `index.md`, `tasks.md`, and `changelog.md` to include the new docs and reflect that several implementation-planning decisions are now intentionally made.

### Why
- The earlier brainstorm doc was useful for choosing a direction, but it was still too broad for a new engineer to implement from directly.
- The intern needs two different kinds of documents:
  1. a conceptual textbook to understand the system and vocabulary
  2. an execution plan to know what to build first and why
- The request explicitly emphasized signals and mutation semantics, so the textbook needed to explain not just API names but the propagation model underneath them.

### What worked
- The textbook naturally clarified several important design decisions that were only implicit in the brainstorm, especially:
  - signals as explicit state cells with propagation semantics
  - batching as a first-class requirement, not a later optimization
  - host-owned animation timelines as the correct default
  - goja adapters as thin bindings over pure-Go domain logic
- The implementation plan now gives a concrete and reviewable build order rather than a vague “we should probably do hybrid retained state”.
- The two-document structure (conceptual deep dive + phased execution plan) fits intern onboarding well.

### What didn't work
- No tooling failed during authoring.
- The only limitation is that the new plan still intentionally leaves some detailed semantic choices open for later implementation passes, such as exact `computed` evaluation policy (lazy vs eager) and reconnect script semantics.

### What I learned
- Once the reactive model is explained carefully, the implementation order becomes much easier to justify. The strongest recurring rule is: **do semantic work in pure Go before binding it into goja**.
- The phrase “mutate a signal” really does need an explicit textbook explanation. Without that, it is too easy for a new engineer to think of signals as just wrapper variables instead of runtime-visible state cells with dependency propagation.
- The implementation plan benefits from being framed in PR-sized milestones, because this system has many layers and it would be easy to create an unreviewable mega-branch otherwise.

### What was tricky to build
- The hardest part was writing a document that is detailed enough for an intern without drifting into an accidental code dump or fake certainty. The solution was to stay very explicit about semantics, responsibilities, and sequencing while still marking a few choices as open questions.
- Another tricky point was keeping the textbook grounded in the current repository rather than inventing an imaginary framework. That is why the documents repeatedly reference `display.go`, `renderer.go`, `writer.go`, `svg_icons.go`, and `cmd/loupe-svg-buttons/main.go`.

### What warrants a second pair of eyes
- A reviewer should challenge the exact first-slice boundary proposed in the execution plan and confirm whether animations belong in the first or second goja milestone.
- The proposed package layout is sensible, but the team may prefer a flatter or more nested arrangement depending on repository conventions.

### What should be done in the future
- Run `docmgr doctor` again after the new docs are in place.
- Commit the expanded ticket docs.
- Upload the expanded bundle to the reMarkable.
- Later, turn `design-doc/03-...` into a narrower implementation RFC for the very first coding slice.

### Code review instructions
- Read in this order:
  - `design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
  - `design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
  - `design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md`
  - `reference/01-javascript-api-example-scripts.md`
- Cross-check architecture references in:
  - `display.go`
  - `renderer.go`
  - `writer.go`
  - `svg_icons.go`

### Technical details
- The most important implementation-order rule in the new plan is:

```text
pure Go reactive core -> retained UI model -> retained visuals -> renderer bridge -> host runtime shell -> goja adapters -> animations/easing JS surface
```

- The most important conceptual rule in the textbook is:

```text
mutating a signal means changing a state cell and propagating invalidation through a dependency graph, not drawing directly to the device
```
