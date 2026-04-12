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
    - Path: examples/js/10-cyb-ito-full-page-all12.js
      Note: |-
        Step 5 source-derived side-strip port and three-display validation target
        Step 5 presenter-driven full-page strip port and hardware smoke validation target
    - Path: runtime/gfx/font.go
      Note: Step 2 Go-side font loader/cache implementation
    - Path: runtime/js/module_gfx/module.go
      Note: Step 3 JS `gfx.font(...)` and text font option implementation
    - Path: runtime/js/runtime_test.go
      Note: JS runtime tests for font handles and kanji collection rendering
    - Path: ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html
      Note: Imported source inspected for exact strip behavior in Step 5
    - Path: ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/scripts/06-render-scene-preview.go
      Note: |-
        Step 7 offscreen framebuffer preview exporter for visual tuning
        Step 7 offscreen stitched PNG preview export workflow
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

## Step 5: Port the side strips from the imported source into the presenter-driven full-page scene

The next useful slice was not “more generic font API work” but a source-driven scene port. The user pointed directly at the imported `cyb-ito.html` and asked me to use it for the side-strip implementation details. That changed the task from a loose sidebar approximation into a concrete fidelity step: inspect the exact strip math and port that behavior into the hardware-backed full-page example.

This step therefore focused on `examples/js/10-cyb-ito-full-page-all12.js`. I kept the presenter-driven architecture intact, but extended the scene so it now drives three real displays again: `left`, `main`, and `right`. The new left/right strips are derived from the imported reference instead of from guesswork.

### Prompt Context

**User prompt (verbatim):** "look at @ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html for implementation details on the side strips"

**Assistant interpretation:** Inspect the imported `cyb-ito.html` directly, extract the exact side-strip behavior, and port that logic into the retained Loupedeck scene instead of improvising new sidebars.

**Inferred user intent:** The user wants source-faithful strip behavior, especially for the sidebar treatment, rather than another approximate horror-themed side decoration.

**Commit (code):** `ea429f8` — `Add source-derived side strips to full-page cyb-ito scene`

### What I did
- Inspected `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-006--full-animated-javascript-uis-for-loupedeck-from-cyb-ito-html-reference/sources/local/cyb-ito.html` around the strip-rendering section.
- Confirmed the exact source behavior:
  - left strip = 12 animated dripping bars
  - right strip = vertically faded horror-kanji scroller
  - both strips = divider lines and flash pips aligned to active rows
- Updated `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js` to:
  - load optional CJK fonts for tile chrome/HUD labels
  - add `leftStrip` and `rightStrip` retained surfaces sized to the real hardware `60x270`
  - add `stripScroll` and `frameCounter` scene state
  - port the source-derived `drip(...)` behavior into the left strip
  - port the source-derived right-strip kanji scroller into the right strip
  - attach `left`, `main`, and `right` displays in the scene page definition
- Ran:
  - `go test ./...`
- Ran a non-interactive hardware smoke command:
  - `timeout 30s go run ./cmd/loupe-js-live --script ./examples/js/10-cyb-ito-full-page-all12.js --duration 5s --send-interval 0ms --stats-interval 2s --log-render-stats`
- Recorded the evidence log:
  - `/tmp/loupe-cyb-ito-font-strips-1776033165.log`

### Why
- The imported HTML source already contained the side-strip semantics, so using that as the design source is more faithful and less error-prone than inventing a new sidebar treatment.
- The presenter-driven full-page scene was the right place to add this because it is the current serious hardware-backed scene, not just a prototype.

### What worked
- The source inspection gave a very concrete implementation target.
- `go test ./...` still passed after the scene expansion.
- The hardware smoke run exited cleanly with `reason=timeout` and showed all three expected display flushes:
  - `Display=left xoff=0 yoff=0 width=60 height=270`
  - `Display=main xoff=0 yoff=0 width=360 height=270`
  - `Display=right xoff=0 yoff=0 width=60 height=270`

### What didn't work
- I did not perform a human visual fidelity check in this step, so the evidence is a successful hardware-backed smoke run rather than a final aesthetic sign-off.
- The right-strip glyph appearance still depends on the local system font path:
  - `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`

### What I learned
- The side strips in the original file are simpler and more specific than they first appear. The left strip is not generic texture; it is a 12-segment meter with drips. The right strip is not free-form text; it is a 20px-step vertically faded kanji scroller.
- Porting from the source file directly is much better than reconstructing from memory when the goal is fidelity.

### What was tricky to build
- The tricky part was mapping browser-strip logic designed around `STRIP=36` into the hardware’s `60x270` side displays without losing the original behavior. I preserved the source semantics—segment count, drip logic, 20px kanji steps, divider lines, and row pips—while widening the drawable area to fit the real device side displays.
- Another subtle point was monotonic strip motion. The existing `phase` signal loops from `0..1`, but the source strip scroller uses a continuously increasing `scrollOff`. I added separate retained scene state (`stripScroll`, `frameCounter`) so the strip animation and drip wobble could behave more like the original source.

