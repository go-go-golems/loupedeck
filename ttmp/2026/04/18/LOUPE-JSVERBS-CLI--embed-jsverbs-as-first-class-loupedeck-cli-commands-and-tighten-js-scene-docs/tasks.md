# Tasks

## Ticket setup and documentation

- [x] Create ticket `LOUPE-JSVERBS-CLI`
- [x] Write an intern-oriented analysis and implementation guide
- [x] Write an investigation diary capturing the evidence and rationale
- [x] Upload the initial ticket bundle to reMarkable
- [x] Revise the design after the product decision that `loupedeck verbs ...` should be the execution namespace
- [x] Extend the design to use a repository model inspired by `sqleton` and Glazed config plans

## Scope confirmation

- [x] Confirm that jsverbs CLI embedding is feasible in loupedeck
- [x] Confirm that execution must stay on the live hardware-owned loupedeck runtime/session
- [x] Confirm that `run` should become the plain-file runner and `verbs` should become the annotated-scene runner
- [x] Confirm that built-in scripts and user repositories should both be supported
- [x] Explicitly defer unrelated JS follow-ups from the earlier review

## Phase 1: lock the repository/config contract

### 1.1 Define the app config shape

- [ ] Add or choose the Go types that represent loupedeck app config for verb repositories
- [ ] Model `verbs.repositories` as structured objects, not just raw strings
- [ ] Include fields for at least `path`, and decide whether `name` and `enabled` ship in v1
- [ ] Decide whether v1 accepts a string-list shorthand in addition to object form, or whether that is deferred
- [ ] Document the final YAML shape in the design doc and public help/docs

### 1.2 Define repository sources and precedence

- [ ] Write down the final precedence contract in code comments/tests:
  1. embedded internal repository
  2. repositories from resolved app config
  3. repositories from `LOUPEDECK_VERB_REPOSITORIES`
  4. repositories from repeated `--verbs-repository`
- [ ] Decide whether an explicit app config flag is needed for bootstrapping, or whether v1 only uses default app config locations plus env + repository flags
- [ ] Decide whether missing configured repository paths are hard errors or soft warnings
- [ ] Decide whether disabled repositories remain in the parsed config model but are filtered before scanning

### 1.3 Add repository normalization helpers

- [ ] Add helpers that trim whitespace, expand `~`, normalize relative paths, dedupe paths, and preserve stable ordering
- [ ] Ensure normalization behavior is shared by app config, env, and explicit CLI repositories
- [ ] Ensure error messages retain the original source of the repository path (config/env/CLI)

### 1.4 Add tests for the contract

- [ ] Add tests for YAML parsing of the chosen app config shape
- [ ] Add tests for path normalization and dedupe behavior
- [ ] Add tests for precedence between app config, env, and explicit CLI repositories
- [ ] Add tests for empty/blank/duplicate entries

## Phase 2: implement repository discovery bootstrap

### 2.1 Add app-config discovery via Glazed config plans

- [ ] Add a small bootstrap/discovery unit responsible for loading loupedeck app config before final Cobra registration
- [ ] Build a Glazed config plan using the v1 sources for app config discovery
- [ ] Use system/XDG/home app config paths in the plan
- [ ] If adopted, add an explicit app config override path into the plan
- [ ] Keep repo-root and cwd config discovery out of v1 unless the design is explicitly changed

### 2.2 Add env and raw-arg bootstrap parsing

- [ ] Parse `LOUPEDECK_VERB_REPOSITORIES` using the platform path-list separator
- [ ] Add repeated root-level `--verbs-repository` parsing early enough that the command tree can be built before Cobra executes
- [ ] Decide where that early parsing lives so the logic remains testable and does not bloat `main.go`
- [ ] Ensure bootstrap parsing does not consume or corrupt the normal Cobra argument flow

### 2.3 Define a concrete repository descriptor for bootstrap output

- [ ] Introduce a struct that describes one discovered repository, including at least:
  - type/source (`embedded`, `config`, `env`, `cli`)
  - repository name if available
  - filesystem path or embedded root
  - normalized identity used for dedupe/collision reporting
