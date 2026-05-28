---
Title: 'Tile UI DSL Improvements: Text Newlines, Word Wrapping, and Per-Tile Invalidation'
Ticket: LOUPE-016
Status: active
Topics:
    - javascript
    - goja
    - ui
    - tiles
    - rendering
    - text
    - performance
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: runtime/gfx/surface.go
      Note: Grayscale pixel buffer with drawing primitives and change listeners
    - Path: runtime/gfx/text.go
      Note: THE TEXT RENDERER - single-line DrawString
    - Path: runtime/js/module_gfx/module.go
      Note: JS bridge for loupedeck/gfx - surface.text() options parsing
    - Path: runtime/js/module_ui/module.go
      Note: JS bridge for loupedeck/ui - tile.surface() already at line 322
    - Path: runtime/present/runtime.go
      Note: Invalidate-on-demand presentation loop
    - Path: runtime/reactive/runtime.go
      Note: Reactive dependency tracking and batch flush
    - Path: runtime/reactive/signal.go
      Note: Mutable reactive value with notification
    - Path: runtime/render/visual_runtime.go
      Note: Retained renderer - drawsCenteredLabel
    - Path: runtime/ui/display.go
      Note: Display struct with surface/layers/tiles
    - Path: runtime/ui/tile.go
      Note: Tile struct with text/icon/visible/surface properties and per-tile dirty tracking
    - Path: runtime/ui/ui.go
      Note: UI orchestrator with dirty tile/display sets
ExternalSources: []
Summary: Fix text newline rendering, add word wrapping, and improve per-tile invalidation ergonomics in the Loupedeck tile UI DSL
LastUpdated: 2026-05-28T09:00:00-04:00
WhatFor: Guide implementation of three tile UI improvements for the Loupedeck JS DSL
WhenToUse: When implementing text wrapping, newline rendering, or per-tile invalidation features
---


# Tile UI DSL Improvements: Text Newlines, Word Wrapping, and Per-Tile Invalidation

## Overview

This ticket addresses three issues with the Loupedeck tile UI DSL:

1. **Newlines in text are not rendered** — `font.Drawer.DrawString()` treats the entire string as one line, so `\n` characters render as missing-glyph boxes or are ignored
2. **No text word wrapping** — long text overflows tile boundaries instead of wrapping at word boundaries within the available width
3. **Full-display invalidation is slow** — when using display-level surfaces, any pixel change re-renders the entire 360×270 display; per-tile surfaces already work but are undocumented

**Key discovery:** Per-tile surfaces with per-tile dirty tracking already exist in both Go and the JS bridge. The problem is documentation, ergonomics, and example migration, not missing infrastructure.

## Documents

### Design
- [01 — Comprehensive Analysis and Implementation Guide](./design/01-tile-ui-dsl-improvements-comprehensive-analysis-and-implementation-guide.md) — **Read this one.** Full intern-ready analysis with architecture, pseudocode, API design, file references, and test plan

### Reference
- [01 — Investigation Diary](./reference/01-investigation-diary.md) — Chronological research process and findings

### Scripts
- [01 — Text Newline Probe](./scripts/01-text-newline-probe.js) — JS test script for multi-line text behavior
- [02 — Per-Tile Invalidation Probe](./scripts/02-per-tile-invalidation-probe.js) — JS test script comparing retained vs surface invalidation

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active** — analysis and design complete, implementation not started

## Topics

- javascript
- goja
- ui
- tiles
- rendering
- text
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
