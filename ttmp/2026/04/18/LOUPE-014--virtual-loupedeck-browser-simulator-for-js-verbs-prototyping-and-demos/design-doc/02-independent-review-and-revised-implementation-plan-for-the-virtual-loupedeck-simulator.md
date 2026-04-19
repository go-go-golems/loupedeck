---
Title: Independent review and revised implementation plan for the virtual Loupedeck simulator
Ticket: LOUPE-014
Status: active
Topics:
    - loupedeck
    - web
    - simulation
    - javascript
    - jsverbs
    - ui
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: loupedeck/README.md
      Note: Defines pkg/device as the low-level protocol package and supports the package-boundary critique
    - Path: loupedeck/cmd/loupedeck/cmds/run/command.go
      Note: Shows the real extraction point in runSceneSession and the existing raw/verb bootstrap split
    - Path: loupedeck/examples/cmd/loupe-js-demo/main.go
      Note: Existing non-hardware DrawTarget example that supports the revised plan
    - Path: loupedeck/runtime/js/jsverbs_integration_test.go
      Note: Proves jsverbs can build UI state without hardware
    - Path: loupedeck/runtime/js/runtime_test.go
      Note: Existing fake input-source tests that support a smaller composition-based simulator
    - Path: loupedeck/runtime/present/runtime.go
      Note: Render-before-flush ordering that affects when browser notifications should fire
    - Path: loupedeck/runtime/render/visual_runtime.go
      Note: Proves the renderer emits both full-display draws and partial tile patches
    - Path: loupedeck/ttmp/2026/04/18/LOUPE-014--virtual-loupedeck-browser-simulator-for-js-verbs-prototyping-and-demos/design-doc/01-virtual-loupedeck-simulation-architecture-and-implementation-guide.md
      Note: First-pass design being reviewed and corrected
ExternalSources: []
Summary: Second-pass design document that reviews the first simulator proposal, corrects scope and package-boundary issues, and proposes a more incremental implementation plan centered on a simulator command, input hub, frame compositor, and browser bridge.
LastUpdated: 2026-04-18T13:59:03-04:00
WhatFor: Provide an independent architectural review of the first simulator design and replace it with a more concrete, lower-risk implementation plan.
WhenToUse: Read before starting implementation if you want the shortest path to a working browser simulator without over-generalizing the device package or rewriting the hardware runner.
---


# Independent review and revised implementation plan for the virtual Loupedeck simulator

## Executive Summary

This document is a second-pass review of the first simulator design written for LOUPE-014. The first design is **directionally correct** on the most important point: the simulator should operate at the **semantic runtime level** rather than trying to emulate the serial wire protocol. That is the right foundation.

However, after re-reading the code and comparing it to the first plan, I think the original proposal is still **too large, too generic, and slightly misplaced in package boundaries**. In particular:

- it leans too hard toward inventing a new all-purpose virtual backend interface,
- it suggests putting simulator code under `pkg/device`, even though the repository explicitly describes `pkg/device` as the low-level device/protocol package,
- it does not fully account for how `runtime/render.Renderer.Flush()` emits **partial tile patches**, not just full display frames,
- and it reaches for a browser “cockpit” sooner than necessary.

My revised proposal is more incremental:

1. **Do not create a virtual `pkg/device.Loupedeck` clone first.**
2. **Do not move simulator concerns into `pkg/device` unless a later need clearly emerges.**
3. **Start with a new `simulate` command plus a `runtime/sim` package** that contains:
   - an input hub implementing `host.EventSource`,
   - a frame compositor implementing named `render.DrawTarget`s,
   - a browser bridge (HTTP + websocket),
   - and a small server UI.
4. **Reuse the current script and jsverb bootstrap path** instead of inventing a parallel load model.
5. **Defer browser-side verb forms to phase 2**; use CLI flags first.

That route gets to a working browser simulator faster, with fewer architectural commitments and less risk of muddying the public `pkg/device` boundary.

