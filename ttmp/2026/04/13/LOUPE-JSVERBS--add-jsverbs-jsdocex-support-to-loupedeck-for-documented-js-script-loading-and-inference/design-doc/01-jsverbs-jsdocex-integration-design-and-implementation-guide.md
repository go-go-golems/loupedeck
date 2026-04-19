---
Title: jsverbs + jsdocex Integration Design and Implementation Guide
Ticket: LOUPE-JSVERBS
Status: active
Topics:
    - loupedeck
    - jsverbs
    - jsdoc
    - goja
    - documentation
    - script-loading
    - inference
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/factory.go
      Note: Canonical runtime composition API via FactoryBuilder, runtime registrars, and runtime initializers
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/runtime.go
      Note: Canonical owned runtime type to adopt inside loupedeck
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/module_roots.go
      Note: Script-path-derived require root helpers for local module resolution
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/scan.go
      Note: Scanner behavior and actual sentinel syntax semantics
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/binding.go
      Note: Binding-plan logic that must remain the single source of truth for invocation
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/command.go
      Note: Command-description generation and current default command wrappers
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go
      Note: Current default jsverbs runtime path; creates and closes a runtime per invocation
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsdoc/extract/extract.go
      Note: Documentation extraction for __doc__, __example__, and doc templates
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/runtimebridge/runtimebridge.go
      Note: Upstream runtime bridge to standardize on for owner/context/loop access
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/runtimeowner/runner.go
      Note: Upstream runtime owner implementation to standardize on
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/doc/09-jsverbs-example-fixture-format.md
      Note: Actual supported jsverbs fixture syntax and semantics
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/testdata/jsverbs/basics.js
      Note: Canonical examples showing __verb__("functionName", {...}) string syntax
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/runtime.go
      Note: Loupedeck-local runtime wrapper to remove after migration
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/pkg/runtimebridge/runtimebridge.go
      Note: Loupedeck-local runtime bridge copy to remove after migration
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/pkg/runtimeowner/runner.go
      Note: Loupedeck-local runtime owner copy to remove after migration
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/env/env.go
      Note: Environment lookup API to preserve while changing implementation
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/pkg/jsmetrics/jsmetrics.go
      Note: Metrics lookup currently coupled to loupedeck-local runtimebridge Values
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/run/command.go
      Note: Current long-lived scene runner that must switch to engine.Runtime
ExternalSources: []
Summary: Revised design that standardizes loupedeck on go-go-goja engine/runtime packages, removes duplicated runtime ownership code, adds loupedeck-specific engine composition through runtime registrars, and integrates jsverbs/jsdoc through upstream extension points suitable for long-lived hardware scenes.
LastUpdated: 2026-04-13T20:05:00-04:00
WhatFor: Provide the corrected plan for importing go-go-goja directly, replacing loupedeck's copied runtime stack, and integrating jsverbs/jsdoc without fighting the current runtime lifecycle mismatch.
WhenToUse: Use before implementing runtime convergence, jsverbs integration, jsdoc extraction, or deleting loupedeck's local runtime bridge/owner/runtime wrapper.
---

# Revised jsverbs + jsdocex Integration Design and Implementation Guide

## Executive Summary

This revised plan deliberately changes direction from the first pass.

The new decision is:

1. **Standardize loupedeck on `go-go-goja` runtime infrastructure** (`engine`, `pkg/runtimeowner`, `pkg/runtimebridge`) instead of maintaining local copies.
2. **Remove loupedeck's bespoke runtime wrapper** in `runtime/js/runtime.go` after migration.
3. **Import `pkg/jsverbs` and `pkg/jsdoc` directly** rather than copying their internals into loupedeck.
4. **Upstream the missing jsverbs extension points** required for long-lived loupedeck scenes.
5. **Keep loupedeck-specific host state in loupedeck-specific bridges**, not in a forked runtime stack.

This is the cleaner long-term architecture, but it only works if we acknowledge one important truth that the first pass underplayed:

> loupedeck scenes are long-lived runtimes with active UI/event callbacks, while `pkg/jsverbs` currently creates and closes a fresh runtime per command invocation.

