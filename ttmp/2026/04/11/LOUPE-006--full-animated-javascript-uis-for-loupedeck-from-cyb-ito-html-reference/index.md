---
Title: Full animated JavaScript UIs for Loupedeck from cyb-ito HTML reference
Ticket: LOUPE-006
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - animation
    - rendering
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html
      Note: Imported HTML canvas reference that defines the target animated scene style
    - Path: ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/design/01-textbook-full-animated-javascript-ui-runtime-from-cyb-ito-reference.md
      Note: Main intern-facing analysis, design, and implementation guide for this ticket
    - Path: runtime/js/module_ui/module.go
      Note: Current JS UI API that must grow from simple tiles to multi-region animated scene support
    - Path: runtime/render/visual_runtime.go
      Note: Current retained tile renderer that demonstrates the retained model but not yet full animated scenes
    - Path: renderer.go
      Note: Current invalidation scheduler that must remain below any new JS-driven animated scene layer
    - Path: writer.go
      Note: Current transport ownership layer that JavaScript must continue to avoid bypassing
ExternalSources:
    - local:cyb-ito.html
Summary: Ticket for expanding the current Loupedeck goja runtime into a full animated retained scene system capable of expressing the imported cyb-ito HTML canvas reference without giving JavaScript raw transport ownership.
LastUpdated: 2026-04-11T23:31:00-04:00
WhatFor: Track the analysis, design, and eventual implementation of multi-region animated JavaScript UIs on Loupedeck hardware.
WhenToUse: Use when orienting on the cyb-ito-inspired animated UI work, reading the imported source artifact, or planning the next runtime expansion beyond simple retained text/icon tiles.
---


# Full animated JavaScript UIs for Loupedeck from cyb-ito HTML reference

## Overview

LOUPE-006 is the next expansion step after the first working goja runtime in `LOUPE-005`. The imported `cyb-ito.html` source is a procedural animated canvas scene with a `4×3` grid of `90×90` tiles, animated side strips, ripple overlays, and touch-reactive scene behavior. This ticket exists to turn that reference into a proper Loupedeck-native design and implementation plan.

The central architecture rule is that JavaScript still should **not** own raw rendering or transport. Instead, the runtime should grow by adding retained display regions, Go-owned graphics surfaces, and layered composition while preserving the current writer and invalidation ownership already established in the Go package.

## Key Links

- **Main design guide**: `design/01-textbook-full-animated-javascript-ui-runtime-from-cyb-ito-reference.md`
- **Implementation diary**: `reference/01-implementation-diary.md`
- **Imported source**: `sources/local/cyb-ito.html`
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Current completion state:
- Ticket created
- `cyb-ito.html` imported as a tracked ticket source artifact
- Source analyzed carefully
- Detailed intern-facing analysis/design/implementation guide written
- Initial diary entry written
- Ticket validated with `docmgr doctor`
- Design bundle uploaded to reMarkable and verified remotely
- Phase B first slice complete: retained JS-facing display regions now exist for `left`, `main`, and `right`, and the live runner can flush all three retained display targets

## Topics

- loupedeck
- goja
- javascript
- animation
- rendering

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
