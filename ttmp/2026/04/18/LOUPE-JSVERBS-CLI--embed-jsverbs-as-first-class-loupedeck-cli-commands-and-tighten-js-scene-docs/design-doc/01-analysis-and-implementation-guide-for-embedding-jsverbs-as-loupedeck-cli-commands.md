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
      Note: Current live hardware scene execution path and existing raw/verb split that will be simplified
    - Path: cmd/loupedeck/cmds/verbs/command.go
      Note: Current inspection-only verbs namespace that will become the dynamic execution namespace
    - Path: cmd/loupedeck/main.go
      Note: Current static root command assembly that the new verbs bootstrap will extend
    - Path: docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md
      Note: Current public docs for annotated scenes that will need tightening
ExternalSources: []
Summary: "Revised design for making `loupedeck verbs ...` the primary dynamic execution namespace for annotated jsverbs scene commands, scanning configured roots at startup and dropping the earlier compatibility-oriented `scene` parent plan."
LastUpdated: 2026-04-18T11:30:09.66691294-04:00
WhatFor: "Use when implementing the revised product direction where `loupedeck verbs ...` directly executes annotated scene verbs discovered from configured roots and `run` returns to being the plain-file runner."
WhenToUse: "Read before changing the root Cobra wiring, replacing the current static verbs list/help commands, or deciding how loupedeck discovers and registers annotated scripts across multiple roots."
---

# Analysis and implementation guide for embedding jsverbs as loupedeck CLI commands

## Executive summary

Yes: loupedeck can expose annotated jsverbs as first-class CLI verbs in the style of `go-go-goja/cmd/jsverbs-example`, and the best user-facing shape is:

```bash
loupedeck verbs documented configure --title OPS
```

rather than a separate `scene` parent or continued reliance on `run --verb`.

This is the right product direction if the goals are:

1. load and expose all annotated scripts from one or more configured roots,
2. make annotated scene entrypoints feel like real CLI commands,
3. avoid backward-compatibility baggage from the transitional `run --verb` UX,
4. keep `run` focused on plain JavaScript file execution.

However, there is one architectural constraint that still matters: loupedeck should **not** execute these commands by directly reusing upstream `registry.Commands()` unchanged. The upstream generated commands still route through `registry.invoke(...)`, which creates and closes an ephemeral runtime per command invocation. Loupedeck scene commands must instead execute inside a live device/runtime/presenter session using `InvokeInRuntime(...)`.

So the correct implementation is:

- keep the **command-tree idea** from `jsverbs-example`,
- keep the **schema/help generation** from `CommandDescriptionForVerb(...)`,
- keep the **live-runtime invocation** path from `InvokeInRuntime(...)`,
- change the product namespace so `verbs` becomes the dynamic execution tree,
- stop designing around backward compatibility for `run --verb`.

This ticket should therefore implement a new model where:

- `loupedeck run <file.js>` is the plain-file runner,
- `loupedeck verbs ...` is the annotated-scene command namespace,
- the CLI boot process scans configured roots before Cobra registration and mounts all discovered annotated verbs under `verbs`.

## Problem statement and scope

### Problem

LOUPE-JSVERBS delivered working jsverbs/jsdoc integration, but the shipped UX is still transitional:

- `loupedeck run --script ... --verb ...` executes annotated scene verbs,
- `loupedeck verbs list --script ...` lists explicit verbs,
- `loupedeck verbs help --script ... --verb ...` renders generated help,
- `loupedeck doc --script ...` exports docs.

That split was useful during the runtime-convergence phase, but it is not the best final user experience for annotated scene scripting.

The user has now clarified the desired end state:

- no special `scene` wrapper namespace,
- no compatibility-oriented emphasis on `run --verb`,
- `verbs` itself should be the real execution namespace,
- the system should load and expose all annotated scripts, not just a single selected file.

### Requested outcome

The new goal is to redesign the follow-up ticket around this command shape:

```bash
loupedeck verbs documented configure
```

with full generated help/flags and hardware-backed execution.

### In scope

- redesign the CLI plan so `verbs` becomes the dynamic execution namespace,
- explain how loupedeck can scan and expose multiple annotated scripts,
- define the discovery/bootstrap model for script roots,
- define how execution still uses loupedeck’s live hardware/runtime session,
- fold the JS docs/example tightening work into the same ticket.

### Out of scope

