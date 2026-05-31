---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: go-go-goja/pkg/xgoja/app/run.go
      Note: Evidence that built-in run command closes runtime after script
    - Path: go-go-goja/pkg/xgoja/providerapi/commands.go
      Note: Recommended CommandSetProvider recovery mechanism
    - Path: go-go-goja/pkg/xgoja/providers/http/http.go
      Note: xgoja HTTP provider lifecycle and config sections
    - Path: loupedeck-server/main.go
      Note: Manual host spike being reviewed; bypasses xgoja generation
    - Path: loupedeck-server/pkg/xgoja/serverprovider/provider.go
      Note: Implemented recovery plan as xgoja CommandSetProvider
    - Path: loupedeck-server/server.js
      Note: REST API script and current API/hardware behavior
    - Path: loupedeck-server/xgoja.yaml
      Note: Original xgoja buildspec; now stale relative to manual main.go
    - Path: loupedeck/runtime/js/env/device_control.go
      Note: Physical Loupedeck adapter for DeviceControl
    - Path: loupedeck/runtime/js/module_hw/module.go
      Note: Clean loupedeck/hw JavaScript hardware API
    - Path: loupedeck/runtime/js/module_hw/module_test.go
      Note: Unit tests for hardware API success and validation
    - Path: loupedeck/runtime/js/module_ui/module.go
      Note: Current JS API and WIP hardware-control build break
    - Path: loupedeck/runtime/js/provider/provider.go
      Note: Loupedeck xgoja provider hardware capability to reuse
    - Path: loupedeck/ttmp/2026/05/31/LDCK-API-001--loupedeck-remote-control-http-api-xgoja-binary/scripts/02-start-server.sh
      Note: Updated tmux launcher to call generated serve command
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---




# Code Review: xgoja HTTP Server Attempt and Recovery Plan

## Executive Summary

The current `loupedeck-server` work proves several important pieces are viable: `gojahttp` can serve express-style routes, `loupedeck/ui` can receive hardware button events, and SQLite-backed polling is a reasonable first integration path. However, the implementation drifted away from the original goal: it is no longer an xgoja-generated binary. `loupedeck-server/main.go` is a manually written host application that duplicates xgoja/provider responsibilities, while `loupedeck-server/xgoja.yaml` remains as a stale artifact.

That drift is not evidence that xgoja is fundamentally wrong for this task. It is evidence of a lifecycle mismatch and a documentation/API gap: xgoja's built-in `run` command is intentionally short-lived. It creates a runtime, runs a script, and closes the runtime. An HTTP server needs a long-running command that loads the script, leaves the runtime alive, and exits only on signal/context cancellation. xgoja already has most of the extension mechanism needed for this: provider-owned `CommandSetProvider`s can create long-running commands and reuse xgoja's selected modules, runtime factory, and provider configuration sections. The correct recovery path is to implement a proper xgoja command provider (or a first-class xgoja `serve` command) instead of continuing to hand-write `main.go`.

The current workspace also contains uncommitted WIP changes that break the build. Before any implementation continues, either revert the WIP or complete it in a focused branch.

---

## Current State and Evidence

### Repository / workspace state

The root workspace is not itself a Git repository. The newly created `loupedeck-server/` directory is its own Git repository with three commits:

- `ed576d5` — `feat: standalone loupedeck-server binary with REST API`
- `cf57f16` — `chore: add .gitignore, remove db from tracking`
- `2b226e4` — `feat: add debug logging for hardware event callbacks`

There are also uncommitted changes outside those commits:

- `loupedeck/runtime/js/env/env.go` — adds a new `DeviceControl` interface field to `LoupeDeckEnvironment`.
- `loupedeck/runtime/js/env/device_control.go` — new adapter for `*device.Loupedeck`.
- `loupedeck/runtime/js/module_ui/module.go` — adds `setButtonColor` and `setBrightness`, but currently fails to compile due an unused `button` variable.
- `loupedeck-server/main.go` — debug logging was added, and a local edit removed `deckConn` declaration, leaving references unresolved after the module build error is fixed.

