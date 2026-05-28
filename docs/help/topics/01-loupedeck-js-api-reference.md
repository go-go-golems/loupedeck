---
Title: Loupedeck JavaScript runtime API reference
Slug: loupedeck-js-api-reference
Short: Reference for the currently implemented goja modules, retained UI primitives, animation helpers, and live-runner behavior.
Topics:
- loupedeck
- javascript
- goja
- api
- runtime
- animation
- reactive-ui
Commands:
- loupedeck
Flags:
- script
- duration
- flush-interval
- queue-size
- send-interval
- log-events
- exit-on-circle
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This reference describes the **currently implemented** JavaScript runtime in this repository, not the broader brainstorm API from the early LOUPE-005 design docs. The important distinction is that the real runtime is intentionally narrow: it exposes retained pages, reactive state, and animation helpers above the Go-owned renderer and writer stack, and it does not expose raw framebuffer or transport operations.

This matters because the Loupedeck transport is fragile enough that letting scripts own low-level I/O would recreate the exact architecture problems the Go refactor was built to remove. Treat this page as the source of truth for what JavaScript can do **today**.

## Runtime model in one page

The current JS layer runs inside a goja VM that is owned by an explicit runtime-owner abstraction. Hardware events, animation callbacks, and reactive closures are all serialized back onto that owner thread before they execute JavaScript.

At a high level:

```text
script
-> goja runtime
-> loupedeck/state, loupedeck/ui, loupedeck/gfx, loupedeck/anim, loupedeck/easing
-> pure-Go reactive runtime + retained UI + host runtime + animation runtime
-> retained tile renderer
-> Go display/writer/transport stack
-> hardware
```

The practical consequence is simple: mutate state and retained UI from JavaScript, and let Go own everything below that boundary.

## Two rendering paths

Scripts can render content in two fundamentally different ways. Understanding the distinction is critical for writing performant scenes.

### Retained tile path (recommended for most scenes)

You declare *what* to show — `tile.text("HELLO")`, `tile.icon("circle")` — and Go decides *how* and *when* to render it. When a tile property changes, **only that tile** is re-rendered and sent to the hardware (90×90 pixels ≈ 32KB per tile). This is the fastest path for per-tile updates.

### Surface path (for custom pixel content)

You create a `gfx.surface(width, height)`, draw pixels into it from JavaScript, and assign it to a display or tile. When the surface changes, the display or tile it belongs to is re-rendered. You can assign surfaces at two levels:

- **Display-level surface**: `display.surface(fullDisplaySurface)` — the entire display (360×270 for main) is re-rendered when any pixel changes. Use this for full-display effects that span tile boundaries (ripples, scanlines, cross-tile animations).
- **Per-tile surface**: `tile.surface(tileSurface)` — **only that tile** is re-rendered when its surface changes (90×90). Use this for independent tile content (clocks, meters, status indicators).

**Performance rule of thumb:** If each tile's content is independent, use per-tile surfaces. If content spans tile boundaries, use a display-level surface. Per-tile surfaces send 12× less data per update on the main display.

## Module overview

| Module | Purpose | Main exports |
|---|---|---|
| `loupedeck/state` | Reactive values and watchers | `signal`, `computed`, `batch`, `watch` |
| `loupedeck/ui` | Retained pages, tiles, and hardware event subscriptions | `page`, `show`, `onButton`, `onTouch`, `onKnob` |
| `loupedeck/gfx` | Pixel surfaces and text rendering | `surface`, `font` |
| `loupedeck/present` | Frame scheduling and invalidation | `invalidate`, `onFrame` |
| `loupedeck/anim` | Numeric tweens, loops, and sequential timelines | `to`, `loop`, `timeline` |
| `loupedeck/easing` | Easing functions for animation | `linear`, `inOutQuad`, `inOutCubic`, `outBack`, `steps` |
| `loupedeck/metrics` | Low-level counters and timings | `inc`, `observeMillis`, `time`, `counted`, `now` |
| `loupedeck/scene-metrics` | Reusable scene-oriented metrics helpers on top of `loupedeck/metrics` | `create`, `reasonCategory` |

## `loupedeck/state`

The state module is the reactive core exposed to JavaScript. Use it whenever values should drive text, visibility, animation targets, or event-driven updates. The module is intentionally tiny because the real value comes from how it plugs into retained UI bindings.

### `state.signal(initial)`

