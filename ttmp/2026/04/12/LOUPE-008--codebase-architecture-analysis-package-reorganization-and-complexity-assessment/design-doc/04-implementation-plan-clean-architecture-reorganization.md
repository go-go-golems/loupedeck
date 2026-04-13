---
Title: Implementation Plan - Clean Architecture Reorganization
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
    - Path: loupedeck.go
      Note: God struct mixes transport, events, font, scheduler, legacy widget state
    - Path: connect.go
      Note: Model="foo", SetDisplays() not called during connect
    - Path: display.go
      Note: Product-based display profile switch should move to connect-time
    - Path: inputs.go
      Note: Type definitions + legacy Bind* API + duplicated name maps in callers
    - Path: watchedint.go
      Note: Obsolete observable, superseded by runtime/reactive
    - Path: intknob.go
      Note: Obsolete widget, only used by feature-tester and touchdials
    - Path: multibutton.go
      Note: Obsolete widget, only used by feature-tester
    - Path: touchdials.go
      Note: Obsolete widget, only used by feature-tester
    - Path: displayknob.go
      Note: Obsolete CT widget, not used by any cmd/
    - Path: listeners.go
      Note: Dual dispatch (legacy Bind* + On* subscription) should simplify to On* only
    - Path: svg_icons.go
      Note: Utility with no device dependency, belongs in pkg/
    - Path: cmd/loupe-feature-tester/main.go
      Note: Only consumer of obsolete widget stack
    - Path: cmd/loupe-js-live/main.go
      Note: Primary app, proves runtime/ is the real architecture
    - Path: cmd/loupe-fps-bench/main.go
      Note: Heavy benchmark binary, duplicates helpers from loupe-js-live
    - Path: cmd/loupe-svg-buttons/main.go
      Note: 600-line standalone, needs svg_icons move
    - Path: runtime/js/module_ui/module.go
      Note: Duplicates button/touch/knob name maps that should live on input types
    - Path: runtime/host/events.go
      Note: Bridges to root-package Subscription/EventSource, correct dependency direction
    - Path: pkg/jsmetrics/jsmetrics.go
      Note: 260 lines of JS module registration that could merge into runtime/js/
    - Path: pkg/runtimebridge/runtimebridge.go
      Note: sync.Map glue for VM→bindings lookup, small and correct
    - Path: pkg/runtimeowner/
      Note: Owner-thread serialization, well-tested, keep as-is
ExternalSources: []
Summary: >
    Concrete implementation plan for reorganizing the loupedeck Go codebase into a clean
    three-layer architecture: a slim hardware driver (root package), a composable runtime
    framework (runtime/), and thin wiring binaries (cmd/). Deletes the obsolete widget
    stack, fixes connect-time initialization, removes font code from the driver, and
    consolidates duplicated name mappings. No backward compatibility required.
LastUpdated: 2026-04-13T14:10:00-04:00
WhatFor: Execute the cleanup in ordered phases with clear validation gates
WhenToUse: When deciding what to implement, in what order, and how to validate each step
---

# Implementation Plan — Clean Architecture Reorganization

## Executive Summary

This document is a fresh analysis of the current codebase (post-LOUPE-008 through LOUPE-012
work) and a concrete, phased implementation plan for cleaning it up. The three prior
analyses ("three brothers") identified the right problems. This plan is about **doing the
work**.

The codebase has converged on a clear architectural direction:

1. **Root package** = hardware driver (serial, websocket, protocol, display writes)
2. **`runtime/`** = application framework (reactive state, retained UI, animation, JS)
3. **`cmd/`** = thin wiring binaries
4. **`pkg/`** = reusable infrastructure (runtimeowner, runtimebridge)

The problem is that the root package still carries ~900 lines of dead widget code,
duplicated callback systems, font rendering, and a connect-time initialization gap that
forces every binary to remember `SetDisplays()`. The fix is surgical deletion and a few
small abstractions.

**What changes:**
- Delete 5 obsolete widget files (842 lines)
- Delete the legacy `Bind*` callback API
- Remove font code from the driver
- Move `SetDisplays()` into connect-time profile resolution
- Add `String()`/`Parse*()` to input types to eliminate duplicated name maps
- Move `svg_icons.go` to `pkg/svgicons`
- Simplify `cmd/loupe-js-live` by extracting a shared runner helper
- Delete `cmd/loupe-feature-tester` (only user of the dead widget stack)
- Clean up `pkg/jsmetrics` (merge thin wrapper modules into `runtime/js/`)

**What stays:**
- `runtime/` architecture — already clean, keep it
- `pkg/runtimeowner` and `pkg/runtimebridge` — correct, well-tested, keep as-is
- All JS module APIs — unchanged
- All `examples/js/` scripts — continue working unchanged

## Problem Statement

### The root package is a god package

`Loupedeck` struct has 30+ fields spanning transport, event dispatch, font rendering,
legacy widget state, and device profiling. It exposes two parallel callback systems
(`Bind*` and `On*`) that do the same thing. Five files (`watchedint.go`, `intknob.go`,
`multibutton.go`, `touchdials.go`, `displayknob.go`) implement a widget model that has
been entirely superseded by `runtime/ui + runtime/reactive`.

### Connect-time initialization is broken