### Build status is currently red

Command:

```bash
cd /home/manuel/workspaces/2026-05-27/better-loupedeck-tiles/loupedeck-server
go build ./...
```

Observed error:

```text
# github.com/go-go-golems/loupedeck/runtime/js/module_ui
../loupedeck/runtime/js/module_ui/module.go:134:4: declared and not used: button
```

Command:

```bash
cd /home/manuel/workspaces/2026-05-27/better-loupedeck-tiles/loupedeck
go test ./runtime/js/... ./runtime/host/... ./pkg/device/...
```

Observed error:

```text
# github.com/go-go-golems/loupedeck/runtime/js/module_ui
runtime/js/module_ui/module.go:134:4: declared and not used: button
FAIL github.com/go-go-golems/loupedeck/runtime/js [build failed]
```

This must be fixed or reverted before further runtime conclusions are trusted.

---

## Architecture Map: The System Pieces

### 1. Loupedeck hardware layer

The low-level device driver lives under `loupedeck/pkg/device/`.

Key responsibilities:

- Detect and open the serial-over-USB device.
- Perform the websocket-like handshake.
- Send display framebuffer messages.
- Dispatch button/knob/touch events.

Important files:

- `loupedeck/pkg/device/connect.go` — device detection and connection setup.
- `loupedeck/pkg/device/loupedeck.go` — core `Loupedeck` struct, `SetBrightness`, `SetButtonColor`, `Close`.
- `loupedeck/pkg/device/display.go` — `Display.Draw(image, xoff, yoff)` writes framebuffer and draw commands.
- `loupedeck/pkg/device/listen.go` — `Listen()` loop reads protocol messages and dispatches events.
- `loupedeck/pkg/device/listeners.go` — `OnButton`, `OnKnob`, `OnTouch` registration.

Hardware event path:

```text
USB serial message
  -> device.Loupedeck.Listen()
  -> dispatchButton / dispatchKnob / dispatchTouch
  -> registered device listener callbacks
```

### 2. Retained UI layer

The retained UI model lives under `loupedeck/runtime/ui/` and `loupedeck/runtime/render/`.

Important files:

- `loupedeck/runtime/ui/ui.go` — pages, active page, dirty tile/display tracking.
- `loupedeck/runtime/ui/page.go` — page-level tiles and displays.
- `loupedeck/runtime/ui/tile.go` — tile text/icon/surface state and dirty marking.
- `loupedeck/runtime/render/visual_runtime.go` — renders dirty tiles/displays to images and calls `Draw` targets.
- `loupedeck/runtime/present/runtime.go` — invalidation loop that calls a `FlushFunc` when dirty.

Display path:

```text
JavaScript: ui.page(... tile.text(...))
  -> runtime/ui Tile.SetText / markDirtyTile
  -> ui.SetDirtyHandler callback
  -> present.Runtime.Invalidate(reason)
  -> present loop calls FlushFunc
  -> render.Renderer.Flush()
  -> device.Display.Draw(image, x, y)
  -> outbound writer sends framebuffer to hardware
```

This path is conceptually sound. If the screen does not update, the fastest diagnostic is to instrument each boundary: dirty handler hit? present wake? renderer flush count? display draw calls? writer stats?

### 3. JavaScript module layer

The JS-facing Loupedeck modules are under `loupedeck/runtime/js/`.

Important files:

- `loupedeck/runtime/js/provider/provider.go` — xgoja provider registration and hardware capability.
- `loupedeck/runtime/js/module_ui/module.go` — exports `page`, `show`, `invalidate`, `onButton`, `onTouch`, `onKnob`.
- `loupedeck/runtime/js/env/env.go` — `LoupeDeckEnvironment`, shared runtime services.
- `loupedeck/runtime/js/registrar.go` — older runtime registrar for non-xgoja loupedeck runtime.

The existing provider already knows how to connect hardware in xgoja mode. In `loupedeck/runtime/js/provider/provider.go`, the hardware capability's `InitRuntimeFromSections` connects the device, attaches it to the host runtime, starts `Listen()`, builds a renderer, and starts presentation flushing. This means a correct xgoja-based server should reuse that capability instead of duplicating the connection/render setup in an application `main.go`.