That means we cannot simply call `registry.Commands()` and be done. We need a host-controlled runtime lifecycle, backed by `go-go-goja/engine`, and jsverbs must expose APIs that let loupedeck reuse a live runtime instead of forcing ephemeral invocation.

## Problem Statement

We want all of the following at once:

- `go-go-goja` as the single runtime composition and ownership stack
- `jsverbs` for script scanning, schema generation, parameter binding, and command discovery
- `jsdoc` for extracting docs, examples, and prose from scripts
- loupedeck's existing long-lived scene behavior, where a script sets up pages, registers callbacks, starts reactive state, and stays alive while hardware events arrive

The current codebase splits these concerns awkwardly:

- `loupedeck` has local copies of `runtimeowner` and `runtimebridge`
- `loupedeck/runtime/js/runtime.go` manually assembles a goja runtime that overlaps conceptually with `go-go-goja/engine`
- `pkg/jsverbs` already solves scan/bind/describe/invoke, but its default invoke path owns the runtime lifecycle itself and closes it after each invocation
- `pkg/jsdoc` is host-agnostic and already fits direct reuse well

The goal of this ticket is therefore not just “add jsverbs.” It is to **converge runtime infrastructure first**, then layer jsverbs/jsdoc on top in a way that respects loupedeck's long-running host session model.

## Evidence-Based Facts We Must Design Around

### Fact 1: jsverbs sentinel syntax is strict and declarative

Per `pkg/doc/09-jsverbs-example-fixture-format.md`, the supported syntax is:

```js
__verb__("listIssues", {
  sections: ["filters"],
  fields: {
    repo: { argument: true },
    filters: { bind: "filters" },
    meta: { bind: "context" }
  }
});
```

Important constraints:

- `__verb__` takes a **string** function name, not an identifier reference.
- metadata must be a strict literal subset.
- `__section__` is file-local unless Go registers shared sections.

This matters because the updated plan must use the real scanner contract, not an invented variant.

### Fact 2: jsverbs already has the logic we want, but it is packaged around ephemeral runtime ownership

`pkg/jsverbs/runtime.go` currently does this in `Registry.invoke(...)`:

1. build a new `engine.Factory`
2. build a new `engine.Runtime`
3. require the target module
4. invoke the function
5. wait for any promise
6. close the runtime

That lifecycle is fine for CLI-like commands that compute a result and exit.

It is **not** fine for loupedeck scenes, because a scene typically:

- registers button/touch/knob callbacks,
- creates reactive state and computed values,
- binds functions that later re-enter JS,
- needs the runtime and event loop to remain alive while the device session runs.

### Fact 3: `go-go-goja/engine` already models the runtime composition loupedeck needs

`engine.FactoryBuilder` already supports:

- `WithRequireOptions(...)`
- `WithModules(...)`
- `WithRuntimeModuleRegistrars(...)`
- `WithRuntimeInitializers(...)`
- `WithModuleRootsFromScript(...)`

and `Factory.NewRuntime(...)` already returns an owned runtime with:

- `VM`
- `Require`
- `Loop`
- `Owner`
- runtime-scoped `Values`
- explicit `Close(ctx)`

This is the correct foundation to adopt.

### Fact 4: loupedeck's current modules are coupled to local runtime copies

Current loupedeck JS modules import:

- `github.com/go-go-golems/loupedeck/pkg/runtimebridge`
- `github.com/go-go-golems/loupedeck/pkg/runtimeowner`

and `env.Lookup(vm)` currently resolves the environment by reading `runtimebridge.Lookup(vm).Values["environment"]`.

This coupling is the main technical reason migration must happen before jsverbs integration is complete.

## Decision Summary

### Decision 1: Use `go-go-goja/engine.Runtime` as the canonical runtime type

After migration, the canonical JS runtime in loupedeck should be `*engine.Runtime`, not the current local `runtime/js.Runtime` wrapper.

### Decision 2: Delete loupedeck-local `pkg/runtimeowner` and `pkg/runtimebridge` after migration

These are copies of upstream concepts and create unnecessary divergence.

