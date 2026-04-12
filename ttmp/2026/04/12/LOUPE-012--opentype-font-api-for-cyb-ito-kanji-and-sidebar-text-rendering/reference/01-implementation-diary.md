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
RelatedFiles:
    - Path: examples/js/07-cyb-ito-prototype.js
      Note: Step 4 prototype kanji and sidebar integration
    - Path: runtime/gfx/font.go
      Note: Step 2 Go-side font loader/cache implementation
    - Path: runtime/js/module_gfx/module.go
      Note: Step 3 JS `gfx.font(...)` and text font option implementation
    - Path: runtime/js/runtime_test.go
      Note: JS runtime tests for font handles and kanji collection rendering
ExternalSources: []
Summary: Chronological diary for OpenType font loading and kanji rendering support in the Loupedeck JS runtime.
LastUpdated: 2026-04-12T18:15:00-04:00
WhatFor: Use this diary to understand the implementation steps, failures, and validation path for LOUPE-012.
WhenToUse: Use when continuing or reviewing the OpenType/CJK font support work.
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

## Step 2: Implement the Go-side font loader/cache with support for TTC collections

The first code slice focused only on the Go-side font-loading foundation. I wanted a clean `runtime/gfx` abstraction for loading fonts that could support both regular `.ttf/.otf` files and multi-font `.ttc` collections before exposing anything to JS.

That choice kept the first code commit small and testable. It also forced the font-loading semantics to become explicit early: path-based loading, normalized options, cache keys, and collection index selection.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Start the implementation with the lowest-risk substrate: Go-side font loading and caching.

**Inferred user intent:** Build the feature from the bottom up so the JS API and cyb-ito integration rest on a solid base.

**Commit (code):** `07d8f6d` — `Add gfx OpenType font loader and cache`

### What I did
- Added `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/font.go`.
- Added:
  - `FontOptions`
  - `LoadedFont`
  - `LoadFont(path, opts)`
  - option normalization
  - a package-level cache keyed by path/size/dpi/index/hinting
- Used `opentype.ParseCollection(...)` first so `.ttc` collections are supported.
- Fell back to `opentype.Parse(...)` for regular `.ttf/.otf` files.
- Added tests in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/font_test.go` for:
  - regular font cache reuse
  - distinct cache entries for different options
  - collection loading using `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc` when available
- Ran:
  - `go test ./runtime/gfx/... ./...`

### Why
- The JS API should not need to know how to parse TTC collections or normalize default font options.
- The machine already has a useful CJK collection installed, so collection support should be first-class rather than an afterthought.

### What worked
- The loader and cache were straightforward to isolate.
- The system font collection path exists on this machine, so the collection-loading test could be made real rather than hypothetical.

### What didn't work
- N/A

### What I learned
- The `golang.org/x/image/font/opentype` package already exposes everything needed, including collection support via `ParseCollection`/`Collection.Font(i)`.
- The hardest part of font loading here is not glyph rasterization but API boundary design and cache semantics.

### What was tricky to build
- The subtle design choice was whether to cache a parsed face or just parsed font bytes. I kept the initial version simple by caching the loaded face object directly because it is enough for the current runtime usage model and lets the higher layers progress.

### What warrants a second pair of eyes
- If the runtime later starts drawing text concurrently across multiple goroutines, the shared-face caching strategy should be revisited.

### What should be done in the future
- Expose the loaded font through the JS bridge as the next slice.

### Code review instructions
- Review `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/font.go` first.
- Validate with `go test ./runtime/gfx/... ./...`.

### Technical details
- Primary system CJK font used for collection testing: `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`

## Step 3: Expose `gfx.font(...)` to JS and allow `surface.text(..., { font })`

Once the Go-side loader existed, the next step was to bring it through the JS boundary without destabilizing the existing `gfx` API. The design stayed deliberately narrow: add a font handle object to `loupedeck/gfx`, and let text options optionally carry that handle.

The crucial rule for this slice was preserving the current behavior when no font is provided. Existing scripts should still use `basicfont.Face7x13` by default.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Build the JS-facing font handle API on top of the loader/cache and keep the old bitmap-font fallback intact.

**Inferred user intent:** Make real font rendering available to scenes without breaking the rest of the runtime.

**Commit (code):** `8091029` — `Add JS gfx font handles and text font option`

### What I did
- Updated `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go`.
- Added:
  - `gfx.font(path, opts)`
  - JS font handle objects carrying `__font`
  - `LoadedFontFromValue(...)`
  - parsing of `font` from `surface.text(..., { font })`
- Kept `basicfont.Face7x13` as the fallback when no font is supplied.
- Added JS runtime tests in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go` covering:
  - loading a regular font handle and using it in `surface.text(...)`
  - loading a collection font and rendering a kanji glyph (`渦`) when the system collection is available