### What warrants a second pair of eyes
- The adaptation from source `STRIP=36` to hardware `SIDE_W=60` is intentionally source-inspired rather than byte-for-byte geometric equivalence. A visual review should confirm whether the widened strip still reads like the original.
- Adding left/right strip flushes means the scene now sends three display updates instead of only the main one. That is architecturally fine, but it should be watched during longer interactive runs.

### What should be done in the future
- Do a human visual check of the new side-strip fidelity on hardware.
- Decide whether to import and pin a repo-local CJK font asset instead of relying on the system font path.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`.
- Review these parts in order:
  - top-level strip/font constants and retained surfaces
  - `drip(...)`
  - `renderLeftStrip(...)`
  - `renderRightStrip(...)`
  - page display wiring for `left` / `main` / `right`
  - `anim.loop(...)` updates to `stripScroll` and `frameCounter`
- Validate with:
  - `go test ./...`
  - `timeout 30s go run ./cmd/loupe-js-live --script ./examples/js/10-cyb-ito-full-page-all12.js --duration 5s --send-interval 0ms --stats-interval 2s --log-render-stats`

### Technical details
- Source strip section inspected:
  - left strip around lines 659–672
  - right strip around lines 674–685
  - pip logic around lines 688–694
- Evidence log:
  - `/tmp/loupe-cyb-ito-font-strips-1776033165.log`

## Step 6: Tune kanji alignment and sidebar readability from hardware feedback

Once the source-derived strips were live, the next refinement came directly from user feedback on the real device. The user reported two concrete visual issues: the kanji characters should sit lower and further left, and the right scrolling sidebar felt too faint. That gave me a precise tuning target instead of vague “make it better” iteration.

This step stayed intentionally narrow. I did not change the underlying scene architecture again; I only adjusted the kanji font sizing and text-box placement inside the already-working full-page scene.

### Prompt Context

**User prompt (verbatim):** "oh actually i see it it's very faint on the right and the characters need to be larger"

**Assistant interpretation:** Keep the strip architecture intact, but make the right-side scrolling kanji visibly stronger and adjust kanji placement to align better on hardware.

**Inferred user intent:** The user wants the new CJK/strip work to be not just present but clearly legible on the physical device.

**Commit (code):** `58966ed` — `Tune full-page kanji alignment and sidebar visibility`

### What I did
- Updated `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`.
- Added explicit placement constants for kanji alignment:
  - `KANJI_TILE_DX`
  - `KANJI_TILE_DY`
  - `KANJI_HUD_DX`
  - `KANJI_HUD_DY`
- Added explicit right-strip layout constants:
  - `SIDEBAR_KANJI_X`
  - `SIDEBAR_KANJI_Y`
  - `SIDEBAR_KANJI_W`
  - `SIDEBAR_KANJI_H`
- Increased font sizes:
  - main kanji font from `16` to `18`
  - small kanji font from `12` to `14`
  - new sidebar kanji font at `20`
- Moved the right-strip kanji box left inside the `60px` strip and increased strip brightness substantially.
- Ran:
  - `go test ./...`

### Why
- The most direct evidence about visual quality here is user hardware feedback, not just my assumptions from code.
- The right strip was already structurally present; the problem was visibility and placement, so a layout/brightness/font-size pass was the right fix.

### What worked
- The tuning was localized to one scene file.
- The main-grid, HUD, and right-strip kanji can all be tuned through text-box placement and font-size changes without touching the font API or renderer internals.

### What didn't work
- N/A

### What I learned
- The initial right-strip placement was too far to the right for the actual `60px` hardware strip. Even when the glyphs existed, they were not occupying the most visible part of the strip.
- Hardware feedback is especially valuable for CJK layout because small box-position mistakes are very noticeable on a narrow, bezel-constrained display.

### What was tricky to build
- The tricky part was remembering that the strip text box in the current `drawText(...)` helper uses `x`/`y` as the top-left corner of a centered text box, not as a glyph anchor. That means “move the characters left” really means moving the entire text box left, not just changing a center point.

### What warrants a second pair of eyes
- The right-strip brightness curve is now much stronger than the source HTML. That was intentional for readability, but it may deserve one more visual pass for style balance.

### What should be done in the future
- Add an offscreen preview/export workflow so framebuffer layout can be inspected without always relying on live hardware.

### Code review instructions
- Review `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`.
- Focus on:
  - top-level kanji placement constants
  - font size changes
  - `drawTileChrome(...)`
  - `renderRightStrip(...)`
  - `renderHUDLayer(...)`
- Validate with `go test ./...`.

### Technical details
- Right-strip visibility problem came from the glyph box being too far right and too dim for the hardware strip width.

## Step 7: Add offscreen framebuffer PNG export and use it for preview-driven tuning

After the hardware-feedback tuning pass, the user suggested something very practical: export the framebuffer to PNG and inspect it with image tools. That was a good idea, because it turns scene-layout debugging into something repeatable and reviewable without requiring a live hardware session for every tiny text-position change.

I implemented that workflow as a ticket-local helper script rather than as ad hoc shell history. The helper boots the JS runtime offscreen, runs the scene, captures the `left`, `main`, and `right` display images through the normal renderer, stitches them into one `480x270` image, and writes a PNG.

### Prompt Context

**User prompt (verbatim):** "you can btw maybe export the framebuffer as png and then look at it to see if things are looking good, using your image analysis / toolss"

**Assistant interpretation:** Add a reproducible offscreen preview/export path for the JS scene and use it to inspect layout quality without hardware.

**Inferred user intent:** The user wants a faster debug loop for visual tuning, and wants the agent to use its image-analysis capabilities where possible.

**Commit (code):** `cfd341f` — `Add offscreen preview export for font tuning`

### What I did
- Added `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/scripts/06-render-scene-preview.go`.
- The helper:
  - creates a JS runtime/environment
  - installs a normal retained renderer with offscreen capture targets for `left`, `main`, and `right`
  - starts the presenter runtime
  - runs a JS scene file
  - waits briefly for initial render/animation state
  - stitches the captured displays into one `480x270` PNG
- Ran:
  - `go run ./ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/scripts/06-render-scene-preview.go --script ./examples/js/10-cyb-ito-full-page-all12.js --out /tmp/loupe-cyb-ito-preview.png --wait 400ms`
- Exported:
  - `/tmp/loupe-cyb-ito-preview.png`
- Attempted image analysis on the preview twice, but both attempts failed with the same transient backend error:
  - `API error (1033)`
  - `system error`
- Used the preview workflow anyway and applied one additional small kanji Y-offset adjustment in the scene.
- Ran:
  - `go test ./...`
  - the first run failed with the known flaky unrelated timing test:
    - `--- FAIL: TestAnimModuleLoopCanDriveReactiveUpdates (0.04s)`
    - `expected loop to update visible text, got "0"`
  - reran `go test ./...` successfully
- Regenerated the preview:
  - `/tmp/loupe-cyb-ito-preview-v2.png`

### Why
- A preview PNG is a durable artifact. It can be inspected later, attached to docs, and reviewed without needing the device.
- The helper uses the actual retained JS/runtime/render pipeline instead of inventing a fake alternate path, so the preview is meaningful.

### What worked
- The exporter script worked on the first real render after a small import fix.
- The stitched preview workflow gives one image containing the actual left/main/right composition.
- The second full test run passed after the known flaky timing test was rerun.

### What didn't work
- The image-analysis backend failed twice with the same transient error:
  - `API error (1033)`
  - `system error`
- The first `go test ./...` rerun after the preview-driven tweak hit the known flaky test:
  - `TestAnimModuleLoopCanDriveReactiveUpdates`

### What I learned
- The preview export path is absolutely worth keeping, even if image analysis is temporarily unavailable, because it shortens the visual-debugging loop dramatically.
- The current repo already has enough runtime separation to render the scene offscreen cleanly without any hardware-specific hacks.

### What was tricky to build
- The main tricky point was making the helper drive the normal presenter path correctly: it needed a real JS runtime environment, a started presenter, and a flush callback wired into the renderer. Without that, the scene would boot but not actually produce retained-frame output.
- Another subtle point was copying the drawn images inside the capture target. The renderer hands out images over time, so the capture target clones each frame instead of holding references that might later change.

### What warrants a second pair of eyes
- If this preview workflow becomes more central, it may deserve promotion from a ticket-local helper into a more discoverable repo tool or command.
- The preview script currently waits a fixed amount of time; a future version could wait for a minimum draw count or explicit scene readiness signal instead.

### What should be done in the future
- Optionally promote the preview exporter into a reusable repo-local helper.
- Retry image-analysis review later when the backend is healthy.

### Code review instructions
- Review `/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/scripts/06-render-scene-preview.go` first.
- Then review the tiny kanji Y-offset tweak in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js`.
- Validate with:
  - `go test ./...`
  - `go run ./ttmp/2026/04/12/LOUPE-012--opentype-font-api-for-cyb-ito-kanji-and-sidebar-text-rendering/scripts/06-render-scene-preview.go --script ./examples/js/10-cyb-ito-full-page-all12.js --out /tmp/loupe-cyb-ito-preview-v2.png --wait 400ms`

### Technical details
- Preview output size: `480x270`
  - `left 60 + main 360 + right 60`
- Preview artifacts:
  - `/tmp/loupe-cyb-ito-preview.png`
  - `/tmp/loupe-cyb-ito-preview-v2.png`
- Transient image-analysis failures observed:
  - `Trace-Id: 062b522e7b896a43b44612a9ac938fb4`
  - `Trace-Id: 062b524763cac70a5474bc1b4a8d7764`
