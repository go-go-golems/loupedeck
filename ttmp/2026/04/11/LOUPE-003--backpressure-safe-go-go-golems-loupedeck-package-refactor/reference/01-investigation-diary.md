---
Title: Investigation diary
Ticket: LOUPE-003
Status: active
Topics:
    - loupedeck
    - go
    - serial
    - websocket
    - backpressure
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go.mod
      Note: Diary records the root module creation
    - Path: listen.go
      Note: Diary records the lifecycle change from panic to error
    - Path: listeners.go
      Note: Diary records the composable listener implementation
    - Path: listeners_test.go
      Note: Diary records the Phase 1 validation tests
    - Path: sources/loupedeck-repo/display.go
      Note: Evidence discussed in diary for direct draw behavior
    - Path: sources/loupedeck-repo/inputs.go
      Note: Evidence discussed in diary for single-slot binding behavior
    - Path: ttmp/2026/04/11/LOUPE-002--loupedeck-live-feature-tester-comprehensive-hardware-exercise/scripts/feature_tester.go
      Note: Application evidence referenced while writing the guide
    - Path: ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md
      Note: Primary design deliverable recorded by this diary
    - Path: ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/tasks.md
      Note: Diary tracks task progression across phases
ExternalSources: []
Summary: Chronological diary for the package refactor analysis, including ticket setup, evidence gathering, and authoring of the implementation guide.
LastUpdated: 2026-04-11T22:10:00-04:00
WhatFor: Preserve the exact analysis path, commands, and decisions that produced the LOUPE-003 design guide.
WhenToUse: Use when resuming work on the package refactor or when reviewing why B-lite was selected as the first implementation phase.
---



# Diary

## Goal

Capture the creation of LOUPE-003, the evidence-gathering process, and the reasoning behind the phased recommendation to do B-lite first, then B, and only later assess whether C is needed.

## Step 1: Create the package-refactor ticket and write the primary design guide

This step converted the earlier LOUPE-002 discussion into a formal package-level follow-up ticket. The focus shifted from patching one feature tester to designing a maintainable, top-level `github.com/go-go-golems/loupedeck` package that can own transport pacing, composable inputs, and future rendering behavior.

The work in this step was documentation-heavy on purpose. Before touching code, I wanted to anchor the refactor plan in concrete evidence from the current upstream clone and from the current feature tester, so that the next engineer can understand the existing system and the proposed direction without rediscovering the same architectural pain points.

### Prompt Context

**User prompt (verbatim):**
> alright, create a new docmgr ticket to do B-lite first, then B. We'll then assess if we need C. We can now stsart working in the toplevel and make a real github.com/go-go-golems/loupedeck package.
>
> reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
>   references.
>   It should be very clear and detailed. Store in the ticket and the nupload to remarkable.

**Assistant interpretation:** Create a new ticket for the package refactor, write a detailed onboarding-quality design guide centered on B-lite then B, keep diary/bookkeeping, and prepare the deliverable for reMarkable upload.

**Inferred user intent:** Move from experimental app-level fixes to a package-owned transport/render design, while producing documentation detailed enough that a new engineer can implement the plan safely.

**Commit (code):** N/A — documentation phase only so far.

### What I did
- Loaded the relevant workflow skills:
  - `/home/manuel/.pi/agent/skills/ticket-research-docmgr-remarkable/SKILL.md`
  - `/home/manuel/.pi/agent/skills/docmgr/SKILL.md`
  - `/home/manuel/.pi/agent/skills/remarkable-upload/SKILL.md`
  - `/home/manuel/.pi/agent/skills/ticket-research-docmgr-remarkable/references/writing-style.md`
- Checked docmgr repository status with:
  - `docmgr status --summary-only`
- Inventoried the repository with:
  - `find . -maxdepth 3 -type f | sort | sed 's#^./##' | head -200`
  - `rg --files . | sort | sed 's#^./##' | head -300`
- Gathered module and file-size evidence with:
  - `rg -n '^module ' -S .`
  - `wc -l sources/loupedeck-repo/{connect.go,dialer.go,display.go,inputs.go,intknob.go,listen.go,loupedeck.go,message.go,multibutton.go,touchdials.go,watchedint.go}`
