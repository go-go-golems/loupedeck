# Changelog

## 2026-04-18

- Initial workspace created.
- Added an intern-oriented design doc explaining why loupedeck can embed jsverbs as CLI commands, and later revised that plan so `loupedeck verbs ...` becomes the primary execution namespace instead of the earlier `scene`-parent proposal.
- Added an investigation diary recording the evidence from the current `loupedeck` root command, the current `run`/`verbs` split, and the upstream `jsverbs-example` embedding pattern.
- Recorded the implementation phases and docs/example tightening tasks for the follow-up ticket.
- Updated the ticket scope to drop the earlier backward-compatibility-oriented `run --verb` emphasis and instead scan configured roots to expose annotated commands directly under `loupedeck verbs ...`.
- Expanded `tasks.md` into a detailed implementation checklist covering repository/config modeling, bootstrap discovery, embedded + filesystem repository scanning, reusable live-scene execution helpers, dynamic Cobra registration, docs updates, and validation gates.
- Clarified the product decision that this ticket should be a clean cutover: remove `run --verb`, remove the old inspection-only `verbs list/help` flow, and avoid compatibility shims or wrapper-preserving logic.

## 2026-04-18

Created the follow-up ticket for jsverbs CLI embedding in loupedeck, documented why the feature is feasible, scoped it to embedded scene commands plus docs/example tightening, and recorded a phased implementation guide for a new intern.

### Related Files

- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/cmd/jsverbs-example/main.go — Reference pattern for dynamic jsverbs Cobra registration
- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/jsverbs/runtime.go — Reference split between ephemeral invoke and live-runtime InvokeInRuntime
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/run/command.go — Existing live hardware scene execution path that must stay authoritative
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/main.go — Static root command assembly that will need early dynamic scene bootstrap

