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
- Later, turn `design-doc/03-...` into a narrower implementation RFC for the very first coding slice.

## Step 3: Validate, commit, and upload the expanded intern bundle

After the new textbook and implementation-plan docs were written and the ticket index/tasks/changelog were updated, the next job was continuity and delivery: validate the ticket, commit the docs, upload the expanded bundle to the reMarkable, and verify the remote folder contents.

### Prompt Context

**User prompt (verbatim):** "ok, create a detailed implementat plan now to build it and stor ein ticket and upload to remarkable, same again for our intern."

**Assistant interpretation:** Finish the documentation package operationally, not just editorially: validate it with `docmgr`, commit the new ticket docs, upload the expanded LOUPE-005 bundle to the reMarkable, and record the result.

**Inferred user intent:** Make the plan durable and reviewable in all of the usual places used for this project: ticket docs, git history, and reMarkable delivery.

**Commit (code):** `a3761dc` — `Expand LOUPE-005 with reactive textbook and implementation plan`

### What I did
- Ran:

```bash
docmgr doctor --ticket LOUPE-005 --stale-after 30
```

- Result: all checks passed.
- Committed the expanded LOUPE-005 docs, including:
  - `design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
  - `design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
  - updated `index.md`, `tasks.md`, `changelog.md`, and this diary
- Performed a dry-run reMarkable bundle upload including:
  - brainstorm design doc
  - reactive textbook
  - implementation plan
  - example scripts
  - diary
- Uploaded the real bundle with the document name:

```text
LOUPE-005 Reactive Goja UI Runtime textbook and implementation plan
```

- Target reMarkable folder:

```text
/ai/2026/04/11/LOUPE-005
```

- Verified remote contents with:

```bash
remarquee cloud ls "/ai/2026/04/11/LOUPE-005/" --long --non-interactive
```

- Verified that the folder now contains:
  - `LOUPE-005 Goja JavaScript API brainstorm`
  - `LOUPE-005 Reactive Goja UI Runtime textbook and implementation plan`

### Why
- The request explicitly asked for ticket storage and reMarkable delivery, so stopping after local markdown authoring would have left the task incomplete.
- Validation and commit history matter because this ticket is evolving into the design record for a future implementation effort.

### What worked
- `docmgr doctor` passed cleanly after the updates.
- The bundle upload worked on the first real attempt after a successful dry-run.
- The remote listing confirmed the document landed in the intended LOUPE-005 folder next to the earlier brainstorm bundle.

### What didn't work
- No tooling failure occurred in this step.
- The only remaining bookkeeping item was to mark the tasks complete and extend the changelog/diary with the delivery evidence, which is what this step is doing.

### What I learned
- The new LOUPE-005 package is now strong enough to function as a genuine onboarding bundle: concept, examples, phased plan, and diary are all in one place and mirrored to reMarkable.
- It is useful to keep the brainstorm bundle and the reactive-textbook bundle as separate reMarkable artifacts because they serve slightly different reading goals.

### What warrants a second pair of eyes
- A reviewer may want to confirm whether the uploaded bundle should also include `index.md` and `tasks.md` in future deliveries, or whether the current design-doc/reference-only bundle is the better reader experience.

### Code review instructions
- Review the committed docs at commit `a3761dc`.
- Confirm the remote reMarkable folder contains both LOUPE-005 PDFs.
- If desired, derive the first actual implementation RFC directly from `design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`.

### Technical details
- Validation command used:

```bash
docmgr doctor --ticket LOUPE-005 --stale-after 30
```

- Upload command family used:

```bash
remarquee upload bundle ... --name "LOUPE-005 Reactive Goja UI Runtime textbook and implementation plan" --remote-dir "/ai/2026/04/11/LOUPE-005" --toc-depth 2
```

- Verification command used:

```bash
remarquee cloud ls "/ai/2026/04/11/LOUPE-005/" --long --non-interactive
```

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

## Step 4: Implement milestone A as a pure-Go reactive core

The first real coding step for LOUPE-005 was intentionally below goja. The point was to implement the semantic heart of the runtime first: signals, computed values, dirty propagation, batching, and eager watchers/effects. This was the right starting point because it let me validate the reactive model in isolation, without mixing in any JS-module or runtime-loader complexity yet.

This step also tested whether the implementation order from the new textbook and plan was actually practical. It was: building the reactive core first forced the API to stay small and clear, and it immediately surfaced one concrete Go constraint that would have been more annoying to discover after building adapters.

### Prompt Context

**User prompt (verbatim):** "Add detailed tasks for the impllementation to the ticket, and then implement them one by one, committing at appropriat eintervals, and keeping a detailed diary"

**Assistant interpretation:** Expand the ticket into execution-grade implementation tasks, then start performing those milestones in order with real code, tests, commits, and diary continuity.