---

## Part I — In-depth review of the first design

## 1. What the first design gets right

The first design's biggest strength is that it found the right conceptual center of gravity. The repository is not “just a serial driver.” The runtime already has a semantic model above transport, and the first design correctly recognized that the simulator should attach there.

### 1.1 It correctly rejects wire-protocol emulation as the starting point

This is the correct call.

Evidence from the code:

- real hardware connection is still tied to serial port enumeration and a websocket dial over serial (`pkg/device/connect.go:13-52`, `pkg/device/connect.go:74-217`),
- `pkg/device/dialer.go` is explicitly about `SerialWebSockConn` and USB serial open logic,
- display output is encoded into protocol messages (`pkg/device/display.go:81-157`),
- input is decoded from binary messages in `Listen()` (`pkg/device/listen.go:11-95`).

A browser simulator that starts by emulating all of that would be solving the wrong problem. The user asked for a better **prototype/demo workflow** for JS verbs, not a firmware emulator.

### 1.2 It correctly identifies the best existing semantic seams

The first design was right to call out:

- `runtime/host.EventSource` for inputs (`runtime/host/runtime.go:10-14`),
- `runtime/render.DrawTarget` for output (`runtime/render/visual_runtime.go:23-25`),
- `runtime/ui.UI` for retained state (`runtime/ui/ui.go:16-209`),
- `runtime/js` and the JS modules for scene-level APIs.

This is strongly supported by current tests and examples:

- `runtime/js/runtime_test.go` already uses a `fakeSource` to drive button callbacks into the JS runtime without real hardware (`runtime/js/runtime_test.go:21-69`, `runtime/js/runtime_test.go:127-206`),
- `examples/cmd/loupe-js-demo/main.go` already renders retained UI into a non-hardware output target (`examples/cmd/loupe-js-demo/main.go:17-65`),
- `runtime/js/jsverbs_integration_test.go` proves an annotated verb can build UI state in a live runtime without attaching any physical device (`runtime/js/jsverbs_integration_test.go:11-63`).

That is the strongest evidence in the repository that the simulator does **not** need to start from the transport layer.

### 1.3 It correctly preserves the real hardware path

The first design does not propose deleting or rewriting the serial path immediately. That is also correct.

The current `runSceneSession()` function is explicitly real-device oriented: it connects hardware, starts `Listen()`, attaches the device as the host event source, and renders to the hardware displays (`cmd/loupedeck/cmds/run/command.go:339-468`). That path still matters for final hardware validation even if the simulator becomes the preferred prototyping flow.

### 1.4 It correctly recognizes Live-first scope

The README calls out Loupedeck Live (`product 0004`) as the current hardware focus (`README.md:22-26`), and `pkg/device/profile.go` already defines the Live display layout (`pkg/device/profile.go:45-53`). Starting the simulator with Live as the default profile is the right pragmatic scope.

---

## 2. Where the first design is too broad or slightly off

The first design's core idea is good, but I would not implement it exactly as written.

## 2.1 It overstates the need for a new generic `Backend` abstraction

The first design proposes a new generic backend interface combining `host.EventSource`, `render.DrawTarget`, `Close()`, and `Snapshot()`. I understand the motivation, but I do not think that should be the first abstraction we commit to.

### Why I think this is too early

The current runtime already works with **two independent seams**, not one:

- input comes through `host.EventSource`,
- rendering goes to `render.DrawTarget`.

Those are already sufficient to run scenes. The proof is in the existing code:

- `runtime/js/runtime_test.go` uses only a fake event source to test interaction (`runtime/js/runtime_test.go:127-206`),
- `examples/cmd/loupe-js-demo/main.go` uses only a render target to test rendering (`examples/cmd/loupe-js-demo/main.go:17-65`).

That means a simulator can be built as **composition of small pieces**, not necessarily as a monolithic “virtual device” object.

### Better first abstraction

Split the simulator into two primary components:

