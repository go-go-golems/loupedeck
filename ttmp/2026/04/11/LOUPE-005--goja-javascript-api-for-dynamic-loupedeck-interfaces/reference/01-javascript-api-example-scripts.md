---
Title: JavaScript API example scripts
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
      Note: Main brainstorm doc that motivates the example APIs here
    - Path: cmd/loupe-svg-buttons/main.go
      Note: Current Go demo whose concepts are reimagined here as script-facing APIs
    - Path: svg_icons.go
      Note: Existing asset pipeline that could underpin named icon/image APIs
ExternalSources: []
Summary: Copy/paste-style example scripts for multiple possible goja JavaScript API designs, including declarative pages, imperative effects, reactive state, timelines, easing curves, page transitions, and interactive scenarios.
LastUpdated: 2026-04-11T20:40:45-04:00
WhatFor: Provide concrete examples of what a future JavaScript API could feel like under different design choices.
WhenToUse: Use when evaluating API ergonomics or selecting an implementation direction for the goja runtime.
---

# JavaScript API example scripts

## Goal

Provide concrete script examples for several possible JavaScript API styles so API ergonomics can be compared against real scenarios rather than only abstract descriptions.

## Context

These examples are intentionally exploratory. They are not all meant to be implemented exactly as written. Their purpose is to show what kinds of code a goja-based Loupedeck runtime could enable.

The examples assume modules like:

- `require("loupedeck")`
- `require("loupedeck/ui")`
- `require("loupedeck/state")`
- `require("loupedeck/anim")`
- `require("loupedeck/easing")`
- `require("loupedeck/assets")`

## Implemented live examples in this repository

The exploratory examples below are broader than the currently implemented runtime, but the repository now also contains a concrete live example pack under:

- `examples/js/01-hello.js`
- `examples/js/02-counter-button.js`
- `examples/js/03-knob-meter.js`
- `examples/js/04-touch-feedback.js`
- `examples/js/05-pulse-animation.js`
- `examples/js/06-page-switcher.js`

Those scripts are intended to run through:

- `cmd/loupe-js-live/main.go`

and the following were validated on actual Loupedeck Live hardware during this ticket:

- `01-hello.js` — static retained page render
- `02-counter-button.js` — Circle-button counter updates (run with `--exit-on-circle=false`)
- `03-knob-meter.js` — `Knob1` updates reactive numeric state
- `04-touch-feedback.js` — `Touch1`, `Touch6`, and `Touch12` update a status tile
- `05-pulse-animation.js` — auto-running animated state updates
- `06-page-switcher.js` — `Button1` / `Button2` switch retained pages

Note: `04-touch-feedback.js` was corrected during hardware validation so its visible tile labels now match the actual touched regions (`Touch1` at top-left, `Touch6` in the middle row, second tile, and `Touch12` at bottom-right).

## Quick reference

### Example 1: smallest possible imperative script

```javascript
const deck = require("loupedeck")

deck.page("home")
deck.tile(0, 0).icon("finder")
deck.tile(1, 0).icon("trash")
deck.tile(2, 0).text("REC")
deck.flush()
```

### Example 2: declarative page with static icons

```javascript
const ui = require("loupedeck/ui")

ui.page("home", page => {
  page.tile(0, 0, t => t.icon("finder"))
  page.tile(1, 0, t => t.icon("trash"))
  page.tile(2, 0, t => t.icon("clock"))
  page.tile(3, 0, t => t.icon("document"))
})

ui.show("home")
```

### Example 3: reactive state bound to a knob

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")

const value = state.signal(64)

ui.page("levels", page => {
  page.tile(0, 0, t => t.text(() => String(value.get())))
  page.tile(1, 0, t => t.progress(() => value.get() / 127))
})

ui.onKnob("Knob1", delta => {
  value.set(Math.max(0, Math.min(127, value.get() + delta)))
})

ui.show("levels")
```

### Example 4: touch feedback with easing curve

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")
const easing = require("loupedeck/easing")

const page = ui.page("home")
const tile = page.tile(0, 0, t => t.icon("finder"))

ui.onTouch("Touch1", () => {
  anim.timeline()
    .to(tile, { scale: 1.16, y: -4 }, 120, easing.outBack)
    .to(tile, { scale: 1.0, y: 0 }, 180, easing.inOutCubic)
    .play()
})

ui.show("home")
```

### Example 5: looping pulse animation

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")
const easing = require("loupedeck/easing")

const page = ui.page("armed")
const tile = page.tile(0, 0, t => t.icon("record").text("REC"))

anim.timeline({ repeat: Infinity, yoyo: true })
  .to(tile, { scale: 1.08 }, 450, easing.inOutSine)
  .to(tile, { scale: 1.0 }, 450, easing.inOutSine)
  .play()

ui.show("armed")
```

### Example 6: steps easing for retro blinking

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")
const easing = require("loupedeck/easing")

const page = ui.page("retro")
const tile = page.tile(0, 0, t => t.icon("clock"))

anim.to(tile, { opacity: 0 }, 600, easing.steps(2), {
  repeat: Infinity,
  yoyo: true,
})

ui.show("retro")
```

### Example 7: curated icon bank browser

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")
const assets = require("loupedeck/assets")

assets.loadIconLibrary("macos1-icon-library.html")

