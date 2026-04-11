# Changelog

## 2026-04-11

- Initial workspace created


## 2026-04-11

Created the package-refactor ticket and wrote the primary backpressure-safe architecture/implementation guide for B-lite then B.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/design-doc/01-go-go-golems-loupedeck-package-backpressure-safe-architecture-and-implementation-guide.md — Primary implementation guide for the new package


## 2026-04-11

Related evidence files, added the backpressure topic vocabulary entry, passed docmgr doctor, and uploaded the ticket bundle to reMarkable.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-003--backpressure-safe-go-go-golems-loupedeck-package-refactor/tasks.md — Task state updated after validation and upload


## 2026-04-11

Implemented Phase 0 root-module port and Phase 1 composable listeners/safe lifecycle in the new root package.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/go.mod — Root go-go-golems module created
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/listen.go — Read loop now returns error instead of panicking
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/listeners.go — Composable listener registration and dispatch
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/listeners_test.go — Listener dispatch and unsubscribe tests


## 2026-04-11

Implemented Phase 2 B-lite transport ownership: single outbound writer, configurable pacing, command abstraction, and writer tests.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/connect.go — Connect helpers now accept writer options
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/display.go — Display draws now enqueue a single logical display draw command
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/writer.go — B-lite outbound writer
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/writer_test.go — Writer ordering and pacing tests


## 2026-04-11

Added a root-level feature tester command that uses the new package APIs and package-owned pacing instead of app-level sleep-based throttling.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-feature-tester/main.go — Root command for exercising the new package on hardware

