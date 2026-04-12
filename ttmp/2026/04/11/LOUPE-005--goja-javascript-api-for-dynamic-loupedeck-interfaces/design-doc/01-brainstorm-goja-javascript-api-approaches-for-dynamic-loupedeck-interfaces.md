---
Title: 'Brainstorm: goja JavaScript API approaches for dynamic Loupedeck interfaces'
Ticket: LOUPE-005
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
    - Path: display.go
      Note: Current image-to-blit entry point that should likely remain below the JS API boundary
    - Path: renderer.go
      Note: Current render scheduler whose semantics strongly shape any future dynamic JS runtime
    - Path: writer.go
      Note: Transport ownership/pacing layer that scripts should not bypass
    - Path: cmd/loupe-svg-buttons/main.go
      Note: Current dynamic button-bank demo that hints at the kinds of interfaces scripts should be able to express
    - Path: svg_icons.go
      Note: Existing icon asset pipeline that might become script-visible as an icon/image registry
ExternalSources: []
Summary: Deep technical brainstorm covering multiple possible goja JavaScript API designs for dynamic Loupedeck interfaces, including animation models, easing/timeline ideas, scene/runtime lifecycles, and recommendation tradeoffs.
LastUpdated: 2026-04-11T20:40:45-04:00
WhatFor: Explore how to layer an elegant JavaScript runtime over the current Go renderer/writer architecture without repeating the earlier mistake of exposing raw transport policy at the top level.
WhenToUse: Use when deciding the API shape for a future goja-based dynamic UI system or when comparing scripting approaches and animation models.
---

# Brainstorm: goja JavaScript API approaches for dynamic Loupedeck interfaces

## Executive Summary

The current Go codebase already has the two hardest low-level ingredients required for a scriptable dynamic UI runtime: a package-owned transport writer and a render scheduler that can coalesce repeated updates before they hit the device. That means the next major design question is not “how do we talk to the hardware from JavaScript?” but “what should JavaScript be allowed to describe?”

This document explores multiple possible API shapes for a goja-based runtime embedded inside the Go process. The best answer is probably not a purely low-level imperative API and not a fully magical React-like retained scene system on day one. The strongest direction appears to be a **hybrid** model:

1. a small low-level runtime module for events, pages, assets, and lifecycle,
2. a retained screen/tile model for ordinary UI state,
3. an animation/timeline module with easing curves,
4. a narrow escape hatch for imperative effects,
5. all of it compiled down to the current Go-side renderer/writer stack.

The reason this hybrid looks strongest is simple: the Loupedeck device is dynamic enough that scripts need expressive power, but constrained enough that the runtime must still control scheduling, coalescing, pacing, and state recovery after reconnects.

## Problem Statement

If the project adds a goja VM so JavaScript can build dynamic Loupedeck interfaces, several naive approaches are available—and most of them are wrong in interesting ways.

Wrong idea #1 is to expose raw transport functions to JavaScript:

```javascript
deck.sendRawFramebuffer(...)
deck.sendDraw(...)
```

That would simply push the previous architectural problems up one layer. Scripts would accidentally recreate redraw storms, transport coupling, and fragile sequencing.

Wrong idea #2 is to jump directly to a giant retained declarative framework without understanding what the device actually needs. That would likely create a lot of code before the team validates the right state model, animation model, and recovery semantics.

The real problem is therefore to find an API boundary that is:

- expressive enough for dynamic interfaces,
- constrained enough to preserve transport safety,
- elegant enough that scripts read like UI code rather than device-driver glue,
- capable of describing animation with easing curves and timelines,
- friendly to future asset pipelines such as SVG icon loading,
- compatible with the current renderer/writer architecture.

## Constraints from the current system

Any future JS API must respect the current Go-side realities.

### Constraint 1: transport must remain package-owned

The current writer in `writer.go` centralizes websocket writes and pacing. This should remain true. Scripts should never directly decide how low-level framebuffer and draw messages are timed.

### Constraint 2: rendering is region-based

The current system naturally supports rectangular subregion updates (`90×90` tiles, knob strips, etc.). A future script API should use that shape rather than pretending it has a free-form GPU canvas.

