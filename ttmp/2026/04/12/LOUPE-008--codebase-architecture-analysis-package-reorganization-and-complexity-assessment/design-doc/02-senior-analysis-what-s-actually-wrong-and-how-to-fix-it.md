---
Title: Senior Analysis - What's Actually Wrong and How to Fix It
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
    - Path: cmd/loupe-js-live/main.go
      Note: |-
        402-line main that wires everything, duplicates name maps
        402-line main that wires everything - proves legacy widgets are unused in primary app
    - Path: connect.go
      Note: doConnect initializes everything - hardware, events, fonts, writer, renderer
    - Path: display.go
      Note: Conflates framebuffer protocol with render scheduler integration
    - Path: displayknob.go
      Note: Entire widget system for CT knob that JS runtime never uses
    - Path: inputs.go
      Note: Hardware constants that get remapped 3 times across codebase
    - Path: loupedeck.go
      Note: God struct - 30+ fields mixing hardware, events, UI, fonts, drag tracking
    - Path: runtime/js/module_ui/module.go
      Note: Remaps button/touch/knob names that already exist in inputs.go
    - Path: watchedint.go
      Note: Dead observable system - replaced by runtime/reactive
ExternalSources: []
Summary: 'The previous analysis identified symptoms (big files) without diagnosing the disease. This document identifies the real architectural issues: a god package, a dead widget system, triplicated name mappings, and a missing hardware/framework boundary.'
LastUpdated: 2026-04-12T16:30:00-04:00
WhatFor: Provide an honest, evidence-backed diagnosis and concrete reorganization plan
WhenToUse: When actually planning and executing a refactoring effort
---


# Senior Analysis — What's Actually Wrong and How to Fix It

## Executive Summary

The previous analysis (Design Doc 01) treated file size as a proxy for complexity. It concluded that `displayknob.go` (426 lines) is the most complex file and should be split, and that `module_ui/module.go` has too much "boilerplate." Neither conclusion is wrong, but both miss the point.

This codebase has one real problem: **it contains two complete, independent UI systems, and the first one was never retired.** The root `loupedeck` package is a god package that grew organically across three eras of development, absorbing hardware protocol, font management, SVG parsing, event dispatch, render scheduling, and a widget framework — none of which were ever extracted when the second system (`runtime/`) was built on top.

The fix is not to shuffle files around. It's to **draw the boundary between what the hardware driver needs to be and what the application framework already is.**

## Diagnosis

### Problem 1: Two Dead-End UI Systems

The codebase contains two completely separate widget/value systems that never interact:

**System A — "Legacy widgets" (root package):**
- `WatchedInt` — observable integer with callback-based watchers
- `IntKnob` — maps knob rotation to an integer range, click-to-reset
- `MultiButton` — cycles through images on touch
- `TouchDial` — renders knob values on the side LCDs with drag-to-adjust-all
- `DisplayKnob` / `DKWidget` / `DKAnalogWidget` / `WidgetHolder` — full widget system for the CT knob display

**System B — "Reactive UI" (`runtime/`):**
- `reactive.Runtime` — dependency-tracked signal/computed/effect graph (SolidJS-inspired)
- `ui.UI` / `ui.Page` / `ui.Display` / `ui.Tile` — component tree with dirty tracking
- `render.Renderer` — UI-to-image flush with layer compositing and themes
- `host.Runtime` — event routing with composable subscriptions
- `anim.Runtime` — tweening, looping, timelines
- `gfx.Surface` — 8-bit pixel surface with batching and change notifications

**Evidence that System A is dead:**

`cmd/loupe-js-live/main.go` is the primary application. It:
1. Creates an `env.Environment` (which creates `host.Runtime`, `ui.UI`, `reactive.Runtime`)
2. Attaches the hardware via `env.Host.Attach(deckConn)` — using `host.Runtime`, NOT the root package's `BindButton`/`BindKnob` etc.
3. Creates a `render.Renderer` against the UI tree
4. Runs a flush loop against `renderer.Flush()`

It never touches `IntKnob`, `MultiButton`, `TouchDial`, `DisplayKnob`, `DKWidget`, or `WatchedInt`.