Creates a mutable signal and returns an object with `get()`, `set(value)`, and `update(fn)`.

```javascript
const state = require("loupedeck/state");

const count = state.signal(0);

count.get();          // 0
count.set(1);         // count is now 1
count.update(v => v + 1);  // count is now 2
```

Why you use it:

- to hold the source-of-truth value for the current page
- to let multiple tile bindings read the same state
- to provide an animation target via `get()` and `set()`

### `signal.get()`

Returns the current exported JS value.

```javascript
const value = count.get();
```

If you call `get()` inside a reactive binding such as `tile.text(() => ...)`, the underlying Go reactive runtime tracks that dependency so later mutations re-run the binding.

### `signal.set(value)`

Sets a new value immediately.

```javascript
count.set(42);
```

If the new value is equal to the old one under the default equality check, the reactive runtime does not mark dependents dirty.

### `signal.update(fn)`

Reads the current value, calls your JS updater on the owner thread, and stores the returned value.

```javascript
count.update(v => v + 1);
```

Use `update(...)` when the next value depends on the current one. This avoids duplicating `get()` / `set()` logic in JavaScript and keeps the mutation shape clear.

### `state.computed(fn)`

Creates a derived value and returns an object with `get()`.

```javascript
const double = state.computed(() => count.get() * 2);
double.get();
```

Use `computed(...)` when you want one place to encode derived logic that several bindings can read. The current implementation keeps the surface minimal: there is no setter and no custom equality hook in JS.

### `state.batch(fn)`

Runs a group of state mutations as one reactive batch.

```javascript
state.batch(() => {
  left.set(10);
  right.set(20);
});
```

Why it matters:

- it reduces intermediate reactive churn
- it makes multi-value updates feel atomic at the UI layer
- it keeps related mutations grouped conceptually

### `state.watch(fn)`

Registers an eager watcher and returns an object with `stop()`.

```javascript
const sub = state.watch(() => {
  console.log(count.get());
});

sub.stop();
```

The function runs through the owner-thread bridge, just like other deferred JS callbacks. Use `watch(...)` when you want a side effect that follows reactive changes. Do **not** use it as your primary rendering API; tile bindings are the better fit for UI updates.

## `loupedeck/ui`

The UI module is the retained UI surface. It lets scripts declare named pages, add tiles to those pages, bind tile properties to reactive values, and listen to hardware events.

### `ui.page(name, fn)`

Creates or reuses a named page and optionally passes a page object to a configuration callback.

```javascript
const ui = require("loupedeck/ui");

ui.page("home", page => {
  page.tile(0, 0, tile => {
    tile.text("HELLO");
  });
});
```

The current runtime targets the **4×3 main display tile grid**. Tile coordinates are zero-based: column 0–3, row 0–2.

### `page.tile(col, row, fn)`

Creates or reuses a tile on the page and optionally configures it.

```javascript
page.tile(1, 2, tile => {
  tile.text("BOTTOM");
});
```

The tile callback receives a tile object with the methods documented below.

### `page.display(name, fn)`

Creates or reuses a display on the page and optionally configures it. Display names are `"left"`, `"main"`, and `"right"`.

```javascript
ui.page("scene", page => {
  page.display("left", display => {
    display.text("VU");
  });
  page.display("main", display => {
    display.surface(mainSurface);
  });
  page.display("right", display => {
    display.text("INFO");
  });
});
```

### `tile.text(valueOrFn)`

Sets static text or binds text to a reactive closure.

Static:

```javascript
tile.text("READY");
```

Reactive:

```javascript
tile.text(() => `COUNT ${count.get()}`);
```

Use the reactive form when the tile should update automatically after signal changes. The closure executes on the owner thread and its signal dependencies are tracked by the Go reactive runtime. When the signal changes, the tile is **marked dirty individually** — only that 90×90 tile is re-rendered and sent to hardware.

### `tile.icon(valueOrFn)`

Sets or binds the tile icon string.

```javascript
tile.icon("circle");
tile.icon(() => mode.get() === "armed" ? "record" : "stop");
```

Important current limitation: the JS renderer does **not** yet map icon names to the SVG asset pipeline. The retained renderer currently displays the icon string as centered placeholder text. This is still useful for structure and testing, but it is not yet the final asset story.

### `tile.visible(boolOrFn)`

Sets static visibility or binds visibility to a reactive boolean closure.