1. **Input hub**
   - implements `host.EventSource`,
   - exposes `InjectButton`, `InjectTouch`, `InjectKnob` methods for the browser/server side.

2. **Frame compositor**
   - owns retained RGBA buffers for `left`, `main`, and `right`,
   - exposes named `render.DrawTarget`s,
   - tracks versions and snapshots for HTTP/websocket delivery.

A higher-level `Simulator` struct can later compose those pieces, but the internal implementation should start from the narrower seams the code already uses.

## 2.2 It places simulator code too close to `pkg/device`

The first design suggests files such as `pkg/device/virtual.go`. I think that is the wrong home for the first implementation.

Evidence:

- the README explicitly says `pkg/device` is the “active low-level device/protocol implementation” and the main public Go package boundary (`README.md:17-20`, `README.md:92-100`),
- `pkg/device` is currently about transport, message framing, listeners, profiles, and hardware drawing.

A browser simulator is not primarily a low-level device/protocol feature. It is a **runtime/demo surface**. Putting it into `pkg/device` would blur the current package story and make the public boundary less coherent.

### Better package placement

For the first implementation, I would put the simulator under a new non-public or semi-public runtime-facing package such as:

- `runtime/sim/`

or, if we want to keep it CLI-only at first:

- `cmd/loupedeck/internal/sim/`

My preference is `runtime/sim/`, because the simulator is conceptually a runtime adjunct and may later deserve reuse outside the CLI.

## 2.3 It under-specifies the output side: partial patches matter

This is the most important technical gap in the first design.

The first design talks about storing display PNGs and sending frames to the browser, but it does not fully deal with the fact that `runtime/render.Renderer.Flush()` sends both:

- full-display renders for dirty displays,
- **partial tile renders** for dirty main-display tiles.

Evidence:

- dirty displays are rendered with `target.Draw(r.renderDisplay(display), 0, 0)` (`runtime/render/visual_runtime.go:71-80`),
- dirty tiles are rendered with `target.Draw(r.renderTile(tile), rect.Min.X, rect.Min.Y)` (`runtime/render/visual_runtime.go:82-92`).

That means a naive “display name → last PNG blob” store is insufficient. The simulator must support **compositing patch draws into a retained full-display framebuffer**.

### Concrete consequence

The frame store should own full RGBA images per display, and each target draw should behave like:

```go
func (t *displayTarget) Draw(im image.Image, xoff, yoff int) {
    t.store.mu.Lock()
    defer t.store.mu.Unlock()

    dst := t.store.buffers[t.name] // full RGBA canvas for the display
    rect := image.Rect(xoff, yoff, xoff+im.Bounds().Dx(), yoff+im.Bounds().Dy())
    draw.Draw(dst, rect, im, im.Bounds().Min, draw.Src)
    t.store.version[t.name]++
    t.store.dirty[t.name] = true
}
```

This is not optional. Without it, main-display tile updates will not reconstruct a correct browser-visible full screen.

## 2.4 Its pseudocode suggests the wrong notification moment

In the first design's pseudocode, the present render callback is shown as the place where browser-side notification could happen. That is risky.

Evidence:

`present.Runtime.loop()` runs work in this order:

1. call `render(reason)` if present,
2. then call `flush()` if present (`runtime/present/runtime.go:95-117`).

So if we notify the browser from `RenderFunc`, the simulator may announce a frame change **before** the frame compositor has actually applied the new patch or image in `FlushFunc`.

### Better rule

- Use `RenderFunc` only for pre-flush bookkeeping if needed.
- Publish websocket frame/version updates **after `renderer.Flush()` completes**, inside or immediately after the flush path.

That guarantees the browser never chases a frame version that has not been materialized yet.

## 2.5 It reaches for browser-side verb UX too early

The first design wants the browser UI to include a script selector, verb picker, value entry, and a full operator cockpit immediately. That is attractive, but I think it is one scope step too far for phase 1.