Only `cmd/loupe-feature-tester` (the hardware test tool) uses the legacy widgets. And even it is a development-time utility, not a production application.

**System A's types also have design problems that System B solved:**
- `WatchedInt` has no thread safety (no mutex), no unsubscription, no batching
- `WatchedInt` calls watchers synchronously inside `Set()`, which can cause re-entrant calls
- Legacy widget constructors (`l.IntKnob(...)`, `l.NewMultiButton(...)`) are methods on the `Loupedeck` god struct, coupling widget creation to hardware initialization
- The `Bind*` / `On*` duality in `inputs.go`/`listeners.go` shows the legacy system was retrofitted with a subscription pattern that coexists awkwardly with the original single-callback design

### Problem 2: The God Package and God Struct

The root `loupedeck` package (3,434 lines across 15+ files) is responsible for:

| Responsibility | Files | Should Be |
|---|---|---|
| USB serial enumeration + connection | `dialer.go` | `loupedeck` (hardware driver) |
| WebSocket-over-serial transport | `dialer.go` (net.Conn impl) | `loupedeck` (hardware driver) |
| Message protocol (types, framing, transaction IDs) | `message.go` | `loupedeck` (hardware driver) |
| Connection lifecycle (connect, reset, handshake) | `connect.go` | `loupedeck` (hardware driver) |
| Event reading + dispatch | `listen.go` | `loupedeck` (hardware driver) |
| Display protocol (framebuffer encoding, endianness) | `display.go` | `loupedeck` (hardware driver) |
| Event type definitions (Button, Knob, TouchButton) | `inputs.go` | `loupedeck` (hardware driver) |
| Outbound writer (queue, pacing, retries) | `writer.go` | `pkg/` or `loupedeck` internal |
| Render scheduler (coalescing, invalidation) | `renderer.go` | `pkg/` or separate |
| Font loading + text rendering | `loupedeck.go` (TextInBox, FontDrawer) | `pkg/` or remove (runtime/gfx/text.go replaces it) |
| SVG icon parsing + rasterization | `svg_icons.go` (252 lines) | `pkg/svgicons` or `loupedeck/icons` |
| Event subscription management (Bind*/On*/dispatch*) | `inputs.go` + `listeners.go` | Split: hardware dispatch stays, subscription routing moves |
| Legacy widgets | `intknob.go`, `multibutton.go`, `touchdials.go`, `displayknob.go`, `watchedint.go` | `loupedeck/legacy` or delete |

The `Loupedeck` struct has **30+ fields**:

```go
type Loupedeck struct {
    // Hardware identification
    Vendor, Product, Model, Version, SerialNo string
    
    // Font management (3 fields)
    font *opentype.Font
    face font.Face
    fontdrawer *font.Drawer
    
    // Transport (2 fields)
    serial *SerialWebSockConn
    conn wsConn
    
    // Pipeline (2 fields)
    writer *outboundWriter
    renderer *renderScheduler
    
    // Configuration (2 fields)
    writerOptions WriterOptions
    renderOptions RenderOptions
    
    // Legacy single-callback bindings (5 maps)
    buttonBindings, buttonUpBindings map[Button]ButtonFunc
    knobBindings map[Knob]KnobFunc
    touchBindings, touchUpBindings map[TouchButton]TouchFunc
    
    // New subscription-based listeners (5 maps)
    buttonListeners, buttonUpListeners map[Button]map[uint64]ButtonFunc
    knobListeners map[Knob]map[uint64]KnobFunc
    touchListeners, touchUpListeners map[TouchButton]map[uint64]TouchFunc
    
    // CT knob drag state (4 fields)
    touchDKBindings TouchDKFunc
    dragDKBinding DragDisplayKnobFunc
    dragDKStarted bool
    dragDKStartX, dragDKStartY uint16
    dragDKStartTime time.Time
    
    // Coordination (3 fields)
    listenerMutex sync.RWMutex
    listenerID uint64
    
    // Transactions (3 fields)
    transactionID uint8
    transactionMutex sync.Mutex
    transactionCallbacks map[byte]transactionCallback
    
    // Displays
    displays map[string]*Display
}
```

