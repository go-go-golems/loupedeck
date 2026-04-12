---
Title: Implementation plan for layered full-frame effects on the presenter-driven cyb-ito runtime
Ticket: LOUPE-011
Status: active
Topics:
    - javascript
    - rendering
    - animation
    - performance
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Plan the next cyb-ito runtime step: reintroduce multiple logical scene layers while preserving the presenter-driven single full-frame flush model that fixed the rebuild storm."
LastUpdated: 2026-04-12T17:20:07.173201699-04:00
WhatFor: "Use this document when implementing layered full-page visual effects on top of the presenter-driven cyb-ito runtime without regressing into multi-rebuild flush storms."
WhenToUse: "Use when adding scanlines, noise, ripple overlays, HUD elements, cached chrome, or similar full-frame effects to the current cyb-ito full-page JavaScript scene."
---

# Implementation plan for layered full-frame effects on the presenter-driven cyb-ito runtime

## Goal

Bring back richer animated full-frame effects for the `10-cyb-ito-full-page-all12.js` scene now that `LOUPE-010` has established the correct simulation/presentation model. The new work should preserve the presenter-driven architecture:

- simulation updates state independently
- presentation is flush-gated
- one presenter-driven render per visible frame
- one final full-page hardware flush per presented frame
- repeated invalidations coalesce to the latest state instead of causing a rebuild storm

## Non-goals

This ticket does **not** try to preserve the old direct `renderAll()`-from-loop model. There is no backwards-compatibility goal for the pre-presenter full-page scene architecture.

This ticket also does **not** try to reintroduce multiple device-visible flushes. The output contract remains one final `360x270` full-page frame per presentation step.

## Current starting point

The current full-page scene already works well enough on hardware with:

- `loupedeck/present`
- a single `main` retained surface attached to the UI display
- one rebuild per non-empty flush in the new trace evidence
- smooth enough motion when the run is started with a much more aggressive writer configuration such as `--send-interval 0ms`

The missing piece is scene richness. The current script draws everything into one surface in one pass. That works, but it makes it awkward to reason about:

- static chrome vs dynamic art
- interaction overlays vs persistent scene content
- per-frame post effects such as scanlines and grain
- future caching opportunities for layers that do not need a full redraw every frame

## Correct architecture for this step

The correct next model is:

1. Maintain multiple **logical software layers** as `gfx.surface(...)` values inside the JS scene.
2. Build or rebuild those layers during presenter-driven frame production.
3. Composite the layers into one final full-page surface.
4. Flush only that final full-page surface through the existing presenter-driven path.

In other words:

- many layers in memory
- one composed frame
- one hardware-visible full-page flush

## Proposed initial layer set

### 1. Base layer
Purpose:
- quiet static background
- faint tile-region shading
- subtle global structure that does not depend on time or input

This layer should be cacheable and rebuilt only when the scene structure changes.

### 2. Scene-art layer
Purpose:
- the 12 cyb-ito tile renderers
- the core procedural tile art that changes with `phase`

This layer remains the main animated content layer.

### 3. Chrome layer
Purpose:
- borders
- active-tile glow
- labels/dividers

This layer may depend on the active tile, but it is conceptually separate from the art.

### 4. FX layer
Purpose:
- scanlines
- grain/noise
- active-tile ripple or sweep
- subtle whole-frame modulation

This is where the “multiple layers again” goal really lives, but it should be implemented as software composition, not separate output flushes.

### 5. HUD layer
Purpose:
- status text
- active tile marker
- interaction breadcrumbs such as the last event

### 6. Final frame surface
Purpose:
- compositing target attached to `page.display("main", display => display.surface(frame))`

## Composition order

Recommended first-pass order:

1. base
2. scene-art
3. chrome
4. FX
5. HUD

The final frame should be built atomically with `frame.batch(() => { ... })`.

## First implementation slice

The first implementation slice should be intentionally narrow and hardware-oriented:

- refactor the full-page script to use multiple internal surfaces
- preserve the current presenter-driven `present.onFrame(...)` model
- add a subtle first FX pass (scanlines + noise + active tile overlay)
- keep one final full-page flush
- validate on hardware with the fluid `--send-interval 0ms` run style

This is enough to prove that the new layered model works without regressing the main pacing fix from `LOUPE-010`.

## Future slices after the first one

Once the first layered full-page slice is stable, good next directions are:

- cached static base/chrome invalidation instead of rebuilding every layer every frame
- scene-wide sweep/glitch passes
- stronger active-tile ripples/selection pulses
- side-strip reintroduction on the same composed-frame model
- more faithful post-processing inspired by the original `cyb-ito.html` scene

## Validation approach

### Code validation
- `go test ./...`

### Hardware validation
Run the live runner with the full-page scene and aggressive writer pacing:

```bash
go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration 60s \
  --send-interval 0ms
```

If needed, also enable render/writer stats while tuning:

```bash
go run ./cmd/loupe-js-live \
  --script ./examples/js/10-cyb-ito-full-page-all12.js \
  --duration 20s \
  --send-interval 0ms \
  --stats-interval 2s \
  --log-render-stats \
  --log-writer-stats
```

## Success criteria

The first pass is successful if:

- the scene still uses the presenter-driven full-page model
- the device still shows fluid motion on hardware
- multiple internal logical layers are present in the script
- the scene regains visibly richer frame effects
- the ticket contains a reproducible plan, tasks, scripts, and diary updates
