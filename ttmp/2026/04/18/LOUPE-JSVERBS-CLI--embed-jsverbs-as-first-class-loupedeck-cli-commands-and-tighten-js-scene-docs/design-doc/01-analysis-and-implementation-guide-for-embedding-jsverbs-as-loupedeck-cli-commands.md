---
Title: Analysis and implementation guide for embedding jsverbs as loupedeck CLI commands
Ticket: LOUPE-JSVERBS-CLI
Status: active
Topics:
    - loupedeck
    - javascript
    - goja
    - cli
    - documentation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-goja/cmd/jsverbs-example/main.go
      Note: Upstream dynamic Cobra embedding pattern being adapted
    - Path: ../../../../../../../go-go-goja/pkg/jsverbs/command.go
      Note: Upstream command-description and generated-command behavior
    - Path: ../../../../../../../go-go-goja/pkg/jsverbs/runtime.go
      Note: Upstream runtime ownership split between invoke and InvokeInRuntime
    - Path: cmd/loupedeck/cmds/run/command.go
      Note: Current live hardware scene execution path that must remain the runtime owner
    - Path: cmd/loupedeck/cmds/verbs/command.go
      Note: Current inspection-only verbs list/help split that the new UX may partially supersede
    - Path: cmd/loupedeck/main.go
      Note: Current static root command assembly that the new scene embedding will extend
    - Path: docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md
      Note: Current public docs for annotated scenes that will need tightening
ExternalSources: []
Summary: Evidence-based analysis and intern-oriented implementation guide for exposing scanned jsverbs as first-class loupedeck CLI commands while preserving the long-lived hardware runtime model and tightening the related JS docs/examples.
LastUpdated: 2026-04-18T11:30:09.66691294-04:00
WhatFor: Use when implementing the follow-up CLI embedding work after LOUPE-JSVERBS, especially if the goal is to make annotated scene verbs feel like first-class commands rather than a run --verb submode.
WhenToUse: Read before changing loupedeck root Cobra wiring, replacing verbs help/list flows, or adapting the go-go-goja jsverbs-example embedding pattern to the hardware-owned runtime path.
---


# Analysis and implementation guide for embedding jsverbs as loupedeck CLI commands

## Executive summary

Yes: loupedeck can expose scanned jsverbs as CLI verbs in the style of `go-go-goja/cmd/jsverbs-example`, but it should not do so by directly reusing `registry.Commands()` for scene execution.

The reason is architectural. The upstream `jsverbs-example` flow is designed for ephemeral command runtimes: it scans a directory, generates Glazed commands, and lets each generated command invoke JavaScript through `registry.invoke(...)`, which creates and closes its own runtime for each execution. Loupedeck scene execution is different. The `run` path opens a real device connection, attaches event listeners, starts a long-lived presenter loop, and invokes a selected verb inside that already-live runtime with `InvokeInRuntime(...)`.

So the correct direction is a loupedeck-specific embedding layer that reuses the upstream schema-generation APIs (`CommandDescriptionForVerb(...)`) and live-runtime invocation APIs (`InvokeInRuntime(...)`), but wraps them in hardware-aware Cobra commands. The recommended user-facing shape is a dedicated static parent command, for example `loupedeck scene ...`, under which dynamically scanned verb subcommands are registered after a shallow early scan of `os.Args`. This keeps the root command stable, avoids collisions with existing static commands (`run`, `verbs`, `doc`, `help`), and still gives users a jsverbs-example-style command experience.

This ticket is intentionally narrow. It should cover:

1. first-class CLI embedding of annotated scene verbs,
2. coexistence with the existing `run`, `verbs`, and `doc` commands,
3. docs/example tightening related to the new flow.

It should explicitly defer broader JS follow-ups such as doc server mode, multi-script package registries, richer interactive verb value UX, and generic error-reporting polish.

## Problem statement and scope

### Problem

LOUPE-JSVERBS added working jsverbs support to loupedeck, but the current UX is still split across three different static command surfaces:

- `loupedeck run --script ... --verb ...` for actual execution,
- `loupedeck verbs list --script ...` for discovery,
- `loupedeck verbs help --script ... --verb ...` for metadata-accurate help.

That split was the right implementation choice for the first ticket because it preserved the hardware runtime model and avoided overcomplicating `run`. But it is not yet the most ergonomic final CLI shape for annotated scenes.

### Requested outcome

The new goal is to evaluate and plan a second-stage UX where annotated scene verbs appear as first-class CLI commands, similar to `go-go-goja/cmd/jsverbs-example`, while still honoring loupedeck’s long-lived device/runtime lifecycle.

### In scope

- analyze whether jsverbs can be embedded as CLI verbs in loupedeck,
- explain the full context to a new intern,
- propose the command-tree shape and implementation strategy,
- produce ticket tasks for the implementation work,
- include docs/example tightening as part of the same ticket.

### Out of scope

- generic JS runtime error-reporting polish,
- doc server / `--serve` support,
- advanced multi-script scene package registries,
- richer interactive value prompts beyond existing config/JSON resolution,
- revisiting shorthand script-path support as a product feature.

## Current-state architecture and evidence

### 1. Loupedeck currently has a static root command tree

The current root command is assembled in `cmd/loupedeck/main.go` and only adds three static command families:

- `run`
- `verbs`
- `doc`

Evidence:

- `cmd/loupedeck/main.go:17-41` constructs the root, installs logging/help, builds the Glazed-backed `run` command, then adds `verbs` and `doc` as static Cobra commands.

This means there is currently no dynamic registration step driven by script scanning before `rootCmd.Execute()`.

### 2. The current jsverbs UX is intentionally split

The `verbs` command is inspection-only today:

- `verbs list` scans a script or directory and prints discovered explicit verb paths
- `verbs help` scans a script or directory, finds one verb, builds its `CommandDescription`, then renders help for that synthetic command

Evidence:

- `cmd/loupedeck/cmds/verbs/command.go:24-32` defines a static `verbs` parent command.
- `cmd/loupedeck/cmds/verbs/command.go:35-65` implements `list --script <path>`.
- `cmd/loupedeck/cmds/verbs/command.go:68-105` implements `help --script <path> --verb <name>` by calling `registry.CommandDescriptionForVerb(...)` and rendering `--help` through a generated Cobra command.

The docs also describe the split explicitly:

- `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md:79-117` shows `run --verb`, `verbs list`, and `verbs help` as separate workflows.

### 3. Actual scene execution happens in a caller-owned hardware/runtime session

The `run` command does not hand control to `jsverbs.Commands()`. Instead, it decides between two boot paths:

- raw script boot
- jsverbs verb boot

Evidence:

- `cmd/loupedeck/cmds/run/command.go:259-264` switches between raw and verb execution.
- `cmd/loupedeck/cmds/run/command.go:300-328` shows the verb bootstrap path: scan registry, find verb, build description, parse values, compose engine options, and invoke the verb with `registry.InvokeInRuntime(...)`.
- `cmd/loupedeck/cmds/run/command.go:339-360` shows the hardware session beginning: connect device, hold the session open, and then continue into the long-lived scene loop.

This is a key difference from the upstream example. Loupedeck verb execution is not just “call a JS function”; it is “call a JS function inside an already-live hardware/runtime/presenter session.”

### 4. Raw script execution is file-oriented, even though scene metadata scanning can work on directories

The raw boot path resolves a target, but still requires an actual JavaScript file entrypoint.

Evidence:

- `cmd/loupedeck/cmds/run/command.go:266-289` resolves the target and then rejects directory-only raw execution with `raw script execution requires a JavaScript file`.
- `cmd/loupedeck/cmds/run/command.go:121` still describes `script` as “Path to the JavaScript file or scene directory to execute”, which is accurate for the combined feature set but not equally true for all submodes.