### Decision 3: Preserve the public `env.Lookup(vm)` API, but reimplement it without runtimebridge `Values`

We do **not** need to fork upstream `runtimebridge` just to carry loupedeck-specific host objects.

Instead, loupedeck should keep host-only state in a small loupedeck-specific bridge keyed by `*goja.Runtime`.

That bridge should store:

- `*env.Environment`
- optionally direct metrics lookup if we do not derive it from environment

This keeps upstream `runtimebridge` generic and lets loupedeck-specific modules stay loupedeck-specific.

### Decision 4: jsverbs integration requires upstream host-runtime extension points

We should not copy jsverbs internals into loupedeck. Instead, we should add upstream APIs that let hosts:

- reuse jsverbs scan/binding/description logic,
- provide a caller-owned runtime,
- avoid the default create-call-close runtime lifecycle.

### Decision 5: jsdoc can be integrated directly and independently of runtime convergence

`pkg/jsdoc` does not own runtime lifecycle and can be consumed much more directly than jsverbs.

## Proposed Architecture

## 1. Loupedeck runtime composition becomes an engine registrar

Create a loupedeck runtime registrar that implements `engine.RuntimeModuleRegistrar`.

### Responsibilities

1. Register loupedeck native modules into the require registry:
   - `loupedeck/ui`
   - `loupedeck/state`
   - `loupedeck/easing`
   - `loupedeck/anim`
   - `loupedeck/gfx`
   - `loupedeck/present`
   - metrics modules via `pkg/jsmetrics`

2. Ensure and store the loupedeck environment.
3. Seed runtime-scoped values for host-side inspection.
4. Register cleanup hooks for bridge teardown.

### Sketch

```go
type Registrar struct {
    Env *envpkg.Environment
}

func (r Registrar) ID() string { return "loupedeck-runtime" }

func (r Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
    env := envpkg.Ensure(r.Env)

    envpkg.Store(ctx.VM, env)
    ctx.SetValue("environment", env)
    ctx.SetValue("metricsCollector", env.Metrics)

    _ = ctx.AddCloser(func(context.Context) error {
        envpkg.Delete(ctx.VM)
        return nil
    })

    module_state.Register(reg)
    module_ui.Register(reg)
    module_easing.Register(reg)
    module_anim.Register(reg)
    module_gfx.Register(reg)
    module_present.Register(reg)
    jsmetrics.RegisterModules(reg, "loupedeck")
    return nil
}
```

This turns loupedeck from “a project with its own runtime bootstrap” into “a host-specific engine composition.”

That is the right abstraction boundary.

## 2. `env.Lookup(vm)` stays, but changes implementation

Current `env.Lookup(vm)` depends on loupedeck-local `runtimebridge.Values`.

The revised plan is:

- keep `env.Lookup(vm)` as the public API used by native modules,
- reimplement it using an internal `sync.Map` keyed by `*goja.Runtime`,
- add `Store(vm, env)` and `Delete(vm)` used only by the registrar.

### Why this is better than forking upstream runtimebridge

- `environment` is a loupedeck host concept, not a generic go-go-goja runtime concept
- `metricsCollector` is also loupedeck-specific in this repo
- upstream `runtimebridge` should stay small and generic: context, loop, owner
- loupedeck modules already import `envpkg.Lookup(vm)`; changing that implementation is low-churn

## 3. `pkg/jsmetrics` should resolve through `env.Lookup(vm)`

Current `pkg/jsmetrics.Lookup(vm)` also depends on `runtimebridge.Values`.

Revised plan:

```go
func Lookup(vm *goja.Runtime) (*metrics.Collector, bool) {
    env, ok := envpkg.Lookup(vm)
    return env.Metrics, ok && env != nil && env.Metrics != nil
}
```

This lets us remove the separate metrics binding key entirely and avoids another host-specific runtime bridge.

## 4. Replace `runtime/js/runtime.go` with engine-based construction

The current loupedeck runtime wrapper should be retired.

Two acceptable end states:

### Preferred end state

Callers work directly with `*engine.Runtime`.

### Transitional end state

