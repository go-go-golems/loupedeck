---
Title: Virtual Loupedeck simulation architecture and implementation guide
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
      Note: Current repo status
    - Path: loupedeck/cmd/loupedeck/main.go
      Note: Root CLI command tree that the simulator command would join
    - Path: loupedeck/pkg/device/connect.go
      Note: Serial/websocket hardware connection path that the virtual backend should bypass
    - Path: loupedeck/pkg/scriptmeta/scriptmeta.go
      Note: Script and verb discovery helper reused by the simulator load flow
    - Path: loupedeck/runtime/host/runtime.go
      Note: Semantic input-source seam for simulated button/touch/knob events
    - Path: loupedeck/runtime/render/visual_runtime.go
      Note: DrawTarget seam for feeding browser-visible frames from retained UI state
ExternalSources: []
Summary: Design for a browser-driven virtual Loupedeck backend that replaces the serial dependency for prototyping and demos while reusing the existing JS runtime, retained UI, and rendering layers.
LastUpdated: 2026-04-18T12:45:39-04:00
WhatFor: Explain how to add a virtual device backend and browser simulator so JS verbs can be prototyped, demonstrated, and tested without a physical Loupedeck attached.
WhenToUse: Read before adding a simulator command, extracting device backend interfaces, wiring a browser UI, or deciding how the runtime should behave when no serial device is present.
---


# Virtual Loupedeck simulation architecture and implementation guide

## Executive Summary

This ticket proposes a browser-driven **virtual Loupedeck** that lets the existing JavaScript scene runtime run without a physical device attached. Instead of talking to a USB serial port and the firmware 2.x serial-WebSocket protocol, the simulator would keep the same retained UI/runtime model in memory, expose simulated buttons/knobs/touch input, and render the current screen state into a web UI.

The important design idea is: **simulate the semantics, not the wire protocol**. The existing codebase already has the right semantic layers for this: `runtime/ui` stores pages, tiles, and display state; `runtime/host` attaches button/touch/knob sources; `runtime/render` converts retained UI into images; and `runtime/js` exposes those systems to JavaScript. The missing piece is a device backend that is not tied to `go.bug.st/serial` and `websocket` handshake logic.

The payoff is a much better loop for prototyping and demos:

- run a JS verb from the browser,
- click or drag simulated controls,
- see the resulting page state and rendered output immediately,
- share a demo without needing hardware on the desk.

The proposed implementation is intentionally conservative: keep the real hardware path in place, add a separate virtual backend, and connect both to the same JS/runtime/UI stack. That gives us a simulator that is useful immediately and does not force the serial implementation to become browser-aware.

## Problem Statement

The repository is still centered on real hardware. The top-level README says the project is a Go library and CLI for talking directly to Loupedeck hardware over the firmware 2.x serial-WebSocket protocol, and the supported surfaces are the `cmd/loupedeck` CLI and the `pkg/device` package (`README.md:1-26`). The quick-start path also assumes a connected device and a serial path such as `--device /dev/ttyACM0` (`README.md:36-47`). That is fine for release work, but it creates friction for prototyping, demos, and onboarding.

The user request here is more specific than “make the code easier to test.” The request is to **replace reliance on the serial interface** with a virtual Loupedeck and then expose that virtual device through a **web UI** so JS verbs can be demonstrated interactively. That implies three concrete goals:

1. scenes and verbs must still run through the real JS runtime,
2. the user must be able to inject Loupedeck-like input without USB hardware,
3. the browser must show a believable representation of the device state and rendered UI.

The simulator should therefore be a product-facing tool, not just a unit-test helper. It should support the same authoring model that the repository already uses for real scenes, especially the annotated JS verb flow shown in `examples/js/12-documented-scene.js` and documented in `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md`.

## Current-State Analysis

### 1. The real device path is still tightly coupled to serial transport

The low-level device package is implemented around a concrete `Loupedeck` struct that owns serial, websocket, writer, renderer, listeners, and display objects (`pkg/device/loupedeck.go:16-59`). Connection setup happens through `ConnectAuto` / `ConnectPath`, which always end up opening a `SerialWebSockConn` and dialing a websocket over that serial connection (`pkg/device/connect.go:13-52`, `pkg/device/connect.go:74-217`). The serial transport itself is backed by `go.bug.st/serial` enumeration and open calls (`pkg/device/dialer.go:13-164`).

