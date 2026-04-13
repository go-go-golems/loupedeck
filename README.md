# go-go-golems/loupedeck

A Go library and CLI for talking directly to Loupedeck hardware over the firmware 2.x serial-WebSocket protocol, with a retained Go/JavaScript runtime for building and running live scenes on real devices.

## Status

This repository is now organized around a narrower, release-facing surface, but it is still **pre-1.0**.

That means:
- the main supported CLI is `loupedeck`,
- the main supported Go package is `pkg/device`,
- examples and dev tools are intentionally separated from the primary shipped surface,
- behavior and APIs may still change as the first release is tightened.

## Supported surface

The main supported surfaces are:

- `cmd/loupedeck` — the primary CLI binary
- `pkg/device` — the low-level device/protocol package

The current hardware focus is:

- **Loupedeck Live** (`product 0004`)

Other devices may partially work through the protocol abstractions already present in the repo, but they are not yet treated as the primary release target.

## Quick start

Build or run the main CLI:

```bash
go run ./cmd/loupedeck --help
```

Run a JavaScript scene on a connected device:

```bash
go run ./cmd/loupedeck run --script ./examples/js/01-hello.js --duration 5s
```

Useful runtime flags:

- `--device /dev/ttyACM0` — connect to a specific serial device path instead of auto-detecting
- `--send-interval 0ms` — remove writer pacing delays
- `--flush-interval 20ms` — tune retained render scheduler cadence
- `--with-glaze-output --output json` — emit structured Glazed output

To inspect the embedded help system:

```bash
go run ./cmd/loupedeck help
go run ./cmd/loupedeck run --help
```

## Installation and development build

This repo is not yet presented here as a stable `go install ...@latest` release target. For now, the most reliable path is to build from source:

```bash
go build ./cmd/loupedeck
```

Or use the included Makefile:

```bash
make build
```

## Repository layout

- `cmd/loupedeck/` — primary release CLI
- `pkg/device/` — active low-level device/protocol implementation
- `runtime/` — retained UI, rendering, runtime host, metrics, and JS integration
- `examples/js/` — JavaScript scene examples intended to run on real hardware
- `examples/cmd/` — example/demo binaries kept out of the main release path
- `dev-tools/` — benchmark and developer-only tooling
- `docs/help/` — embedded Glazed help pages for the CLI
- `sources/loupedeck-repo/` — local upstream reference clone kept for protocol comparison
- `ttmp/` — ticket docs, design notes, and implementation diary material

## Development vs. non-primary binaries

These paths are intentionally **not** the main shipped CLI:

- `dev-tools/loupe-fps-bench/` — raw hardware throughput benchmark
- `examples/cmd/loupe-js-demo/` — retained runtime PNG demo command
- `examples/cmd/loupe-svg-buttons/` — animated SVG demo command

They are useful for investigation and development, but they should be treated as secondary surfaces rather than the primary end-user entrypoint.

## Go package usage

The main public Go package boundary is:

```go
import "github.com/go-go-golems/loupedeck/pkg/device"
```

That package contains the active low-level device/protocol implementation used by the CLI.

## Upstream protocol reference

A major practical reference for the device-level implementation in this repo was Scott Laird’s original Loupedeck repository:

- <https://github.com/scottlaird/loupedeck>

A local clone is preserved under:

- `sources/loupedeck-repo/`

That upstream repo was heavily used as the main source of protocol implementation detail while writing and validating the low-level `pkg/device` layer here. Credit where due: it was an important reference for understanding serial/WebSocket framing, message flow, and basic device behavior.

## Release plumbing

The repo now includes the standard go-go-golems release/build plumbing:

- GitHub Actions workflows under `.github/workflows/`
- `.goreleaser.yaml`
- `.golangci.yml`
- `.golangci-lint-version`
- `Makefile`

The current release binary configured in GoReleaser is:

- `loupedeck` from `./cmd/loupedeck`

## Validation

Useful local validation commands:

```bash
make lint
go test ./...
make build
```

For CLI smoke checks:

```bash
go run ./cmd/loupedeck --help
go run ./cmd/loupedeck run --help
```

## JavaScript runtime docs

The repo includes Glazed-formatted help pages for the current goja-based Loupedeck runtime under:

- `docs/help/tutorials/01-build-your-first-live-loupedeck-js-script.md` — step-by-step guide for writing and running a real script on hardware
- `docs/help/topics/01-loupedeck-js-api-reference.md` — API reference for `loupedeck/state`, `loupedeck/ui`, `loupedeck/anim`, `loupedeck/easing`, and the live runner
- `docs/help/topics/02-reusable-goja-js-metrics-subpackage.md` — guide to the reusable JS metrics collector/binding packages and how to embed them into your own goja runtime

Those pages are embedded into the main `loupedeck` CLI help system.

## Project reports and deeper design docs

For the detailed architecture and performance investigation trail, see:

- `ttmp/2026/04/12/LOUPE-008--codebase-architecture-analysis-package-reorganization-and-complexity-assessment/design-doc/01-codebase-architecture-analysis-package-reorganization-and-complexity-assessment.md`
- `ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/design/01-project-report-cleanup-and-performance.md`
