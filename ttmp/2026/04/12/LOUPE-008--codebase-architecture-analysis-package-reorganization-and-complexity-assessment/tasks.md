# Tasks

## Completed architecture cleanup slices

- [x] Move the former root driver package into `pkg/device`
- [x] Delete the obsolete widget/value stack and the `loupe-feature-tester` binary
- [x] Move device display profiling into connect-time initialization
- [x] Add canonical input naming/parsing APIs and remove duplicated input-name maps

## Next slice: `cmd/loupe-js-live` decomposition

### Planning

- [x] Write a dedicated design doc for decomposing `cmd/loupe-js-live`
- [x] Write a dedicated implementation plan for the decomposition
- [x] Define a phased task list for the decomposition work

### Phase A: options parsing split

- [ ] Add `cmd/loupe-js-live/options.go`
- [ ] Introduce an `options` struct covering all current flags
- [ ] Move flag parsing and validation into `parseOptions()`
- [ ] Keep the current CLI behavior and error messages intact

### Phase B: main flow extraction

- [ ] Add `cmd/loupe-js-live/run.go`
- [ ] Move script loading into the runner path
- [ ] Move device connection and display validation into the runner path
- [ ] Move runtime/env/renderer setup into the runner path
- [ ] Move the main `select` loop into the runner path
- [ ] Shrink `main.go` to parse + call + exit

### Phase C: stats helper extraction

- [ ] Add `cmd/loupe-js-live/stats.go`
- [ ] Move `renderStatsWindow` into `stats.go`
- [ ] Move writer diff helpers into `stats.go`
- [ ] Move JS counter/timing formatting helpers into `stats.go`
- [ ] Move trace filtering/formatting helpers into `stats.go`

### Phase D: logging and cleanup extraction

- [ ] Add `cmd/loupe-js-live/logging.go`
- [ ] Move `registerEventLogging` into `logging.go`
- [ ] Add `cmd/loupe-js-live/cleanup.go`
- [ ] Move `clearDisplays` into `cleanup.go`

### Phase E: validation and bookkeeping

- [ ] Run `gofmt -w cmd/loupe-js-live/*.go`
- [ ] Run `go test ./...`
- [ ] Update the LOUPE-008 diary for the decomposition slice
- [ ] Update the LOUPE-008 changelog for the decomposition slice
- [ ] Commit the code slice
- [ ] Commit the bookkeeping slice

## Future follow-up candidates

- [ ] Consider whether `cmd/loupe-fps-bench` should be decomposed similarly after `loupe-js-live` stabilizes
- [ ] Consider whether any stats/logging helpers become worthy of a shared package only after a second consumer appears