That means the current device stack assumes:

- USB port discovery,
- serial port open/close,
- websocket framing over that serial port,
- firmware-specific reset/version/serial handshakes.

The `doConnect` path makes this especially clear: it opens a websocket connection, resolves the profile from the detected product ID, constructs the device object, installs the outbound writer and render scheduler, then sends reset/brightness/version/serial messages (`pkg/device/connect.go:103-217`). There is no alternative path that says “pretend to be a Loupedeck.”

### 2. The device message model is low-level and hardware-shaped

Display drawing on the real device is implemented as protocol messages. `Display.Draw` converts an image into RGB565 bytes, emits a `WriteFramebuff` message, then emits a `Draw` message (`pkg/device/display.go:81-157`). Incoming events are decoded in `Listen`, which parses binary messages and dispatches button, knob, and touch callbacks (`pkg/device/listen.go:1-95`). Listener registration and cleanup are handled inside `pkg/device/listeners.go` (`pkg/device/listeners.go:1-189`).

This is all correct for hardware, but it is the wrong abstraction level for a browser simulator. A web UI should not need to know about transaction IDs, protocol message types, or the order in which `WriteFramebuff` and `Draw` messages must be sent.

### 3. The runtime/UI layer is already a good semantic boundary

The retained UI stack already gives us most of the semantics we need for a simulator. `runtime/ui.UI` tracks pages, active page, dirty tiles, and dirty displays (`runtime/ui/ui.go:16-209`). `runtime/ui.Page` manages displays and tiles (`runtime/ui/page.go:1-51`), and `runtime/ui.Display` / `runtime/ui.Tile` store text, icons, visibility, and surfaces (`runtime/ui/display.go:1-320`, `runtime/ui/tile.go:1-89`).

The rendering layer turns that retained state into images. `runtime/render.Renderer` consumes a `ui.UI`, asks for dirty displays and tiles, and draws them to a `DrawTarget` (`runtime/render/visual_runtime.go:14-95`). `DrawTarget` is already a clean interface:

```go
type DrawTarget interface {
    Draw(im image.Image, xoff, yoff int)
}
```

That is an excellent seam for a virtual device. A browser simulator can receive exactly the same images the hardware renderer would produce, without needing the serial wire format.

### 4. The host runtime already abstracts input sources

`runtime/host.Runtime` is the event side of the same seam. It defines an `EventSource` interface with `OnButton`, `OnTouch`, and `OnKnob` (`runtime/host/runtime.go:10-74`). The runtime can `Attach` a source, register subscriptions, and later `Show` pages or `ReplayActivePage` (`runtime/host/runtime.go:49-118`, `runtime/host/pages.go:1-25`).

This is the other key fact that makes a virtual device practical: the JS runtime does not need to know whether button events came from a USB controller or a browser canvas. It only needs an `EventSource`.

### 5. The JS runtime is already wired through the retained UI and host layers

The environment object created by `runtime/js/env.Ensure` bundles `Reactive`, `UI`, `Host`, `Anim`, `Present`, and `Metrics` into one coherent runtime context (`runtime/js/env/env.go:13-68`). The `UI` dirty handler is already wired to `Present.Invalidate("ui-dirty")`, which means retained UI changes can wake the presentation loop automatically (`runtime/js/env/env.go:59-63`).

The JS modules expose the exact APIs a simulator needs to exercise:

- `loupedeck/ui` for pages, tiles, displays, and event hooks (`runtime/js/module_ui/module.go:20-320`),
- `loupedeck/state` for signals, computed values, batch, and watch (`runtime/js/module_state/module.go:13-91`),
- `loupedeck/anim` for timing/tweening (`runtime/js/module_anim/module.go:14-140`),
- `loupedeck/present` for invalidation and frame callbacks (`runtime/js/module_present/module.go:14-47`).

The module API is already browser-friendly in the semantic sense: JS code can create pages and tiles, register callbacks, and mutate retained state without caring about transport.

### 6. The CLI already has a strong loading and metadata story

The current top-level CLI wires `run`, `verbs`, and `doc` in `cmd/loupedeck/main.go` (`cmd/loupedeck/main.go:17-41`). The `run` command is currently hardware-oriented and accepts `--script`, `--verb`, `--device`, pacing flags, and tracing flags (`cmd/loupedeck/cmds/run/command.go:100-240`). The `verbs` command currently only inspects metadata (`list` and `help`) (`cmd/loupedeck/cmds/verbs/command.go:24-106`).

