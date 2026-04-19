---
Title: Lazy command pattern for dynamic verbs loading
Ticket: LOUPE-JSVERBS-CLI
Status: active
Topics:
    - loupedeck
    - javascript
    - goja
    - cli
    - documentation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/main.go
      Note: Root command now mounts a lightweight placeholder instead of eagerly building the full verbs tree
    - Path: /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/verbs/bootstrap.go
      Note: Bootstrap discovery can be performed from command context at execution time rather than only from raw process args
    - Path: /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/verbs/command.go
      Note: Lazy placeholder command, dynamic resolved command tree construction, and help/output adoption live here
ExternalSources: []
Summary: Reference note explaining the lazy dynamic-command pattern used for `loupedeck verbs`, why it avoids eager startup failures, how the placeholder-to-resolved-command handoff works, and where the pattern is reusable for other kinds of dynamic verb loading.
LastUpdated: 2026-04-18T20:55:00-04:00
WhatFor: Read this when reusing or extending the lazy `verbs` bootstrap pattern, especially if command discovery depends on config, repositories, plugins, or other runtime state.
WhenToUse: Use this note when you need dynamic subcommands without making unrelated root commands fail, slow down, or lose rich help/output behavior.
---

# Lazy command pattern for dynamic verbs loading

## Executive summary

Yes: the lazy command pattern is broadly useful anywhere a CLI wants to expose a large or dynamic verb tree whose final shape depends on configuration, repositories, plugins, extensions, or other runtime state.

The basic idea is simple:

1. mount a **cheap static placeholder** command in the root tree,
2. wait until that namespace is actually invoked,
3. compute the real bootstrap from the current command context,
4. build the real command tree on demand,
5. forward execution into the resolved tree while preserving help, output, and context.

That is now the pattern used by `loupedeck verbs`.

It is useful because it decouples **root command startup** from **dynamic subcommand discovery**. In practice that means a broken or expensive verbs repository no longer prevents unrelated commands like `loupedeck run`, `loupedeck doc`, or root `--help` from working.

## Why eager dynamic registration was a problem

The earlier dynamic `verbs` cutover initially discovered repositories and built the resolved `verbs` command tree during root startup in `main.go`.

That had one desirable property: once the process started, the final `verbs` subtree already existed.

But it also had a bad product consequence: any failure in verbs bootstrap became a failure for the entire CLI process, including commands that were not trying to use verbs at all.

Examples of such failures include:

- a missing repository path from app config,
- a stale `LOUPEDECK_VERB_REPOSITORIES` environment variable,
- a duplicate verb-path collision in configured repositories,
- an expensive scan step that makes root startup noticeably slower.

Those are valid failures **for the `verbs` namespace**, but they are not valid failures for the rest of the CLI.

That asymmetry is what motivates the lazy pattern.

## The pattern in one sentence

Treat dynamic verbs loading as a **namespace-local bootstrap concern**, not as a **root-process startup requirement**.

## How the lazy pattern works in loupedeck

### Step 1: Mount a lightweight placeholder under the root

`cmd/loupedeck/main.go` now mounts a cheap `verbs` placeholder instead of eagerly building the resolved verbs tree.

Conceptually:

```go
rootCmd.AddCommand(verbscmd.NewLazyCommand())
```

At this point, the root only knows that there is a `verbs` namespace. It does not yet know which repositories exist or which specific nested commands will be available.

### Step 2: Defer bootstrap until the namespace is invoked

The placeholder command is implemented in `cmd/loupedeck/cmds/verbs/command.go`.

When the user actually invokes something under `loupedeck verbs ...`, the placeholder:

1. computes bootstrap state from the current Cobra command context,
2. constructs the real resolved command tree,
3. forwards execution into that resolved tree.

In loupedeck, that bootstrap includes repository discovery from:

- the embedded built-in repository,
- app config,
- `LOUPEDECK_VERB_REPOSITORIES`,
- repeated `--verbs-repository` flags.

### Step 3: Discover bootstrap from the command context, not just raw process args

A useful refinement is that lazy loading should not be forced to depend only on `os.Args`.

For loupedeck, `cmd/loupedeck/cmds/verbs/bootstrap.go` now exposes `DiscoverBootstrapFromCommand(cmd *cobra.Command)`, so the lazy dispatcher can compute repository/bootstrap state from the invoked command context.

That matters because once a CLI has a real root command with global flags, output streams, and parsed state, the lazy subtree should derive as much as possible from that existing command context instead of re-parsing the process from scratch.

This improves reuse and makes the bootstrap logic easier to test.

### Step 4: Build the real command tree only after bootstrap succeeds

Once bootstrap discovery succeeds, the placeholder creates the fully resolved verbs tree via the normal dynamic builder.

Conceptually:

```go
bootstrap, err := DiscoverBootstrapFromCommand(cmd)
resolvedCmd, err := NewCommand(bootstrap)
```

Only here does loupedeck perform the real repository scan and register all discovered verbs.

This means failures are now localized correctly:

- `loupedeck verbs ...` can fail if repository bootstrap is invalid,
- `loupedeck run ...` and `loupedeck doc ...` remain unaffected.

### Step 5: Preserve help, usage, and output behavior when forwarding

One subtle but important implementation detail is that the resolved subtree must not lose the root command’s UX behavior.

