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
    - Path: cmd/loupe-feature-tester/main.go
      Note: Diary records the root-level feature tester migration
    - Path: cmd/loupe-fps-bench/main.go
      Note: Diary records the raw display throughput benchmark harness and measured FPS ceilings
    - Path: display.go
      Note: Diary records logical draw command grouping
    - Path: go.mod
      Note: Diary records the root module creation
    - Path: listen.go
      Note: Diary records the lifecycle change from panic to error
    - Path: listeners.go
      Note: Diary records the composable listener implementation
    - Path: listeners_test.go
      Note: Diary records the Phase 1 validation tests
    - Path: message.go
      Note: Diary records Send/SendWithCallback routing through the writer
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
    - Path: writer.go
      Note: Diary records the B-lite writer implementation
    - Path: writer_test.go
      Note: Diary records ordering and pacing tests
ExternalSources: []
Summary: Chronological diary for the package refactor analysis and implementation, including lifecycle fixes, hardware validation, and raw display throughput benchmarking.
LastUpdated: 2026-04-11T19:11:59-04:00
WhatFor: Preserve the exact analysis path, commands, commits, and hardware evidence that shaped the evolving go-go-golems Loupedeck package.
WhenToUse: Use when resuming work on the package refactor or when reviewing why B-lite came first, how reconnect issues were diagnosed, and what the current hardware throughput limits look like.
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

## Step 5: Phase 2 — add the B-lite outbound writer, pacing, and transport tests

This step introduced the first backpressure-specific transport control in the new package. The objective was deliberately modest: do not build a full coalescing renderer yet, but do centralize outbound websocket writes behind a single queue-owning writer with configurable pacing and basic metrics.

The implementation keeps `Send()` synchronous from the caller’s point of view. That is important because the existing package behavior assumes that `Send()` reports transport errors directly. Internally, however, the actual websocket writes now happen in one goroutine, which gives us an owned place to enforce ordering, insert pacing, and collect stats.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue into the next implementation task and land the first package-owned backpressure mechanism as a focused commit.

**Inferred user intent:** Replace ad hoc or implicit write behavior with a reviewable B-lite transport layer before moving to more advanced rendering changes.

**Commit (code):** `a3056ae` — `"Phase 2: add B-lite outbound writer and pacing"`

### What I did
- Added `writer.go` implementing:
  - `WriterOptions`
  - `DefaultWriterOptions`
  - `WriterStats`
  - outbound command interface
  - single-message command wrapper
  - queue-owned `outboundWriter`
  - pacing via configurable `SendInterval`
  - synchronous enqueue-and-wait semantics for callers
- Introduced `wsConn` as a narrow connection interface so writer behavior can be tested without real hardware.
- Updated `Loupedeck` in `loupedeck.go` to hold:
  - `conn wsConn`
  - `writer *outboundWriter`
  - `writerOptions WriterOptions`
- Added `WriterStats()` on `Loupedeck`.
- Updated `connect.go` to add:
  - `ConnectAutoWithOptions`
  - `ConnectPathWithOptions`
  - root connections now start the writer using the configured options
  - legacy `ConnectAuto` / `ConnectPath` now delegate to defaults
- Updated `message.go` so:
  - `Send()` and `SendWithCallback()` route through `EnqueueCommand()`
  - transaction callback map access is synchronized through helper methods
- Updated `display.go` so a display draw now enqueues a single logical `displayDrawCommand` containing both:
  - `WriteFramebuff`
  - `Draw`
- Added `writer_test.go` covering:
  - send ordering
  - pacing interval behavior
  - display draw command ordering (`WriteFramebuff` before `Draw`)
- Ran:
  - `gofmt -w *.go`
  - `go test ./...`
- Committed the Phase 2 implementation.

### Why
- B-lite needed to move write ownership into one place before any higher-level render work can be trusted.
- Keeping `Send()` synchronous preserves compatibility while still giving the package a single writer goroutine underneath.
- The new writer also creates a natural insertion point for future queue metrics, throttling, coalescing, or stricter flow-control experiments.

### What worked
- The package compiled cleanly after the writer refactor.
- The writer tests passed without needing real hardware.
- `Display.Draw()` now routes as one logical command through the writer instead of two independent ad hoc sends.
- The new connect helpers make send interval and queue size configurable without breaking the old API.

### What didn't work
- No external failures occurred in this step, but I did tighten one implementation detail during the refactor: transaction callback registration and retrieval were previously unsynchronized relative to the read loop. Since the writer introduced clearer transport ownership, it made sense to clean that up in the same phase.

### What I learned
- A narrow connection interface (`wsConn`) makes transport behavior much easier to test than hard-coding `*websocket.Conn` everywhere.
- The B-lite writer can preserve the existing synchronous `Send()` contract while still serializing outbound traffic internally.
- Grouping framebuffer and draw into one command is a good transitional abstraction even before full dirty-region coalescing exists.

### What was tricky to build
- The main design tension was between compatibility and ownership. If `Send()` became fully asynchronous immediately, it would change error-reporting behavior at the same time that we are trying to stabilize the transport. The chosen design—queue internally, but wait for the queued command result—keeps the surface area stable while still moving the actual write side under one goroutine.
- Another subtle point was deciding the unit of pacing. For B-lite, pacing is applied per outbound command, not per individual protocol message. That means a logical display draw can still write framebuffer + draw back-to-back, but multiple draw commands are spaced. This matches the current working hypothesis about where the device gets overwhelmed.

### What warrants a second pair of eyes
- The default pacing interval (`25ms`) is a reasonable starting point, but it has not yet been hardware-validated against the current feature tester.
- The writer currently provides coarse stats and simple queue behavior. If queue saturation becomes an issue, we may want more explicit backpressure/error semantics rather than a buffered-channel wait.
- Once the feature tester is ported, we should inspect whether any draw patterns still create unacceptable traffic bursts even with the single writer in place.

### What should be done in the future
- Start Phase 3 by porting the feature tester to the new root package.
- Run hardware tests with B-lite to collect send-rate and stability evidence.
- Use those results to decide how aggressively full B should coalesce display invalidations.

### Code review instructions
- Start with:
  - `writer.go`
  - `writer_test.go`
  - `message.go`
  - `display.go`
  - `connect.go`
- Then review the root type changes in:
  - `loupedeck.go`
- Validate with:
  - `gofmt -w *.go`
  - `go test ./...`

### Technical details
- New B-lite transport API pieces introduced in this step:

```go
type WriterOptions struct {
    QueueSize    int
    SendInterval time.Duration
}

func ConnectAutoWithOptions(writerOptions WriterOptions) (*Loupedeck, error)
func ConnectPathWithOptions(serialPath string, writerOptions WriterOptions) (*Loupedeck, error)
func (l *Loupedeck) EnqueueCommand(cmd outboundCommand) error
func (l *Loupedeck) WriterStats() WriterStats
```

- The display path now groups the two protocol messages for one draw as one logical outbound command.

## Step 6: Phase 3 — add a root-level feature tester command using the new package

This step moved the feature tester out of the ticket-local script module and into the root repository as a real command that consumes the new package. That migration matters because the package changes from Phases 1 and 2 are most valuable when an actual application uses them the way a future user would.

The migrated command also exercises the new composable listener model directly. Unlike the old feature tester, it can log knob/touch/button events while the widgets continue to function, because those behaviors now compose through `On*` subscriptions instead of fighting over `Bind*` slots.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue into the next ticket task by moving the feature tester to the new root package and rely on the package-owned pacing instead of application sleeps.

**Inferred user intent:** Prove that the refactor is usable from a real command in the repository root, not only from isolated internal tests.

**Commit (code):** `8bd9e43` — `"Phase 3: add root feature tester command"`

### What I did
- Added a new root command:
  - `cmd/loupe-feature-tester/main.go`
- Ported the current feature tester behavior into the command while updating it to use the new package APIs:
  - `ConnectAutoWithOptions(...)`
  - `OnKnob(...)`
  - `OnTouch(...)`
  - `OnTouchUp(...)`
  - `OnButton(...)`
  - `OnButtonUp(...)`
- Removed the old application-level `time.Sleep()` pacing used during MultiButton setup.
- Used package-owned pacing instead by supplying `WriterOptions` during connect.
- Preserved the main interactive feature set:
  - left/right TouchDial strips
  - 12 MultiButtons on the main display
  - touch flash overlays on press
  - button LED color cycling
  - Circle button exit path
  - event logging
- Started the listener loop with explicit error handling:

```go
listenErrCh := make(chan error, 1)
go func() {
    listenErrCh <- l.Listen()
}()
```

- Logged writer stats at startup and shutdown.
- Ran:
  - `gofmt -w cmd/loupe-feature-tester/main.go`
  - `go test ./...`
- Committed the new command.

### Why
- We needed at least one real application in the root module to validate that the refactored package API is usable.
- The migrated feature tester is the best practical bridge between package-only tests and later hardware validation.
- Removing app-level sleeps is one of the clearest ways to show that pacing ownership has moved into the package.

### What worked
- The new command compiled cleanly as `github.com/go-go-golems/loupedeck/cmd/loupe-feature-tester`.
- `go test ./...` continued to pass after adding the command.
- The new package API was expressive enough to port the tester without falling back to the old single-slot binding model.

### What didn't work
- No build or test failures occurred in this step.
- This step did **not** yet include real hardware validation, so the remaining open question is behavioral rather than compilation-related: we still need to see how the migrated command behaves on the actual device with B-lite pacing.

### What I learned
- The new `On*` listener model materially improves application ergonomics. The feature tester can now subscribe for logging/flash/exit behavior without disrupting the internal widget bindings.
- The package-owned writer options are a better place to express pacing intent than scattered sleeps in app code.
- The root command now gives us a concrete harness for the decision gate about whether C is necessary later.

### What was tricky to build
- The main challenge was deciding how much cleanup to do while porting. I kept the command close enough to the previous feature tester to preserve its role as a hardware exercise, but I updated the event wiring to take advantage of the new listener model. That balance keeps the migration understandable while still demonstrating the value of the refactor.
- The other subtle point was handling `Listen()` now that it returns an error. The command needs to treat the read loop as a goroutine with an explicit error path rather than as an infinite background side effect.

### What warrants a second pair of eyes
- The chosen writer interval in the command (`40ms`) is still a hypothesis until hardware testing confirms it is stable enough and responsive enough.
- The command currently logs writer stats but does not yet expose deeper diagnostics or render metrics; that may be useful once full B starts.

### What should be done in the future
- Run the new root-level feature tester on real hardware.
- Compare its behavior with the earlier ticket-local feature tester.
- Use those observations to decide whether to proceed directly to full B or whether any transport tuning is needed first.

### Code review instructions
- Start with:
  - `cmd/loupe-feature-tester/main.go`
- Cross-check the APIs it depends on in:
  - `writer.go`
  - `listeners.go`
  - `connect.go`
- Validate with:
  - `gofmt -w cmd/loupe-feature-tester/main.go`
  - `go test ./...`

### Technical details
- The command now depends on the root package instead of a ticket-local module.
- The most important migration difference is that pacing is package-owned via `WriterOptions`, not enforced in the app with `Sleep` calls.

## Step 7: Phase 4 — add render invalidation and region coalescing on top of the writer

After the root command was in place, I moved on to the first part of full B: coalescing repeated display updates before they hit the outbound writer. This does not replace the writer; it sits above it. The writer still owns transport ordering and pacing, while the renderer owns the question “which display state is still worth sending?”.

This phase focused specifically on display traffic. It does not yet try to coalesce arbitrary protocol commands. The concrete change is that `Display.Draw()` no longer always forwards a draw command immediately. When the renderer is present, it turns the draw into a keyed invalidation, and the scheduler flushes only the latest pending draw per region key at each tick.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue past B-lite and implement the first real full-B mechanism: keyed render invalidation and coalescing.

**Inferred user intent:** Reduce redundant display traffic, especially for repeated knob-strip redraws and touch flash/repaint sequences, without giving up the new transport ownership model.

**Commit (code):** `c6cd4e2` — `"Phase 4: add render scheduler and region coalescing"`

### What I did
- Added `renderer.go` implementing:
  - `RenderOptions`
  - `DefaultRenderOptions`
  - `RenderStats`
  - `renderScheduler`
  - keyed invalidation queue
  - periodic flush loop
  - replacement/coalescing metrics
- Updated `Loupedeck` to hold:
  - `renderer *renderScheduler`
  - `renderOptions RenderOptions`
- Added `RenderStats()` accessor on `Loupedeck`.
- Updated `connect.go` so new connections start the render scheduler alongside the writer.
- Updated `display.go` so `Display.Draw()` now:
  - builds the same framebuffer + draw command as before,
  - computes a region key from display name and geometry,
  - sends the command through the render scheduler when enabled,
  - falls back to direct enqueue if no renderer exists.
