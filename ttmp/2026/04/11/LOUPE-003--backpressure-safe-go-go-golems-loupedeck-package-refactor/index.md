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
Summary: Planning ticket for turning the current Loupedeck experiments into a real backpressure-safe go-go-golems package, starting with B-lite and then full B.
LastUpdated: 2026-04-11T22:12:00-04:00
WhatFor: Track the package refactor that will move transport pacing and render control out of app-level experiments and into a reusable module.
WhenToUse: Use when orienting on LOUPE-003, finding the primary design doc, or reviewing the implementation sequence for B-lite, B, and the later C decision gate.
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
- Ticket created
- Detailed design/implementation guide written
- Diary started
- Bookkeeping partially complete
- Validation and reMarkable upload pending

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
