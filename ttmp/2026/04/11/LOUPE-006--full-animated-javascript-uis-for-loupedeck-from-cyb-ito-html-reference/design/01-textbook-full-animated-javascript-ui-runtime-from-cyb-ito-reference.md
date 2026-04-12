---
Title: 'Textbook: full animated JavaScript UI runtime for Loupedeck from the cyb-ito HTML reference'
Ticket: LOUPE-006
Status: active
Topics:
    - loupedeck
    - go
    - goja
    - javascript
    - animation
    - rendering
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html
      Note: Imported source artifact that defines the reference animated UI to be analyzed and re-expressed on Loupedeck hardware
    - Path: runtime/js/runtime.go
      Note: Current owned goja runtime bootstrap that all new JS-facing animated UI APIs must build on top of
    - Path: runtime/js/module_ui/module.go
      Note: Current retained-page and input-event API that will need to grow beyond simple tiles
    - Path: runtime/js/module_state/module.go
      Note: Current reactive state API that should remain the main JS state primitive for animated interfaces
    - Path: runtime/js/module_anim/module.go
      Note: Current animation module whose loop/tween/timeline model should remain the preferred JS-facing time system
    - Path: runtime/render/visual_runtime.go
      Note: Current retained tile renderer bridge that proves the retained rendering model but is too limited for cyb-ito-style scenes
    - Path: renderer.go
      Note: Current keyed invalidation scheduler that must remain below any new JS-driven animated scene system
    - Path: writer.go
      Note: Current transport ownership and pacing layer that JavaScript must not bypass
    - Path: display.go
      Note: Current region-draw boundary that future retained surfaces or display regions should ultimately target
    - Path: svg_icons.go
      Note: Existing Go-side asset/raster pipeline that informs how future JS-facing visual assets should remain Go-owned
    - Path: cmd/loupe-js-live/main.go
      Note: Current live hardware runner that should eventually run the full animated cyb-ito-inspired demo
ExternalSources:
    - local:cyb-ito.html
Summary: Detailed intern-facing analysis, design, and phased implementation guide for extending the current Loupedeck goja runtime from retained text/icon tiles into full animated JavaScript-driven UI scenes based on the imported cyb-ito HTML canvas reference.
LastUpdated: 2026-04-11T23:31:00-04:00
WhatFor: Teach a new intern what the cyb-ito reference actually is, why the current JS runtime cannot reproduce it directly, and how to implement a correct full animated UI system without violating Go-owned rendering and transport control.
WhenToUse: Use when planning or implementing the next JS runtime expansion beyond simple retained tiles, especially for animated raster scenes, side strips, overlays, and touch-driven visual effects.
---

# Textbook: full animated JavaScript UI runtime for Loupedeck from the cyb-ito HTML reference

## Executive Summary

This document is the detailed guide for the next major step in the Loupedeck JavaScript runtime: moving from a small retained text/icon tile model to a **full animated JavaScript UI system** capable of expressing dense procedural visuals like the imported `cyb-ito.html` reference.

The first and most important thing a new intern must understand is that `cyb-ito.html` is **not** a normal webpage UI and **not** a DOM-based design that can be translated into buttons, labels, and CSS transitions. It is a single-file procedural canvas renderer that:

- draws its own grayscale framebuffer pixel by pixel,
- maintains a `4×3` grid of `90×90` animated tiles,
- draws animated left and right side strips,
- overlays spiral ripple effects,
- drives state from a render loop and touch interaction,
- and composes everything into one synthetic display image.

That means the current JS API in this repository is **structurally on the right path** but **not yet sufficient** for a faithful implementation. The current runtime already has:

- owned goja execution,
- reactive signals,
- retained pages,
- button/touch/knob callbacks,
- numeric animation loops and tweens,
- and a hardware live runner.

However, it still lacks the next layer that `cyb-ito.html` requires:

1. **display regions beyond the main tile grid**
2. **Go-owned retained raster surfaces or scene layers**
3. **JS-facing graphics primitives above those retained surfaces**
4. **overlay composition**
5. **side-strip rendering support**

The key thesis of this document is:

> **We should not port the HTML file by exposing raw pixels or raw transport to JavaScript. We should port it by extending the retained JS runtime with Go-owned display regions, surfaces, layers, and graphics primitives, while keeping all rendering policy, pacing, and transport ownership in Go.**

That preserves the architectural wins from `LOUPE-003` and `LOUPE-005` instead of undoing them.