This struct is the hardware connection, the event bus, the font manager, the display registry, and the CT knob drag state machine — all in one. Every new feature (animations? metrics? new device types?) has to thread through this struct.

### Problem 3: Name Mappings Defined Three Times

The hardware input types (`Button`, `Knob`, `TouchButton`) are defined with constants in `inputs.go`. Their names are then remapped in two other places:

**In `runtime/js/module_ui/module.go`:**
```go
var (
    buttons = map[string]deck.Button{
        "Circle":  deck.Circle,
        "Button1": deck.Button1,
        // ... 8 entries
    }
    touches = map[string]deck.TouchButton{
        "Touch1":  deck.Touch1,
        // ... 12 entries
    }
    knobs = map[string]deck.Knob{
        "Knob1": deck.Knob1,
        // ... 6 entries
    }
)
```

**In `cmd/loupe-js-live/main.go`:**
```go
func buttonName(b loupedeck.Button) string {
    names := map[loupedeck.Button]string{
        loupedeck.Circle:  "Circle",
        loupedeck.Button1: "Button1",
        // ... 8 entries AGAIN
    }
    // ...
}
func touchName(t loupedeck.TouchButton) string { /* same pattern */ }
func knobName(k loupedeck.Knob) string { /* same pattern */ }
```

These three locations all express the same mapping: `Button(7) ↔ "Circle"`. Adding a new button type requires editing three files. This is the kind of coupling that doesn't show up in import graphs but causes real bugs.

### Problem 4: The `Display.Draw()` Method Does Too Much

`display.go`'s `Draw` method (lines 128-185) does four things:
1. Translates coordinates from display-relative to device-relative
2. Encodes the pixel data into RGB565 with endianness handling
3. Creates a framebuffer + draw command pair
4. Decides whether to route through the render scheduler or write directly

Steps 1-2 are hardware protocol. Step 4 is scheduling policy. Step 3 is a command abstraction. These should not all live in one method on a struct that also holds display metadata.

### Problem 5: `connect.go`'s `doConnect` is a Mega-Constructor

`doConnect()` (lines 70-160) creates the `Loupedeck` struct, initializes all 10+ map fields, creates the writer, creates the render scheduler, loads the font, sends reset and brightness commands, and queries version/serial. This is a constructor that knows about every subsystem. Adding a new subsystem requires modifying this function.

---

## What Should Actually Happen

### Principle: The root package should be a hardware driver, not an application framework.

A hardware driver's job:
- Open a connection (serial enumeration, websocket handshake)
- Send commands (message framing, transaction IDs)
- Receive events (parsing, dispatching)
- Manage display protocol (framebuffer encoding, device-specific quirks)

Everything else is application-level and already exists in `runtime/`.

### Concrete Steps

#### Step 1: Add `String()` methods to input types (eliminates triplication)

**File:** `inputs.go`

```go
func (b Button) String() string {
    switch b {
    case Circle:  return "Circle"
    case Button1: return "Button1"
    // ...
    default:      return fmt.Sprintf("Button(%d)", b)
    }
}

func (k Knob) String() string { /* same pattern */ }
func (t TouchButton) String() string { /* same pattern */ }
```

Then add reverse-lookup helpers:

```go
func ParseButton(s string) (Button, error) { /* single lookup map */ }
func ParseKnob(s string) (Knob, error) { /* single lookup map */ }
func ParseTouchButton(s string) (TouchButton, error) { /* single lookup map */ }
```

**Impact:** Delete ~60 lines of duplicate maps from `module_ui/module.go` and ~60 lines from `cmd/loupe-js-live/main.go`. One source of truth for names.

#### Step 2: Move legacy widgets to `loupedeck/legacy`

Move these files into a `legacy/` subdirectory:
- `watchedint.go`
- `intknob.go`
- `multibutton.go`
- `touchdials.go`
- `displayknob.go`

This makes the boundary explicit. `cmd/loupe-feature-tester` imports `legacy`, nothing else does.

**Important nuance:** Don't refactor these files. Don't try to make them use `reactive.Signal`. Just move them. The goal is clarity, not perfection. If the feature tester is the only consumer, the code can stay exactly as-is in a clearly labeled package.

