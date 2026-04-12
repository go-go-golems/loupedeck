# Tasks

## Analysis and design

- [x] Create the LOUPE-006 ticket workspace
- [x] Import `~/Downloads/cyb-ito.html` as a tracked source artifact
- [x] Read and analyze the imported source carefully
- [x] Write a detailed intern-facing analysis / design / implementation guide
- [x] Write an implementation diary entry for continuity
- [x] Upload the design bundle to reMarkable
- [x] Verify the uploaded reMarkable files

## Planned implementation phases

- [ ] Add retained JS-facing display regions for `left`, `main`, and `right`
- [ ] Preserve the existing `page.tile(...)` API as a convenience layer on top of `main`
- [ ] Add a pure-Go retained graphics/surface package (likely `runtime/gfx`)
- [ ] Add a JS-facing `loupedeck/gfx` module on top of the pure-Go graphics layer
- [ ] Add retained surface/layer composition for overlays and multi-pass visuals
- [ ] Add a first cyb-ito-inspired main-display animated demo script
- [ ] Add left/right strip scene support for the cyb-ito-inspired demo
- [ ] Validate the animated scene demo on actual Loupedeck Live hardware
- [ ] Decide whether renderer scheduling or pacing needs adjustment under dense animated workloads