**Inferred user intent:** Move LOUPE-005 out of design-only mode and into disciplined implementation mode, with visible task progress and durable documentation after each meaningful coding slice.

**Commit (code):** `afcc313` — `Add pure Go reactive runtime core`

### What I did
- Expanded `tasks.md` with concrete milestone sections for:
  - pure-Go reactive core
  - retained UI
  - renderer bridge
  - host shell
  - goja adapters
  - animation/easing
  - reconnect replay
- Added the new package:

```text
runtime/reactive/
```

- Implemented:
  - `runtime/reactive/runtime.go`
  - `runtime/reactive/graph.go`
  - `runtime/reactive/signal.go`
  - `runtime/reactive/computed.go`
  - `runtime/reactive/effect.go`
- Added tests in:

```text
runtime/reactive/runtime_test.go
```

- Supported behaviors include:
  - `Runtime` batching and flush scheduling
  - generic `Signal[T]` with `Get`, `Set`, `Update`
  - generic `Computed[T]` with lazy reevaluation and dependency tracking
  - eager `Watch`/effect support with unsubscribe via `Stop()`
  - dirty propagation through dependency graphs
  - cycle/reentrancy panic guards for computed/effect evaluation
- Ran:

```bash
gofmt -w runtime/reactive/*.go
go test ./...
```

### Why
- This is the smallest slice that validates the core reactive semantics promised by the textbook and implementation plan.
- By keeping the package pure Go and goja-free, the tests answer the important question first: “does the reactive graph behave the way we want?”

### What worked
- The package structure stayed small and readable.
- The tests cover the most important semantic scenarios from the design docs:
  - equal-value no-op updates
  - invalidation chains
  - diamond dependency graphs
  - batching
  - stop/detach behavior
  - cycle/reentrancy guards
- The full repository test suite passed once the API was adjusted for the current Go version.

### What didn't work
- My first API draft used generic methods on `Runtime`:

```go
func (r *Runtime) Signal[T any](initial T) *Signal[T]
func (r *Runtime) Computed[T any](fn func() T) *Computed[T]
```

- That failed under the repository’s Go toolchain with:

```text
runtime/reactive/runtime.go:17:25: method must have no type parameters
runtime/reactive/runtime.go:21:34: method must have no type parameters
runtime/reactive/runtime.go:32:27: method must have no type parameters
```

- Command that failed:

```bash
go test ./...
```

- Fix: switch to package-level generic constructors instead:

```go
func NewSignal[T any](r *Runtime, initial T) *Signal[T]
func NewSignalWithEqual[T any](r *Runtime, initial T, equal func(a, b T) bool) *Signal[T]
func NewComputed[T any](r *Runtime, fn func() T) *Computed[T]
```

### What I learned
- The implementation-order rule from the plan is already proving itself. By starting in pure Go, I found and resolved the API-shape issue before any JS adapter code existed.
- Eager `Watch` support is worth having in the core package even before the JS layer, because it makes batching and propagation behavior much easier to test and reason about.

### What was tricky to build
- The dependency graph needs two kinds of participants: dependency **sources** (signals/computeds) and dirty **dependents** (computeds/effects). Getting that split explicit early made the implementation much cleaner.
- Another tricky point was making computed dependencies dynamic rather than one-time-bound. The solution was to clear old dependencies before each reevaluation and rebuild them while evaluating the computed in a collector context.
- The panic guards are intentionally simple for now: they detect obvious cyclic computed evaluation and reentrant effect execution rather than attempting a sophisticated cycle-resolution system.

### What warrants a second pair of eyes
- The current computed semantics are lazy and may over-invalidate downstream dependents when a source changes but a computed value ends up equal after reevaluation. That is acceptable for the first slice, but a reviewer should keep an eye on whether later milestones need equality-aware downstream suppression.
- The runtime is currently intentionally single-threaded. If future host code tries to mutate it from multiple goroutines, additional synchronization or scheduling discipline will be needed.

### What should be done in the future
- Build milestone B on top of this package: retained pages/tiles with reactive property bindings.
- Decide whether future computed values need configurable equality semantics or whether the current always-dirty downstream policy is sufficient.

### Code review instructions
- Start with:
  - `runtime/reactive/runtime.go`
  - `runtime/reactive/signal.go`
  - `runtime/reactive/computed.go`
  - `runtime/reactive/effect.go`
  - `runtime/reactive/runtime_test.go`
- Validate with:

```bash
gofmt -w runtime/reactive/*.go
go test ./...
```

### Technical details
- The implementation uses a collector-scoping pattern:

```text
signal/computed Get() -> register dependency with current collector
signal Set() -> mark dependents dirty
computed/effect reevaluation -> clear old deps, rebuild deps while collecting Get() calls
```

- The milestone-A public constructor shape is now:

```go
rt := reactive.NewRuntime()
count := reactive.NewSignal(rt, 0)
label := reactive.NewComputed(rt, func() string {
    return fmt.Sprintf("%d", count.Get())
})
sub := rt.Watch(func() {
    _ = label.Get()
})
sub.Stop()
```

## Step 5: Implement milestone B as a retained page/tile UI layer

With the reactive core in place, the next layer was the smallest useful retained UI model: pages and `4x3` touchscreen tiles. This step intentionally stayed pure Go again. The goal was to prove that reactive state can drive retained node properties and dirty-node tracking before any renderer bridge or goja adapter exists.

This was a good second slice because it forced the project to define the host-owned UI state shape early. Instead of talking abstractly about “some future page model”, the code now has a concrete page registry, active-page concept, tile coordinates, tile property bindings, and dirty-tile collection.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue executing the phased implementation plan after milestone A, keeping the same discipline of tests, commits, and diary continuity.

**Inferred user intent:** Build the reactive runtime incrementally in the intended order, with each layer becoming concrete and validated before the next one starts.

**Commit (code):** `1eefa17` — `Add retained UI runtime model`

### What I did
- Added the new package:

```text
runtime/ui/
```

- Implemented:
  - `runtime/ui/ui.go`
  - `runtime/ui/page.go`
  - `runtime/ui/tile.go`
  - `runtime/ui/ui_test.go`
- Added a host-owned retained UI model with:
  - page registration via `AddPage`
  - active-page switching via `Show`
  - `4x3` main-touchscreen coordinate validation
  - tile nodes with `text`, `icon`, and `visible` properties
  - static setters (`SetText`, `SetIcon`, `SetVisible`)
  - reactive property binding helpers (`BindText`, `BindIcon`, `BindVisible`)
  - dirty-tile collection and active-page filtering
- Made `ui.New(nil)` create its own reactive runtime so tests and future callers can use a simple default.
- Ran:

```bash
gofmt -w runtime/ui/*.go
go test ./...
```

### Why
- The implementation plan explicitly called for retained UI before the renderer bridge. That order matters because the renderer needs concrete retained nodes to render.
- The first node type was intentionally kept small: tiles only, no strips or overlays yet. That keeps the API understandable for the intern and reduces early surface area.

### What worked
- The retained model fits naturally on top of `runtime/reactive`.
- The reactive tile bindings are simple because they reuse `Runtime.Watch(...)` rather than inventing a second binding mechanism.
- The tests cover the most important milestone-B semantics:
  - page switching
  - dirty-tile filtering by active page
  - static property updates
  - reactive property updates through bound signals

### What didn't work
- No failing test or build error occurred in this milestone once the files were implemented.
- The main limitation is intentional: the model currently tracks dirtiness but does not yet know how to render or clear tiles visually. That is the next milestone’s job.

### What I learned
- The `visible` property should still mark tiles dirty even when the tile becomes hidden, because the eventual renderer bridge will need to know that a region must be redrawn/cleared.
- Filtering `DirtyTiles()` by the active page keeps the retained model simple while still letting hidden-page changes accumulate harmlessly until that page is shown.

### What was tricky to build
- The subtle point in this slice is that dirty-node tracking and visible-page filtering are different concerns. A tile can be dirty while its page is hidden; that should not be lost. The current design keeps the dirty bit on the tile and only filters at readout time (`DirtyTiles()`), which is a good fit for a later bridge.
- Another small but important design choice was whether `Show(name)` should mark the page’s tiles dirty. It should, because page activation is itself a visibility change and the newly active page must be rendered by the future bridge.

### What warrants a second pair of eyes
- The current property set is intentionally tiny. A reviewer should confirm that milestone C should continue with just text/icon/visible rendering rather than adding transforms or style concepts prematurely.
- Hidden-page dirty handling is reasonable for now, but future milestones may want a more explicit dirty-reason model if page transitions become more complex.

### What should be done in the future
- Build milestone C: map dirty retained tiles to concrete main-display rectangles and feed them into the existing `Display.Draw()` path.
- Decide the minimal first tile visual composition strategy (text-only, icon-only, or icon-plus-text) for the renderer bridge.

### Code review instructions
- Start with:
  - `runtime/ui/ui.go`
  - `runtime/ui/tile.go`
  - `runtime/ui/ui_test.go`
- Validate with:

```bash
gofmt -w runtime/ui/*.go
go test ./...
```

### Technical details
- The retained tile coordinate contract is now explicitly:

```text
4 columns x 3 rows on the 360x270 main display
=> each tile maps conceptually to a 90x90 region
```

- The current public shape of the layer is approximately:

```go
rt := reactive.NewRuntime()
ui := ui.New(rt)
page := ui.AddPage("home")
tile := page.AddTile(0, 0)
label := reactive.NewSignal(rt, "IDLE")
tile.BindText(func() string { return label.Get() })
_ = ui.Show("home")
dirty := ui.DirtyTiles()
```