Add a very thin loupedeck helper that just builds an engine factory and returns `*engine.Runtime`, without redefining ownership concepts.

Example:

```go
func NewRuntime(ctx context.Context, env *envpkg.Environment, opts ...engine.Option) (*engine.Runtime, error) {
    builder := engine.NewBuilder(opts...).
        WithRuntimeModuleRegistrars(jsruntime.NewRegistrar(env))
    factory, err := builder.Build()
    if err != nil {
        return nil, err
    }
    return factory.NewRuntime(ctx)
}
```

This is acceptable as a helper. It is **not** acceptable as another bespoke runtime type.

## 5. Upstream jsverbs must expose host-runtime hooks

This is the most important design correction.

Using `registry.Commands()` as-is is not sufficient for loupedeck because those commands use `registry.invoke(...)`, which creates and closes its own runtime.

We need upstream jsverbs APIs that separate:

1. **scanning**
2. **schema/description generation**
3. **argument binding**
4. **runtime invocation**
5. **runtime lifecycle ownership**

### Proposed upstream jsverbs API additions

#### A. Exported command-description builder

```go
func (r *Registry) CommandDescriptionForVerb(verb *VerbSpec) (*cmds.CommandDescription, error)
```

This exposes the existing `buildDescription(...)` logic without forcing the default command wrapper.

#### B. Exported require loader for scanned sources

```go
func (r *Registry) RequireLoader() func(modulePath string) ([]byte, error)
```

This lets hosts compose the jsverbs overlay loader into their own engine runtime.

#### C. Exported invocation against an existing live runtime

```go
func (r *Registry) InvokeInRuntime(
    ctx context.Context,
    rt *engine.Runtime,
    verb *VerbSpec,
    parsedValues *values.Values,
) (interface{}, error)
```

This should internally reuse the existing binding-plan and argument-marshaling logic from `binding.go` + `runtime.go`, but **must not** create or close the runtime.

#### D. Optional convenience API for custom invokers

```go
type Invoker interface {
    Invoke(ctx context.Context, reg *Registry, verb *VerbSpec, parsedValues *values.Values) (interface{}, error)
}

func (r *Registry) CommandsWithInvoker(invoker Invoker) ([]cmds.Command, error)
```

This is optional but would make host integration much cleaner.

### Why these hooks matter

Once these exist, loupedeck can:

- use jsverbs for schema/help generation,
- create one long-lived engine runtime with the loupedeck registrar,
- invoke the entry verb inside that runtime,
- keep the runtime alive while device events continue flowing.

That matches loupedeck's actual execution model.

## 6. Run command lifecycle with the revised runtime model

The long-lived `run` command should become:

1. connect to device
2. build environment
3. build `engine.Runtime` using loupedeck registrar
4. if jsverbs metadata exists and `--verb` is selected:
   - scan the script root
   - build jsverbs registry
   - invoke selected verb inside the live runtime
5. else:
   - run the script in compatibility mode (raw script path)
6. start renderer / present loop
7. keep runtime alive until timeout, signal, or Circle button exit
8. close runtime and device cleanly

### Important consequence

The renderer/present loop remains owned by the host process, not by jsverbs. jsverbs only supplies discovery, binding, and invocation.

## 7. Script-root and local `require()` behavior

We need both:

- jsverbs overlay loading for scanned modules
- normal local filesystem resolution for non-scanned helper modules and adjacent files

The engine already supports both with require options.

### Revised composition

When running a script or verb rooted at `/path/to/examples/js/scene.js`:

```go
builder := engine.NewBuilder(
    engine.WithModuleRootsFromScript(scriptPath, engine.DefaultModuleRootsOptions()),
    engine.WithRequireOptions(require.WithLoader(registry.RequireLoader())),
).WithRuntimeModuleRegistrars(jsruntime.NewRegistrar(env))
```

Because `require.WithLoader(...)` falls back when the loader returns `ModuleFileDoesNotExistError`, this supports:

- overlay-injected scanned modules,
- normal local `require("./helper")` resolution,
- native loupedeck modules.

## 8. jsdoc integration path