`doConnect()` sets `Model: "foo"` and does **not** configure displays. Every binary must
remember to call `l.SetDisplays()` separately, or nothing renders. This is a mandatory
setup step that should not be possible to forget.

### Input name mappings are duplicated everywhere

Three separate places maintain string↔enum maps for buttons, knobs, and touch buttons:
- `runtime/js/module_ui/module.go` (for JS `onButton("Circle", ...)` API)
- `cmd/loupe-js-live/main.go` (for event logging)
- `cmd/loupe-feature-tester/main.go` (for event logging)

These should be methods on the types themselves.

### Font code bloats the driver

The root package imports `golang.org/x/image/font/opentype` and `gofont/goregular` for
`TextInBox()`, `FontDrawer()`, `Face()`, `SetDefaultFont()`. The only non-test consumers
are `touchdials.go` and `displayknob.go` — both files being deleted. The runtime has its
own text path in `runtime/gfx/text.go`.

## Current State Analysis (April 2026)

### File inventory — root package (3,434 lines total)

| File | Lines | Role | Status |
|------|------:|------|--------|
| `loupedeck.go` | 243 | Device struct, font, brightness, button colors | **Trim needed** — font code removable |
| `connect.go` | 203 | Serial/websocket connect, retry logic | **Fix** — Model="foo", no display init |
| `display.go` | 193 | Display protocol, SetDisplays profile switch | **Refactor** — move profile into connect |
| `message.go` | 212 | Protocol message types, framing | **Keep** — correct transport layer |
| `writer.go` | 220 | Outbound queue, pacing, coalescing | **Keep** — core transport pipeline |
| `renderer.go` | 148 | Render scheduler, invalidation | **Keep** — core transport pipeline |
| `listeners.go` | 241 | On* subscription API + dispatch | **Simplify** — remove Bind* dual dispatch |
| `listen.go` | 102 | WebSocket read loop, event parsing | **Keep** — correct |
| `inputs.go` | 220 | Input enums, Bind* API, coord mapping | **Simplify** — delete Bind*, add String() |
| `dialer.go` | 126 | Serial port discovery | **Keep** — correct |
| `svg_icons.go` | 252 | SVG→image rasterization | **Move** to `pkg/svgicons` |
| `displayknob.go` | 426 | CT display knob widget | **Delete** |
| `touchdials.go` | 145 | Touch dial widget | **Delete** |
| `multibutton.go` | 135 | Multi-state touch button widget | **Delete** |
| `intknob.go` | 83 | Integer knob widget | **Delete** |
| `watchedint.go` | 59 | Observable integer | **Delete** |

### File inventory — pkg/ (972 lines)

| File | Lines | Role | Status |
|------|------:|------|--------|
| `pkg/runtimeowner/runner.go` | 230 | Owner-thread serialization | **Keep** — well-tested |
| `pkg/runtimeowner/runner_test.go` | 315 | Tests | **Keep** |
| `pkg/runtimeowner/runner_race_test.go` | 41 | Race tests | **Keep** |
| `pkg/runtimeowner/types.go` | 33 | Interfaces | **Keep** |
| `pkg/runtimeowner/errors.go` | 10 | Error types | **Keep** |
| `pkg/jsmetrics/jsmetrics.go` | 260 | JS metrics module registration | **Simplify** — thin wrappers can merge |
| `pkg/runtimebridge/runtimebridge.go` | 50 | VM→bindings sync.Map | **Keep** |
| `pkg/runtimebridge/runtimebridge_test.go` | 33 | Tests | **Keep** |

### File inventory — runtime/ (5,722 lines total)

| Package | Files | Lines | Role | Status |
|---------|------:|------:|------|--------|
| `runtime/reactive/` | 6 | 557 | Signal, computed, effect, graph | **Keep** — solid |
| `runtime/ui/` | 5 | 903 | Retained UI tree (pages, tiles, displays) | **Keep** — solid |
| `runtime/render/` | 2 | 456 | UI→image renderer with themes | **Keep** — solid |
| `runtime/gfx/` | 4 | 767 | Surface, font, text | **Keep** — solid |
| `runtime/host/` | 4 | 590 | Event routing, timers, pages | **Keep** — solid |
| `runtime/anim/` | 2 | 219 | Tween, loop, timeline | **Keep** — solid |
| `runtime/easing/` | 2 | 89 | Easing functions | **Keep** — solid |
| `runtime/present/` | 2 | 284 | Frame scheduling | **Keep** — solid |
| `runtime/metrics/` | 2 | 273 | Counter/timing/trace collector | **Keep** — solid |
| `runtime/js/` | 10 | 2,184 | JS runtime + module bindings | **Keep** — clean up name maps |

### File inventory — cmd/ (2,015 lines)

| Binary | Lines | Role | Status |
|--------|------:|------|--------|
| `cmd/loupe-svg-buttons/` | 602 | SVG icon animation demo | **Keep** — update imports |
| `cmd/loupe-fps-bench/` | 525 | Throughput benchmark | **Keep** — extract shared helpers |
| `cmd/loupe-js-live/` | 479 | Primary JS runtime runner | **Simplify** — extract runner helper |
| `cmd/loupe-feature-tester/` | 284 | Legacy widget demo | **Delete** |
| `cmd/loupe-js-demo/` | 65 | Headless JS→PNG renderer | **Keep** — simple and useful |

### Dependency graph — who imports what

