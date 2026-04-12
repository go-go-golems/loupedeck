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
- loupe-js-live
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

This tutorial walks through the smallest useful end-to-end workflow for the current JavaScript runtime: write a script, run it through the live runner, press hardware controls, and watch the retained UI update on the device. The important idea is that your script does **not** talk to the serial transport directly. It mutates state and retained UI objects, and the Go runtime owns rendering, flushing, and pacing.

This matters because the Loupedeck Live is sensitive to transport behavior. The runtime exists to give scripts a convenient API **without** pushing writer, renderer, or connection policy into JavaScript.

## What you'll build

You will build a tiny page with four tiles:

- a title tile
- a live counter tile
- a hint tile
- an exit hint tile

`Button1` increments the counter. The Circle button still exits the runner, which makes the example easy to start and stop.

## Prerequisites

Before you start, make sure the hardware and repository state are sane. The live runner expects a real Loupedeck Live connected over USB serial, and stale processes can temporarily keep `/dev/ttyACM0` busy.

You need:

- the repo checked out at `/home/manuel/code/wesen/2026-04-11--loupedeck-test`
- a connected Loupedeck Live
- `go test ./...` passing in the repo
- no other process currently owning the device

A quick validation loop is:

```bash
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test
go test ./...
```

If the hardware is busy, stop older `loupe-js-live`, `loupe-svg-buttons`, or `loupe-feature-tester` runs before continuing.

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
- `tile.text(() => ...)` binds retained tile text to reactive state
- `ui.onButton("Button1", ...)` registers a hardware callback
- `ui.show("counter")` makes the page active so the renderer can flush it

If you skip `ui.show(...)`, the page exists but nothing becomes active, so the live runner has no active page to flush.

## Step 2 — Run the script on the device

The live hardware entry point is `cmd/loupe-js-live`. It loads the script into the owned goja runtime, attaches the host runtime to the deck, and flushes retained UI to the main display on a timer.

Run:

```bash
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test
go run ./cmd/loupe-js-live \
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

A typical event log looks like:

```text
INFO button event button=Button1 status=down
```

The important semantic detail is that the button callback does not mutate pixels directly. It mutates `count`, which re-runs the bound text closure, which marks the tile dirty, which the retained renderer flushes on the next tick.

## Step 4 — Stop the run cleanly

Press Circle to exit, or stop the process from the terminal with `Ctrl-C`.

If you want to keep Circle for your own script logic instead of exit behavior, disable the default exit hook:

```bash
go run ./cmd/loupe-js-live \
  --script /tmp/loupedeck-button1-counter.js \
  --duration 0 \
  --exit-on-circle=false \
  --log-events
```

This matters for scripts like `examples/js/02-counter-button.js`, which intentionally use Circle as the application input.

## Step 5 — Use the built-in example pack

Once the first custom script works, switch to the repository examples. These are useful because they match the current implementation, and several have already been validated on real hardware.

Examples currently in the repo:

- `examples/js/01-hello.js`
- `examples/js/02-counter-button.js`
- `examples/js/03-knob-meter.js`
- `examples/js/04-touch-feedback.js`
- `examples/js/05-pulse-animation.js`
- `examples/js/06-page-switcher.js`

Try the page-switcher example:

```bash
go run ./cmd/loupe-js-live \
  --script ./examples/js/06-page-switcher.js \
  --duration 0 \
  --log-events
```

This is a good next step because it proves that retained page switching works, not just simple text updates.

## How the current runtime thinks about your script

The current JavaScript API is easiest to understand as a layered system:

```text
script
-> require("loupedeck/ui"), require("loupedeck/state"), require("loupedeck/anim")
-> owned goja runtime
-> pure-Go reactive runtime and retained UI model
-> retained tile renderer
-> live runner flush loop
-> package-owned display/writer/transport stack
-> hardware
```

That layering is why the API feels high-level even though the device transport is fragile. JavaScript talks to state, pages, tiles, and animations. Go keeps ownership of transport and rendering policy.

## Complete example

If you want a slightly richer example that includes animation, the built-in pulse demo is the simplest current reference:

```javascript
const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const anim = require("loupedeck/anim");
const easing = require("loupedeck/easing");

const pulse = state.signal(0);

ui.page("pulse", page => {
  page.tile(0, 0, tile => {
    tile.text("PULSE");
  });
  page.tile(1, 0, tile => {
    tile.text(() => `${Math.round(easing.inOutCubic(pulse.get()) * 100)}%`);
  });
  page.tile(2, 0, tile => {
    tile.text("LOOP");
  });
  page.tile(3, 0, tile => {
    tile.text("RUN");
  });
});

anim.loop(1200, t => {
  pulse.set(t);
});

ui.show("pulse");
```

Run it with:

```bash
go run ./cmd/loupe-js-live \
  --script ./examples/js/05-pulse-animation.js \
  --duration 10s \
  --log-events
```

## Current limitations you should know up front

The current runtime is useful, but it is still the first real slice rather than the final platform.

Important current constraints:

- the JS-facing UI targets the **main 4x3 tile grid** only
- `tile.icon(...)` currently stores a string and the placeholder renderer draws that string as text; it is not yet a full SVG/icon asset pipeline in the JS layer
- timers are host-owned internally, but they are not yet exposed as JS `setTimeout` / `setInterval`
- there is no JS `assets` module yet
- scripts do not get raw transport access, by design
- the goja VM is treated as **single-threaded** and all callbacks are serialized through the owner runner

These constraints are not accidents. They preserve the transport and rendering boundaries that keep the system stable.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `connect: unable to open port "/dev/ttyACM0"` | Another process still owns the device | Stop older `loupe-js-live` or related runs, then retry |
| `malformed HTTP response ...` during connect | The device is in a fragile reconnect state after an earlier run | Wait a moment, reconnect, and prefer clean exits when switching demos |
| The screen stays blank | The script defined pages but never called `ui.show(...)` | Call `ui.show("page-name")` after building the page |
| Button presses appear in logs but the screen does not update | The callback is not mutating reactive or retained state | Update a `state.signal(...)` or a tile property from the event callback |
| Circle exits the app when you wanted to use it as input | The live runner defaults to `--exit-on-circle=true` | Re-run with `--exit-on-circle=false` |
| A tile bound with `tile.text(() => ...)` never changes | The closure is not reading reactive state, so there is nothing to invalidate it | Read a signal or computed value inside the closure, such as `count.get()` |
| You expected icons but only see words | The current JS renderer uses placeholder text rendering for `tile.icon(...)` | Treat icon strings as labels for now; full JS asset support is future work |

## See Also

- [Loupedeck JavaScript runtime API reference](../topics/01-loupedeck-js-api-reference.md) — Detailed reference for modules, methods, events, and live-runner behavior
- `examples/js/` — Built-in example scripts that match the current implementation
- `cmd/loupe-js-live/main.go` — The live hardware runner used in this tutorial
- `runtime/js/module_ui/module.go` — JS-facing page, tile, and event bindings
- `runtime/js/module_state/module.go` — JS-facing reactive state bindings
