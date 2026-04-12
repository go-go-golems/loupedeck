# Tasks

## Phase A: ticket setup and architecture plan

- [x] Create `LOUPE-012`
- [x] Write the implementation plan for OpenType font loading and kanji rendering
- [x] Create the implementation diary
- [x] Define the phased task list

## Phase B: Go-side font loading and caching

- [ ] Add a `runtime/gfx` font loader abstraction
- [ ] Support `.ttf` / `.otf` loading
- [ ] Support `.ttc` collection loading with face index selection
- [ ] Add caching keyed by path/size/dpi/index
- [ ] Add unit tests for font loading and cache reuse
- [ ] Run `go test ./...`

## Phase C: JS font API

- [ ] Expose `gfx.font(path, opts)` from `runtime/js/module_gfx/module.go`
- [ ] Allow `surface.text(..., { font })`
- [ ] Keep `basicfont.Face7x13` as the fallback when no font is supplied
- [ ] Add JS runtime tests for font handle use across the bridge
- [ ] Run `go test ./...`

## Phase D: first cyb-ito integration

- [ ] Update one cyb-ito example to render actual kanji labels with a loaded CJK font
- [ ] Add first sidebar or side-strip text rendering experiment with the loaded font
- [ ] Keep the rendering inside the normal retained surface pipeline
- [ ] Archive reproducibility commands in `scripts/`

## Phase E: continuity and validation

- [ ] Update the diary after each code slice
- [ ] Update the ticket changelog and index
- [ ] Run `docmgr doctor --ticket LOUPE-012 --stale-after 30`
- [ ] Commit code slices separately from bookkeeping slices
