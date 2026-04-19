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
    - Path: ../../../../../../../go-go-goja/pkg/jsverbs/scan.go
      Note: Upstream support for scanning both filesystem directories and embedded fs.FS trees
    - Path: ../../../../../../../sqleton/cmd/sqleton/config.go
      Note: Sqleton app-level repository discovery via config file plus environment merge
    - Path: ../../../../../../../sqleton/cmd/sqleton/main.go
      Note: Sqleton composition of embedded repository content with user-provided filesystem repositories
    - Path: ../../../../../../../sqleton/cmd/sqleton/doc/topics/06-query-commands.md
      Note: Sqleton user-facing repository model and separation between app config discovery and per-command config
    - Path: ../../../../../../../glazed/pkg/config/plan.go
      Note: Declarative config-plan API with layered discovery and provenance-aware resolution
    - Path: ../../../../../../../glazed/pkg/config/plan_sources.go
      Note: Built-in system/XDG/home/git-root/cwd/explicit config source constructors
    - Path: ../../../../../../../glazed/pkg/doc/topics/24-config-files.md
      Note: Guidance for explicit config discovery and parser integration
    - Path: ../../../../../../../glazed/pkg/doc/topics/27-declarative-config-plans.md
      Note: Reference for layered config plans and why they are preferable to hidden discovery helpers
    - Path: cmd/loupedeck/cmds/run/command.go
      Note: Current live hardware scene execution path and existing raw/verb split that will be simplified
    - Path: cmd/loupedeck/cmds/verbs/command.go
      Note: Current inspection-only verbs namespace that will become the dynamic execution namespace
    - Path: cmd/loupedeck/main.go
      Note: Current static root command assembly that the new verbs bootstrap will extend
    - Path: docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md
      Note: Current public docs for annotated scenes that will need tightening
ExternalSources: []
Summary: "Revised design for making `loupedeck verbs ...` the primary dynamic execution namespace for annotated jsverbs scene commands, using a sqleton-style repository model with both embedded builtins and user-provided repositories, discovered through Glazed config plans."
LastUpdated: 2026-04-18T11:30:09.66691294-04:00
WhatFor: "Use when implementing the revised product direction where `loupedeck verbs ...` directly executes annotated scene verbs discovered from both embedded builtins and configured repositories, while `run` returns to being the plain-file runner."
WhenToUse: "Read before changing the root Cobra wiring, replacing the current static verbs list/help commands, or deciding how loupedeck discovers, configures, and registers annotated scripts across embedded and filesystem repositories."
---

# Analysis and implementation guide for embedding jsverbs as loupedeck CLI commands

## Executive summary

Yes: loupedeck can expose annotated jsverbs as first-class CLI verbs in the style of `go-go-goja/cmd/jsverbs-example`, and the best user-facing shape is:

```bash
loupedeck verbs documented configure --title OPS
```

rather than a separate `scene` parent or continued reliance on `run --verb`.

This is the right product direction if the goals are:

1. load and expose all annotated scripts from one or more configured repositories,
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
- the CLI boot process discovers configured repositories before Cobra registration and mounts all discovered annotated verbs under `verbs`.

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
- define the discovery/bootstrap model for repositories and app config,
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
- `go-go-goja/pkg/jsverbs/scan.go:17-71` provides `ScanDir(...)` for filesystem roots
- `go-go-goja/pkg/jsverbs/scan.go:73-110` provides `ScanFS(...)` for embedded `fs.FS` trees

So loupedeck already has the primitives it needs to build native dynamic commands under `verbs` without reintroducing runtime duplication, and it can support both embedded builtins and filesystem repositories without inventing a second scanner.

### 8. Sqleton already uses the closest repository-discovery model we want

Sqleton separates repository discovery from command execution.

Evidence:

- `sqleton/cmd/sqleton/config.go:12-57` defines a tiny app config with a `repositories:` list, discovers it from app config paths, merges repository paths from config plus environment, and normalizes/dedupes the results.
- `sqleton/cmd/sqleton/config_test.go:16-56` proves that repository paths are trimmed, deduped, and merged from config plus env.
- `sqleton/cmd/sqleton/main.go:217-265` loads repository paths from config/env, appends an embedded repository directory, appends valid filesystem repositories, and then loads all commands from the composed repository set.
- `sqleton/cmd/sqleton/doc/topics/06-query-commands.md:48-94` documents the user model explicitly: embedded repositories plus filesystem repositories discovered from app config and env, while per-command settings still come from explicit command config.

This is a strong precedent for loupedeck: one app-level repository-discovery mechanism, one always-present embedded repository, and separate per-command config for actual command execution.