- Ran:
  - `go test ./runtime/js ./...`

### Why
- The bridge needed to expose only a minimal API surface: font creation and font reuse in text calls.
- Preserving the old fallback was important so existing example scripts continue to boot unchanged.

### What worked
- The JS bridge accepted the new handle cleanly.
- The runtime tests now prove a real kanji glyph can flow through the JS `gfx` path when backed by a collection font.

### What didn't work
- The first implementation hit a concrete Go type mismatch:
  - `cannot use loaded.Face() (value of interface type font.Face) as *basicfont.Face value in assignment: need type assertion`
- Fix: widen the local variable to `var face font.Face = basicfont.Face7x13` so both the bitmap fallback and real loaded faces fit the same slot.

### What I learned
- The most important real proof point was not just creating a font object but showing that a kanji glyph could be rasterized through the existing `Surface.Text(...)` path.
- The JS-facing API can stay very small and still unlock the full feature.

### What was tricky to build
- The bridge code itself was not complicated, but the type interaction between the fallback font and loaded OpenType faces was a good reminder not to let concrete fallback types leak into the API layer.

### What warrants a second pair of eyes
- The current JS API is path-based. That is fine for this stage, but future ergonomics work may want font discovery or a repo-local asset convention.

### What should be done in the future
- Use the new API in a real cyb-ito example so the feature is not only test-proven but visually relevant.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_gfx/module.go`.
- Then review the new runtime tests in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/runtime_test.go`.
- Validate with `go test ./runtime/js ./...`.

### Technical details
- The CJK collection test uses `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc` with `index: 0`.

## Step 4: Use the new font API in a real cyb-ito example for kanji labels and sidebar text

With the API proven, the next step was to make one of the cyb-ito examples actually use it. The prototype scene was the best target because it already had `left`, `main`, and `right` display surfaces. That meant it could exercise both kanji tile labels and side-strip text without needing a new example file.

This slice keeps the integration focused and honest: one real scene, one real system font path, and no new rendering subsystem. The scene still uses the normal `gfx` text path; it simply uses a CJK-capable face now.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Demonstrate the new API in a concrete cyb-ito example by replacing placeholder ASCII labels/sidebar text with real CJK rendering.

**Inferred user intent:** Make the feature materially useful for the actual Junji Ito graphics work, not just available in theory.

**Commit (code):** `0aa9fc2` — `Render cyb-ito prototype kanji with loaded font`

### What I did
- Updated `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js`.
- Added:
  - `CJK_FONT_PATH = "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc"`
  - `CJK_FONT = gfx.font(...)`
  - `CJK_FONT_SMALL = gfx.font(...)`
- Replaced tile metadata so each tile now has a real kanji title:
  - `眼`, `渦`, `歯`, `溶`, `穴`, `狂`, `蟲`, `砂`, `歪`, `裂`, `脈`, `闇`
- Updated `drawTileChrome(...)` to render kanji labels with the loaded font.
- Replaced the right-strip ASCII scroller with a kanji scroller derived from the original `cyb-ito.html` horror string.
- Ran:
  - `go test ./...`

### Why
- The point of the API is to make the cyb-ito scenes better, so the ticket needed a real usage example before stopping.
- The prototype scene already had side displays, which made it the most natural place to test both tile-title kanji and sidebar-style text.

### What worked
- The example still booted under the example-script test suite.
- The new font API was sufficient to render both tile titles and sidebar glyphs without any further runtime changes.

### What didn't work
- The scene currently assumes the system font path exists on this machine. That is fine for this repo/machine workflow, but it is not yet a generalized asset-management solution.

### What I learned
- The API is already useful enough for real scene code after only two infrastructure slices.
- The retained `gfx` text pipeline did not need to be rewritten; it only needed access to a real face.

### What was tricky to build
- The main tradeoff was whether to create a new cyb-ito example or update the prototype. Updating the prototype was the better move because it let the feature prove itself in a scene that already had side displays.
- The font path is intentionally explicit. That makes the current behavior reproducible on this machine, even if it is not yet the final portability story.

### What warrants a second pair of eyes
- The system-font-path assumption should be revisited later if these examples are expected to travel to other machines unchanged.
- The chosen CJK face/style may not be the best final art direction for Junji Ito text on device; this slice proves capability, not final typography.

### What should be done in the future
- Add ticket-local scripts and ticket bookkeeping.
- Later, decide whether to move from a system font path to a repo-imported font asset.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/07-cyb-ito-prototype.js`.
- Review:
  - font declarations at the top
  - `tiles` metadata changes
  - `drawTileChrome(...)`
  - `renderRight()`
- Validate with `go test ./...`.

### Technical details
- First concrete CJK font path used in scene code: `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`