## Problem Statement

The repository now has a real goja-based Loupedeck runtime, but it was designed and validated against a deliberately small first slice of UI problems:

- static retained pages,
- tile labels,
- button-driven counters,
- knob-driven numeric state,
- touch feedback,
- page switching,
- and simple loop-driven animation.

That first slice was the correct scope for proving:

- owner-thread callback safety,
- reactive semantics,
- retained page switching,
- hardware event integration,
- and live-runner viability.

But `cyb-ito.html` is qualitatively different.

It is not asking for:

- one tile whose text changes,
- one value that tweens from `0` to `100`,
- or one page switch on button press.

It is asking for:

- procedural raster art,
- per-frame scene redraw,
- multiple animation layers,
- side-strip activity outside the main `4×3` tile grid,
- touch-triggered ripple overlays,
- and a composition model that is closer to a retained animated scene than to a dashboard of labels.

The problem is therefore not “how do we write a cool JS demo?”

The real problem is:

> **How do we extend the current safe retained JS runtime into a full animated scene runtime without breaking the current Go-owned writer/renderer/transport ownership boundary?**

That is the architecture question this ticket must answer.

## Who this document is for

This document is written for a new intern who:

- can read Go and JavaScript,
- has seen the existing Loupedeck runtime packages,
- understands basic animation concepts,
- but does not yet know how to design or implement a retained animated UI runtime for constrained hardware.

The intern should come away understanding:

- what the imported HTML actually does,
- what pieces of the current runtime are reusable,
- what pieces are missing,
- what new APIs should exist,
- what should remain in Go,
- and what implementation order is safest.

## The current baseline architecture

Before discussing the cyb-ito reference, the intern must understand the current stack that already exists in this repository.

```mermaid
flowchart TD
    A[JS script] --> B[owned goja runtime]
    B --> C[state/ui/anim/easing modules]
    C --> D[reactive runtime]
    C --> E[retained pages + tiles]
    C --> F[host runtime]
    C --> G[animation runtime]
    E --> H[retained tile renderer]
    H --> I[Display.Draw(image, x, y)]
    I --> J[keyed invalidation scheduler]
    J --> K[single outbound writer]
    K --> L[serial websocket transport]
    L --> M[Loupedeck Live]
```

Important current files:

- `runtime/js/runtime.go`
  - owned runtime bootstrap
- `runtime/js/module_state/module.go`
  - `signal`, `computed`, `batch`, `watch`
- `runtime/js/module_ui/module.go`
  - `page`, `show`, `onButton`, `onTouch`, `onKnob`, tile bindings
- `runtime/js/module_anim/module.go`
  - `to`, `loop`, `timeline`
- `runtime/render/visual_runtime.go`
  - retained tile renderer bridge
- `renderer.go`
  - keyed invalidation scheduler
- `writer.go`
  - paced single-writer transport owner
- `display.go`
  - region draw boundary

That architecture is already correct in one crucial way:

> **JavaScript owns UI semantics, not transport bytes.**

This must remain true after the cyb-ito expansion.

## What `cyb-ito.html` actually is

The imported file lives at:

- `ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html`

The first thing to understand is that it is a **single-file canvas scene program**, not a componentized browser app.

### High-level structure

It contains:

- one `<canvas id="c">`,
- one inline `<script>`,
- no external assets,
- no frameworks,
- no SVG assets,
- no CSS transitions,
- one `requestAnimationFrame(render)` loop.

### Geometry model

At the top of the script it defines:

```javascript
const BS=90,COLS=4,ROWS=3,STRIP=36;
const TW=STRIP+COLS*BS+STRIP,TH=ROWS*BS;
```

This means the virtual scene is:

- left strip: `36×270`
- main grid: `360×270`
- right strip: `36×270`
- total canvas: `432×270`

The `90×90` tile size is significant because it already matches the Loupedeck Live main display’s `4×3` tile layout perfectly.

The strip width does **not** match real hardware, because the real device gives us:

- left display: `60×270`
- main display: `360×270`
- right display: `60×270`

So the reference geometry is a strong conceptual match, but not a one-to-one physical match.

### Rendering model

The file draws into raw pixel memory:

```javascript
const img=ctx.createImageData(TW,TH);
const d=img.data;
```

and uses helper functions like:

- `setP(x,y,v)`
- `lineH(...)`
- `lineV(...)`
- `rTxt(...)`
- `drawText(...)`
- `crosshatch(...)`
- `drawSpiral(...)`
- `drip(...)`