```
cmd/*
  ├── pkg/device/                    ← hardware driver (connect, protocol, display, events)
  │     connect, dialer, display, listen, message,
  │     writer, renderer, inputs, listeners, profile
  │
  ├── runtime/js/
  │     └── pkg/runtimebridge (VM bindings)
  │     └── pkg/runtimeowner (owner thread)
  │     └── pkg/jsmetrics (module registration)
  │     └── runtime/js/env/ (environment wiring)
  │     └── runtime/js/module_* (JS API modules)
  │           └── module_ui imports pkg/device (Button, etc)
  │           └── module_gfx imports runtime/gfx
  │           └── module_anim imports runtime/anim
  │           └── module_state imports runtime/reactive
  │           └── module_easing imports runtime/easing
  │           └── module_present imports runtime/present
  │
  ├── runtime/host/
  │     └── imports pkg/device (Button, Knob, TouchButton, Subscription, EventSource)
  └── runtime/render/
        └── imports runtime/ui, runtime/gfx
```

The key dependency that must be preserved: `runtime/host` and `runtime/js/module_ui` depend on
`pkg/device` for input type definitions (`Button`, `Knob`, `TouchButton`, `ButtonStatus`)
and the `Subscription` + `EventSource` interfaces. This is the correct dependency direction —
runtime depends on driver types, not the reverse.

### Import path change impact

The driver package moves from `github.com/go-go-golems/loupedeck` to
`github.com/go-go-golems/loupedeck/pkg/device`.

**Affected import sites (11 total):**

| File | Current import | New import |
|------|---------------|------------|
| `runtime/host/events.go` | `deck "github.com/go-go-golems/loupedeck"` | `deck "github.com/go-go-golems/loupedeck/pkg/device"` |
| `runtime/host/runtime.go` | same | same |
| `runtime/host/runtime_test.go` | same | same |
| `runtime/js/module_ui/module.go` | same | same |
| `runtime/js/runtime_test.go` | same | same |
| `cmd/loupe-js-live/main.go` | `loupedeck "github.com/go-go-golems/loupedeck"` | `"github.com/go-go-golems/loupedeck/pkg/device"` |
| `cmd/loupe-fps-bench/main.go` | same | same |
| `cmd/loupe-svg-buttons/main.go` | same | same |
| `cmd/loupe-svg-buttons/main_test.go` | same | same |
| `pkg/device/*_test.go` (was root `*_test.go`) | `"github.com/go-go-golems/loupedeck"` | `"github.com/go-go-golems/loupedeck/pkg/device"` |

This is a mechanical find-and-replace. No logic changes required.

### What is actually used by whom

**Legacy widget stack usage (the deletion target):**

| Widget | Used by |
|--------|---------|
| `WatchedInt` | `IntKnob`, `MultiButton`, `TouchDial`, `DisplayKnob`, `cmd/loupe-feature-tester` |
| `IntKnob` | `TouchDial` (internal), `DisplayKnob` (internal) |
| `MultiButton` | `cmd/loupe-feature-tester` only |
| `TouchDial` | `cmd/loupe-feature-tester` only |
| `DisplayKnob` | **Nobody** (zero external consumers) |

**Legacy `Bind*` API usage:**

| API | Used by |
|-----|---------|
| `BindButton` | `displayknob.go` (internal) |
| `BindButtonUp` | Nobody |
| `BindKnob` | `displayknob.go` (internal) |
| `BindTouch` | Nobody |
| `BindTouchUp` | Nobody |
| `BindTouchCT` | `displayknob.go` (internal) |

**Root font API usage:**

| API | Used by |
|-----|---------|
| `SetDefaultFont()` | `connect.go` (during init), `touchdials.go` (being deleted) |
| `FontDrawer()` | `touchdials.go`, `displayknob.go` (both being deleted) |
| `Face()` | Nobody directly |
| `TextInBox()` | Nobody directly |

**`SetDisplays()` usage (the connect-time fix target):**

Every single binary calls it: `loupe-feature-tester`, `loupe-fps-bench`, `loupe-js-live`, `loupe-svg-buttons`.

## Proposed Solution

### Target architecture

**No `.go` files in the repository root.** Everything lives under `pkg/`, `runtime/`, or `cmd/`.

```
(root)/                            ← Only go.mod, go.sum, README.md, LICENSE, docs/, examples/

pkg/
├── device/                        ← Slim hardware driver (was root package)
│   ├── connect.go                 ← Connect returns fully initialized Device
│   ├── dialer.go                  ← Serial port discovery
│   ├── display.go                 ← Display write protocol
│   ├── profile.go                 ← NEW: device profile table + resolution
│   ├── inputs.go                  ← Input enums + String()/Parse*()
│   ├── listen.go                  ← WebSocket read loop
│   ├── listeners.go               ← On* subscription API only (no Bind*)
│   ├── loupedeck.go               ← Slim device struct
│   ├── message.go                 ← Protocol framing
│   ├── writer.go                  ← Outbound queue + pacing
│   ├── renderer.go                ← Render scheduler
│   └── svg_icons.go               ← SVG rasterization (only used by loupe-svg-buttons)
├── runtimeowner/                  ← Keep as-is
└── runtimebridge/                 ← Keep as-is

runtime/                           ← Keep as-is (already clean)
├── anim/
├── easing/
├── gfx/
├── host/
├── js/
│   ├── env/
│   └── module_*/
├── metrics/
├── present/
├── reactive/
├── render/
└── ui/

cmd/
├── loupe-js-live/                 ← Simplified, uses shared runner
├── loupe-js-demo/                 ← Keep as-is
├── loupe-fps-bench/               ← Keep, uses shared runner
├── loupe-svg-buttons/             ← Keep, updated imports
└── (loupe-feature-tester DELETED)
```

