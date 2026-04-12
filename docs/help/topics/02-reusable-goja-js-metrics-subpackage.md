---
Title: Reusable Goja JavaScript metrics subpackage
Slug: reusable-goja-js-metrics-subpackage
Short: Use the extracted `pkg/jsmetrics` and `runtime/metrics` packages to add reusable counters and timings to a goja runtime, including scene-oriented helpers and prefix-based module registration.
Topics:
- goja
- javascript
- runtime
- api
- metrics
- performance
- instrumentation
Commands:
- loupe-js-live
Flags:
- log-js-stats
- log-render-stats
- log-writer-stats
- stats-interval
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The reusable JavaScript metrics stack in this repository is now split into two layers on purpose. `runtime/metrics` is the generic in-process collector for counters and timing windows. `pkg/jsmetrics` is the goja-facing bridge that looks up a collector through `runtimebridge`, registers native modules, and exposes both a low-level `metrics` API and a higher-level `scene-metrics` helper API. This split matters because it keeps the instrumentation substrate portable: the current Loupedeck runtime registers these modules under the `loupedeck/...` prefix, but the underlying implementation is no longer conceptually owned by the Loupedeck environment and is intended to be movable into `go-go-goja` later.

## Why this package exists

Adding counters and timings directly to one application runtime is easy. Reusing them across multiple goja hosts is harder unless the implementation is deliberate about boundaries. The main problem is that app-specific module code tends to accidentally depend on app-specific environment objects, which makes later extraction painful.

This repository now avoids that trap by separating concerns:

- `runtime/metrics` stores measurements in a plain Go collector.
- `pkg/jsmetrics` exposes that collector to JavaScript through `runtimebridge`.
- the current Loupedeck runtime chooses the module names (`loupedeck/metrics`, `loupedeck/scene-metrics`) as a registration detail, not as a baked-in implementation dependency.

That architecture is what makes the package reusable in future goja runtimes, including a later move into `go-go-goja`.

## Package layout

The reusable metrics implementation is split across these files:

- `runtime/metrics/metrics.go` — collector, snapshots, and timing aggregation
- `pkg/jsmetrics/jsmetrics.go` — generic goja/runtimebridge integration and module registration
- `runtime/js/runtime.go` — current concrete runtime that binds a collector and registers the modules under the `loupedeck` prefix
- `runtime/js/module_metrics/module.go` — thin compatibility wrapper for `loupedeck/metrics`
- `runtime/js/module_scene_metrics/module.go` — thin compatibility wrapper for `loupedeck/scene-metrics`

The important design point is that the real logic lives in `pkg/jsmetrics`, not in the Loupedeck-specific wrappers.

## Core concepts

The reusable stack has three conceptual layers.

### 1. Collector layer

`runtime/metrics.Collector` stores named counters and named timing windows.

It currently supports:

- `Inc(name, delta)`
- `ObserveDuration(name, d)`
- `ObserveMillis(name, ms)`
- `Snapshot()`
- `SnapshotAndReset()`

This layer does not know anything about goja, module names, or application semantics. It is just a concurrency-safe place to accumulate observations.

### 2. Binding layer

`pkg/jsmetrics` uses `runtimebridge` to find a collector from a running goja VM.

The binding key is:

```go
const BindingKeyCollector = "metricsCollector"
```

This matters because the binding layer is what makes the package portable. A runtime does not need to import any Loupedeck environment type to expose metrics to JavaScript. It only needs to make sure `runtimebridge.Values` contains a collector under that key.

### 3. Module layer

`pkg/jsmetrics` registers two native modules:

- a low-level metrics module
- a higher-level scene-oriented helper module

The registration is prefix-based:

```go
jsmetrics.RegisterModules(registry, "loupedeck")
```

That call currently creates:

- `loupedeck/metrics`
- `loupedeck/scene-metrics`

A future runtime could choose a different prefix or none at all.

## Low-level JavaScript API

The low-level module is for generic counters and timers. It is intentionally small and reusable.

### `metrics.inc(name, delta = 1)`

Increment a named counter.

```javascript
const metrics = require("loupedeck/metrics");
metrics.inc("scene.frames");
metrics.inc("scene.activations", 2);
```

### `metrics.observeMillis(name, value)`

Record a timing sample in milliseconds.

```javascript
metrics.observeMillis("scene.renderAll", 12.5);
```

### `metrics.time(name, fn)`

Measure a synchronous block and record the elapsed milliseconds.