### 4. HTTP / express layer

The HTTP stack lives in `go-go-goja/pkg/gojahttp/` and `go-go-goja/modules/express/`.

Important files:

- `go-go-goja/modules/express/express.go` — `require("express").app()` route registration.
- `go-go-goja/pkg/gojahttp/host.go` — `Host.ServeHTTP` routes HTTP requests into goja.
- `go-go-goja/pkg/gojahttp/route_registry.go` — route patterns such as `/api/v1/buttons/:name`.
- `go-go-goja/pkg/xgoja/providers/http/http.go` — xgoja HTTP provider: config section, host creation, server lifecycle.

HTTP request path:

```text
net/http request
  -> gojahttp.Host.ServeHTTP
  -> registry.Match(method, path)
  -> runtime owner Call("http-handler", ...)
  -> JavaScript route handler(req, res)
  -> res.json / res.send
```

Important implication: JavaScript route handlers run on the goja runtime owner. If a script blocks the owner goroutine with `while(true)`, HTTP handlers cannot run.

### 5. xgoja composition layer

xgoja is supposed to compose provider packages into a generated binary.

Important files:

- `go-go-goja/pkg/xgoja/app/root.go` — generated root command construction.
- `go-go-goja/pkg/xgoja/app/run.go` — built-in short-lived `run` command.
- `go-go-goja/pkg/xgoja/app/command_providers.go` — provider-owned command mounting.
- `go-go-goja/pkg/xgoja/providerapi/commands.go` — `CommandSetProvider`, `CommandSetContext`, `RuntimeFactory`.
- `loupedeck-server/xgoja.yaml` — proposed buildspec.

The built-in `run` command is the source of the confusion. It is not a server command.

Evidence in `go-go-goja/pkg/xgoja/app/run.go`:

- `runScriptFileWithInitializers(...)` creates a runtime.
- It defers `rt.Close(ctx)` immediately after runtime creation.
- It runs the script inside `rt.Owner.Call(...)`.
- It returns after the script finishes.

For one-shot scripts this is correct. For an HTTP server it is fatal: after `server.js` registers routes, `run` returns and the runtime close hook shuts down the HTTP server.

---

## Code Review Findings

## 1. The current binary is not xgoja-generated

### Problem

`loupedeck-server/main.go` is a manually written application. It does not use the generated `main.go` produced from `loupedeck-server/xgoja.yaml`. The resulting binary may be useful as a spike, but it no longer tests xgoja composition or validates the buildspec.

### Where to look

- `loupedeck-server/main.go`
- `loupedeck-server/xgoja.yaml`
- `go-go-goja/pkg/xgoja/app/root.go`

### Example

`loupedeck-server/main.go` manually creates a Cobra root, engine runtime, loupedeck environment, HTTP server, and native modules.

The xgoja buildspec still declares providers and runtime modules:

```yaml
packages:
  - id: loupedeck
    import: github.com/go-go-golems/loupedeck/pkg/xgoja/provider
  - id: go-go-goja-http
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/http
  - id: go-go-goja-host
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host
```

But the manually built binary does not consume this YAML.

### Why it matters

The manual app duplicates provider wiring and bypasses xgoja's declarative capability model. That creates two systems:

1. `xgoja.yaml` says one thing.
2. `main.go` does another.

A new intern will not know which one is authoritative. Future module additions must be made twice, or one path silently diverges.

### Cleanup sketch

Either delete the manual binary path or explicitly label it as a temporary spike. The recommended path is to move the long-running server lifecycle into a provider-owned xgoja command:

```text
loupedeck-server/
  xgoja.yaml                     # authoritative buildspec
  server.js                      # REST route script
  pkg/serverprovider/provider.go # registers CommandSetProvider{Name:"http-server"}
```

Then `xgoja build -f xgoja.yaml` should be the only way to build the product binary.

---