#### Step 3: Extract SVG icon loading to `pkg/svgicons`

`svg_icons.go` (252 lines) is self-contained: it parses HTML, extracts SVG fragments, rasterizes them. It has zero dependency on the `Loupedeck` struct or any hardware concept. It imports `oksvg`, `rasterx`, and stdlib image packages.

Move to `pkg/svgicons/svg_icons.go`. Update `cmd/loupe-svg-buttons` import.

#### Step 4: Extract font management from the `Loupedeck` struct

`loupedeck.go` currently holds:
- `font *opentype.Font`
- `face font.Face`
- `fontdrawer *font.Drawer`
- `SetDefaultFont()` — parses a Go embed font
- `FontDrawer()` — returns a configured drawer
- `Face()` — returns the font face
- `TextInBox()` — auto-sizes text into a bounding box

This font system is used by:
- `touchdials.go` (legacy widget)
- `displayknob.go` (legacy widget)
- `drawCenteredStringAt` / `drawRightJustifiedStringAt` in `touchdials.go`

The new runtime doesn't use it — `runtime/gfx/text.go` has its own text rendering with `basicfont.Face7x13`.

Move font management into the legacy package alongside the widgets that need it. The root package should not be a font manager.

#### Step 5: Simplify the `Loupedeck` struct

After Steps 2-4, the struct should shed:

| Removed | Fields |
|---|---|
| Legacy widgets | `dragDKStarted`, `dragDKStartX`, `dragDKStartY`, `dragDKStartTime`, `dragDKBinding`, `touchDKBindings` |
| Font management | `font`, `face`, `fontdrawer` |
| Legacy single-callback bindings | `buttonBindings`, `buttonUpBindings`, `knobBindings`, `touchBindings`, `touchUpBindings` |

That's ~15 fields removed. The remaining struct has:
- Connection state (`serial`, `conn`, `writer`, `renderer`)
- Subscription-based event routing (`*Listeners` maps, `listenerMutex`, `listenerID`)
- Transaction management (`transactionID`, `transactionMutex`, `transactionCallbacks`)
- Display registry (`displays`)
- Device info (`Vendor`, `Product`, `Model`, `Version`, `SerialNo`)

This is still not tiny, but every remaining field serves the hardware driver role.

#### Step 6: Remove the `Bind*` methods, keep only `On*`

The root package currently has both:
- `BindButton(b, f)` — sets a single "primary" callback (legacy)
- `OnButton(b, f) Subscription` — adds a composable subscription (current)

And the dispatch code checks both:
```go
func (l *Loupedeck) dispatchButton(button Button, status ButtonStatus) bool {
    var primary ButtonFunc         // from BindButton
    listeners := make([]ButtonFunc, 0)  // from OnButton
    // ... checks both maps
}
```

Once the legacy widgets move to `legacy/`, no code outside the package uses `Bind*`. Remove the 5 `Bind*` methods and 5 `*Bindings` maps. Simplify dispatch to only use the subscription maps.

#### Step 7: Move `outboundWriter` and `renderScheduler` to `pkg/`

These are generic queueing/coalescing mechanisms with no dependency on Loupedeck protocol specifics:

- `outboundWriter` takes a `wsConn` interface and an `outboundCommand` interface
- `renderScheduler` takes an `outboundWriter` and a key string

Both are useful infrastructure that shouldn't be buried in the root package. Move to `pkg/writer/` and `pkg/rendersched/` (or keep together in `pkg/pipeline/`).

The `Display` struct would then depend on a `DrawTarget`-like interface from `pkg/`, rather than directly on the writer.

---

## What NOT to Do

### Don't refactor the JS module boilerplate

The previous analysis suggested code generation or reflection helpers for `module_ui/module.go`. This is wrong.

The "boilerplate" in that file is the actual goja binding API. Each `exports.Set("name", func(...))` call is a JavaScript API surface definition. Making it "cleaner" via code generation would:
- Hide the actual JS API behind templates
- Make debugging harder (stack traces point to generated code)
- Add a build step for no user-visible benefit
- Not actually reduce complexity — just move it

