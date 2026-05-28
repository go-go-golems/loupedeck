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