## 2. xgoja `run` was used for a long-running server, but `run` is intentionally short-lived

### Problem

The initial xgoja attempt failed because the built-in xgoja `run` command closes the runtime as soon as the script finishes. An HTTP server needs the runtime to remain alive after route registration.

### Where to look

- `go-go-goja/pkg/xgoja/app/run.go:81-120`
- `go-go-goja/pkg/xgoja/providers/http/http.go:78-156`

### Example

Relevant lifecycle in `run.go`:

```go
rt, err := factory.NewRuntime(ctx, profile, requireOpt)
if err != nil { ... }
defer func() { _ = rt.Close(ctx) }()

if vals != nil && len(selectedModules) > 0 {
    if err := initRuntimeFromSections(ctx, vals, rt, selectedModules); err != nil { ... }
}

_, err = rt.Owner.Call(ctx, "xgoja.run", func(_ context.Context, vm *goja.Runtime) (any, error) {
    return rt.Require.Require(scriptPath)
})
```

The HTTP provider registers a runtime closer in `InitRuntimeFromSections`; when `rt.Close` runs, the HTTP server is shut down.

### Why it matters

This is the central lifecycle issue. It is not enough for JavaScript to call `express.app().get(...)`. The host process must keep the runtime alive and allow the runtime owner to process HTTP callbacks.

Blocking the JavaScript script with `while(true)` is not correct either: it occupies the owner and prevents `gojahttp.Host.ServeHTTP` from scheduling route handlers.

### Cleanup sketch

Add a long-running command using xgoja's `CommandSetProvider` mechanism:

```go
func Register(reg *providerapi.Registry) error {
    return reg.Package("loupedeck-server",
        providerapi.CommandSetProvider{
            Name: "http-server",
            DefaultMount: "",
            New: newServerCommandSet,
        },
    )
}

func newServerCommandSet(ctx providerapi.CommandSetContext) (*providerapi.CommandSet, error) {
    sections := providerutil.CollectConfigSections(ctx.SelectedModules, ...)
    cmd := cmds.NewBareCommand(descWithScriptArgAndSections(sections), func(runCtx context.Context, vals *values.Values) error {
        requireOpt := engine.RequireOptionWithModuleRootsFromScript(scriptPath, engine.DefaultModuleRootsOptions())
        rt, err := ctx.RuntimeFactory.NewRuntime(runCtx, ctx.RuntimeProfile, requireOpt)
        if err != nil { return err }
        defer rt.Close(context.Background())

        err = providerutil.InitRuntimeFromSections(runCtx, vals, runtimeHandle{rt}, ctx.SelectedModules)
        if err != nil { return err }

        _, err = rt.Owner.Call(runCtx, "server.load", func(_ context.Context, vm *goja.Runtime) (any, error) {
            return rt.Require.Require(absScript)
        })
        if err != nil { return err }

        waitForSignalOrContext(runCtx)
        return nil
    })
    return &providerapi.CommandSet{Commands: []cmds.Command{cmd}}, nil
}
```

This preserves xgoja composition and fixes the lifecycle issue.

---

## 3. Current manual `main.go` duplicates provider logic and will diverge

### Problem

`loupedeck-server/main.go` manually reconstructs behavior that already exists in provider packages: hardware connection, environment creation, renderer setup, express module loading, database module loading, metadata sentinels, and cleanup.

### Where to look

- `loupedeck-server/main.go:174-294` — manual runtime/environment/server setup.
- `loupedeck-server/main.go:310-353` — manual module registration.
- `loupedeck/runtime/js/provider/provider.go:242+` — existing xgoja hardware capability.
- `go-go-goja/pkg/xgoja/providers/http/http.go:78+` — existing xgoja HTTP capability.
- `go-go-goja/pkg/xgoja/providers/host/host.go` — existing guarded database module.

### Example

Manual module registration:

```go
loupedeckenv.Store(ctx.VM, s.env)
loupedeckui.Register(reg)
loader := gojaexpress.NewLoader(s.httpHost)
reg.RegisterNativeModule("express", loader)
dbMod := modules.GetModule("database")
reg.RegisterNativeModule("database", dbMod.Loader)
```

