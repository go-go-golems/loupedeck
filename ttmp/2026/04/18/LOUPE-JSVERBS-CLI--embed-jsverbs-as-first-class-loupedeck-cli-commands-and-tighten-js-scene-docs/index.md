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
Summary: Follow-up ticket focused on making `loupedeck verbs ...` the primary execution namespace for annotated jsverbs scene commands discovered from configured roots, while keeping execution on the hardware-owned runtime path and tightening the related JS docs/examples.
LastUpdated: 2026-04-18T11:30:09.59772471-04:00
WhatFor: 'Use this ticket when implementing the revised CLI UX after LOUPE-JSVERBS: `loupedeck verbs ...` as the annotated-scene execution tree plus the corresponding docs/example cleanup.'
WhenToUse: Open this workspace when you need the intern handoff docs, task checklist, or changelog for the jsverbs CLI embedding follow-up.
---



# Embed jsverbs as first-class loupedeck CLI commands and tighten JS scene docs

## Overview

This ticket picks up the next UX step after `LOUPE-JSVERBS`.

The previous ticket made annotated scripts work correctly on the shared `go-go-goja` runtime and shipped a transitional split between `run --verb`, `verbs list/help`, and `doc`.

This follow-up replaces that transitional shape with a clean cutover. The intended end state is that annotated scene commands execute directly under `loupedeck verbs ...`, for example `loupedeck verbs documented configure`, while `run` returns to being the plain-file runner and the old wrapper/inspection surfaces are removed.

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

- first-class CLI embedding of annotated scene verbs under `loupedeck verbs ...`
- reuse of `CommandDescriptionForVerb(...)` + `InvokeInRuntime(...)` through the native loupedeck hardware execution path
- repository discovery so all annotated scripts can be exposed
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
