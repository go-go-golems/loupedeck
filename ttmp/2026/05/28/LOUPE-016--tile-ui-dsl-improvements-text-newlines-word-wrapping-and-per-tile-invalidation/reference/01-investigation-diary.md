---
Title: Investigation Diary
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
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - runtime/gfx/text.go
    - runtime/gfx/surface.go
    - runtime/ui/tile.go
    - runtime/ui/display.go
    - runtime/ui/ui.go
    - runtime/render/visual_runtime.go
    - runtime/js/module_ui/module.go
    - runtime/js/module_gfx/module.go
    - runtime/present/runtime.go
    - runtime/reactive/runtime.go
    - runtime/reactive/signal.go
ExternalSources: []
Summary: "Chronological investigation diary for LOUPE-016 tile UI DSL improvements"
LastUpdated: 2026-05-28T08:30:00-04:00
WhatFor: "Track investigation process, findings, and decisions"
WhenToUse: "When reviewing what was investigated and why decisions were made"
---

# Investigation Diary — LOUPE-016

## 2026-05-28 — Initial Codebase Exploration

### What I did

- Read the entire runtime layer: `runtime/ui/`, `runtime/gfx/`, `runtime/render/`, `runtime/js/`, `runtime/present/`, `runtime/reactive/`, `runtime/host/`, `runtime/anim/`
- Read all JS module bridges: `module_ui`, `module_gfx`, `module_present`, `module_state`, `module_anim`
- Read all example scripts: `01-hello.js` through `12-documented-scene.js`
- Read the API reference doc and existing LOUPE-008 and LOUPE-013 tickets
- Read all Go test files in gfx, render, ui packages

### Key findings

1. **Text newline bug**: `gfx.Surface.Text()` (`runtime/gfx/text.go`) calls `font.Drawer.DrawString(text)` on a single line. The `font.Drawer` in Go's `golang.org/x/image/font` package treats the entire string as one line — `\n` characters are not handled. They render as missing-glyph boxes or are simply ignored, causing visual corruption.

2. **No word wrapping**: The `TextOptions` struct has `Width` and `Height` fields that are used only for the temporary alpha mask dimensions, not for wrapping. Text that exceeds the surface or specified width simply overflows/clips — there is no line-breaking logic anywhere in the stack.

3. **Whole-page invalidation for per-tile updates**: The current architecture has two rendering paths:
   - **Retained tile mode** (e.g., `01-hello.js`): Uses `tile.text()`, `tile.icon()`. The renderer in `render/visual_runtime.go` already supports per-tile dirty tracking via `DirtyTiles()` and `ClearDirtyTiles()`. When a tile's text changes, only that tile is re-rendered and sent. This path **already works for per-tile invalidation**.
   - **Surface mode** (e.g., `11-cyb-os-tiles.js`): Uses `display.surface(mainSurface)` to draw the entire 360×270 framebuffer manually from JS. When any pixel changes, the entire display is marked dirty via `display.markDirty()`. The `present.onFrame()` callback redraws everything, and `present.invalidate()` causes a full re-render. **This is where the slowness lives** — there is no way to invalidate a sub-rect of a display-level surface.

4. **The present runtime** (`runtime/present/runtime.go`) is a simple invalidate-on-demand loop. It calls `render(reason)` then `flush()`. The render function is set by `present.onFrame()` which is a single JS callback. The flush function is set by the host to call `renderer.Flush()`. There is no concept of "which region changed" — it's all-or-nothing.

5. **Two coexisting UI paradigms create confusion**:
   - The *retained tile DSL* (`tile.text()`, `tile.icon()`) where Go owns rendering
   - The *imperative surface drawing* (`gfx.surface()`, `display.surface()`) where JS owns rendering
   - These are not composited — a tile with a surface set skips the retained text/icon rendering entirely

### Architecture understanding confirmed

The data flow for a timer-driven tile update (e.g., clock showing seconds) currently works like this for the **surface path**:

```
anim.loop() → JS callback → modifies signal → reactive flush → present.invalidate()
→ present loop wakes → render("loop") → JS onFrame() redraws ALL surfaces
→ flush() → renderer.Flush() → DirtyDisplays() finds main display dirty
→ renderDisplay() → full 360×270 image → Draw() → send to hardware
```

For the **retained tile path**:

```
anim.loop() → JS callback → modifies signal → reactive flush
→ tile.BindText closure re-runs → tile.SetText() → markDirtyTile()
→ UI.SetDirtyHandler → present.invalidate("ui-dirty")
→ present loop wakes → render("ui-dirty") → JS onFrame() (usually not set)
→ flush() → renderer.Flush() → DirtyTiles() finds only changed tiles
→ renderTile() for each dirty tile → 90×90 image → Draw() at tile offset
→ send to hardware
```

The retained path is **already per-tile**. The problem is that when users want custom rendering (not just text/icon), they fall off the retained path and into the surface path, which is all-or-nothing.

### Decisions so far

- The design doc needs to explain both paths clearly and propose how to bridge them
- Per-tile surface drawing is the key missing feature
- Text wrapping and newline handling should work in both the retained renderer and the gfx surface text API

## 2026-05-28 — Key Discovery: tile.surface() Already Exists

### What I found

While writing the design doc, I discovered that `tile.surface()` is **already implemented** in the JS bridge:

- `runtime/js/module_ui/module.go` line 322: `obj.Set("surface", ...)` for tile objects
- `runtime/ui/tile.go`: `SetSurface()` with `OnChange` listener that calls `t.markDirty()`
- `runtime/render/visual_runtime.go`: `renderTile()` already checks `tile.Surface()` and renders it

This means **per-tile invalidation with custom surfaces already works**. The problem is purely:
1. No documentation of this feature
2. No examples using this pattern
3. No ergonomic helper (`tile.draw()`) for the common case

### Implication for the design

Design 3 (per-tile invalidation) is largely about documentation and ergonomics, not new infrastructure. The real implementation work is:
1. Add `tile.draw(fn)` as a convenience method
2. Add `tile.invalidate()` as an explicit dirty-marking method
3. Write examples and documentation
4. The underlying Go code already does the right thing

## 2026-05-28 — Design Document Written

### What I did

- Wrote the comprehensive design doc (`design/01-tile-ui-dsl-improvements-comprehensive-analysis-and-implementation-guide.md`)
- 960 lines covering:
  - Executive summary
  - System overview with hardware layout diagram
  - Architecture deep dive (rendering pipeline, JS DSL layer, reactive runtime, present loop)
  - Three problem analyses (newlines, wrapping, invalidation)
  - Three implementation designs with pseudocode and Go code sketches
  - Current and proposed API references
  - File reference map
  - Test plan
  - Migration and backward compatibility analysis
  - Data flow diagrams for all three rendering paths
  - Go type reference appendix

### Probe scripts written

- `scripts/01-text-newline-probe.js`: Tests multi-line text in retained tiles
- `scripts/02-per-tile-invalidation-probe.js`: Demonstrates retained tile mode per-tile invalidation

### Next steps

1. ~~Commit the current state~~ ✓ Done
2. ~~Upload to reMarkable~~ ✓ Done — uploaded as "LOUPE-016 Tile UI DSL Improvements.pdf" to /ai/2026/05/28
3. ~~Update ticket index and tasks~~ ✓ Done
4. Implementation of the three features (see tasks.md)

## 2026-05-28 — Documentation and Examples Pass

### What I did

- Rewrote the API reference (`docs/help/topics/01-loupedeck-js-api-reference.md`) with:
  - New "Two rendering paths" section explaining retained vs surface vs per-tile surface paths
  - Full `loupedeck/gfx` module documentation (surface, font, all drawing methods, text options)
  - Full `loupedeck/present` module documentation (invalidate, onFrame)
  - `tile.surface()` documentation with performance guidance
  - Display object documentation (text, icon, visible, surface, layer, tile)
  - `page.display()` documentation
  - Updated module overview table with gfx and present
  - New example patterns: per-tile surface clock, full-display surface, hybrid layers
  - New troubleshooting entries for newline text, long text overflow, slow full-display scenes
  - Updated "Current limitations" to mention LOUPE-016 newline/wrapping bugs