This is a grayscale additive raster pipeline, not a DOM paint tree.

### Tile set

The file defines twelve animated tiles under `const tiles=[ ... ]`:

1. `眼 / EYE`
2. `渦 / SPIRAL`
3. `歯 / TEETH`
4. `溶 / MELT`
5. `穴 / HOLE`
6. `狂 / FACE`
7. `蟲 / WORM`
8. `砂 / NOISE`
9. `歪 / WARP`
10. `裂 / CRACK`
11. `脈 / PULSE`
12. `闇 / VOID`

Each tile object includes a `draw(ox, oy, t, active)` function that draws its own contents procedurally.

### Global effects outside the tiles

The source is not only a grid of twelve independent mini-scenes. It also includes:

- expanding **spiral ripple** overlays
- animated **left strip dripping bars**
- animated **right strip scrolling horror kanji**
- global **scanline darkening**
- per-tile flash and scan-sweep effects

This matters because it means the correct target architecture is **not** only “better tiles.” The correct target is a multi-region retained scene runtime.

### Input model

The source handles:

- pointer/touch down
- pointer/touch move
- pointer/touch up

On down it:

- creates ripple spirals,
- activates the touched tile,
- sets flash,
- starts a tile-local scan sweep.

On move it:

- creates more ripples.

On up it:

- clears the tile `active` states.

So the interaction model is closer to a touch-reactive scene system than to a simple UI form.

## Why the current JS runtime is not enough yet

The current JS runtime is already good enough for:

- text changes,
- visibility toggles,
- simple named-page switching,
- event-driven state changes,
- numeric signal animation.

But a faithful cyb-ito implementation needs more.

### Things we already have that are useful

We should preserve and reuse:

- `state.signal(...)`
- `state.computed(...)`
- `state.batch(...)`
- `anim.loop(...)`
- `anim.timeline(...)`
- `ui.onTouch(...)`
- `ui.onButton(...)`
- owner-thread callback serialization
- Go-owned writer and invalidation scheduler

### Things we do not yet have

We currently lack:

- JS-facing left/right display regions
- retained display-sized surfaces
- overlay/layer composition
- Go-owned raster graphics primitives exposed to JS
- region-wide or display-wide post-processing effects
- JS-side scene descriptions richer than tile text/icon/visible

### Why not just expose `setPixel(...)` to JS?

A new intern may reasonably ask: why not add a JS API like this?

```javascript
surface.setPixel(x, y, brightness)
```

The answer is: that would be the wrong ownership boundary.

Problems with a naïve JS per-pixel API:

- enormous Go↔goja call overhead
- hard to optimize or batch
- encourages JS to think in transport-era draw loops
- difficult to coalesce at the scene level
- too easy to accidentally push the runtime toward “JS owns the framebuffer”

The right design is to give JS **coarse scene and graphics operations**, not raw low-level pixel loops.

## The target architecture

The cyb-ito ticket should extend the current runtime like this:

```mermaid
flowchart TD
    A[JS state + event handlers] --> B[JS scene declarations]
    B --> C[gfx module: Go-owned retained surfaces]
    B --> D[ui module: pages + display regions + layers]
    A --> E[anim module]
    E --> A
    C --> F[render scene to display images]
    D --> F
    F --> G[Display.Draw(image, x, y)]
    G --> H[renderer.go invalidation scheduler]
    H --> I[writer.go single writer]
    I --> J[Loupedeck Live]
```

The key rule is still:

> **Go owns realization. JS owns state, structure, and animation intent.**

## New concepts we need

The next runtime slice should introduce four new concepts.

### 1. Display regions

Today JS can create only pages with tiles in the main `4×3` grid.

For cyb-ito, JS needs named display regions:

- `main`
- `left`
- `right`

Proposed API sketch:

```javascript
const ui = require("loupedeck/ui");

ui.page("cyb", page => {
  page.display("left", display => {
    // left strip content
  });

  page.display("main", display => {
    // main scene content
  });

  page.display("right", display => {
    // right strip content
  });
});
```

### 2. Retained surfaces

JS needs a way to describe and update display-sized graphics content without owning transport.

Proposed API sketch:

```javascript
const gfx = require("loupedeck/gfx");

const main = gfx.surface(360, 270, { mode: "mono-additive" });
const left = gfx.surface(60, 270, { mode: "mono-additive" });
const right = gfx.surface(60, 270, { mode: "mono-additive" });
```