- Read and line-inspected the main implementation files:
  - `sources/loupedeck-repo/display.go`
  - `sources/loupedeck-repo/message.go`
  - `sources/loupedeck-repo/listen.go`
  - `sources/loupedeck-repo/inputs.go`
  - `sources/loupedeck-repo/intknob.go`
  - `sources/loupedeck-repo/touchdials.go`
  - `sources/loupedeck-repo/multibutton.go`
  - `sources/loupedeck-repo/dialer.go`
  - `ttmp/.../LOUPE-002.../scripts/feature_tester.go`
- Created the new ticket:
  - `docmgr ticket create-ticket --ticket LOUPE-003 --title "Backpressure-safe go-go-golems loupedeck package refactor" --topics loupedeck,go,serial,websocket,backpressure`
- Verified the ticket exists with:
  - `docmgr ticket list`
- Added two documents to the ticket:
  - design doc: `01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
  - diary: `01-investigation-diary.md`
- Replaced the design doc template with a long-form analysis covering:
  - current-state architecture
  - evidence-backed failure analysis
  - package layout proposal
  - B-lite plan
  - full B plan
  - later C decision gate
  - API sketches, pseudocode, diagrams, and file references
- Replaced this diary template with the current chronological entry.

### Why
- The earlier issue was identified correctly as a backpressure problem, but the fix direction needed package-level structure, not more app-level sleeps.
- The current codebase is still arranged as ticket-local experiments plus a cloned upstream dependency. That is the wrong shape for the next phase of work.
- A new engineer needs a document that explains both the current system and the migration path. Without that, they will likely make local fixes instead of addressing the architectural bottlenecks.

### What worked
- `docmgr ticket create-ticket` successfully created `LOUPE-003`.
- `docmgr ticket list` confirmed the ticket workspace existed even though the first `doc add` calls failed immediately afterward.
- The current codebase was small enough to inspect thoroughly in one pass (~1575 LOC in the upstream clone), which made it feasible to write a detailed evidence-backed guide.
- The earlier LOUPE-002 investigation already contained enough context to support a structured package-level recommendation.

### What didn't work
- Immediately after creating the ticket, the first attempt to add docs failed with:

```text
Error: failed to find ticket directory: ticket not found: LOUPE-003
```

- This happened on both of the initial `docmgr doc add` calls even though ticket creation had just reported success.
- Running `docmgr ticket list` showed that the ticket did in fact exist, and retrying the `doc add` commands afterward worked.
- No code was changed yet, so transport/hardware behavior was not exercised in this step.

### What I learned
- There may be a short-lived docmgr visibility/indexing delay immediately after ticket creation, or at least a transient lookup inconsistency. The safe move is to verify with `docmgr ticket list` and retry.
- The most important architectural problem is not only transport pacing. The single-slot callback model (`BindKnob`, `BindTouch`, `BindButton`) makes widget composition unsafe and increases debugging confusion.
- The strongest migration story is “B-lite first” because it addresses lifecycle, event composition, and writer ownership before introducing a larger render scheduler.

### What was tricky to build
- The biggest documentation challenge was balancing breadth with actionable specificity. The guide had to explain transport, protocol, display drawing, widget behavior, and repository shape clearly enough for a new engineer, while also preserving concrete file evidence and not drifting into speculation.
- The other tricky part was keeping the phase recommendations disciplined. It was tempting to jump directly to a fully coalesced renderer or to a strict ack-gated transport design, but the current code evidence supports a more conservative sequence: first own the writer path and event fanout, then add render coalescing, then evaluate whether strict in-flight gating is still necessary.

### What warrants a second pair of eyes
- The proposed package layout may want a flatter or more nested structure depending on repository conventions in go-go-golems projects.
- The public API sketch should be reviewed once implementation starts to ensure it matches the team’s preference for root-package versus subpackage exposure.
- The eventual choice of default pacing interval in B-lite will need real hardware validation rather than documentation-only reasoning.

### What should be done in the future
- Implement the root module `github.com/go-go-golems/loupedeck`.
- Port connection/message/display code into the new package.
- Replace single-slot binds with multi-listener subscriptions.
- Add a single outbound writer with configurable pacing.
- Port the feature tester to the new package.
- Add keyed render invalidation and coalescing.
- Only after those steps, decide whether stricter C-style gating is required.

### Code review instructions
- Start with the new design doc:
  - `ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
