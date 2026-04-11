---
Title: Backpressure-safe go-go-golems loupedeck package refactor
Ticket: LOUPE-003
Status: active
Topics:
    - loupedeck
    - go
    - serial
    - websocket
    - backpressure
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md
      Note: Primary architecture and implementation guide
    - Path: ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/reference/01-investigation-diary.md
      Note: Chronological record of ticket creation and analysis
ExternalSources: []
Summary: Active implementation ticket for the backpressure-safe go-go-golems Loupedeck package, including Phases 0-4, clean-exit hardware validation, and measured raw display throughput benchmarks.
LastUpdated: 2026-04-11T19:11:59-04:00
WhatFor: Track the package refactor that moved transport pacing and render control out of app-level experiments and now records the remaining reconnect and flow-control decision points with real hardware evidence.
WhenToUse: Use when orienting on LOUPE-003, finding the primary design doc, or reviewing completed implementation phases plus the remaining C decision gate.
---


# Backpressure-safe go-go-golems loupedeck package refactor

## Overview

LOUPE-003 is the follow-up ticket to the LOUPE-001 and LOUPE-002 experiments. Its purpose is to stop treating transport stability as an application-level workaround problem and instead build a real package, `github.com/go-go-golems/loupedeck`, that owns event fanout, lifecycle, outbound pacing, and eventually render coalescing.

The immediate implementation strategy is intentionally phased:

1. **B-lite first** — create the root package, fix lifecycle behavior, replace single-slot callbacks with multi-listener fanout, and add a single paced outbound writer.
2. **Then full B** — add keyed render invalidation and coalescing so widgets do not emit direct draw storms.
3. **Then evaluate C** — only if bounded, coalesced traffic still triggers protocol failures.

The primary design guide in this ticket is written for a new engineer and should be treated as the starting document for implementation.

## Key Links

- **Primary design doc**: `design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
- **Diary**: `reference/01-investigation-diary.md`
- **Previous feature-tester ticket**: `../LOUPE-002--loupedeck-live-feature-tester-comprehensive-hardware-exercise/`
- **Upstream reference implementation**: `/home/manuel/code/wesen/2026-04-11--loupedeck-test/sources/loupedeck-repo/`

## Status

Current status: **active**

Current completion state:
- Phases 0-4 implemented in the root package
- Root feature tester migrated and hardware-smoke-tested
- Clean-exit rerun validated with Circle-button shutdown
- Raw throughput benchmarks captured for full-screen and tile animation workloads
- Reconnect/reset hygiene still under investigation
- Strict C-style flow control still undecided

## Topics

- loupedeck
- go
- serial
- websocket
- backpressure

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design-doc/` - Architecture and implementation plans
- `reference/` - Diary and future quick-reference material
- `playbooks/` - Future operator/test procedures if needed
- `scripts/` - Ticket-local helper scripts if implementation work creates any
- `sources/` - Optional local evidence snapshots or external references
- `various/` - Scratch notes if needed
- `archive/` - Superseded artifacts
