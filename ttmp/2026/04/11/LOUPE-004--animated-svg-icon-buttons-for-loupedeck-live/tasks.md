# Tasks

## TODO

### Ticket setup and source provenance
- [x] Create LOUPE-004 ticket workspace
- [x] Import the icon-library HTML into the ticket sources via `docmgr import file`
- [x] Replace template docs with a real design plan and diary

### SVG extraction and normalization
- [ ] Extract icon-cell SVG fragments and labels from the imported HTML library
- [ ] Extract shared defs and CSS color variables from the HTML source
- [ ] Normalize SVG fragments for Go rasterization (vars, dither fills, style cleanup, namespace)
- [ ] Add tests for extraction and normalization

### Rasterization and tile composition
- [ ] Add Go-side SVG rasterization support
- [ ] Trim transparent bounds so icons scale by visible content, not excess padding
- [ ] Compose scaled `90×90` button frames with consistent styling
- [ ] Add tests for scaling/composition helpers

### Animated demo program
- [ ] Add a root command that loads the imported library and renders 12 animated icon buttons on the Loupedeck Live
- [ ] Make the command configurable via a library path flag
- [ ] Add Circle-button exit handling and safe cleanup
- [ ] Validate the command builds with `go test ./...`
- [ ] Run the demo on actual hardware

### Documentation and continuity
- [ ] Update the diary after each major implementation step
- [ ] Update changelog and related-file bookkeeping
- [ ] Run `docmgr doctor --ticket LOUPE-004 --stale-after 30`
