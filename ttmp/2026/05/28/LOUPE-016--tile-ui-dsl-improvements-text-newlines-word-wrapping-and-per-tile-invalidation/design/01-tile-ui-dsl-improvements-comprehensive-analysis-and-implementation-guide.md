---
Title: Tile UI DSL Improvements — Comprehensive Analysis and Implementation Guide
Ticket: LOUPE-016
Status: active
Topics:
    - javascript
    - goja
    - ui
    - tiles
    - rendering
    - text
    - performance
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - runtime/gfx/text.go
    - runtime/gfx/surface.go
    - runtime/ui/tile.go
    - runtime/ui/display.go
    - runtime/ui/ui.go
    - runtime/render/visual_runtime.go
    - runtime/js/module_ui/module.go
    - runtime/js/module_gfx/module.go
    - runtime/present/runtime.go
    - runtime/reactive/runtime.go
    - runtime/reactive/signal.go
ExternalSources: []
Summary: "Intern-ready analysis, design, and implementation guide for fixing text newlines, adding word wrapping, and enabling per-tile invalidation in the Loupedeck tile UI DSL"
LastUpdated: 2026-05-28T09:00:00-04:00
WhatFor: "Guide implementation of LOUPE-016 tile UI improvements"
WhenToUse: "When implementing text wrapping, newline rendering, or per-tile invalidation features"
---

# Tile UI DSL Improvements — Comprehensive Analysis and Implementation Guide

