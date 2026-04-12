# Tasks

## TODO

### Ticket setup and source provenance
- [x] Create LOUPE-004 ticket workspace
- [x] Import the icon-library HTML into the ticket sources via `docmgr import file`
- [x] Replace template docs with a real design plan and diary

### SVG extraction and normalization
- [x] Extract icon-cell SVG fragments and labels from the imported HTML library
- [x] Extract shared defs and CSS color variables from the HTML source
- [x] Normalize SVG fragments for Go rasterization (vars, dither fills, style cleanup, namespace)
- [x] Add tests for extraction and normalization

### Rasterization and tile composition
- [x] Add Go-side SVG rasterization support
- [x] Trim transparent bounds so icons scale by visible content, not excess padding
- [x] Compose scaled `90×90` button frames with consistent styling
- [x] Add tests for scaling/composition helpers

### Animated demo program
- [x] Add a root command that loads the imported library and renders 12 animated icon buttons on the Loupedeck Live
- [x] Make the command configurable via a library path flag
- [x] Add Circle-button exit handling and safe cleanup
- [x] Validate the command builds with `go test ./...`
- [x] Run the demo on actual hardware

### Icon-bank navigation and selection
- [x] Add `--offset` support for starting from a later icon in the selected list
- [x] Add `--icons` support for curated comma-separated icon subsets
- [x] Add automatic page cycling between banks of 12 via `--page-every`
- [x] Add physical-button bank controls (previous/next/toggle-cycle)
- [x] Add touch-based bank controls (previous/next/toggle-cycle)
- [x] Add tests for selection ordering and bank padding behavior
- [x] Run the banked/curated SVG demo on actual hardware

### Documentation and continuity
- [x] Update the diary after each major implementation step
- [x] Update changelog and related-file bookkeeping
- [x] Run `docmgr doctor --ticket LOUPE-004 --stale-after 30`