**Import path change:** The driver package moves from
`github.com/go-go-golems/loupedeck` to `github.com/go-go-golems/loupedeck/pkg/device`.
This affects 11 import sites across `runtime/host`, `runtime/js/module_ui`, and all `cmd/` binaries.

**Why `pkg/device/` and not `pkg/loupedeck/`:** The module is already named `loupedeck`.
Writing `import "github.com/go-go-golems/loupedeck/pkg/loupedeck"` is redundant.
`device` describes what the package *does* (hardware device driver), not what brand it is.

### What gets deleted

| What | Lines removed | Why |
|------|-------------:|-----|
| `watchedint.go` | 59 | Superseded by `runtime/reactive` Signal |
| `intknob.go` | 83 | Only used by other deleted files |
| `multibutton.go` | 135 | Only used by `cmd/loupe-feature-tester` |
| `touchdials.go` | 145 | Only used by `cmd/loupe-feature-tester` |
| `displayknob.go` | 426 | Zero external consumers |
| `cmd/loupe-feature-tester/` | 284 | Only consumer of deleted widget stack |
| Root font code | ~65 | No remaining callers after widget deletion |
| `Bind*` APIs + fields | ~80 | Dual dispatch eliminated |
| CT drag state fields | ~15 | Only for DisplayKnob |
| **Total** | **~1,292** | |

### What gets created

| What | Purpose |
|------|---------|
| `profile.go` | Table-driven device specs, called during connect |
| `String()` / `Parse*()` on input types | Authoritative name mappings |
| Shared runner helper (in `pkg/` or `cmd/`) | Reduce duplication in loupe-js-live and loupe-fps-bench |

### What gets refactored

| What | Change |
|------|--------|
| `connect.go` | Call `resolveProfile()` during `doConnect()`, remove `Model: "foo"` |
| `display.go` | Move the product switch into `profile.go`, simplify `SetDisplays()` to package-internal |
| `loupedeck.go` | Remove font fields, `SetDefaultFont()`, `FontDrawer()`, `Face()`, `TextInBox()`; remove widget fields |
| `inputs.go` | Remove `Bind*` methods, add `String()` / `Parse*()` |
| `listeners.go` | Simplify dispatch to only use `On*` listener maps, remove `ensureListenerMapsLocked()` branches for Bind maps |
| `listen.go` | Remove `TouchCT` / `TouchEndCT` handling (only relevant for deleted DisplayKnob) |
| `runtime/js/module_ui/module.go` | Use `Button.String()`, `Knob.String()`, `TouchButton.String()` instead of hand-maintained maps |
| `cmd/loupe-js-live/main.go` | Remove `buttonName()`, `touchName()`, `knobName()` helpers; extract runner helper |

### Design Decisions

#### Decision 0: No .go files in the repository root
The root directory should only contain metadata files (go.mod, go.sum, README.md, LICENSE)
and non-Go directories (docs/, examples/, sources/). All Go code belongs under `pkg/`,
`runtime/`, or `cmd/`. This is the standard Go project layout for multi-package modules.

#### Decision 1: Delete the widget stack, don't quarantine it
Moving dead code to `legacy/` preserves the cognitive overhead. The user explicitly does not need
backward compatibility. If CT knob UI is needed in the future, it should be rebuilt on top of
`runtime/ui + runtime/render`, not by reviving `DisplayKnob`.

#### Decision 2: Keep TouchCT/TouchEndCT message parsing but remove the dispatch
The CT touch messages are valid protocol events. The right approach is to keep parsing them in
`listen.go` and dispatching through the normal `OnTouch` subscription system (or a new
dedicated `OnTouchCT` subscription), rather than through the deleted `touchDKBindings` field.
However, since no current binary needs CT touch, the simplest correct move is to **log them
as unhandled** until someone needs them, matching how unknown buttons are handled today.

#### Decision 3: `profile.go` as a standalone file in `pkg/device/`
Device profiling is integral to the driver. It belongs in the device package. A separate file
keeps it discoverable without adding import complexity.

#### Decision 4: `String()` / `Parse*()` as methods on the `pkg/device` types
These types (`Button`, `Knob`, `TouchButton`) are defined in the device package. Adding methods
there is the correct Go pattern. No new package needed.

#### Decision 5: `svg_icons.go` moves into `pkg/device/`
Since all Go files are leaving the root, `svg_icons.go` moves alongside the rest of the driver.
It has zero internal dependency on device internals but it's a small file (252 lines) that isn't
worth its own package for a single consumer. If a second consumer appears later, it can be
extracted to `pkg/svgicons/` then.