That matters because the simulator should reuse the same script loading semantics rather than inventing a new one. `pkg/scriptmeta.ResolveTarget` already handles file-or-directory script roots and shorthand resolution (`pkg/scriptmeta/scriptmeta.go:31-57`), `ScanVerbRegistry` discovers annotated verbs, and `EngineOptionsForTarget` sets up module roots and require loaders for scripts or directories (`pkg/scriptmeta/scriptmeta.go:87-179`).

The existing examples also help explain the target user experience. `examples/cmd/loupe-js-demo/main.go` already renders a JS scene into PNG files using the retained runtime and a `DrawTarget`-style output sink. `examples/cmd/loupe-svg-buttons/main.go` shows a real-time interactive loop with button/touch events, animated rendering, and control handling. The simulator should feel closer to the second example, but without the hardware dependency.

## Gap Analysis

The current codebase is close, but not there yet.

### What we already have

- A retained UI model that can represent pages, tiles, and displays.
- A host runtime that can attach an input source and dispatch callbacks.
- A renderer that can produce device-sized images from retained state.
- A JS runtime with modules for UI, state, animation, and presentation.
- Script metadata helpers for plain scripts and annotated verbs.

### What is missing

1. **A virtual device backend**
   - There is no in-memory equivalent to `pkg/device.Loupedeck`.
   - There is no object that simultaneously acts as an event source and a render sink without touching serial transport.

2. **A browser communication protocol**
   - No HTTP server or websocket channel exists for streaming rendered frames or accepting input events from a browser.
   - No client-side state model exists for the active page, display contents, button states, or knob values.

3. **A dedicated simulator CLI surface**
   - `run` still assumes real hardware and currently mixes raw-script execution with verb execution.
   - `verbs` is still inspection-only, not a first-class execution surface for an interactive simulator.

4. **A UI for device interaction**
   - There is no browser-visible control grid, touch pad, or knob widget.
   - There is no way to click a simulated Circle button, rotate a dial, or touch a screen region from the web.

5. **A shared snapshot model**
   - There is no canonical “what the simulator looks like right now” object that can be sent to the browser or used in tests.

The main architectural opportunity is that none of these gaps require rewriting the JS runtime or the retained UI model. They require a new seam at the device boundary.

## Proposed Solution

### Overview

Create a new virtual device subsystem that sits between the JS/runtime layer and the browser UI.

At a high level:

- the JS runtime keeps using `runtime/js`, `runtime/ui`, `runtime/host`, `runtime/render`, and `runtime/present`,
- the virtual deck implements the `EventSource` role that `host.Runtime.Attach` expects,
- the renderer pushes images into the virtual deck instead of into a hardware websocket,
- the browser UI reads those images and sends simulated input events back to the server.

### Architecture Diagram

```text
                  ┌──────────────────────────────┐
                  │        Browser UI           │
                  │  - canvases / controls       │
                  │  - verb picker / logs        │
                  │  - button / knob / touch UI  │
                  └──────────────┬───────────────┘
                                 │ websocket + HTTP
                                 ▼
                  ┌──────────────────────────────┐
                  │   Simulator Server / CLI      │
                  │  - loads script / verb        │
                  │  - owns runtime lifecycle     │
                  │  - serves snapshots / frames  │
                  └──────────────┬───────────────┘
                                 │
                 simulated input  │   rendered images
                                 ▼
                  ┌──────────────────────────────┐
                  │     Virtual Loupedeck         │
                  │  - EventSource                │
                  │  - DrawTarget / frame store   │
                  │  - profile / display layout   │
                  └──────────────┬───────────────┘
                                 │
                                 ▼
                  ┌──────────────────────────────┐
                  │  runtime/host + runtime/ui    │
                  │  retained pages / tiles       │
                  │  button / touch / knob hooks  │
                  └──────────────┬───────────────┘
                                 │
                                 ▼
                  ┌──────────────────────────────┐
                  │   runtime/js + JS modules     │
                  │   scene code / verbs / docs   │
                  └──────────────────────────────┘
```

### Core Design Principle