### 9. Glazed already provides the config-discovery API we should build on

Sqleton currently uses `ResolveAppConfigPath(...)` directly, which is simple but narrow. Glazed now has a better long-term model: declarative config plans with explicit layers and provenance.

Evidence:

- `glazed/pkg/config/plan.go:13-119` defines the layered plan model and resolved-file/report outputs.
- `glazed/pkg/config/plan_sources.go:14-88` provides the built-in app-config source constructors:
  - `SystemAppConfig(...)`
  - `XDGAppConfig(...)`
  - `HomeAppConfig(...)`
  - `ExplicitFile(...)`
- `glazed/pkg/config/plan_sources.go:90-132` also provides repo/cwd discovery via `WorkingDirFile(...)` and `GitRootFile(...)`.
- `glazed/pkg/doc/topics/24-config-files.md:68-170` explains how `CobraParserConfig.AppName` and `ConfigPlanBuilder` fit together, and why config discovery should be explicit.
- `glazed/pkg/doc/topics/27-declarative-config-plans.md:1-115` explains why plans are preferable to hidden helpers and how layered discovery remains readable and debuggable.
- `glazed/cmd/examples/config-plan/main.go:37-69` shows a concrete layered plan with repo, cwd, and explicit files.

This means loupedeck does not need to invent a bespoke repository-config story. It can use a sqleton-style repository list, but discover the app config itself through Glazed’s explicit config-plan layer model.

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

1. discover configured repositories before Cobra registration,
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

## Configuration and repository model

### Core requirement

If users should be able to type:

```bash
loupedeck verbs documented configure
```

without `--script`, then loupedeck must know which script repositories to scan **before** it assembles the final Cobra command tree.

### Recommended repository model

Borrow the product model from sqleton, but apply it to jsverbs repositories instead of SQL query repositories:

1. always register one **embedded internal repository** shipped with loupedeck,
2. allow users to add **filesystem repositories** through app config, environment, and explicit CLI bootstrap flags,
3. scan all repositories into one combined `verbs` command tree,
4. fail fast on duplicate full verb paths.

This is a better fit than a bare list of roots because it cleanly separates:

- built-in shipped content,
- user-provided filesystem content,
- app-level discovery policy,
- per-command runtime execution.

### Embedded internal repository

Because `jsverbs` already supports both `ScanDir(...)` and `ScanFS(...)`, the internal built-in scripts can be embedded and scanned directly from an `embed.FS` tree.

That should be the loupedeck equivalent of sqleton’s embedded `queriesFS` repository in `sqleton/cmd/sqleton/main.go:231-258`.

Recommendation:

- always include one internal repository in code,
- give it a stable name such as `builtin` or `loupedeck`,
- treat it as the lowest-precedence repository for discovery purposes.

### User-provided repositories

User repositories should be filesystem-backed directories that are scanned with `jsverbs.ScanDir(...)`.

These should be configured through an app-level config file plus environment and explicit CLI overrides.

### Recommended app config shape

Unlike sqleton’s current flat `repositories: []string`, loupedeck should start with a slightly more structured repository spec because jsverbs repositories may eventually need metadata like display names or enable/disable controls.

Recommended YAML shape:

```yaml
verbs:
  repositories:
    - name: team-scenes
      path: ~/code/acme/loupedeck-scenes
    - name: local-scenes
      path: ~/.loupedeck/verbs
```

Suggested Go structs:

```go
type AppConfig struct {
    Verbs VerbsConfig `yaml:"verbs"`
}

type VerbsConfig struct {
    Repositories []RepositorySpec `yaml:"repositories"`
}

type RepositorySpec struct {
    Name    string `yaml:"name,omitempty"`
    Path    string `yaml:"path"`
    Enabled *bool  `yaml:"enabled,omitempty"`
}
```

A string-only shorthand could be added later if desired, but object form is a better starting point for a new feature.

### App config discovery

Use Glazed config plans rather than an ad hoc path helper.

Recommended v1 discovery plan for the app config itself:

```go
plan := config.NewPlan(
    config.WithLayerOrder(
        config.LayerSystem,
        config.LayerUser,
        config.LayerExplicit,
    ),
    config.WithDedupePaths(),
).Add(
    config.SystemAppConfig("loupedeck").Named("system-app-config"),
    config.XDGAppConfig("loupedeck").Named("xdg-app-config"),
    config.HomeAppConfig("loupedeck").Named("home-app-config"),
    config.ExplicitFile(explicitConfigPath).Named("explicit-app-config").InLayer(config.LayerExplicit),
)
```

