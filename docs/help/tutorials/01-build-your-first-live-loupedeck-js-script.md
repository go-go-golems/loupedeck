---
Title: Build your first live Loupedeck JavaScript script
Slug: loupedeck-js-first-live-script
Short: Build and run a small reactive script on a real Loupedeck Live with the current goja-based runtime.
Topics:
- loupedeck
- javascript
- goja
- tutorial
- runtime
- reactive-ui
Commands:
- loupedeck
Flags:
- script
- duration
- flush-interval
- log-events
- exit-on-circle
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial walks through the smallest useful end-to-end workflow for the current JavaScript runtime: write a script, run it through the live runner, press hardware controls, and watch the retained UI update on the device.

The important idea is that your script does **not** talk to the serial transport directly. It mutates state and retained UI objects, and the Go runtime owns rendering, flushing, and pacing. This layering exists because the Loupedeck Live transport is sensitive to timing and framing — letting JavaScript own the writer would recreate the exact problems the Go refactor was built to remove.

## What you'll build

You will build a tiny page with four tiles:

- a title tile
- a live counter tile
- a hint tile
- an exit hint tile

`Button1` increments the counter. The Circle button still exits the runner, which makes the example easy to start and stop.

## Prerequisites

Before you start, make sure the hardware and repository state are sane. The live runner expects a real Loupedeck Live connected over USB serial, and stale processes can temporarily keep the serial device busy.

You need:

- this repository checked out
- a connected Loupedeck Live
- `go test ./...` passing in the repo
- no other process currently owning the device

A quick validation loop:

```bash
go test ./...
```

If the hardware is busy, stop older `loupedeck run` sessions or other example/dev-tool runs before continuing.

## Step 1 — Write a minimal reactive script

The current runtime exposes modules through `require(...)`. For a first script, the most useful pair is:

- `require("loupedeck/state")` for reactive values
- `require("loupedeck/ui")` for pages, tiles, and hardware events

Create a file such as `/tmp/loupedeck-button1-counter.js` with this content:

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");

const count = state.signal(0);

ui.page("counter", page => {
  page.tile(0, 0, tile => {
    tile.text("BUTTON1");
  });

  page.tile(1, 0, tile => {
    tile.text(() => `COUNT ${count.get()}`);
  });

  page.tile(2, 0, tile => {
    tile.text("PRESS");
  });

  page.tile(3, 0, tile => {
    tile.text("CIRCLE EXIT");
  });
});

ui.onButton("Button1", () => {
  count.update(v => v + 1);
});

ui.show("counter");
```

Why this shape works:

- `state.signal(0)` creates the mutable counter state cell
- `tile.text(() => ...)` binds retained tile text to reactive state — when the signal changes, **only that tile** is re-rendered and sent to the hardware
- `ui.onButton("Button1", ...)` registers a hardware callback
- `ui.show("counter")` makes the page active so the renderer can flush it

If you skip `ui.show(...)`, the page exists but nothing becomes active, so the live runner has no active page to flush.

## Step 2 — Run the script on the device

The live hardware entry point is `cmd/loupedeck`, with the hardware runner exposed as the `run` subcommand. It loads the script into the owned goja runtime, attaches the host runtime to the deck, and flushes retained UI to the main display on a timer.

Run:

```bash
go run ./cmd/loupedeck run \
  --script /tmp/loupedeck-button1-counter.js \
  --duration 0 \
  --log-events
```

Why these flags matter:

- `--script` points at the JS file to execute
- `--duration 0` means "run until interrupted" instead of timing out
- `--log-events` prints high-level button, touch, and knob events so you can verify what the hardware delivered

When the script starts cleanly, you should see initial draw activity and then a log line similar to:

```text
INFO Loupedeck JS live runner started script=/tmp/loupedeck-button1-counter.js duration=0s flush_interval=16ms
```

## Step 3 — Interact on the hardware

At this point, the top row of the main display should show the four tiles you defined. Press `Button1` a few times.

What should happen in practice:

- the counter tile changes from `COUNT 0` to `COUNT 1`, `COUNT 2`, and so on
- the log prints high-level button events
- the runner keeps going until you press Circle or interrupt it from the terminal

The important semantic detail is that the button callback does not mutate pixels directly. It mutates `count`, which re-runs the bound text closure, which marks the tile dirty, which the retained renderer flushes on the next tick. **Only the changed tile is re-rendered** — the other three tiles are untouched, saving bandwidth and CPU.

## Step 4 — Stop the run cleanly

Press Circle to exit, or stop the process from the terminal with `Ctrl-C`.

If you want to keep Circle for your own script logic instead of exit behavior, disable the default exit hook:

```bash
go run ./cmd/loupedeck run \
  --script /tmp/loupedeck-button1-counter.js \
  --duration 0 \
  --exit-on-circle=false \
  --log-events