The simulator should model **the device that scene authors think they are using**, not the USB protocol that the hardware happens to speak.

That means the simulator must expose:

- page show/hide state,
- tile text/icon/surface updates,
- display text/icon/surface updates,
- button/touch/knob events,
- a visible active page,
- rendered output for the browser.

It does **not** need to expose raw transaction IDs, protocol message types, or serial port retries.

### Proposed Internal API Shape

A practical first cut is a small interface around the semantic device boundary:

```go
type Backend interface {
    host.EventSource
    render.DrawTarget
    Close() error
    Snapshot() Snapshot
}

type Snapshot struct {
    Profile      device.DeviceProfile
    ActivePage   string
    Displays     map[string]DisplaySnapshot
    Buttons      map[string]ButtonSnapshot
    Knobs        map[string]KnobSnapshot
    LastRenderAt time.Time
}

type DisplaySnapshot struct {
    Name      string
    Width     int
    Height    int
    Visible   bool
    ImagePNG  []byte
}
```

The first implementation does not need to be perfect or complete. It needs to be stable enough that the browser can render the current state and the tests can assert against it.

### Proposed Browser Protocol

Use a simple websocket protocol first:

- **Server → browser**
  - `snapshot` — complete state summary when a client connects or reloads,
  - `frame` — one or more display images have changed,
  - `log` — runtime or simulator logs,
  - `error` — script/verb or runtime failures,
  - `metrics` — optional stats and counters.

- **Browser → server**
  - `buttonDown` / `buttonUp`,
  - `touchDown` / `touchMove` / `touchUp`,
  - `knobTurn`,
  - `showPage`,
  - `reloadScene`,
  - `invokeVerb`.

A minimal JSON shape is fine for the first version. If frame traffic becomes expensive, the transport can later switch from full PNG snapshots to region diffs, but that should not block the first usable simulator.

### Proposed Runtime Wiring

The runtime wiring should look like this:

```go
func RunVirtualScene(ctx context.Context, opts Options) error {
    target, registry, err := scriptmeta.ScanVerbRegistry(opts.ScriptPath)
    if err != nil {
        return err
    }

    env := envpkg.Ensure(nil)
    deck := virtual.New(opts.Profile)
    env.Host.Attach(deck)

    renderer := render.NewWithDisplays(env.UI, deck.DrawTargets())
    env.Present.SetRenderFunc(func(reason string) error {
        // optional browser-side notification hook
        return nil
    })
    env.Present.SetFlushFunc(func() (int, error) {
        return renderer.Flush(), nil
    })
    env.Present.Start(ctx)

    rt, err := jsruntime.OpenRuntime(ctx, env, scriptmeta.EngineOptionsForTarget(target, registry)...)
    if err != nil {
        return err
    }
    defer rt.Close(ctx)

    if opts.Verb != "" {
        verb, err := scriptmeta.FindVerb(target, registry, opts.Verb)
        if err != nil {
            return err
        }
        desc, err := registry.CommandDescriptionForVerb(verb)
        if err != nil {
            return err
        }
        parsed, err := scriptmeta.ParseVerbValues(desc, opts.VerbConfig, opts.VerbValuesJSON)
        if err != nil {
            return err
        }
        _, err = registry.InvokeInRuntime(ctx, rt.Runtime, verb, parsed)
        return err
    }

    _, err = rt.RunString(ctx, opts.ScriptSource)
    return err
}
```

This pseudocode shows the shape we want:

- the scene loader stays shared,
- the JS runtime stays shared,
- the host and UI stay shared,
- only the backend changes.

### How the Browser UI Should Behave

The browser UI should be the operator-facing cockpit for a virtual device. A good first version should include:

- a **main display canvas** showing the 4×3 tile area,
- left and right display previews for the current product profile,
- a **button strip** for Circle and other hardware buttons,
- a **touch overlay** for the touch regions,
- controls for one or more knobs,
- a script/verb selector,
- a log pane,
- a small state inspector showing the active page and current display labels.

For JS verbs specifically, the web UI should make it easy to:

1. choose a script file or scene directory,
2. choose an annotated verb,
3. enter verb values/configuration,
4. start the scene,
5. interact with the simulated controls,
6. observe the rendered result.

That flow is what makes the simulator valuable for demos.

## Design Decisions

### 1. Simulate semantics, not the serial wire protocol