- Added `renderer_test.go` covering:
  - repeated invalidations to the same region,
  - “latest image wins” behavior,
  - coalesced replacement counts,
  - flushed command counts.
- Ran:
  - `gofmt -w *.go cmd/loupe-feature-tester/main.go`
  - `go test ./...`
- Committed the render scheduler groundwork.

### Why
- The B-lite writer solved ownership and pacing, but it still allowed every draw request to become a queued command.
- Knob-strip redraws and touch flash/restore sequences naturally create overwriteable intermediate states.
- Region-keyed invalidation is the smallest useful full-B mechanism because it reduces redundant display work without requiring a full scene graph or a large widget rewrite.

### What worked
- The package continued to compile and test cleanly after the renderer was inserted above the writer.
- The renderer test confirmed that repeated invalidations to the same region collapse to one flushed command and that the last image survives.
- The change required only a small surface-area modification to `Display.Draw()`, which means existing widgets benefit automatically.

### What didn't work
- No implementation failures occurred in this step.
- The current renderer only coalesces by region key and flush interval. It does not yet implement priorities, explicit flush-now semantics, or broader retained-mode composition.

### What I learned
- The existing `Display.Draw()` API is a surprisingly good interception point for the first full-B pass. Because all current widgets already flow through it, we can gain meaningful coalescing without redesigning every widget first.
- The combination of renderer + writer gives a nice layered model:

```text
widgets/app -> Display.Draw -> renderer invalidation -> writer pacing -> websocket -> device
```

- This is already enough to reduce redundant strip/tile traffic in theory; the next question is how it behaves on real hardware.

### What was tricky to build
- The main design choice was deciding the coalescing key. I used `display name + x + y + width + height`, which is simple and fits the current widget layout well. It is not the only possible keying strategy, but it is concrete, deterministic, and sufficient for the existing touch-grid and strip-based UI.
- Another subtle point was layering. The renderer should not replace the writer or duplicate pacing logic. Its job is to decide what to flush, not how the websocket is paced.

### What warrants a second pair of eyes
- The default render flush interval (`40ms`) may need tuning alongside the writer interval after more real-device testing.
- If we later introduce explicit widget-owned region names instead of geometry-derived keys, we should make sure migration is straightforward.

### What should be done in the future
- Re-run the hardware tester with the renderer-enabled package.
- Compare the observed behavior with the earlier B-lite-only run.
- Decide whether reconnect/reset behavior is a separate lifecycle problem or a sign that stricter C-style flow control is still needed.

### Code review instructions
- Start with:
  - `renderer.go`
  - `renderer_test.go`
  - `display.go`
- Then verify the root type and connection wiring in:
  - `loupedeck.go`
  - `connect.go`
- Validate with:
  - `gofmt -w *.go cmd/loupe-feature-tester/main.go`
  - `go test ./...`

### Technical details
- The renderer currently coalesces by a key shaped like:

```text
<display-name>:<x>:<y>:<width>:<height>
```

- Each flush still sends commands through the existing B-lite writer, so transport pacing remains centralized.

## Step 8: Hardware smoke test results after B-lite and after the first full-B groundwork

I ran the root-level feature tester against the reconnected Loupedeck Live to gather evidence from the real device instead of relying only on unit tests. The result was mixed in an informative way: the first run came up cleanly and exercised knobs and touch without websocket failures in the observed window, while a later reconnect attempt failed during the websocket handshake with a malformed HTTP response.

That split is useful because it suggests the steady-state draw storm problem is improving, but device reset/reconnect state is still fragile, especially after non-clean process termination or rapid reruns.

### Prompt Context

**User prompt (verbatim):** "go aheaed."

**Assistant interpretation:** Continue with the next practical step after implementation by exercising the new root-level command against real hardware.

**Inferred user intent:** Validate the package changes on the actual Loupedeck and use that evidence to guide whether C is necessary.

**Commit (code):** N/A — hardware validation and documentation step.

### What I did
- First attempted a root-level feature tester run before reconnect and saw:

```text
connection failed: unable to open port "/dev/ttyACM0"
```

- After the user reconnected the device, ran:

```bash
timeout 10s go run ./cmd/loupe-feature-tester
```

- Observed a clean startup with no websocket crash during initialization.
- Observed successful runtime events in the log, including:
  - MultiButton initialization
  - writer stats at readiness
  - knob value updates
  - raw knob delta logs
  - touch press/release logs
  - MultiButton state cycling
- Then, after the coalescing scheduler was added, ran a short smoke test again with:

```bash
timeout 8s go run ./cmd/loupe-feature-tester
```

- That second later run failed at connect with a malformed HTTP response, indicating the device/websocket handshake state is still vulnerable after the prior run.

### Why
- This was the first real-device check of the new root-level package and command.
- The decision about C should be based on actual hardware behavior, not just architecture taste.

### What worked
- The first real run after reconnect started cleanly and remained stable long enough to exercise knobs and touch without the earlier `bad opcode 4` / `FIN not set on control` style crash during the observed interval.
- The ready log showed healthy writer behavior at startup:

```text
Feature tester ready writer_stats="{QueuedCommands:18 SentCommands:18 SentMessages:32 FailedCommands:0 MaxQueueDepth:1}"
```

- The runtime log clearly showed working behavior such as:
  - knob value tracking (`[KNOB N] value=...`)
  - raw delta logging (`raw_event=true`)
  - touch press/release logging
  - MultiButton state transitions

### What didn't work
- The later rerun after the coalescing implementation failed during connect with:

```text
malformed HTTP response "\x82\x05\x05\x00\x00\b\x00\x82\x05\x05\x00\x00\t\x01\x82\x05\x05\x00\x00"
```

- That is consistent with the older observation that the device can remain in a bad websocket/serial state after an abrupt stop or a quick rerun.
- Because the test command was run under `timeout`, the process did not get a graceful device-side shutdown path. That means the reconnect failure is not strong evidence against the new pacing/coalescing architecture by itself.

### What I learned
- B-lite plus the first full-B groundwork appear to have improved the steady-state startup/runtime behavior enough that the feature tester can initialize and handle interactive events without immediately tripping the old websocket failure during the observed run.
- The remaining fragility appears strongly tied to reconnect/reset behavior after abrupt process termination.
- That means the current case for C is weaker than before. Right now, the bigger open problem is likely device reset/handshake hygiene rather than insufficient steady-state flow control.

