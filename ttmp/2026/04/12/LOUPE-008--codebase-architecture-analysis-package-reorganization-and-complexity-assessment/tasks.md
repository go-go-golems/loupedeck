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

- [x] Add `cmd/loupe-js-live/options.go`
- [x] Introduce an `options` struct covering all current flags
- [x] Move flag parsing and validation into `parseOptions()`
- [x] Keep the current CLI behavior and error messages intact

### Phase B: main flow extraction

- [x] Add `cmd/loupe-js-live/run.go`
- [x] Move script loading into the runner path
- [x] Move device connection and display validation into the runner path
- [x] Move runtime/env/renderer setup into the runner path
- [x] Move the main `select` loop into the runner path
- [x] Shrink `main.go` to parse + call + exit

### Phase C: stats helper extraction

- [x] Add `cmd/loupe-js-live/stats.go`
- [x] Move `renderStatsWindow` into `stats.go`
- [x] Move writer diff helpers into `stats.go`
- [x] Move JS counter/timing formatting helpers into `stats.go`
- [x] Move trace filtering/formatting helpers into `stats.go`

### Phase D: logging and cleanup extraction

- [x] Add `cmd/loupe-js-live/logging.go`
- [x] Move `registerEventLogging` into `logging.go`
- [x] Add `cmd/loupe-js-live/cleanup.go`
- [x] Move `clearDisplays` into `cleanup.go`

### Phase E: validation and bookkeeping

- [x] Run `gofmt -w cmd/loupe-js-live/*.go`
- [x] Run `go test ./...`
- [x] Update the LOUPE-008 diary for the decomposition slice
- [x] Update the LOUPE-008 changelog for the decomposition slice
- [x] Commit the code slice
- [x] Commit the bookkeeping slice

## Future follow-up candidates

- [ ] Consider whether `cmd/loupe-fps-bench` should be decomposed similarly after `loupe-js-live` stabilizes
- [ ] Consider whether any stats/logging helpers become worthy of a shared package only after a second consumer appears