The first lazy version restored startup isolation but accidentally regressed the rich help rendering for `loupedeck verbs --help`, because the resolved command tree was executed as a standalone root and therefore did not inherit the main root’s custom help renderer/templates.

The fix was to explicitly adopt the original command’s:

- output writer,
- error writer,
- help function,
- usage function,
- help template,
- usage template.

In loupedeck that handoff now happens before executing the resolved subtree.

This is a general lesson of the pattern: lazy execution is not just about discovering commands late; it is also about preserving the parent CLI’s UX contract.

### Step 6: Forward args and context into the resolved tree

After adopting help/output behavior, the placeholder forwards the actual args and execution context:

```go
resolvedCmd.SetArgs(args)
return resolvedCmd.ExecuteContext(cmd.Context())
```

That keeps cancellation, command-scoped context, and runtime wiring aligned with the original invocation.

## Minimal pseudocode

A reusable version of the pattern looks like this:

```go
func NewLazyNamespaceCommand() *cobra.Command {
    return &cobra.Command{
        Use:                "verbs",
        Short:              "Run dynamic verbs",
        DisableFlagParsing: true,
        Args:               cobra.ArbitraryArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            bootstrap, err := DiscoverBootstrapFromCommand(cmd)
            if err != nil {
                return err
            }

            resolvedCmd, err := NewResolvedCommand(bootstrap)
            if err != nil {
                return err
            }

            AdoptHelpAndOutput(cmd, resolvedCmd)
            resolvedCmd.SetArgs(args)
            return resolvedCmd.ExecuteContext(cmd.Context())
        },
    }
}
```

The important parts are not the exact function names. The important parts are the responsibilities:

- cheap placeholder creation,
- deterministic bootstrap discovery,
- on-demand resolved-tree construction,
- explicit help/output adoption,
- final execution handoff.

## Why this pattern is reusable beyond loupedeck

This pattern is a good fit whenever the final subcommand tree is not purely static.

Examples:

- JavaScript verbs discovered from embedded or filesystem repositories
- plugin directories that contribute commands
- extension packages loaded from config
- remote or cached registries that describe available operations
- project-local automation directories such as `.tool/verbs/`
- generated commands whose schemas depend on runtime metadata

The shared property across all of these cases is that the CLI wants a real command UX, but the available commands are not known cheaply and safely at compile time.

## Benefits

### 1. Failures are localized to the namespace that owns them

This is the biggest product win.

Dynamic verb failures still surface clearly, but they no longer take down unrelated commands.

### 2. Faster and more reliable root startup

Root `--help`, unrelated commands, and general CLI initialization stay cheap because they do not perform repository scans or dynamic registry work unless the relevant namespace is actually used.

### 3. Cleaner ownership boundaries

The root CLI owns global setup.
The namespace owner owns namespace-specific discovery.
The dynamic builder owns registration.
The executor owns runtime/session invocation.

This makes the code easier to reason about and easier to extend.

### 4. Better testability

Bootstrap discovery, lazy dispatch, resolved command construction, and help adoption can all be tested independently.

## Tradeoffs and caveats

The pattern is useful, but it is not free.

### Root help will only know about the placeholder namespace

If the command tree is truly lazy, top-level help cannot eagerly enumerate every dynamically discovered grandchild without performing the same discovery work you were trying to defer.

That tradeoff is usually acceptable when the namespace itself is discoverable and has its own rich help.

### Help and output inheritance must be preserved deliberately

This was the main subtle bug loupedeck hit. If the resolved subtree is executed as an independent root, plain Cobra defaults can leak back in unless you copy the parent CLI’s configured help and output behavior.

### Command-path wording may differ if the resolved subtree behaves as its own root

Depending on how help rendering is implemented, a lazily executed subtree may render itself as `verbs ...` rather than `loupedeck verbs ...` unless the surrounding UX is adjusted deliberately.

That is not necessarily wrong, but it is something to review consciously.

### Completion may need extra design

If you want shell completion to know the full discovered subtree ahead of time, lazy registration can complicate the completion path. In some CLIs this is acceptable; in others it motivates a hybrid strategy.

### Bootstrap discovery should be deterministic for a given invocation

If the bootstrap logic depends on mutable global state in surprising ways, the lazy path becomes harder to reason about. Keep the bootstrap input model clear: config, env, flags, and explicit repository sources should have well-defined precedence.

## Recommendations for future reuse

If another CLI namespace wants to use this pattern, the recommended checklist is:

1. Keep the placeholder cheap and side-effect free.
2. Keep bootstrap discovery in a separately testable helper.
3. Prefer command-context discovery over ad hoc `os.Args` re-parsing.
4. Treat help/output adoption as a first-class requirement, not a polish step.
5. Keep failures namespace-local.
6. Document the precedence rules for all dynamic sources.
7. Add regressions for:
   - lazy construction with broken dynamic config,
   - namespace help rendering,
   - nested command help rendering,
   - output writer forwarding.

## Why this belongs in the ticket

This note is worth keeping because the pattern is more general than the one `loupedeck verbs` fix that prompted it.

The current implementation already demonstrates a reusable architecture for:

- dynamic command trees that should not block root startup,
- config-dependent command discovery,
- repository-backed verbs loading,
- preserving a custom CLI UX across late-bound command execution.

That makes it useful not just for future loupedeck verbs work, but also for any other tool in this ecosystem that wants dynamic command namespaces without eager global bootstrap.