### Constraint 3: dynamic interfaces will need retained state

As soon as multiple script callbacks, animations, and pages exist, a purely immediate-mode “draw this image now” model will become awkward. The Go side likely needs some retained UI state beneath the scripting surface.

### Constraint 4: reconnects exist

The transport still has lifecycle quirks. A script API that assumes transport continuity forever is too optimistic. The runtime should eventually be able to re-render state after reconnect.

### Constraint 5: goja favors explicit module contracts

A future module surface should look like `require("loupedeck")`, `require("loupedeck/ui")`, `require("loupedeck/anim")`, etc., with lowerCamelCase keys and explicit thrown errors or return values.

## Design goals

A good JS API for this project should:

- let scripts describe **pages**, **tiles**, **widgets**, and **animations**
- make ordinary code concise for common scenarios
- expose events from buttons, touches, and knobs cleanly
- provide animation primitives like:
  - duration
  - delay
  - repeat
  - yoyo
  - easing curves
  - timelines/sequences
- preserve Go-side control over:
  - rendering
  - invalidation
  - scheduling
  - transport pacing
  - reconnect recovery
- support both static assets (icons, text) and computed visuals
- allow multiple styles of programming without exploding complexity

## Approach A: low-level imperative scripting API

### Shape

```javascript
const deck = require("loupedeck")

deck.connect()
deck.onTouch("Touch1", () => {
  deck.tile(0, 0).setIcon("finder")
  deck.tile(0, 0).setScale(1.1)
  deck.flush()
})
```

### Mental model

Scripts call methods directly in response to events. The runtime exposes device objects and mutation methods.

### Strengths

- easy to explain initially
- small API surface to start
- maps well to current demos
- low implementation cost

### Weaknesses

- state quickly becomes scattered across callbacks
- animation orchestration becomes manual and repetitive
- page switching and recovery become difficult
- easy to accidentally write imperative “do everything now” spaghetti

### Best fit

- diagnostics
- operator macros
- quick prototypes
- tiny one-off control scripts

### Verdict

Good as an escape hatch or base layer, but not elegant enough as the primary model.

## Approach B: retained declarative scene/page API

### Shape

```javascript
const ui = require("loupedeck/ui")

ui.page("home", page => {
  page.tile(0, 0, t => t.icon("finder"))
  page.tile(1, 0, t => t.icon("trash"))
  page.tile(2, 0, t => t.text("REC").blink())
  page.tile(3, 0, t => t.icon("clock").animate("pulse"))
})

ui.show("home")
```

### Mental model

JavaScript declares what the UI is, not how low-level draws should happen. The Go side owns retained scene state and diffs/re-renders as needed.

### Strengths

- clean mental model
- easy to restore after reconnect
- good match for page-based device UIs
- natural fit for retained renderer evolution
- encourages stability and composability

### Weaknesses

- larger implementation investment
- may feel heavy for tiny scripts
- requires deciding how mutable/dynamic declarations become at runtime

### Best fit

- dashboards
- button banks
- page-based workflows
- long-running interactive control surfaces

### Verdict

Very attractive as the main UI model if implemented gradually.

## Approach C: reactive signal/store API

### Shape

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")

const level = state.signal(0)

ui.page("meter", page => {
  page.tile(0, 0, t => t.text(() => String(level.get())))
  page.tile(1, 0, t => t.progress(() => level.get() / 100))
})

ui.onKnob("Knob1", delta => {
  level.set(Math.max(0, Math.min(100, level.get() + delta)))
})
```

### Mental model

Scripts mutate signals; the UI depends on those signals; the runtime schedules the resulting updates.

### Strengths

- elegant for dynamic controls
- natural fit for knob/touch-driven UIs
- makes animation state and UI state explicit
- composes well with a retained renderer

### Weaknesses

- requires more runtime machinery
- needs careful invalidation semantics
- can become magical if dependency tracking is opaque

### Best fit

- live control dashboards
- input-bound visualizations
- stateful page UIs

### Verdict

Likely excellent as part of a hybrid design, especially if the dependency model stays simple and explicit.

## Approach D: timeline-centric animation API

### Shape

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")

const tile = ui.page("home").tile(0, 0)
tile.icon("finder")

anim.timeline()
  .to(tile, { scale: 1.15, y: -4 }, 180, "outBack")
  .to(tile, { scale: 1.0, y: 0 }, 220, "inOutCubic")
  .play()
```