These surfaces should be Go-owned image buffers or logical drawing command buffers, not JS-owned pixel arrays.

### 3. Layers / overlays

The source has overlay behavior:

- touch ripples on top of tiles
- scanline darkening across the scene
- strip pips mirroring tile activity

That means we need a compositional model.

Proposed concept:

```javascript
page.display("main", display => {
  display.layer("base", () => mainSurface);
  display.layer("ripple", () => rippleSurface);
  display.layer("scanlines", () => scanSurface);
});
```

Or a simpler retained model where one display owns several named surfaces.

### 4. Graphics primitives above raw pixels

The current source uses drawing building blocks that are more meaningful than pixels:

- line
- text
- spiral
- drip
- crosshatch
- noise field
- scanline effect

So the future `gfx` API should expose meaningful coarse operations.

Example sketch:

```javascript
surface.clear(0);
surface.text("EYE", { x: 4, y: 3, font: "serif", size: 9, brightness: 160 });
surface.line({ x1: 2, y1: 13, x2: 87, y2: 13, brightness: 30 });
surface.spiral({ x: 45, y: 46, turns: 6, radius: 32, brightness: 180 });
surface.crosshatch({ x: 4, y: 16, w: 18, h: 50, density: 3, brightness: 20 });
```

This gives JS authoring power without JS becoming the raster engine.

## How the imported reference maps to the target runtime

The intern should think of the imported HTML not as something to “run,” but as something to **translate into the new retained animated scene model**.

### Mapping table

| Source concept | Meaning in the reference | Correct runtime target |
|---|---|---|
| `BS=90` tile size | one animated horror tile | main-grid retained region or sub-surface |
| `STRIP=36` | visual side bars in browser mockup | left/right display regions adapted to real `60×270` hardware |
| `img.data` direct raster writes | procedural scene output | Go-owned retained surfaces / draw command realization |
| `draw(ox,oy,t,active)` per tile | tile-local procedural art | named tile renderer or JS scene function that populates a Go surface |
| `spirals[]` | touch ripple overlay state | reactive overlay state list |
| `flash`, `active`, `scanning` | per-tile runtime state | JS reactive state bound into retained scene layers |
| `requestAnimationFrame(render)` | global time progression | host-owned `anim.loop(...)` and retained redraw invalidation |
| `scrollOff` | strip scroll state | signal-driven strip animation |
| `onDown`, `onMove`, `onUp` | touch interaction | `ui.onTouch(...)` plus possibly later gesture helpers |

## Why the current `anim` and `state` APIs are still correct

This ticket does **not** mean the existing JS APIs were wrong. In fact, the current `state` and `anim` APIs are precisely the pieces we should keep.

The source scene has lots of dynamic values:

- current frame/time
- strip scroll position
- ripple ages
- tile flash strengths
- tile active flags
- scan sweep positions
- internal tile-specific animation state

That should still be modeled with:

- signals
- computed values
- loop-driven updates
- occasionally timelines or tweens

The thing that must change is **how those values become visuals**, not how they become state.

## Recommended API direction

The next JS-facing modules should likely look like this:

### Existing modules kept

- `require("loupedeck/state")`
- `require("loupedeck/ui")`
- `require("loupedeck/anim")`
- `require("loupedeck/easing")`

### New module added

- `require("loupedeck/gfx")`

### Possible additions inside `ui`

- `page.display(name, fn)`
- `display.layer(name, fn)` or `display.surface(...)`
- maybe `page.tile(col, row, fn)` retained as a convenience abstraction on top of `main`

### Possible additions inside `gfx`

- `surface(width, height, options)`
- `clear(value)`
- `text(...)`
- `line(...)`
- `rect(...)`
- `crosshatch(...)`
- `spiral(...)`
- `noise(...)`
- `drip(...)`
- `scanlines(...)`
- `composite(...)`

The exact surface API can evolve, but the module boundary itself is the important decision.

## Pseudocode for the intended scene model

A future cyb-ito-style script should feel more like this:

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");

const ripples = state.signal([]);
const scroll = state.signal(0);
const tiles = state.signal(initialTileState());

const main = gfx.surface(360, 270, { mode: "mono-additive" });
const left = gfx.surface(60, 270, { mode: "mono-additive" });
const right = gfx.surface(60, 270, { mode: "mono-additive" });
const overlay = gfx.surface(360, 270, { mode: "mono-additive" });

