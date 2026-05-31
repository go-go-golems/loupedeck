---
Title: Tasks
Ticket: LOUPE-016
Status: active
---

# Tasks

## Design & Analysis

- [x] Read and analyze the complete rendering pipeline (reactive → UI → present → renderer → hardware)
- [x] Identify the three bugs/limitations (newlines, wrapping, invalidation)
- [x] Discover that tile.surface() already exists in the JS bridge
- [x] Write comprehensive design document with pseudocode and implementation guide
- [x] Write probe scripts for testing

## Implementation — Text Newlines

- [x] Add `LineGap` field to `gfx.TextOptions` in `runtime/gfx/text.go`
- [x] Modify `Surface.Text()` to split on `\n` and render multiple lines
- [x] Modify `drawCenteredLabel()` in `runtime/render/visual_runtime.go` to split on `\n`
- [x] Add `lineGap` property parsing in `runtime/js/module_gfx/module.go`
- [x] Write unit tests for multi-line text rendering
- [ ] Verify with probe script `01-text-newline-probe.js`

## Implementation — Word Wrapping

- [x] Add `wrapText()` function to `runtime/gfx/text.go`
- [x] Add `WrapWidth` field to `gfx.TextOptions`
- [x] Integrate `wrapText()` into `Surface.Text()`
- [x] Add `drawWrappedLabel()` to `runtime/render/visual_runtime.go`
- [x] Add `wrapWidth` property parsing in `runtime/js/module_gfx/module.go`
- [x] Add `tile.text(valueOrFn, { wrap: true })` support in `runtime/js/module_ui/module.go`
- [x] Write unit tests for word wrapping
- [x] Write integration test script for wrapped text

## Implementation — Per-Tile Invalidation Ergonomics

- [ ] Add `tile.draw(fn)` method to `runtime/ui/tile.go`
- [ ] Add `tile.draw(fn)` JS bridge in `runtime/js/module_ui/module.go`
- [ ] Add `tile.invalidate()` method to `runtime/ui/tile.go`
- [ ] Add `tile.invalidate()` JS bridge in `runtime/js/module_ui/module.go`
- [ ] Export `module_gfx.surfaceObject()` as public helper or create shared function
- [ ] Write example script `13-per-tile-clock.js`
- [ ] Update API reference documentation

## Documentation

- [ ] Update `docs/help/topics/01-loupedeck-js-api-reference.md` with:
  - `tile.surface()` documentation
  - `tile.draw()` documentation
  - `tile.invalidate()` documentation
  - `surface.text()` `lineGap` and `wrapWidth` options
  - Multi-line text support
- [ ] Update `docs/help/tutorials/01-build-your-first-live-loupedeck-js-script.md` with per-tile pattern