#### Decision 6: Merge thin JS module wrappers
`runtime/js/module_metrics/` (12 lines) and `runtime/js/module_scene_metrics/` (12 lines) are
pure delegation wrappers to `pkg/jsmetrics`. The real question is whether `pkg/jsmetrics` at
260 lines is worth its own package. Since it has zero dependency on the device package and is
only imported from `runtime/js/`, the cleanest option is to inline it into `runtime/js/` as
a single `metrics_modules.go` file. But this is low priority — the current structure works.

**Verdict:** Leave `pkg/jsmetrics` and its thin wrappers as-is for now. Not worth the churn.

### Alternatives Considered

#### Alternative A: Keep widget stack in `pkg/widgets/`
Rejected. The widget stack is architecturally incompatible with the `runtime/ui` model.
Keeping it around creates confusion about which UI system to use for new work.

#### Alternative B: Make `SetDisplays()` private but keep the two-step pattern
Better than today but still wrong. The profile should be resolved during connect because
device capabilities are part of the connection contract.

#### Alternative C: Move `writer.go` and `renderer.go` to `internal/pipeline/`
Rejected for now. They're close to the device boundary and not imported by any code outside
the root package. Moving them adds import complexity without reducing cognitive load. The
root package is already their correct home.

#### Alternative D: Create a `loupedeck/types` sub-package for input types
Rejected. The types are small enums. A sub-package would force every consumer to add an
extra import for no real gain.

## Implementation Plan

### Phase 0 — Baseline validation (5 min)

Before touching anything:

```bash
go test ./...
go vet ./...
```

Confirm all tests pass. Record the baseline.

---

### Phase 1 — Move root package to `pkg/device/` (20 min)

This is the structural foundation for everything else. Do it first so all subsequent
phases work in the target layout.

**Step 1.1: Create `pkg/device/` and move all retained .go files**

```bash
mkdir -p pkg/device
# Move driver files
git mv connect.go dialer.go display.go inputs.go listen.go listeners.go \
       loupedeck.go message.go writer.go renderer.go svg_icons.go pkg/device/
# Move their tests
git mv loupedeck_test.go listeners_test.go renderer_test.go \
       writer_test.go svg_icons_test.go pkg/device/
```

(Tests for the widget files are skipped — those files get deleted in Phase 2.)

**Step 1.2: Update `package` declaration**

All moved files declare `package loupedeck`. Change to `package device` to match the directory.

```bash
sed -i 's/^package loupedeck$/package device/' pkg/device/*.go
```

**Step 1.3: Update all import paths**

Mechanical find-and-replace across the codebase:

```bash
find runtime/ cmd/ pkg/ -name '*.go' -exec \
  sed -i 's|"github.com/go-go-golems/loupedeck"|"github.com/go-go-golems/loupedeck/pkg/device"|g' {} +
```

**Step 1.4: Update package aliases where needed**

Some files import as `loupedeck "..."` or `deck "..."` — those aliases still work.
Some cmd/ files reference `loupedeck.Something` — verify the alias still resolves.

**Step 1.5: Update `pkg/device/*_test.go` imports**

Tests that import the module path externally need updating. Internal tests
(`package device`) need no import change.

**Step 1.6: Validate**

```bash
go build ./...
go test ./...
go vet ./...
```

At this point, the root directory should have **zero `.go` files**.

```bash
ls *.go 2>/dev/null | wc -l   # should be 0
```

---

### Phase 2 — Delete the obsolete widget stack (30 min)

This is the single highest-impact change. It removes 842 lines of dead code and
eliminates the entire `WatchedInt` → `IntKnob` → `TouchDial`/`MultiButton`/`DisplayKnob`
chain.

**Step 2.1: Delete files**

```
rm pkg/device/watchedint.go
rm pkg/device/intknob.go
rm pkg/device/multibutton.go
rm pkg/device/touchdials.go
rm pkg/device/displayknob.go
```

**Step 2.2: Remove widget-related fields from `Loupedeck` struct (in `pkg/device/loupedeck.go`)**

Remove:
- `touchDKBindings   TouchDKFunc`
- `dragDKBinding     DragDisplayKnobFunc`
- `dragDKStarted     bool`
- `dragDKStartX      uint16`
- `dragDKStartY      uint16`
- `dragDKStartTime   time.Time`

**Step 2.3: Remove widget-related types and methods from `inputs.go`**

Remove (in `pkg/device/inputs.go`):
- `TouchDKFunc` type
- `DragEvent` type + constants (`DragClick`, `DragDone`)
- `DragDisplayKnobFunc` type (was in `displayknob.go`)
- `BindTouchCT()` method

**Step 2.4: Clean up CT touch dispatch in `listen.go`**

In `pkg/device/listen.go`, replace the `TouchCT` / `TouchEndCT` cases.
log them as debug messages (matching how unhandled events are treated):

```go
case TouchCT:
    x := binary.BigEndian.Uint16(message[4:])
    y := binary.BigEndian.Uint16(message[6:])
    slog.Debug("Received CT touch (no handler)", "x", x, "y", y)
case TouchEndCT:
    x := binary.BigEndian.Uint16(message[4:])
    y := binary.BigEndian.Uint16(message[6:])
    slog.Debug("Received CT touch end (no handler)", "x", x, "y", y)
```

**Step 2.5: Delete `cmd/loupe-feature-tester/`**

```bash
rm -rf cmd/loupe-feature-tester/
```

**Step 2.6: Remove `doConnect` init for deleted Bind maps**

