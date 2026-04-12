# Tasks

## Phase A: ticket setup and architecture plan

- [x] Create `LOUPE-012`
- [x] Write the implementation plan for OpenType font loading and kanji rendering
- [x] Create the implementation diary
- [x] Define the phased task list

## Phase B: Go-side font loading and caching

- [x] Add a `runtime/gfx` font loader abstraction
- [x] Support `.ttf` / `.otf` loading
- [x] Support `.ttc` collection loading with face index selection
- [x] Add caching keyed by path/size/dpi/index
- [x] Add unit tests for font loading and cache reuse
- [x] Run `go test ./...`

## Phase C: JS font API

- [x] Expose `gfx.font(path, opts)` from `runtime/js/module_gfx/module.go`
- [x] Allow `surface.text(..., { font })`
- [x] Keep `basicfont.Face7x13` as the fallback when no font is supplied
- [x] Add JS runtime tests for font handle use across the bridge
- [x] Run `go test ./...`

## Phase D: first cyb-ito integration

- [x] Update one cyb-ito example to render actual kanji labels with a loaded CJK font
- [x] Add first sidebar or side-strip text rendering experiment with the loaded font
- [x] Keep the rendering inside the normal retained surface pipeline
- [x] Archive reproducibility commands in `scripts/`

## Phase E: continuity and validation

- [x] Update the diary after each code slice
- [x] Update the ticket changelog and index
- [x] Run `docmgr doctor --ticket LOUPE-012 --stale-after 30`
- [x] Commit code slices separately from bookkeeping slices

## Phase F: source-driven side-strip fidelity slice

- [x] Inspect the imported `cyb-ito.html` side-strip implementation details directly
- [x] Port the source-derived left dripping-bar strip behavior into the presenter-driven full-page scene
- [x] Port the source-derived right horror-kanji scroller into the presenter-driven full-page scene
- [x] Add the side strips to the real `left` and `right` hardware displays in `examples/js/10-cyb-ito-full-page-all12.js`
- [x] Run `go test ./...`
- [x] Run a non-interactive hardware smoke test and record the evidence log
