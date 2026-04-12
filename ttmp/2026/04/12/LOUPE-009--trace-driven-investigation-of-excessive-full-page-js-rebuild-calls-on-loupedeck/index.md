---
Title: Trace-driven investigation of excessive full-page JS rebuild calls on Loupedeck
Ticket: LOUPE-009
Status: active
Topics:
    - loupedeck
    - goja
    - javascript
    - animation
    - rendering
    - benchmarking
    - performance
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js
      Note: Current full-page all-12 scene whose rebuild-call path motivated this investigation
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/anim/runtime.go
      Note: Defines the scene-global animation loop cadence currently suspected of over-producing rebuild requests
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/metrics/metrics.go
      Note: Existing collector that should gain bounded ordered trace support
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/jsmetrics/jsmetrics.go
      Note: Reusable JS binding layer where generic trace APIs should be added
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go
      Note: Live-runner surface where JS/Go trace output should be exposed and dumped
    - Path: /tmp/loupe-cyb-ito-full10-reasons-1776023397.log
      Note: Current strongest evidence log showing loop-dominant rebuild reasons in the no-input run
ExternalSources: []
Summary: Ticket for designing and later implementing trace instrumentation that explains where excessive full-page rebuild calls originate and how they correlate with renderer flushes and writer/device output.
LastUpdated: 2026-04-12T16:26:00-04:00
WhatFor: Track the focused follow-on investigation into rebuild-call origin and event ordering as a separate stream from the broader pacing-analysis ticket.
WhenToUse: Use when orienting on the rebuild-call tracing plan, expected evidence shapes, or future implementation of bounded JS/Go trace capture.
---

# Trace-driven investigation of excessive full-page JS rebuild calls on Loupedeck

## Overview

LOUPE-009 is a narrow follow-on ticket dedicated to one question: where are the many full-page rebuild calls actually coming from, and how do those calls line up against renderer flushes and hardware-visible output? Existing evidence from `LOUPE-007` already showed that rebuild reasons are dominated by the animation loop and that the writer queue stays calm, but counters and timing summaries still do not provide an ordered event timeline. This ticket defines the trace instrumentation needed to answer that question directly.

## Key Links

- **Main design guide**: `design/01-textbook-trace-driven-investigation-of-excessive-full-page-rebuild-calls.md`
- **Operational runbook**: `playbooks/01-trace-capture-runbook.md`
- **Implementation diary**: `reference/01-implementation-diary.md`
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Current completion state:
- Ticket created
- Main trace-analysis design guide written
- Operational trace-capture runbook written
- Diary created
- Task breakdown drafted
- Ticket-local reproducibility scripts archived under `scripts/` with numeric `XX-...` prefixes
- Generic trace collector substrate implemented
- Reusable JS trace bindings implemented
- Scene-level breadcrumb instrumentation implemented in the full-page all-12 workload
- Live-runner JS/Go trace dump controls implemented
- First no-input hardware trace captured and analyzed
- Current evidence says the full-page scene averages about `33.6` rebuilds per non-empty flush, with a median of `27` and a worst observed flush bucket of `119`
- Ticket validated with `docmgr doctor`
- ReMarkable bundle uploaded and verified

## Topics

- loupedeck
- goja
- javascript
- animation
- rendering
- benchmarking
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