In `pkg/device/connect.go`, remove initialization of:
- `l.buttonBindings`
- `l.buttonUpBindings`
- `l.knobBindings`
- `l.touchBindings`
- `l.touchUpBindings`

(these become empty in Phase 2 but removing now keeps Phase 1 clean)

**Step 2.7: Validate**

```bash
go build ./...
go test ./...
go vet ./...
```

Expected: all build and test failures are from references to deleted types/methods.
Fix any remaining references (there shouldn't be any outside the deleted files).

---

### Phase 3 — Remove legacy `Bind*` API (20 min)

The dual dispatch (`Bind*` + `On*`) is now unnecessary since the only consumers of `Bind*`
were the deleted widget files.

**Step 3.1: Remove `Bind*` methods from `pkg/device/inputs.go`**

Delete:
- `BindButton()`
- `BindButtonUp()`
- `BindKnob()`
- `BindTouch()`
- `BindTouchUp()`

**Step 3.2: Remove legacy fields from `Loupedeck` struct (in `pkg/device/loupedeck.go`)**

Delete:
- `buttonBindings     map[Button]ButtonFunc`
- `buttonUpBindings   map[Button]ButtonFunc`
- `knobBindings       map[Knob]KnobFunc`
- `touchBindings      map[TouchButton]TouchFunc`
- `touchUpBindings    map[TouchButton]TouchFunc`

**Step 3.3: Simplify `pkg/device/listeners.go`**

- Remove `ensureListenerMapsLocked()` branches for the deleted Bind maps
- Simplify `dispatchButton()`, `dispatchKnob()`, `dispatchTouch()` to only iterate
  the listener maps (no primary callback check)
- Remove from `doConnect()`: the deleted map initializations

**Step 3.4: Validate**

```bash
go build ./...
go test ./...
```

---

### Phase 4 — Remove font code from device package (15 min)

After Phase 2+3, no code calls `FontDrawer()`, `Face()`, `TextInBox()`, or `SetDefaultFont()`
outside of `connect.go`'s init. The runtime has its own font/text stack.

**Step 4.1: Remove font fields and methods from `pkg/device/loupedeck.go`**

Delete:
- `font *opentype.Font`
- `face font.Face`
- `fontdrawer *font.Drawer`
- `FontDrawer()` method
- `Face()` method
- `TextInBox()` method
- `SetDefaultFont()` method

**Step 4.2: Remove font imports from `pkg/device/loupedeck.go`**

Remove:
- `"golang.org/x/image/font"`
- `"golang.org/x/image/font/gofont/goregular"`
- `"golang.org/x/image/font/opentype"`
- `"golang.org/x/image/math/fixed"`
- `"image/color"` (if no longer needed)
- `"image/draw"` (if no longer needed)

**Step 4.3: Remove `SetDefaultFont()` call from `pkg/device/connect.go`**

Delete the block:
```go
err = l.SetDefaultFont()
if err != nil {
    return nil, fmt.Errorf("Unable to set default font: %v", err)
}
```

**Step 4.4: Run `go mod tidy`**

```bash
go mod tidy
```

**Step 4.5: Validate**

```bash
go build ./...
go test ./...
```

---

### Phase 5 — Fix connect-time device profiling (30 min)

This is the most structurally important change. It makes `Connect*()` return a fully
initialized device.

**Step 5.1: Create `pkg/device/profile.go`**

```go
package loupedeck

import "fmt"

// DeviceProfile describes the hardware capabilities of a specific Loupedeck model.
type DeviceProfile struct {
    ProductID string
    Name      string
    Displays  []DisplaySpec
}

// DisplaySpec describes one physical display region.
type DisplaySpec struct {
    Name      string
    ID        byte
    Width     int
    Height    int
    OffsetX   int
    OffsetY   int
    BigEndian bool
}

// profiles is the table of known Loupedeck devices.
var profiles = map[string]DeviceProfile{
    "0003": {
        ProductID: "0003",
        Name:      "Loupedeck CT v1",
        Displays: []DisplaySpec{
            {"left", 'L', 60, 270, 0, 0, false},
            {"main", 'A', 360, 270, 60, 0, false},
            {"right", 'R', 60, 270, 420, 0, false},
            {"dial", 'W', 240, 240, 0, 0, true},
        },
    },
    "0007": {
        ProductID: "0007",
        Name:      "Loupedeck CT v2",
        Displays: []DisplaySpec{
            {"left", 'M', 60, 270, 0, 0, false},
            {"main", 'M', 360, 270, 60, 0, false},
            {"right", 'M', 60, 270, 420, 0, false},
            {"all", 'M', 480, 270, 0, 0, false},
            {"dial", 'W', 240, 240, 0, 0, true},
        },
    },
    "0004": {
        ProductID: "0004",
        Name:      "Loupedeck Live",
        Displays: []DisplaySpec{
            {"left", 'L', 60, 270, 0, 0, false},
            {"main", 'A', 360, 270, 0, 0, false},
            {"right", 'R', 60, 270, 0, 0, false},
        },
    },
    "0006": {
        ProductID: "0006",
        Name:      "Loupedeck Live S",
        Displays: []DisplaySpec{
            {"left", 'M', 60, 270, 0, 0, false},
            {"main", 'M', 360, 270, 60, 0, false},
            {"right", 'M', 60, 270, 420, 0, false},
            {"all", 'M', 480, 270, 0, 0, false},
        },
    },
    "0d06": {
        ProductID: "0d06",
        Name:      "Razer Stream Controller",
        Displays: []DisplaySpec{
            {"left", 'M', 60, 270, 0, 0, false},
            {"main", 'M', 360, 270, 60, 0, false},
            {"right", 'M', 60, 270, 420, 0, false},
            {"all", 'M', 480, 270, 0, 0, false},
        },
    },
}

func resolveProfile(product string) (DeviceProfile, error) {
    p, ok := profiles[product]
    if !ok {
        return DeviceProfile{}, fmt.Errorf("unknown device product ID: %q", product)
    }
    return p, nil
}
```

**Step 5.2: Update `doConnect()` in `pkg/device/connect.go` to apply profile**

Replace the `Model: "foo"` and empty `displays` init with:

```go
profile, err := resolveProfile(c.Product)
if err != nil {
    return nil, err
}

l := &Loupedeck{
    // ...
    Vendor:   c.Vendor,
    Product:  c.Product,
    Model:    profile.Name,
    displays: map[string]*Display{},
}

// Apply device profile
for _, spec := range profile.Displays {
    l.addDisplay(spec.Name, spec.ID, spec.Width, spec.Height,
        spec.OffsetX, spec.OffsetY, spec.BigEndian)
}
```

**Step 5.3: Remove `SetDisplays()` from public API**

Delete it from `pkg/device/display.go`.

**Step 5.4: Update all cmd/ binaries**

Remove `l.SetDisplays()` calls from:
- `cmd/loupe-fps-bench/main.go` (3 call sites)
- `cmd/loupe-js-live/main.go` (1 call site)
- `cmd/loupe-svg-buttons/main.go` (1 call site)

**Step 5.5: Validate**

```bash
go build ./...
go test ./...
```

Then manually test with hardware:
```bash
go run ./cmd/loupe-js-live --script examples/js/01-hello.js --duration 10s
```

---

### Phase 6 — Authoritative input name mappings (20 min)

**Step 6.1: Add `String()` methods to `pkg/device/inputs.go`**

```go
func (b Button) String() string { /* table lookup */ }
func (k Knob) String() string   { /* table lookup */ }
func (t TouchButton) String() string { /* table lookup */ }
func (s ButtonStatus) String() string { /* "down"/"up" */ }
```

**Step 6.2: Add `Parse*()` functions to `pkg/device/inputs.go`**

```go
func ParseButton(string) (Button, error)
func ParseKnob(string) (Knob, error)
func ParseTouchButton(string) (TouchButton, error)
```

**Step 6.3: Remove duplicated maps from `runtime/js/module_ui/module.go`**

Replace the hand-maintained `buttons`, `touches`, `knobs` maps with `ParseButton()`,
`ParseTouchButton()`, `ParseKnob()`.

**Step 6.4: Remove duplicated helpers from `cmd/loupe-js-live/main.go`**

Delete `buttonName()`, `touchName()`, `knobName()`, `buttonStatusName()`.
Use the new `String()` methods directly.

**Step 6.5: Verify no other name-map duplication**

This binary doesn't have name helpers currently (it doesn't log events), but
verify no duplication remains.