This matters for docs tightening. The user explicitly clarified that raw script usage should really be treated as filename-oriented, not as a shorthand/directory-first UX goal.

### 5. Upstream jsverbs-example does dynamic Cobra embedding at startup

The upstream example does almost exactly what the user is asking about conceptually:

1. discover the target directory from raw `os.Args`,
2. scan it before building the command tree,
3. ask the registry for generated commands,
4. add those commands to the Cobra root.

Evidence:

- `go-go-goja/cmd/jsverbs-example/main.go:23-40` scans the directory and builds `registry.Commands()`.
- `go-go-goja/cmd/jsverbs-example/main.go:74-85` adds those generated commands to the root with `cli.AddCommandsToRootCommand(...)`.
- `go-go-goja/cmd/jsverbs-example/main.go:98-110` shows the early argument sniffing used to determine scan directory before command registration.

So the short answer to “can we do that?” is: yes, mechanically, because the same registry already knows how to generate Glazed/Cobra commands.

### 6. But upstream generated commands currently own ephemeral runtimes

The generated upstream commands are not suitable for loupedeck scene execution as-is.

Evidence:

- `go-go-goja/pkg/jsverbs/command.go:37-47` exposes `Registry.Commands()` by wrapping each verb in a generated Glazed/Writer command.
- `go-go-goja/pkg/jsverbs/command.go:326-356` shows `RunIntoGlazeProcessor(...)` and `RunIntoWriter(...)` both delegating to `registry.invoke(...)`.
- `go-go-goja/pkg/jsverbs/runtime.go:18-35` shows `invoke(...)` creating a new runtime through `engine.NewBuilder()`, then closing it after invocation.

That behavior is correct for jsverbs-example and generic CLI scripting, but wrong for loupedeck hardware scenes because it would destroy the live runtime immediately after the verb returns.

### 7. Loupedeck already has the upstream APIs needed for a host-owned adaptation

The previous ticket added exactly the upstream APIs needed to solve this cleanly.

Evidence:

- `go-go-goja/pkg/jsverbs/command.go:49-53` exports `CommandDescriptionForVerb(...)`.
- `go-go-goja/pkg/jsverbs/runtime.go:38-42` exports `RequireLoader()`.
- `go-go-goja/pkg/jsverbs/runtime.go:44-108` exports `InvokeInRuntime(...)`.

That means loupedeck does not need to copy or fork the jsverbs command builder. It can reuse the description/schema generation and the module-loader/runtime-invocation plumbing separately.

## Gap analysis

### What exists today

- annotated scenes can be executed on hardware via `run --verb`
- discovered verbs can be listed
- generated help can be rendered accurately
- docs can be exported
- raw scenes still work

### What is missing relative to the requested UX

- annotated verbs are not first-class Cobra subcommands in the loupedeck CLI tree
- command registration is static, not script-driven
- there is no early scan/bootstrap phase in `cmd/loupedeck/main.go`
- help/docs still present `run --verb` as the main execution path for annotated scenes
- docs/examples do not yet explain the eventual “first-class command” model because it does not exist yet

### Why direct reuse of `registry.Commands()` is insufficient

Because `registry.Commands()` still routes through ephemeral runtime ownership. Loupedeck needs generated command descriptions without generated runtime ownership.

## Answer to the core product question

## Can we embed jsverbs and expose them as CLI verbs like `cmd/jsverbs-example`?

Yes, with one important caveat:

- **Yes** for command discovery, schema generation, help rendering, and Cobra registration.
- **No** if the plan is to reuse `registry.Commands()` unchanged for actual hardware scene execution.

For loupedeck, the right implementation is:

1. scan before Cobra registration,
2. build a command description per verb,
3. wrap each description in a loupedeck-specific execution command that opens the hardware scene session and then calls `InvokeInRuntime(...)` inside that live runtime.

So the answer is “yes, but via a loupedeck-specific adapter layer, not by directly mounting upstream runtime-owning commands.”

## Proposed solution

## High-level design

