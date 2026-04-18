# Tasks

## Ticket setup and documentation

- [x] Create ticket `LOUPE-JSVERBS-CLI`
- [x] Write an intern-oriented analysis and implementation guide
- [x] Write an investigation diary capturing the evidence and rationale
- [x] Upload the ticket bundle to reMarkable

## Scope definition

- [x] Confirm that jsverbs CLI embedding is feasible in loupedeck
- [x] Revise the ticket to use `loupedeck verbs ...` as the primary execution namespace
- [x] Remove the earlier `scene`-parent / compatibility-oriented assumptions from the plan
- [x] Explicitly defer the other JS follow-ups from the earlier review

## Phase 1: reusable live-scene verb execution

- [ ] Refactor `cmd/loupedeck/cmds/run/command.go` so the live hardware scene execution path can be reused by dynamic `verbs` commands
- [ ] Introduce a reusable helper for invoking one parsed jsverb inside a caller-owned loupedeck runtime/session
- [ ] Simplify `run` toward plain-file execution responsibilities

## Phase 2: startup discovery and dynamic `verbs` bootstrap

- [ ] Add an early bootstrap layer in `cmd/loupedeck/main.go` that discovers jsverbs repositories before final Cobra registration
- [ ] Add app-level repository discovery using a Glazed config plan for `loupedeck` config files
- [ ] Always register one embedded internal scripts repository and merge it with configured filesystem repositories
- [ ] Add env + explicit CLI repository overrides (for example `LOUPEDECK_VERB_REPOSITORIES` and repeated `--verbs-repository` flags)
- [ ] Scan all configured repositories and collect all explicit verbs
- [ ] Detect duplicate full verb paths and fail fast with clear source references
- [ ] Build the dynamic `loupedeck verbs ...` command tree from the discovered full paths

## Phase 3: native loupedeck execution commands under `verbs`

- [ ] Replace or deeply refactor the current static inspection-only `cmd/loupedeck/cmds/verbs/command.go`
- [ ] Generate native loupedeck execution commands from `CommandDescriptionForVerb(...)`
- [ ] Invoke selected verbs through `Registry.InvokeInRuntime(...)` inside the live device/runtime session
- [ ] Decide whether `verbs list` remains as a debugging helper once `verbs` becomes the execution namespace
- [ ] Decide whether `verbs help` remains as a debugging helper or is fully superseded by normal nested Cobra help

## Phase 4: docs and examples tightening

- [ ] Update help/docs to show `loupedeck verbs ...` as the primary annotated-scene UX
- [ ] Position `run` as the plain-file runner
- [ ] Use filename-oriented raw script examples consistently
- [ ] Tighten JS scene docs/examples so they do not imply shorthand or directory-first raw execution is the intended public UX

## Phase 5: tests and validation

- [ ] Add tests for startup root discovery and dynamic command registration
- [ ] Add tests for duplicate full-path collision detection
- [ ] Add tests for executing an embedded `loupedeck verbs ...` command against the annotated example
- [ ] Verify coexistence with `run`, `doc`, and root help
- [ ] Verify `loupedeck verbs --help` and nested help output quality
- [ ] Run `go test ./...` before closing the ticket
- [ ] Run `docmgr doctor --ticket LOUPE-JSVERBS-CLI --stale-after 30` before final review
