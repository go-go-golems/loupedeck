# Tasks

## Phase A: ticket setup and architecture plan

- [x] Create `LOUPE-011`
- [x] Write the implementation plan for layered full-frame effects on the presenter-driven runtime
- [x] Create the implementation diary
- [x] Define a detailed phased task list

## Phase B: first layered full-page implementation slice

- [ ] Refactor `examples/js/10-cyb-ito-full-page-all12.js` to use multiple internal `gfx.surface(...)` layers
- [ ] Preserve the current presenter-driven full-page flush model
- [ ] Add a first software FX layer with at least scanlines, grain/noise, or active-tile overlay effects
- [ ] Keep the final scene attached as one full-page display surface
- [ ] Run `go test ./...`

## Phase C: first hardware tuning and validation

- [ ] Archive concrete run/test commands in `scripts/`
- [ ] Run the layered full-page scene on hardware with aggressive writer pacing
- [ ] Inspect whether the effect pass hurts visible smoothness
- [ ] Tune effect intensity or composition cost if needed

## Phase D: ticket continuity

- [ ] Update the diary after the first code slice
- [ ] Update the ticket changelog and index
- [ ] Run `docmgr doctor --ticket LOUPE-011 --stale-after 30`
- [ ] Commit the code slice
- [ ] Commit the ticket bookkeeping slice

## Future work candidates after the first slice

- [ ] Cache static base/chrome layers instead of rebuilding everything every frame
- [ ] Reintroduce more faithful global sweep/glitch passes
- [ ] Reintroduce side-strip scene composition under the same one-frame-flush model
- [ ] Produce a longer technical write-up once the layered scene stabilizes
