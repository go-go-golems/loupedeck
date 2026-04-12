---
Title: Big Brother Analysis - Grade the Prior Reviews and Refactor Without Legacy Baggage
Ticket: LOUPE-008
Status: active
Topics:
    - architecture
    - refactoring
    - analysis
    - code-quality
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupe-feature-tester/main.go
      Note: |-
        The only binary exercising most legacy widget APIs
        Only real consumer of obsolete widget stack
    - Path: cmd/loupe-js-live/main.go
      Note: Primary app path proves runtime/ is the real future architecture
    - Path: connect.go
      Note: Connect path hardcodes Model="foo" and leaves display setup to callers
    - Path: display.go
      Note: |-
        Device profile switch and display setup should move into connection/profile resolution
        Public SetDisplays() reveals device setup is happening too late in the lifecycle
    - Path: inputs.go
      Note: Type definitions plus legacy Bind* API; names should become the single source of truth
    - Path: loupedeck.go
      Note: God struct mixes transport, events, font, scheduler, and legacy CT state
    - Path: runtime/ui/ui.go
      Note: Good retained UI architecture worth keeping and expanding
    - Path: watchedint.go
      Note: Obsolete observable type superseded by runtime/reactive
ExternalSources: []
Summary: Assessment of the previous two reviews plus a stronger refactoring plan that assumes no legacy compatibility requirement. Recommends deleting the old widget stack instead of preserving it, tightening the hardware-driver boundary, and moving device profiling/setup into connect-time.
LastUpdated: 2026-04-12T16:50:00-04:00
WhatFor: Choose the real refactor plan now that legacy compatibility is explicitly not required
WhenToUse: When prioritizing implementation work and deciding what to delete vs preserve
---



# Big Brother Analysis — Grade the Prior Reviews and Refactor Without Legacy Baggage

## Executive Summary

The first two analyses were directionally useful but incomplete:

- **Design Doc 01** identified symptoms, not the disease.
- **Design Doc 02** identified the disease, but prescribed a compatibility-preserving treatment.

Now that the user has stated **we do not need legacy compatibility**, the correct refactoring is more aggressive and simpler:

1. **Delete the entire legacy widget/value stack** from the root package.
2. **Keep and strengthen the `runtime/` architecture** as the only application/UI layer.
3. **Turn the root `loupedeck` package into a narrow hardware driver**.
4. **Make device profiling/setup happen during connect**, not as a follow-up call every binary must remember.
5. **Make input names a first-class API** (`String()` / `Parse*()`), so the JS/runtime/cmd layers stop duplicating maps.

The goal is not “better file organization.” The goal is this boundary:

- `loupedeck` = hardware/transport/protocol/device-profile/output pipeline
- `runtime/*` = retained UI, reactive state, animation, JS bindings, app behavior
- `cmd/*` = small wiring binaries

That is the architecture the repo is already drifting toward. The cleanup should finish that move instead of preserving the abandoned branch.

## Grade the Prior Reviews

### Review 01 — Grade: **C+**

### What it got right
- It correctly spotted that `displayknob.go` is a bad place for a widget system.
- It correctly noticed that the root package has mixed concerns.
- It produced usable inventory material: file sizes, subsystem list, major package map.

### What it got wrong
- It overfit on **file size as complexity**.
- It treated `module_ui/module.go` boilerplate as a structural issue when it is mostly just binding glue.
- It proposed moving code around without first establishing which subsystem is active, obsolete, or strategic.

### Why the grade is not lower
Because the inventory work is still useful. It is just not sufficient for a real refactor plan.

---

### Review 02 — Grade: **B+**

### What it got right
- It found the real architectural issue: **two UI systems coexist**.
- It correctly identified the root package as a **god package** and `Loupedeck` as a **god struct**.
- It correctly called out duplicated name mappings and the awkward coexistence of `Bind*` and `On*` APIs.
- It correctly distinguished good complexity (`runtime/reactive`, `runtime/gfx/surface`, `runtimeowner`) from bad complexity.

### Where it held back
It assumed legacy compatibility still mattered, so it recommended:
- moving old widgets into `legacy/`,
- preserving old structures in a more explicit place,
- keeping some migration-oriented compromise layers.

That was a reasonable answer under uncertainty, but now it is too conservative.

