---
Title: Virtual Loupedeck browser simulator for JS verbs prototyping and demos
Ticket: LOUPE-014
Status: active
Topics:
    - loupedeck
    - web
    - simulation
    - javascript
    - jsverbs
    - ui
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Browser-driven virtual Loupedeck research ticket focused on simulating the device without serial hardware so JS verbs can be prototyped and demoed interactively."
LastUpdated: 2026-04-18T13:59:03-04:00
WhatFor: "Track the architecture, design, and implementation plan for a virtual Loupedeck backend plus web UI simulator that reuses the existing JS runtime and retained UI stack."
WhenToUse: "Use when orienting a new engineer to the simulator work, reviewing the virtual device/backend seam, or planning the browser-based JS verb demo flow."
---

# Virtual Loupedeck browser simulator for JS verbs prototyping and demos

## Overview

LOUPE-014 explores a browser-driven virtual Loupedeck that can run the existing JS runtime, expose simulated controls, and render the current device state without requiring a physical serial connection. The ticket is intended to make JS verb prototyping and demos easier while keeping the real hardware path available for validation.

## Key Links

- **Design doc 01**: [Virtual Loupedeck simulation architecture and implementation guide](./design-doc/01-virtual-loupedeck-simulation-architecture-and-implementation-guide.md)
- **Design doc 02**: [Independent review and revised implementation plan for the virtual Loupedeck simulator](./design-doc/02-independent-review-and-revised-implementation-plan-for-the-virtual-loupedeck-simulator.md)
- **Investigation diary**: [Investigation diary](./reference/01-investigation-diary.md)
- **Tasks**: [tasks.md](./tasks.md)
- **Changelog**: [changelog.md](./changelog.md)

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- loupedeck
- web
- simulation
- javascript
- jsverbs
- ui

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