`pkg/jsdoc` is much easier to integrate because it does not own the runtime.

### Phase-appropriate use

- scan the target script or script directory using `extract.ParseSource(...)` or batch mode
- build a `DocStore`
- expose docs through:
  - JSON output
  - markdown export
  - later, optional doc browser server

### Recommended sequencing

Do not block runtime convergence on the doc browser UI.

Integrate jsdoc in this order:

1. raw extraction for one script / directory
2. `loupedeck doc --script ... --format json|markdown`
3. later, `--serve` using `pkg/jsdoc/server`

## Gap Analysis Against This Revised Direction

### Gap 1: duplicated runtime infrastructure

**Current:** loupedeck has local runtime owner/bridge/runtime assembly.

**Needed:** use `go-go-goja/engine`, `pkg/runtimeowner`, and `pkg/runtimebridge` directly.

### Gap 2: environment lookup is tied to local runtimebridge values

**Current:** `env.Lookup(vm)` and `jsmetrics.Lookup(vm)` depend on loupedeck-local bridge values.

**Needed:** move to loupedeck-specific env bridge keyed by VM.

### Gap 3: jsverbs default execution model is ephemeral

**Current:** one runtime per invocation, always closed afterward.

**Needed:** host-controlled invocation into an already-live runtime.

### Gap 4: run command is still built around bespoke runtime wrapper + `RunString`

**Current:** `runtime/js/runtime.go` creates its own runtime shape.

**Needed:** `engine.Runtime` plus loupedeck registrar composition.

### Gap 5: docs are not yet wired into scripts or CLI surfaces

**Current:** examples are plain scripts with no jsdoc/jsverbs metadata.

**Needed:** start with one annotated reference example, then expand.

## Correct Script-Side Example

This is the correct jsverbs-style script syntax for this plan:

```javascript
__package__({
  name: "cyb-ito",
  short: "CYB ITO dashboard scenes"
});

__section__("display", {
  title: "Display options",
  fields: {
    refreshRate: { type: "integer", default: 30, help: "Refresh rate in Hz" },
    theme: { type: "choice", choices: ["dark", "light"], default: "dark" }
  }
});

__doc__("configureDashboard", {
  summary: "Configure the dashboard scene",
  params: [
    { name: "layout", type: "string", description: "Layout name" }
  ]
});

function configureDashboard(layout, display, meta) {
  const ui = require("loupedeck/ui");
  ui.page("home", page => {
    page.tile(0, 0, tile => tile.text(layout));
  });
  ui.show("home");
  return { layout, theme: display.theme, rootDir: meta.rootDir };
}

__verb__("configureDashboard", {
  name: "configure",
  parents: ["cyb-ito"],
  sections: ["display"],
  fields: {
    layout: { argument: true },
    display: { bind: "display" },
    meta: { bind: "context" }
  }
});

doc`---
symbol: configureDashboard
---
# Dashboard configuration

Sets up the dashboard scene and activates the home page.
`;
```

## Revised Phased Implementation Plan

### Phase 0: dependency and toolchain convergence

**Goal:** make direct import of go-go-goja technically safe.

**Tasks:**
1. Add `github.com/go-go-golems/go-go-goja` as a dependency in loupedeck.
2. Bump loupedeck's module `go` directive to match the upstream minimum if required.
3. Resolve `goja` / `goja_nodejs` version selection explicitly and document the chosen versions.
4. Add `tree-sitter` dependencies needed by jsverbs/jsdoc.

**Acceptance criteria:**
- `go test ./...` still builds in loupedeck after adding the dependency, before runtime migration work begins.

### Phase 1: runtime convergence onto go-go-goja engine

**Goal:** remove duplicated runtime infrastructure.

**Tasks:**
1. Add loupedeck runtime registrar implementing `engine.RuntimeModuleRegistrar`.
2. Rework `env.Lookup(vm)` to use a loupedeck-specific VM→Environment bridge.
3. Rework `pkg/jsmetrics.Lookup(vm)` to derive the collector from `env.Lookup(vm)`.
4. Switch modules to import upstream `pkg/runtimebridge` and `pkg/runtimeowner`.
5. Convert `runtime/js/runtime.go` into either a thin engine helper or remove it outright.
6. Delete loupedeck-local `pkg/runtimebridge` and `pkg/runtimeowner` after migration is complete.