```javascript
tile.visible(true);

tile.visible(() => count.get() > 0);
```

If a tile is invisible, the retained renderer currently clears it to the themed background instead of drawing accent/text content.

### `tile.surface(surfaceOrNull)`

Sets a per-tile pixel surface, or clears it by passing `null`/`undefined`. When a tile has a surface, the renderer draws the surface content instead of the retained text/icon layout. When the surface changes, **only that tile** is re-rendered.

```javascript
const gfx = require("loupedeck/gfx");
const ui = require("loupedeck/ui");

const tileSurface = gfx.surface(90, 90);

ui.page("clock", page => {
  page.tile(0, 0, tile => {
    tile.surface(tileSurface);
  });
});

// Draw into the surface — only tile (0,0) is re-rendered
tileSurface.clear(0);
tileSurface.text("12:34", { x: 0, y: 20, width: 90, height: 20, center: true });
```

**Why this matters:** A per-tile surface gives you custom pixel content (charts, meters, patterns, custom fonts) while retaining per-tile invalidation. When the surface changes, the Go runtime marks only that tile dirty, and the renderer sends only 90×90 pixels (≈ 32KB) instead of the full display (≈ 389KB). This is 12× less data per update on the main display.

**When to use per-tile vs. display-level surfaces:**

| Scenario | Use | Why |
|---|---|---|
| Each tile has independent content (clock, meter, status) | `tile.surface(s)` | Only changed tiles are re-rendered |
| Full-display effects span tile boundaries (ripples, scanlines) | `display.surface(s)` | The effect covers the entire display anyway |
| Both independent tile content and full-display effects | Both | Use display layers for the background effect, per-tile surfaces for foreground content |

Pass `null` or `undefined` to remove a surface and revert to the retained text/icon rendering:

```javascript
tile.surface(null);
```

### Display objects

Display objects represent the three physical screen areas: `left` (60×270), `main` (360×270), and `right` (60×270). The main display contains the 4×3 tile grid.

#### `display.text(valueOrFn)`

Sets or binds display-level text. Shown as centered text on the display.

```javascript
display.text("STATUS");
display.text(() => mode.get());
```

#### `display.icon(valueOrFn)`

Sets or binds the display icon string. Same placeholder limitation as `tile.icon()`.

#### `display.visible(boolOrFn)`

Sets or binds display visibility.

#### `display.surface(surfaceOrNull)`

Sets a display-level pixel surface. When set, the renderer draws the surface content instead of the retained text/icon layout. **When any pixel in the surface changes, the entire display is re-rendered** — so use this for full-display content, not per-tile content.

```javascript
const gfx = require("loupedeck/gfx");

const mainSurface = gfx.surface(360, 270);

ui.page("full", page => {
  page.display("main", display => {
    display.surface(mainSurface);
  });
});
```

#### `display.layer(name, surfaceOrNull, opts?)`

Adds or removes a named compositing layer on the display. Layers are drawn on top of the base surface in the order they were added. Each layer can have a foreground tint.

```javascript
const base = gfx.surface(360, 270);
const overlay = gfx.surface(360, 270);

display.surface(base);
display.layer("fx", overlay);

// Tint the layer red
display.layer("accent", redOverlay, { r: 255, g: 0, b: 0, a: 255 });
```

Layers are useful for compositing effects on top of a base surface — for example, a background animation layer with a static content layer on top.

#### `display.tile(col, row, fn)`

Creates or reuses a tile on the main display. Same as `page.tile()` but called on the display object directly. Only available on the `"main"` display.

### `ui.show(name)`

Makes the named page active.

```javascript
ui.show("home");
```

This is the call that turns your retained page into something the renderer can actually flush. If you forget it, the script may build pages successfully but nothing becomes visible on the hardware.

### `ui.invalidate(reason)`

Marks the current frame as needing re-render. Use this when you need to trigger a redraw from outside of reactive state changes.

```javascript
const present = require("loupedeck/present");
present.invalidate("manual-update");
```

The `reason` string is passed through to the present runtime for debugging and metrics, but does not affect which regions are re-rendered — that is determined by the dirty tile/display tracking in the UI layer.

### `ui.onButton(name, fn)`

Registers a button handler and returns a subscription object with `close()`.

```javascript
const sub = ui.onButton("Button1", event => {
  console.log(event.name, event.status);
});

sub.close();
```

Event object fields:

