# Tasks

## Ticket setup and documentation

- [x] Create ticket `LOUPE-JSVERBS-CLI`
- [x] Write an intern-oriented analysis and implementation guide
- [x] Write an investigation diary capturing the evidence and rationale
- [x] Upload the ticket bundle to reMarkable

## Scope definition

- [x] Confirm that jsverbs CLI embedding is feasible in loupedeck
- [x] Narrow the ticket scope to CLI embedding + docs/example tightening
- [x] Explicitly defer the other JS follow-ups from the earlier review

## Phase 1: extract reusable live-scene verb execution

- [ ] Refactor `cmd/loupedeck/cmds/run/command.go` so embedded commands can reuse the existing live scene session logic
- [ ] Introduce a reusable helper for invoking one parsed jsverb inside a caller-owned loupedeck runtime/session
- [ ] Keep `run --script ... --verb ...` behavior unchanged while sharing the new helper

## Phase 2: dynamic scene command bootstrap

- [ ] Add an early argument sniffing/bootstrap layer in `cmd/loupedeck/main.go` for the new scene command family
- [ ] Create a static parent command for embedded scene verbs (recommended: `scene`)
- [ ] Scan the selected script before Cobra registration and register embedded child commands for the entry-file verbs
- [ ] Provide a clear stub/help path when the scene command is invoked without enough information to scan

## Phase 3: loupedeck-specific jsverb command adapter

- [ ] Add a loupedeck command wrapper that uses `CommandDescriptionForVerb(...)` but executes through the live hardware/runtime path
- [ ] Reuse `scriptmeta` target scanning and engine option composition from the current run path
- [ ] Invoke selected verbs through `Registry.InvokeInRuntime(...)` inside the live scene session
- [ ] Decide whether `verbs help` remains public, becomes a debug helper, or is partially superseded by embedded commands

## Phase 4: docs and examples tightening

- [ ] Update help/docs to show the new embedded scene-command flow
- [ ] Keep `run --verb` documented as a fallback/compatibility path
- [ ] Use filename-oriented raw script examples consistently
- [ ] Tighten JS scene docs/examples so they do not imply shorthand or directory-first raw execution is the intended public UX

## Phase 5: tests and validation

- [ ] Add tests for early scene bootstrap and dynamic command registration
- [ ] Add tests for executing an embedded scene command against the annotated example
- [ ] Verify coexistence with `run`, `verbs`, `doc`, and `help`
- [ ] Verify help output quality for the new scene command family
- [ ] Run `go test ./...` before closing the ticket
- [ ] Run `docmgr doctor --ticket LOUPE-JSVERBS-CLI --stale-after 30` before final review