The file is 383 lines of straightforward, debuggable code. That's fine.

### Don't try to unify WatchedInt and reactive.Signal

They serve different purposes and different lifecycles. `WatchedInt` goes away with the legacy widgets. `reactive.Signal` is the correct abstraction for the reactive UI system. No bridge needed.

### Don't split `runtime/js/module_*` into subdirectories

The current flat layout is fine. There are 7 modules, each is a single file. Adding directory nesting (`runtime/js/modules/ui/`) would increase import path depth and file count without improving readability. The naming convention `module_<name>` is clear enough.

### Don't extract a "pkg/display" abstraction

The previous analysis suggested creating a `pkg/display/` package with `DrawTarget`, `Renderer`, and `Theme`. But `render.Renderer` already exists and does exactly this. The issue isn't missing abstractions — it's that the root package's `Display.Draw()` bypasses the render layer. Fix the routing, don't add another layer.

---

## Implementation Priority

| Priority | Step | Effort | Impact |
|---|---|---|---|
| **1** | Add `String()`/`Parse*()` to input types | 1 hour | Eliminates 3-way duplication immediately |
| **2** | Move legacy widgets to `legacy/` | 2 hours | Makes god struct obviously smaller |
| **3** | Remove `Bind*` methods + `*Bindings` maps | 1 hour | Simplifies dispatch, removes dead API surface |
| **4** | Extract SVG icons to `pkg/svgicons` | 30 min | Removes largest self-contained subsystem from root |
| **5** | Move font management to legacy | 1 hour | Root package no longer imports opentype |
| **6** | Extract writer + render scheduler to `pkg/` | 2 hours | Clean hardware/framework boundary |
| **7** | Simplify `doConnect` | 1 hour | Constructor only initializes hardware |

**Total: ~8.5 hours of focused work.**

After these steps, the root `loupedeck` package would be:
- `message.go` — protocol types and framing
- `connect.go` — connection lifecycle (simplified)
- `dialer.go` — serial transport
- `listen.go` — event reading + dispatch (simplified)
- `inputs.go` — type definitions with `String()`/`Parse*()`
- `listeners.go` — subscription management (simplified)
- `display.go` — display protocol (framebuffer encoding only)

That's ~1,200 lines for a complete hardware driver. Everything else is in the right place already.

---

## What the Runtime Packages Get Right

The previous analysis didn't acknowledge what's working well. For the record:

**`runtime/reactive/`** — A clean SolidJS-inspired reactive graph with signals, computed values, effects, batch mutations, and automatic dependency tracking. 300 lines total. No global state. No mutexes (intentionally single-threaded). This is well-designed.

**`runtime/gfx/surface.go`** — A thread-safe 8-bit graphics surface with batching, change notifications, and a clean public API. The mutex + condition variable pattern for batching is correct. 385 lines is justified for what it does.

**`runtime/ui/`** — Clean component tree (UI → Page → Display → Tile) with dirty tracking. Display layers are composable. Reactive bindings (`BindText`, `BindIcon`, etc.) integrate naturally with the reactive graph. No mutexes in individual components — dirtiness is tracked centrally by UI. Good design.

**`runtime/host/`** — Clean event routing with composable subscriptions and hot-swappable `EventSource` (via `Attach`). The detach/reattach pattern for device reconnection is smart.

**`pkg/runtimeowner/`** — Correct solution to the goja single-threading constraint. `Call` (blocking) vs `Post` (fire-and-forget), owner context detection via goroutine ID, configurable panic recovery. 230 lines of careful concurrent code that works.

**`pkg/runtimebridge/`** — Minimal, correct. A `sync.Map` from VM to bindings. 50 lines. Not everything needs to be complex.

---

## Dependency Graph After Refactoring

