---
Title: Layered full-frame effects on the presenter-driven cyb-ito runtime
Ticket: LOUPE-011
Status: active
Topics:
    - javascript
    - rendering
    - animation
    - performance
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Track the next cyb-ito milestone: reintroduce multiple software-composited full-frame effect layers while preserving the presenter-driven single-flush model."
LastUpdated: 2026-04-12T17:19:49.724539411-04:00
WhatFor: "Use this ticket when implementing richer layered visual effects for the presenter-driven full-page cyb-ito runtime."
WhenToUse: "Use when working on software-composited layered effects, scanlines, HUD overlays, chrome caching, and other full-frame rendering improvements for the cyb-ito full-page scene."
---

# Layered full-frame effects on the presenter-driven cyb-ito runtime

## Overview

This ticket captures the next step after `LOUPE-010`: bring back richer animated full-frame layering for the cyb-ito full-page scene without regressing the presenter-driven pacing model that fixed the rebuild storm.

The target model is:
- multiple logical software layers inside the JS scene
- one atomically composed final `360x270` frame
- one presenter-driven full-page hardware flush per presented frame

## Key Links

- Design plan: [design/01-implementation-plan-for-layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime.md](./design/01-implementation-plan-for-layered-full-frame-effects-on-the-presenter-driven-cyb-ito-runtime.md)
- Diary: [reference/01-implementation-diary.md](./reference/01-implementation-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**

Current completion state:
- Ticket created
- Implementation plan written
- Detailed task list written
- Diary created
- First layered full-page code slice implemented
- Full-page scene refactored into internal software layers (`base`, `chrome`, `scene`, `fx`, `hud`, final `frame`)
- Presenter-driven single full-page flush model preserved
- First FX slice added: scanlines, grain/noise, and active-tile sweep/ripple overlays
- Color-tinted display-layer support added for overlays
- Selected tile now uses a red accent layer
- A large touch-triggered spiral ripple overlay now spans the whole screen
- `go test ./...` passed after the accent/ripple runtime extension and scene update
- Hardware smoke validation was previously successful for the layered compositor slice; the new red/ripple slice is ready for user verification via the archived interactive run script

## Topics

- javascript
- rendering
- animation
- performance

## Structure

- design/ - architecture and implementation planning
- reference/ - chronological diary and supporting reference material
- playbooks/ - hardware validation procedures
- scripts/ - archived commands used during implementation and validation
- sources/ - imported external artifacts if needed later
- various/ - working notes and auxiliary material
- archive/ - retired drafts and old artifacts