**Ticket:** LOUPE-016
**Date:** 2026-05-28
**Audience:** New intern — this document assumes Go familiarity but no prior knowledge of this codebase

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Overview: The Big Picture](#system-overview-the-big-picture)
3. [Architecture Deep Dive](#architecture-deep-dive)
4. [Problem 1: Newlines in Text Are Not Rendered](#problem-1-newlines-in-text-are-not-rendered)
5. [Problem 2: No Text Wrapping — Long Text Overflows Tiles](#problem-2-no-text-wrapping--long-text-overflows-tiles)
6. [Problem 3: Full-Display Invalidation Is Slow](#problem-3-full-display-invalidation-is-slow)
7. [Implementation Design](#implementation-design)
8. [API Reference: Current State](#api-reference-current-state)
9. [API Reference: Proposed Changes](#api-reference-proposed-changes)
10. [File Reference Map](#file-reference-map)
11. [Test Plan](#test-plan)
12. [Migration and Backward Compatibility](#migration-and-backward-compatibility)

---

## Executive Summary

This document describes three improvements to the Loupedeck tile UI system:

- **Text with newlines** (`\n`) is silently broken today — Go's `font.Drawer.DrawString()` treats the entire string as a single line, so newline characters render as missing-glyph boxes or invisible gaps. We need a multi-line text renderer that splits on `\n` and draws each line at a properly offset Y position.

- **Text wrapping** does not exist. When text exceeds the width of a tile (90×90 pixels at `basicfont.Face7x13`, roughly 12 characters), it clips or overflows. We need word-wrapping logic that breaks text at word boundaries (with a fallback to character boundaries for extremely long words) within a specified bounding box.

- **Per-tile invalidation** works for the *retained tile path* (`tile.text()`, `tile.icon()`) but is completely absent for the *imperative surface path* (`display.surface()`, `present.onFrame()`). When a script draws custom pixel content onto a single tile's surface and changes just one tile, the entire 360×270 display is re-rendered and re-sent to the hardware. We need per-tile surfaces with per-tile dirty tracking so that a clock script updating one tile only re-renders that 90×90 region.

**Key finding:** The JS `tile.surface()` method already exists in the bridge code (`module_ui/module.go` line 322), and the Go infrastructure for per-tile dirty tracking is fully built. The per-tile invalidation work is primarily about documentation, ergonomics, and example migration, not new infrastructure.

---

## System Overview: The Big Picture

The Loupedeck Live is a hardware controller with a 360×270 pixel main display (divided into a 4×3 grid of 90×90 tiles) and two 60×270 side displays. This codebase provides a Go library and JavaScript runtime for building live UIs on that hardware.

### Hardware Layout

```
┌──────┬────────────────────────────────────┬──────┐
│      │  (0,0)  │ (1,0) │ (2,0) │ (3,0)  │      │
│      ├──────────┼───────┼───────┼────────┤      │
│ LEFT │  (0,1)  │ (1,1) │ (2,1) │ (3,1)  │ RIGHT│
│60×270├──────────┼───────┼───────┼────────┤60×270│
│      │  (0,2)  │ (1,2) │ (2,2) │ (3,2)  │      │
│      │          │       │       │        │      │
└──────┴────────────────────────────────────┴──────┘
          Main display: 360×270 (4 cols × 3 rows)
          Each tile: 90×90 pixels
```

### The Core Idea

You write JavaScript. The Go runtime runs it. The hardware shows the result. The system is split into two rendering philosophies:

1. **Retained mode**: You declare *what* to show (`tile.text("HELLO")`), and Go decides *how* and *when* to render it. This is the simpler path, and it already supports per-tile invalidation.

2. **Imperative mode**: You create pixel surfaces (`gfx.surface(360, 270)`) and draw into them from JavaScript every frame. Go ships the pixels. This gives full control but currently invalidates the entire display every frame.

---

## Architecture Deep Dive

### The Go Rendering Pipeline

The pipeline from JavaScript state change to hardware pixels looks like this:

```
JavaScript
  │
  ▼
┌──────────────────────────────────────┐
│  Reactive Runtime                     │
│  runtime/reactive/                    │
│  - Signal.Set() → notify dependents  │
│  - Effect.run() → re-execute closure │
│  - Batch() → coalesce updates        │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│  UI Layer (retained)                 │
│  runtime/ui/                         │
│  - Tile.SetText() → markDirtyTile()  │
│  - Display.SetSurface() → markDirty  │
│  - UI.DirtyTiles() → list changed   │
│  - UI.DirtyDisplays() → list changed│
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│  Present Runtime                     │
│  runtime/present/runtime.go          │
│  - Invalidate(reason) → wake loop    │
│  - loop() → render() → flush()       │
│  - render: calls JS onFrame callback │
│  - flush: calls renderer.Flush()     │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│  Renderer                            │
│  runtime/render/visual_runtime.go    │
│  - Flush() → DirtyTiles + DirtyDisp │
│  - For dirty tiles: renderTile()    │
│    → 90×90 RGBA image → Draw()      │
│  - For dirty displays:              │
│    renderDisplay() → full RGBA → Draw│
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│  DrawTarget (hardware bridge)        │
│  - Draw(image, xoff, yoff)          │
│  → serializes pixels to device      │
└──────────────────────────────────────┘
```

#### Key Files in the Rendering Pipeline

| File | Lines | Role |
|------|-------|------|
| `runtime/reactive/runtime.go` | 124 | Reactive dependency tracking, batch flush, watcher execution |
| `runtime/reactive/signal.go` | 34 | Mutable reactive value with dependency notification |
| `runtime/reactive/effect.go` | ~40 | Eager side-effect that re-runs when dependencies change |
| `runtime/ui/ui.go` | 222 | UI orchestrator: pages, dirty tile/display sets, dirty handler |
| `runtime/ui/tile.go` | 113 | Single 90×90 tile with text/icon/visible/surface properties |
| `runtime/ui/display.go` | 275 | Display (main/left/right) with text/icon/layers/surfaces/tiles |
| `runtime/ui/page.go` | ~50 | Named page containing displays |
| `runtime/gfx/surface.go` | 382 | Grayscale pixel buffer with drawing primitives, change listeners |
| `runtime/gfx/text.go` | 91 | **THE TEXT RENDERER** — single-line `DrawString` only |
| `runtime/gfx/font.go` | ~100 | Font loading with caching, OpenType/collection support |
| `runtime/render/visual_runtime.go` | 202 | Retained renderer: renders dirty tiles/displays to RGBA, sends via DrawTarget |
| `runtime/present/runtime.go` | 142 | Invalidate-on-demand loop: render → flush → wait |
| `runtime/host/runtime.go` | ~80 | Event routing, timer management, button/touch/knob subscriptions |
| `runtime/host/timers.go` | ~50 | SetTimeout/SetInterval wrappers |
| `runtime/anim/runtime.go` | ~150 | Tween/loop/timeline animation primitives |
| `runtime/js/module_ui/module.go` | 390 | JS bridge for loupedeck/ui module |
| `runtime/js/module_gfx/module.go` | 248 | JS bridge for loupedeck/gfx module |
| `runtime/js/module_present/module.go` | ~60 | JS bridge for loupedeck/present module |
| `runtime/js/module_state/module.go` | ~80 | JS bridge for loupedeck/state module |
| `runtime/js/module_anim/module.go` | ~60 | JS bridge for loupedeck/anim module |
| `runtime/js/env/env.go` | ~60 | Environment struct linking UI, Host, Anim, Present, Metrics |
| `runtime/js/env/bridge.go` | ~20 | Sync.Map for VM→Environment lookup |

### The JavaScript DSL Layer

The JavaScript API is exposed through Go modules that use the [goja](https://github.com/dop251/goja) ECMAScript runtime. Each module is a Go file that:
- Registers a native module name (e.g., `"loupedeck/ui"`)
- Provides a `Loader()` function that creates Go objects and exposes them as JS properties
- Marshals JS values to Go types and vice versa

```
┌────────────────────────────────────────────────────────┐
│  JavaScript (running in goja VM)                       │
│                                                        │
│  const ui = require("loupedeck/ui");     → module_ui  │
│  const gfx = require("loupedeck/gfx");   → module_gfx │
│  const state = require("loupedeck/state");→ module_state│
│  const anim = require("loupedeck/anim"); → module_anim │
│  const present = require("loupedeck/present"); → present│
│  const easing = require("loupedeck/easing"); → easing  │
└────────────────────────────────────────────────────────┘
         │ goja.FunctionCall → Go function call
         ▼
┌────────────────────────────────────────────────────────┐
│  Go Module Bridges                                     │
│                                                        │
│  runtime/js/module_ui/module.go  (390 lines)          │
│  - page(), show(), invalidate()                        │
│  - onButton(), onTouch(), onKnob()                     │
│  - pageObject → displayObject → tileObject             │
│                                                        │
│  runtime/js/module_gfx/module.go  (248 lines)         │
│  - surface(), font()                                   │
│  - surfaceObject: clear, set, add, fillRect, line,     │
│    crosshatch, text, compositeAdd, at, batch            │
└────────────────────────────────────────────────────────┘
```

#### How `tile.text("HELLO")` Actually Works

1. JavaScript calls `tile.text("HELLO")`
2. Go's `module_ui.tileObject` receives this as a `goja.FunctionCall`
3. The argument is either a string (static) or a function (reactive binding)
4. For a static string: Go calls `tile.SetText("HELLO")` directly
5. For a reactive function: Go calls `tile.BindText(func() string { ... })` which creates a reactive `Effect` that re-runs when tracked signals change
6. `tile.SetText()` calls `tile.markDirty()` → `ui.markDirtyTile(tile)` → adds tile to `dirtyTiles` set → calls `onDirty()` handler
7. The `onDirty` handler (set up in `env.Ensure()`) calls `present.Invalidate("ui-dirty")`
8. The present loop wakes, calls `render()` then `flush()`
9. `renderer.Flush()` calls `ui.DirtyTiles()`, finds the changed tile, renders it via `renderTile()`, sends the 90×90 image via `DrawTarget.Draw()`

#### How `surface.text("HELLO", opts)` Actually Works

1. JavaScript creates a surface: `const s = gfx.surface(90, 90)`
2. JavaScript draws text: `s.text("HELLO", { x: 0, y: 0, width: 90, height: 20, center: true })`
3. Go's `module_gfx.surfaceObject` maps this to `surface.Text("HELLO", gfx.TextOptions{...})`
4. `gfx.Surface.Text()` (in `runtime/gfx/text.go`):
   - Creates an `image.Alpha` mask of size `width × height`
   - Creates a `font.Drawer` with that mask as destination
   - Computes a baseline Y position (vertically centered in the height)
   - If `center` is true, computes X offset to center the text
   - Calls `drawer.DrawString(text)` — **single line, no newline handling**
   - Copies the alpha mask pixels into the surface's grayscale pixel buffer, applying brightness scaling
   - Calls `markChangedLocked()` → notifies `OnChange` listeners

### The Reactive Runtime

The reactive runtime is a lightweight signal/computed/effect system inspired by SolidJS:

```
┌─────────────┐     track      ┌─────────────┐     notify    ┌─────────────┐
│   Signal     │ ───────────── │   Effect     │ ◄────────── │   Signal     │
│  (source)    │               │  (consumer)  │             │  (changed)   │
└─────────────┘               └──────┬──────┘             └─────────────┘
                                     │
                                     │ re-runs
                                     ▼
                              ┌──────────────┐
                              │ tile.SetText  │  ← side effect of the effect
                              │ display.SetX  │
                              └──────────────┘
```

**How it works:**

- When a `Signal.Get()` is called inside an `Effect.run()`, the signal registers the effect as a dependent
- When `Signal.Set()` is called, it notifies all dependent effects by marking them dirty
- `Runtime.Flush()` (or `maybeFlush()` when not batching) runs all pending dirty effects
- `Runtime.Batch(fn)` groups multiple signal mutations so effects only run once after the batch completes

**Key implication for our work:** Reactive bindings (`tile.text(() => ...)`) automatically create the right dirty-tracking. When we add per-tile surfaces, the reactive binding pattern will work naturally — a signal change will mark the tile dirty, and only that tile will be re-rendered.

### The Presentation Loop

The present runtime (`runtime/present/runtime.go`) is the frame scheduler:

```
                 ┌──────────────┐
  Invalidate() ──│  dirty=true  │
                 │  wake channel│
                 └──────┬───────┘
                        │
                        ▼
               ┌────────────────┐
               │   loop()       │
               │                │
               │  1. render()  │  ← calls JS onFrame callback OR is a no-op
               │  2. flush()   │  ← calls renderer.Flush() → sends pixels to hardware
               │  3. loop back │
               └────────────────┘
```

**Critical detail:** The present runtime has **no concept of regions**. It's binary: either the frame is dirty or it isn't. The `Invalidate(reason)` method takes a reason string for debugging/metrics, but the reason does not affect what gets re-rendered. The `renderer.Flush()` method decides what to re-render based on `DirtyTiles()` and `DirtyDisplays()`.

This means: **per-tile invalidation already exists in the renderer, but only for retained tiles.** The gap is that the surface/imperative path often bypasses the per-tile dirty tracking by using display-level surfaces.

---

## Problem 1: Newlines in Text Are Not Rendered

### The Bug

When you pass text containing `\n` to any text rendering API, the newline characters are not handled. Here is exactly what happens:

**In the retained renderer** (`runtime/render/visual_runtime.go`, `drawCenteredLabel`):

```go
func drawCenteredLabel(dst draw.Image, text string, baseline int, fg color.Color) {
    face := basicfont.Face7x13
    d := &font.Drawer{
        Dst:  dst,
        Src:  &image.Uniform{fg},
        Face: face,
    }
    width := d.MeasureString(text).Round()
    x := (dst.Bounds().Dx() - width) / 2
    d.Dot = fixed.P(x, baseline)
    d.DrawString(text)  // ← treats entire string as one line
}
```

**In the surface text API** (`runtime/gfx/text.go`, `Surface.Text`):

```go
d := &font.Drawer{
    Dst:  alpha,
    Src:  image.White,
    Face: face,
}
// ... compute baseline ...
d.Dot = fixed.P(x, baseline)
d.DrawString(text)  // ← same problem: single-line DrawString
```

Go's `font.Drawer.DrawString()` does **not** interpret `\n`. It advances the dot horizontally for each glyph, but newline characters are either rendered as missing glyphs (tofu □) or ignored entirely, depending on the font face. The result is garbled or missing output.

### What the user expects

```javascript
tile.text("LINE1\nLINE2");
```

Should render:

```
┌─────────┐
│  LINE1  │
│  LINE2  │
└─────────┘
```

What actually renders:

```
┌─────────┐
│ LINE1□LINE2 │  ← □ = missing glyph for \n, text overflows tile
└─────────┘
```

### Where it matters

- **Retained tile path**: `tile.text("A\nB")` → `drawCenteredLabel()` → garbled
- **Surface text path**: `surface.text("A\nB", { ... })` → `gfx.Surface.Text()` → garbled
- **Display text**: `display.text("A\nB")` → `drawCenteredLabel()` → garbled

All three code paths need the same fix.

---

## Problem 2: No Text Wrapping — Long Text Overflows Tiles

### The Bug

There is no word-wrapping logic anywhere in the rendering stack. When text is wider than the available width, it simply extends past the tile boundary. The `Width` field in `TextOptions` is only used for the alpha mask dimensions — it does not cause line breaks.

### What the user expects

```javascript
tile.text("HELLO WORLD FROM LOUPEDECK");
```

Should render within a 90×90 tile as:

```
┌─────────┐
│  HELLO   │
│  WORLD   │
│  FROM    │
│ LOUPEDECK│
└─────────┘
```

What actually renders:

```
┌─────────┐
│ HELLO WORLD FROM LOUPEDECK  ← runs off edge, clipped by image bounds
└─────────┘
```

### Why it matters for tiles

A tile is only 90 pixels wide. With `basicfont.Face7x13` (7px per character), you get roughly 12 characters per line. Any text longer than that is invisible beyond the tile edge.

### Current text measurement

Go's `font.Drawer.MeasureString(text)` returns the width of a string in fixed-point pixels. This is available and correct — it's just never used for wrapping decisions.

---

## Problem 3: Full-Display Invalidation Is Slow

### The Problem

When using the imperative surface path (drawing pixels manually from JS), any change to *any* pixel on the display causes the *entire* 360×270 display to be re-rendered and re-sent to the hardware. This is approximately 97,200 pixels per frame, or about 388,800 bytes of RGBA data.

For a clock that updates one tile per second, this means sending 12× more data than necessary.

### The Two Paths and Their Invalidation Behavior

```
                    ┌──────────────────────┐
                    │  What changed?        │
                    └───────┬──────────────┘
                            │
                ┌───────────┴───────────┐
                │                       │
        Retained tile path      Surface/imperative path
        (tile.text, tile.icon)  (display.surface + present.onFrame)
                │                       │
                ▼                       ▼
        markDirtyTile()          display.markDirty()
                │                       │
                ▼                       ▼
        Only changed tile        Entire display marked dirty
        is in DirtyTiles()       is in DirtyDisplays()
                │                       │
                ▼                       ▼
        renderTile() sends        renderDisplay() sends
        90×90 = 8,100px          360×270 = 97,200px
        ≈ 32KB RGBA               ≈ 389KB RGBA
                │                       │
                ▼                       ▼
        FAST ✓                   SLOW ✗
```

### Why scripts use the surface path

The retained tile path only supports static text and icon strings. For anything custom — animations, charts, patterns, kanji, pixel art — scripts must create their own `gfx.surface()` and draw into it. The example `11-cyb-os-tiles.js` creates three full-size surfaces (left, main, right) and redraws all of them every 2 seconds.

### Per-tile surfaces — Already Works!

The `tile.surface()` method is **already wired** in the JS bridge at `module_ui/module.go` line 322:

```go
// runtime/js/module_ui/module.go line 322
_ = obj.Set("surface", func(call goja.FunctionCall) goja.Value {
    arg := call.Argument(0)
    if goja.IsNull(arg) || goja.IsUndefined(arg) {
        tile.SetSurface(nil)
    } else {
        tile.SetSurface(module_gfx.SurfaceFromValue(arg, runtime))
    }
    return goja.Undefined()
})
```

And the Go infrastructure is fully built:

```go
// runtime/ui/tile.go — already supports per-tile surfaces
func (t *Tile) SetSurface(surface *gfx.Surface) {
    // ...
    t.surfaceSub = surface.OnChange(func() {
        t.markDirty()  // ← Only THIS tile is marked dirty!
    })
    // ...
}

// runtime/render/visual_runtime.go — already renders tile surfaces
func (r *Renderer) renderTile(tile *ui.Tile) image.Image {
    // ...
    if surface := tile.Surface(); surface != nil {
        return surface.ToRGBA(r.Theme.Foreground, r.Theme.Background)
    }
    // ...
}
```

This means **a script can already do this today:**

```javascript
const gfx = require("loupedeck/gfx");
const ui = require("loupedeck/ui");

const tileSurface = gfx.surface(90, 90);

ui.page("clock", page => {
    page.tile(0, 0, tile => {
        tile.surface(tileSurface);  // ← already works!
    });
});

// Later, modify tileSurface and only tile (0,0) is re-rendered
tileSurface.clear(0);
tileSurface.text("12:34", { x: 0, y: 20, width: 90, height: 20, center: true });
```

**The real problem is that this path is undocumented and no examples use it.** Users reach for the display-level surface path because that's what the examples show. The fix has three parts:

1. **Documentation**: Update the API reference and examples to show per-tile surface usage
2. **Ergonomics**: Add a convenience method so scripts don't need to manually create 90×90 surfaces
3. **Example migration**: Update example scripts to use per-tile surfaces where possible

### What about display-level animation effects?

Some scripts (like `11-cyb-os-tiles.js`) draw full-display animation effects such as ripples and scanlines that span across tiles. These genuinely need a full-display surface. The per-tile surface approach doesn't replace display-level surfaces — it complements them. The recommendation is:

- **Use per-tile surfaces** when each tile's content is independent (clocks, meters, status indicators)
- **Use display-level surfaces** when content spans tile boundaries (ripples, scanlines, cross-tile animations)
- **Hybrid approach**: Use display-level surfaces for background effects, and per-tile surfaces for the foreground content of each tile (via display layers)

The display layer system already supports this:

```go
// runtime/ui/display.go
func (d *Display) SetLayer(name string, surface *gfx.Surface) { ... }
```

Layers composite on top of each other. A script could:

1. Set a display-level surface for the background effect (ripples, scanlines)
2. Set per-tile surfaces for each tile's foreground content
3. Get per-tile invalidation for the content tiles while the background layer is a shared surface

This is already supported by the `renderDisplayLayers` path in the renderer.

---

## Implementation Design

### Design 1: Multi-Line Text Rendering

#### Strategy

Split the input string on `\n` and render each line separately at incrementing Y offsets.

#### Where to change

We need to fix **two** rendering call sites:

1. `runtime/gfx/text.go` — `Surface.Text()` — used by the surface text API
2. `runtime/render/visual_runtime.go` — `drawCenteredLabel()` — used by the retained renderer for tile and display text

#### Pseudocode for `Surface.Text()` multi-line support

```go
func (s *Surface) Text(text string, opts TextOptions) {
    if s == nil || text == "" {
        return
    }
    face := opts.Face
    if face == nil {
        face = basicfont.Face7x13
    }

    // Split on newlines
    lines := strings.Split(text, "\n")

    lineHeight := face.Metrics().Height.Ceil()
    if lineHeight < 1 {
        lineHeight = 13 // fallback for basicfont
    }
    lineGap := opts.LineGap
    if lineGap <= 0 {
        lineGap = lineHeight / 4  // auto inter-line gap
    }

    w := opts.Width
    if w <= 0 {
        w = s.width
    }
    h := opts.Height
    if h <= 0 {
        h = len(lines) * (lineHeight + lineGap) - lineGap + 4
    }

    brightness := opts.Brightness
    if brightness == 0 {
        brightness = 255
    }

    alpha := image.NewAlpha(image.Rect(0, 0, w, h))
    draw.Draw(alpha, alpha.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

    d := &font.Drawer{
        Dst:  alpha,
        Src:  image.White,
        Face: face,
    }

    // Compute total block height for vertical centering
    totalTextHeight := len(lines) * (lineHeight + lineGap) - lineGap
    startY := (h - totalTextHeight) / 2 + face.Metrics().Ascent.Ceil()
    if startY < face.Metrics().Ascent.Ceil() {
        startY = face.Metrics().Ascent.Ceil()
    }

    for i, line := range lines {
        if line == "" {
            continue  // empty line, skip but still advance Y
        }
        baseline := startY + i * (lineHeight + lineGap)
        if baseline > h - face.Metrics().Descent.Ceil() {
            break  // no more room
        }
        x := 0
        if opts.Center {
            lineW := d.MeasureString(line).Round()
            x = (w - lineW) / 2
            if x < 0 {
                x = 0
            }
        }
        d.Dot = fixed.P(x, baseline)
        d.DrawString(line)
    }

    // Copy alpha mask to surface pixels (same as current code)
    s.mu.Lock()
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            a := alpha.AlphaAt(x, y).A
            if a == 0 {
                continue
            }
            px := opts.X + x
            py := opts.Y + y
            if !s.inBounds(px, py) {
                continue
            }
            v := scaleUint8(a, brightness)
            s.addLocked(px, py, v)
        }
    }
    listeners := s.markChangedLocked()
    s.mu.Unlock()
    notifyListeners(listeners)
}
```

#### Pseudocode for `drawCenteredLabel()` multi-line support

```go
func drawCenteredLabel(dst draw.Image, text string, baseline int, fg color.Color) {
    if text == "" {
        return
    }
    face := basicfont.Face7x13
    lines := strings.Split(text, "\n")
    lineHeight := face.Metrics().Height.Ceil()
    lineGap := lineHeight / 4

    for i, line := range lines {
        if line == "" {
            continue
        }
        y := baseline + i * (lineHeight + lineGap)
        d := &font.Drawer{
            Dst:  dst,
            Src:  &image.Uniform{fg},
            Face: face,
        }
        width := d.MeasureString(line).Round()
        x := (dst.Bounds().Dx() - width) / 2
        if x < 0 {
            x = 0
        }
        d.Dot = fixed.P(x, y)
        d.DrawString(line)
    }
}
```

#### JS API: No changes needed for newlines

The existing `tile.text("LINE1\nLINE2")` and `surface.text("A\nB", opts)` APIs will just work once the Go implementation handles newlines. JavaScript `\n` in string literals is a real newline character, and Go receives it correctly via goja.

#### TextOptions struct changes

Add `LineGap` and `WrapWidth` fields to `TextOptions`:

```go
type TextOptions struct {
    X          int
    Y          int
    Width      int
    Height     int
    Brightness uint8
    Face       font.Face
    Center     bool
    LineGap    int     // NEW: pixels between lines (0 = auto = lineHeight/4)
    WrapWidth  int     // NEW: enable word wrapping at this width (0 = no wrap)
}
```

---

### Design 2: Word-Wrap Text Layout

#### Strategy

Add a `wrapText` function that breaks a string into lines that fit within a specified pixel width, respecting word boundaries, with a character-break fallback for words wider than the available width.

#### Where to add

Add the wrapping logic to `runtime/gfx/text.go` as a helper function. Both `Surface.Text()` and `drawCenteredLabel()` will use it.

#### Word-wrap algorithm (pseudocode)

```
function wrapText(text, face, maxWidth):
    if maxWidth <= 0:
        return [text]  // no wrapping

    paragraphs = text.split("\n")
    result = []

    for paragraph in paragraphs:
        if paragraph == "":
            result.append("")
            continue

        words = paragraph.split(" ")
        currentLine = ""

        for word in words:
            if currentLine == "":
                currentLine = word
                continue

            testLine = currentLine + " " + word
            testWidth = measureString(face, testLine)

            if testWidth <= maxWidth:
                currentLine = testLine
            else:
                // Word doesn't fit on current line
                wordWidth = measureString(face, word)
                if wordWidth > maxWidth:
                    // Force character-level break
                    if currentLine != "":
                        result.append(currentLine)
                        currentLine = ""
                    for char in word:
                        if measureString(face, currentLine + char) > maxWidth:
                            result.append(currentLine)
                            currentLine = char
                        else:
                            currentLine += char
                else:
                    // Start a new line with this word
                    result.append(currentLine)
                    currentLine = word

        if currentLine != "":
            result.append(currentLine)

    return result
```

#### Go implementation sketch

```go
package gfx

import (
    "strings"
    "golang.org/x/image/font"
)

// wrapText breaks text into lines that fit within maxWidth pixels.
// It first splits on \n, then wraps each paragraph at word boundaries.
// Words wider than maxWidth are force-broken at character boundaries.
func wrapText(text string, face font.Face, maxWidth int) []string {
    if maxWidth <= 0 || text == "" {
        return []string{text}
    }

    var lines []string
    paragraphs := strings.Split(text, "\n")

    for _, para := range paragraphs {
        if para == "" {
            lines = append(lines, "")
            continue
        }

        words := strings.Fields(para)
        if len(words) == 0 {
            lines = append(lines, "")
            continue
        }

        var currentLine string
        drawer := &font.Drawer{Face: face}

        for _, word := range words {
            if currentLine == "" {
                currentLine = word
                continue
            }
            test := currentLine + " " + word
            if drawer.MeasureString(test).Round() <= maxWidth {
                currentLine = test
            } else {
                wordWidth := drawer.MeasureString(word).Round()
                if wordWidth > maxWidth && currentLine != "" {
                    // Emit current line, force-break the word
                    lines = append(lines, currentLine)
                    currentLine = ""
                    // Character-level break
                    for _, r := range word {
                        ch := string(r)
                        if currentLine != "" &&
                           drawer.MeasureString(currentLine + ch).Round() > maxWidth {
                            lines = append(lines, currentLine)
                            currentLine = ch
                        } else {
                            currentLine += ch
                        }
                    }
                } else {
                    lines = append(lines, currentLine)
                    currentLine = word
                }
            }
        }
        if currentLine != "" {
            lines = append(lines, currentLine)
        }
    }

    return lines
}
```

#### Integration with Surface.Text()

```go
func (s *Surface) Text(text string, opts TextOptions) {
    // ... face, width, height setup ...

    // Determine effective wrap width
    wrapWidth := opts.WrapWidth
    if wrapWidth <= 0 && opts.Width > 0 {
        // If the caller specified a width but not a wrap width,
        // wrapping is still off by default (backward compatible).
        // Users must explicitly set wrapWidth > 0.
    }

    // Get lines (either from newlines only, or from word wrapping)
    var lines []string
    if wrapWidth > 0 {
        lines = wrapText(text, face, wrapWidth)
    } else {
        lines = strings.Split(text, "\n")
    }

    // ... rest of multi-line rendering as in Design 1 ...
}
```

#### Integration with drawCenteredLabel()

The retained renderer calls `drawCenteredLabel` with a fixed baseline and the display/tile width. For tiles, the available text area is approximately 90px wide (the full tile width). We need to add wrap width awareness:

```go
// New function with wrap support
func drawWrappedLabel(dst draw.Image, text string, baseline int, fg color.Color, wrapWidth int) {
    if text == "" {
        return
    }
    face := basicfont.Face7x13

    var lines []string
    if wrapWidth > 0 {
        lines = wrapText(text, face, wrapWidth)
    } else {
        lines = strings.Split(text, "\n")
    }

    lineHeight := face.Metrics().Height.Ceil()
    lineGap := lineHeight / 4

    for i, line := range lines {
        if line == "" {
            continue
        }
        y := baseline + i * (lineHeight + lineGap)
        d := &font.Drawer{
            Dst:  dst,
            Src:  &image.Uniform{fg},
            Face: face,
        }
        width := d.MeasureString(line).Round()
        x := (dst.Bounds().Dx() - width) / 2
        if x < 0 {
            x = 0
        }
        d.Dot = fixed.P(x, y)
        d.DrawString(line)
    }
}

// Backward-compatible wrapper (no wrap by default)
func drawCenteredLabel(dst draw.Image, text string, baseline int, fg color.Color) {
    drawWrappedLabel(dst, text, baseline, fg, 0)
}
```

#### JS API: Text wrapping option

For the surface text API, add a `wrapWidth` option:

```javascript
surface.text("Long text that should wrap", {
    x: 0,
    y: 0,
    width: 90,
    height: 90,
    wrapWidth: 88,  // NEW: wrap text at this pixel width
    center: true,
});
```

For the retained tile text API, add an options parameter:

```javascript
// Current: static string or reactive function
tile.text("Hello");
tile.text(() => `Count: ${count.get()}`);

// Proposed: object form with wrap option
tile.text("Long text that should wrap", { wrap: true });
tile.text(() => `Status: ${status.get()}`, { wrap: true });
```

The Go bridge would detect whether the first argument is a string, function, or object and dispatch accordingly.

---

### Design 3: Per-Tile Surface Drawing and Invalidation

#### Strategy

The infrastructure already exists. The work is:
1. Document the existing `tile.surface()` API
2. Add ergonomic helpers for common per-tile patterns
3. Write examples demonstrating per-tile surface usage
4. Consider adding `tile.draw(fn)` as a shorthand for creating a surface and drawing into it

#### Proposed `tile.draw(fn)` ergonomic helper

The most common pattern for per-tile custom drawing is:

```javascript
// Current (verbose)
const tileSurface = gfx.surface(90, 90);
tile.surface(tileSurface);
tileSurface.clear(0);
tileSurface.text("12:34", { ... });

// Proposed (ergonomic)
tile.draw(surface => {
    surface.clear(0);
    surface.text("12:34", { x: 0, y: 20, width: 90, height: 20, center: true });
});
```

The `tile.draw(fn)` method would:
1. Create a 90×90 surface if one doesn't exist on the tile
2. Pass that surface to the callback
3. After the callback returns, the surface's `OnChange` listener marks the tile dirty

#### Go implementation for `tile.draw(fn)`

In `runtime/ui/tile.go`:

```go
func (t *Tile) Draw(fn func(*gfx.Surface)) {
    if t.surface == nil {
        t.SetSurface(gfx.NewSurface(TileWidth, TileHeight))
    }
    fn(t.surface)
}
```

In `runtime/js/module_ui/module.go`, add to `tileObject`:

```go
_ = obj.Set("draw", func(call goja.FunctionCall) goja.Value {
    fn, ok := goja.AssertFunction(call.Argument(0))
    if !ok {
        panic(runtime.NewTypeError("tile.draw requires a function"))
    }
    if tile.Surface() == nil {
        tile.SetSurface(gfx.NewSurface(ui.TileWidth, ui.TileHeight))
    }
    surface := tile.Surface()
    surfaceObj := module_gfx.SurfaceObject(runtime, surface) // need to expose this
    _, err := fn(goja.Undefined(), surfaceObj)
    if err != nil {
        panic(runtime.NewGoError(err))
    }
    return goja.Undefined()
})
```

Note: `module_gfx.surfaceObject` is currently a private function. We would need to either export it or create a shared helper.

#### Per-tile reactive drawing pattern

For reactive per-tile updates (e.g., a clock), the recommended pattern would be:

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const anim = require("loupedeck/anim");

const seconds = state.signal(0);

ui.page("clock", page => {
    page.tile(0, 0, tile => {
        // Bind text reactively — only this tile invalidates
        tile.text(() => {
            const d = new Date();
            return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
        });
    });
});

// Animation loop updates the signal every second
anim.loop(1000, () => {
    const d = new Date();
    seconds.set(d.getSeconds());
});

ui.show("clock");
```

With this pattern:
- `anim.loop()` calls the JS callback every 1000ms
- The callback updates the `seconds` signal
- The reactive binding in `tile.text(() => ...)` re-evaluates
- `tile.SetText()` is called → `markDirtyTile()` → only tile (0,0) is re-rendered
- `renderer.Flush()` renders only that tile → sends 90×90 = 32KB instead of 389KB

**This is 12× less data per update.**

For custom pixel drawing, the pattern is:

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");

const level = state.signal(0);

ui.page("meter", page => {
    page.tile(0, 0, tile => {
        // Create per-tile surface
        const tileSurface = gfx.surface(90, 90);
        tile.surface(tileSurface);

        // Bind level reactively
        // NOTE: We need a "draw on change" pattern here
        // Currently, the JS author must arrange this themselves.
    });
});

// Option: use present.onFrame to check if level changed
// and redraw only the tile surface
let lastLevel = -1;
present.onFrame(() => {
    const current = level.get();
    if (current !== lastLevel) {
        tileSurface.clear(0);
        // draw the meter
        lastLevel = current;
    }
});
```

The ergonomic gap here is that there's no built-in way to say "when signal X changes, redraw tile Y's surface." This is where `tile.draw(fn)` with reactive bindings could help. A future extension could support:

```javascript
// Hypothetical: reactive tile drawing
tile.draw(surface => {
    surface.clear(0);
    const l = level.get();  // signal dependency tracked
    surface.fillRect(10, 90 - l, 70, l, 200);
});
```

The `tile.draw(fn)` callback would be wrapped in a reactive `Watch()` so that when `level.get()` changes, the callback re-runs, redraws the surface, and the tile is automatically marked dirty. This is the ideal ergonomic API but requires careful design around when to re-run the draw callback.

---

## API Reference: Current State

### `loupedeck/ui` module — current exports

| Export | Type | Description |
|--------|------|-------------|
| `page(name, fn)` | function | Create/configure a named page |
| `show(name)` | function | Switch to a named page |
| `invalidate(reason)` | function | Mark the current frame as needing re-render |
| `onButton(name, fn)` | function | Subscribe to button events |
| `onTouch(name, fn)` | function | Subscribe to touch events |
| `onKnob(name, fn)` | function | Subscribe to knob events |

### Page object — current methods

| Method | Description |
|--------|-------------|
| `page.tile(col, row, fn)` | Create/configure a tile on the main display |
| `page.display(name, fn)` | Create/configure a display (left/main/right) |
| `page.AddTile(col, row)` | Go-only: add tile without callback |

### Tile object — current methods

| Method | JS signature | Description |
|--------|-------------|-------------|
| `tile.text(valueOrFn)` | `tile.text("HELLO")` or `tile.text(() => ...)` | Set or bind text |
| `tile.icon(valueOrFn)` | `tile.icon("circle")` or `tile.icon(() => ...)` | Set or bind icon |
| `tile.visible(boolOrFn)` | `tile.visible(true)` or `tile.visible(() => ...)` | Set or bind visibility |
| `tile.surface(surfaceOrNull)` | `tile.surface(s)` or `tile.surface(null)` | Set or clear per-tile surface (**already works!**) |

### Display object — current methods

| Method | Description |
|--------|-------------|
| `display.text(valueOrFn)` | Set or bind display text |
| `display.icon(valueOrFn)` | Set or bind display icon |
| `display.visible(boolOrFn)` | Set or bind display visibility |
| `display.surface(surfaceOrNull)` | Set or clear display-level surface |
| `display.layer(name, surface, opts?)` | Set/clear a named compositing layer |
| `display.tile(col, row, fn)` | Create/configure a tile (main display only) |

### `loupedeck/gfx` module — current exports

| Export | Type | Description |
|--------|------|-------------|
| `surface(width, height)` | function | Create a new grayscale pixel surface |
| `font(path, opts)` | function | Load an OpenType font |

### Surface object — current methods

| Method | Description |
|--------|-------------|
| `surface.width` | Get surface width |
| `surface.height` | Get surface height |
| `surface.clear(v)` | Fill all pixels with value |
| `surface.batch(fn)` | Group operations, coalesce change notifications |
| `surface.set(x, y, v)` | Set a pixel value |
| `surface.add(x, y, v)` | Add to a pixel value (saturating) |
| `surface.fillRect(x, y, w, h, v)` | Fill a rectangle |
| `surface.line(x1, y1, x2, y2, v)` | Draw a line (Bresenham) |
| `surface.crosshatch(x, y, w, h, density, v)` | Draw a crosshatch pattern |
| `surface.text(text, opts)` | Draw text (see options below) |
| `surface.compositeAdd(other, x, y)` | Composite another surface |
| `surface.at(x, y)` | Read pixel value |

### Surface text options — current fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `x` | int | 0 | X offset within surface |
| `y` | int | 0 | Y offset within surface |
| `width` | int | surface width | Alpha mask width |
| `height` | int | auto | Alpha mask height |
| `brightness` | int | 255 | Brightness multiplier (0–255) |
| `center` | bool | false | Center text horizontally |
| `font` | font object | basicfont.Face7x13 | Custom font to use |

---

## API Reference: Proposed Changes

### `surface.text()` — new options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `lineGap` | int | 0 (auto) | Pixels between lines. 0 = auto = lineHeight/4 |
| `wrapWidth` | int | 0 (off) | Enable word wrapping at this pixel width. 0 = no wrapping |

**Example:**

```javascript
surface.text("HELLO WORLD FROM LOUPEDECK", {
    x: 0,
    y: 0,
    width: 90,
    height: 90,
    wrapWidth: 88,
    center: true,
    lineGap: 2,
});
```

### `tile.text()` — new object form

Currently:
```javascript
tile.text("string")           // static
tile.text(() => ...)          // reactive
```

Proposed addition:
```javascript
tile.text("string", { wrap: true })           // static with wrap
tile.text(() => ..., { wrap: true })          // reactive with wrap
```

When `wrap: true`, the retained renderer uses the tile width (90px minus padding) as the wrap width.

### `tile.draw()` — new method

```javascript
tile.draw(surface => {
    surface.clear(0);
    surface.fillRect(10, 50, 70, 30, 150);
    surface.text("HELLO", { x: 0, y: 20, width: 90, height: 20, center: true });
});
```

Creates a per-tile 90×90 surface if one doesn't exist, passes it to the callback. After the callback, the tile is automatically marked dirty.

### `tile.invalidate()` — new method

```javascript
tile.invalidate();  // Mark this tile as dirty, forcing re-render
```

This is useful when a script has modified a tile's surface outside of a `tile.draw()` callback and wants to ensure the tile is re-rendered on the next flush.

---

## File Reference Map

### Files that must be modified for newlines + wrapping

| File | Change | Complexity |
|------|--------|------------|
| `runtime/gfx/text.go` | Add `LineGap`, `WrapWidth` fields to `TextOptions`. Add `wrapText()` function. Modify `Surface.Text()` to split on `\n` and apply word wrapping. | Medium |
| `runtime/render/visual_runtime.go` | Modify `drawCenteredLabel()` to split on `\n` and render multiple lines. Add `drawWrappedLabel()` with wrap support. Modify `renderTile()` and `renderDisplay()` to pass wrap width when appropriate. | Medium |
| `runtime/js/module_gfx/module.go` | Add `lineGap` and `wrapWidth` property parsing in `textOptionsFromValue()`. Add `intProp(obj, "lineGap")` and `intProp(obj, "wrapWidth")` calls. | Easy |
| `runtime/js/module_ui/module.go` | Modify `tile.text()` to accept an optional second argument (options object with `wrap` boolean). Add `tile.draw()` method. Add `tile.invalidate()` method. | Medium |

### Files that need modifications for per-tile ergonomics

| File | Change | Complexity |
|------|--------|------------|
| `runtime/ui/tile.go` | Add `Draw(fn func(*gfx.Surface))` method. Export `TileWidth` and `TileHeight` constants or make them accessible from the tile. | Easy |
| `runtime/js/module_ui/module.go` | Add `tile.draw()` JS bridge. Add `tile.invalidate()` JS bridge. May need to expose `module_gfx.surfaceObject()` as a public helper. | Medium |
| `runtime/render/visual_runtime.go` | May need to pass tile wrap width to `drawWrappedLabel()` for retained tile text rendering. | Easy |

### Files that need documentation updates

| File | Change |
|------|--------|
| `docs/help/topics/01-loupedeck-js-api-reference.md` | Document `tile.surface()`, `tile.draw()`, `tile.invalidate()`, `surface.text()` wrap options, newline support |
| `examples/js/` | Add a new example demonstrating per-tile surfaces, e.g., `13-per-tile-clock.js` |

### No changes needed

| File | Reason |
|------|--------|
| `runtime/gfx/surface.go` | The surface data structure is fine. `OnChange` already works. |
| `runtime/ui/ui.go` | Per-tile dirty tracking already works. `DirtyTiles()` and `ClearDirtyTiles()` are correct. |
| `runtime/present/runtime.go` | The present loop is region-agnostic and works correctly. |
| `runtime/reactive/` | The reactive runtime works correctly for per-tile invalidation. |

---

## Test Plan

### Unit tests for newline rendering

**File:** `runtime/gfx/text_test.go` (new file)

```go
func TestSurfaceTextMultiLine(t *testing.T) {
    s := NewSurface(80, 40)
    s.Text("LINE1\nLINE2", TextOptions{
        X:      0,
        Y:      0,
        Width:  80,
        Height: 40,
        Center: true,
    })
    // Verify that pixels are drawn at two distinct Y ranges
    topRow, bottomRow := 0, 0
    for y := 0; y < 20; y++ {
        for x := 0; x < 80; x++ {
            if s.At(x, y) > 0 { topRow++; break }
        }
    }
    for y := 20; y < 40; y++ {
        for x := 0; x < 80; x++ {
            if s.At(x, y) > 0 { bottomRow++; break }
        }
    }
    if topRow == 0 || bottomRow == 0 {
        t.Fatal("expected text on both lines when \\n is used")
    }
}
```

### Unit tests for word wrapping

**File:** `runtime/gfx/text_test.go`

```go
func TestWrapTextBreaksAtWordBoundaries(t *testing.T) {
    face := basicfont.Face7x13
    lines := wrapText("HELLO WORLD", face, 40) // 40px ≈ 5 chars
    if len(lines) < 2 {
        t.Fatalf("expected word wrap to produce multiple lines, got %d", len(lines))
    }
    if lines[0] != "HELLO" {
        t.Fatalf("expected first wrapped line to be HELLO, got %q", lines[0])
    }
}

func TestWrapTextPreservesNewlines(t *testing.T) {
    face := basicfont.Face7x13
    lines := wrapText("A\nB C", face, 1000) // wide enough, no wrapping needed
    if len(lines) != 2 || lines[0] != "A" || lines[1] != "B C" {
        t.Fatalf("expected newline to be preserved, got %v", lines)
    }
}

func TestWrapTextForceBreaksLongWords(t *testing.T) {
    face := basicfont.Face7x13
    lines := wrapText("SUPERLONGWORD", face, 30) // 30px ≈ 4 chars
    if len(lines) < 3 {
        t.Fatalf("expected force-break of long word, got %d lines: %v", len(lines), lines)
    }
}
```

### Unit tests for per-tile surface invalidation

**File:** `runtime/ui/ui_test.go` (additions)

```go
func TestTileSurfaceChangeOnlyInvalidatesThatTile(t *testing.T) {
    rt := reactive.NewRuntime()
    ui := New(rt)
    page := ui.AddPage("home")

    tileA := page.AddTile(0, 0)
    tileB := page.AddTile(1, 0)

    surfaceA := gfx.NewSurface(90, 90)
    tileA.SetSurface(surfaceA)

    surfaceB := gfx.NewSurface(90, 90)
    tileB.SetSurface(surfaceB)

    if err := ui.Show("home"); err != nil {
        t.Fatalf("show: %v", err)
    }
    ui.ClearDirty()

    // Only modify tile A's surface
    surfaceA.FillRect(0, 0, 10, 10, 100)

    dirty := ui.DirtyTiles()
    if len(dirty) != 1 || dirty[0] != tileA {
        t.Fatalf("expected only tileA dirty, got %v", dirty)
    }
}
```

### Integration test script

**File:** `ttmp/.../scripts/03-multi-line-wrap-probe.js`

A test script that exercises all three features:
- Multi-line text in retained tiles
- Word-wrapped text in surfaces
- Per-tile surface invalidation with a clock

---

## Migration and Backward Compatibility

### Backward compatibility guarantees

- **Newline handling**: Adding `\n` splitting is backward compatible because any text that previously worked (no `\n`) produces identical output. Text with `\n` was already broken, so fixing it doesn't change working behavior.
- **Word wrapping**: Must be opt-in. The default `WrapWidth: 0` means "no wrapping," preserving current behavior. Scripts must explicitly set `wrapWidth > 0` or `{ wrap: true }` to activate wrapping.
- **Per-tile surfaces**: Already works. Adding `tile.draw()` and `tile.invalidate()` are new methods that don't affect existing scripts.
- **`drawCenteredLabel()`**: The new multi-line behavior changes the visual output for text containing `\n`, but since that text was already garbled, this is a fix, not a regression. For single-line text, output is identical.

### Migration steps for existing scripts

1. **Scripts using `display.surface()` for per-tile content**: Migrate to `tile.surface()` or `tile.draw()` for better invalidation performance.
2. **Scripts with long text in tiles**: Add `{ wrap: true }` to enable word wrapping.
3. **Scripts with `\n` in text**: No code changes needed — they will automatically render correctly.

### Constants for reference

```go
// runtime/render/visual_runtime.go
const (
    TileWidth         = 90
    TileHeight        = 90
    MainDisplayWidth  = 360
    MainDisplayHeight = 270
    SideDisplayWidth  = 60
    SideDisplayHeight = 270
)

// runtime/ui/ui.go
const (
    MainDisplayColumns = 4
    MainDisplayRows    = 3
)
```

---

## Appendix: Data Flow Diagrams

### Retained tile update (already fast)

```
  anim.loop(1000, callback)
       │
       ▼
  JS: seconds.set(d.getSeconds())
       │
       ▼
  reactive.Signal.Set()
       │
       ▼
  notify dependents
       │
       ▼
  Effect.run() → tile.BindText closure
       │
       ▼
  tile.SetText("12:34")
       │
       ▼
  tile.markDirty()
       │
       ▼
  ui.markDirtyTile(tile)
       │
       ▼
  dirtyTiles[tile] = struct{}{}
       │
       ▼
  onDirty() → present.Invalidate("ui-dirty")
       │
       ▼
  present loop wakes
       │
       ▼
  renderer.Flush()
       │
       ▼
  DirtyTiles() → [tile(0,0)]
       │
       ▼
  renderTile(tile) → 90×90 RGBA
       │
       ▼
  DrawTarget.Draw(image, 0, 0)
       │
       ▼
  Hardware: only 90×90 pixels sent ✓
```

### Display-level surface update (currently slow)

```
  anim.loop(2000, callback)
       │
       ▼
  JS: updateAnimationState()
       │
       ▼
  present.invalidate("loop")
       │
       ▼
  present loop wakes
       │
       ▼
  JS: onFrame(reason) → renderAll()
       │
       ▼
  JS: redraws left, main, right surfaces
       │
       ▼
  renderer.Flush()
       │
       ▼
  DirtyDisplays() → [left, main, right]
       │
       ▼
  renderDisplay(left)  → 60×270 RGBA
  renderDisplay(main)  → 360×270 RGBA  ← 97,200 pixels!
  renderDisplay(right) → 60×270 RGBA
       │
       ▼
  Hardware: 480×270 = 129,600 pixels sent ✗
```

### Per-tile surface update (proposed fast path)

```
  anim.loop(1000, callback)
       │
       ▼
  JS: seconds.set(d.getSeconds())
       │
       ▼
  reactive: Effect.run() → redraw tileSurface
       │
       ▼
  tileSurface.onChange → tile.markDirty()
       │
       ▼
  ui.markDirtyTile(tile)
       │
       ▼
  onDirty() → present.Invalidate("ui-dirty")
       │
       ▼
  present loop wakes
       │
       ▼
  renderer.Flush()
       │
       ▼
  DirtyTiles() → [tile(0,0)]
       │
       ▼
  renderTile(tile) → 90×90 RGBA
       │
       ▼
  DrawTarget.Draw(image, 0, 0)
       │
       ▼
  Hardware: only 90×90 pixels sent ✓
```

---

## Appendix: Key Go Types Reference

```go
// runtime/gfx/text.go
type TextOptions struct {
    X          int
    Y          int
    Width      int
    Height     int
    Brightness uint8
    Face       font.Face
    Center     bool
    // PROPOSED:
    // LineGap    int
    // WrapWidth  int
}

// runtime/ui/tile.go
type Tile struct {
    display *Display
    Col     int
    Row     int
    text       string
    icon       string
    visible    bool
    dirty      bool
    surface    *gfx.Surface
    surfaceSub gfx.Subscription
    textSub    reactive.Subscription
    iconSub    reactive.Subscription
    visibleSub reactive.Subscription
}

// runtime/ui/display.go
type Display struct {
    page       *Page
    Name       string
    text       string
    icon       string
    visible    bool
    dirty      bool
    configured bool
    surface    *gfx.Surface
    surfaceSub gfx.Subscription
    layers     map[string]*DisplayLayer
    layerOrder []string
    tiles      map[TileCoord]*Tile
    // ... reactive subscriptions
}

// runtime/render/visual_runtime.go
type Renderer struct {
    UI      *ui.UI
    Targets map[string]DrawTarget
    Theme   Theme
}

// runtime/gfx/surface.go
type Surface struct {
    width  int
    height int
    pixels []uint8        // grayscale pixel buffer
    mu             sync.Mutex
    cond           *sync.Cond
    batchDepth     int
    changedInBatch bool
    nextListenerID uint64
    listeners      map[uint64]func()
}

// runtime/present/runtime.go
type Runtime struct {
    mu          sync.Mutex
    render      RenderFunc    // func(reason string) error
    flush       FlushFunc     // func() (int, error)
    dirty       bool
    dirtyReason string
    wakeCh      chan struct{}
    closeCh     chan struct{}
    // ...
}

// runtime/reactive/signal.go
type Signal[T any] struct {
    rt    *Runtime
    value T
    equal func(a, b T) bool
    src   sourceNode
}

// runtime/js/env/env.go
type LoupeDeckEnvironment struct {
    Reactive *reactive.Runtime
    UI       *ui.UI
    Host     *host.Runtime
    Anim     *anim.Runtime
    Present  *present.Runtime
    Metrics  *metrics.Collector
}
```

---

*End of document.*