### What was tricky to build
- The main difficulty in interpreting the hardware results is separating three different failure classes:
  1. burst-induced transport instability during normal operation,
  2. reconnect/handshake corruption after abrupt stop,
  3. ordinary device-open failures when the serial port is unavailable.
- The logs now suggest class (1) is improving, while class (2) still needs explicit lifecycle attention.

### What warrants a second pair of eyes
- We should do at least one clean-exit hardware cycle—ideally exiting with the Circle button rather than `timeout`—before deciding too strongly about C.
- It may be worth adding an explicit device-reset/recovery helper or startup drain step if reconnect instability remains common.

### What should be done in the future
- Re-run the tester with a clean exit path to separate abrupt-stop effects from normal reconnect behavior.
- Decide whether to open a dedicated lifecycle/reset ticket if reconnect instability remains after clean exits.
- Keep C undecided for now; current evidence does not justify it as the immediate next move.

### Code review instructions
- Review the new runtime harness:
  - `cmd/loupe-feature-tester/main.go`
- Review the new transport stack:
  - `writer.go`
  - `renderer.go`
  - `display.go`
- Hardware validation commands used in this step:

```bash
timeout 10s go run ./cmd/loupe-feature-tester
timeout 8s go run ./cmd/loupe-feature-tester
```

### Technical details
- The most important positive hardware signal from the successful run was:

```text
QueuedCommands:18 SentCommands:18 SentMessages:32 FailedCommands:0 MaxQueueDepth:1
```

- The most important remaining negative hardware signal was the malformed HTTP response on a later rerun, which points more toward reconnect state corruption than immediate draw-storm failure.

## Step 9: Write the technical deep-dive project report article

After the implementation and diary phases had reached a useful stopping point, I wrote a separate project-report-style technical deep dive focused on the new frontend implementation itself. The goal of this document is different from the design guide and different from the diary. The design guide explains the plan; the diary explains the sequence of work; the deep-dive article explains the resulting implementation as a system.

This document is aimed at readers who want a cohesive technical narrative about the new frontend—its layers, algorithms, runtime contracts, and engineering tradeoffs—without reading through the entire ticket history first.

### Prompt Context

**User prompt (verbatim):** "after the diary, write a detailed project report technical deep dive article about the new implementation of the loupedeck go frontend and especially the algorithms and updates"

**Assistant interpretation:** Add a new long-form document in the ticket that reads like a project report and technical deep dive into the new frontend implementation, with emphasis on algorithms and implementation updates.

**Inferred user intent:** Capture the state of the new implementation in a form that is useful for engineers, reviewers, and future readers who want the architecture explained clearly in one place.

**Commit (code):** N/A — documentation step.

### What I did
- Added a new LOUPE-003 design document:
  - `design-doc/02-technical-deep-dive-the-new-go-go-golems-loupedeck-frontend-implementation.md`
- Wrote a long-form report covering:
  - the shift from experiment to root package
  - listener fanout algorithm
  - error-returning lifecycle model
  - single-writer pacing algorithm
  - grouped display draw commands
  - render invalidation and coalescing algorithm
  - transaction callback synchronization
  - widget behavior on top of the new stack
  - root feature tester as a frontend harness
  - hardware observations and implications
  - what remains intentionally incomplete
- Structured the article as a technical report rather than as a task plan.

### Why
- The design guide is implementation-oriented and phase-based; the deep dive is architecture-oriented and explanatory.
- The diary is chronological; the deep dive is conceptual.
- Having all three forms of documentation makes the ticket more useful for different audiences.

### What worked
- The article format fit well with the current implementation state.
- There was enough completed code to write concretely about the frontend as it exists now rather than only speculatively.

### What didn't work
- Nothing failed technically in this step.
- The main challenge was avoiding duplication between the design guide and the new article while still making the new piece self-contained.

### What I learned
- The package is now mature enough that the frontend can be described in terms of layered runtime algorithms, not only as a set of helper functions.
- The best way to explain the implementation is to treat listener fanout, writer pacing, and render invalidation as three distinct but cooperating algorithms.

### What was tricky to build
- The tricky part was choosing the right narrative structure. A project report should not read like a task list, but it also must stay anchored to real code and real runtime behavior. The solution was to organize the article around the runtime pipeline and its algorithms rather than around ticket chronology.

### What warrants a second pair of eyes
- The deep-dive article should be reviewed after future render/reconnect work to keep the “intentionally incomplete” section honest.
- If the package layout changes significantly, the article’s repository-shape discussion may need refreshing.

### What should be done in the future
- Keep the deep dive updated as full B evolves beyond the current groundwork.
- Consider uploading a refreshed bundle to reMarkable once enough additional implementation has accumulated.

### Code review instructions
- Read the new article:
  - `ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/02-technical-deep-dive-the-new-go-go-golems-loupedeck-frontend-implementation.md`
- Cross-check the implementation claims against:
  - `listeners.go`
  - `writer.go`
  - `renderer.go`
  - `display.go`
  - `cmd/loupe-feature-tester/main.go`

### Technical details
- The article complements, rather than replaces:
  - the primary design guide (`design-doc/01-...`)
  - the chronological diary (`reference/01-investigation-diary.md`)

## Step 10: Fix connect callback shape assumptions, repair timeout retry lifecycles, and validate a clean-exit hardware cycle

After the first successful root-package hardware smoke run, I pushed one layer deeper into the remaining instability and found that the next failures were no longer the original draw-storm problem. Instead, they clustered around connection lifecycle edges: short async callback payloads during the connect handshake, and retry logic that re-used a serial/websocket connection after a timeout instead of reopening a fresh one.

This was a useful shift in failure mode. It meant the package had become stable enough in the steady state that the next bugs were now connect/reconnect hygiene problems. I fixed those concretely, then reran the root feature tester in the background and exited it cleanly with the Circle button so I could separate abrupt-stop corruption from normal runtime behavior.

### Prompt Context

**User prompt (verbatim):** "The conversation history before this point was compacted into the following summary:

<summary>
## Goal
Create a comprehensive Loupedeck Live hardware feature tester program that exercises all device capabilities: 6 knob encoders, physical button presses, 4×3 touchscreen, scrolling sliders (TouchDial), round button LED colors, and individual touch button icons. Program must include proper docmgr ticket structure with analysis diary and be tested on actual hardware.

Expanded goal: turn the experimental work into a real root Go package, `github.com/go-go-golems/loupedeck`, with package-owned backpressure handling, composable listeners, safer lifecycle behavior, a root feature tester command, detailed ticket docs, and an Obsidian technical deep dive.