- broad JS runtime error-reporting work,
- doc server / `--serve`,
- advanced multi-script package registry features beyond initial root scanning and collision handling,
- richer interactive value-entry UX,
- shorthand-path UX for raw script execution.

## Current-state architecture and evidence

### 1. Loupedeck currently has a static root command tree

`cmd/loupedeck/main.go:17-41` builds a static root command and adds:

- `run`
- `verbs`
- `doc`

There is currently no dynamic scan-and-register step before `rootCmd.Execute()`.

### 2. The current `verbs` namespace is inspection-only

`cmd/loupedeck/cmds/verbs/command.go:24-32` defines a static `verbs` parent.

Under it:

- `cmd/loupedeck/cmds/verbs/command.go:35-65` implements `list --script <path>`
- `cmd/loupedeck/cmds/verbs/command.go:68-105` implements `help --script <path> --verb <name>`

So the `verbs` namespace already exists, but it is currently metadata inspection only. It is not the actual execution surface.

### 3. The actual execution path for annotated scenes currently lives under `run`

`cmd/loupedeck/cmds/run/command.go:259-264` branches between raw scripts and verb mode.

For annotated verbs:

- `cmd/loupedeck/cmds/run/command.go:300-328` scans the registry, finds a verb, parses values, and invokes it with `registry.InvokeInRuntime(...)`.
- `cmd/loupedeck/cmds/run/command.go:339-360` begins the live hardware scene session.

This means the loupedeck execution architecture we want already exists, but it is behind the wrong product-facing command shape.

### 4. Raw script execution is file-oriented and should stay that way

`cmd/loupedeck/cmds/run/command.go:266-289` requires an actual JavaScript file for raw execution and rejects directory-only raw execution.

That is compatible with the user clarification: raw execution should be filename-oriented, not shorthand/directory-oriented.

### 5. Upstream `jsverbs-example` proves dynamic embedding is feasible

`go-go-goja/cmd/jsverbs-example/main.go:23-40` scans before command registration.

`go-go-goja/cmd/jsverbs-example/main.go:74-85` adds generated commands to the Cobra root.

`go-go-goja/cmd/jsverbs-example/main.go:98-110` uses early raw-arg inspection to discover the scan target before final command registration.

So the core dynamic embedding pattern is already proven upstream.

### 6. Upstream generated commands are still runtime-owning

`go-go-goja/pkg/jsverbs/command.go:37-47` exposes `Registry.Commands()`.

`go-go-goja/pkg/jsverbs/command.go:326-356` routes execution through `registry.invoke(...)`.

`go-go-goja/pkg/jsverbs/runtime.go:18-35` shows that `invoke(...)` opens and closes a fresh runtime.

This is the main reason we still cannot just mount upstream generated commands directly for loupedeck hardware scenes.

### 7. The upstream APIs needed for native loupedeck execution already exist

- `go-go-goja/pkg/jsverbs/command.go:49-53` exports `CommandDescriptionForVerb(...)`
- `go-go-goja/pkg/jsverbs/runtime.go:38-42` exports `RequireLoader()`
- `go-go-goja/pkg/jsverbs/runtime.go:44-108` exports `InvokeInRuntime(...)`

So loupedeck already has the primitives it needs to build native dynamic commands under `verbs` without reintroducing runtime duplication.

## Gap analysis

### What exists today

- annotated scenes execute correctly in a live loupedeck runtime
- metadata-accurate command descriptions can be generated
- docs can be extracted
- raw plain-file scripts still work

### What is missing relative to the revised desired UX

- `verbs` is not yet the actual execution namespace
- no configured-root scan exists at startup
- loupedeck cannot yet load and expose all annotated scripts automatically
- the current docs still describe `run --verb` and `verbs list/help` as the main annotated-scene workflow
- collision handling and discovery policy for multiple annotated scripts are not yet defined

## Answer to the product question

## Would `loupedeck verbs documented configure` work?

Yes, that would work.

In fact, it is a cleaner final UX than the earlier `scene` proposal if the product direction is:

- load and expose all annotated scripts,
- avoid backward-compatibility obligations for the transitional `run --verb` flow,
- keep one stable namespace for annotated scene commands.

The only architectural caveat remains the same:

- **yes** to dynamic command registration under `verbs`
- **no** to directly reusing upstream runtime-owning generated commands for actual execution

So the recommended execution model is:

1. scan configured roots before Cobra registration,
2. discover all annotated verbs,
3. mount them as subcommands under `loupedeck verbs ...`,
4. execute them through loupedeck’s live hardware/runtime session using `InvokeInRuntime(...)`.

## Proposed solution

## High-level product model

### Plain scripts

```bash
loupedeck run ./examples/js/02-counter-button.js
```

This is the plain-file, non-annotated runner.

### Annotated scene commands

```bash
loupedeck verbs documented configure --title OPS
```

This becomes the primary execution path for annotated scenes.

### Docs export

```bash
loupedeck doc --script ./examples/js/12-documented-scene.js --format markdown
```

This remains a separate extraction surface.

## Namespace decision

Use the existing `verbs` top-level namespace as the dynamic command tree.

This is better than:

- direct root injection, because that risks collisions with `run`, `doc`, and future static commands,
- a separate `scene` parent, because the user explicitly does not want that extra wrapper namespace.

So `verbs` should evolve from:

- inspection-only namespace

to:

- dynamic execution namespace for all annotated scene commands.

## Discovery model

### Core requirement

If users should be able to type:

```bash
loupedeck verbs documented configure
```

without `--script`, then loupedeck must know which roots to scan **before** it assembles the final Cobra command tree.

### Recommended first implementation

Support one or more configured scan roots with deterministic precedence.

Recommended precedence order:

1. explicit CLI bootstrap roots discovered by early raw-arg sniffing,
2. environment/config-driven roots,
3. documented conventional fallback roots if the product wants them.

The exact source of truth for roots is still a product decision, but the design should be explicit that dynamic verbs require **startup-time root discovery**.

### Minimum viable product recommendation

For v1 of this ticket, implement one clear project-level mechanism for roots, for example:

- a persistent root-level `--verbs-root` / repeated roots parsed early from raw args, and/or
- an environment/config-backed root list.

Even if later product work hides this behind defaults, the implementation should first establish a deterministic startup bootstrap path.

## Multi-script exposure model

Scan all configured roots and expose every discovered explicit verb under `loupedeck verbs`.

### Command path

The command path should be the verb’s full jsverbs path, including package/parents.

Example:

- JS metadata full path: `documented configure`
- CLI path: `loupedeck verbs documented configure`

This is the simplest mental model and aligns well with the current jsverbs metadata conventions.

### Collision policy

If two scanned scripts produce the same full verb path, registration should fail fast with a clear error that names both sources.

Do not silently shadow one command with another.

## Execution model

### Important non-goal

Do not execute loupedeck scene verbs through upstream `registry.Commands()`.

That path still owns ephemeral runtimes.

### Required execution model

For each discovered verb:

1. ask the registry for its `CommandDescription`,
2. generate a loupedeck-native command from that description,
3. when executed, open the live device/runtime scene session,
4. invoke the selected verb inside that live runtime with `InvokeInRuntime(...)`,
5. keep the runtime/session alive for callbacks, presenter flushes, and reactive updates.

This preserves the correct scene semantics while still making the command look like a native CLI verb.

## Implementation architecture

### A. Replace the current static `verbs` implementation

Today, `verbs` is a static parent with hand-written `list` and `help` subcommands.

The new design should turn `verbs` into a bootstrapped command tree assembled in `cmd/loupedeck/main.go`.

Recommended structure:

- `loupedeck verbs` root command always exists
- if scan roots resolve successfully, it gets dynamic child commands for discovered verbs
- optional debugging helpers like `list`/`help` can be retained only if they still add value

### B. Scan before final registration

Model this on `jsverbs-example`, but under the existing `verbs` namespace.

Pseudo-bootstrap flow:

```go
func main() {
    bootstrap := discoverVerbBootstrap(os.Args[1:])

    root := newStaticRoot()
    root.AddCommand(buildRunCmd())
    root.AddCommand(buildDocCmd())

    verbsCmd, err := buildVerbsCommand(bootstrap)
    cobra.CheckErr(err)
    root.AddCommand(verbsCmd)

    cobra.CheckErr(root.Execute())
}
```

### C. Build loupedeck-native dynamic verb commands

For each discovered verb:

- call `registry.CommandDescriptionForVerb(verb)`
- convert that description into a Cobra command using the normal Glazed/Cobra builder
- attach a loupedeck-native execution implementation underneath

Suggested command type:

```go
type EmbeddedVerbCommand struct {
    *cmds.CommandDescription
    Registry *jsverbs.Registry
    Verb     *jsverbs.VerbSpec
    Roots    []string
}
```

This is not a user-facing wrapper namespace. It is just the internal execution type that binds a generated command description to the loupedeck runtime/session path.

### D. Reuse the live hardware scene-session code from `run`

Refactor `cmd/loupedeck/cmds/run/command.go` so the live-session logic is reusable from the new `verbs` dynamic commands.

The reusable helper should:

- open the device
- attach the environment
- open the JS runtime with the right module roots/loaders
- invoke the selected verb with parsed values
- leave the session alive afterward

### E. Simplify product responsibility boundaries

After this redesign:

- `run` should be treated as the plain-file runner
- `verbs` should be treated as the annotated-scene runner
- `doc` should remain the doc-export surface

The implementation does not need to preserve `run --verb` as a compatibility commitment if the product no longer wants it.

## Pseudocode and key flows

### 1. Build the `verbs` namespace

```go
func buildVerbsCommand(cfg VerbBootstrap) (*cobra.Command, error) {
    verbsRoot := &cobra.Command{
        Use:   "verbs",
        Short: "Run annotated loupedeck scene verbs",
    }

    registries, err := scanConfiguredRoots(cfg.Roots)
    if err != nil {
        return nil, err
    }

    discovered, err := collectAllVerbs(registries)
    if err != nil {
        return nil, err
    }

    for _, discoveredVerb := range discovered {
        desc, err := discoveredVerb.Registry.CommandDescriptionForVerb(discoveredVerb.Verb)
        if err != nil {
            return nil, err
        }
        cmd := &EmbeddedVerbCommand{
            CommandDescription: desc,
            Registry:           discoveredVerb.Registry,
            Verb:               discoveredVerb.Verb,
            Roots:              cfg.Roots,
        }
        cobraCmd, err := loupedeckcmdcommon.BuildCobraCommandDualMode(cmd)
        if err != nil {
            return nil, err
        }
        verbsRoot.AddCommand(cobraCmd)
    }

    return verbsRoot, nil
}
```

### 2. Execute one dynamic embedded verb

```go
func (c *EmbeddedVerbCommand) Run(ctx context.Context, parsed *values.Values) error {
    return runpkg.RunEmbeddedVerb(ctx, runpkg.EmbeddedVerbOptions{
        Registry: c.Registry,
        Verb:     c.Verb,
        Values:   parsed,
    })
}
```

### 3. Collision check

```go
func collectAllVerbs(registries []*jsverbs.Registry) ([]DiscoveredVerb, error) {
    seen := map[string]DiscoveredVerb{}
    for _, registry := range registries {
        for _, verb := range registry.Verbs() {
            path := verb.FullPath()
            if prev, ok := seen[path]; ok {
                return nil, fmt.Errorf(
                    "duplicate jsverb path %q from %s and %s",
                    path,
                    prev.Verb.SourceRef(),
                    verb.SourceRef(),
                )
            }
            seen[path] = DiscoveredVerb{Registry: registry, Verb: verb}
        }
    }
    return stableSorted(seen), nil
}
```

## Design decisions

### Decision 1: Use `verbs` itself as the dynamic execution namespace

**Why:** this matches the desired product UX and avoids the extra `scene` wrapper namespace.

### Decision 2: Stop designing around compatibility for `run --verb`

**Why:** the user explicitly said backward compatibility is not needed for this follow-up direction. That frees the implementation to simplify product boundaries.

### Decision 3: Scan all configured roots, not just one selected file

**Why:** the desired command shape implies global discovery of annotated scene commands rather than file-local execution.

### Decision 4: Fail fast on duplicate full verb paths

**Why:** ambiguous command registration is worse than a loud startup error.

### Decision 5: Keep execution loupedeck-native even though command discovery is jsverbs-native

**Why:** command discovery and command execution are different layers; only the discovery layer should mirror `jsverbs-example` directly.

## Alternatives considered

### Alternative A: Keep the earlier `scene --script ...` proposal

Pros:

- easier to parameterize per file
- avoids multi-root discovery immediately

Cons:

- not the desired product UX
- adds an extra namespace layer the user does not want
- still keeps annotated scenes feeling semi-transitional

Rejected.

### Alternative B: Continue using `run --verb` as the main UX