Provider-based equivalent already exists:

```yaml
runtimes:
  server:
    modules:
      - package: go-go-goja-http
        name: express
        as: express
      - package: loupedeck
        name: loupedeck/ui
        as: loupedeck/ui
      - package: go-go-goja-host
        name: db
        as: db
        config:
          allowConfigure: true
```

### Why it matters

Manual wiring loses these xgoja properties:

- runtime profile capability isolation,
- provider config sections (`--deck-*`, `--http-*`),
- module config schema validation,
- command-provider reuse,
- generated module listing and help consistency,
- provider-owned cleanup hooks.

Any future provider change must be manually mirrored.

### Cleanup sketch

Delete manual module registration after the xgoja command provider exists. Keep `server.js` as the application script, and let xgoja own the Go module graph:

```yaml
commandProviders:
  - id: loupedeck-http-server
    package: loupedeck-server
    name: http-server
    mount: ""
    runtimeProfile: server
```

---

## 4. The display API returns success before proving anything reached hardware

### Problem

`server.js` returns `{"ok": true}` after updating in-memory retained UI state. It does not verify that dirty state flushed, that renderer produced operations, or that `Display.Draw` wrote to hardware.

### Where to look

- `loupedeck-server/server.js` display/page endpoints.
- `loupedeck/runtime/ui/ui.go:51-66` — active page and invalidation.
- `loupedeck/runtime/render/visual_runtime.go:55-86` — renderer flushes dirty displays/tiles.
- `loupedeck/runtime/present/runtime.go:53-113` — invalidate loop.

### Example

`server.js`:

```javascript
ui.page(pageName, (p) => { ... })
ui.show(pageName)
pages[pageName] = { tiles: tiles }
res.json({ ok: true, page: pageName })
```

This only means the JavaScript calls did not throw. It does not mean any pixels were sent.

### Why it matters

The user observed exactly this failure mode: `show` returns success, but the hardware screen stays blank. Without instrumentation at the dirty/render/device boundary, the API gives a false positive.

### Cleanup sketch

Add a narrow metrics/debug surface before changing architecture:

```go
// In the long-running command's flush function:
env.Present.SetFlushFunc(func() (int, error) {
    dirtyDisplays := len(env.UI.DirtyDisplays())
    dirtyTiles := len(env.UI.DirtyTiles())
    n := renderer.Flush()
    log.Info().Int("dirtyDisplays", dirtyDisplays).Int("dirtyTiles", dirtyTiles).Int("flushed", n).Msg("loupedeck flush")
    return n, nil
})
```

Add debug endpoints:

```http
GET /api/v1/debug/render
```

Response sketch:

```json
{
  "activePage": "demo",
  "dirtyTiles": 0,
  "dirtyDisplays": 0,
  "lastFlushOps": 4,
  "lastFlushAt": "...",
  "writer": { "queuedCommands": 0, "sentCommands": 10, "failedCommands": 0 }
}
```

---

## 5. Button color and brightness endpoints were known no-ops

### Problem

The public API includes button color and brightness endpoints, but `server.js` only updates local state. The hardware call is commented out.

### Where to look

- `loupedeck-server/server.js` brightness endpoint.
- `loupedeck-server/server.js` button color endpoint.
- `loupedeck/pkg/device/loupedeck.go` — `SetBrightness`, `SetButtonColor`.

### Example

```javascript
currentBrightness = body.value
// Will be implemented when setBrightness is added to ui module
// ui.setBrightness(currentBrightness)
res.json({ value: currentBrightness })
```

```javascript
buttonColors[name] = { r: body.r, g: body.g, b: body.b }
// Will be implemented when setButtonColor is added to ui module
// ui.setButtonColor(name, body.r, body.g, body.b)
res.json({ name: name, r: body.r, g: body.g, b: body.b })
```

### Why it matters

This is not a runtime mystery. The API returns success while intentionally not calling hardware. The smoke tests passed because they only asserted HTTP status and JSON response, not device side effects.