- Rewrote the tutorial (`docs/help/tutorials/01-build-your-first-live-loupedeck-js-script.md`) with:
  - New Step 6: Custom pixel content with per-tile surfaces
  - "Two rendering paths explained" section with data flow diagrams for all three paths
  - Complete per-tile surface pulse-meter example
  - Updated example table to include new examples
  - Updated troubleshooting table with per-tile surface and text entries
  - Fixed stale repository path references

- Created new examples:
  - `examples/js/13-per-tile-clock.js`: 12-tile clock with per-tile surfaces, anim.loop, batch(), demonstrates per-tile invalidation
  - `examples/js/14-tile-surface-drawing.js`: 12-tile showcase of all gfx primitives (fillRect, line, text, crosshatch, set/add, compositeAdd, alignment, brightness gradient, borders, checkerboard, animated wave, animated pulse rings)

- Updated `examples/js/02-counter-button.js`: minor cleanup

- Re-uploaded combined docs bundle (API reference + tutorial + design doc) to reMarkable as "LOUPE-016 Tile UI DSL - API Reference, Tutorial, and Design.pdf"

## 2026-05-28 — External Script Improvements

### What I did

Improved all 5 scripts in the goja-ast-analysis loupedeck-code-nav directory to use per-tile surfaces instead of the broken `tile.text("LINE1\nLINE2")` pattern.

**Key improvements across all scripts:**
1. Replaced `tile.text("A\nB")` with per-tile `gfx.surface(90, 90)` + `surface.text()` calls — works around the newline rendering bug
2. Two-line labels: category label on top (dimmer), value on bottom (bright) — much more readable
3. Accent bars: each tile gets a colored top bar indicating its role (bright = interactive, dim = static)
4. Per-tile invalidation: only the tiles that actually change are re-rendered when a knob turns
5. `surface.batch()` used consistently for grouped drawing operations
6. Static tiles drawn once (control labels, hints) instead of reactively
7. `drawParamTile()` / `drawStaticTile()` helpers reduce code duplication

**Per-script changes:**
- `01-log-interactions.js`: Counter tile gets a progress bar. Touched tiles highlight with bright accent. Per-tile surfaces show touch number + last interaction type.
- `02-query-state-console.js`: Three rows — parameter tiles (reactive), knob assignment tiles (static), action tiles. Parameter tiles show KIND/methods, SCOPE/package, FILTER/exported, LIMIT/10.
- `03-ast-query-console.js`: Same layout as 02 but with real AST queries. Knob handlers only redraw the parameter tile they affect, not all tiles.
- `04-result-tiles.js`: The big winner — result tiles show receiver + name on separate lines, file + line number clearly. Mode tile highlighted with bright accent. Touched tiles highlight briefly.
- `05-web-remote-browser.js`: Info row shows package, item index, kind+label, file+line. Static control row. Status tile with two-line status display.

## 2026-05-30 — Step 6: Implement Text Newline Support

Implemented the first feature from the design doc: multi-line text rendering via `\n` splitting. This covers tasks #6–#10.

### Prompt Context

**User prompt (verbatim):** "go ahead. keep a diary as you work. commit at appropriate intervals, check tasks off as you go."

**Assistant interpretation:** Begin implementing the LOUPE-016 features in order, starting with text newline support. Keep the investigation diary updated, commit at logical boundaries, and check off tasks.

**Inferred user intent:** Execute the implementation plan from the design doc, maintaining full traceability.

**Commit (code):** 95a46a1 — "LOUPE-016: Implement text newline support in Surface.Text() and drawCenteredLabel()"