```

This matters for scripts like `examples/js/02-counter-button.js`, which intentionally use Circle as the application input.

## Step 5 — Use the built-in example pack

Once the first custom script works, switch to the repository examples. These match the current implementation, and several have been validated on real hardware.

Examples in the repo:

| Example | What it shows |
|---|---|
| `01-hello.js` | Minimal static tile text |
| `02-counter-button.js` | Reactive counter with button input |
| `03-knob-meter.js` | Knob-driven numeric display |
| `04-touch-feedback.js` | Touch event handling |
| `05-pulse-animation.js` | Animation loop driving a signal |
| `06-page-switcher.js` | Multi-page navigation |
| `13-per-tile-clock.js` | Per-tile surfaces with custom drawing |
| `14-tile-surface-drawing.js` | All gfx primitives showcase |
| `15-tile-draw-clock.js` | `tile.draw()` and `tile.invalidate()` pattern |

Try the page-switcher example:

```bash
go run ./cmd/loupedeck run \
  --script ./examples/js/06-page-switcher.js \
  --duration 0 \
  --log-events
```

## Step 6 — Custom pixel content with per-tile surfaces

The retained tile path (`tile.text()`, `tile.icon()`) handles text and simple labels. For anything custom — charts, meters, patterns, custom fonts — use a per-tile surface.

A per-tile surface is a 90×90 pixel buffer you draw into from JavaScript. When it changes, **only that tile** is re-rendered and sent to the hardware. This is much faster than redrawing the entire 360×270 display.

### The `tile.draw()` shorthand

The easiest way to draw custom content on a tile is `tile.draw(fn)`. It auto-creates a per-tile surface if needed and passes it to your drawing function:

```javascript
const ui = require("loupedeck/ui");

ui.page("demo", page => {
  page.tile(0, 0, tile => {
    tile.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 8, 100);  // accent bar
      s.text("HELLO", { x: 0, y: 30, width: 90, height: 20, center: true });
    });
  });
});
```

### The `tile.surface()` pattern (for animation updates)

When you need to redraw the surface on each frame (e.g., animation), create the surface separately and keep a reference so you can modify it later:

```javascript
const gfx = require("loupedeck/gfx");
const ui = require("loupedeck/ui");
const anim = require("loupedeck/anim");

const tileSurface = gfx.surface(90, 90);

ui.page("meter", page => {
  page.tile(0, 0, tile => {
    tile.surface(tileSurface);
  });
});

const level = state.signal(50);

// Redraw the tile surface when the level changes
anim.loop(100, () => {
  tileSurface.batch(() => {
    tileSurface.clear(0);
    const h = Math.round(level.get() * 0.8);
    tileSurface.fillRect(10, 80 - h, 70, h, 180);
  });
  // Only tile (0,0) is re-rendered — 32KB instead of 389KB
});

ui.show("meter");
```

### Multi-line and wrapped text

Both `tile.text()` and `surface.text()` support newlines and word wrapping:

```javascript
// Multi-line text with \n
tile.text("LINE 1\nLINE 2");

// Word-wrapped text (wraps to tile width)
tile.text("Very long label text", { wrap: true });

// Multi-line surface text with line gap
surface.text("ABOVE\nBELOW", {
  x: 0, y: 0, width: 90, center: true,
  lineGap: 4,  // 4 extra pixels between lines
});

