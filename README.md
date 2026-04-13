# go-go-golems/loupedeck

A Go module and CLI for talking directly to Loupedeck hardware over the firmware 2.x serial-WebSocket protocol, with a retained Go/JavaScript runtime for live scene development on real devices.

## Current status

This repository has moved from an exploratory hardware workspace toward a releaseable shape.

The main supported surfaces are now:

- `cmd/loupedeck` — the primary CLI binary
- `pkg/device` — the low-level device/protocol package

The current focus is:

1. keep the hardware driver clean and reusable,
2. keep the retained runtime practical for real device work,
3. make the repository release-ready with standard CI/release plumbing,
4. keep demos, experiments, and benchmark tools clearly separated from the main release binary.

## Supported hardware focus

The main hardware target right now is:

- **Loupedeck Live** (`product 0004`)

Other devices may partially work through the existing protocol abstractions, but they are not yet the primary release target.

## Main CLI

The main binary is now:

```bash
go run ./cmd/loupedeck --help
```

The primary runtime command is:

```bash
go run ./cmd/loupedeck run --script ./examples/js/01-hello.js --duration 5s
```

Useful tuning flags include:

- `--device /dev/ttyACM0` — connect to a specific serial device path
- `--send-interval 0ms` — remove writer pacing delays
- `--flush-interval 20ms` — tune retained render scheduler cadence
- `--with-glaze-output --output json` — emit structured command output through Glazed

## Repository layout

- `cmd/loupedeck/` — primary release CLI
- `pkg/device/` — active low-level device/protocol implementation
- `runtime/` — retained UI, rendering, host/runtime, and JS integration
- `examples/js/` — JavaScript scene examples
- `examples/cmd/` — example/demo binaries kept out of the main release path
- `dev-tools/` — benchmark and developer-only tooling
- `docs/help/` — embedded Glazed help pages for the CLI
- `sources/loupedeck-repo/` — upstream reference clone kept for comparison and protocol implementation guidance
- `ttmp/` — ticket documentation and implementation diary

## Development vs release commands

The following paths are intentionally **not** the primary release binary:

- `dev-tools/loupe-fps-bench/` — raw hardware throughput benchmark
- `examples/cmd/loupe-js-demo/` — retained runtime PNG demo command
- `examples/cmd/loupe-svg-buttons/` — animated SVG example/demo command

These remain useful for development and investigation, but they are treated as dev tools / examples rather than the main product surface.

## Go package usage

The most important public package boundary is:

```go
import "github.com/go-go-golems/loupedeck/pkg/device"
```

That package contains the active low-level device/protocol implementation.

## Protocol reference and shout-out

A major reference for the device-level implementation in this repository was the original `loupedeck-repo` code preserved under `sources/loupedeck-repo/`.

That repository was heavily used as the main source of protocol implementation details while writing and validating the low-level `pkg/device` layer here. Credit where due: it was an important practical reference for understanding the serial/WebSocket framing, message flow, and basic device behavior.

## Release readiness plumbing

The repository now includes standard go-go-golems release plumbing copied from the template repo:

- GitHub Actions workflows under `.github/workflows/`
- `.goreleaser.yaml`
- `.golangci.yml`
- `.golangci-lint-version`
- `Makefile`

The release binary configured in GoReleaser is:

- `loupedeck` from `./cmd/loupedeck`

## Development

```bash
go test ./...
make lint
make build
```

## JavaScript runtime docs

The repository includes Glazed-formatted help pages for the current goja-based Loupedeck runtime under:

- `docs/help/tutorials/01-build-your-first-live-loupedeck-js-script.md` — step-by-step user guide for writing and running a real script on hardware
- `docs/help/topics/01-loupedeck-js-api-reference.md` — detailed API reference for `loupedeck/state`, `loupedeck/ui`, `loupedeck/anim`, `loupedeck/easing`, and the live runner
- `docs/help/topics/02-reusable-goja-js-metrics-subpackage.md` — standalone guide to the reusable JS metrics collector/binding packages, how they work, and how to embed them into your own goja runtime

Those pages are now embedded into the main `loupedeck` CLI help system.

For the detailed architecture and implementation plan, see:

- `ttmp/2026/04/12/LOUPE-008--codebase-architecture-analysis-package-reorganization-and-complexity-assessment/design-doc/01-codebase-architecture-analysis-package-reorganization-and-complexity-assessment.md`
- `ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/design/01-project-report-cleanup-and-performance.md`