Introduce a new static parent command in loupedeck, tentatively named `scene`, whose child commands are discovered dynamically from the selected script file.

Recommended UX:

```bash
loupedeck scene --script ./examples/js/12-documented-scene.js documented configure --title OPS
```

This preserves the feel of jsverbs-example while avoiding top-level command collisions with:

- `run`
- `verbs`
- `doc`
- `help`

### Why a static parent command is recommended

A direct root-level embedding like:

```bash
loupedeck --script ./examples/js/12-documented-scene.js documented configure --title OPS
```

is technically possible, but it is less desirable for loupedeck because:

1. the root already contains stable product commands,
2. dynamic command names could collide with future static command names,
3. the hardware workflow benefits from one obvious namespace for scene-entry commands.

The dedicated parent gives us the jsverbs-example mechanics without destabilizing the product root.

## Proposed command tree

### Keep

- `loupedeck run` for plain raw script execution and as the stable low-level hardware runner
- `loupedeck doc` for docs export
- `loupedeck verbs list` as a lightweight inspection/debug helper

### Add

- `loupedeck scene --script <file> <dynamic-jsverb-path> [verb flags]`

### Potential later cleanup

Once dynamic embedded scene commands exist, `loupedeck verbs help` may become redundant for normal users, though it can remain as a debugging/introspection tool.

## Implementation architecture

### A. Early scene-script discovery before dynamic registration

Add a small pre-parser in `cmd/loupedeck/main.go` that looks for the `scene` command plus a `--script` value in `os.Args` before final Cobra assembly.

This should be modeled on `go-go-goja/cmd/jsverbs-example/main.go:23-40` and `:98-110`, but adapted to the loupedeck command layout.

### B. Build a registry for the selected script

Use the existing `scriptmeta` helpers:

- resolve the script file,
- scan the registry,
- limit discovered commands to entry-file verbs rather than all verbs in the scanned root unless the product explicitly wants directory-wide exposure.

Recommendation: start with **entry-file-only** command embedding so the dynamic scene tree matches the file the user selected.

### C. Generate command descriptions, but not upstream runtime-owning commands

For each selected verb:

- call `registry.CommandDescriptionForVerb(verb)`
- wrap that description in a new loupedeck-specific command type

Proposed adapter shape:

```go
type SceneVerbCommand struct {
    *cmds.CommandDescription
    ScriptPath string
    Verb       *jsverbs.VerbSpec
    Registry   *jsverbs.Registry
}
```

The command implementation should:

1. decode Glazed values from Cobra,
2. translate them into `*values.Values`,
3. call the existing hardware session runner,
4. reuse `runSceneSession(...)` with a verb bootstrap equivalent to today’s `prepareVerbBootstrap(...)`.

### D. Reuse the existing hardware runner instead of duplicating it

Refactor the existing `run` package slightly so the live hardware boot path is reusable from both:

- static `run --verb`
- dynamic embedded scene commands

Recommendation: extract one reusable helper such as:

```go
func RunVerbSceneWithValues(ctx context.Context, opts SceneSessionOptions, scriptPath string, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsed *values.Values) error
```

This helper should:

- compose engine options via `scriptmeta.EngineOptionsForTarget(...)`
- open the device session
- open the loupedeck JS runtime
- call `registry.InvokeInRuntime(...)`
- keep the presenter/event loop alive afterward

### E. Keep `run --verb` during the transition

Do not remove `run --verb` in the first implementation.

Why:

- it is already shipped,
- it is a useful low-level fallback,
- it keeps a stable automation interface while the dynamic scene command UX settles.

## Pseudocode and key flows

### 1. Root assembly flow

```go
func main() {
    sceneBootstrap := discoverSceneBootstrap(os.Args[1:])

    root := newStaticRoot()
    root.AddCommand(buildRunCmd())
    root.AddCommand(buildDocCmd())
    root.AddCommand(buildVerbsCmd())

    if sceneBootstrap.Enabled {
        sceneCmd, err := buildEmbeddedSceneCommand(sceneBootstrap)
        cobra.CheckErr(err)
        root.AddCommand(sceneCmd)
    } else {
        root.AddCommand(buildSceneStubCommand()) // explains --script requirement
    }

    cobra.CheckErr(root.Execute())
}
```

