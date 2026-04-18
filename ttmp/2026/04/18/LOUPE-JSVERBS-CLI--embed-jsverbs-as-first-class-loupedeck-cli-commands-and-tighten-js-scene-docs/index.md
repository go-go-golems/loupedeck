---
Title: Embed jsverbs as first-class loupedeck CLI commands and tighten JS scene docs
Ticket: LOUPE-JSVERBS-CLI
Status: active
Topics:
    - loupedeck
    - javascript
    - goja
    - cli
    - documentation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupedeck/cmds/run/command.go
      Note: Existing hardware-owned scene execution path that future embedded commands must reuse
    - Path: cmd/loupedeck/main.go
      Note: Static root command assembly that frames the follow-up ticket
ExternalSources: []
Summary: Follow-up ticket focused on exposing annotated jsverbs scene entrypoints as first-class loupedeck CLI commands, using a hardware-aware command adapter rather than upstream runtime-owning jsverbs commands, and tightening the related JS docs/examples.
LastUpdated: 2026-04-18T11:30:09.59772471-04:00
WhatFor: 'Use this ticket when implementing the next CLI UX step after LOUPE-JSVERBS: dynamically embedded scene commands plus the corresponding docs/example cleanup.'
WhenToUse: Open this workspace when you need the intern handoff docs, task checklist, or changelog for the jsverbs CLI embedding follow-up.
---



# Embed jsverbs as first-class loupedeck CLI commands and tighten JS scene docs

## Overview

This ticket picks up the next UX step after `LOUPE-JSVERBS`.

The previous ticket made annotated scripts work correctly on the shared `go-go-goja` runtime and added `run --verb`, `verbs list/help`, and `doc`. This follow-up asks whether loupedeck can go one step further and expose annotated scene verbs as first-class CLI commands in the style of `go-go-goja/cmd/jsverbs-example`.

The answer is yes, but with one important constraint: loupedeck must keep control of the live hardware/runtime session. So this ticket focuses on a loupedeck-specific embedding layer that reuses jsverbs metadata and live-runtime invocation APIs without reusing the upstream ephemeral runtime-owning command wrappers directly.

The same ticket also includes the postponed docs/example tightening work related to the JS scene flow.

## Key Links

- Design doc: `design-doc/01-analysis-and-implementation-guide-for-embedding-jsverbs-as-loupedeck-cli-commands.md`
- Investigation diary: `reference/01-investigation-diary-for-jsverbs-cli-embedding-follow-up.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **active**

## Scope

### In scope

- first-class CLI embedding of annotated scene verbs
- reuse of `CommandDescriptionForVerb(...)` + `InvokeInRuntime(...)` through a loupedeck-specific adapter
- docs/example tightening related to the new flow

### Deferred

- generic JS error-reporting polish
- doc server / `--serve`
- multi-script registry work
- richer interactive verb value UX

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.
