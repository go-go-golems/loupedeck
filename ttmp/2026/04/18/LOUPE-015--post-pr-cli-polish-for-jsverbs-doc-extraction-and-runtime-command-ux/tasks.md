# Tasks

## Ticket setup

- [x] Create ticket `LOUPE-015`
- [x] Write a design doc for the post-PR polish pass
- [x] Create an implementation diary for the work

## PR comment correctness fixes

- [x] Fix `BuildDocStore(...)` so file targets only extract docs from the selected file
- [x] Add regression coverage for file-scoped doc extraction
- [x] Fix `FindVerb(...)` so explicit selector lookups stay restricted to entry-file verbs for file targets
- [x] Add regression coverage for explicit file-scoped verb lookup
- [x] Record the PR comment rationale in the implementation diary

## Runtime command UX simplification

- [x] Make `loupedeck run` a plain runtime command instead of a structured Glazed result command
- [x] Remove `run` help/examples that advertise structured output
- [x] Keep `run` argument parsing working without introducing ad hoc flag duplication
- [x] Change the default session duration to `0s` and update the help text accordingly

## Generated verbs UX simplification

- [x] Remove structured-output UX from generated `loupedeck verbs ...` commands
- [x] Keep jsverbs metadata-driven argument parsing and session flag parsing intact
- [x] Ensure loupedeck session flags are shown in short help for generated verbs
- [x] Ensure the default duration for generated verbs is also `0s`
- [x] Add tests covering help visibility for session flags and the absence of old structured-output toggles

## Validation and ticket hygiene

- [x] Run focused tests while iterating
- [x] Run `go test ./...` in `/home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck`
- [x] Update changelog and diary with final implementation details
- [x] Run `docmgr doctor --ticket LOUPE-015 --stale-after 30`
