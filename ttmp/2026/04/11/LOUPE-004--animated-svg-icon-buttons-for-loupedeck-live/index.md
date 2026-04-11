---
Title: Animated SVG icon buttons for Loupedeck Live
Ticket: LOUPE-004
Status: active
Topics:
    - loupedeck
    - go
    - svg
    - animation
    - rendering
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupe-svg-buttons/main.go
      Note: Root command for rendering the animated SVG-backed button grid on hardware
    - Path: svg_icons.go
      Note: Core loader and rasterization path for the imported icon library
    - Path: ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/design-doc/01-animated-svg-icon-button-rendering-plan.md
      Note: Primary design plan for the SVG extraction and animation work
    - Path: ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/reference/01-implementation-diary.md
      Note: Chronological implementation record for the ticket
ExternalSources:
    - local:macos1-icon-library.html
Summary: Ticket for importing an HTML SVG icon library, extracting/normalizing its icons in Go, and rendering properly scaled animated buttons on the Loupedeck Live, including a working root demo command and hardware validation.
LastUpdated: 2026-04-11T19:24:30-04:00
WhatFor: Track the implementation and documentation for animated SVG-backed touch-button rendering on the Loupedeck Live.
WhenToUse: Use when orienting on the SVG renderer work, locating the imported icon source, or running the hardware demo.
---

# Animated SVG icon buttons for Loupedeck Live

## Overview

LOUPE-004 builds on the new root `github.com/go-go-golems/loupedeck` package by adding a higher-level visual demo: animated touch-button icons sourced from an imported HTML SVG library. The ticket covers source provenance, SVG extraction/normalization, Go-side rasterization, tile scaling, and the final hardware demo command.

## Key Links

- **Design plan**: `design-doc/01-animated-svg-icon-button-rendering-plan.md`
- **Diary**: `reference/01-implementation-diary.md`
- **Imported icon library**: `sources/local/macos1-icon-library.html`

## Status

Current status: **active**

Current completion state:
- Ticket created and source library imported into docmgr
- SVG extraction/normalization/rasterization implemented
- Root animated-button demo command added
- Demo run on actual hardware
- Lifecycle warnings still inherited from the lower-level package

## Topics

- loupedeck
- go
- svg
- animation
- rendering

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design-doc/` - Architecture and implementation planning
- `reference/` - Chronological diary and quick references
- `sources/` - Imported asset sources and later evidence
- `scripts/` - Ticket-local helpers if needed
- `archive/` - Superseded artifacts
