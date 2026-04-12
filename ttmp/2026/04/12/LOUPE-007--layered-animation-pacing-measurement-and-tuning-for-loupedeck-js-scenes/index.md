---
Title: Layered animation pacing measurement and tuning for Loupedeck JS scenes
Ticket: LOUPE-007
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
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go
      Note: Future instrumentation surface for renderer and writer statistics during real JS scene runs
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go
      Note: Retained display composition layer where flush timing and display-level work can be measured
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/display.go
      Note: Current layered display model whose base surfaces and named overlays define scene density
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/writer.go
      Note: Go-owned write queue and pacing layer that must be observed under layered load
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go
      Note: Existing raw hardware throughput benchmark used as the control baseline
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js
      Note: Current layered prototype scene that should become the main density-sweep workload
ExternalSources: []
Summary: Ticket for designing and later implementing instrumentation and runbooks that measure how layered retained JS scenes affect renderer pacing, writer pressure, and perceived responsiveness on real Loupedeck hardware.
LastUpdated: 2026-04-12T07:01:00-04:00
WhatFor: Track the analysis, design, and future implementation of layered-scene pacing measurement and tuning work as a separate stream from the active cyb-ito implementation ticket.
WhenToUse: Use when orienting on future pacing instrumentation, scene-density sweeps, or the distinction between raw transport ceilings and retained-scene behavior.
---

# Layered animation pacing measurement and tuning for Loupedeck JS scenes

## Overview

LOUPE-007 is the follow-on ticket for performance measurement and tuning strategy around the new layered retained-scene runtime. The immediate goal is not optimization for its own sake. The goal is to build a reliable framework for answering whether a slowdown is caused by JavaScript scene updates, Go-side composition, writer queue pressure, or hardware transport limits.

This ticket exists separately from `LOUPE-006` so that active cyb-ito scene implementation can continue without losing the future measurement and interpretation plan.

## Key Links

- **Main design guide**: `design/01-textbook-measuring-layered-animation-density-pacing-and-tuning-for-loupedeck-js-scenes.md`
- **Project technical report**: `design/02-project-technical-report-performing-the-12-tile-javascript-canvas-cyb-ito-port.md`
- **Operational runbook**: `playbooks/01-layered-density-measurement-runbook.md`
- **Implementation diary**: `reference/01-implementation-diary.md`
- **Related Files**: See frontmatter RelatedFiles field

## Status

Current status: **active**

Current completion state:
- Ticket created
- Main intern-facing design and implementation guide written
- Operational runbook written
- Diary created
- Task breakdown drafted
- Full project technical report written for the 12-tile cyb-ito performance investigation
- First concrete runtime precondition slice implemented: retained `gfx.Surface` batching now exists so full-page scene construction can be made frame-atomic before real pacing instrumentation starts
- First live instrumentation slice implemented: `cmd/loupe-js-live` now supports periodic renderer stats, writer stats, and JS-scene metrics logging, and the JS runtime now exposes reusable `metrics` / `scene-metrics` modules for scene-side timing
- Follow-up hardware evidence captured after adding rebuild-reason metrics; the measured rebuild stream is overwhelmingly loop-driven (`scene.renderAll.reason.loop`) with only one startup `initial` rebuild in the fresh no-input run
- Fresh evidence continues to show a calm writer queue but highly variable and often very slow full-page flush windows, reinforcing that the current full-page bottleneck is upstream of queue buildup
- Ticket validation rerun complete after the new report addition
- Updated reMarkable bundle uploaded and verified, including the new project technical report

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