| Field | Type | Meaning |
|---|---|---|
| `name` | string | The symbolic button name you subscribed to |
| `status` | string | `"down"` or `"up"` |

Supported button names:

- `Circle`
- `Button1` through `Button7`

### `ui.onTouch(name, fn)`

Registers a touch-region handler and returns a subscription object with `close()`.

```javascript
ui.onTouch("Touch6", event => {
  console.log(event.name, event.x, event.y);
});
```

Event object fields:

| Field | Type | Meaning |
|---|---|---|
| `name` | string | The symbolic touch region name |
| `status` | string | `"down"` or `"up"` |
| `x` | number | Touch X coordinate in device-space pixels |
| `y` | number | Touch Y coordinate in device-space pixels |

Supported touch names: `Touch1` through `Touch12`

### `ui.onKnob(name, fn)`

Registers a knob handler and returns a subscription object with `close()`.

```javascript
ui.onKnob("Knob1", event => {
  level.update(v => Math.max(0, Math.min(100, v + event.value)));
});
```

Event object fields:

| Field | Type | Meaning |
|---|---|---|
| `name` | string | The symbolic knob name |
| `value` | number | Signed delta from the hardware event |

Supported knob names: `Knob1` through `Knob6`

### Subscription objects

`ui.onButton(...)`, `ui.onTouch(...)`, and `ui.onKnob(...)` return a small subscription object:

```javascript
const sub = ui.onButton("Button1", () => {
  // ...
});

sub.close();
```

Use `close()` when a script installs a temporary handler that should not live for the entire process.

## `loupedeck/gfx`

The graphics module provides pixel surfaces for custom drawing and text rendering. Surfaces are grayscale (8-bit per pixel, 0–255). The renderer converts them to themed RGBA at flush time.

### `gfx.surface(width, height)`

Creates a new pixel surface and returns a surface object with the drawing methods documented below.

```javascript
const gfx = require("loupedeck/gfx");

const s = gfx.surface(90, 90);
```

Typical sizes:

| Surface | Width | Height | Use |
|---|---|---|---|
| Per-tile surface | 90 | 90 | `tile.surface(s)` — one tile |
| Main display | 360 | 270 | `display.surface(s)` — full main display |
| Left side display | 60 | 270 | `display.surface(s)` — left strip |
| Right side display | 60 | 270 | `display.surface(s)` — right strip |

### `gfx.font(path, opts)`

Loads an OpenType font and returns a font object. The font can be passed to `surface.text()` for custom rendering.

```javascript
const font = gfx.font("/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc", {
  size: 14,
  dpi: 72,
  index: 0,  // font index within a .ttc collection
});
```

Options:

| Option | Type | Default | Meaning |
|---|---|---|---|
| `size` | number | 12 | Font size in points |
| `dpi` | number | 72 | DPI for rasterization |
| `index` | number | 0 | Font index within a collection (.ttc) file |

Fonts are cached by path+options, so calling `gfx.font()` with the same arguments returns the same Go object.

### Surface drawing methods

#### `surface.width` / `surface.height`

Returns the surface dimensions.

```javascript
surface.width;   // 90
surface.height;  // 90
```

#### `surface.clear(v)`

Fills every pixel with the given value (0–255).

```javascript
surface.clear(0);    // all black
surface.clear(255);  // all white (within the theme)
```

#### `surface.batch(fn)`

Groups multiple drawing operations so that change notifications are coalesced into a single notification. Use this when making many changes to a surface at once.

```javascript
surface.batch(() => {
  surface.clear(0);
  surface.fillRect(0, 0, 30, 30, 100);
  surface.text("HELLO", { x: 0, y: 10, width: 90, height: 20, center: true });
});
// Only one OnChange notification is fired after the batch completes
```

**Performance note:** Always use `batch()` when making multiple changes to a surface between frames. Without batching, each individual drawing operation triggers a separate change notification, which can cause redundant re-renders.

#### `surface.set(x, y, v)`

Sets a pixel to a specific value (0–255), overwriting the previous value.

```javascript
surface.set(10, 20, 200);
```

#### `surface.add(x, y, v)`

Adds to a pixel value (saturating at 255). Use this for additive blending.

```javascript
surface.add(10, 20, 50);
```

#### `surface.fillRect(x, y, width, height, v)`

Fills a rectangle with a specific value.

```javascript
surface.fillRect(5, 60, 80, 10, 150);
```

#### `surface.line(x1, y1, x2, y2, v)`