// Word-wrapped surface text
surface.text("This is a very long label", {
  x: 4, y: 0, width: 90, center: true,
  wrapWidth: 82,  // wrap at 82 pixels (90 - 8 padding)
});
```

Key points about per-tile surfaces:

- Create a `gfx.surface(90, 90)` — the tile is 90×90 pixels
- Assign it with `tile.surface(tileSurface)`, or use `tile.draw(fn)` for the shorthand
- Draw into it with `surface.text()`, `surface.fillRect()`, `surface.line()`, etc.
- Always use `surface.batch(fn)` when making multiple drawing calls — this coalesces change notifications so the tile is only re-rendered once
- When the surface changes, only that tile is re-rendered — other tiles are unaffected
- Use `tile.invalidate()` to explicitly mark a tile dirty (usually not needed — surface changes auto-mark the tile)

**When to use per-tile vs. display-level surfaces:**

| Scenario | Use this | Why |
|---|---|---|
| Each tile has independent content (clock, meter, status) | `tile.surface(s)` or `tile.draw(fn)` | Only changed tiles are re-rendered |
| Full-display effects span tile boundaries (ripples, scanlines) | `display.surface(s)` | The effect covers the entire display anyway |
| Both independent tiles and full-display effects | Both | Use display layers for background, per-tile for foreground |

## How the runtime thinks about your script

The current JavaScript API is easiest to understand as a layered system:

```text
script
-> require("loupedeck/ui"), require("loupedeck/state"), require("loupedeck/anim")
-> owned goja runtime
-> pure-Go reactive runtime and retained UI model
-> retained tile renderer (per-tile dirty tracking)
-> live runner flush loop
-> package-owned display/writer/transport stack
-> hardware
```

That layering is why the API feels high-level even though the device transport is fragile. JavaScript talks to state, pages, tiles, and animations. Go keeps ownership of transport and rendering policy.

### Two rendering paths explained

The system has two rendering paths with different invalidation behavior:

**Retained tile path** (fast, per-tile):

```text
signal.set(value) → reactive flush → tile.BindText() closure re-runs
→ tile.SetText() → markDirtyTile() → only that tile is dirty
→ renderer.Flush() → renderTile() → 90×90 pixels sent
```

**Display-level surface path** (flexible, full-display):

```text
surface modification → markChanged() → display.markDirty()
→ renderer.Flush() → renderDisplay() → 360×270 pixels sent
```

**Per-tile surface path** (flexible + fast):

```text
tileSurface modification → markChanged() → tile.markDirty()
→ renderer.Flush() → renderTile() → 90×90 pixels sent
```

For most scenes, the per-tile surface path gives you the custom drawing flexibility of surfaces with the per-tile efficiency of the retained path.

## Complete example: animated pulse with per-tile surface

This example combines animation, reactive state, and per-tile surface drawing:

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");
const easing = require("loupedeck/easing");

const pulse = state.signal(0);
const tileSurface = gfx.surface(90, 90);

ui.page("pulse-meter", page => {
  page.tile(0, 0, tile => {
    tile.surface(tileSurface);
  });
  page.tile(1, 0, tile => {
    tile.text(() => `${Math.round(easing.inOutCubic(pulse.get()) * 100)}%`);
  });
});

anim.loop(1200, t => {
  pulse.set(t);

  // Redraw the meter tile surface
  const level = easing.inOutCubic(t);
  const h = Math.round(level * 70);
  tileSurface.batch(() => {
    tileSurface.clear(0);
    tileSurface.fillRect(10, 80 - h, 70, h, 180);
    tileSurface.line(10, 10, 80, 10, 40);
  });
});

ui.show("pulse-meter");
```

Run it:

```bash
go run ./cmd/loupedeck run \
  --script /tmp/pulse-meter.js \
  --duration 10s \
  --log-events
```

## Current limitations you should know up front

The current runtime is useful, but it is still the first real slice rather than the final platform.

Important current constraints:

- the JS-facing UI targets the **main 4×3 tile grid** and the left/right side displays
- `tile.icon(...)` currently stores a string and the placeholder renderer draws that string as text; it is not yet a full SVG/icon asset pipeline in the JS layer
- timers are host-owned internally, but they are not yet exposed as JS `setTimeout` / `setInterval`
- there is no JS `assets` module yet
- scripts do not get raw transport access, by design
- the goja VM is treated as **single-threaded** and all callbacks are serialized through the owner runner

These constraints are not accidents. They preserve the transport and rendering boundaries that keep the system stable.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `connect: unable to open port` | Another process still owns the device | Stop older `loupedeck run` or related runs, then retry |
| `malformed HTTP response ...` during connect | The device is in a fragile reconnect state | Wait a moment, reconnect, and prefer clean exits when switching demos |
| The screen stays blank | The script defined pages but never called `ui.show(...)` | Call `ui.show("page-name")` after building the page |
| Button presses appear in logs but the screen does not update | The callback is not mutating reactive or retained state | Update a `state.signal(...)` or a tile property from the event callback |
| Circle exits the app when you wanted to use it as input | The live runner defaults to `--exit-on-circle=true` | Re-run with `--exit-on-circle=false` |
| A tile bound with `tile.text(() => ...)` never changes | The closure is not reading reactive state, so there is nothing to invalidate it | Read a signal or computed value inside the closure, such as `count.get()` |
| You expected icons but only see words | The current JS renderer uses placeholder text rendering for `tile.icon(...)` | Treat icon strings as labels for now |
| Full-display surface scene feels slow | Every frame redraws 360×270 pixels | Switch to per-tile surfaces or `tile.draw()` if tiles are independent |
| Long text overflows tile boundaries | Text was not configured to wrap | Use `tile.text("long label", { wrap: true })` or `surface.text("...", { wrapWidth: 82 })` |
| Multi-line text needed | Use `\n` in text strings | Both `tile.text()` and `surface.text()` support `\n` for multi-line rendering |

## See Also

- [Loupedeck JavaScript runtime API reference](../topics/01-loupedeck-js-api-reference.md) — Detailed reference for modules, methods, events, and live-runner behavior
- `examples/js/` — Built-in example scripts that match the current implementation
- `cmd/loupedeck/main.go` — The main CLI entry point
- `cmd/loupedeck/cmds/run/command.go` — The live hardware runner used in this tutorial
- `runtime/js/module_ui/module.go` — JS-facing page, tile, and event bindings
- `runtime/js/module_gfx/module.go` — JS-facing surface and font bindings
- `runtime/js/module_state/module.go` — JS-facing reactive state bindings