**Step 6.6: Validate**

```bash
go build ./...
go test ./...
```

---

### Phase 7 — Simplify `cmd/loupe-js-live` (30 min, optional)

This phase reduces the 479-line `main.go` by extracting reusable wiring logic.
Lower priority than Phases 1-6 but worth doing while we're here.

**Step 7.1: Extract `pkg/runner/` helper**

Create a helper that encapsulates the common pattern:
1. Parse flags
2. Connect to device (now returns fully initialized device, no `SetDisplays()`)
3. Start listen goroutine
4. Create JS environment + runtime
5. Set up present/flush pipeline
6. Wire signal handling
7. Run main loop

The extracted helper should be usable by both `loupe-js-live` and any future runner.

**Step 7.2: Move stats formatting into the helper**

`renderStatsWindow`, `diffWriterStats`, `formatJSCounters`, etc. are generic
enough to live in the runner package.

**Step 7.3: Validate**

```bash
go build ./...
go test ./...
```

Manual test:
```bash
go run ./cmd/loupe-js-live --script examples/js/01-hello.js --duration 10s
go run ./cmd/loupe-js-live --script examples/js/07-cyb-ito-prototype.js --duration 30s
```

## Testing and Validation Strategy

### Automated validation (every phase)

```bash
# 1. Build everything
go build ./...

# 2. Run all tests
go test ./...

# 3. No .go files in root
ls *.go 2>/dev/null | wc -l   # must be 0

# 4. Check for lingering references to deleted code
rg -n "\b(WatchedInt|IntKnob|MultiButton|TouchDial|DisplayKnob|DKWidget|WidgetHolder)\b" \
   --include='*.go' --glob '!ttmp/**' --glob '!sources/**'

# 5. Check for lingering Bind* references
rg -n "\bBind(Button|ButtonUp|Knob|Touch|TouchUp|TouchCT)\b" \
   --include='*.go' --glob '!ttmp/**' --glob '!sources/**'

# 6. Check no binary still calls SetDisplays
rg -n "\.SetDisplays\(" cmd/ --include='*.go'

# 7. Check font code is gone from device package
rg -n "FontDrawer|TextInBox|SetDefaultFont|gofont/goregular" \
   --include='*.go' --glob '!ttmp/**' --glob '!sources/**'

# 8. Vet
go vet ./...
```

