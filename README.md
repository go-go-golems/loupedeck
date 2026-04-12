# go-go-golems/loupedeck

A Go package for talking directly to Loupedeck hardware over USB serial using the firmware 2.x serial-WebSocket protocol.

## Current status

This repository started as a hardware investigation and feature-tester workspace. It now contains the beginning of the real package module:

```text
module github.com/go-go-golems/loupedeck
```

The current focus is:

1. establish a root package instead of ticket-local scripts,
2. preserve the working baseline implementation,
3. improve lifecycle safety and event composition,
4. add package-owned backpressure handling (B-lite first, then full B).

## Supported hardware focus

The main hardware target right now is:

- **Loupedeck Live** (`product 0004`)

Other devices may partially work through the existing protocol abstractions, but they are not the primary focus of the initial refactor.

## Repository layout

- root package files (`*.go`) — active package implementation
- `sources/loupedeck-repo/` — upstream reference clone kept for comparison
- `ttmp/` — ticket documentation and implementation diary

## Near-term roadmap

- Phase 0: root module and baseline package port
- Phase 1: composable event listeners and safe lifecycle behavior
- Phase 2: single outbound writer and pacing for B-lite
- Phase 3: migrate the feature tester to the new package
- Phase 4: dirty-region invalidation and render coalescing

## Development

```bash
go test ./...
```

## JavaScript runtime docs

The repository now includes Glazed-formatted help pages for the current goja-based Loupedeck runtime under:

- `docs/help/tutorials/01-build-your-first-live-loupedeck-js-script.md` — step-by-step user guide for writing and running a real script on hardware
- `docs/help/topics/01-loupedeck-js-api-reference.md` — detailed API reference for `loupedeck/state`, `loupedeck/ui`, `loupedeck/anim`, `loupedeck/easing`, and the live runner

These pages are authored in Glazed help format so they can be loaded into a future Cobra/Glazed root help system when this repo grows one.

For the detailed architecture and implementation plan, see:

- `ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md`
