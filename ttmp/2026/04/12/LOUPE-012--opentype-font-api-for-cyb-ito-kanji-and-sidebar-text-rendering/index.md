---
Title: OpenType font API for cyb-ito kanji and sidebar text rendering
Ticket: LOUPE-012
Status: active
Topics:
    - javascript
    - rendering
    - fonts
    - unicode
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Track the addition of a JS-facing OpenType font API for proper kanji and sidebar text rendering in cyb-ito scenes."
LastUpdated: 2026-04-12T18:15:00-04:00
WhatFor: "Use this ticket when implementing real font loading and CJK-capable text rendering in the Loupedeck JS runtime."
WhenToUse: "Use when working on `gfx.font(...)`, kanji labels, or sidebar text support for cyb-ito scenes."
---

# OpenType font API for cyb-ito kanji and sidebar text rendering

## Overview

This ticket captures the addition of a JS-facing real-font API for the Loupedeck retained graphics runtime. The immediate goal is to support kanji labels and sidebar text from the original `cyb-ito.html` reference through the normal `gfx` text rasterization pipeline.

## Key Links

- Design plan: [design/01-implementation-plan-for-opentype-font-loading-and-kanji-rendering-in-the-loupedeck-js-runtime.md](./design/01-implementation-plan-for-opentype-font-loading-and-kanji-rendering-in-the-loupedeck-js-runtime.md)
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
- Go-side font loader/cache implemented in `runtime/gfx`
- `.ttf/.otf` and `.ttc` collection loading supported
- JS `gfx.font(path, opts)` API implemented
- `surface.text(..., { font })` support implemented with `basicfont.Face7x13` fallback
- JS runtime tests now prove both generic font-handle use and collection-font kanji rendering
- First cyb-ito integration completed in `examples/js/07-cyb-ito-prototype.js`
- The prototype now uses a CJK font to render actual kanji tile labels and a kanji sidebar scroller
- `go test ./...` passes after the first three implementation phases

## Topics

- javascript
- rendering
- fonts
- unicode