The repository already has working CLI surfaces for raw scripts and annotated verbs:

- raw path bootstrap via `prepareRawScriptBootstrap()` (`cmd/loupedeck/cmds/run/command.go:266-298`),
- verb path bootstrap via `prepareVerbBootstrap()` (`cmd/loupedeck/cmds/run/command.go:300-336`),
- jsverbs execution into a live runtime via `registry.InvokeInRuntime(...)` (`cmd/loupedeck/cmds/run/command.go:321-327`, `runtime/js/jsverbs_integration_test.go:39-63`).

### Better first scope

Phase 1 should allow this:

```bash
go run ./cmd/loupedeck simulate \
  --script ./examples/js/12-documented-scene.js \
  --verb "documented configure" \
  --verb-values-json '{"default":{"title":"OPS"}}'
```

and then open a browser UI for interaction.

That already solves the user's goal of easier prototype/demo flows for JS verbs. Browser-side form generation can come later.

## 2.6 It does not put enough emphasis on the actual extraction point in the current CLI

The first design identifies the seams but does not focus enough on the most valuable place to cut the code: `runSceneSession()`.

Evidence:

`runSceneSession()` currently combines all of this in one place (`cmd/loupedeck/cmds/run/command.go:339-519`):

- real device acquisition,
- display lookup,
- `Listen()` goroutine management,
- environment creation,
- event-source attachment,
- runtime open/close,
- renderer construction,
- presentation loop startup,
- exit handling.

That function is where simulator and hardware paths are still physically welded together. If we want a browser simulator without rewriting the whole command tree, that is the highest-value extraction point.

---

## 3. Review verdict

### Summary table

| Topic | Verdict on first design | My review |
|---|---|---|
| Semantic-not-wire principle | Correct | Keep it |
| Preserve hardware path | Correct | Keep it |
| Live-first scope | Correct | Keep it |
| Generic virtual backend interface | Too broad | Delay / avoid as first abstraction |
| Put simulator in `pkg/device` | Wrong package boundary | Move to `runtime/sim` or CLI-internal sim package |
| Frame snapshot model | Under-specified | Must support patch compositing |
| Browser verb cockpit in phase 1 | Too ambitious | Defer form generation to phase 2 |
| Notification in `RenderFunc` | Risky | Publish after flush |
| Main extraction point | Under-emphasized | Focus on `runSceneSession()` and bootstrap reuse |

### Bottom line

The first design is a **good strategy memo**, but not yet the lowest-risk implementation plan. I would keep its big idea and replace its implementation shape.

---

## Part II — My own proposal, analysis, and implementation plan

## 4. Problem statement

We want a browser-visible virtual Loupedeck that makes JS verbs easier to prototype and demo **without** requiring serial hardware, while preserving the current runtime behavior and without destabilizing the hardware runner.

The shortest credible path is:

- reuse current script/bootstrap logic,
- reuse current JS runtime and retained UI,
- reuse current presenter and renderer,
- add a simulator-specific input hub and frame compositor,
- add a small browser bridge,
- keep the real `run` command intact.

## 5. Revised design goals

### Must-have goals

1. Run the existing plain-script and jsverb paths without hardware.
2. Display the current retained UI in a browser.
3. Inject button, touch, and knob events from the browser into the real JS runtime.
4. Keep the implementation incremental and easy to validate.

### Non-goals for phase 1

1. Perfect emulation of the serial transport.
2. Full browser-side script discovery and form generation.
3. A stable public Go simulator API in `pkg/device`.
4. Support for every profile on day one.

---

## 6. Revised architecture

## 6.1 High-level design