### Cleanup sketch

Do not expose endpoints as "working" until the module supports them. For the next commit, either:

1. Return `501 Not Implemented` when hardware control is unavailable, or
2. Implement a supported JS module API and call it.

Preferred API:

```javascript
ui.setBrightness(value)
ui.setButtonColor("Button1", { r: 255, g: 0, b: 0 })
```

Go-side export should validate bounds rather than masking values with `& 0xFF`.

---

## 6. The WIP hardware control change is conceptually useful but currently broken and too broad

### Problem

The WIP adds `DeviceControl` to `LoupeDeckEnvironment`, which may be a good direction, but it is incomplete and currently breaks the build.

### Where to look

- `loupedeck/runtime/js/env/env.go` — new `DeviceControl` interface.
- `loupedeck/runtime/js/env/device_control.go` — adapter.
- `loupedeck/runtime/js/module_ui/module.go` — new exports.
- `loupedeck-server/main.go` — intended wiring not completed.

### Example

Current `module_ui` WIP includes an unused local:

```go
button, err := deck.ParseButton(name)
if err != nil { ... }
...
if err := env.DeviceControl.SetButtonColor(name, c); err != nil { ... }
```

The parsed `button` is unused; the compiler rejects this.

### Why it matters

Because `loupedeck-server` depends on the local `loupedeck` checkout, this breaks both the main server and the loupedeck JS packages. It also changes the shared environment type for all loupedeck JS runtime users, not just this server.

### Cleanup sketch

Make it a small, tested change:

```go
// env/env.go
type DeviceControl interface {
    SetButtonColor(device.Button, color.RGBA) error
    SetBrightness(int) error
}
```

Then avoid double-parsing strings:

```go
button, err := deck.ParseButton(name)
if err != nil { panic(runtime.NewTypeError(err.Error())) }
if err := env.DeviceControl.SetButtonColor(button, c); err != nil { ... }
```

Adapter:

```go
type LoupedeckDeviceControl struct { Deck *device.Loupedeck }
func (c LoupedeckDeviceControl) SetButtonColor(b device.Button, col color.RGBA) error {
    if c.Deck == nil { return ErrNoDevice }
    return c.Deck.SetButtonColor(b, col)
}
```

Tests:

- compile package,
- mock `DeviceControl` receives parsed button and color,
- no hardware returns clear JS error,
- values outside 0–255 reject instead of wrapping.

---

## 7. The smoke tests validate transport, not device behavior

### Problem

The smoke scripts are useful for verifying HTTP status codes and in-memory state. They do not prove anything rendered on the physical Loupedeck or that button LEDs changed.

### Where to look

- `ttmp/.../scripts/01-smoke-test.sh`
- `ttmp/.../scripts/02-start-server.sh`
- `ttmp/.../scripts/03-event-poll-demo.sh`

### Example

`PUT /api/v1/buttons/Circle/color` passes because the route returns JSON after updating `buttonColors`, not because hardware confirmed the LED changed.

### Why it matters

The tests gave false confidence. This is exactly why the user could run a curl command, get success, and still see no display/button change.

### Cleanup sketch

Split smoke tests into two classes:

```text
scripts/
  01-api-contract-smoke.sh          # no hardware required
  02-start-server.sh                # tmux launcher
  03-event-poll-demo.sh             # operator/manual event flow
  04-stop-server.sh
  05-hardware-operator-check.py     # asks user to confirm visible changes
```

Hardware operator script should print a checklist:

```text
1. Setting Button1 red. Did it turn red? [y/N]
2. Drawing DEMO page. Do you see text? [y/N]
3. Press Button1. Did polling return an event? [y/N]
```

---

## Root Cause: xgoja Deficiency or Misuse?

### It is not fundamentally an xgoja mismatch

xgoja can compose the required modules. The buildspec validated, and the generated binary could load `express` and `db`. The problem was not module selection.

### It is a lifecycle/documentation gap