- Then cross-check the evidence against these files:
  - `sources/loupedeck-repo/display.go`
  - `sources/loupedeck-repo/message.go`
  - `sources/loupedeck-repo/listen.go`
  - `sources/loupedeck-repo/inputs.go`
  - `sources/loupedeck-repo/touchdials.go`
  - `sources/loupedeck-repo/multibutton.go`
  - `ttmp/.../LOUPE-002.../scripts/feature_tester.go`
- Validation commands used in this step:
  - `docmgr ticket list`
  - `docmgr doc list --ticket LOUPE-003`

### Technical details
- The design doc highlights these specific observed issues in the current code:
  - immediate framebuffer + draw send in `display.go`
  - direct websocket writes in `message.go`
  - panic on read errors in `listen.go`
  - incomplete serial close in `dialer.go`
  - single-slot event bindings in `inputs.go`
  - widget-triggered direct redraws in `touchdials.go` and `multibutton.go`
- The guide proposes the following phased abstractions:

```text
Phase B-lite:
  app/widgets -> client enqueue -> single writer -> protocol -> serial websocket -> device

Phase B:
  app/widgets -> invalidate regions -> render scheduler -> single writer -> protocol -> device

Phase C (only if needed later):
  render scheduler -> strict in-flight gate / ack discipline -> device
```

- The intended next package path is:

```text
module github.com/go-go-golems/loupedeck
```

## Step 2: Relate evidence, validate the ticket, and upload the bundle to reMarkable

After the primary design guide was written, I completed the ticket hygiene work so the deliverable could be handed off cleanly. This step focused on docmgr relationships, vocabulary validation, and the reMarkable bundle upload.

This is an important part of the workflow because the value of a detailed architecture document drops quickly if it is not easy to discover later, not checked by the local docs tooling, or not actually delivered to the reading device the user asked for.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the documentation workflow properly by wiring the ticket, validating it, and delivering the bundle to reMarkable.

**Inferred user intent:** Make the new ticket not just exist, but be complete, navigable, and available for offline reading/review.

**Commit (code):** N/A — documentation and delivery phase only.

### What I did
- Added `docmgr doc relate` links from the design doc to the main evidence files in:
  - `display.go`
  - `message.go`
  - `listen.go`
  - `inputs.go`
  - `touchdials.go`
  - `multibutton.go`
  - `dialer.go`
  - `feature_tester.go`
- Added `docmgr doc relate` links from the diary to the design doc and key evidence files.
- Added ticket-level related files on the LOUPE-003 index.
- Ran:
  - `docmgr doctor --ticket LOUPE-003 --stale-after 30`
- Saw a vocabulary warning for the `backpressure` topic and fixed it with:
  - `docmgr vocab add --category topics --slug backpressure --description "Outbound transport pacing, flow control, queueing, and render coalescing concerns in device or protocol integrations."`
- Re-ran doctor successfully.
- Confirmed reMarkable tooling with:
  - `remarquee status`
  - `remarquee cloud account --non-interactive`
- Performed bundle dry-run with:
  - `remarquee upload bundle --dry-run ... --name "LOUPE-003 Backpressure-safe go-go-golems loupedeck package refactor" --remote-dir "/ai/2026/04/11/LOUPE-003" --toc-depth 2`
- Uploaded the real bundle with the same arguments minus `--dry-run`.
- Verified the uploaded document with:
  - `remarquee cloud ls "/ai/2026/04/11/LOUPE-003/" --long --non-interactive`
- Corrected an accidental task-check mistake by editing `tasks.md` directly so the ticket reflects actual completion status.

