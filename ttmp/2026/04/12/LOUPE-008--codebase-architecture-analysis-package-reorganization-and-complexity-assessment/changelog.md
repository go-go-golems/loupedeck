# Changelog

## 2026-04-12

- Initial workspace created


## 2026-04-12

Step 1: Completed codebase structure mapping - identified 3434 lines in root, 4219 in runtime/, key complex files

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/displayknob.go — Most complex file (426 lines) - mixes hardware


## 2026-04-12

Step 2: Deep analysis of complex files - surface.go, module_ui/module.go, runner.go, visual_runtime.go

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — JS UI bindings with repetitive boilerplate (383 lines)


## 2026-04-12

Step 3: Package boundary analysis - identified split opportunities in runtime/js/ and root package

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/runner.go — Well-structured thread-safe JS execution


## 2026-04-12

Step 5: Senior analysis completed - identified god package, dead widget system, triplicated names, missing hardware/framework boundary. Wrote Design Doc 02 with concrete 7-step refactoring plan (~8.5 hours effort).

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/loupedeck.go — God struct with 30+ fields - needs shedding legacy subsystems


## 2026-04-12

Step 6: Big-brother review completed. Graded prior docs (C+ and B+), verified obsolete API usage with ripgrep, and wrote Design Doc 03 with a no-compatibility refactor plan centered on deletion, connect-time profiling, and a narrower hardware driver boundary.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/connect.go — Connect path currently leaves device partially initialized and hardcodes Model=foo


## 2026-04-13

Checkpointed pkg/device migration and removed legacy widget stack (commit 8018a20)

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-feature-tester/main.go — Deleted obsolete binary that depended on removed widget stack
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/listeners.go — Listener-only event model after dead-code removal
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/loupedeck.go — Driver package simplified after widget/font removal


## 2026-04-13

Initialized device profiles during connect and removed manual SetDisplays calls (commit 2dac4b1)

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — No more post-connect display initialization
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/connect.go — Connect path now resolves model and displays
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/profile.go — Device profile table and display specs


## 2026-04-13

Added canonical input naming/parsing APIs and removed duplicated JS/live-runner maps (commit 41d4f67)

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/cmd/loupe-js-live/main.go — Live runner logs canonical device names
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/device/inputs.go — Single source of truth for button/knob/touch naming
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go — JS event registration now validates through device parse helpers

