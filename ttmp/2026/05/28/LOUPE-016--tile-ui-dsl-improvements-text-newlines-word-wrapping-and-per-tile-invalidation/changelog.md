---
Title: Changelog
Ticket: LOUPE-016
Status: active
---

# Changelog

## 2026-05-28 — Initial Analysis and Design

- Created ticket LOUPE-016 for tile UI DSL improvements
- Read entire rendering pipeline: reactive → UI → present → renderer → hardware
- Identified three bugs: newline rendering, text wrapping, full-display invalidation
- **Key discovery**: `tile.surface()` already exists in JS bridge (module_ui/module.go line 322), per-tile dirty tracking already works in Go
- Wrote comprehensive design document (55KB, ~960 lines) covering:
  - Architecture deep dive with data flow diagrams
  - Three problem analyses with visual diagrams
  - Three implementation designs with Go pseudocode
  - Current and proposed API references
  - File reference map with complexity estimates
  - Test plan with unit test sketches
  - Migration and backward compatibility analysis
- Wrote two probe scripts for testing:
  - `scripts/01-text-newline-probe.js` — multi-line text in retained tiles
  - `scripts/02-per-tile-invalidation-probe.js` — retained vs surface invalidation
- Wrote investigation diary documenting research process

## 2026-05-28 — Documentation and Examples Pass

- Rewrote API reference (01-loupedeck-js-api-reference.md):
  - New "Two rendering paths" section (retained, display surface, per-tile surface)
  - Full loupedeck/gfx module docs (surface, font, drawing methods, text options)
  - Full loupedeck/present module docs (invalidate, onFrame)
  - tile.surface() docs with performance guidance
  - Display object docs (text, icon, visible, surface, layer, tile)
  - page.display() docs
  - New example patterns (per-tile clock, full-display scene, hybrid)
  - New troubleshooting entries
  - Updated limitations for LOUPE-016
- Rewrote tutorial (01-build-your-first-live-loupedeck-js-script.md):
  - New Step 6: Custom pixel content with per-tile surfaces
  - Data flow diagrams for all three rendering paths
  - Complete per-tile surface example
  - Updated troubleshooting
- Created examples/js/13-per-tile-clock.js: 12-tile clock with per-tile surfaces
- Created examples/js/14-tile-surface-drawing.js: 12-tile gfx primitives showcase
- Updated 02-counter-button.js
- Re-uploaded combined docs bundle to reMarkable
