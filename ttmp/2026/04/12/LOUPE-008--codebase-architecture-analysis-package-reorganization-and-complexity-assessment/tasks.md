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

## Release-facing CLI and repo plumbing

- [x] Replace `cmd/loupe-js-live` with a new `cmd/loupedeck` main binary
- [x] Implement the main runtime flow as a Glazed/Cobra `loupedeck run` command
- [x] Embed help docs into the new root CLI
- [x] Move `loupe-fps-bench` into `dev-tools/`
- [x] Move `loupe-js-demo` and `loupe-svg-buttons` into `examples/cmd/`
- [x] Copy and adapt GitHub Actions, GoReleaser, golangci-lint config, and Makefile from the go-template baseline
- [x] Rewrite `README.md` around the new main binary and supported surface
- [x] Tighten `README.md` into a more consumer-facing release document and add the explicit upstream repo URL
- [x] Make the copied lint pipeline pass without weakening the new checks
- [x] Review the copied GitHub Actions workflows against the actual shipped CLI surface
- [x] Review and modernize `.goreleaser.yaml` until `goreleaser check` passes without deprecation failures
- [x] Make example-script boot tests tolerate missing host CJK fonts so GitHub runners do not require distro-specific font packages
- [x] Validate the new CLI and release plumbing with `go run ./cmd/loupedeck --help`, `go run ./cmd/loupedeck run --help`, `make lint`, `go test ./...`, `make build`, `goreleaser check`, and `goreleaser release --snapshot --clean --skip=sign --skip=publish --single-target`

## Future follow-up candidates

- [ ] Decide whether `loupedeck run` is the final first-release UX or whether the root command should eventually execute scripts directly
- [ ] Add install/release notes after the first tagged release path is exercised end-to-end
- [ ] Decide whether `loupe-fps-bench` needs a true three-display benchmark mode before making broader performance claims
- [ ] Investigate the remaining serial-WebSocket handshake flake separately from the release-surface work