### What I did
- Added `LineGap` field to `gfx.TextOptions` in `runtime/gfx/text.go`
- Refactored `Surface.Text()` to detect `\n` in the text and render each line separately with proper vertical offsets
- Extracted `renderLine()` as the single-line rendering function (original `Text()` logic)
- Added `splitLines()` helper that preserves empty lines for vertical spacing
- Modified `drawCenteredLabel()` in `runtime/render/visual_runtime.go` to split on `\n` and render each line
- Extracted `drawSingleLine()` helper for the renderer
- Added `lineGap` property parsing in `runtime/js/module_gfx/module.go` `textOptionsFromValue()`
- Wrote comprehensive unit tests in `runtime/gfx/text_test.go`: 8 new tests covering multi-line, line gap, empty lines, edge cases
- Saved verification script `scripts/03-verify-text-newline.sh` to ticket

### Why
The Go `font.Drawer.DrawString()` treats the entire string as one line — `\n` characters render as missing-glyph boxes. Users need multi-line text in both the surface API and the retained tile renderer.

### What worked
- Splitting `Text()` into `Text()` (dispatcher) + `renderLine()` (single line) was clean
- For multi-line, setting `lineOpts.Height = lineH + 4` per line gave consistent positioning
- The `drawCenteredLabel()` change was straightforward — same pattern as `Surface.Text()`

### What didn't work
- First attempt used the original `opts.Height` for each line's alpha mask in multi-line mode — caused the second line to be vertically centered in a too-tall mask, pushing pixels outside the test's expected region. Fixed by overriding `lineOpts.Height = lineH + 4` per line.
- Initial test `TestSurfaceTextNewlineRendersMultipleLines` used a fixed row-45 split boundary, but Face7x13 has `Height.Ceil() = 13`, so two lines only span ~26 rows. Rewrote test to use the actual `lineH` as the separator.

### What I learned
- Face7x13 metrics: Height=13, Ascent=11, Descent=2
- When auto-calculating `Height` (height <= 0 → lineH + 4), the baseline calculation `h/2` can exceed the actual text extent — this works for single-line centering but is wrong for multi-line where we want each line to occupy exactly one line-height

### What was tricky to build
- The `Height` field serves two purposes: alpha mask size and vertical centering computation. For single-line centering, a larger Height is fine. For multi-line layout, each line needs its Height set to exactly `lineH + 4` so the baseline calculation produces consistent per-line positioning.
- The renderer's `drawCenteredLabel()` uses a `baseline` parameter that's an absolute Y position in the destination image, not relative — so multi-line layout there is just `baseline + i*lineH`.

### What warrants a second pair of eyes
- The `lineOpts.Height = lineH + 4` override in the multi-line path — is 4px the right margin? It matches the original single-line auto-height but may be too tight for larger fonts.
- The `drawCenteredLabel()` uses `lineH` (font height) with no gap between lines — should we add a configurable gap in the retained renderer too? Currently the tile renderer doesn't have a `LineGap` option.

### What should be done in the future
- Add `LineGap` support to the retained tile renderer's `drawCenteredLabel()` (currently only the surface `Text()` API supports it)
- Consider adding a `WrapWidth` field for word wrapping (next step in the plan)

### Code review instructions
- Start with `runtime/gfx/text.go`: new `LineGap` field, `splitLines()`, `renderLine()`, modified `Text()` dispatcher
- Then `runtime/render/visual_runtime.go`: modified `drawCenteredLabel()`, new `drawSingleLine()`
- Then `runtime/js/module_gfx/module.go`: `LineGap` line in `textOptionsFromValue()`
- Run: `go test ./runtime/gfx/ -count=1 -v -run "TestSurfaceText|TestSplitLines"`

### Technical details
- `splitLines()` uses `strings.Split(text, "\n")` — preserves empty lines
- Multi-line Y layout: `opts.Y + i*(lineH+gap)` where `lineH = face.Metrics().Height.Ceil()`
- Per-line Height override: `lineH + 4` (4px margin below descent line)
- Face7x13: Height=13, so two lines with gap=0 span 26 rows, three lines span 39 rows