### Why not an A
Because the strongest refactor is not “move the dead code to another package.” It is **delete the dead code** and shrink the root package to the hardware driver it already wants to be.

---

## My Assessment

## The Real Architectural Shape of the Codebase

This repo has three broad layers:

### 1. Hardware driver layer — root package
What belongs here:
- serial discovery and opening (`dialer.go`)
- websocket-over-serial transport
- protocol messages + transactions (`message.go`)
- event parsing and dispatch (`listen.go`, parts of `listeners.go`, `inputs.go`)
- physical display write protocol (`display.go`)
- queueing / pacing / coalescing near the device boundary (`writer.go`, `renderer.go`)
- resolved device capabilities/profile

### 2. App/runtime layer — `runtime/`
What already belongs here and is good:
- reactive state graph (`runtime/reactive`)
- retained UI tree (`runtime/ui`)
- rendering from UI to images (`runtime/render`)
- animation runtime (`runtime/anim`)
- host event/timer orchestration (`runtime/host`)
- JS modules and environment (`runtime/js`)

### 3. Tool/demo layer — `cmd/`
What belongs here:
- example/demo binaries
- experimental tools
- feature-testing utilities
- benchmarks

The problem is that the root package still contains a large amount of former app/runtime code from before `runtime/` existed.

## What should be considered obsolete immediately

These files are not “legacy but maybe useful.” They are **obsolete architecture** now that compatibility is not required:

- `watchedint.go`
- `intknob.go`
- `multibutton.go`
- `touchdials.go`
- `displayknob.go`

### Why delete instead of move?
Because they are not a second supported UI model. They are a precursor to the real one.

Evidence:
- `cmd/loupe-js-live/main.go` uses `runtime/js`, `runtime/ui`, `runtime/render`, `runtime/host`, `runtime/metrics`
- grep confirms only `cmd/loupe-feature-tester/main.go` uses `WatchedInt` / `MultiButton` / `TouchDial`
- `Bind*` APIs are only used by tests and `displayknob.go`

So the correct choice is:
- either **delete `cmd/loupe-feature-tester`**, or
- **rewrite it on top of `runtime/`**

Do not preserve the old widget abstractions as a supported path.

## The strongest missed issue: connect-time device setup is wrong

This is the biggest concrete problem neither prior review emphasized enough.

### Evidence
In `connect.go`:
- `Model` is set to `"foo"`
- displays are **not** configured during connect
- every caller must remember to call `SetDisplays()` manually afterwards

That leads to this repeated binary pattern:

```go
l, err := loupedeck.ConnectAutoWithOptions(...)
...
l.SetDisplays()
```

That is a design smell.

### Why this matters
Device capability/profile resolution is part of connecting to the hardware. If callers can forget it, then the object is not fully initialized after connect.

### Correct direction
Replace:
- `Connect*() -> *Loupedeck`
- caller separately calls `SetDisplays()`

with:
- `Connect*() -> *Device` that is already fully profiled and display-capable

Concretely:
- move the `Product` switch out of `SetDisplays()` into a `resolveProfile(product string)` helper
- apply the profile during `doConnect()`
- delete `SetDisplays()` from the public workflow

## Proposed Solution

## 1. Delete the obsolete widget/value system entirely

Delete:
- `watchedint.go`
- `intknob.go`
- `multibutton.go`
- `touchdials.go`
- `displayknob.go`

Delete the corresponding fields from `Loupedeck`:
- `touchDKBindings`
- `dragDKBinding`
- `dragDKStarted`
- `dragDKStartX`
- `dragDKStartY`
- `dragDKStartTime`

Delete related APIs:
- `BindTouchCT`
- `TouchDKFunc`
- `DragDisplayKnobFunc`
- `DragEvent`, `DragClick`, `DragDone`

If CT knob UI is needed later, rebuild it on top of `runtime/ui + runtime/render + runtime/host`, not by reviving this code.

## 2. Delete the old single-callback API

Delete from `inputs.go`:
- `BindButton`
- `BindButtonUp`
- `BindKnob`
- `BindTouch`
- `BindTouchUp`

Delete from `Loupedeck`:
- `buttonBindings`
- `buttonUpBindings`
- `knobBindings`
- `touchBindings`
- `touchUpBindings`