Draws a line between two points using Bresenham's algorithm. Uses additive blending.

```javascript
surface.line(0, 0, 89, 89, 180);
```

#### `surface.crosshatch(x, y, width, height, density, v)`

Draws a crosshatch pattern within the given region. `density` controls the spacing between lines.

```javascript
surface.crosshatch(0, 0, 90, 90, 2, 30);
```

#### `surface.text(text, opts)`

Draws text onto the surface. The text is rendered as alpha-masked glyphs and blended into the grayscale pixel buffer.

```javascript
surface.text("HELLO", {
  x: 0,
  y: 10,
  width: 90,
  height: 20,
  brightness: 200,
  center: true,
  font: myFont,
});
```

Text options:

| Option | Type | Default | Meaning |
|---|---|---|---|
| `x` | int | 0 | X offset within the surface |
| `y` | int | 0 | Y offset within the surface |
| `width` | int | surface width | Width of the text rendering area |
| `height` | int | auto | Height of the text rendering area |
| `brightness` | int | 255 | Brightness multiplier (0–255) |
| `center` | bool | false | Center text horizontally within the width |
| `font` | font object | built-in 7×13 | Custom font to use |

**Current limitation:** Text is rendered as a single line. Newline characters (`\n`) are not handled correctly — they may render as missing glyphs. Word wrapping is not yet supported. These are planned improvements (see LOUPE-016).

#### `surface.compositeAdd(other, x, y)`

Adds the pixels of another surface onto this surface at the given offset, using additive blending.

```javascript
const overlay = gfx.surface(20, 20);
overlay.fillRect(0, 0, 20, 20, 100);
surface.compositeAdd(overlay, 10, 10);
```

#### `surface.at(x, y)`

Reads the pixel value at the given coordinates. Returns 0 for out-of-bounds.

```javascript
const v = surface.at(10, 20);
```

## `loupedeck/present`

The present module controls frame scheduling. It bridges between the JS side (which decides *when* to invalidate) and the Go side (which decides *what* to re-render based on dirty tracking).

### `present.invalidate(reason)`

Marks the current frame as needing re-render. The `reason` string is used for debugging and metrics only — it does not affect which tiles or displays are re-rendered. The actual re-render scope is determined by the UI dirty tracking.

```javascript
const present = require("loupedeck/present");
present.invalidate("tick");
```

**Important:** If you are using the retained tile path with reactive bindings, you typically do **not** need to call `invalidate()` manually. Signal changes automatically mark tiles dirty, and the UI dirty handler calls `invalidate()` for you. You only need `invalidate()` when:

- You are using display-level surfaces and need to trigger a full re-render
- You are modifying per-tile surfaces outside of reactive bindings

### `present.onFrame(fn)`

Registers a callback that runs on every invalidated frame, before the renderer flushes. The callback receives the reason string.

```javascript
const present = require("loupedeck/present");

present.onFrame(reason => {
  // Redraw all surfaces
  mainSurface.batch(() => {
    mainSurface.clear(0);
    drawScene();
  });
});
```

**When to use `onFrame()`:** Use it when you have display-level surfaces that need to be redrawn every frame (e.g., full-display animation scenes like `11-cyb-os-tiles.js`). For per-tile content with reactive bindings, you typically don't need `onFrame()` — signal changes drive tile updates automatically.

## `loupedeck/anim`

The animation module gives JavaScript access to the host-owned animation runtime. These helpers operate on **numeric targets**, not arbitrary tiles or scene objects. A valid target is any object that exposes `get()` and `set(value)`.

Signals are the intended primary target.

### `anim.to(target, to, durationMs, easeFn?)`

Tweens a numeric target to a new value and returns a handle with `stop()`.

```javascript
const anim = require("loupedeck/anim");
const easing = require("loupedeck/easing");

const opacity = state.signal(0);

const handle = anim.to(opacity, 1, 250, easing.inOutCubic);
```

Parameter meanings:

| Parameter | Meaning |
|---|---|
| `target` | Object with `get()` and `set(value)` |
| `to` | Final numeric value |
| `durationMs` | Duration in milliseconds |
| `easeFn` | Optional easing function; defaults to linear |

### `anim.loop(durationMs, fn)`

Runs a repeating loop that calls your function with a normalized phase from `0.0` to `< 1.0` and returns a handle with `stop()`.