```javascript
metrics.time("scene.renderAll", () => {
  renderAll();
});
```

### `metrics.counted(name, fn)`

Increment a counter and then run a synchronous block.

```javascript
metrics.counted("scene.frames", () => {
  renderAll();
});
```

### `metrics.now()`

Return the current wall-clock time in milliseconds.

```javascript
const t0 = metrics.now();
```

Use the low-level module when you need raw flexibility and do not want any opinionated naming helpers.

## Higher-level scene helper API

The higher-level helper exists because scene authors quickly end up repeating the same naming logic. A scene usually wants consistent prefixes, rebuild-reason counters, activation counters, loop tick counters, and per-tile timing. Repeating that naming by hand works, but it creates noisy scripts and inconsistent metric names.

The helper module is therefore still generic enough to be reusable, but opinionated enough to save work in UI/scene runtimes.

### Create a helper

```javascript
const sceneMetrics = require("loupedeck/scene-metrics").create("scene");
```

That helper automatically prefixes metrics with `scene.`.

### `sceneMetrics.time(suffix, fn)`

```javascript
sceneMetrics.time("renderAll", () => {
  renderAll();
});
```

This records timing under:

- `scene.renderAll`

### `sceneMetrics.timeTile(name, fn)`

```javascript
sceneMetrics.timeTile("SPIRAL", () => {
  drawSpiralTile(...);
});
```

This records timing under:

- `scene.tile.SPIRAL`

### `sceneMetrics.recordLoopTick()`

This increments:

- `scene.loopTicks`

### `sceneMetrics.recordActivation(reason)`

```javascript
sceneMetrics.recordActivation("T3");
sceneMetrics.recordActivation("B1");
```

This records:

- `scene.activations`
- `scene.activations.touch` or `scene.activations.button`

### `sceneMetrics.recordRebuild(reason, fn)`

```javascript
sceneMetrics.recordRebuild("loop", () => {
  renderAll();
});
```

This records:

- `scene.renderAll.calls`
- `scene.renderAll.reason.loop`
- `scene.renderAll.reasonExact.loop`
- and, when `fn` is provided, timing under `scene.renderAll`

### `sceneMetrics.reasonCategory(reason)`

The current helper maps reasons into these categories:

- `initial`
- `loop`
- `touch`
- `button`
- `other`
- `unknown`

This is useful when your event reasons are concrete values like `T12` or `B1` but you still want stable category counters.

## How it is implemented

The implementation works by binding a Go collector into the current VM through `runtimebridge` and then letting the module code look it up lazily.

At a high level:

```text
collector in Go
-> runtimebridge.Values["metricsCollector"]
-> pkg/jsmetrics.Lookup(vm)
-> native module exports
-> JavaScript counters and timers
```

This is the critical portability trick. The JavaScript modules do not need to know about Loupedeck pages, host runtimes, or custom environment types. They only need a collector in `runtimebridge`.

## Integrating it into your own goja runtime

The easiest way to reuse this package in another goja setup is to follow the same shape as the current Loupedeck runtime.

### Step 1 — create a collector

Start by creating a collector that will accumulate your per-runtime measurements.

```go
collector := metrics.New()
```

### Step 2 — register the modules

Register the reusable modules with the module prefix you want your scripts to use.

```go
registry := new(require.Registry)
jsmetrics.RegisterModules(registry, "myapp")
registry.Enable(vm)
```

This gives your scripts:

- `myapp/metrics`
- `myapp/scene-metrics`

If you want a different naming scheme, change the prefix before enabling the registry.

### Step 3 — bind the collector through runtimebridge

Store the collector in the VM bindings.

```go
runtimebridge.Store(vm, runtimebridge.Bindings{
    Context: ctx,
    Loop:    loop,
    Owner:   owner,
    Values: map[string]any{
        jsmetrics.BindingKeyCollector: collector,
    },
})
```

This step is the actual integration point. Without it, the modules load but panic when used because they cannot find a collector.

### Step 4 — execute your script

Once the bindings and registry are in place, JavaScript can use the modules normally.

```javascript
const metrics = require("myapp/metrics");
const sceneMetrics = require("myapp/scene-metrics").create("scene");

sceneMetrics.recordRebuild("loop", () => {
  metrics.inc("custom.work");
});
```

### Step 5 — read snapshots on the Go side

Your host process can periodically inspect or reset the collector.

```go
snapshot := collector.SnapshotAndReset()
for _, key := range metrics.CounterKeys(snapshot) {
    fmt.Printf("%s=%d\n", key, snapshot.Counters[key])
}
```