```text
                  ┌───────────────────────────┐
                  │   loupedeck simulate      │
                  │  - parses script/verb     │
                  │  - opens runtime          │
                  │  - starts HTTP/WS server  │
                  └─────────────┬─────────────┘
                                │
                                ▼
                  ┌───────────────────────────┐
                  │       runtime/sim         │
                  │                           │
                  │  InputHub      BrowserHub │
                  │     │               ▲     │
                  │     ▼               │     │
                  │  host.EventSource   │     │
                  │                     │     │
                  │  FrameCompositor ───┘     │
                  │     ▲                     │
                  │     │ named DrawTargets   │
                  └─────┼─────────────────────┘
                        │
                        ▼
             ┌───────────────────────────┐
             │ runtime/ui + render + js  │
             │ retained scene execution  │
             └───────────────────────────┘
```

## 6.2 Main idea

Do **not** model the simulator as a fake transport driver first.

Instead, create four smaller pieces:

### A. `InputHub`

Responsibilities:

- implement `host.EventSource`,
- store button/touch/knob subscribers,
- allow browser/server code to inject semantic events.

Sketch:

```go
type InputHub struct {
    // maps of listeners similar to fakeSource + device listener style
}

func (h *InputHub) OnButton(device.Button, device.ButtonFunc) device.Subscription
func (h *InputHub) OnTouch(device.TouchButton, device.TouchFunc) device.Subscription
func (h *InputHub) OnKnob(device.Knob, device.KnobFunc) device.Subscription

func (h *InputHub) InjectButton(device.Button, device.ButtonStatus)
func (h *InputHub) InjectTouch(device.TouchButton, device.ButtonStatus, uint16, uint16)
func (h *InputHub) InjectKnob(device.Knob, int)
```

### B. `FrameCompositor`

Responsibilities:

- own a full retained RGBA framebuffer per display,
- expose a named `render.DrawTarget` for each display,
- apply partial patch draws at `xoff,yoff`,
- track display versions and last update timestamps,
- export full PNGs or snapshots to the browser bridge.

Sketch:

```go
type FrameCompositor struct {
    profile  device.DeviceProfile
    buffers  map[string]*image.RGBA
    versions map[string]uint64
    dirty    map[string]bool
}

func (c *FrameCompositor) Target(name string) render.DrawTarget
func (c *FrameCompositor) PNG(name string) ([]byte, error)
func (c *FrameCompositor) SnapshotMeta() SnapshotMeta
func (c *FrameCompositor) DirtyVersions() map[string]uint64
```

### C. `BrowserHub`

Responsibilities:

- serve `index.html`, JS, and CSS,
- expose `/api/snapshot` and `/api/display/{name}.png`,
- maintain websocket clients,
- translate incoming browser events into `InputHub.Inject...` calls,
- publish version updates after flushes.

### D. `simulate` command

Responsibilities:

- reuse existing raw/verb bootstrap logic,
- open the runtime,
- attach `InputHub`,
- create renderer over `FrameCompositor` targets,
- start the browser bridge,
- optionally open the browser.

---

## 7. Why this revised shape is better

## 7.1 It matches existing code better

The codebase already proves input and render can be separated:

- fake input source in tests (`runtime/js/runtime_test.go:21-69`),
- custom render target in PNG demo (`examples/cmd/loupe-js-demo/main.go:17-65`).

So we should build the simulator as those two ideas combined, not as a fake hardware device clone.

## 7.2 It keeps `pkg/device` clean

The README's package story remains intact:

- `pkg/device` stays the low-level protocol boundary,
- simulator code lives under runtime/CLI-facing packages,
- future users of the repo still understand what `pkg/device` means.

## 7.3 It gets to a working browser faster

A minimal browser bridge plus versioned PNG fetches is far less work than a full browser-side command-discovery UI.

## 7.4 It minimizes architectural regret

If we later discover a strong need for a shared “backend” abstraction, we can introduce it after the simulator works. It is much easier to generalize a successful concrete design than to make a speculative abstraction concrete.

---

## 8. Browser transport proposal

I recommend a **hybrid HTTP + websocket** transport.

### Why not send PNG blobs directly over websocket first?

Because the simulator already needs:

- HTML/JS/CSS asset serving,
- snapshot JSON,
- browser event handling.