### Manual validation (after Phase 4+)

1. **loupe-js-live** — connect, render JS scenes, respond to button/touch/knob events
2. **loupe-fps-bench** — run full-screen sweep, confirm it connects and draws
3. **loupe-svg-buttons** — confirm SVG icon loading and animation still works
4. **loupe-js-demo** — confirm headless PNG rendering still works

### Regression risk per phase

| Phase | Risk | Mitigation |
|-------|------|------------|
| 1 (Move to pkg/device) | Low — mechanical import path change | `go build` catches all issues |
| 2 (Delete widgets) | Very low — deleted code has no active consumers | grep confirms |
| 3 (Remove Bind*) | Low — `On*` subscription system is separate and tested | `listeners_test.go` exercises `On*` |
| 4 (Remove font) | Low — only deleted files used it | grep confirms |
| 5 (Connect profile) | Medium — changes init path for every binary | Table-driven profile with test coverage |
| 6 (Input names) | Very low — adding methods, changing map lookups | Unit test `String()`/`Parse*()` round-trips |
| 7 (Runner helper) | Low — mechanical extraction | Manual test with hardware |

## Risks

### Risk: `loupe-fps-bench` calls `SetDisplays()` inside display-closure pattern

In `loupe-fps-bench`, `SetDisplays()` is called inside a closure passed to
`runSingleRegionSweep`. After Phase 4, this becomes automatic. The closure
just needs to use `l.GetDisplay("main")` without the `SetDisplays()` call.
Simple but must be done carefully across all 3 call sites.

### Risk: CT touch events become dead code

After removing `touchDKBindings`, `TouchCT`/`TouchEndCT` events are parsed but
dispatched to nowhere. This is fine — they're CT-specific and no current user
has a CT. If CT support is added later, an `OnTouchCT` subscription API can be
added to `pkg/device/listeners.go` using the same `On*` pattern.

### Risk: Profile table might be incomplete

The current `SetDisplays()` switch handles 4 product IDs. If there are Loupedeck
variants we haven't seen, `resolveProfile()` will return an error at connect time
instead of panicking. This is an improvement.

## Open Questions

1. **Should `writer.go` and `renderer.go` be unexported?** Currently they're
   public types used by cmd/ binaries indirectly (through the `Loupedeck` struct).
   They could become internal implementation details. Low priority.

2. **Should the `sources/loupedeck-repo/` directory be kept?** It's the original
   upstream code preserved for reference. It's in `.git` as a submodule/copy.
   Not a code dependency. Keep for historical reference, but consider adding
   a README noting it's frozen.

3. **Should `pkg/jsmetrics` be inlined into `runtime/js/`?** The thin wrappers
   (`module_metrics`, `module_scene_metrics`) are 12 lines each. But the
   `pkg/jsmetrics` implementation is 260 lines with its own internal structure.
   Current structure works — not worth the churn right now.

## What NOT to do

These were considered and explicitly rejected:

- **Don't** leave any `.go` files in the repository root
- **Don't** create a `pkg/loupedeck/` package (redundant with the module name)
- **Don't** create a `pkg/device/types` sub-package
- **Don't** preserve widget code in a `legacy/` or `pkg/widgets/` directory
- **Don't** add CT-specific APIs now (add when needed)
- **Don't** merge `pkg/jsmetrics` into `runtime/js/` (works fine as-is)
- **Don't** change any JS-facing API (all `examples/js/` scripts must keep working)

## Execution Summary

| Phase | Description | Est. Time | Lines Removed | Lines Added |
|------:|-------------|----------:|--------------:|------------:|
| 0 | Baseline validation | 5 min | 0 | 0 |
| 1 | Move root → `pkg/device/` | 20 min | 0 (move) | ~5 (import paths) |
| 2 | Delete widget stack | 30 min | ~842 | ~10 |
| 3 | Remove Bind* API | 20 min | ~80 | ~5 |
| 4 | Remove font code | 15 min | ~65 | ~2 |
| 5 | Connect-time profile | 30 min | ~30 | ~90 |
| 6 | Input name mappings | 20 min | ~30 | ~60 |
| 7 | Simplify loupe-js-live | 30 min | ~100 | ~80 |
| **Total** | | **~2.8 hr** | **~1,147** | **~252** |

Net result: ~900 fewer lines of code, zero `.go` files in root, clean three-layer
architecture, and zero mandatory setup steps after `Connect*()`.

## References

- **Design Doc 01** — `LOUPE-008/design-doc/01-*` — Initial inventory and symptom analysis
- **Design Doc 02** — `LOUPE-008/design-doc/02-*` — Senior analysis identifying two coexisting UI systems
- **Design Doc 03** — `LOUPE-008/design-doc/03-*` — Big brother analysis proposing deletion over quarantine
- **LOUPE-003** — Backpressure-safe architecture (writer/renderer design rationale)
- **LOUPE-005** — JS runtime API design (module_* architecture)
- **LOUPE-006** — Full animated JS UI runtime design
- **LOUPE-009** — Trace-driven investigation of render performance
- **LOUPE-010** — Simulation-paced state with flush-gated presentation
