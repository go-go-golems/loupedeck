# Tasks

## TODO

### Ticket setup and research framing
- [x] Create LOUPE-005 ticket workspace
- [x] Add brainstorm design doc, examples doc, and diary
- [x] Capture current renderer/writer constraints that shape the future JS API

### API brainstorming
- [x] Compare multiple JavaScript API styles (imperative, declarative, reactive, timeline-centric, hybrid)
- [x] Brainstorm animation and easing models
- [x] Brainstorm scene/page/state/runtime lifecycle models
- [x] Include tradeoffs and likely recommendation direction

### Example scenarios
- [x] Write multiple example scripts across different approaches
- [x] Include examples for easing curves, timelines, and interactive controls
- [x] Include dynamic paging, overlays, and reactive scenarios

### Future implementation planning
- [x] Choose a preferred JS module contract
- [x] Decide low-level vs high-level module boundaries (`loupedeck`, `ui`, `anim`, `easing`, etc.)
- [x] Decide host runtime semantics (`setTimeout`, `requestAnimationFrame`, fixed-timestep scheduler, etc.)
- [x] Decide state ownership model (JS-owned, Go-owned retained scene, or hybrid)
- [x] Decide what asset pipeline to expose to JS
- [ ] Decide how/if scripts persist across reconnects
- [x] Write a detailed intern-oriented reactive textbook
- [x] Write a detailed phased implementation plan

### Documentation and continuity
- [x] Start the diary
- [x] Update index/changelog/tasks coherently
- [ ] Run `docmgr doctor --ticket LOUPE-005 --stale-after 30`
- [ ] Commit ticket docs and bookkeeping
- [ ] Upload the expanded document bundle to reMarkable