## Constraints & Preferences
- Use `github.com/scottlaird/loupedeck` Go library over USB serial (firmware 2.x) as the upstream reference, but now work in the repo root as a real package
- Must work on Loupedeck Live (product ID 0004)
- Include rate-limiting / backpressure control to prevent `\"websocket: bad opcode 4\"` and related protocol errors
- Touchscreen button presses should flash corresponding **screen button areas** (4×3 grid), not physical LEDs
- Maintain docmgr ticket structure with design docs, diary, changelog, tasks, related files
- Commit incrementally at appropriate intervals and update the diary as work proceeds
- Upload major ticket docs to reMarkable when requested
- Save project/deep-dive notes into the Obsidian vault
- Keep exact file references and evidence-backed analysis
- Root package should build with `go test ./...`

## Progress
### Done
- [x] Created LOUPE-001 ticket with minimal hello_world.go (tested on hardware)
- [x] Created LOUPE-002 ticket for comprehensive feature tester
- [x] Implemented feature_tester.go with all 6 hardware features
- [x] Fixed WebSocket error by adding 50ms delays between MultiButtons + 100ms after batch
- [x] Fixed encoder value updates by removing double-binding (TouchDial already creates IntKnobs internally)
- [x] Added touch-to-flash mapping: Touch 1-4 → Circle/Btn1-3, Touch 5-8 → Btn4-7, Touch 9-12 → cycles back
- [x] Successfully tested on actual Loupedeck Live hardware - all features working
- [x] Committed all changes to git (earlier LOUPE-001/002 work)
- [x] Wrote LOUPE-002 postmortem and detailed diary:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-002--loupedeck-live-feature-tester-comprehensive-hardware-exercise/design-doc/02-postmortem.md`
  - `/home/manuel/code/wesen/2026/04/11/LOUPE-002--loupedeck-live-feature-tester-comprehensive-hardware-exercise/reference/02-detailed-diary.md`
- [x] Updated Obsidian project note with LOUPE-002 status and WebSocket limitations:
  - `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/11/PROJ - Loupedeck Live Hello World - Serial Go Driver.md`
- [x] Analyzed LOUPE-002 backpressure problem and identified structural issues in upstream library:
  - immediate draw sends in `sources/loupedeck-repo/display.go`
  - direct send path in `sources/loupedeck-repo/message.go`
  - panic in `sources/loupedeck-repo/listen.go`
  - single-slot bindings in `sources/loupedeck-repo/inputs.go`
  - incomplete close in `sources/loupedeck-repo/dialer.go`
- [x] Created new ticket LOUPE-003:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/`
- [x] Wrote primary LOUPE-003 architecture/design guide:
  - `.../design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
- [x] Wrote and maintained LOUPE-003 diary:
  - `.../reference/01-investigation-diary.md`
- [x] Uploaded LOUPE-003 bundle to reMarkable:
  - remote dir `/ai/2026/04/11/LOUPE-003/`
  - document `LOUPE-003 Backpressure-safe go-go-golems loupedeck package refactor`
- [x] Added `backpressure` to docmgr vocabulary and passed `docmgr doctor --ticket LOUPE-003 --stale-after 30`
- [x] Phase 0: created root module `github.com/go-go-golems/loupedeck`
  - root files copied from upstream reference into repo root
  - root `README.md` added
  - root `go.mod` and `go.sum` added
  - `go test ./...` passes at root
- [x] Phase 1: added composable listeners and safe lifecycle
  - new file `/home/manuel/code/wesen/2026-04-11--loupedeck-test/listeners.go`
  - new tests `/home/manuel/code/wesen/2026-04-11--loupedeck-test/listeners_test.go`
  - `Listen()` now returns `error` instead of panicking
  - `SerialWebSockConn.Close()` now closes the actual port
  - widgets (`intknob.go`, `touchdials.go`, `multibutton.go`) now use `On*` APIs
- [x] Phase 2: added B-lite outbound writer and pacing
  - new file `/home/manuel/code/wesen/2026-04-11--loupedeck-test/writer.go`
  - new tests `/home/manuel/code/wesen/2026-04-11--loupedeck-test/writer_test.go`
  - added `WriterOptions`, `WriterStats`, `ConnectAutoWithOptions`, `ConnectPathWithOptions`
  - all outbound websocket writes now go through a single writer goroutine
  - `display.go` now groups `WriteFramebuff` + `Draw` into one logical `displayDrawCommand`
- [x] Phase 3: added root feature tester command
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-feature-tester/main.go`
  - uses new root package APIs (`On*`, `ConnectAutoWithOptions`)
  - removed old app-level sleep-based pacing from the tester
  - `go test ./...` passes with command present
- [x] Hardware smoke test with B-lite/new root command succeeded after reconnect:
  - `timeout 10s go run ./cmd/loupe-feature-tester`
  - observed clean startup and runtime logs without immediate websocket failure during the observed window
  - writer stats logged: `QueuedCommands:18 SentCommands:18 SentMessages:32 FailedCommands:0 MaxQueueDepth:1`
- [x] Phase 4: added render scheduler and region coalescing (full-B groundwork)
  - new file `/home/manuel/code/wesen/2026-04-11--loupedeck-test/renderer.go`
  - new tests `/home/manuel/code/wesen/2026-04-11--loupedeck-test/renderer_test.go`
  - `Display.Draw()` now routes through keyed invalidation when renderer is enabled
  - region key format: `<display-name>:<x>:<y>:<width>:<height>`
  - `go test ./...` passes
- [x] Wrote LOUPE-003 technical deep dive article in ticket:
  - `.../design-doc/02-technical-deep-dive-the-new-go-go-golems-loupedeck-frontend-implementation.md`
- [x] Saved deep-dive article to Obsidian vault:
  - `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/11/ARTICLE - Loupedeck - Backpressure-Safe Go Frontend Deep Dive.md`
- [x] Maintained incremental commit history for LOUPE-003:
  - `1c706c4 Add LOUPE-003 package refactor ticket and task breakdown`
  - `e99d144 Phase 0: create root go-go-golems loupedeck module`
  - `c828131 Phase 1: add composable listeners and safe lifecycle`
  - `0bd7fc9 Update LOUPE-003 diary and bookkeeping for Phase 0 and Phase 1`
  - `a3056ae Phase 2: add B-lite outbound writer and pacing`
  - `1ed5c25 Update LOUPE-003 diary and bookkeeping for Phase 2`
  - `8bd9e43 Phase 3: add root feature tester command`
  - `9b12ff6 Update LOUPE-003 diary and bookkeeping for Phase 3`
  - `c6cd4e2 Phase 4: add render scheduler and region coalescing`