This is the most important decision in the design.

Why:

- The serial protocol is only useful if the goal is hardware compatibility.
- The browser UI needs stable semantic state, not transaction bytes.
- Reimplementing the wire protocol would add complexity without improving the prototype experience.

Implication:

- The virtual backend should render to `image.Image` / PNGs and expose input events directly at the button/touch/knob level.
- The real hardware path can continue using `WriteFramebuff` / `Draw` without changes.

### 2. Reuse the retained UI stack as the source of truth

`runtime/ui.UI` already stores pages, displays, tiles, and dirty regions (`runtime/ui/ui.go:16-209`). That makes it the correct shared model for both real hardware and simulator UI.

Implication:

- The simulator should not invent its own parallel “scene state” model.
- The browser should display the retained UI as it exists in `runtime/ui`, plus any extra metadata needed for controls and logs.

### 3. Use `host.EventSource` and `render.DrawTarget` as the main seams

The two existing abstractions are already in the right shape:

- input source: `runtime/host.EventSource` (`runtime/host/runtime.go:10-14`),
- output sink: `runtime/render.DrawTarget` (`runtime/render/visual_runtime.go:23-25`).

Implication:

- The virtual deck should implement both roles.
- The rest of the runtime should remain unaware of whether a physical or virtual device is attached.

### 4. Default to the Loupedeck Live profile first

The repository already treats Loupedeck Live (`product 0004`) as the main hardware focus (`README.md:22-26`). `pkg/device/profile.go` also encodes the Live display layout directly (`pkg/device/profile.go:45-53`).

Implication:

- The first simulator profile should be Live.
- CT and Live S support can come later because the profile data already exists.

### 5. Keep the real device path unchanged initially

The real `pkg/device` stack is already working and is still needed for hardware validation.

Implication:

- The simulator should be additive.
- Do not delete or rewrite the serial path just to support the browser.
- If a small seam needs to be extracted, keep it narrow and local.

### 6. Use full-frame PNG transfer first

The images are small enough that full-frame PNG snapshots are the simplest transport for the first browser UI.

Implication:

- Start with correctness and simplicity.
- Optimize later only if profile switching or animation makes the websocket too chatty.

## Alternatives Considered

### Alternative 1: Proxy the serial protocol into the browser

This would mean emulating the Loupedeck transport byte-for-byte and teaching the browser about websocket/serial messages.

Why it is not the right first step:

- it would entangle the simulator with firmware-specific implementation details,
- it would not make the UI easier for prototype/demo work,
- it would duplicate logic already present in `pkg/device`.

### Alternative 2: Build a browser-only mock of the JS APIs

This would mean reimplementing `loupedeck/ui`, `loupedeck/state`, `loupedeck/anim`, and `loupedeck/present` separately for the browser.

Why it is not the right first step:

- the repository already has a runtime and retained state model that works,
- duplicating the runtime would create behavior drift,
- it would make the browser demo less faithful to hardware.

### Alternative 3: Add only a PNG preview command

The repository already has a non-interactive PNG demo (`examples/cmd/loupe-js-demo/main.go`), but that is not enough for verb prototyping.

Why it is not the right first step:

- there is no live input loop,
- there is no control surface,
- it cannot serve as a demo cockpit.

### Alternative 4: Keep only hardware demos and avoid a simulator

This is the current state, but it does not solve the user request.

Why it is not acceptable:

- it keeps demos hardware-dependent,
- it raises the cost of trying verbs or scene variations,
- it makes onboarding slower.

## Implementation Plan

### Phase 1: Extract the simulator seam

Goal: identify the minimum interface and data model that the browser UI will need.

Suggested files:

- `runtime/host/runtime.go` — confirm the EventSource boundary,
- `runtime/render/visual_runtime.go` — confirm the DrawTarget boundary,
- `runtime/ui/ui.go` / `runtime/ui/display.go` / `runtime/ui/tile.go` — confirm what state must be serialized,
- `pkg/device/profile.go` — profile-driven dimensions/layout.

Work to do:

1. define a `virtual` or `sim` package name,
2. decide what a snapshot contains,
3. decide whether the browser consumes PNG bytes or RGBA diffs,
4. decide how the simulator will identify the active profile.

Validation:

- new unit tests should describe the snapshot shape,
- no behavior should change for real hardware yet.