```javascript
const pulse = state.signal(0);

const handle = anim.loop(1200, t => {
  pulse.set(t);
});
```

Use `loop(...)` when the script wants to derive its own animated value instead of asking the runtime for a target tween.

### `anim.timeline()`

Creates a sequential timeline builder. The builder supports `.to(...)` chaining and `.play()`.

```javascript
const timeline = anim.timeline()
  .to(level, 100, 200, easing.outBack)
  .to(level, 50, 180, easing.inOutCubic);

const handle = timeline.play();
```

This is a **sequential** timeline in the current implementation. Each step begins after the previous tween finishes.

### Animation handles

All current animation entry points return a handle object with `stop()`.

```javascript
const handle = anim.loop(1000, t => {
  pulse.set(t);
});

handle.stop();
```

Stopping the handle stops future timer-driven updates. It does not rewind the signal automatically.

## `loupedeck/easing`

The easing module exposes pure functions that map `t` in `[0, 1]` to an eased `t` in `[0, 1]`.

### Available functions

| Function | Purpose |
|---|---|
| `linear(t)` | Straight interpolation |
| `inOutQuad(t)` | Smooth quadratic ease-in/ease-out |
| `inOutCubic(t)` | Smooth cubic ease-in/ease-out |
| `outBack(t)` | Overshoots before settling |
| `steps(n)` | Returns a stepped easing function |

Example:

```javascript
const easing = require("loupedeck/easing");

const blink = easing.steps(2);
const value = blink(0.75);
```

These functions are useful both as tween easing functions and as ordinary numeric transforms inside reactive bindings.

## `loupedeck/metrics`

The low-level metrics module is the narrow bridge from JavaScript into the Go-owned in-process metrics collector. It is intentionally small and generic.

### `metrics.inc(name, delta = 1)`

Increments a named counter.

```javascript
const metrics = require("loupedeck/metrics");
metrics.inc("scene.frames");
metrics.inc("scene.activations", 2);
```

### `metrics.observeMillis(name, value)`

Records a timing sample in milliseconds.

```javascript
metrics.observeMillis("scene.renderAll", 12.5);
```

### `metrics.time(name, fn)`

Times a synchronous block and records the elapsed milliseconds.

```javascript
metrics.time("scene.renderAll", () => {
  renderAll();
});
```

### `metrics.counted(name, fn)`

Increments a counter and then executes a synchronous block.

```javascript
metrics.counted("scene.frames", () => {
  renderAll();
});
```

### `metrics.now()`

Returns the current wall-clock time in milliseconds.

```javascript
const t0 = metrics.now();
```

## `loupedeck/scene-metrics`

The scene-metrics module is the reusable higher-level helper package for scene authors. Use it when you want consistent metric names and common patterns like rebuild-reason tracking, activation counting, loop tick counting, and per-tile timing without repeating string-building logic in every scene.

### `sceneMetrics = require("loupedeck/scene-metrics").create(prefix)`

Creates a helper object whose counters and timings are automatically namespaced under `prefix`.

```javascript
const sceneMetrics = require("loupedeck/scene-metrics").create("scene");
```

### `sceneMetrics.time(suffix, fn)`

Times a block and records it under `prefix + "." + suffix`.

```javascript
sceneMetrics.time("renderAll", () => {
  renderAll();
});
```

### `sceneMetrics.timeTile(name, fn)`

Records per-tile timing under `prefix + ".tile." + name`.

```javascript
sceneMetrics.timeTile("SPIRAL", () => {
  drawSpiralTile(...);
});
```

### `sceneMetrics.recordLoopTick()`

Increments `prefix + ".loopTicks"`.

### `sceneMetrics.recordActivation(reason)`

Records `prefix + ".activations"` plus a categorized activation counter such as `prefix + ".activations.touch"` or `prefix + ".activations.button"`.

```javascript
sceneMetrics.recordActivation("T3");
sceneMetrics.recordActivation("B1");
```

### `sceneMetrics.recordRebuild(reason, fn)`

Tracks a rebuild cause and, when a function is provided, times the rebuild body.

Recorded counters include:
- `prefix + ".renderAll.calls"`
- `prefix + ".renderAll.reason.<category>"`
- `prefix + ".renderAll.reasonExact.<reason>"`

If `fn` is provided, the timing is recorded as:
- `prefix + ".renderAll"`

```javascript
sceneMetrics.recordRebuild("loop", () => {
  renderAll();
});
```