**Acceptance criteria:**
- existing loupedeck JS runtime tests pass using `engine.Runtime`
- no imports remain from `github.com/go-go-golems/loupedeck/pkg/runtimebridge`
- no imports remain from `github.com/go-go-golems/loupedeck/pkg/runtimeowner`

### Phase 2: migrate the run command to engine.Runtime

**Goal:** keep current scene behavior while using the shared runtime stack.

**Tasks:**
1. Update `cmd/loupedeck/cmds/run/command.go` to create an `engine.Runtime`.
2. Keep renderer/present lifecycle identical to current behavior.
3. Ensure signal/button/timeout exit paths still clean up correctly.
4. Preserve event callback behavior under the new runtime.

**Acceptance criteria:**
- current example scenes still run on hardware
- callback-heavy examples still respond correctly
- runtime close semantics remain clean

### Phase 3: upstream jsverbs host-runtime APIs

**Goal:** make jsverbs usable in long-lived host runtimes.

**Tasks in go-go-goja:**
1. Add `CommandDescriptionForVerb(...)`.
2. Add `RequireLoader()` or equivalent exported loader hook.
3. Add `InvokeInRuntime(...)` that reuses a live runtime.
4. Optionally add `CommandsWithInvoker(...)` for host-defined execution.
5. Keep `Commands()` and current default behavior as backward-compatible convenience APIs.

**Acceptance criteria:**
- jsverbs tests still pass upstream
- new host-runtime tests show a caller-owned runtime can invoke a verb without being closed inside jsverbs

### Phase 4: loupedeck jsverbs integration

**Goal:** run annotated scene scripts through jsverbs while preserving the long-lived runtime.

**Tasks:**
1. Scan the entry script root with jsverbs.
2. Compose engine runtime with:
   - module roots from script path,
   - jsverbs require loader,
   - loupedeck runtime registrar.
3. Add `--verb` selection to `run`.
4. If the script contains verbs:
   - select the configured verb,
   - build parsed values from CLI,
   - invoke that verb inside the existing live runtime.
5. If the script has no verbs:
   - keep compatibility mode for raw script execution.

**Acceptance criteria:**
- `loupedeck run --script ... --verb ...` configures a scene and keeps it alive
- callbacks continue to fire after verb invocation returns
- Glazed help/flags match jsverbs metadata

### Phase 5: jsdoc extraction surfaces

**Goal:** add docs and inference outputs.

**Tasks:**
1. Add `loupedeck doc --script ... --format json|markdown`.
2. Wire `pkg/jsdoc` extraction to the same script root used for verbs.
3. Add one fully annotated example under `examples/js/`.
4. Later, optionally add `--serve` using the jsdoc server.

**Acceptance criteria:**
- docs for the reference example extract successfully
- params/returns/prose are visible in JSON or markdown output

## File-Level Guidance

### New files expected in loupedeck

| File | Purpose |
|------|---------|
| `runtime/js/registrar.go` | Engine runtime registrar for loupedeck modules + env seeding |
| `runtime/js/factory.go` or `runtime/js/engine.go` | Thin builder helper around `engine.NewBuilder()` if needed |
| `runtime/js/env/bridge.go` | VM→Environment bridge backing `env.Lookup(vm)` |
| `cmd/loupedeck/cmds/doc/command.go` | jsdoc extraction CLI |

### Existing files expected to change heavily

| File | Change |
|------|--------|
| `cmd/loupedeck/cmds/run/command.go` | swap from local runtime to engine runtime; add `--verb` path |
| `runtime/js/env/env.go` | preserve API, change lookup implementation |
| `pkg/jsmetrics/jsmetrics.go` | stop depending on bridge `Values`; derive collector from env |
| `runtime/js/module_*/module.go` | switch imports to upstream runtimebridge/runtimeowner |
| `runtime/js/runtime.go` | shrink to compatibility helper or delete |
| `go.mod` | add go-go-goja + tree-sitter deps, align go/toolchain as needed |