Pros:

- minimal additional code

Cons:

- not first-class command embedding
- keeps annotated scenes as a submode rather than a true command tree

Rejected as the primary end state.

### Alternative C: Inject dynamic commands directly at the root

Pros:

- closest to raw jsverbs-example behavior

Cons:

- higher collision risk with static product commands
- less clear mental model than `loupedeck verbs ...`

Rejected.

### Alternative D: Mount upstream `registry.Commands()` directly under `verbs`

Pros:

- smallest code change

Cons:

- wrong runtime ownership for hardware scenes
- would close the runtime immediately after invocation

Rejected.

## Phased implementation plan

### Phase 0: revise docs/ticket scope

- update this design doc to the `verbs`-as-execution-namespace model
- update tasks to remove the old `scene`-parent assumptions
- capture the docs/example tightening work in the same ticket

### Phase 1: reusable live-scene execution helper

Files to review first:

- `cmd/loupedeck/cmds/run/command.go`
- `pkg/scriptmeta/scriptmeta.go`

Tasks:

1. extract a reusable helper for executing one parsed jsverb in the live device/runtime session
2. keep the helper independent from the current `run` flag-decoding layout
3. simplify `run` toward plain-file execution responsibilities

### Phase 2: verbs-root bootstrap and startup discovery

Files to review first:

- `cmd/loupedeck/main.go`
- `go-go-goja/cmd/jsverbs-example/main.go`

Tasks:

1. implement early bootstrap discovery for scan roots
2. build registries for all configured roots before final Cobra registration
3. collect and collision-check all explicit verbs
4. assemble the dynamic `verbs` command tree from the discovered full paths

### Phase 3: dynamic verbs execution tree

Files to review first:

- `cmd/loupedeck/cmds/verbs/command.go`
- `go-go-goja/pkg/jsverbs/command.go`
- `go-go-goja/pkg/jsverbs/runtime.go`

Tasks:

1. replace or deeply refactor the current static inspection-only `verbs` command
2. generate native loupedeck execution commands from `CommandDescriptionForVerb(...)`
3. execute each command through the live loupedeck runtime/session using `InvokeInRuntime(...)`
4. decide whether any debugging helpers remain under `verbs`

### Phase 4: docs and example tightening

Files to update:

- `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md`
- `docs/help/topics/01-loupedeck-js-api-reference.md`
- any tutorial that presents the old split as the preferred annotated-scene UX

Tasks:

1. document `loupedeck verbs documented configure` as the primary annotated-scene UX
2. position `run` as the plain-file runner
3. use filename-oriented raw execution examples consistently
4. remove or demote wording that treats shorthand/directory-first raw execution as intended product UX

### Phase 5: tests and validation

Tasks:

1. add tests for bootstrap root discovery and multi-registry command assembly
2. add tests for duplicate full-path collision detection
3. add tests for executing one embedded `verbs ...` command in a live runtime/session
4. verify coexistence with `run`, `doc`, and root help
5. verify `loupedeck verbs --help` and nested help output quality

## Testing and validation strategy

### Unit tests

- bootstrap root discovery from raw args/config sources
- full-path command tree construction
- duplicate path rejection
- stable ordering of dynamic command registration

### Integration tests

- execute one annotated command through `loupedeck verbs ...`
- verify callbacks/presenter remain alive after the verb returns
- verify plain `run <file.js>` still works for non-annotated scripts
- verify `doc` remains unaffected

### Docs validation

- check `loupedeck --help`
- check `loupedeck verbs --help`
- check one nested command help such as `loupedeck verbs documented configure --help`
- ensure raw examples remain filename-oriented

## Risks, alternatives, and open questions

### Risks

1. startup scanning cost grows with the number of configured roots and scripts
2. duplicate full paths may appear once multiple roots are loaded
3. dynamic command registration may complicate help/testing if root discovery is not deterministic
4. some current automation may rely on `run --verb` and would need migration if that path is reduced or removed

### Open questions

1. What is the initial authoritative source of scan roots: raw flags, env, config file, or a combination?
2. Should `verbs list` survive as a debugging aid once `verbs` becomes the real execution namespace?
3. Should the dynamic `verbs` namespace expose only explicit verbs, or also inferred public functions in some later phase?
4. Should command registration happen strictly from configured roots, or should there be documented conventional fallback directories in v1?

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