### `sceneMetrics.reasonCategory(reason)`

Maps a reason like `loop`, `initial`, `T12`, or `B1` into a stable category such as `loop`, `initial`, `touch`, or `button`.

## Live-runner command reference

The main hardware execution path is `cmd/loupedeck`, with the live runner exposed as the `run` subcommand. It is a Cobra/Glazed command, and these flags are the operational surface you will use while developing scripts.

### Basic usage

```bash
go run ./cmd/loupedeck run ./examples/js/01-hello.js --duration 5s
```

### Important flags

| Flag | Meaning | Why you care |
|---|---|---|
| positional `script` argument | Path to the JS file | Required entry point |
| `--device` | Optional serial device override | Use when auto-detect is wrong or unavailable |
| `--duration` | How long to run | Defaults to `0s` (run until interrupted) |
| `--flush-interval` | Retained-render flush cadence | Useful when experimenting with update pacing |
| `--queue-size` | Writer queue size | Useful for stress testing or animation experiments |
| `--send-interval` | Writer pacing interval | Lets you tune device-facing traffic |
| `--log-events` | Logs high-level hardware events | Excellent for hardware validation and debugging |
| `--exit-on-circle` | Circle exits the process when true | Disable when the script wants to use Circle itself |

### Important operational rule

If your script uses Circle as an input, run with:

```bash
--exit-on-circle=false
```

Otherwise the runner will exit before your script callback becomes useful.

## Example patterns

### Reactive counter (retained tile path)

The simplest pattern: button press → signal change → tile text updates automatically.

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");

const count = state.signal(0);

ui.page("counter", page => {
  page.tile(0, 0, tile => tile.text("BUTTON1"));
  page.tile(1, 0, tile => tile.text(() => `COUNT ${count.get()}`));
});

ui.onButton("Button1", () => {
  count.update(v => v + 1);
});

ui.show("counter");
```

### Per-tile surface clock

Use per-tile surfaces for custom drawing that only invalidates the changed tile. This is 12× faster than redrawing the full display.

```javascript
const gfx = require("loupedeck/gfx");
const ui = require("loupedeck/ui");
const anim = require("loupedeck/anim");

// Create per-tile surfaces
const clockTile = gfx.surface(90, 90);

ui.page("clock", page => {
  page.tile(0, 0, tile => {
    tile.surface(clockTile);
  });
});

// Update the clock tile surface every second
anim.loop(1000, () => {
  const d = new Date();
  const h = String(d.getHours()).padStart(2, "0");
  const m = String(d.getMinutes()).padStart(2, "0");
  const s = String(d.getSeconds()).padStart(2, "0");

  clockTile.batch(() => {
    clockTile.clear(0);
    clockTile.text(`${h}:${m}`, { x: 0, y: 20, width: 90, height: 24, center: true, brightness: 220 });
    clockTile.text(s, { x: 0, y: 48, width: 90, height: 14, center: true, brightness: 100 });
  });
  // Only tile (0,0) is re-rendered — 32KB instead of 389KB
});

ui.show("clock");
```

### Full-display surface scene

Use a display-level surface when effects span tile boundaries (ripples, scanlines, cross-tile animations). This is the most flexible but slowest path — every frame redraws the entire 360×270 display.

```javascript
const gfx = require("loupedeck/gfx");
const ui = require("loupedeck/ui");
const present = require("loupedeck/present");
const anim = require("loupedeck/anim");

const main = gfx.surface(360, 270);

ui.page("full", page => {
  page.display("main", display => {
    display.surface(main);
  });
});

let frame = 0;

function drawScene() {
  main.batch(() => {
    main.clear(0);
    // Draw all 12 tiles, background effects, scanlines, etc.
    frame++;
  });
}

present.onFrame(() => {
  drawScene();
});

anim.loop(2000, () => {
  present.invalidate("loop");
});

ui.show("full");
```

### Hybrid: display layers + per-tile surfaces

Combine display-level background effects with per-tile foreground content for the best of both worlds.

```javascript
const gfx = require("loupedeck/gfx");
const ui = require("loupedeck/ui");

const bgEffect = gfx.surface(360, 270);
const tileContent = gfx.surface(90, 90);