- [ ] Ensure the bootstrap output is deterministic and easy to inspect in tests

### 2.4 Add tests for bootstrap discovery

- [ ] Add tests for config-plan discovery across supported app config locations
- [ ] Add tests for env parsing and repeated `--verbs-repository` parsing
- [ ] Add tests proving deterministic repository ordering
- [ ] Add tests proving duplicate repository entries collapse correctly

## Phase 3: add embedded + filesystem repository scanning

### 3.1 Add the embedded internal repository

- [ ] Decide which built-in scripts are part of the embedded repository for v1
- [ ] Add an embedded FS for the built-in annotated scripts
- [ ] Choose the internal repository root and stable repository name (`builtin` or similar)
- [ ] Scan the embedded repository through `jsverbs.ScanFS(...)`

### 3.2 Add filesystem repository scanning

- [ ] Scan each configured filesystem repository through `jsverbs.ScanDir(...)`
- [ ] Preserve repository identity alongside each resulting registry so errors can name the source repository
- [ ] Decide how to handle unreadable directories, nonexistent paths, and non-JS files

### 3.3 Build a merged discovered-verb model

- [ ] Add a data structure representing one discovered verb, including:
  - repository name/source
  - registry pointer
  - verb pointer
  - source file/module path
  - full jsverbs command path
- [ ] Gather verbs from all repositories into one merged collection
- [ ] Filter to explicit verbs only for v1
- [ ] Sort discovered verbs deterministically before building Cobra commands

### 3.4 Add collision detection

- [ ] Detect duplicate full verb paths across repositories before command registration
- [ ] Make the error include both repositories and both source files/modules
- [ ] Decide whether collisions inside a single repository should be treated the same way

### 3.5 Add tests for scanning and merge behavior

- [ ] Add tests for scanning the embedded repository
- [ ] Add tests for scanning multiple filesystem repositories
- [ ] Add tests for stable ordering of discovered verbs
- [ ] Add tests for duplicate full-path collision errors

## Phase 4: extract reusable live-scene execution helpers

### 4.1 Refactor the current `run` implementation

- [ ] Review `cmd/loupedeck/cmds/run/command.go` and identify the reusable parts:
  - scene session setup/teardown
  - runtime opening
  - presenter/render loop wiring
  - signal/timeout handling
  - post-bootstrap steady-state loop
- [ ] Extract those reusable parts into helper functions or a small helper type that is not tightly coupled to `run` flag decoding
- [ ] Keep the helper focused on caller-owned runtime bootstrapping so both raw scripts and dynamic verbs can use it

### 4.2 Extract verb-specific execution reuse

- [ ] Extract the current jsverb execution path so a caller can provide:
  - the target registry
  - the selected verb
  - parsed values
  - runtime options
- [ ] Ensure the helper still calls `Registry.InvokeInRuntime(...)` inside the live runtime/session
- [ ] Ensure the helper returns errors with the full verb path and source context

### 4.3 Simplify the `run` product boundary

- [ ] Remove `run --verb` entirely as part of the cutover
- [ ] Update `run` help text so plain-file execution is the primary and only responsibility
- [ ] Remove any remaining parsing, validation, or execution code paths that only existed for `run --verb`

### 4.4 Add tests for the extracted helper layer

- [ ] Add unit/integration tests for the extracted session helper with raw script bootstrapping
- [ ] Add unit/integration tests for the extracted session helper with verb bootstrapping
- [ ] Keep or update the existing runtime/presenter regression coverage so auto-present behavior remains intact

## Phase 5: replace the static `verbs` command with a dynamic execution tree

### 5.1 Rework root command assembly

- [ ] Refactor `cmd/loupedeck/main.go` so repository discovery happens before final `verbs` command registration
- [ ] Keep `run`, `doc`, logging, and help system wiring intact while swapping in the new `verbs` bootstrap path
- [ ] Ensure startup failures in repository discovery produce actionable errors without obscuring unrelated root command behavior

### 5.2 Rebuild `cmd/loupedeck/cmds/verbs/command.go`