### Mental model

Animation is a first-class system, with keyframes/tweens/timelines operating against retained node properties.

### Strengths

- very expressive for easing-based interfaces
- compact for sequenced motion
- good match for rich interactive polish
- easy to author multiple animation styles

### Weaknesses

- needs a stable property model underneath
- too narrow to be the whole API
- invites complexity if not grounded in a retained scene

### Best fit

- transitions
- button press feedback
- page changes
- continuous ambient motion

### Verdict

Not enough by itself, but extremely desirable as a companion module.

## Approach E: hybrid declarative pages + reactive state + imperative effects + timelines

### Shape

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")
const anim = require("loupedeck/anim")

const page = ui.page("home")
const armed = state.signal(false)

const recTile = page.tile(0, 0, t => {
  t.text(() => armed.get() ? "REC" : "IDLE")
  t.icon(() => armed.get() ? "record" : "pause")
})

ui.onTouch("Touch1", () => {
  armed.set(!armed.get())
  anim.timeline()
    .to(recTile, { scale: 1.18 }, 120, "outBack")
    .to(recTile, { scale: 1.0 }, 180, "inOutQuad")
    .play()
})

ui.show("home")
```

### Why this looks strongest

It gives:

- declarative page layout for ordinary UI
- explicit state for dynamic logic
- timelines/easing for rich animation
- imperative escape hatches for special effects
- a structure that still compiles into retained rendering underneath

### Verdict

This is the most elegant long-term direction.

## Recommended module layout brainstorm

A plausible module family:

### `require("loupedeck")`

Low-level runtime / device module:

- `onButton(name, fn)`
- `onTouch(name, fn)`
- `onKnob(name, fn)`
- `show(pageName)`
- `currentPage()`
- `log(...)`
- `setTimeout(fn, ms)` / `clearTimeout(id)`
- `setInterval(fn, ms)` / `clearInterval(id)`
- `requestAnimationFrame(fn)` or runtime frame subscription

### `require("loupedeck/ui")`

Retained UI description:

- `page(name, builder)`
- `show(name)`
- `tile(col, row, builder?)`
- `strip(side, builder?)`
- `overlay(name, builder?)`
- widget builders such as:
  - `icon(name)`
  - `text(valueOrFn)`
  - `progress(valueOrFn)`
  - `gauge(valueOrFn)`
  - `visible(valueOrFn)`
  - `style({...})`

### `require("loupedeck/state")`

Simple explicit state model:

- `signal(initial)`
- `computed(fn)`
- `watch(signal, fn)`
- maybe `store(object)` later, but signals are enough initially

### `require("loupedeck/anim")`

Animation/tween system:

- `to(target, props, duration, easing)`
- `fromTo(target, from, to, duration, easing)`
- `timeline()`
- `sequence([...])`
- `parallel([...])`
- `cancel(target)`
- `spring(target, props, options)` maybe later

### `require("loupedeck/easing")`

Easing utilities/constants:

- `linear`
- `inQuad`
- `outQuad`
- `inOutQuad`
- `inCubic`
- `outCubic`
- `inOutCubic`
- `outBack`
- `outElastic`
- `steps(n)`
- maybe `bezier(x1, y1, x2, y2)`

### `require("loupedeck/assets")`

Asset registry:

- `icon(name)`
- `loadIconLibrary(path)`
- `image(path)`
- `textRaster(...)` maybe if needed

## Runtime architecture ideas

### Option 1: host-driven fixed timestep

Go owns a fixed animation/render clock, for example 30Hz or 60Hz internally, and JS callbacks run on that clock.

Pros:
- deterministic
- easy to reason about
- aligns with retained render scheduler

Cons:
- more rigid
- long JS frames can block the host loop

### Option 2: event loop + RAF-like subscription

Expose `requestAnimationFrame` semantics to JS and let scripts animate with frame callbacks.

Pros:
- familiar to JS authors
- elegant for animation/tween engines

Cons:
- needs careful integration with Go-side scheduling
- easy to overuse if not paired with retained state

### Option 3: no free-form per-frame callback, only host animations

JS declares timelines/tweens; Go executes them without arbitrary per-frame user code.

Pros:
- safer
- easiest to optimize and coalesce
- strongest separation of intent vs transport

Cons:
- less expressive for procedural visuals

### Recommendation

Use a hybrid:

- ordinary animation through host tween/timeline primitives
- optional RAF-style callback for advanced/procedural cases
- strong guidance that retained nodes and host animations are the default

## State ownership brainstorm

### JS-owned state, Go renders from JS queries

JS is the source of truth. Go asks JS what the UI looks like.

Bad fit:
- hard to recover after reconnect
- too many crossings
- awkward for retained rendering

### Go-owned retained scene, JS mutates via API

JS declares pages/nodes/signals, but Go owns the actual retained scene graph / framebuffer.

Better fit:
- reconnect-friendly
- renderer can diff and coalesce
- transport remains Go-owned

### Hybrid state

- JS owns business/application state
- Go owns retained visual state derived from it

Best fit.

## Easing curve design brainstorm

Animations are a first-class user request, so easing deserves explicit design.

### API style 1: string names

```javascript
anim.to(tile, { scale: 1.1 }, 180, "outBack")
```

Pros:
- concise
- readable

Cons:
- typo-prone

### API style 2: easing module symbols

```javascript
const easing = require("loupedeck/easing")
anim.to(tile, { scale: 1.1 }, 180, easing.outBack)
```

Pros:
- discoverable
- composable
- can expose factories like `easing.steps(4)`

Cons:
- slightly more verbose

### API style 3: CSS-like strings

```javascript
anim.to(tile, { x: 16 }, 240, "cubic-bezier(0.34, 1.56, 0.64, 1)")
```

Pros:
- familiar to frontend authors

Cons:
- parsing burden
- less Goja-native/simple

### Recommendation

Support both named built-ins and `easing.bezier(...)` / `easing.steps(n)` helpers. Avoid a CSS parser initially.

## Animation scenario brainstorm

### Scenario 1: button press pop

- touch down
- scale tile to `1.14`
- slightly lift `y`
- return with `outBack` then `inOutCubic`

### Scenario 2: armed/recording pulse

- animate scale and brightness on loop
- use `inOutSine` or `inOutQuad`
- maybe color alternation if the asset model supports it

### Scenario 3: page transition slide

- current page tiles slide left
- next page tiles slide in from right
- stagger by column

### Scenario 4: knob-controlled easing scrub

- knob changes timeline progress
- tile movement reflects exact eased state

### Scenario 5: icon browser bank carousel

- banks auto-cycle every N seconds
- manual prev/next interrupts timeline
- small ease on incoming/outgoing banks

### Scenario 6: physics-ish spring effect

- touch starts a spring settle on scale/rotation
- likely better as a host primitive than userland physics per frame initially

## Multiple scenario-specific API ideas

### Scenario: simple tile bank

#### Declarative

```javascript
const ui = require("loupedeck/ui")