### Files expected to be deleted after migration

| File |
|------|
| `pkg/runtimebridge/runtimebridge.go` |
| `pkg/runtimeowner/*.go` |
| possibly `runtime/js/runtime.go` if we remove the helper entirely |

## Testing Strategy

### Runtime convergence tests

- existing `runtime/js/runtime_test.go` coverage should be ported to the new engine-backed runtime path
- add tests ensuring `env.Lookup(vm)` still works after migration
- add tests ensuring JS callbacks still re-enter Go correctly after runtime convergence

### jsverbs host-runtime tests

Required upstream tests in go-go-goja:

1. build registry from fixture directory
2. create caller-owned engine runtime with custom modules
3. invoke a verb via `InvokeInRuntime(...)`
4. verify runtime remains usable afterward
5. verify a second callback into JS still works

### End-to-end loupedeck tests

- one annotated reference scene script
- `run --verb configure ...` on a fake or mock host
- verify page/UI state and continued callback operation

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| runtime migration breaks existing scenes | High | do runtime convergence before jsverbs; preserve current test suite behavior |
| jsverbs upstream API changes take longer than expected | High | explicitly stage upstream work in Phase 3 before loupedeck integration |
| env lookup migration breaks native modules | Medium | preserve `env.Lookup(vm)` API and change implementation underneath |
| dependency/toolchain mismatch when importing go-go-goja | Medium | make Phase 0 a dedicated convergence step, not an incidental part of implementation |
| overexpanding script capabilities accidentally via default go-go-goja modules | Medium | start with loupedeck registrar only; do not enable unrelated default modules unless explicitly chosen |

## Alternatives Considered

### Alternative 1: keep the loupedeck-local runtime and only import scan/doc packages

Rejected because it preserves duplicated ownership/bridge/runtime code and misses the user's stated goal of standardizing on go-go-goja.

### Alternative 2: copy jsverbs internals into loupedeck

Rejected because it would recreate the exact duplication problem we are trying to eliminate.

### Alternative 3: use current jsverbs `Commands()` directly

Rejected because its default runtime lifecycle is ephemeral and therefore incompatible with long-lived loupedeck scene sessions.

## Final Recommendation

Proceed with the new direction, but do it in this order:

1. **converge on go-go-goja runtime infrastructure first**
2. **then add the upstream jsverbs host-runtime APIs**
3. **then integrate verbs into the loupedeck run lifecycle**
4. **then add jsdoc extraction/output surfaces**

This is the correct design if the goal is not merely “get jsverbs working,” but to leave loupedeck with a cleaner and more maintainable runtime architecture afterward.

## Implementation notes discovered during execution

The implemented CLI shape ended up being slightly more pragmatic than the early design sketch:

- `loupedeck run --script ...` remains the compatibility path for plain scripts
- `loupedeck run --script ... --verb ...` runs an annotated jsverbs verb inside the same long-lived hardware/runtime session
- `loupedeck verbs list --script ...` lists explicit verbs
- `loupedeck verbs help --script ... --verb ...` renders the generated Glazed/Cobra help for a selected verb, which is the primary metadata-accurate help surface
- `loupedeck doc --script ... --format json|markdown` exports jsdoc/jsdocex output

That split keeps the hardware runner stable while still exposing metadata-accurate help/flag rendering through a dedicated inspection command.

## Follow-up work intentionally left out of this ticket

- doc browser `--serve` mode for jsdoc server integration
- richer multi-script command trees beyond the current script-root scanning model
- more advanced run-path config UX than `--verb-config` / `--verb-values-json`

## Document History

- **2026-04-13:** initial first-pass design written.
- **2026-04-13 (revised):** design replaced with runtime-convergence-first plan after explicit decision to standardize on go-go-goja and remove loupedeck-local runtime copies.
- **2026-04-14:** implementation completed using the runtime-convergence-first approach, with `run --verb`, `verbs list/help`, `doc`, and the annotated reference example landed.