The built-in `run` command is documented as "execute a JavaScript file". It is not documented as a long-running server lifecycle. The HTTP provider can start a server, but a script loaded through `run` immediately returns unless the script deliberately blocks. Blocking in JS is not acceptable because route handlers also need the goja owner.

### xgoja already has the right extension point

`providerapi.CommandSetProvider` exists for domain-specific commands. The loupedeck provider already uses this pattern for scenes. The HTTP server should use the same pattern: a provider-owned long-running `serve` command that creates the runtime, initializes provider capabilities, loads `server.js`, and blocks outside the goja owner until signal/context cancellation.

### Documentation should be improved

xgoja docs should include a page or tutorial:

- "Building long-running xgoja services"
- Explain why `run` is short-lived.
- Show a command provider that blocks on SIGINT while keeping the runtime alive.
- Show an express server example with `go-go-goja-http`.

---

## Recommended Recovery Plan

## Phase 0: Stabilize workspace

1. Stop any running `loupedeck-api` tmux session.
2. Decide whether to keep or revert current WIP hardware-control edits.
3. Restore green build:

```bash
cd loupedeck && go test ./runtime/js/... ./runtime/host/... ./pkg/device/...
cd ../loupedeck-server && go build ./...
```

Minimum action if not finishing hardware control immediately:

```bash
cd loupedeck
git checkout -- runtime/js/env/env.go runtime/js/module_ui/module.go
rm -f runtime/js/env/device_control.go
cd ../loupedeck-server
git checkout -- main.go
```

Because `loupedeck` itself is not a Git checkout root in this workspace listing, use care: inspect `.git` indirection before running destructive commands.

## Phase 1: Convert server lifecycle back to xgoja

Add a package such as:

```text
loupedeck-server/pkg/xgoja/serverprovider/provider.go
```

It should register:

```go
providerapi.CommandSetProvider{
    Name: "server",
    DefaultMount: "server",
    New: newServerCommandSet,
}
```

The command should:

1. receive `CommandSetContext`,
2. collect config sections from `ctx.SelectedModules`,
3. add `script` argument,
4. create runtime using `ctx.RuntimeFactory.NewRuntime`,
5. run `providerutil.InitRuntimeFromSections`,
6. load `server.js`,
7. wait for signal/context,
8. close runtime.

## Phase 2: Update buildspec

Add the server provider to `xgoja.yaml`:

```yaml
packages:
  - id: loupedeck-server
    import: github.com/go-go-golems/loupedeck-server/pkg/xgoja/serverprovider
    register: Register
    replace: .

commandProviders:
  - id: loupedeck-http-server
    package: loupedeck-server
    name: server
    mount: server
    runtimeProfile: server
```

Then the usage becomes:

```bash
xgoja build -f xgoja.yaml --xgoja-replace ../go-go-goja
./dist/loupedeck-server server serve server.js --http-listen :9876 --deck-enabled
```

or, if mounted at root:

```bash
./dist/loupedeck-server serve server.js --http-listen :9876 --deck-enabled
```

## Phase 3: Make API honesty match hardware reality

Until hardware control exists, these endpoints should return 501 when called with hardware actions:

- `PUT /api/v1/brightness`
- `PUT /api/v1/buttons/:name/color`

Once `loupedeck/ui` exposes hardware control, change them to call:

```javascript
ui.setBrightness(value)
ui.setButtonColor(name, { r, g, b })
```

## Phase 4: Add hardware-control JS API cleanly

Keep this separate from xgoja lifecycle work.

Implementation sketch:

```go
// env/device_control.go
type DeviceControl interface {
    SetButtonColor(device.Button, color.RGBA) error
    SetBrightness(int) error
}
```

Expose in `module_ui`:

```go
exports.Set("setButtonColor", func(name string, value map[string]int) error { ... })
exports.Set("setBrightness", func(value int) error { ... })
```

Wire in hardware capability after `deckConn` exists:

```go
environment.DeviceControl = env.LoupedeckDeviceControl{Deck: deckConn}
```

## Phase 5: Instrument rendering before debugging visual output

Add logging or metrics at these points:

- `uiRT.SetDirtyHandler`: reason and page/tile count.
- `present.Invalidate`: reason.
- `FlushFunc`: dirty counts and returned flush count.
- `renderer.Flush`: display/tile target names.
- `device.Display.Draw`: display name, x/y, width/height.
- writer stats: queued/sent/failed.

Expected sequence after `POST /api/v1/pages/show`:

```text
server.js: ui.show("demo")
ui.Show: activePage=demo
ui.invalidatePage: dirtyTiles=N
present.Invalidate: reason=ui-dirty
present.loop: wake
renderer.Flush: flushed=N
Display.Draw: main x=0 y=0 w=90 h=90
writer: sentCommands increases
```

If any line is missing, that is the failing boundary.

## Phase 6: Rewrite tests as contract + hardware operator checks

Keep the current API smoke test, but name it honestly:

```text
01-api-contract-smoke.sh
```

Add an operator-confirmed hardware script:

```python
# 05-hardware-operator-check.py
call("PUT", "/api/v1/buttons/Button1/color", {"r":255,"g":0,"b":0})
ask("Did Button1 turn red?")
call("POST", "/api/v1/pages", {...})
call("POST", "/api/v1/pages/show", {"name":"hardware-check"})
ask("Do you see HARDWARE CHECK on the display?")
ask("Press Button1 now")
resp = call("GET", "/api/v1/events?since=0")
assert event present
```

---

## Proposed Final Architecture

```text
xgoja.yaml
  ├─ packages:
  │   ├─ loupedeck provider          -> loupedeck/ui + hardware capability
  │   ├─ go-go-goja-http provider    -> express + HTTP server capability
  │   ├─ go-go-goja-host provider    -> db module
  │   └─ loupedeck-server provider   -> long-running serve command
  │
  └─ commandProviders:
      └─ loupedeck-server.server

Generated binary
  └─ serve server.js --http-listen :9876 --deck-enabled
      ├─ create runtime from profile
      ├─ initialize providers from CLI sections
      │   ├─ HTTP provider starts net/http.Server
      │   └─ Loupedeck provider connects hardware + starts renderer
      ├─ load server.js
      │   ├─ register express REST routes
      │   └─ register ui.onButton/onKnob/onTouch event persistence
      └─ wait for signal/context
```

The JavaScript script stays small and focused:

```javascript
const express = require("express")
const ui = require("loupedeck/ui")
const db = require("db")

const app = express.app()

db.configure("sqlite3", "loupedeck_events.db")
createTables()

app.get("/api/v1/info", infoHandler)
app.post("/api/v1/pages", createPageHandler)
app.post("/api/v1/pages/show", showPageHandler)
app.get("/api/v1/events", pollEventsHandler)

ui.onButton("Button1", event => recordEvent(event))
```

Go owns lifecycle and hardware. JavaScript owns API behavior.

---

## Intern-Oriented Glossary

- **goja**: JavaScript VM implemented in Go.
- **runtime owner**: serialized owner loop for safe access to a goja VM; all JS runs through it.
- **xgoja**: generator/build system that composes provider packages into custom goja binaries.
- **provider package**: Go package registering modules, config sections, command providers, and optional jsverbs.
- **runtime profile**: xgoja YAML section listing which provider modules are exposed to one command.
- **command provider**: provider extension that contributes Cobra/Glazed commands to the generated binary.
- **gojahttp**: adapter from Go `net/http` to JavaScript route handlers.
- **present runtime**: background invalidation/flush loop for retained UI updates.
- **retained UI**: model where pages/tiles store desired state; renderer redraws dirty state to hardware.

---

## Immediate Recommendation

Do not keep building on the manual `main.go` path. Treat it as a spike that proved the need for a long-running lifecycle.

Next high-confidence task:

1. Restore green build.
2. Implement `loupedeck-server` as an xgoja command provider.
3. Build with `xgoja build` again.
4. Add rendering instrumentation.
5. Only then resume hardware display/button debugging.

This keeps the project aligned with the original design goal: a standalone xgoja binary composed from provider packages, not a one-off hand-written host.