ui.page("icons", page => {
  page.tile(0, 0, t => t.icon("finder"))
  page.tile(1, 0, t => t.icon("trash"))
  page.tile(2, 0, t => t.icon("clock"))
})
```

#### Imperative

```javascript
const deck = require("loupedeck")

deck.tile(0, 0).setIcon("finder")
deck.tile(1, 0).setIcon("trash")
deck.tile(2, 0).setIcon("clock")
deck.flush()
```

### Scenario: easing-based touch feedback

#### Timeline style

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")
const easing = require("loupedeck/easing")

const tile = ui.page("home").tile(0, 0, t => t.icon("finder"))

ui.onTouch("Touch1", () => {
  anim.timeline()
    .to(tile, { scale: 1.15, y: -3 }, 120, easing.outBack)
    .to(tile, { scale: 1.0, y: 0 }, 180, easing.inOutCubic)
    .play()
})
```

#### Host effect shorthand

```javascript
ui.onTouch("Touch1", () => tile.effect("pressPop"))
```

### Scenario: reactive meter page

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")

const level = state.signal(0)

ui.page("meter", page => {
  page.tile(0, 0, t => t.progress(() => level.get() / 127))
  page.tile(1, 0, t => t.text(() => String(level.get())))
})

ui.onKnob("Knob1", delta => {
  level.set(Math.max(0, Math.min(127, level.get() + delta)))
})
```

## Page model brainstorm

### Page stack vs single active page

#### Single active page
- simplest
- good first cut

#### Stack with overlays
- more expressive
- useful for transient modal states

Recommendation:
- start with one active page plus optional overlays

## Error and runtime semantics

### What should happen if JS throws inside a callback?

Options:
1. crash whole runtime
2. log error and keep last good UI
3. disable only the failing callback/page

Recommendation:
- log and isolate as much as possible
- preserve current page if feasible
- expose runtime error hooks

### What should happen if a script loops forever?

Need host safeguards:
- execution time budget per callback/frame
- maybe interrupt support if available
- clear operator-visible error state

A dynamic UI runtime without watchdog semantics would be too fragile.

## Asset pipeline brainstorm

The current SVG icon loader in Go is a strong foundation. The JS runtime could expose icons at multiple levels:

### Level 1: named icons only

```javascript
tile.icon("finder")
```

### Level 2: load library in JS-visible registry

```javascript
const assets = require("loupedeck/assets")
assets.loadIconLibrary("/path/to/library.html")
```

### Level 3: custom image composition in JS

Potentially too much too early.

Recommendation:
- start with host-side asset registry and named icon usage
- keep raw image composition as an advanced future layer

## Recommended implementation direction

The strongest path appears to be:

### Phase 1: host-managed runtime shell

- `require("loupedeck")`
- event registration
- page switching
- timers
- simple asset registry

### Phase 2: retained page/tile API

- `require("loupedeck/ui")`
- tile and strip builders
- simple text/icon/progress widgets

### Phase 3: animation/tween API with easing

- `require("loupedeck/anim")`
- `require("loupedeck/easing")`
- timelines, sequences, repeats, yoyo

### Phase 4: explicit state helpers

- `require("loupedeck/state")`
- signals/computed/watch

### Phase 5: optional procedural hooks

- RAF-like callbacks or advanced canvas-ish effect layer if truly needed

## Design Decisions

### Decision: do not expose raw transport in JS

This is the most important constraint. JS should describe UI and behavior, not byte scheduling.

### Decision: prefer a hybrid retained model

A hybrid retained UI + timeline + signals model gives the best balance of elegance and transport safety.

### Decision: make easing a first-class module

Animations and easing are not garnish here; they are a core user goal. They deserve explicit API shape, not ad hoc callback math.

### Decision: keep a narrow imperative escape hatch

There will be cases where a small imperative hook is useful. It should exist, but it should not be the primary mode.

## Alternatives Considered

### Alternative A: only low-level imperative device API

Rejected as the main interface because it repeats the same architectural mistake that the Go package refactor just corrected.

### Alternative B: giant declarative retained framework immediately

Rejected as the first move because it commits too early to a large state model before validating the minimal retained primitives.

### Alternative C: animation-only API without retained pages/state

Rejected because elegant animation needs stable targets and stateful UI nodes underneath.

## Open Questions

1. Should the first user-facing JS runtime surface be one module or several (`loupedeck`, `ui`, `anim`, `state`, `easing`)?
2. Should timers/RAF semantics be host-driven or script-driven?
3. How much of the asset pipeline should be script-visible initially?
4. Should reconnect recovery be invisible to scripts or observable as lifecycle events?
5. How strict should the runtime watchdog / callback budget be?

## Implementation Plan

1. Choose the preferred module and state model direction.
2. Write a smaller RFC narrowing the initial API to one implementation slice.
3. Implement a tiny host runtime with event registration and page switching only.
4. Add a retained tile/page model.
5. Add the animation/easing module.
6. Only then evaluate more advanced procedural hooks.