ui.page("hybrid", page => {
  page.display("main", display => {
    display.surface(bgEffect);       // background layer (full display)
  });

  page.tile(0, 0, tile => {
    tile.surface(tileContent);       // per-tile content (only this tile redraws)
  });
});
```

When `tileContent` changes, only tile (0,0) is re-rendered. When `bgEffect` changes, the entire display is re-rendered (because the background spans all tiles).

### Knob-driven numeric state

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");

const level = state.signal(50);

ui.page("knob", page => {
  page.tile(0, 0, tile => tile.text(() => `${level.get()}%`));
});

ui.onKnob("Knob1", event => {
  level.update(v => Math.max(0, Math.min(100, v + event.value)));
});

ui.show("knob");
```

### Animation loop driving a signal

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const anim = require("loupedeck/anim");
const easing = require("loupedeck/easing");

const pulse = state.signal(0);

ui.page("pulse", page => {
  page.tile(0, 0, tile => {
    tile.text(() => `${Math.round(easing.inOutCubic(pulse.get()) * 100)}%`);
  });
});

anim.loop(1000, t => {
  pulse.set(t);
});

ui.show("pulse");
```

## Current limitations and non-goals

The current API is deliberately narrower than the long-term brainstorm docs.

What is implemented today:

- retained pages and tiles on the main display
- per-tile surfaces with per-tile invalidation
- display-level surfaces with layers and compositing
- reactive state
- hardware event callbacks
- numeric animation helpers
- easing helpers
- owned-runtime callback serialization
- live hardware execution through `loupedeck run`

What is **not** implemented yet:

- raw transport or framebuffer access from JavaScript
- a JS assets module
- full JS-driven SVG/icon raster asset support
- direct JS timer APIs such as `setTimeout` / `setInterval`
- text newline rendering (planned, see LOUPE-016)
- text word wrapping (planned, see LOUPE-016)
- left/right strip retained UI in the JS layer
- advanced scene-graph widgets beyond simple tiles

These omissions are intentional. The current boundary preserves Go-side transport ownership and keeps the first runtime slice understandable.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `ui.onButton("Unknown", ...)` throws | The symbol is not in the supported button-name table | Use one of `Circle`, `Button1`…`Button7` |
| A touch callback never fires | The region name does not match the current module names | Use `Touch1` through `Touch12` exactly |
| `anim.to(...)` panics about `get()` or `set()` | The target is not a numeric target object | Pass a signal or another object that exposes `get()` and `set(value)` |
| The script builds a page but nothing renders | No active page exists | Call `ui.show("page-name")` |
| The app exits when Circle is pressed | The runner default exit behavior is still active | Run with `--exit-on-circle=false` |
| You see text where you expected icons | `tile.icon(...)` is currently a placeholder string in the JS renderer | Treat icons as labels until the asset layer is wired into JS |
| An animation callback or hardware callback seems to stop after shutdown | The owned runtime suppresses post-close callback execution | Re-run the process; do not expect closed runtimes to keep dispatching work |
| Reconnect sometimes fails with malformed HTTP or closed-port warnings | The device lifecycle is still somewhat fragile after abrupt stops | Retry cleanly, prefer `Ctrl-C` or Circle exits, and avoid piling overlapping runs on the same device |
| Full-display surface scene is slow | Every frame redraws 360×270 pixels | Switch to per-tile surfaces if tiles are independent — see the per-tile clock pattern above |
| Text with `\n` renders garbled | Newline characters are not handled by the text renderer | This is a known bug (LOUPE-016); avoid `\n` in text for now |
| Long text overflows tile boundaries | Word wrapping is not yet implemented | Keep text under ~12 characters per line with the default 7×13 font |

## See Also

- [Build your first live Loupedeck JavaScript script](../tutorials/01-build-your-first-live-loupedeck-js-script.md) — Step-by-step user guide for writing and running a real script
- `runtime/js/module_ui/module.go` — Concrete source of truth for the UI module exports
- `runtime/js/module_gfx/module.go` — Concrete source of truth for the gfx module exports
- `runtime/js/module_present/module.go` — Concrete source of truth for the present module exports
- `runtime/js/module_state/module.go` — Concrete source of truth for the reactive state exports
- `runtime/js/module_anim/module.go` — Concrete source of truth for the animation exports
- `runtime/js/module_easing/module.go` — Concrete source of truth for the easing exports
- `cmd/loupedeck/main.go` — Primary CLI root
- `cmd/loupedeck/cmds/run/command.go` — Current live hardware runner command
- `examples/js/` — Repository example scripts that match this API surface