### In Progress
- [ ] LOUPE-003 Phase 5 documentation/continuity bookkeeping after latest Phase 4 + deep-dive work
- [ ] Decision gate for whether strict C-style in-flight / ack-gated flow control is needed
- [ ] Additional clean-exit hardware validation to separate reconnect/reset issues from steady-state transport behavior

### Blocked
- Reconnect / handshake state after abrupt stop is still fragile. A later rerun after prior execution failed with:
  - `malformed HTTP response "\x82\x05\x05\x00\x00\b\x00\x82\x05\x05\x00\x00\t\x01\x82\x05\x05\x00\x00"`
- Serial device availability can still temporarily block runs:
  - `unable to open port "/dev/ttyACM0"`

## Key Decisions
- **TouchDial vs IntKnob**: TouchDial internally creates IntKnobs; creating them separately caused double-binding conflicts. Solution: Let TouchDial handle knob binding, only add WatchedInt watchers for logging.
- **Rate-limiting for WebSocket stability**: Rapid creation of 12 MultiButtons caused "bad opcode 4" WebSocket errors. Solution: initially 50ms delay between each MultiButton, 100ms after batch completion; later moved pacing responsibility into the package writer.
- **Touch flash mapping**: Map touch buttons 1-12 to physical buttons in repeating pattern (Circle, Btn1-7, then cycle) with corresponding rainbow colors for visual feedback. Later changed per user request to flash the **screen button area** instead of physical LEDs.
- **B-lite before B before C**: Implement package-owned single writer and pacing first, then render invalidation/coalescing, and only then assess whether strict ack-gated flow control (C) is necessary.
- **Root package creation**: Move from ticket-local script modules to a real root package `github.com/go-go-golems/loupedeck` so transport/render control lives in the package, not the app.
- **Composable listeners**: Keep legacy `Bind*` methods for compatibility, but add `OnButton`, `OnButtonUp`, `OnKnob`, `OnTouch`, `OnTouchUp` plus `Subscription` for multi-listener fanout.
- **Safe lifecycle**: Replace panic-based `Listen()` behavior with `Listen() error`; make `SerialWebSockConn.Close()` actually close the underlying port; make `Loupedeck.Close()` return aggregated errors.
- **Single outbound writer**: All websocket writes now go through one writer goroutine in `writer.go`; pacing is applied per outbound command, not per low-level message.
- **Display draw grouping**: Treat framebuffer upload + draw trigger as one logical command (`displayDrawCommand`) instead of two unrelated sends.
- **Render coalescing**: Add renderer above writer; `Display.Draw()` now invalidates a keyed region so repeated draws to the same tile/strip collapse to the latest state before flush.
- **Feature tester migration**: Root command `cmd/loupe-feature-tester/main.go` should use package-owned pacing (`WriterOptions`) and `On*` listeners rather than app-level sleeps and overwrite-prone `Bind*` usage.
- **Current interpretation of hardware results**: Steady-state draw-storm behavior appears improved under B-lite/full-B groundwork, while remaining failures now look more like reconnect/reset corruption after abrupt termination than immediate backpressure collapse.
- **Documentation strategy**: Maintain three doc forms in LOUPE-003—primary design guide, chronological diary, and separate deep-dive article—plus mirror the deep dive into Obsidian for durable project knowledge.

## Next Steps
1. Finish latest LOUPE-003 diary/changelog/task bookkeeping for Phase 4 hardware observations and the new deep-dive article.
2. Link the new Obsidian article from the main project note if desired.
3. Re-run the root feature tester with a **clean exit path** (preferably Circle button, not `timeout`) to isolate reconnect/reset issues.
4. Compare clean-exit rerun behavior against the malformed-handshake case:
   - `malformed HTTP response "\x82\x05\x05\x00\x00\b\x00\x82\x05\x05\x00\x00\t\x01\x82\x05\x05\x00\x00"`
5. Decide whether to open a dedicated lifecycle/reset ticket or continue LOUPE-003 for reconnect hygiene.
6. Keep C undecided for now; only adopt strict in-flight/ack-gated flow control if failures remain after clean-exit validation.

## Critical Context
- **Repository**: `/home/manuel/code/wesen/2026-04-11--loupedeck-test/`
- **Upstream reference library**: `sources/loupedeck-repo/` (frozen reference, not the active implementation)
- **Root module**: `github.com/go-go-golems/loupedeck`
- **Main new root command**: `/home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-feature-tester/main.go`
- **Root package key files**:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/listeners.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/writer.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/renderer.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/display.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/message.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/connect.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/listen.go`
- **Root tests**:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/listeners_test.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/writer_test.go`
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/renderer_test.go`
- **LOUPE-003 ticket path**:
  - `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/`
- **LOUPE-003 primary design doc**:
  - `.../design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
- **LOUPE-003 deep dive doc**:
  - `.../design-doc/02-technical-deep-dive-the-new-go-go-golems-loupedeck-frontend-implementation.md`
- **LOUPE-003 diary**:
  - `.../reference/01-investigation-diary.md`
- **Obsidian deep-dive article**:
  - `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/11/ARTICLE - Loupedeck - Backpressure-Safe Go Frontend Deep Dive.md`
- **Obsidian project note**:
  - `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/11/PROJ - Loupedeck Live Hello World - Serial Go Driver.md`
- **Hardware**: Loupedeck Live product `0004`, usually `/dev/ttyACM0`
- **Protocol**: “mutant WebSocket over serial” over USB CDC ACM
- **Display dimensions**: left `(60×270)`, main `(360×270)`, right `(60×270)`
- **Successful hardware signal after reconnect**:
  - `Feature tester ready writer_stats="{QueuedCommands:18 SentCommands:18 SentMessages:32 FailedCommands:0 MaxQueueDepth:1}"`
- **Important remaining failure modes**:
  - `unable to open port "/dev/ttyACM0"`
  - `malformed HTTP response "\x82\x05\x05\x00\x00\b\x00\x82\x05\x05\x00\x00\t\x01\x82\x05\x05\x00\x00"`
- **Current conclusion**: B-lite + initial full-B groundwork appear to improve steady-state startup/runtime behavior enough that immediate draw-storm websocket failures are less prominent; the biggest remaining issue appears to be reconnect/reset hygiene after abrupt stop rather than obviously insufficient steady-state flow control.

---

**Turn Context (split turn):**

## Original Request
- The user asked to try/run the software on the device.