- [ ] Replace or deeply refactor the current inspection-only implementation
- [ ] Keep the top-level `verbs` namespace stable
- [ ] Remove the old inspection-only `verbs list` and `verbs help` commands as part of the cutover
- [ ] Remove or rewrite any tests/docs that still assume the old inspection-only `verbs` subcommands exist

### 5.3 Generate native loupedeck execution commands

- [ ] For each discovered explicit verb, call `CommandDescriptionForVerb(...)`
- [ ] Convert the description into a Cobra command using the existing Glazed/Cobra builder path
- [ ] Attach a loupedeck-native execution implementation that reuses the live-scene helper from Phase 4
- [ ] Ensure generated command help, examples, and parameter descriptions come from jsverbs/jsdoc metadata rather than hand-written duplication

### 5.4 Preserve correct runtime ownership semantics

- [ ] Ensure execution happens on the live device/runtime session, not upstream ephemeral `registry.Commands()` behavior
- [ ] Ensure callbacks, presenter invalidation, and reactive updates remain alive after the initial verb returns
- [ ] Ensure command execution still works when the selected verb is nested under package/section parents

### 5.5 Add tests for the dynamic command tree

- [ ] Add command-tree tests for one embedded repository only
- [ ] Add command-tree tests for embedded + filesystem repositories together
- [ ] Add tests proving `loupedeck verbs documented configure` executes the expected verb
- [ ] Add tests for nested help output such as `loupedeck verbs documented configure --help`
- [ ] Add tests for collision failures during startup registration

## Phase 6: docs and examples tightening

### 6.1 Update public CLI/help documentation

- [ ] Update `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md`
- [ ] Update any other help topics that still mention `run --verb` or `verbs list/help`
- [ ] Document the new repository model, including:
  - embedded builtins
  - app config locations
  - `LOUPEDECK_VERB_REPOSITORIES`
  - repeated `--verbs-repository`
  - duplicate full-path failure behavior
- [ ] Document the intended split:
  - `loupedeck run <file.js>` for plain scripts
  - `loupedeck verbs ...` for annotated scenes
  - `loupedeck doc ...` for extraction/export

### 6.2 Tighten examples and wording

- [ ] Audit examples for shorthand or directory-first raw-script wording that is no longer the intended product UX
- [ ] Convert raw examples to filename-oriented examples consistently
- [ ] Add or update one end-to-end example that shows configuring a repository and then running a nested annotated verb
- [ ] Ensure built-in annotated examples remain valid for both docs and tests

### 6.3 Update ticket docs after implementation

- [ ] Update the design doc to reflect the final implemented repository/config shape if it differs from the proposal
- [ ] Add implementation notes to the investigation diary or a new implementation diary entry
- [ ] Update the changelog with the final implementation summary and file references

## Phase 7: test matrix and validation

### 7.1 Automated tests

- [ ] Run the relevant `go test` targets while iterating on each phase
- [ ] Run `go test ./...` in `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck` before closing the ticket
- [ ] If any tests must stay targeted rather than full-suite, record the reason in the ticket docs

### 7.2 CLI/help validation

- [ ] Verify `loupedeck --help`
- [ ] Verify `loupedeck verbs --help`
- [ ] Verify help for at least one nested generated command
- [ ] Verify the clean-cutover root command tree: `run`, `verbs`, and `doc`, with no legacy wrapper/inspection subcommands left behind

### 7.3 Manual product validation

- [ ] If hardware is available, manually run one built-in annotated verb on a real device
- [ ] If hardware is available, manually run one filesystem-repository annotated verb on a real device
- [ ] Verify that reactive updates continue after the initial verb invocation
- [ ] Verify that duplicate repository collisions fail early with clear error messages

### 7.4 Ticket hygiene before closure

- [ ] Re-run `docmgr doctor --ticket LOUPE-JSVERBS-CLI --stale-after 30`
- [ ] Re-upload the final ticket bundle to reMarkable
- [ ] Record the final validation commands and results in the ticket docs
- [ ] Close or re-scope any tasks left intentionally deferred