ui.page("cyb", page => {
  page.display("left", display => {
    display.surface(() => renderLeftStrip(left, scroll.get(), tiles.get()));
  });

  page.display("main", display => {
    display.surface(() => renderMainScene(main, tiles.get()));
    display.overlay(() => renderRipples(overlay, ripples.get()));
  });

  page.display("right", display => {
    display.surface(() => renderRightStrip(right, scroll.get(), tiles.get()));
  });
});

ui.onTouch("Touch1", ev => spawnRipple(ev.x, ev.y));
// ... more touch handlers or a generic region mapper later

anim.loop(16, t => {
  advanceTileAnimations(tiles);
  advanceRipples(ripples);
  scroll.set(nextScroll(scroll.get()));
});

ui.show("cyb");
```

The actual final syntax may vary, but this is the right model:

- JS mutates scene state
- JS invokes coarse Go-owned graphics operations
- Go owns actual surfaces, invalidation, and transport

## Why we should not literally run the HTML

A common wrong instinct would be to try to embed the HTML file or execute it directly in some browser-like environment.

Reasons not to do that:

- the target hardware runtime is not a browser
- the HTML canvas API is not the same abstraction boundary we want for Loupedeck JS
- browser `requestAnimationFrame` semantics are not the same as our host-owned loop model
- we need hardware-aware display-region integration, not a fake browser canvas on top of a serial device

The imported file is a **reference artifact**, not the runtime target format.

## Detailed implementation plan

The safest implementation sequence is incremental.

### Phase A: source analysis and target design (this ticket’s current stage)

Deliverables:

- imported HTML tracked under the ticket
- detailed design doc (this document)
- explicit API direction
- mapping from source artifact to runtime features

Acceptance criteria:

- everyone agrees the target is a retained multi-region animated scene runtime, not raw browser canvas emulation

### Phase B: extend the retained UI model to support display regions

Files likely touched:

- `runtime/ui/ui.go`
- `runtime/ui/page.go`
- possibly new files in `runtime/ui/`
- `runtime/js/module_ui/module.go`
- `runtime/render/visual_runtime.go`

Goals:

- add retained display regions for `left`, `main`, `right`
- preserve existing tile API on top of `main`
- keep current tile examples working

Acceptance criteria:

- JS can declare region-level content, not only `main` tiles
- live runner can flush retained left/right scenes too

### Phase C: add a pure-Go graphics/surface package

Likely new package:

- `runtime/gfx/`

Goals:

- provide Go-owned grayscale/monochrome or additive-retained surfaces
- expose coarse drawing ops
- support efficient clearing and composition
- keep enough control to optimize later

Acceptance criteria:

- Go tests prove surfaces and drawing ops work without any goja dependency

### Phase D: add `runtime/js/module_gfx`

Likely files:

- `runtime/js/module_gfx/module.go`
- `runtime/js/runtime.go`

Goals:

- expose surfaces and coarse drawing operations to JS
- preserve owner-thread safety where JS closures are involved
- avoid per-pixel Go↔JS loops as the default authoring path

Acceptance criteria:

- JS can build and update surfaces that are then consumed by retained display regions

### Phase E: add layered composition

Possible files:

- `runtime/render/...`
- `runtime/ui/...`
- `runtime/js/module_ui/module.go`

Goals:

- multiple retained layers per display
- overlay support for ripples, scanlines, highlights
- stable compositing order

Acceptance criteria:

- a simple base + overlay demo works on the main display

### Phase F: build the first cyb-ito-inspired main-scene demo

Likely output:

- `examples/js/07-cyb-ito.js`
- optionally a dedicated live runner example command or helper

Goals:

- reproduce a subset first:
  - 12 animated tiles
  - touch flash/scan
  - ripple overlay
- do not yet require full strip parity in the very first visual milestone

Acceptance criteria:

- main display demo runs on real hardware and is visibly recognizably inspired by the source

### Phase G: add left/right strip scenes

Goals:

- left dripping bars
- right scrolling kanji
- activity pips tied to tile flash state

Acceptance criteria:

- the full visual composition spans left + main + right displays on hardware

### Phase H: optimize and harden

Goals:

- benchmark throughput under dense animation
- decide whether per-display flush cadences or additional coalescing are needed
- refine graphics operations if some are too slow or too allocation-heavy

Acceptance criteria:

- stable interactive demo with acceptable device behavior and no JS ownership regressions

## Concrete file roadmap for the intern

A new intern should expect the work to start in these areas:

### 1. JS/UI surface area

- `runtime/js/module_ui/module.go`
- new `runtime/js/module_gfx/module.go`

### 2. Pure-Go UI/runtime layers

- `runtime/ui/`
- `runtime/render/`
- new `runtime/gfx/`

### 3. Live integration

- `cmd/loupe-js-live/main.go`
- `runtime/js/runtime.go`

### 4. Demo scripts

- `examples/js/`

### 5. Ticket docs and validation evidence

- `ttmp/2026/04/11/LOUPE-006--.../`

## Common failure modes to avoid

### Failure mode 1: exposing raw transport to JS

Bad idea:

```javascript
deck.drawRaw(...)
deck.sendFramebuffer(...)
```

Why it is bad:

- bypasses writer ownership
- bypasses region scheduling
- recreates the old architecture mistakes

### Failure mode 2: building a JS per-pixel inner loop

Bad idea:

```javascript
for (let y = 0; y < 270; y++) {
  for (let x = 0; x < 360; x++) {
    surface.setPixel(x, y, value(x, y));
  }
}
```

Why it is bad:

- too much goja crossing overhead
- hard to optimize globally
- too easy to make performance depend on JS interpreter overhead

### Failure mode 3: abandoning retained scene state

Bad idea:

- recomputing everything as ad hoc immediate draws with no retained model

Why it is bad:

- harder to layer overlays
- harder to reuse current invalidation scheduler
- harder to recover after reconnect

### Failure mode 4: trying to emulate the browser exactly

Bad idea:

- building a tiny HTML canvas clone instead of a Loupedeck-native scene system

Why it is bad:

- wrong target abstraction
- misses the hardware-specific geometry and transport model

## Working rules

> [!important]
> Treat `cyb-ito.html` as a **reference scene specification**, not an executable format we are trying to host directly.

> [!important]
> Keep JavaScript above the rendering/transport ownership boundary. JS can describe state and scene intent; Go still owns drawing realization, region invalidation, writer pacing, and transport safety.

> [!important]
> Prefer pure-Go packages for semantics first (`runtime/gfx`, display-region retention, layer composition), then bind them into goja second.

> [!important]
> Preserve the current simple tile API as a convenience layer. The new scene system should grow the runtime, not replace every existing example and test immediately.

> [!important]
> Always adapt the imported visual design to **real hardware geometry** rather than freezing browser mock geometry such as `STRIP=36` when the device actually provides `60×270` side displays.

## Suggested API reference draft for the next slice

The intern should think in terms of these likely additions.

### `loupedeck/ui`

Possible new exports:

- `page.display(name, fn)`
- maybe `display.surface(fn)`
- maybe `display.layer(name, fn)`

### `loupedeck/gfx`

Possible new exports:

- `surface(width, height, options)`
- `surface.clear(value)`
- `surface.text(text, options)`
- `surface.line(options)`
- `surface.crosshatch(options)`
- `surface.spiral(options)`
- `surface.noise(options)`
- `surface.drip(options)`
- `surface.composite(other, options)`

### Existing modules retained

- `loupedeck/state`
- `loupedeck/anim`
- `loupedeck/easing`

## Review checklist for a future implementation PR

Before approving an implementation PR for this ticket, ask:

- Does JS still avoid raw transport ownership?
- Are the new scene/surface semantics implemented in pure Go first?
- Are left/right display regions part of the retained model, not hacks in the runner?
- Are overlays and composition explicit and testable?
- Can the existing simple JS examples still run?
- Is the cyb-ito-inspired demo recognizably faithful to the source without trying to literally emulate browser APIs?
- Are hardware validation steps documented with exact commands and observed behavior?

## Related files and next reading order

A new intern should read in this order:

1. `ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html`
2. `runtime/js/runtime.go`
3. `runtime/js/module_ui/module.go`
4. `runtime/js/module_state/module.go`
5. `runtime/js/module_anim/module.go`
6. `runtime/render/visual_runtime.go`
7. `renderer.go`
8. `writer.go`
9. this document again

## Final recommendation

The correct next move is **not** to make JavaScript stronger by giving it lower-level access. The correct next move is to make JavaScript stronger by giving it **better retained animated scene abstractions** that are still realized in Go.

That is how this project can support full animated horror-style UIs like `cyb-ito.html` while still respecting the hard-won transport and rendering discipline already established in the repository.