Once an HTTP server exists anyway, static PNG fetches become simpler to debug and inspect.

### Proposed protocol

#### HTTP

- `GET /` → simulator UI
- `GET /api/snapshot` → JSON metadata
- `GET /api/display/main.png?v=12` → current full display PNG
- `GET /api/display/left.png?v=7`
- `GET /api/display/right.png?v=7`

#### Websocket

Server → browser:

```json
{ "type": "hello", "profile": "0004", "displays": ["left","main","right"] }
{ "type": "frames", "versions": { "main": 12, "left": 4, "right": 4 } }
{ "type": "log", "level": "info", "message": "scene booted" }
{ "type": "error", "message": "..." }
```

Browser → server:

```json
{ "type": "button", "name": "Circle", "status": "down" }
{ "type": "button", "name": "Circle", "status": "up" }
{ "type": "touch", "name": "Touch6", "status": "down", "x": 140, "y": 100 }
{ "type": "knob", "name": "Knob2", "delta": 1 }
```

### Why this is a good first phase

- the browser can redraw by fetching only displays whose version changed,
- debugging is easy with a browser network tab,
- the server-side frame store remains the single source of truth.

---

## 9. Revised implementation plan

## Phase 1 — Extract shared scene bootstrap helpers

### Goal

Reuse current raw-script and jsverb startup logic from the simulator without duplicating it.

### Evidence

The existing `run` command already has the right split for bootstrap preparation:

- `prepareRawScriptBootstrap()` (`cmd/loupedeck/cmds/run/command.go:266-298`)
- `prepareVerbBootstrap()` (`cmd/loupedeck/cmds/run/command.go:300-336`)

### Work

Move these helpers, or copies of them, into a shared internal package such as:

- `cmd/loupedeck/cmds/common/scene_bootstrap.go`

Exported helpers could look like:

```go
func PrepareRawScriptBootstrap(scriptPath string) ([]engine.Option, RuntimeBootstrap, error)
func PrepareVerbBootstrap(opts VerbBootstrapOptions) ([]engine.Option, RuntimeBootstrap, error)
```

### Result

Both `run` and `simulate` can use identical scene loading behavior.

---

## Phase 2 — Implement `runtime/sim/input.go`

### Goal

Create the semantic event injection layer.

### Work

Use the same subscription style as:

- `pkg/device/listeners.go` for listener maps and cleanup,
- `runtime/js/runtime_test.go`'s `fakeSource` for minimal semantics.

### API sketch

```go
type InputHub struct { ... }

func NewInputHub() *InputHub
func (h *InputHub) OnButton(...)
func (h *InputHub) OnTouch(...)
func (h *InputHub) OnKnob(...)

func (h *InputHub) InjectButton(name string, status string) error
func (h *InputHub) InjectTouch(name string, status string, x, y uint16) error
func (h *InputHub) InjectKnob(name string, delta int) error
```

### Tests

- multiple listeners receive events,
- unsubscribe works,
- name parsing maps to existing `device.ParseButton`, `device.ParseTouchButton`, `device.ParseKnob`.

---

## Phase 3 — Implement `runtime/sim/frame_compositor.go`

### Goal

Capture full browser-visible display state while accepting partial patch draws from the renderer.

### Work

- allocate full RGBA buffers using the selected profile's display sizes,
- expose one target per display name,
- on each `Draw(im, xoff, yoff)`, composite the patch into the full display buffer,
- bump version counters,
- encode PNG lazily or eagerly depending on simplicity.

### Critical implementation detail

Because `render.Renderer.Flush()` emits both full-display and tile-patch draws (`runtime/render/visual_runtime.go:71-92`), this compositor must preserve full display state across calls.

### Tests

- drawing a 90×90 tile patch updates only that rectangle on the main display,
- full-display redraw replaces the whole display image,
- version bumps only for changed displays.

---

## Phase 4 — Implement `runtime/sim/server.go`