### Phase 2: Implement the virtual deck backend

Goal: create an in-memory device that can receive input and store rendered output.

Suggested files:

- new `pkg/device/virtual.go` or `runtime/virtual/deck.go`,
- maybe a small `runtime/virtual/snapshot.go` helper.

Work to do:

1. store a selected device profile,
2. implement `OnButton`, `OnTouch`, and `OnKnob`,
3. keep display frame buffers in memory,
4. provide a `DrawTarget` implementation for each display,
5. provide a `Snapshot()` method for browser sync and tests,
6. optionally provide `SetBrightness` and `SetButtonColor` semantics if the browser UI wants them.

Implementation note:

- button/touch/knob events should be injected at the same semantic level that `runtime/host` expects.
- for touch, reuse the same coordinate-to-region logic as the real device if possible so the browser matches the current hardware model.

Validation:

- unit tests for event subscription and callback delivery,
- tests for snapshot updates after draw calls,
- tests that the profile dimensions match the selected hardware model.

### Phase 3: Wire the virtual deck into the JS runtime path

Goal: run a real JS scene or verb against the virtual deck.

Suggested files:

- `pkg/scriptmeta/scriptmeta.go` for loading,
- `runtime/js/env/env.go` for environment setup,
- `runtime/js/registrar.go` and the JS module packages for runtime reuse,
- a new scene runner helper if needed.

Work to do:

1. load script metadata and verb definitions from the same entrypoint code used by the real runner,
2. create a JS runtime with the same modules,
3. attach the virtual deck as the host event source,
4. attach the renderer so UI changes flush into the virtual frames,
5. support raw scripts and annotated verbs.

Validation:

- the same example scripts should boot in both hardware and simulator modes,
- annotated verbs should still resolve through `scriptmeta.FindVerb` and `registry.InvokeInRuntime`.

### Phase 4: Build the browser UI and websocket protocol

Goal: expose the virtual deck in a browser.

Suggested files:

- new `cmd/loupedeck/cmds/simulate/command.go` or `cmd/loupedeck/cmds/web/command.go`,
- new `runtime/sim/server.go`,
- static HTML/JS assets under a new `web/` or `assets/` directory.

Work to do:

1. serve the initial page and assets,
2. stream snapshots/frames over websocket,
3. send browser events back to the virtual deck,
4. display the current page and frame content,
5. show logs and runtime errors.

Validation:

- browser connects and receives an initial snapshot,
- button/knob/touch clicks show up in the JS runtime,
- page and tile updates redraw in the browser.

### Phase 5: Add CLI wiring and user-facing documentation

Goal: make the simulator easy to start.

Suggested files:

- `cmd/loupedeck/main.go` — add the simulator command,
- `docs/help/...` — add help pages for simulator usage,
- `examples/js/...` — optionally add one dedicated simulator-friendly example.

Work to do:

1. add a top-level command such as `loupedeck simulate` or `loupedeck web`,
2. preserve the current `run` command for real hardware,
3. decide whether the simulator should default to opening a browser automatically,
4. document the new workflow in the help system.

Validation:

- `loupedeck --help` should show the simulator command,
- the simulator command should work with both a plain JS file and an annotated verb.

### Phase 6: Testing and hardening

Goal: make the simulator trustworthy for demos and regression testing.

Suggested tests:

- backend tests for event delivery and frame storage,
- runtime integration tests that boot an example scene in virtual mode,
- websocket/server tests for message decoding and routing,
- optional browser smoke tests with Playwright.

Validation commands to expect eventually:

```bash
go test ./... -count=1
```

plus targeted simulator tests and at least one browser smoke path.

## Testing and Validation Strategy

The simulator should be validated at three levels.

### 1. Backend tests

Focus:

- event subscription correctness,
- frame storage updates,
- snapshot consistency,
- profile dimensions and layout.

Examples:

- pressing a simulated button triggers the right callback,
- a simulated touch maps to the expected touch region,
- the virtual display buffer changes after a render flush.

### 2. Runtime integration tests

Focus:

- JS runtime boot,
- `ui.page` / `ui.tile` / `ui.show` semantics,
- event callback handling,
- annotated verb invocation,
- retained page replay.

A good integration test should look very close to `runtime/js/runtime_test.go`, but with a virtual backend attached instead of a fake source.

### 3. Browser/UI smoke tests

