# Changelog

## 2026-04-13

- Initial workspace created


## 2026-04-13

Backfilled the initial cyb-os-tiles framerate investigation with hardware smoke evidence, stats-based pacing analysis, a main-only A/B probe, and a raw benchmark follow-up plan.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go — Control benchmark identified for the next measurement pass
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/stats.go — Stats output used to estimate effective scene flush cadence
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/11-cyb-os-tiles.js — Three-display redraw scene used for the initial pacing investigation


## 2026-04-13

Reran the raw hardware FPS benchmark successfully on the current tree and added dedicated JS path probes that exposed the live runner's default 40ms retained flush interval as an approximate 25 FPS cap before scene-specific costs are considered.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-fps-bench/main.go — Fresh raw-writer baseline confirmed ~36 FPS full-screen
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/renderer.go — Default 40ms retained flush interval explains the live-runner cap found by JS probes
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/01-js-path-probe-main-only.js — Main-only probe measured effective live-runner throughput near 25 FPS
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/scripts/02-js-path-probe-three-display.js — Three-display probe showed ~25 frame-equivalents/sec with proportionally more commands


## 2026-04-13

Exposed --flush-interval in loupe-js-live and verified on hardware that reducing the retained flush interval from 40ms to 20ms raises the effective live-runner throughput ceiling.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/options.go — New --flush-interval flag
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/run.go — Live runner now connects with explicit render options so flush interval is controllable


## 2026-04-13

Added a long-form project report that consolidates the cleanup narrative from LOUPE-008 with the later LOUPE-013 performance findings, including the retained flush-interval cap and the raw-vs-live measurement distinction.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/13/LOUPE-013--cyb-os-tiles-framerate-investigation-and-raw-transport-benchmark-refresh/design/01-project-report-cleanup-and-performance.md — Ticket-local project report copy of the cleanup and performance writeup