### Why
- File relationships make the design doc much easier to revisit with docmgr search and related-file navigation.
- Doctor validation catches metadata drift before the ticket is treated as complete.
- The dry-run first/upload second pattern reduces the chance of pushing the wrong bundle to reMarkable.
- Correct task state matters because this ticket is expected to continue into real package implementation work later.

### What worked
- `docmgr doc relate` updated the design doc, diary, and ticket index successfully.
- Adding the `backpressure` topic to the vocabulary resolved the only doctor warning.
- `remarquee upload bundle` succeeded and reported:

```text
OK: uploaded LOUPE-003 Backpressure-safe go-go-golems loupedeck package refactor.pdf -> /ai/2026/04/11/LOUPE-003
```

- Final cloud verification showed:

```text
[f]    LOUPE-003 Backpressure-safe go-go-golems loupedeck package refactor
```

### What didn't work
- My first attempt to verify the upload used a slightly different cloud path form and returned:

```text
Error: no matches for 'LOUPE-003'
```

- A separate `remarquee cloud find` attempt returned:

```text
Error: start directory doesn't exist
```

- The fix was to list the exact directory path with a trailing slash:

```bash
remarquee cloud ls "/ai/2026/04/11/LOUPE-003/" --long --non-interactive
```

- I also accidentally checked the wrong docmgr task IDs once. I corrected `tasks.md` directly afterward.

### What I learned
- The local docmgr vocabulary now needs an explicit `backpressure` topic for future tickets/docs in this area.
- `remarquee cloud ls` path matching can be a little finicky; using the exact folder path with a trailing slash was the reliable form here.
- Ticket bookkeeping should be treated as a real deliverable step, not a postscript.

### What was tricky to build
- The trickiest part of this step was not technical complexity but tool ergonomics: both docmgr and remarquee required a little iteration around exact identifiers and path forms. Because the workflow spans docs, metadata, and cloud delivery, small mismatches can make it look like work failed even when the underlying artifact exists.
- The solution was to verify each stage explicitly: first doctor output, then dry-run upload plan, then actual upload confirmation, then exact-path cloud listing.

### What warrants a second pair of eyes
- It may be worth standardizing a reusable upload/verification command template for ticket bundles so path mismatches happen less often.
- If the team expects many backpressure-related docs, the vocabulary and topic guidance should probably be reviewed once and then reused consistently.

### What should be done in the future
- Start code implementation for Phase 0 in the new top-level module.
- Add a commit covering the LOUPE-003 ticket docs and bookkeeping if/when the user wants the repo state recorded.
- Reuse the upload template for future tickets to keep delivery consistent.

### Code review instructions
- Review the finished deliverables here:
  - `ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
  - `ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/reference/01-investigation-diary.md`
  - `ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/tasks.md`
- Validate with:
  - `docmgr doctor --ticket LOUPE-003 --stale-after 30`
  - `remarquee cloud ls "/ai/2026/04/11/LOUPE-003/" --long --non-interactive`

### Technical details
- Final doctor result:

```text
## Doctor Report (1 findings)

### LOUPE-003