## Early Progress
- The assistant launched the root-level tester `go run ./cmd/loupe-feature-tester` in the background and intended to log to `/tmp`, telling the user to exit with the Circle button.
- The launch command had a shell/log variable bug (`LOG` echoed empty), so the assistant checked whether the process started and found the actual log file in `/tmp/loupedeck-feature-tester-1775947380.log`.
- Log inspection showed the tester did connect and exercise the device, but later crashed due to a callback/data-shape issue rather than the earlier websocket backpressure problem.
- The assistant inspected `connect.go` and found async callbacks for `Version` and `Serial` responses during connect.
- It patched those callbacks to guard against short/empty response data instead of indexing blindly, then ran `gofmt` and `go test ./...` successfully.
- The assistant reran the tester with `timeout 8s go run ./cmd/loupe-feature-tester`; startup succeeded and logs showed normal connect/listen/draw behavior.

## Context for Suffix
- During the rerun, the assistant noticed a separate lifecycle bug: `tryConnect` timed out after 2s, then retried `doConnect` using the same serial/websocket connection while the first attempt was still active, leading to later `Port has been closed` read failures.
- To fix that, it refactored `ConnectAutoWithOptions`/`ConnectPathWithOptions` so `tryConnect` receives an opener function, opens a fresh serial connection for the retry, closes the timed-out one, and then reconnects cleanly.
- After this patch, `gofmt` and `go test ./...` passed.
- The suffix begins with the assistant doing one more test run using the fixed reconnect path, noting a possible unplug/replug may be needed if the device remained in a bad state from prior crashes.
</summary>"

**Assistant interpretation:** Continue the real-device validation, fix the new connect-path bugs that surface under the root package, and record the evidence carefully.

**Inferred user intent:** Get the new root package to a point where it can be run repeatedly on hardware with cleaner lifecycle behavior and documented evidence rather than one-off success.

**Commit (code):** `b9e2f22` — "Fix connect retry to reopen serial connection"

### What I did
- Inspected the root tester logs from the background run and confirmed the earlier software exercised the device but later failed in a callback/data-shape path rather than the old websocket draw-storm path.
- Hardened the async connect callbacks in `connect.go` so `Version` ignores short payloads and `Serial` ignores empty payloads instead of indexing blindly.
- Ran:
  - `gofmt -w connect.go`
  - `go test ./...`
- Found the deeper retry bug in `tryConnect`: after the 2-second timeout it reused the same `SerialWebSockConn` for the second attempt, which later manifested as closed-port reads.
- Refactored `tryConnect` to accept an opener function so the timed-out connection can be closed and the retry can open a fresh serial connection before calling `doConnect` again.
- Committed the retry fix.
- Started the root tester in the background with:

```bash
LOG="/tmp/loupedeck-feature-tester-$(date +%s).log"
nohup go run ./cmd/loupe-feature-tester >"$LOG" 2>&1 &
```

- Verified the live log in `/tmp/loupedeck-feature-tester-1775947611.log`.
- Confirmed a clean interactive session on hardware and a clean exit via the Circle button.
- Captured the final clean-exit writer stats from the log:

```text
Final writer stats stats="{QueuedCommands:160 SentCommands:160 SentMessages:258 FailedCommands:0 MaxQueueDepth:1}"
```

### Why
- The callback guard bug and retry-on-the-same-connection bug would both make reconnect behavior look worse than it really is.
- A clean-exit run was necessary to distinguish steady-state draw behavior from state corruption caused by `timeout` killing the process.

### What worked
- The callback guards prevented the new root package from crashing on short async response payloads.
- Reopening a fresh serial connection for the timeout retry removed the obvious self-poisoning retry path.
- The background run connected, handled real button activity, exited via Circle, and shut down cleanly enough to flush the final writer stats without failed outbound commands.

### What didn't work
- Even with the retry fix, subsequent reconnects in later benchmark runs still produced familiar warnings such as:

```text
time=2026-04-11T18:55:44.846-04:00 level=WARN msg="Read error, exiting" error="Port has been closed"
time=2026-04-11T18:55:46.849-04:00 level=WARN msg="dial failed" err="Port has been closed"
```

- That means the reconnect story is improved but not fully solved; there is still residual device or serial state that can outlive one process.

### What I learned
- The new primary failure mode had indeed moved from immediate draw overload to edge-case lifecycle handling.
- Clean exit matters: the Circle-button path produced much better evidence than `timeout` because it let the application reset button colors and close in order.
- `Version` responses are not guaranteed to be the only thing arriving on that transaction path when the device or websocket state is already slightly off.

### What was tricky to build
- The hardest part was separating a logic bug in our retry code from genuine device-side reset fragility. Both showed up as connection failures, but they have different fixes. The retry bug was internal and deterministic; the residual closed-port behavior is intermittent and likely involves the device or OS serial stack as well.
- Another tricky point was that the connection helpers had already been threaded through the package and command API surface. The retry fix had to preserve the existing `ConnectAutoWithOptions` / `ConnectPathWithOptions` shape while still injecting a fresh-open path underneath.

### What warrants a second pair of eyes
- `tryConnect` now closes the timed-out serial connection before reopening, but a reviewer should confirm there is no race between the goroutine running the first `doConnect` and the close path.
- The callback hardening is intentionally defensive; if stricter transaction validation is added later, this path may want more explicit message-type filtering.

### What should be done in the future
- Continue reconnect-focused validation after clean exits, not only after abrupt-stop runs.
- If intermittent `Port has been closed` warnings continue, consider a dedicated lifecycle/reset ticket or a startup drain/recovery step.

### Code review instructions
- Review:
  - `connect.go`
  - `cmd/loupe-feature-tester/main.go`
- Validate with:

```bash
gofmt -w connect.go
go test ./...
nohup go run ./cmd/loupe-feature-tester >/tmp/loupedeck-feature-tester.log 2>&1 &
# interact with hardware, then exit via Circle
```

### Technical details
- The key retry structural change was turning `tryConnect` into an opener-driven helper so the second attempt can call `open()` again instead of reusing the first `SerialWebSockConn`.
- The clean-exit evidence log lived at:

```text
/tmp/loupedeck-feature-tester-1775947611.log
```

## Step 11: Add a raw FPS benchmark command and measure full-screen, single-tile, and mixed-button throughput on real hardware

Once the clean-exit feature tester run showed that the root package was behaving sanely enough to interact with, I shifted from smoke testing to measurement. The immediate question from the user was not whether the package merely works, but how fast it can push different update granularities on the real Loupedeck Live.

