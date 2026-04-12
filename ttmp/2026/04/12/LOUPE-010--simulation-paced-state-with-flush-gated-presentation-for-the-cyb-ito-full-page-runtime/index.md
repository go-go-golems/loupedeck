---
Title: Simulation-paced state with flush-gated presentation for the cyb-ito full-page runtime
Ticket: LOUPE-010
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - animation
    - rendering
    - performance
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js
      Note: Full-page scene to migrate away from loop-driven redraws
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go
      Note: Live-runner presentation path to refactor around one-frame-in-flight presentation
    - Path: /tmp/loupe-cyb-ito-full10-trace-1776025944.log
      Note: Trace baseline motivating the refactor
ExternalSources: []
Summary: Ticket for replacing loop-driven full-page redraws with simulation-paced state updates and flush-gated one-frame-in-flight presentation.
LastUpdated: 2026-04-12T17:08:00-04:00
WhatFor: Track the forward-only presenter refactor needed to make the cyb-ito full-page runtime architecturally correct.
WhenToUse: Use when implementing or reviewing the new simulation/presentation model.
---

# Simulation-paced state with flush-gated presentation for the cyb-ito full-page runtime

## Overview

LOUPE-010 is the forward-only refactor ticket for the cyb-ito full-page runtime. The goal is to stop using the animation loop as the direct trigger for full-page redraws and replace it with the correct model: simulation-paced state changes plus flush-gated one-frame-in-flight presentation.

## Key Links

- **Main implementation plan**: `design/01-implementation-plan-simulation-paced-state-with-flush-gated-presentation.md`
- **Implementation diary**: `reference/01-implementation-diary.md`
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Current completion state:
- Ticket created
- Main implementation plan written
- Diary created
- Detailed phased task breakdown written
- Ticket-local reproducibility scripts archived under `scripts/`
- Pure-Go presenter runtime implemented
- JS environment and `loupedeck/present` module implemented
- Live runner refactored to use presenter-driven frame production
- Full-page all-12 scene migrated to `loupedeck/present`
- New no-input hardware trace captured and analyzed under the new architecture
- Rebuilds-per-non-empty-flush collapsed from the old traced average of about `33.6` to `1.0` in the new presenter-driven trace

## Topics

- loupedeck
- goja
- javascript
- animation
- rendering
- performance

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
