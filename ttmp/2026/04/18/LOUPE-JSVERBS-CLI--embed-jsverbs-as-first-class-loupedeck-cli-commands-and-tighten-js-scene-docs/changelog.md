# Changelog

## 2026-04-18

- Initial workspace created.
- Added an intern-oriented design doc explaining why loupedeck can embed jsverbs as CLI commands, and later revised that plan so `loupedeck verbs ...` becomes the primary execution namespace instead of the earlier `scene`-parent proposal.
- Added an investigation diary recording the evidence from the current `loupedeck` root command, the current `run`/`verbs` split, and the upstream `jsverbs-example` embedding pattern.
- Recorded the implementation phases and docs/example tightening tasks for the follow-up ticket.
- Updated the ticket scope to drop the earlier backward-compatibility-oriented `run --verb` emphasis and instead scan configured roots to expose annotated commands directly under `loupedeck verbs ...`.
- Expanded `tasks.md` into a detailed implementation checklist covering repository/config modeling, bootstrap discovery, embedded + filesystem repository scanning, reusable live-scene execution helpers, dynamic Cobra registration, docs updates, and validation gates.
- Clarified the product decision that this ticket should be a clean cutover: remove `run --verb`, remove the old inspection-only `verbs list/help` flow, and avoid compatibility shims or wrapper-preserving logic.
- Recorded the upstream `go-go-goja` prerequisite commits that unblocked the downstream implementation:
  - `ad6e30b` — `jsverbs: add pluggable command invokers`
  - `9f2c797` — `docaccess: update glazed help section integration`
  - `4cd7c11` — `runtimeowner: decouple OwnerContext from concrete runner`
- Implemented the loupedeck-side cutover:
  - added an embedded built-in scripts repository
  - added repository discovery from app config, env, and repeated `--verbs-repository`
  - added duplicate full-path collision detection
  - extracted shared live-scene session helpers into `cmd/loupedeck/cmds/run/session.go`
  - removed `run --verb`
  - replaced the static `verbs list/help` implementation with a dynamic command tree under `loupedeck verbs ...`
  - wired generated verbs through the upstream pluggable invoker API back into the live hardware-owned loupedeck runtime/session
  - updated public help/docs to reflect the clean cutover
- Added a reusable reference note documenting the lazy command pattern behind dynamic verb loading:
  - why the `verbs` namespace should bootstrap on demand instead of during root startup
  - how the placeholder-to-resolved-command handoff works
  - why help/output adoption is required to preserve rich CLI UX
  - where the same pattern is applicable to other dynamic namespaces, plugin trees, or repository-backed verb systems
- Validated the implementation with:
  - `go test ./cmd/loupedeck/cmds/run ./cmd/loupedeck/cmds/verbs ./pkg/scriptmeta`
  - `go test ./...`

## 2026-04-18

Created the follow-up ticket for jsverbs CLI embedding in loupedeck, documented why the feature is feasible, scoped it to embedded scene commands plus docs/example tightening, and recorded a phased implementation guide for a new intern.

### Related Files

- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/cmd/jsverbs-example/main.go — Reference pattern for dynamic jsverbs Cobra registration
- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/jsverbs/runtime.go — Reference split between ephemeral invoke and live-runtime InvokeInRuntime
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/run/command.go — Existing live hardware scene execution path that must stay authoritative
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/main.go — Static root command assembly that will need early dynamic scene bootstrap

