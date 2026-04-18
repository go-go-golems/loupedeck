# Tasks

## Ticket setup and documentation

- [x] Create ticket `LOUPE-JSVERBS-CLI`
- [x] Write an intern-oriented analysis and implementation guide
- [x] Write an investigation diary capturing the evidence and rationale
- [x] Upload the initial ticket bundle to reMarkable
- [x] Revise the design after the product decision that `loupedeck verbs ...` should be the execution namespace
- [x] Extend the design to use a repository model inspired by `sqleton` and Glazed config plans
- [x] Record the downstream implementation in a dedicated implementation diary

## Upstream prerequisite

- [x] Add the upstream `go-go-goja` pluggable invoker API needed for host-owned jsverbs execution
- [x] Record the upstream commits that unblocked the loupedeck implementation:
  - `ad6e30b` — `jsverbs: add pluggable command invokers`
  - `9f2c797` — `docaccess: update glazed help section integration`
  - `4cd7c11` — `runtimeowner: decouple OwnerContext from concrete runner`

## Repository/config contract

- [x] Define loupedeck app config structs for `verbs.repositories`
- [x] Use structured repository objects with `name`, `path`, and optional `enabled`
- [x] Keep v1 app config discovery to system/XDG/home app config locations
- [x] Keep v1 repository precedence as:
  1. embedded internal repository
  2. app-config repositories
  3. `LOUPEDECK_VERB_REPOSITORIES`
  4. repeated `--verbs-repository`
- [x] Normalize repository paths with trimming, `~` expansion, relative-path resolution, dedupe, and stable ordering
- [x] Treat invalid configured repository paths as hard errors
- [x] Add tests for config loading, env parsing, CLI parsing, and duplicate repository collapse

## Repository bootstrap and scanning

- [x] Add an early bootstrap layer that discovers repositories before final Cobra registration
- [x] Add Glazed config-plan based app-config discovery
- [x] Add one embedded built-in scripts repository
- [x] Add filesystem repository scanning via `jsverbs.ScanDir(...)`
- [x] Add embedded repository scanning via `jsverbs.ScanFS(...)`
- [x] Add a merged discovered-verb model with deterministic ordering
- [x] Detect duplicate full verb paths and fail fast with clear source references
- [x] Add bootstrap/scanning tests for config, env, CLI, embedded, and collision scenarios

## Runtime/session refactor and clean cutover

- [x] Extract reusable session/runtime helpers from `cmd/loupedeck/cmds/run`
- [x] Reuse the live hardware-owned runtime/session path for generated annotated commands
- [x] Remove `run --verb`
- [x] Remove the old inspection-only `verbs list` and `verbs help` commands
- [x] Keep `run` as the plain-file runner only
- [x] Keep `verbs` as the annotated-scene runner only
- [x] Preserve raw plain-file execution behavior and shorthand file resolution tests

## Dynamic command tree

- [x] Refactor `cmd/loupedeck/main.go` to build `verbs` after repository discovery
- [x] Replace the static `verbs` implementation with a dynamic execution tree
- [x] Generate commands from jsverbs metadata and merge loupedeck session sections into their schemas
- [x] Route generated command execution through the live loupedeck session helper instead of upstream ephemeral runtime ownership
- [x] Add tests for generated help and dynamic command invocation with the embedded built-in repository

## Docs and examples

- [x] Update `docs/help/topics/03-annotated-scene-scripts-and-jsverbs.md`
- [x] Update help text and examples so `run` is filename-oriented plain-file execution
- [x] Document the repository model, app-config shape, env var, and repeated CLI repository flags
- [x] Document the clean-cutover command split:
  - `loupedeck run <file.js>`
  - `loupedeck verbs ...`
  - `loupedeck doc ...`

## Validation

- [x] Run focused package tests while iterating
- [x] Run `go test ./...` in `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck`
- [ ] If hardware is available, manually run one built-in annotated verb on a real device
- [ ] If hardware is available, manually run one filesystem-repository annotated verb on a real device
- [ ] Verify reactive updates on hardware after the initial verb invocation
- [x] Re-run `docmgr doctor --ticket LOUPE-JSVERBS-CLI --stale-after 30`
- [ ] Re-upload the final ticket bundle to reMarkable
- [ ] Record the final validation commands and results in the ticket docs after any hardware/reMarkable follow-up
