# Changelog

## 2026-04-18

- Initial workspace created.
- Added an intern-oriented design doc explaining why loupedeck can embed jsverbs as CLI commands, but should do so through a loupedeck-specific hardware/runtime adapter rather than upstream runtime-owning generated commands.
- Added an investigation diary recording the evidence from the current `loupedeck` root command, the current `run`/`verbs` split, and the upstream `jsverbs-example` embedding pattern.
- Recorded the implementation phases and docs/example tightening tasks for the follow-up ticket.

## 2026-04-18

Created the follow-up ticket for jsverbs CLI embedding in loupedeck, documented why the feature is feasible, scoped it to embedded scene commands plus docs/example tightening, and recorded a phased implementation guide for a new intern.

### Related Files

- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/cmd/jsverbs-example/main.go — Reference pattern for dynamic jsverbs Cobra registration
- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/jsverbs/runtime.go — Reference split between ephemeral invoke and live-runtime InvokeInRuntime
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/run/command.go — Existing live hardware scene execution path that must stay authoritative
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/main.go — Static root command assembly that will need early dynamic scene bootstrap

