---
Title: Post-PR CLI polish for jsverbs, doc extraction, and runtime command UX
Ticket: LOUPE-015
Status: active
Topics:
    - loupedeck
    - javascript
    - jsverbs
    - cli
    - documentation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupedeck/cmds/run/command.go
      Note: Current plain-script runner still exposes structured Glazed behavior that this ticket will simplify
    - Path: cmd/loupedeck/cmds/run/session.go
      Note: Shared loupedeck session settings/defaults and help visibility live here
    - Path: cmd/loupedeck/cmds/verbs/command.go
      Note: Dynamic annotated-verb command generation and current structured-output wrappers
    - Path: pkg/scriptmeta/scriptmeta.go
      Note: Contains the two PR-commented scope bugs for doc extraction and explicit verb lookup
ExternalSources:
    - https://github.com/go-go-golems/loupedeck/pull/1
Summary: Follow-up ticket for small-but-important post-PR fixes: constrain per-file doc/verb scope correctly, simplify run and verb Cobra UX, expose loupedeck session flags in short help, default duration to infinite, and drop unnecessary structured-output framing for runtime commands.
LastUpdated: 2026-04-18T20:05:00-04:00
WhatFor: Use this ticket when reviewing or continuing the post-PR fixes around script scoping, command help, and loupedeck runtime command UX.
WhenToUse: Open this workspace when you need the rationale, implementation diary, and validation notes for the PR-follow-up cleanup pass.
---

# Post-PR CLI polish for jsverbs, doc extraction, and runtime command UX

## Overview

This ticket captures the immediate cleanup pass after the large jsverbs CLI embedding work.

It covers two correctness bugs called out on PR #1 and several product-shape tweaks requested afterward:

- `doc --script <file>` must extract docs only from that file
- explicit verb lookup for a file target must stay scoped to verbs declared in that file
- `run` should be a plain Cobra command, not a structured Glazed result command
- generated `verbs` commands should expose loupedeck session flags in short help
- the default duration should be `0s` (run until interrupted) everywhere
- runtime-facing `verbs` commands do not need structured-output UX or glazed settings sections

## Key Links

- Design doc: `design-doc/01-analysis-and-implementation-guide-for-post-pr-jsverbs-cli-polish.md`
- Implementation diary: `reference/01-implementation-diary-for-post-pr-jsverbs-cli-polish.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **implemented; pending final doctor check and optional commit/bookkeeping follow-up**

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.
