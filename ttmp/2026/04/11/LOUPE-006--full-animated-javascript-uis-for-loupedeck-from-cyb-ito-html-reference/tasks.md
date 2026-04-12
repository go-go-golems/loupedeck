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

### Phase B: retained display-region groundwork

- [x] Add retained UI model support for named display regions: `left`, `main`, and `right`
- [x] Keep `page.tile(...)` working by delegating it to the retained `main` display region
- [x] Add retained per-display dirty tracking without breaking existing tile dirty semantics
- [x] Extend the retained renderer bridge so it can flush `left`, `main`, and `right` display regions
- [x] Update `cmd/loupe-js-live` to flush all retained display regions rather than only the main display
- [x] Add Go tests for retained display-region activation, dirty filtering, and rendering
- [x] Add JS integration tests for `ui.page(...).display(...)` and `ui.show(...)`
- [x] Run `go test ./...` after the display-region slice
- [x] Commit the display-region slice
- [x] Record the display-region slice in the diary/changelog/tasks

### Phase C: pure-Go retained graphics package

- [x] Add a pure-Go `runtime/gfx` package for retained grayscale/additive surfaces
- [x] Define a surface model that supports efficient clear, text, line, crosshatch, and compositing operations
- [x] Keep graphics semantics Go-owned rather than JS-per-pixel
- [x] Add focused unit tests for the graphics package
- [x] Run `go test ./...` after the graphics package slice
- [x] Commit the graphics package slice
- [x] Record the graphics package slice in the diary/changelog/tasks

### Phase D: JS-facing graphics module

- [x] Add `runtime/js/module_gfx/module.go`
- [x] Register `loupedeck/gfx` from `runtime/js/runtime.go`
- [x] Expose retained surfaces and coarse drawing ops to JS
- [x] Preserve owner-thread safety for any JS callbacks/closures involved in graphics composition
- [x] Add JS integration tests for the `loupedeck/gfx` surface API
- [x] Run `go test ./...` after the JS graphics module slice
- [x] Commit the JS graphics module slice
- [x] Record the JS graphics module slice in the diary/changelog/tasks

### Phase E: retained surface/layer composition

- [x] Add retained display-owned surface attachment and dirty propagation
- [x] Extend the retained renderer so attached display surfaces flow through the existing Go-owned invalidation/writer stack
- [x] Add tests for retained display-surface rendering and dirty propagation
- [ ] Add retained multi-layer composition support for overlays and multi-pass visuals
- [ ] Define a stable ordering model for base surfaces, overlays, and transient effects
- [x] Run `go test ./...` after the first retained-surface composition slice
- [x] Commit the first retained-surface composition slice
- [x] Record the retained-surface composition slice in the diary/changelog/tasks

### Phase F: first cyb-ito-inspired main-scene demo

- [x] Add a first JS demo script for the main animated scene
- [x] Port the 12-tile scene structure to the retained scene/surface model in prototype form
- [ ] Add touch-driven ripple and tile activation behavior on the main display
- [ ] Validate the main-scene demo locally via the live runner
- [x] Commit the main-scene demo slice
- [x] Record the main-scene demo slice in the diary/changelog/tasks

### Phase G: left/right strip scenes

- [x] Add left-strip dripping-bar scene support in prototype form
- [x] Add right-strip scrolling-kanji scene support in prototype form
- [ ] Add cross-display scene coordination such as activity pips or mirrored tile activation signals
- [ ] Validate the multi-display animated scene locally via the live runner
- [x] Commit the strip-scene slice
- [x] Record the strip-scene slice in the diary/changelog/tasks

### Phase H: hardware validation and tuning

- [ ] Validate the full animated scene demo on actual Loupedeck Live hardware
- [ ] Measure whether the denser animation workload stresses the current renderer/writer pacing model
- [ ] Decide whether renderer scheduling or pacing needs adjustment under dense animated workloads
- [ ] Commit any hardware-driven tuning changes separately from pure feature work
- [ ] Record hardware validation and tuning results in the diary/changelog/tasks