Simplify `dispatchButton`, `dispatchKnob`, `dispatchTouch` in `listeners.go` to only use the subscription maps.

This will shrink both the API and the internal state significantly.

## 3. Remove root-level font rendering

Delete from `loupedeck.go`:
- `font`
- `face`
- `fontdrawer`
- `FontDrawer()`
- `Face()`
- `TextInBox()`
- `SetDefaultFont()`

Delete the `opentype` / gofont dependency from the root package.

Why this is safe now:
- the only non-test uses are in the obsolete widget files being deleted
- the retained runtime has its own text path in `runtime/gfx/text.go`

## 4. Promote device profile resolution to a first-class concept

Create something like:

```go
type DeviceProfile struct {
    ProductID string
    Name      string
    Displays  []DisplaySpec
}

type DisplaySpec struct {
    Name      string
    ID        byte
    Width     int
    Height    int
    OffsetX   int
    OffsetY   int
    BigEndian bool
}
```

Then:
- `resolveProfile(product string) (DeviceProfile, error)`
- call it inside `doConnect()`
- populate `l.displays` during construction
- set `Model`/`Name` from the profile instead of `"foo"`

Delete public `SetDisplays()` or reduce it to an internal helper.

This change has outsized value because it removes a mandatory gotcha from every binary.

## 5. Make input names authoritative in `inputs.go`

Add:

```go
func (b Button) String() string
func (k Knob) String() string
func (t TouchButton) String() string

func ParseButton(string) (Button, error)
func ParseKnob(string) (Knob, error)
func ParseTouchButton(string) (TouchButton, error)
```

Then replace:
- manual maps in `runtime/js/module_ui/module.go`
- `buttonName`, `touchName`, `knobName` in `cmd/loupe-js-live/main.go`

This is a small change with immediate cleanup effect.

## 6. Move non-driver utilities out of the root package

### `svg_icons.go`
Move to `pkg/svgicons`.

Why:
- zero dependency on device internals
- only used by `cmd/loupe-svg-buttons`
- dragging `oksvg` and `rasterx` through the root package is unnecessary

### `writer.go` + `renderer.go`
These are more nuanced.

They are close to the device boundary, so leaving them under `loupedeck` is acceptable, but they should be treated as **internal transport pipeline**, not part of the conceptual driver API.

Best option:
- move to `internal/pipeline/`
- root package owns them, commands don’t import them directly

This avoids turning them into general-purpose public packages prematurely.

## 7. Shrink the primary binaries

### `cmd/loupe-js-live/main.go`
It is currently 402 lines because it mixes:
- CLI parsing
- device connection
- runtime construction
- listen loop
- flush loop
- stats logging
- signal handling
- event-name formatting

Refactor into:
- `runner.go` — wiring the deck + runtime + renderer together
- `stats.go` — logging helpers
- `names.go` disappears once `String()` exists on input types
- `main.go` becomes mostly flags + one `Run(opts)` call

### `cmd/loupe-feature-tester`
If compatibility is not needed, choose one:

**Preferred:** delete it.

If you still want a hardware exerciser, rewrite it against:
- `runtime/host` for events
- `runtime/ui` for retained display state
- `runtime/render` for visuals

Do not keep it as a justification for preserving the old root-package widget model.

## Design Decisions

### Decision 1: Delete obsolete code instead of quarantining it
**Why:** The user explicitly removed the compatibility constraint. Keeping dead architecture around imposes cognitive cost every day.

### Decision 2: Keep `runtime/` as the application framework center
**Why:** It already solves the problems the deleted code tried to solve, and it does so in a cleaner, more composable way.

### Decision 3: Make connect return a fully initialized device
**Why:** This is the cleanest fix for the `SetDisplays()` gotcha and the fake `Model` field.

### Decision 4: Use `internal/` before `pkg/` for transport pipeline helpers
**Why:** `writer.go` and `renderer.go` are implementation details of the device pipeline, not proven reusable libraries.

## Alternatives Considered

### Alternative A: Move old widgets to `legacy/`
Rejected because the user said compatibility is not needed. Moving dead code still preserves dead mental models.

### Alternative B: Keep `SetDisplays()` public but call it from connect
Better than today, but still preserves a confusing public method that should not be part of normal client flow.

