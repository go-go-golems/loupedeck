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
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md
      Note: Example-script reference pack for multiple API styles and scenarios
    - Path: renderer.go
      Note: Current render scheduler constraints discussed in the brainstorm
    - Path: writer.go
      Note: Current transport ownership layer that the JS API must not bypass
ExternalSources: []
Summary: Chronological diary for the goja JavaScript API design ticket, including ticket creation, design-space mapping, and example-script brainstorming.
LastUpdated: 2026-04-11T20:40:45-04:00
WhatFor: Preserve the exact design framing, commands, and rationale behind the first brainstorming pass for a future scriptable Loupedeck runtime.
WhenToUse: Use when resuming the JS API design work or when reviewing why a hybrid retained/timeline/state model was favored in the brainstorm.
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
