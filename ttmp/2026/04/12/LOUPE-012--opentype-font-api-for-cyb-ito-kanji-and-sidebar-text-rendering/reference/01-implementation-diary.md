---
Title: Implementation diary
Ticket: LOUPE-012
Status: active
Topics:
    - javascript
    - rendering
    - fonts
    - unicode
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary for OpenType font loading and kanji rendering support in the Loupedeck JS runtime."
LastUpdated: 2026-04-12T18:15:00-04:00
WhatFor: "Use this diary to understand the implementation steps, failures, and validation path for LOUPE-012."
WhenToUse: "Use when continuing or reviewing the OpenType/CJK font support work."
---

# Diary

## Goal

Capture the implementation of a real font-loading API for the Loupedeck JS runtime so cyb-ito scenes can render proper kanji and sidebar text through the normal retained graphics pipeline.

## Step 1: Create the ticket, write the implementation plan, and define the phased task sequence

The current JS text path is already close to what we need technically: it rasterizes text into an alpha bitmap and then copies that into the grayscale `gfx.Surface`. The real blocker is the lack of a real font-loading API on the JS side, not the lack of a text rasterizer. Because of that, this work deserves a dedicated ticket focused on fonts rather than another ad hoc cyb-ito scene patch.

This first step creates the ticket, writes the plan, and defines the task order before any code changes. The implementation sequence is intentionally bottom-up: font loader first, JS API second, scene integration third.

### Prompt Context

**User prompt (verbatim):** "now create a new ticket to add font rendering API to allow us to properly render the sidebar and kanji from the original cyb-ito.html, and add a detailed implementation plan and add tasks to the ticket and then work on them one by one, committing ata ppropriate intervals and keeping a detailed diary"

**Assistant interpretation:** Create a dedicated ticket for OpenType/CJK font support in the JS runtime, write the design and task breakdown, and then start implementing the plan in focused slices with commits and diary updates.

**Inferred user intent:** The user wants proper kanji and sidebar text rendering treated as a real subsystem addition rather than a scene-local hack, with the same disciplined ticketing process used in the other Loupedeck work.

### What I did
- Created ticket `LOUPE-012`.
- Wrote the design/implementation plan.
- Created this diary.
- Decided on the initial implementation order:
  - Go-side font loader/cache
  - JS `gfx.font(...)` API
  - cyb-ito integration

### Why
- The work touches runtime/gfx, JS bindings, and scene code, so a ticketed approach is justified.
- The core problem is architectural/API-level, not just visual polish.

### What worked
- The plan is narrow and practical.
- The machine already has a likely useful system font path for CJK experiments:
  - `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`

### What didn't work
- N/A

### What I learned
- The real work here is exposing a proper `font.Face` into the current raster path rather than building a new renderer.

### What was tricky to build
- The main subtlety in planning is remembering that the repo has two text pipelines: the older root package already uses OpenType, while the JS `gfx` pipeline still uses `basicfont.Face7x13`. The ticket therefore has to focus on bridging that gap instead of rediscovering generic font rasterization from zero.

### What warrants a second pair of eyes
- The initial scope should stay focused on path-based font loading and JS font handles. It would be easy to drift into a larger text-layout system too early.

### What should be done in the future
- Implement the Go-side font loader/cache as the next code slice.

### Code review instructions
- Start with the design doc and check that the implementation order is bottom-up and does not overreach.

### Technical details
- Ticket path: `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/`
- Likely first useful font path on this machine: `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`