### Alternative C: Merge `runtime/` back into root
Rejected immediately. `runtime/` is the best-structured part of the repo.

## Concrete Target Structure

```text
/root
  connect.go
  dialer.go
  display.go
  inputs.go
  listen.go
  listeners.go
  loupedeck.go        # slimmed device struct + stats accessors
  message.go
  profile.go          # NEW: device profiles/specs/resolution

/internal/pipeline
  writer.go
  scheduler.go

/pkg/svgicons
  svg_icons.go

/runtime
  anim/
  easing/
  gfx/
  host/
  js/
  metrics/
  reactive/
  render/
  ui/

/cmd
  loupe-js-live/      # simplified wiring binary
  loupe-fps-bench/
  loupe-svg-buttons/
  # loupe-feature-tester removed or rewritten
```

## Implementation Plan

### Phase 1 — Remove dead architecture
1. Delete the five obsolete files:
   - `watchedint.go`
   - `intknob.go`
   - `multibutton.go`
   - `touchdials.go`
   - `displayknob.go`
2. Delete CT touch/drag fields and types from root package.
3. Delete `Bind*` APIs and corresponding internal maps.
4. Update tests accordingly.

### Phase 2 — Fix connect/profile correctness
1. Add `profile.go` with table-driven device specs.
2. Resolve profile in `doConnect()`.
3. Populate displays during connect.
4. Replace `Model: "foo"` with real profile data.
5. Remove `SetDisplays()` from normal client use.

### Phase 3 — Clean boundaries
1. Remove font code from root.
2. Move `svg_icons.go` to `pkg/svgicons`.
3. Move writer/scheduler to `internal/pipeline`.
4. Update imports and tests.

### Phase 4 — Polish callers
1. Add `String()` / `Parse*()` helpers for input types.
2. Simplify `module_ui/module.go` by using parsing helpers.
3. Simplify `cmd/loupe-js-live/main.go` and other binaries.
4. Delete or rewrite `cmd/loupe-feature-tester`.

## Testing and Validation Strategy

### Structural validation
```bash
# legacy APIs gone
rg -n "\bBind(Button|ButtonUp|Knob|Touch|TouchUp|TouchCT)\b" . --glob '!ttmp/**'

# obsolete widget types gone
rg -n "\b(WatchedInt|IntKnob|MultiButton|TouchDial|DisplayKnob|DKWidget|WidgetHolder)\b" . --glob '!ttmp/**'

# callers no longer need manual setup
rg -n "\.SetDisplays\(" cmd/ . --glob '!ttmp/**'
```

### Behavior validation
```bash
go test ./...
```

Then manually validate:
1. `loupe-js-live` still connects, renders, and responds to inputs.
2. `loupe-fps-bench` still records writer/render stats.
3. `loupe-svg-buttons` still loads rasterized icons after import path update.

## Risks

### Risk: deleting feature-tester removes a useful manual tool
Mitigation: either keep it deleted and rely on JS/runtime-driven tools, or rewrite it quickly on top of `runtime/`. Do not keep the old architecture alive for one binary.

### Risk: public API churn
Accepted. The user explicitly removed the compatibility requirement.

### Risk: device profile mistakes break specific hardware variants
Mitigation: make the new `profile.go` table-driven and add focused tests around resolved display specs per product id.

## References

### Most important evidence files
- `connect.go` — connect path, fake model, initialization responsibilities
- `display.go` — product/profile switch and manual display setup
- `loupedeck.go` — overgrown device struct
- `inputs.go` — input enums plus dead callback API
- `cmd/loupe-feature-tester/main.go` — only real consumer of the old widget stack
- `cmd/loupe-js-live/main.go` — the future architecture already in use
- `runtime/ui/ui.go` — retained UI foundation worth building on

## Bottom Line

If compatibility were required, Review 02 would be close to the right plan.

But compatibility is **not** required.

So the right move is:
- **delete** the old widget/value system,
- **delete** the old single-callback API,
- **delete** root-level font support,
- **move** profile resolution into connect-time,
- **narrow** the root package to a proper hardware driver,
- **keep** `runtime/` as the only real application framework.

That is the simplest architecture, the smallest surface area, and the least confusing codebase going forward.