```
Before (current):
                                                                        
  loupedeck (god package)                                               
  ├── message protocol                                                  
  ├── serial transport                                                  
  ├── connection lifecycle                                              
  ├── event types + dispatch                                           
  ├── subscription management                                           
  ├── font management                                                   
  ├── SVG icon parsing                                                  
  ├── legacy widgets (WatchedInt, IntKnob, MultiButton, ...)           
  ├── CT knob widget system (DKWidget, WidgetHolder)                   
  ├── display protocol + render scheduling                              
  └── outbound writer pipeline                                          
       │                                                                
       ▼                                                                
  runtime/                                                              
  ├── js/ ← depends on root package for types + hardware                
  ├── host/ ← depends on root package for EventSource interface         
  ├── ui/ ← depends on reactive/ + gfx/                                 
  ├── render/ ← depends on ui/                                          
  └── ...                                                               


After (proposed):

  loupedeck (hardware driver, ~1200 lines)                              
  ├── message.go      — protocol types, framing, transactions           
  ├── connect.go      — connection lifecycle                            
  ├── dialer.go       — serial transport + enumeration                  
  ├── listen.go       — event reading + simplified dispatch             
  ├── inputs.go       — type definitions + String() + Parse*()          
  ├── listeners.go    — subscription management (On* only)              
  └── display.go      — framebuffer encoding + endianness               

  loupedeck/legacy/ (dead widgets)                                      
  ├── watchedint.go, intknob.go, multibutton.go                         
  ├── touchdials.go, displayknob.go                                     
  └── font.go (extracted from loupedeck.go)                             

  pkg/svgicons/   — SVG parsing and rasterization                       
  pkg/pipeline/   — outboundWriter + renderScheduler                    
  pkg/runtimeowner/ — (unchanged)                                       
  pkg/runtimebridge/ — (unchanged)                                      
  pkg/jsmetrics/     — (unchanged)                                      

  runtime/ (unchanged — already well-structured)                        
  ├── reactive/, ui/, render/, host/, anim/, gfx/, easing/, metrics/    
  └── js/                                                               
```

---

## File-Level Move List

| Current Location | New Location | Notes |
|---|---|---|
| `watchedint.go` | `legacy/watchedint.go` | package `legacy` |
| `intknob.go` | `legacy/intknob.go` | |
| `multibutton.go` | `legacy/multibutton.go` | |
| `touchdials.go` | `legacy/touchdials.go` | includes `drawCenteredStringAt`, `drawRightJustifiedStringAt` |
| `displayknob.go` | `legacy/displayknob.go` | includes `DKWidget`, `DKAnalogWidget`, `WidgetHolder` |
| Font code from `loupedeck.go` | `legacy/font.go` | `SetDefaultFont`, `FontDrawer`, `Face`, `TextInBox`, `drawCenteredStringAt` (root copy) |
| `svg_icons.go` | `pkg/svgicons/svg_icons.go` | |
| `writer.go` | `pkg/pipeline/writer.go` | `outboundWriter`, `WriterOptions`, `WriterStats`, interfaces |
| `renderer.go` | `pkg/pipeline/scheduler.go` | `renderScheduler`, `RenderOptions`, `RenderStats` |

---

## Risks

| Risk | Mitigation |
|---|---|
| `cmd/loupe-feature-tester` breaks | It imports `legacy/` explicitly. Test it. |
| `Display.Draw()` currently routes through render scheduler | After moving scheduler to `pkg/pipeline`, `Display` depends on the interface, not concrete type. Same behavior. |
| Someone is using `Bind*` methods externally | They're not exported in the public API sense — only used internally and by legacy widgets. Safe to remove. |
| The `sources/loupedeck-repo/` directory | This is a vendored copy of an older version of the root package. Don't touch it — it's reference material, not active code. |

---

## Open Questions

1. **Should `legacy/` be `internal/legacy/`?** — Probably yes. Go's `internal` packages prevent external consumers from depending on code you intend to remove.

2. **Should the `Display` struct's `Draw` method accept an interface instead of directly routing to the scheduler?** — Yes, but that's Step 6, not Step 1. Do the moves first, then refine the interfaces.

3. **Is `loupedeck.go`'s `SetButtonColor` the only non-legacy method that uses font/rendering?** — Yes. `SetButtonColor` just sends a protocol message. It stays in the root package.

4. **What about `module_ui/module.go`'s name maps?** — After Step 1, replace the manual maps with `ParseButton(name)`, `ParseKnob(name)`, `ParseTouchButton(name)`. The module becomes cleaner without code generation.