### 2. Embedded scene command builder

```go
func buildEmbeddedSceneCommand(cfg SceneBootstrap) (*cobra.Command, error) {
    target, registry, err := scriptmeta.ScanVerbRegistry(cfg.ScriptPath)
    if err != nil { ... }

    verbs := scriptmeta.EntryVerbs(target, registry)
    sceneRoot := &cobra.Command{Use: "scene", Short: "Run annotated scene verbs"}
    sceneRoot.PersistentFlags().String("script", cfg.ScriptPath, "Scene script file")

    for _, verb := range verbs {
        desc, err := registry.CommandDescriptionForVerb(verb)
        if err != nil { ... }
        cmd := &SceneVerbCommand{CommandDescription: desc, ScriptPath: cfg.ScriptPath, Registry: registry, Verb: verb}
        cobraCmd, err := loupedeckcmdcommon.BuildCobraCommandDualMode(cmd)
        if err != nil { ... }
        sceneRoot.AddCommand(cobraCmd)
    }

    return sceneRoot, nil
}
```

### 3. Dynamic scene command execution

```go
func (c *SceneVerbCommand) Run(ctx context.Context, parsed *values.Values) error {
    opts := DefaultSceneSessionOptionsFromRootFlags(...)
    return RunVerbSceneWithValues(ctx, opts, c.ScriptPath, c.Registry, c.Verb, parsed)
}
```

## Design decisions

### Decision 1: Use a dedicated static parent (`scene`) instead of direct root injection

**Why:** safer coexistence with loupedeck’s existing root command family.

### Decision 2: Reuse upstream description/invocation APIs, not upstream generated runtime-owning commands

**Why:** keeps runtime ownership correct for hardware scenes and avoids a fork of jsverbs command-generation logic.

### Decision 3: Keep `run --verb` during the first migration phase

**Why:** stable fallback, easier validation, lower rollout risk.

### Decision 4: Limit embedded commands to the selected file’s entry verbs initially

**Why:** predictable UX and fewer surprises than exposing every verb from an entire scanned tree on day one.

### Decision 5: Include docs/example tightening in this ticket

**Why:** once the CLI shape changes, docs need to explain the new primary path, the continuing role of `run`, and the filename-oriented expectation for raw script usage.

## Alternatives considered

### Alternative A: Keep the current split forever

Pros:

- lowest code churn
- already works

Cons:

- not the ergonomic CLI the user asked for
- keeps annotated scenes feeling like a special submode rather than first-class commands

### Alternative B: Directly mount `registry.Commands()` under loupedeck

Pros:

- minimal code
- very close to jsverbs-example

Cons:

- wrong runtime ownership model for hardware scenes
- would create/close a runtime per command invocation
- would bypass the device-attached scene session that loupedeck needs

Rejected.

### Alternative C: Inject dynamic verbs at the absolute root

Pros:

- most similar to jsverbs-example

Cons:

- command-name collision risk
- less clear product surface
- root help becomes more volatile depending on scan target

Not recommended for the first implementation.

## Phased implementation plan

### Phase 0: Ticket prep and documentation

- land this design doc and the investigation diary
- define explicit ticket scope and out-of-scope items
- capture the docs/example tightening work as part of the same ticket

### Phase 1: Reusable scene-verb execution adapter

Files to review first:

- `cmd/loupedeck/cmds/run/command.go`
- `pkg/scriptmeta/scriptmeta.go`

Tasks:

1. extract a reusable helper for executing one parsed verb in the live hardware session
2. ensure the helper does not depend on `run`-specific flag parsing internals
3. keep `run --verb` behavior unchanged by reusing the new helper from the old path

