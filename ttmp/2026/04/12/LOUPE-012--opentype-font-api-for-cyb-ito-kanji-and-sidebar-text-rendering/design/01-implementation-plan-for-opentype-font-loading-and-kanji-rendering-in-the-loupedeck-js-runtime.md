---
Title: Implementation plan for OpenType font loading and kanji rendering in the Loupedeck JS runtime
Ticket: LOUPE-012
Status: active
Topics:
    - javascript
    - rendering
    - fonts
    - unicode
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/js/07-cyb-ito-prototype.js
      Note: First cyb-ito example using the new CJK font API
    - Path: runtime/gfx/font.go
      Note: New Go-side font loading/caching layer designed and implemented under this ticket
    - Path: runtime/js/module_gfx/module.go
      Note: JS font API bridge and text font option
ExternalSources: []
Summary: Plan the addition of a JS-facing OpenType font API so cyb-ito scenes can render proper kanji and sidebar text through the retained gfx surface pipeline.
LastUpdated: 2026-04-12T18:15:00-04:00
WhatFor: Use this document when implementing real font loading and CJK-capable text rendering for the Loupedeck JS runtime.
WhenToUse: Use when adding `gfx.font(...)`, font caching, kanji rendering, or cyb-ito sidebar/title text support.
---


# Implementation plan for OpenType font loading and kanji rendering in the Loupedeck JS runtime

## Goal

Add a JS-facing font API that allows the retained Loupedeck scene runtime to load real OpenType/TrueType fonts and use them in `gfx.surface(...).text(...)`. The immediate product goal is to render proper kanji and sidebar text from the original `cyb-ito.html` reference instead of falling back to ASCII-only placeholders or `?` glyphs.

## Current problem

The current JS text path is structurally correct but font-limited.

- `runtime/js/module_gfx/module.go` exposes `surface.text(...)`
- `runtime/gfx/text.go` rasterizes text into an `image.Alpha` and copies it into the grayscale `gfx.Surface`
- but `module_gfx` hardcodes `basicfont.Face7x13`

So the current scene pipeline can only reliably render tiny bitmap-font glyphs. That is sufficient for ASCII labels and test UI, but not for the kanji and scrolling sidebar text from the source `cyb-ito.html` artifact.

## Important architectural constraint

This ticket does **not** change the presentation or transport model.

The text/font work must stay inside the existing raster pipeline:

1. load font in Go
2. create a `font.Face`
3. rasterize glyphs into a bitmap in Go
4. copy bitmap intensity into `gfx.Surface`
5. keep the existing retained surface -> renderer -> RGB565 -> Loupedeck pipeline unchanged

The device should still only receive pixels.

## Existing useful foundations

The repo already has most of the lower-level pieces needed:

- `runtime/gfx/text.go` already rasterizes text through `font.Drawer`
- `loupedeck.go` already demonstrates OpenType loading with `opentype.Parse(...)` and `opentype.NewFace(...)`
- the current JS `gfx` API already separates text content from text options
- the current scene model already supports retained/cached surfaces, which is useful for font-heavy scene elements like sidebars

## Proposed JS API

### Minimal API

```javascript
const gfx = require("loupedeck/gfx");

const horror = gfx.font("/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc", {
  size: 14,
  dpi: 72,
  index: 0,
});

surface.text("渦", {
  x: 4,
  y: 4,
  width: 20,
  height: 18,
  brightness: 220,
  center: true,
  font: horror,
});
```

### Notes

- `path` is explicit and should point to a local `.ttf`, `.otf`, or `.ttc`
- `size` is in typographic points, matching `opentype.FaceOptions`
- `dpi` defaults to a sensible value if omitted
- `index` selects a face from a font collection like `.ttc`

## Go-side design

### 1. Add a font-loading abstraction in `runtime/gfx`

Add something like:

- `runtime/gfx/font.go`

Possible types/functions:

- `type LoadedFont struct { ... }`
- `func LoadFont(path string, opts FontOptions) (*LoadedFont, error)`
- `func (f *LoadedFont) Face() font.Face`
- `func (f *LoadedFont) Close() error` (only if needed)

### 2. Support TTF/OTF and TTC

The implementation should use:

- `opentype.Parse(...)` for regular single-font files
- `opentype.ParseCollection(...)` for `.ttc` collections

This is important because a system CJK font on this machine is available at:

- `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`

So the implementation should not assume a single-face font file only.

### 3. Add caching

Font loading should be cached by at least:

- path
- size
- dpi
- index
- hinting (if exposed)

Reason:

- repeated `gfx.font(...)` calls in JS should not repeatedly parse and instantiate the same face
- the scene system may redraw often, but font parsing should not happen per frame

### 4. Expose font handles to JS

In `runtime/js/module_gfx/module.go`:

- add `gfx.font(path, opts)`
- return a JS object with a hidden Go payload like `__font`
- teach `textOptionsFromValue(...)` to inspect `opts.font` and pull out the underlying loaded face

### 5. Keep `runtime/gfx/text.go` mostly unchanged

This file is already doing the right thing conceptually:

- allocate an alpha bitmap
- rasterize with `font.Drawer`
- copy alpha into grayscale surface pixels

Once the `Face` is no longer forced to `basicfont.Face7x13`, the same raster path should work for kanji-capable fonts.

## Immediate cyb-ito target

Once the generic API exists, the first concrete scene win should be to update one cyb-ito example to use real kanji text.

Suggested first target:

- the tile title labels in `examples/js/10-cyb-ito-full-page-all12.js`

Suggested second target:

- the sidebar/scroller text inspired by `cyb-ito.html`

## Recommended implementation order

## Phase A: ticket/docs/setup
- create ticket
- write plan
- write diary
- define tasks

## Phase B: Go-side font loader and cache
- add `runtime/gfx/font.go`
- support `.ttf`, `.otf`, `.ttc`
- add unit tests for cache hits and collection loading

## Phase C: JS `gfx.font(...)` API
- expose font handles from `runtime/js/module_gfx/module.go`
- allow `surface.text(..., { font })`
- add JS runtime tests proving a font handle survives the bridge

## Phase D: first cyb-ito integration
- update the relevant example scene to load a CJK-capable font
- replace placeholder ASCII labels with actual kanji for tile titles
- optionally add first sidebar text experiment

## Phase E: validation/docs
- run `go test ./...`
- add ticket-local scripts for testing
- update diary/changelog/index/tasks
- capture hardware/user validation once the scene is runnable with kanji

## Success criteria

The first milestone is successful if:

- JS can load a real font via `gfx.font(...)`
- `surface.text(...)` can use that font
- at least one cyb-ito example renders proper kanji glyphs through the normal `gfx` pipeline
- the retained scene/render/presenter/transport stack remains unchanged outside the font/text layer

## Anti-patterns to avoid

- Do not special-case kanji by hardcoding one-off glyph draw paths into the renderer.
- Do not make the device aware of fonts or Unicode; keep the output pixel-based.
- Do not parse fonts per frame.
- Do not bypass `runtime/gfx/text.go` with an unrelated rendering subsystem unless that becomes unavoidable later.
