---
Title: Animated SVG icon button rendering plan
Ticket: LOUPE-004
Status: active
Topics:
    - loupedeck
    - go
    - svg
    - animation
    - rendering
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/sources/local/macos1-icon-library.html
      Note: Imported source library containing the inline SVG icon set to extract and render
ExternalSources:
    - local:macos1-icon-library.html
Summary: Plan for extracting SVG icons from the imported System 1.0 icon library, rasterizing them safely in Go, scaling them to 90×90 touch-button tiles, and animating them on the Loupedeck Live.
LastUpdated: 2026-04-11T19:11:59-04:00
WhatFor: Provide the implementation plan for loading the imported icon library and rendering animated SVG buttons on actual Loupedeck Live hardware.
WhenToUse: Use when implementing or reviewing the SVG button renderer and demo command.
---

# Animated SVG icon button rendering plan

## Executive Summary

This ticket adds a new end-to-end capability on top of the root `github.com/go-go-golems/loupedeck` package: load an imported HTML icon library containing inline SVGs, extract individual icons from it, rasterize them in Go, scale them cleanly to the Loupedeck Live’s `90×90` touch-button tiles, and animate a full 4×3 bank of buttons on the real device.

The design deliberately avoids trying to execute the HTML page’s original CSS animation model inside Go. Instead, it treats the imported HTML file as an SVG asset library. The plan is to extract the static icon geometry from each tile, normalize the SVG markup so a Go rasterizer can consume it, render crisp base sprites, trim excess transparent padding, and then apply lightweight runtime animation transforms such as pulse, bob, blink, and invert on the device-facing images.

## Problem Statement

The imported icon library is not a directory of standalone `.svg` files. It is an HTML document containing:

- CSS custom properties such as `var(--white)` and `var(--black)`
- one global hidden `<svg>` block with shared dither-pattern defs
- many inline `<svg>` fragments inside `.icon-cell` tiles
- animation styles written for the browser (`animation: ...`)

That format is excellent for browsing in a browser, but it is not directly usable by the Loupedeck package. The package needs `image.Image` values sized for device tiles, and the command must update those images at animation-friendly rates without depending on a browser runtime.

## Proposed Solution

### 1. Preserve the source in docmgr

Use the imported file stored in:

- `ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/sources/local/macos1-icon-library.html`

This keeps the asset provenance explicit and makes the ticket self-contained.

### 2. Extract icon entries from HTML

Implement a loader that reads the HTML file and extracts, for each `.icon-cell`:

- the inline `<svg>...</svg>` fragment
- the visible icon label from `.icon-label`

The loader should also extract:

- root CSS color variables (currently `--white` and `--black`)
- the shared `<defs>` block for dither patterns

### 3. Normalize SVG fragments for Go-side rasterization

Each inline SVG should be normalized before rasterization:

- replace `var(--white)` / `var(--black)` with actual colors
- inject the shared `<defs>` block into the icon SVG when needed
- replace browser-only class-based dither usage with explicit `fill="url(#ditherXX)"`
- strip browser animation declarations from `style="..."`
- ensure the root SVG contains the XML namespace if needed

The goal is not to preserve the browser animation model. The goal is to preserve the static geometry and fill/stroke appearance faithfully enough for device rendering.

### 4. Rasterize SVGs into base sprites

Use a Go SVG rasterizer to render each normalized icon into a transparent RGBA image. The initial sprite size can stay close to the source viewbox geometry (for example `48×48` or a modest multiple), because the device-facing animation layer will perform final placement/scaling.

### 5. Trim transparent bounds and scale to fit tiles

To make the icons feel properly scaled on `90×90` Loupedeck tiles, compute the non-transparent bounds of each rasterized sprite and scale the trimmed sprite to fit within a consistent inner box, preserving aspect ratio. This prevents icons with large empty margins in their original SVG viewbox from appearing too small on-device.

### 6. Animate at the image-composition layer

Do not attempt to run browser CSS keyframes. Instead, compose each `90×90` frame in Go using:

- off-white / black button styling consistent with the imported library
- nearest-neighbor sprite scaling to keep the pixel-art feel
- per-icon phase offsets and lightweight transforms such as:
  - pulse (slight scale oscillation)
  - bob (vertical motion)
  - blink/invert accents
  - border emphasis

### 7. Provide a runnable root command

Add a new command that:

- connects to the Loupedeck Live
- loads the imported HTML library
- selects 12 icons
- renders them to the `4×3` main touchscreen grid
- animates them until the user exits with the Circle button

## Design Decisions

### Decision: Treat the HTML file as an asset library, not as an executable browser scene

**Why:** Running the original CSS animation model inside Go would require an embedded browser or a much more complex SVG/CSS animation engine. That is unnecessary for the actual device goal, which is animated icon buttons.

### Decision: Scale based on trimmed alpha bounds, not raw viewbox bounds alone

**Why:** Proper visual scaling on `90×90` tiles matters more than preserving the original page’s internal whitespace. Trimmed-bounds scaling gives more uniform button presence across heterogeneous icons.

### Decision: Use lightweight animation transforms over pre-rasterized icon sprites

**Why:** The device only needs short, crisp per-frame image updates. Applying simple transforms to cached sprites is cheaper and easier to reason about than reinterpreting browser keyframes each frame.

### Decision: Keep the command configurable via a library path flag

**Why:** The imported docmgr path should be the default reference, but a flag makes experimentation easier and keeps the loader reusable.

## Alternatives Considered

### Alternative A: Screenshot the HTML in a headless browser

Rejected because it introduces a browser dependency, complicates deployment, and mixes page layout behavior with icon extraction.

### Alternative B: Reimplement every original CSS animation from the source HTML

Rejected because it is not necessary for the ticket goal. The Loupedeck needs animated buttons, not a perfect browser emulation of the HTML page.

### Alternative C: Manually copy a handful of SVGs into standalone files

Rejected because it loses the provenance and breadth of the imported icon library and does not satisfy the request to import the full library via docmgr.

## Implementation Plan

1. Replace LOUPE-004 templates with real ticket docs and import the HTML library.
2. Add SVG extraction/normalization code and tests.
3. Add SVG rasterization support and tests.
4. Add tile-scaling and sprite-composition helpers.
5. Add a root command to animate a 12-button icon bank on the Loupedeck Live.
6. Run `go test ./...`.
7. Run the command on hardware and record the results in the diary/changelog.

## Open Questions

1. Should the demo show icon labels on-device or icon-only buttons for maximum visual clarity?
2. What default animation FPS best balances smoothness and transport conservatism for this icon workload?
3. Should the icon-loader API live in the root package long-term, or remain an implementation detail of the demo command until it stabilizes?

## References

- Imported asset library: `ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/sources/local/macos1-icon-library.html`
- Root hardware package: `github.com/go-go-golems/loupedeck`