### Phase 2: Dynamic command bootstrap under `scene`

Files to review first:

- `cmd/loupedeck/main.go`
- `go-go-goja/cmd/jsverbs-example/main.go`

Tasks:

1. implement early `os.Args` sniffing for `scene --script <file>`
2. scan the selected script before full Cobra registration
3. build a static `scene` parent command plus dynamic child commands
4. add a useful stub/usage message when `scene` is invoked without enough information to scan

### Phase 3: Loupedeck-specific jsverb command wrapper

Files to review first:

- `go-go-goja/pkg/jsverbs/command.go`
- `go-go-goja/pkg/jsverbs/runtime.go`
- `cmd/loupedeck/cmds/common/build.go`

Tasks:

1. define `SceneVerbCommand`
2. wrap `CommandDescriptionForVerb(...)` outputs in loupedeck execution commands
3. convert parsed Cobra/Glazed values into the live runtime invocation path
4. confirm both Glaze and bare/writer output modes are either supported or explicitly constrained for scene commands

### Phase 4: Docs and examples tightening

Files to update:

- `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md`
- `docs/help/topics/01-loupedeck-js-api-reference.md`
- any JS tutorial that references the old flow as the primary annotated-scene UX

Tasks:

1. document the new embedded scene command path
2. keep `run --verb` documented as fallback/compatibility mode
3. use filename-oriented examples for raw script execution
4. tighten examples so they do not imply that shorthand or directory-style raw invocation is the intended public UX

### Phase 5: Validation and cleanup

Tasks:

1. add tests for early scene bootstrap and command registration
2. add tests for one embedded scene command executing into the live runtime/session
3. add tests for coexistence with `run`, `verbs`, `doc`, and `help`
4. verify help output remains sane when dynamic scene commands are present
5. keep `verbs list` and `verbs help` only if they still provide debugging value after embedding

## Testing and validation strategy

### Unit tests

- command bootstrap argument parsing
- dynamic command tree assembly for one annotated scene file
- description generation and command naming for nested verb paths
- stable behavior when no verbs are present

### Integration tests

- execute an embedded command against the existing annotated example
- verify the command invokes the verb inside the live runtime and leaves callbacks/presenter active
- verify raw `run --script file.js` still works unchanged
- verify `run --script file.js --verb ...` still works unchanged

### Docs validation

- check `loupedeck --help`
- check `loupedeck scene --help`
- check one embedded verb `--help`
- update examples/help topics to use filename-oriented raw examples

## Risks, alternatives, and open questions

### Risks

1. dynamic command registration can make root help harder to reason about if the scan target is missing or invalid
2. nested verb parents may create command paths that are awkward in the presence of a static `scene` parent
3. Glazed dual-mode command generation may expose output-oriented flags that do not add much value for hardware scene commands
4. keeping both `run --verb` and embedded commands may temporarily duplicate user-facing pathways

### Open questions

1. Should the embedded parent be named `scene`, `scenes`, or something else?
2. Should embedded commands only support Glazed/bare execution, or should some output modes be hidden for hardware scene commands?
3. After this lands, should `verbs help` remain a public command or become mostly an internal/debugging surface?
4. Should the first implementation require `--script <file>` explicitly, or support a configurable default scene file later?

## References

### Key files

- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/main.go`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/run/command.go`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/verbs/command.go`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/cmd/jsverbs-example/main.go`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/jsverbs/command.go`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/jsverbs/runtime.go`

### Related prior ticket

- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/ttmp/2026/04/13/LOUPE-JSVERBS--add-jsverbs-jsdocex-support-to-loupedeck-for-documented-js-script-loading-and-inference/design-doc/01-jsverbs-jsdocex-integration-design-and-implementation-guide.md`
- `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/ttmp/2026/04/13/LOUPE-JSVERBS--add-jsverbs-jsdocex-support-to-loupedeck-for-documented-js-script-loading-and-inference/reference/03-implementation-diary-phase-0-and-phase-1-runtime-convergence.md`