- ✅ All checks passed
```

- Final reMarkable location:

```text
/ai/2026/04/11/LOUPE-003/
```

## Step 3: Phase 0 — create the real root module and baseline package port

This step moved the work from ticket-local example modules into a real repository-root Go module. The goal was not yet to improve behavior; it was to establish the code in the place where the next refactor phases will actually live.

I intentionally kept the first code step conservative. Instead of redesigning and porting at the same time, I first copied the current upstream implementation into the root package, updated the module path, and made sure the repository could build as `github.com/go-go-golems/loupedeck` before changing semantics.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Start implementation in the repo root and create the real package module as the base for the later backpressure refactor.

**Inferred user intent:** Stop treating the code as an experiment around a cloned dependency and make this repository the authoritative implementation location.

**Commit (code):** `e99d144` — `"Phase 0: create root go-go-golems loupedeck module"`

### What I did
- Copied the baseline implementation files from `sources/loupedeck-repo/` into the repository root:
  - `connect.go`
  - `dialer.go`
  - `display.go`
  - `displayknob.go`
  - `inputs.go`
  - `intknob.go`
  - `listen.go`
  - `loupedeck.go`
  - `message.go`
  - `multibutton.go`
  - `touchdials.go`
  - `watchedint.go`
  - `loupedeck_test.go`
  - `LICENSE`
- Wrote a new root `go.mod` with:

```text
module github.com/go-go-golems/loupedeck
```

- Added a new root `README.md` describing the package focus and roadmap.
- Updated `loupedeck_test.go` to import `github.com/go-go-golems/loupedeck` instead of the upstream module path.
- Ran:
  - `go mod tidy`
  - `go test ./...`
- Committed the root module baseline.

### Why
- We need a root package before we can apply package-owned architectural fixes like listener fanout and outbound pacing.
- A baseline port isolates future behavior changes. It gives us a stable “ported as-is” reference point before deeper refactoring.

### What worked
- The baseline port compiled immediately after the module-path adjustment.
- `go test ./...` passed at the root module level.
- The repository now has a canonical Go package location for future work.

### What didn't work
- Nothing materially failed in this step. The work was mostly mechanical and low risk.

### What I learned
- The upstream package surface is small enough to port directly without introducing an intermediate shim layer.
- Preserving the baseline first makes the next diffs easier to review because Phase 1 and Phase 2 changes can be read as true behavior changes rather than mixed port+refactor noise.

### What was tricky to build
- The main subtlety was sequencing, not code complexity. It would have been easy to combine the root-module port with the listener/lifecycle refactor, but doing so would have made both review and debugging harder. Keeping Phase 0 narrow produced a cleaner history.

### What warrants a second pair of eyes
- The root package layout is intentionally flat right now. If the team strongly prefers subpackages immediately, that should be decided before Phase 2 grows the API surface.
- The README is intentionally short; it may want expansion once the package behavior diverges more from the upstream baseline.

### What should be done in the future
- Proceed to Phase 1: listener fanout and safe lifecycle handling.
- Keep Phase 2 separate so the B-lite writer queue is reviewed as its own transport change.

### Code review instructions
- Review these root files first:
  - `go.mod`
  - `README.md`
  - `connect.go`
  - `display.go`
  - `listen.go`
- Validate with:
  - `go test ./...`

### Technical details
- Phase 0 established the authoritative package path:

```text
github.com/go-go-golems/loupedeck
```

- This step intentionally did not change transport semantics yet.

## Step 4: Phase 1 — add composable listeners, unsubscribe support, and safe lifecycle behavior

This step introduced the first real behavior changes in the new root package. The key theme was composability: widget code, app code, and diagnostics should be able to observe the same inputs without overwriting each other, and transport read failures should no longer panic the entire process.

I kept backward compatibility where practical. The existing `Bind*` methods still exist as legacy primary bindings, but I added `On*` subscription APIs and moved internal widget code over to them so widgets no longer depend on single-slot binding behavior.

### Prompt Context

**User prompt (verbatim):** "add detailed tasks to ticket. then go ahead. task by task, commit at appropriate intervals, update the diary as you go."

**Assistant interpretation:** Expand the ticket into a concrete implementation checklist, then execute the refactor step-by-step with focused commits and diary updates.

**Inferred user intent:** Build the package incrementally with reviewable history and high-quality continuity notes.

**Commit (code):** `c828131` — `"Phase 1: add composable listeners and safe lifecycle"`

### What I did
- Expanded the LOUPE-003 `tasks.md` into a detailed phased checklist.
- Added `listeners.go` implementing:
  - `Subscription`
  - `OnButton`
  - `OnButtonUp`
  - `OnKnob`
  - `OnTouch`
  - `OnTouchUp`
  - dispatch helpers for button/knob/touch events
- Updated `Loupedeck` in `loupedeck.go` to hold:
  - legacy single-slot bindings
  - new multi-listener registries
  - listener mutex and listener ID state
- Updated `inputs.go` so `Bind*` methods are legacy primary bindings protected by the listener mutex, while new listener maps remain available for composition.
- Changed `listen.go` so:
  - `Listen()` now returns `error`
  - read failures no longer panic
  - inbound events dispatch through the new helper methods
- Updated `dialer.go` so `SerialWebSockConn.Close()` closes the underlying serial port.
- Updated `loupedeck.go` so `Loupedeck.Close()` returns an aggregated `error` rather than silently discarding close failures.
- Ported internal widget code to the new subscription model:
  - `intknob.go` now uses `OnKnob` and `OnButton`
  - `touchdials.go` now uses `OnTouch` and `OnTouchUp`
  - `multibutton.go` now uses `OnTouch`
- Added `listeners_test.go` covering:
  - multiple button listeners
  - unsubscribe cleanup
  - coexistence of `BindKnob` and `OnKnob`
  - touch-up listener cleanup
- Ran:
  - `gofmt -w *.go`
  - `go test ./...`
- Committed the Phase 1 implementation.

### Why
- The current single-slot binding model was one of the biggest structural problems found during LOUPE-002 analysis.
- Package consumers need a way to add logging or app-level behavior without clobbering widget behavior.
- Panicking on malformed or closed reads is not acceptable for a reusable package.
- Proper serial close is important for repeated runs and device recovery.

### What worked
- The listener fanout model compiled cleanly and passed the initial unit tests.
- Existing widget abstractions could be moved to the composable `On*` API with small targeted changes.
- `Listen()` can now fail normally instead of terminating the process via panic.

### What didn't work
- I caught one design bug during implementation before it escaped into tests: the first version of the dispatch helpers called listener callbacks while still holding the read lock. That would have made unsubscribe/register-inside-callback behavior risky or deadlock-prone.
- I fixed that by snapshotting the listeners under lock, unlocking, and only then invoking callbacks.

### What I learned
- Preserving `Bind*` as a legacy primary binding while moving internals to `On*` gives a good compatibility bridge.
- Dispatch helpers make the listener model testable without hardware, which is important for the next transport-focused phases.
- The package is now in a much better position for B-lite because outbound pacing can be added without simultaneously fighting callback overwrite bugs.

### What was tricky to build
- The trickiest part was making the new listener model safe without rewriting the entire public API at once. A hard API break would have made migration noisier, but leaving only `Bind*` would have preserved the same composition problems. The hybrid approach—legacy `Bind*` plus composable `On*`—balances compatibility and forward progress.
- The second tricky point was zero-value safety. The subscription layer now initializes maps as needed so tests and helper instances do not panic on nil maps.

### What warrants a second pair of eyes
- The current listener fanout covers the main button/knob/touch paths, but CT-specific touch behavior still uses the older single callback field. That is acceptable for now, but it should be revisited if CT support becomes a priority.
- `transactionCallbacks` still do not have their own dedicated synchronization strategy. That was not the focus of Phase 1, but it may matter once Phase 2 introduces a writer goroutine.

### What should be done in the future
- Start Phase 2: add the single outbound writer and route sends through it.
- Revisit transaction callback synchronization as part of the writer refactor.
- Port the feature tester later so it uses `On*` rather than relying on legacy `Bind*` semantics.

### Code review instructions
- Start with:
  - `listeners.go`
  - `listeners_test.go`
  - `listen.go`
  - `inputs.go`
  - `loupedeck.go`
  - `dialer.go`
- Then verify the internal widget migration in:
  - `intknob.go`
  - `touchdials.go`
  - `multibutton.go`
- Validate with:
  - `gofmt -w *.go`
  - `go test ./...`

### Technical details
- New composable listener API added in Phase 1:

```go
func (l *Loupedeck) OnButton(b Button, f ButtonFunc) Subscription
func (l *Loupedeck) OnButtonUp(b Button, f ButtonFunc) Subscription
func (l *Loupedeck) OnKnob(k Knob, f KnobFunc) Subscription
func (l *Loupedeck) OnTouch(b TouchButton, f TouchFunc) Subscription
func (l *Loupedeck) OnTouchUp(b TouchButton, f TouchFunc) Subscription
```

- Key lifecycle changes:

```go
func (l *Loupedeck) Listen() error
func (l *Loupedeck) Close() error
func (l *SerialWebSockConn) Close() error
```