### Goal

Expose the simulator state to a browser.

### Work

- serve embedded static assets,
- serve snapshot JSON,
- serve display PNG endpoints,
- manage websocket clients,
- translate browser events into `InputHub` injections,
- publish frame-version updates after flushes.

### Important rule

Publish `frames` events after compositor updates are durable. That means after `renderer.Flush()`, not before.

---

## Phase 5 — Add `cmd/loupedeck/cmds/simulate/command.go`

### Goal

Create an end-user simulator command without disturbing the hardware `run` command.

### Suggested flags

- `--script`
- `--verb`
- `--verb-config`
- `--verb-values-json`
- `--listen 127.0.0.1:0`
- `--open-browser`
- `--profile live`
- `--duration`

### Runtime setup pseudocode

```go
func runSimulate(ctx context.Context, opts options) error {
    runtimeOptions, bootstrap, err := common.PrepareSceneBootstrap(...)
    if err != nil { return err }

    env := envpkg.Ensure(&envpkg.LoupeDeckEnvironment{Metrics: metrics.NewWithTraceLimit(opts.TraceLimit)})

    input := sim.NewInputHub()
    env.Host.Attach(input)

    compositor := sim.NewFrameCompositor(deviceProfiles["0004"])
    renderer := render.NewWithDisplays(env.UI, compositor.Targets())

    server := sim.NewServer(input, compositor, env)
    if err := server.Start(...); err != nil { return err }
    defer server.Close()

    rt, err := jsruntime.OpenRuntime(ctx, env, runtimeOptions...)
    if err != nil { return err }
    defer rt.Close(ctx)

    if err := bootstrap(rt.Context(), rt); err != nil { return err }

    env.Present.SetRenderFunc(func(reason string) error { return nil })
    env.Present.SetFlushFunc(func() (int, error) {
        n := renderer.Flush()
        server.BroadcastDirty(compositor.DirtyVersions())
        return n, nil
    })
    env.Present.Start(rt.Context())
    defer env.Present.Close()

    return server.Wait(ctx)
}
```

Note that `BroadcastDirty(...)` happens after `renderer.Flush()`.

---

## Phase 6 — Add minimal browser UI

### Goal

Ship a working demo surface, not a perfect product UI.

### Browser features for phase 1

- canvases for left / main / right,
- buttons for Circle and Button1..Button7,
- click overlay for Touch1..Touch12,
- simple knob controls (`-` / `+` buttons or mouse wheel areas),
- current script/verb name label,
- log pane.

### Deferred phase-2 features

- auto-generated jsverb forms,
- script repository picker,
- profile switching in the UI,
- frame timeline and replay controls.

---

## 10. File-by-file implementation guide

### New files

- `loupedeck/runtime/sim/input.go`
- `loupedeck/runtime/sim/input_test.go`
- `loupedeck/runtime/sim/frame_compositor.go`
- `loupedeck/runtime/sim/frame_compositor_test.go`
- `loupedeck/runtime/sim/server.go`
- `loupedeck/runtime/sim/server_test.go`
- `loupedeck/cmd/loupedeck/cmds/simulate/command.go`
- `loupedeck/cmd/loupedeck/cmds/common/scene_bootstrap.go`
- `loupedeck/web/simulator/index.html`
- `loupedeck/web/simulator/app.js`
- `loupedeck/web/simulator/styles.css`

### Existing files to touch lightly

- `loupedeck/cmd/loupedeck/main.go` — register the new command.
- `loupedeck/docs/help/...` — document the simulator once it works.

### Files I would avoid changing in phase 1

- `loupedeck/pkg/device/connect.go`
- `loupedeck/pkg/device/dialer.go`
- `loupedeck/pkg/device/message.go`
- `loupedeck/pkg/device/display.go`

Those are transport/hardware code. The simulator does not need them for the first useful version.

---

## 11. Testing strategy

## 11.1 Unit tests

### Input hub