That is exactly the pattern the current live runner uses when it logs periodic JS-side stats.

## Complete integration example

This stripped-down example shows the whole pattern in one place. It omits unrelated app details and focuses on the metrics integration itself.

```go
vm := goja.New()
loop := eventloop.NewEventLoop()
go loop.Start()

registry := new(require.Registry)
jsmetrics.RegisterModules(registry, "myapp")
registry.Enable(vm)

collector := metrics.New()
owner := runtimeowner.NewRunner(vm, loop, runtimeowner.Options{Name: "myapp-runtime"})
ctx := context.Background()

runtimebridge.Store(vm, runtimebridge.Bindings{
    Context: ctx,
    Loop:    loop,
    Owner:   owner,
    Values: map[string]any{
        jsmetrics.BindingKeyCollector: collector,
    },
})

_, err := owner.Call(ctx, "vm.run", func(_ context.Context, vm *goja.Runtime) (any, error) {
    return vm.RunString(`
        const sceneMetrics = require("myapp/scene-metrics").create("scene");
        sceneMetrics.recordLoopTick();
        sceneMetrics.recordRebuild("initial", () => {
          for (let i = 0; i < 1000; i++) {}
        });
    `)
})
if err != nil {
    panic(err)
}

snapshot := collector.SnapshotAndReset()
fmt.Println(snapshot.Counters["scene.loopTicks"])
fmt.Println(snapshot.Counters["scene.renderAll.calls"])
```

The important lesson is not the exact logging output. The important lesson is that the metrics module only needs a collector bound through `runtimebridge`, so it can travel with any owner-thread goja runtime.

## How the current Loupedeck runtime uses it

The current runtime wires the reusable package here:

- `runtime/js/runtime.go`

It binds the collector like this conceptually:

- `runtimebridge.Values[jsmetrics.BindingKeyCollector] = env.Metrics`

And it registers the JS modules with the `loupedeck` prefix:

- `loupedeck/metrics`
- `loupedeck/scene-metrics`

The live runner then reads snapshots periodically and logs them when requested with:

- `--log-js-stats`
- `--stats-interval`

This repo-specific usage is just one concrete integration of the generic package.

## Design constraints and non-goals

The package is reusable, but it is intentionally not a full profiler.

What it is good at:

- counters
- timing windows
- per-runtime snapshots
- lightweight scene/workload instrumentation
- narrow goja-native module exposure

What it is not trying to be:

- a sampling profiler
- a cross-process metrics system
- a transport for arbitrary structured logs from JS
- a substitute for host-side tracing systems

That narrowness is a feature. It makes the package easy to carry into other goja runtimes without dragging in a large policy surface.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `metrics module requires collector binding` | The VM bindings do not contain a collector under `jsmetrics.BindingKeyCollector` | Store the collector in `runtimebridge.Values` before running JS |
| `require("myapp/metrics")` fails | The modules were not registered with the prefix your script expects | Call `jsmetrics.RegisterModules(registry, "myapp")` before `registry.Enable(vm)` |
| Counters stay at zero even though JS ran | You are reading a different collector instance than the one bound into the VM | Verify the same `*metrics.Collector` is both bound and later inspected |
| `scene-metrics` names do not match your workload | The helper uses opinionated default names like `renderAll` and `tile.<name>` | Either use the low-level `metrics` module directly or wrap the helper with your own naming conventions |
| Moving the package into another runtime feels coupled to Loupedeck | You are still importing the wrapper modules instead of the generic package | Depend on `pkg/jsmetrics` and register your own prefix; treat `runtime/js/module_*metrics` as compatibility wrappers |

## See Also

- [Loupedeck JavaScript runtime API reference](./01-loupedeck-js-api-reference.md) — Current runtime module surface, including the concrete `loupedeck/metrics` and `loupedeck/scene-metrics` exports
- [Build your first live Loupedeck JavaScript script](../tutorials/01-build-your-first-live-loupedeck-js-script.md) — Step-by-step live-runner tutorial for the current repo runtime
- `pkg/jsmetrics/jsmetrics.go` — Source of truth for the reusable goja binding and module registration logic
- `runtime/metrics/metrics.go` — Source of truth for the underlying collector implementation
- `runtime/js/runtime.go` — Current concrete example of binding a collector and registering prefixed modules
- `cmd/loupe-js-live/main.go` — Example host process that periodically snapshots and logs JS-side metrics
