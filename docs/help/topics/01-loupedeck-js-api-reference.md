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
-> loupedeck/state, loupedeck/ui, loupedeck/anim, loupedeck/easing
-> pure-Go reactive runtime + retained UI + host runtime + animation runtime
-> retained tile renderer
-> Go display/writer/transport stack
-> hardware
```

The practical consequence is simple: mutate state and retained UI from JavaScript, and let Go own everything below that boundary.

## Module overview

| Module | Purpose | Main exports |
|---|---|---|
| `loupedeck/state` | Reactive values and watchers | `signal`, `computed`, `batch`, `watch` |
| `loupedeck/ui` | Retained pages, tiles, and hardware event subscriptions | `page`, `show`, `onButton`, `onTouch`, `onKnob` |
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

The current runtime targets the **4x3 main display tile grid**. Tile coordinates are zero-based and must fit that grid.

### `page.tile(col, row, fn)`

Creates or reuses a tile on the page and optionally configures it.

```javascript
page.tile(1, 2, tile => {
  tile.text("BOTTOM");
});
```

The tile callback receives a tile object. That object currently exposes `text(...)`, `icon(...)`, and `visible(...)`.

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

Use the reactive form when the tile should update automatically after signal changes. The closure executes on the owner thread and its signal dependencies are tracked by the Go reactive runtime.

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

### `ui.show(name)`

Makes the named page active.

```javascript
ui.show("home");
```

This is the call that turns your retained page into something the renderer can actually flush. If you forget it, the script may build pages successfully but nothing becomes visible on the hardware.

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

Supported button names in the current module:

- `Circle`
- `Button1`
- `Button2`
- `Button3`
- `Button4`
- `Button5`
- `Button6`
- `Button7`

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

Supported touch names in the current module:

- `Touch1` through `Touch12`

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

Supported knob names in the current module:

- `Knob1`
- `Knob2`
- `Knob3`
- `Knob4`
- `Knob5`
- `Knob6`

### Subscription objects

`ui.onButton(...)`, `ui.onTouch(...)`, and `ui.onKnob(...)` return a small subscription object:

```javascript
const sub = ui.onButton("Button1", () => {
  // ...
});

sub.close();
```

Use `close()` when a script installs a temporary handler that should not live for the entire process.

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

## Live-runner command reference

The main hardware execution path is now `cmd/loupedeck`, with the live runner exposed as the `run` subcommand. It is a Cobra/Glazed command, and these flags are the operational surface you will use while developing scripts.

### Basic usage

```bash
go run ./cmd/loupedeck run --script ./examples/js/01-hello.js --duration 5s
```

### Important flags

| Flag | Meaning | Why you care |
|---|---|---|
| `--script` | Path to the JS file | Required entry point |
| `--device` | Optional serial device override | Use when auto-detect is wrong or unavailable |
| `--duration` | How long to run | Use `0` for run-until-interrupted |
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

## Example patterns that match the current implementation

### Reactive counter

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

## See Also

- [Build your first live Loupedeck JavaScript script](../tutorials/01-build-your-first-live-loupedeck-js-script.md) — Step-by-step user guide for writing and running a real script
- `runtime/js/module_ui/module.go` — Concrete source of truth for the UI module exports
- `runtime/js/module_state/module.go` — Concrete source of truth for the reactive state exports
- `runtime/js/module_anim/module.go` — Concrete source of truth for the animation exports
- `runtime/js/module_easing/module.go` — Concrete source of truth for the easing exports
- `cmd/loupedeck/main.go` — Primary CLI root
- `cmd/loupedeck/cmds/run/command.go` — Current live hardware runner command
- `examples/js/` — Repository example scripts that match this API surface