Focus:

- the browser can connect,
- it receives a snapshot,
- it can inject button/touch/knob events,
- the display updates reflect JS changes.

These tests are more expensive, so they should be small and targeted. The goal is not to exhaustively test every browser widget; it is to make sure the end-to-end simulator loop works.

## Risks and Tradeoffs

### Risk 1: Divergence between virtual and real behavior

If the virtual backend becomes too abstract, it may drift away from hardware semantics.

Mitigation:

- keep the device profile data shared,
- keep the same runtime and UI model,
- reuse the same touch/button naming and event types.

### Risk 2: The browser UI becomes a second product

A simulator can accidentally turn into a generic dashboard with too many controls.

Mitigation:

- keep the first UI focused on run, inspect, and interact,
- do not overinvest in presentation chrome before the semantic loop works.

### Risk 3: Websocket traffic becomes noisy

If every tiny UI change sends a full frame at high frequency, the browser can become chatty.

Mitigation:

- start with full-frame PNGs,
- rely on the existing retained renderer and dirty-state coalescing,
- optimize transport only after correctness is proven.

### Risk 4: The simulator command competes with `run`

The CLI already has a `run` command that can execute both plain scripts and verbs.

Mitigation:

- keep `run` as the hardware-facing command for now,
- make the simulator command explicitly about browser-based interaction,
- document the difference clearly.

## Open Questions

1. Should the simulator be called `simulate`, `web`, `virtual`, or something else in the final CLI?
2. Should the first implementation default to Loupedeck Live only, or should it expose a profile selector immediately?
3. Should the browser receive full PNG frames, raw RGBA, or tiled diffs?
4. Should the simulator support hot-reload of the script file, or only a restart button?
5. Should we model button LEDs and brightness controls in the first browser UI, or defer them?
6. Do we want the simulator to be headless-friendly so it can run in CI and feed tests without a browser?

## References

### Key current files

- `cmd/loupedeck/main.go` — root command wiring for `run`, `verbs`, and `doc`.
- `cmd/loupedeck/cmds/run/command.go` — current hardware runner and verb execution path.
- `cmd/loupedeck/cmds/verbs/command.go` — current metadata-only verb inspection command.
- `pkg/device/connect.go` — serial/WebSocket connection bootstrap for real hardware.
- `pkg/device/dialer.go` — USB serial discovery and opening.
- `pkg/device/display.go` — protocol-level display drawing implementation.
- `pkg/device/listeners.go` — real-device input subscriptions and dispatch.
- `pkg/device/profile.go` — device profiles and display layouts.
- `pkg/device/message.go` — low-level protocol messages and transaction IDs.
- `runtime/js/env/env.go` — JS environment construction and dirty-handler wiring.
- `runtime/js/registrar.go` — JS module registration and runtime sentinels.
- `runtime/js/module_ui/module.go` — JS UI surface for pages, tiles, displays, and events.
- `runtime/js/module_state/module.go` — reactive JS state primitives.
- `runtime/js/module_anim/module.go` — animation/tween helpers for live scenes.
- `runtime/js/module_present/module.go` — render invalidation and frame callbacks.
- `runtime/ui/ui.go` — retained UI model and dirty tracking.
- `runtime/ui/page.go` — page-level retained structure.
- `runtime/ui/display.go` — retained display structure and layers.
- `runtime/ui/tile.go` — retained tile structure.
- `runtime/host/runtime.go` — host runtime and event-source attachment.
- `runtime/host/pages.go` — page show hooks and active-page replay.
- `runtime/render/visual_runtime.go` — retained renderer and DrawTarget abstraction.
- `pkg/scriptmeta/scriptmeta.go` — script root resolution, verb scanning, and verb value parsing.
- `examples/cmd/loupe-js-demo/main.go` — PNG-based retained runtime demo.
- `examples/cmd/loupe-svg-buttons/main.go` — interactive real-hardware demo with input loop.
- `examples/js/12-documented-scene.js` — canonical annotated-verb scene example.
- `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md` — current user-facing annotated-scene guide.
- `README.md` — current repo status, release surface, and hardware focus.

### Ticket-local documents

- `design-doc/01-virtual-loupedeck-simulation-architecture-and-implementation-guide.md` — this document.
- `reference/01-investigation-diary.md` — chronological research record and continuation notes.
