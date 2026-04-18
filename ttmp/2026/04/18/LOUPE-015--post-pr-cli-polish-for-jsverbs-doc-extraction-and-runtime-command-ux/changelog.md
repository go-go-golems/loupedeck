# Changelog

## 2026-04-18

- Initial workspace created.
- Added a focused follow-up design doc for the post-PR polish pass.
- Scoped the ticket to two correctness fixes from PR #1 and four runtime command UX simplifications:
  - file-scoped doc extraction
  - file-scoped explicit verb lookup
  - plain `run` command behavior
  - bare runtime `verbs` behavior
  - visible session flags in short help
  - infinite-duration defaults
- Implemented the PR-commented scope fixes in `pkg/scriptmeta`:
  - `BuildDocStore(...)` now limits file targets to the selected file
  - `FindVerb(...)` now keeps explicit selector resolution inside `EntryVerbs(...)` for file targets
- Simplified runtime command UX:
  - `run` is now a plain bare command
  - generated `verbs` commands no longer expose Glazed structured-output toggles or command-settings sections
  - loupedeck session flags are visible directly in command help
  - default session duration is now `0s`
- Added/updated tests for:
  - file-scoped doc extraction
  - file-scoped explicit verb lookup
  - runtime command help without structured-output toggles
  - generated verb help showing session flags and `0s` defaults
- Validated with:
  - `go test ./cmd/loupedeck/cmds/run ./cmd/loupedeck/cmds/verbs ./pkg/scriptmeta`
  - `go test ./...`