const icons = [
  "finder", "trash", "clock", "document",
  "disk", "folder", "key", "mic",
  "music", "network", "pause", "play",
  "record", "speaker", "stop",
]

const bank = state.signal(0)

ui.page("browser", page => {
  for (let row = 0; row < 3; row++) {
    for (let col = 0; col < 4; col++) {
      const slot = row * 4 + col
      page.tile(col, row, t => {
        t.icon(() => icons[bank.get() * 12 + slot] || null)
      })
    }
  }
})

ui.onButton("Button1", () => bank.set(Math.max(0, bank.get() - 1)))
ui.onButton("Button2", () => bank.set(bank.get() + 1))
ui.show("browser")
```

### Example 8: page transition with staggered easing

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")
const easing = require("loupedeck/easing")

const home = ui.page("home", page => {
  page.tile(0, 0, t => t.icon("finder"))
  page.tile(1, 0, t => t.icon("trash"))
  page.tile(2, 0, t => t.icon("clock"))
})

const mixer = ui.page("mixer", page => {
  page.tile(0, 0, t => t.text("A"))
  page.tile(1, 0, t => t.text("B"))
  page.tile(2, 0, t => t.text("C"))
})

ui.transition("home", "mixer", ctx => {
  ctx.out.tiles().stagger(25).to({ x: -100, opacity: 0 }, 220, easing.inOutCubic)
  ctx.in.tiles().from({ x: 100, opacity: 0 }).stagger(25).to({ x: 0, opacity: 1 }, 260, easing.outCubic)
})
```

### Example 9: knob scrubs an animation timeline

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")
const anim = require("loupedeck/anim")

const progress = state.signal(0)
const page = ui.page("scrub")
const tile = page.tile(0, 0, t => t.icon("finder"))

const tl = anim.timeline({ paused: true })
  .to(tile, { x: 18, scale: 1.14 }, 400, "outBack")
  .to(tile, { x: 0, scale: 1.0 }, 400, "inOutCubic")

ui.onKnob("Knob1", delta => {
  progress.set(Math.max(0, Math.min(1, progress.get() + delta / 100)))
  tl.seek(progress.get())
})

ui.show("scrub")
```

### Example 10: simple spring-like effect primitive

```javascript
const ui = require("loupedeck/ui")
const anim = require("loupedeck/anim")

const page = ui.page("spring")
const tile = page.tile(0, 0, t => t.icon("key"))

ui.onTouch("Touch1", () => {
  anim.spring(tile, { scale: 1.2, rotation: 8 }, {
    stiffness: 180,
    damping: 14,
    settleTo: { scale: 1.0, rotation: 0 },
  })
})
```

### Example 11: procedural frame callback for advanced users

```javascript
const deck = require("loupedeck")
const easing = require("loupedeck/easing")

let t = 0
const tile = deck.page("wave").tile(0, 0)
tile.icon("waveform")

deck.requestAnimationFrame(function frame(dt) {
  t += dt
  const phase = (t % 1000) / 1000
  tile.setScale(1 + 0.06 * easing.inOutSine(phase))
  tile.setY(Math.sin(phase * Math.PI * 2) * 3)
  deck.requestAnimationFrame(frame)
})
```

### Example 12: dynamic page router

```javascript
const ui = require("loupedeck/ui")
const state = require("loupedeck/state")

const current = state.signal("home")

ui.page("home", page => {
  page.tile(0, 0, t => t.icon("finder"))
  page.tile(1, 0, t => t.text("NEXT"))
})

ui.page("mixer", page => {
  page.tile(0, 0, t => t.text("VOL"))
  page.tile(1, 0, t => t.text("BACK"))
})

ui.onTouch("Touch2", () => current.set("mixer"))
ui.onTouch("Touch6", () => current.set("home"))
ui.show(() => current.get())
```

## Usage Examples

### Compare three styles for the same press-pop effect

#### Imperative style

```javascript
deck.onTouch("Touch1", () => {
  const tile = deck.tile(0, 0)
  tile.setScale(1.15)
  deck.setTimeout(() => tile.setScale(1.0), 140)
})
```

#### Declarative + effect shorthand

```javascript
ui.onTouch("Touch1", () => tile.effect("pressPop"))
```

#### Timeline + easing style

```javascript
ui.onTouch("Touch1", () => {
  anim.timeline()
    .to(tile, { scale: 1.15, y: -4 }, 110, easing.outBack)
    .to(tile, { scale: 1.0, y: 0 }, 170, easing.inOutCubic)
    .play()
})
```

### Compare state models for a knob-driven counter

#### Imperative mutable variable

```javascript
let value = 0
ui.onKnob("Knob1", d => {
  value = Math.max(0, Math.min(100, value + d))
  tile.text(String(value))
})
```

#### Signal-based state

```javascript
const value = state.signal(0)
ui.onKnob("Knob1", d => value.set(clamp(value.get() + d, 0, 100)))
page.tile(0, 0, t => t.text(() => String(value.get())))
```

The signal version is likely better once multiple widgets depend on the same value.

## Additional scenarios to think about

- audio meters driven by external host data
- transport controls with looping armed/record states
- macro pads with temporary overlays
- icon banks with animated page carousel transitions
- live dashboard pages whose values update from Go-side events
- onboarding/demo pages that sequence themselves automatically
- touch gestures that trigger easing-based slide transitions

## Related

- `design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md`