- listener registration and cleanup,
- semantic event injection,
- correct callback fan-out.

### Frame compositor

- tile-patch composition,
- full display replacement,
- version increments,
- PNG export.

## 11.2 Integration tests

### Plain script path

- boot a simple scene,
- ensure compositor receives tile updates,
- confirm snapshot or PNG reflects expected text/icon changes.

### jsverb path

Use the same example as `runtime/js/jsverbs_integration_test.go` and prove that the simulator path can invoke the verb and publish the resulting UI state.

## 11.3 Browser/server tests

- websocket connects,
- snapshot endpoint works,
- button/touch/knob websocket events reach the input hub,
- dirty frame notifications fire only after flush.

Browser automation can come later. The first pass can stop at server-level integration tests if time is tight.

---

## 12. Risks, tradeoffs, and mitigations

### Risk: phase-1 UI feels too bare-bones

Mitigation:

- solve the real user pain first: no hardware dependency and interactive controls,
- add better operator UX in phase 2.

### Risk: command duplication with `run`

Mitigation:

- share bootstrap helpers,
- keep `simulate` narrowly focused on browser-backed execution,
- do not try to unify every flag and metric immediately.

### Risk: we later want a public generic backend API

Mitigation:

- do not pre-commit to that abstraction now,
- let the first simulator implementation reveal which surfaces truly want to be shared.

---

## 13. Alternatives considered in my revised plan

### Alternative A: modify `run` to support both hardware and browser in one command immediately

I do not recommend this as the first step. `runSceneSession()` is already busy and strongly hardware-shaped (`cmd/loupedeck/cmds/run/command.go:339-519`). A dedicated `simulate` command keeps risk lower.

### Alternative B: use only browser-side JS and skip the Go runtime

Rejected for the same reason as in the first design: it would create behavior drift and would no longer be a faithful prototype path.

### Alternative C: implement simulator inside `pkg/device`

Rejected because it muddies the package boundary defined in the README.

---

## 14. Recommended execution order

If I were handing this to an intern, I would tell them to do the work in this exact order:

1. Extract shared scene bootstrap helpers from `run`.
2. Implement `runtime/sim/input.go` and its tests.
3. Implement `runtime/sim/frame_compositor.go` and its tests.
4. Write one integration test that boots `examples/js/01-hello.js` into the compositor.
5. Add a tiny HTTP server exposing snapshot + PNG endpoints.
6. Add websocket event injection for button clicks.
7. Add the `simulate` command.
8. Only then add browser polish and jsverb browser-side forms.

That order gives fast feedback and keeps architecture grounded in working code.

---

## 15. Final recommendation

Keep the first design's central insight — **simulate semantics, not transport** — but replace the implementation shape with a smaller, sharper, more incremental plan.

The best first implementation is **not** a virtual `pkg/device.Loupedeck`. It is:

- a simulator command,
- a semantic input hub,
- a frame compositor that understands patch draws,
- a browser bridge,
- and strict reuse of the current scene/bootstrap/runtime path.

That will produce a working browser simulator faster, preserve package clarity, and create a better base for later features such as browser-side verb forms and richer demo tooling.

## References

### Previous design being reviewed

- `design-doc/01-virtual-loupedeck-simulation-architecture-and-implementation-guide.md`

### Key repository evidence

- `README.md` — package boundary and hardware focus.
- `cmd/loupedeck/cmds/run/command.go` — current scene bootstrap and hardware session shape.
- `runtime/render/visual_runtime.go` — partial tile patch behavior and `DrawTarget` contract.
- `runtime/host/runtime.go` — `EventSource` contract.
- `runtime/present/runtime.go` — render-before-flush ordering.
- `runtime/js/runtime_test.go` — fake input source pattern and runtime-only tests.
- `runtime/js/jsverbs_integration_test.go` — verb execution without hardware.
- `examples/cmd/loupe-js-demo/main.go` — non-hardware draw target precedent.