I implemented a dedicated benchmark command instead of trying to infer these numbers from the feature tester logs. That let me control writer settings, disable the render scheduler to expose raw transport ceilings, precompute animation frames so CPU-side image generation stayed predictable, and sweep targets until the device or transport began to fall behind.

### Prompt Context

**User prompt (verbatim):** "everything seems to work fine. measure max fps you can get on:
- full screen
- single button
- updating individual buttons playing animations at different framerates (use geometric patterns or something)."

**Assistant interpretation:** Build a repeatable hardware benchmark harness and use it to measure sustainable display-update throughput for the main display, a single tile, and a bank of independently animated touch buttons.

**Inferred user intent:** Replace vague “it seems fast enough” judgments with concrete device-backed throughput numbers that can guide future rendering choices.

**Commit (code):** N/A — code changes for the benchmark command are recorded in the next focused commit after this diary update.

### What I did
- Added optional connect helpers in `connect.go` so callers can supply both writer options and explicit render-scheduler behavior:
  - `ConnectAutoWithWriterAndRenderOptions(...)`
  - `ConnectPathWithWriterAndRenderOptions(...)`
- Used those helpers to make a raw benchmark mode possible by passing `renderOptions=nil`, which disables the render scheduler so each draw goes directly through the writer.
- Added a new root command:
  - `cmd/loupe-fps-bench/main.go`
- Suppressed global slog output in the benchmark command so the existing `INFO`-level draw/read logs in `display.go` and `listen.go` would not dominate the measurements.
- Implemented three benchmark modes:
  1. full main-display sweep (`360×270`),
  2. single-tile sweep (`90×90`),
  3. 12-button mixed-framerate animation sweep using geometric test patterns.
- Precomputed geometric frames to keep raster generation predictable during the measurement runs.
- Measured on the actual device with:

```bash
go run ./cmd/loupe-fps-bench
```

- Extended the target ranges when the first pass did not hit a ceiling, then reran until the command found real “fall behind” thresholds instead of arbitrary lower stable values.
- Captured the resulting ceilings:
  - full-screen main display: `36 FPS` stable, `37.65 FPS` peak achieved before falling behind
  - single `90×90` tile: `320 FPS` stable target, about `314 FPS` practical ceiling
  - 12-tile mixed animation: `288 FPS` total stable aggregate, about `314 FPS` aggregate ceiling before the highest-rate tiles started to flatten

### Why
- The package now has a writer and an optional renderer, so it is important to distinguish raw transport/display throughput from default coalesced-renderer behavior.
- Small-tile performance is especially relevant for the Loupedeck Live because its 4×3 touchscreen layout encourages independent tile updates rather than full-screen repaints.
- A dedicated benchmark harness is reusable evidence for future tickets and optimization decisions.

### What worked
- `cmd/loupe-fps-bench` built cleanly and `go test ./...` remained green.
- The benchmark never recorded failed outbound commands in the successful sweeps.
- The measured numbers were internally consistent: both the single-tile and mixed-tile scenarios saturated around the low-`300 draws/sec` range, while full-screen pushes saturated much earlier due to much larger framebuffer payloads.
- The command successfully used the new “disable renderer” connect mode, which confirms that the package can now expose raw writer behavior separately from coalesced rendering.

### What didn't work
- Repeated connect/close cycles inside the benchmark still surfaced the familiar reconnect warnings between scenarios, including:

```text
time=2026-04-11T18:58:49.855-04:00 level=WARN msg="Read error, exiting" error="Port has been closed"
time=2026-04-11T18:58:51.876-04:00 level=WARN msg="dial failed" err="Port has been closed"
```

- Early sweeps also produced occasional short callback warnings during connect, for example:

```text
time=2026-04-11T18:57:55.693-04:00 level=WARN msg="Received short 'Version' response" message_type=15 length=1 data="\x01"
```

- Those warnings did not invalidate the throughput numbers, but they do reinforce that reconnect/handshake cleanup is still weaker than steady-state drawing.

### What I learned
- Full-screen and tile-update throughput differ by almost an order of magnitude on this device, so future APIs should continue to favor region updates over whole-screen redraws when possible.
- The raw command-rate ceiling for tile-sized updates is currently around `300 draws/sec`, which is comfortably above the rates needed for normal 12-button animation.
- The mixed-button saturation behavior showed a clear fairness effect: at very high aggregate targets, the highest-FPS tiles flatten first while lower-rate tiles continue to track well.

### What was tricky to build
- The biggest measurement risk was accidentally benchmarking logging overhead rather than display throughput. The package currently logs every draw at `INFO`, so I had to silence slog in the benchmark command to avoid distorted numbers.
- Another subtle issue was deciding what “stable” means. I used the benchmark command’s achieved FPS versus target FPS, plus writer failure counts and listen errors, rather than relying only on whether the process crashed. That produces more useful ceilings than a binary pass/fail result.
- It was also necessary to rerun the command with larger target ranges after the first pass, because the initial sweep only proved that lower rates were stable; it did not actually find the limit.

### What warrants a second pair of eyes
- The benchmark currently measures raw writer mode with the render scheduler disabled. A reviewer should keep that distinction clear when using the numbers in design discussions.
- The repeated connect/close loops in the benchmark are a strong stress test for lifecycle behavior. If later refactors change connect semantics, this command should be rerun to see whether the reconnect warnings improve.

### What should be done in the future
- Add a second benchmark mode later for “default renderer enabled” so raw throughput and coalesced-renderer UX throughput can be compared directly.
- Use the current measurements to choose conservative animation rates for demo programs; there is no need to run close to the raw `300 draws/sec` ceiling.
- Keep the C decision open until reconnect hygiene is better understood; the raw throughput numbers do not by themselves justify ack-gated flow control.

### Code review instructions
- Review:
  - `cmd/loupe-fps-bench/main.go`
  - `connect.go`
- Validate with:

```bash
gofmt -w connect.go cmd/loupe-fps-bench/main.go
go test ./...
go run ./cmd/loupe-fps-bench
```

### Technical details
- The benchmark runs in raw-writer mode using:

```go
writerOptions := loupedeck.WriterOptions{QueueSize: 4096, SendInterval: 0}
l, err := loupedeck.ConnectAutoWithWriterAndRenderOptions(writerOptions, nil)
```

- The best stable mixed-rate scenario used these per-button targets:

```text
6, 9, 12, 15, 18, 21, 24, 27, 30, 36, 42, 48 FPS
```