Why this shape is recommended for v1:

- it matches Glazed’s explicit layered discovery model,
- it is easier to explain than hidden path helpers,
- it avoids surprising root-command changes based on arbitrary repository-local config in the current working directory.

### Why not enable repo/cwd config discovery immediately?

Glazed supports `GitRootFile(...)` and `WorkingDirFile(...)`, but I do **not** recommend enabling those for v1 of loupedeck’s dynamic `verbs` namespace.

Reason: they make the available command tree change implicitly when the user changes directories or enters a git repository. That may be useful later, but it is more magic than we need for the first implementation.

### Environment and CLI overrides

Use a sqleton-style environment variable for temporary repositories:

```bash
export LOUPEDECK_VERB_REPOSITORIES=/path/to/repo-a:/path/to/repo-b
```

and repeated bootstrap flags for explicit CLI overrides, for example:

```bash
loupedeck --verbs-repository ./examples/js --verbs-repository ~/.loupedeck/verbs verbs documented configure
```

Recommended precedence for repository discovery:

1. embedded internal repository (always included)
2. app-config repositories from resolved config files
3. env repositories from `LOUPEDECK_VERB_REPOSITORIES`
4. explicit CLI repositories from repeated `--verbs-repository`

Repository paths should be normalized, trimmed, deduped, and expanded consistently, just as sqleton does in `sqleton/cmd/sqleton/config.go:42-72`.

### Separation of concerns

Follow sqleton’s documented separation:

- **app config** is for repository discovery only,
- **command config/values** are still passed explicitly to the selected command.

For loupedeck, that means the app config determines which annotated commands exist, while the generated jsverbs flags/config files still determine how a selected command is executed.

## Multi-repository exposure model

Scan all configured repositories and expose every discovered explicit verb under `loupedeck verbs`.

### Command path

The command path should be the verb’s full jsverbs path, including package/parents.

Example:

- JS metadata full path: `documented configure`
- CLI path: `loupedeck verbs documented configure`

This is the simplest mental model and aligns well with the current jsverbs metadata conventions.

### Collision policy

If two repositories produce the same full verb path, registration should fail fast with a clear error that names both repositories and both source files.

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
- if configured repositories resolve successfully, it gets dynamic child commands for discovered verbs
- the old inspection-only `list`/`help` helpers should be removed as part of the cutover

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
    Registry   *jsverbs.Registry
    Verb       *jsverbs.VerbSpec
    Repository string
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
- `run --verb` should be removed rather than preserved behind a compatibility shim
- the old `verbs list` / `verbs help` inspection flow should be removed rather than carried forward

## Pseudocode and key flows

### 1. Build the `verbs` namespace

```go
func buildVerbsCommand(cfg VerbBootstrap) (*cobra.Command, error) {
    verbsRoot := &cobra.Command{
        Use:   "verbs",
        Short: "Run annotated loupedeck scene verbs",
    }

    repositories, err := discoverVerbRepositories(cfg)
    if err != nil {
        return nil, err
    }

    discovered, err := collectAllVerbs(repositories)
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
            Repository:         discoveredVerb.RepositoryName,
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

### Decision 2: Do a clean cutover instead of preserving compatibility wrappers

**Why:** the user explicitly said backward compatibility is not needed for this follow-up direction. That frees the implementation to simplify product boundaries, remove `run --verb`, and delete the old inspection-only `verbs list/help` flow instead of carrying shim behavior forward.

### Decision 3: Scan all configured repositories, not just one selected file

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
- preserves transitional surface area the product explicitly wants to delete

Rejected.

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

### Phase 2: repository bootstrap and startup discovery

Files to review first:

- `cmd/loupedeck/main.go`
- `go-go-goja/cmd/jsverbs-example/main.go`

Tasks:

1. implement early bootstrap discovery for app config, env, and explicit repository flags
2. build the embedded internal repository plus all configured filesystem repositories before final Cobra registration
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

1. startup scanning cost grows with the number of configured repositories and scripts
2. duplicate full paths may appear once multiple repositories are loaded
3. dynamic command registration may complicate help/testing if repository discovery is not deterministic
4. any current automation that still uses `run --verb` will need to migrate during the cutover because that path should be removed

### Open questions

1. Should the app config schema accept only structured repository objects in v1, or also a string-list shorthand?
2. Should the dynamic `verbs` namespace expose only explicit verbs, or also inferred public functions in some later phase?
3. Should command registration happen strictly from configured repositories, or should there be documented conventional fallback directories in a later phase?
4. Do we want an explicit app-config override flag in v1, or only default app-config locations plus env/CLI repository flags?

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
