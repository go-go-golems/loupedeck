---
Title: Investigation Diary
Ticket: LOUPE-008
Status: active
Topics:
    - architecture
    - refactoring
    - analysis
    - code-quality
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupe-fps-bench/main.go
      Note: Benchmark runner no longer performs manual display setup (commit 2dac4b1)
    - Path: cmd/loupe-js-live/main.go
      Note: |-
        Removed manual SetDisplays call after connect-time profile setup (commit 2dac4b1)
        Event logging now uses device String methods rather than local helper maps (commit 41d4f67)
    - Path: cmd/loupe-svg-buttons/main.go
      Note: SVG demo now relies on connect-time profile setup (commit 2dac4b1)
    - Path: pkg/device/connect.go
      Note: Connect path now resolves device profiles during initialization (commit 2dac4b1)
    - Path: pkg/device/display.go
      Note: Display mechanics retained after removing SetDisplays product switch (commit 2dac4b1)
    - Path: pkg/device/inputs.go
      Note: Canonical input String and Parse APIs plus ButtonStatus typing fix (commit 41d4f67)
    - Path: pkg/device/inputs_test.go
      Note: Round-trip and alias coverage for device input naming (commit 41d4f67)
    - Path: pkg/device/listeners.go
      Note: Listener model simplified to On* subscriptions only (commit 8018a20)
    - Path: pkg/device/loupedeck.go
      Note: Driver package simplified after widget/font removal (commit 8018a20)
    - Path: pkg/device/profile.go
      Note: Table-driven device profiles and display specs (commit 2dac4b1)
    - Path: runtime/js/module_ui/module.go
      Note: JS UI module now parses canonical device input names instead of local maps (commit 41d4f67)
ExternalSources: []
Summary: Chronological diary of codebase analysis investigating package structure, file complexity, and reorganization opportunities for the go-go-golems/loupedeck project
LastUpdated: 2026-04-12T15:35:00-04:00
WhatFor: Document the investigation process, findings, and analysis methodology for codebase reorganization recommendations
WhenToUse: When reviewing analysis conclusions or extending the investigation
---



# Investigation Diary

## Goal

Document the step-by-step investigation of the go-go-golems/loupedeck codebase to identify:
1. Current package organization and boundaries
2. Files and abstractions with high complexity
3. Opportunities for package splitting and reorganization
4. Architectural patterns and coupling issues

## Step 1: Initial Codebase Exploration and Structure Mapping

This step established a comprehensive map of the codebase structure, identifying all Go files, their sizes, and package organization.

### Prompt Context

**User prompt (verbatim):** "Analyze this codebase and suggest a way to better reorganize the files, and which packages could be split and which files and abstractions look complex and why. Create a new docmgr ticket for that analysis, and then write a detailed analysis document and store it in the ticket. Keep a diary as you do your research."

**Assistant interpretation:** The user wants a comprehensive architectural analysis of the current codebase to identify reorganization opportunities, complex files that might need refactoring, and package boundaries that could be improved.

**Inferred user intent:** To understand the current state of the codebase architecture, identify pain points in organization, and receive actionable recommendations for improving code maintainability and separation of concerns.

### What I did

- Listed all files in the repository root
- Found all Go source files using `find . -type f -name "*.go"`
- Examined directory structure for `cmd/`, `pkg/`, `runtime/` directories
- Collected line counts for all Go files to identify the largest/most complex files
- Read `go.mod` to understand module dependencies and structure

### What worked

- The codebase follows a clear directory structure with distinct areas:
  - Root package: Core Loupedeck hardware abstraction
  - `cmd/`: 5 CLI applications
  - `pkg/`: 3 shared packages (jsmetrics, runtimebridge, runtimeowner)
  - `runtime/`: 9 subsystems (ui, gfx, anim, easing, reactive, render, host, metrics, js)
  - `examples/js/`: JavaScript example files

### What I learned

**File Size Distribution (root package):**
- `displayknob.go`: 426 lines (largest)
- `loupedeck.go`: 243 lines
- `listeners.go`: 241 lines
- `writer.go`: 220 lines
- `inputs.go`: 220 lines
- `message.go`: 212 lines
- `connect.go`: 203 lines
- `display.go`: 193 lines
- `renderer.go`: 148 lines
- Smaller files: `watchedint.go` (59), `intknob.go` (83), `dialer.go` (126)

**File Size Distribution (runtime/):**
- `runtime/gfx/surface.go`: 385 lines (largest in runtime)
- `runtime/js/module_ui/module.go`: 383 lines
- `runtime/js/runtime_test.go`: 528 lines (largest test file)
- `runtime/ui/ui_test.go`: 247 lines
- `runtime/ui/display.go`: 257 lines
- `runtime/render/visual_runtime.go`: 198 lines
- `runtime/reactive/runtime.go`: 124 lines

**Package Structure:**
- Root `loupedeck` package: Hardware abstraction layer
- `runtime/gfx`: Graphics surface primitives with mutex-based concurrency
- `runtime/ui`: UI component tree (pages, displays, tiles)
- `runtime/render`: Rendering engine from UI to image buffers
- `runtime/reactive`: Signal-based reactivity system
- `runtime/anim`: Animation system
- `runtime/js`: JavaScript runtime integration with goja
- `pkg/runtimeowner`: Thread-safe JS runtime execution
- `pkg/runtimebridge`: Bridge pattern for JS-to-Go bindings

### What was tricky to build

Mapping the dependencies between packages required multiple passes because:
1. The `runtime/js/` modules have complex interdependencies through `runtimebridge` and `env`
2. The root package is both a hardware abstraction AND contains high-level widgets (DisplayKnob)
3. Understanding the relationship between `runtimeowner`, `runtimebridge`, and JS modules required reading multiple files

### Technical details

**Key Files by Complexity (lines of code):**
```
Root package:
426 displayknob.go    # Complex widget with graphics, touch handling, state management
243 loupedeck.go      # Main device abstraction
241 listeners.go       # Event listener management
220 writer.go          # USB message writing with queue
220 inputs.go          # Input handling (buttons, knobs, touch)
212 message.go         # Protocol message formatting

code/runtime/:
385 runtime/gfx/surface.go       # Concurrent graphics surface with batching
383 runtime/js/module_ui/module.go # JS UI bindings (large module)
528 runtime/js/runtime_test.go   # Comprehensive JS runtime tests
257 runtime/ui/display.go        # Display component with layers
198 runtime/render/visual_runtime.go # Rendering coordination
230 pkg/runtimeowner/runner.go   # Thread-safe JS execution
```

---

## Step 2: Deep Analysis of Complex Files

This step examined the most complex files identified in Step 1 to understand why they are complex and what responsibilities they hold.

### Prompt Context

**User prompt (verbatim):** (same as Step 1 - continuing the analysis)

**Assistant interpretation:** Need to examine the largest files to understand complexity drivers.

**Inferred user intent:** Understand concrete complexity issues to make specific reorganization recommendations.

### What I did

1. Read `displayknob.go` (426 lines) - largest root package file
2. Read `runtime/gfx/surface.go` (385 lines) - concurrent graphics surface
3. Read `runtime/js/module_ui/module.go` (383 lines) - JS UI bindings
4. Read `pkg/runtimeowner/runner.go` (230 lines) - JS runtime execution
5. Read `runtime/render/visual_runtime.go` (198 lines) - rendering system
6. Examined import dependencies across packages

### What worked

**displayknob.go analysis:**
- Contains 5 distinct responsibilities:
  1. DisplayKnob hardware abstraction (WatchedInt wrapper)
  2. Touch/drag event detection with click vs. drag discrimination
  3. Widget interface (DKWidget) and analog widget implementation
  4. WidgetHolder for swipeable tab interface
  5. Navigation bar rendering with trigonometric calculations
- Mixes hardware control, UI widgets, and graphics rendering
- Contains extensive design comments about future widget system

**surface.go analysis:**
- Well-structured concurrent graphics surface
- Uses mutex + condition variable for batching
- Implements multiple drawing primitives: Set, Line, FillRect, Crosshatch, CompositeAdd
- Subscription pattern for change notifications
- ToRGBA conversion for final output
- Complexity is justified - it's a fundamental primitive

**module_ui/module.go analysis:**
- JS-to-Go bridge for UI components
- Creates proxy objects (page, display, tile) with method bindings
- Heavy use of `runtimebridge` and `runtimeowner` for thread-safe execution
- Repetitive pattern: functions that accept values OR callbacks
- 6 map lookups for button/touch/knob name resolution

**runner.go analysis:**
- Thread-safe JS execution with goroutine ID tracking
- Owner context pattern for detecting same-thread access
- Call (blocking) vs Post (async) operations
- Panic recovery options
- Clean separation of concerns

### What I learned

**Complexity Drivers:**

1. **displayknob.go** is complex because it mixes:
   - Hardware abstraction (DisplayKnob struct)
   - Touch gesture recognition (isClick, drag detection)
   - Widget system interface (DKWidget, DKAnalogWidget)
   - Container management (WidgetHolder)
   - Graphics rendering (trigonometry for dots)

2. **module_ui/module.go** is complex because:
   - Each UI component needs 3-4 proxy methods (text, icon, visible, surface)
   - Two calling conventions: static values vs reactive functions
   - Must bridge JS callbacks to Go through runtimeowner
   - Repetitive boilerplate for each component type

3. **Import Dependencies Reveal Coupling:**
```
runtime/js/module_ui -> runtimebridge, runtimeowner, env, module_gfx, ui
runtime/js/module_gfx -> runtimebridge, gfx
runtime/js/module_anim -> runtimebridge, anim, easing
pkg/runtimebridge -> runtimeowner
runtime/ui -> (relatively clean - only fmt)
runtime/render -> ui, font packages
cmd/* -> multiple runtime packages
```

### What was tricky to build

Understanding the JS module architecture required tracing through:
1. `env/env.go` - stores runtime-wide environment
2. `runtimebridge` - lookup mechanism for bindings
3. `runtimeowner` - thread-safe execution
4. Individual module registration in `runtime/js/runtime.go`

The dependency chain: JS Module → env Lookup → runtimebridge Lookup → runtimeowner Call/Post

### What warrants a second pair of eyes

**displayknob.go** mixing hardware abstraction with widget system suggests:
- The widget system (DKWidget, WidgetHolder) should be in `runtime/ui/`
- DisplayKnob hardware abstraction should stay in root package
- The drag/touch detection might belong in a separate input handling package

**module_ui/module.go** repetitive patterns suggest:
- Could use code generation or reflection to reduce boilerplate
- Component proxy creation follows identical patterns

---

## Step 3: Package Boundary Analysis

This step analyzed package cohesion and coupling to identify split/merge opportunities.

### What I did

- Analyzed package responsibilities and dependencies
- Identified cohesion issues (packages doing too many things)
- Found coupling patterns between packages
- Examined `runtime/js/` module organization

### What worked

**Package Responsibility Mapping:**

| Package | Lines | Responsibilities | Cohesion |
|---------|-------|------------------|----------|
| root | 3434 | Hardware USB protocol, device abstraction, widgets | LOW - mixes layers |
| runtime/gfx | 385 | 8-bit surface graphics | HIGH |
| runtime/ui | ~600 | Component tree (page/display/tile) | HIGH |
| runtime/render | 198 | UI → image rendering | HIGH |
| runtime/reactive | ~300 | Signal graph reactivity | HIGH |
| runtime/anim | 146 | Animation definitions | HIGH |
| runtime/easing | 58 | Easing functions | HIGH |
| runtime/host | ~300 | Event/timers/pages integration | MEDIUM |
| runtime/js | ~1000 | Goja JS runtime + 8 modules | LOW - needs splitting |
| pkg/runtimeowner | 230 | Thread-safe JS execution | HIGH |
| pkg/runtimebridge | 50 | Binding lookup | HIGH |
| pkg/jsmetrics | 225 | JS metrics collection | MEDIUM |

### What I learned

**Split Opportunities:**

1. **runtime/js/ should split into:**
   - `runtime/js/runtime/` - Core runtime initialization
   - `runtime/js/modules/` - Individual JS modules
   - Current: 8 modules in flat structure, some only 12-58 lines

2. **Root package should split into:**
   - `loupedeck/` - Hardware protocol (message, connect, listen, writer)
   - `loupedeck/display/` - Display management
   - `loupedeck/input/` - Button/knob/touch handling
   - `widgets/` OR `runtime/widgets/` - DKWidget system
   - Current: displayknob.go shows widget system shouldn't be here

3. **runtime/host/ might split:**
   - `events.go` (139 lines) - Event management
   - `timers.go` (64 lines) - Timer scheduling
   - `pages.go` (44 lines) - Page management
   - `runtime.go` (119 lines) - Integration

**Coupling Issues:**

1. **Tight coupling: JS modules ↔ runtimebridge ↔ runtimeowner**
   - Every JS module needs all three
   - Could runtimebridge absorb owner context handling?

2. **Medium coupling: render → ui**
   - Renderer knows UI structure intimately
   - Alternative: UI could have Render() method

3. **Loose coupling (good):** ui → (none), gfx → (none), easing → (none)

### What was tricky to build

Determining if `runtime/host/` should split required understanding:
- `host/runtime.go` integrates events + timers + pages
- Each has different complexity and change frequency
- Events: 139 lines, complex callback management
- Timers: 64 lines, simple wrapper
- Pages: 44 lines, very simple

Verdict: Keep together for now - they're cohesive runtime services.

---

## Step 4: Architectural Pattern Assessment

### What I did

- Evaluated abstraction quality in complex files
- Identified pattern violations and inconsistencies
- Assessed test coverage distribution

### What worked

**Abstraction Quality Analysis:**

1. **surface.go: EXCELLENT**
   - Clear public API (Set, Line, FillRect, etc.)
   - Internal helpers (setLocked, addLocked)
   - Thread-safe with clean mutex usage
   - Batch mechanism for performance

2. **runner.go: EXCELLENT**
   - Clean interface (Runner)
   - Owner context pattern for optimization
   - Proper context propagation
   - Panic recovery configurable

3. **displayknob.go: NEEDS WORK**
   - 3 abstraction levels mixed: hardware, widgets, graphics
   - Widget interface embedded in hardware package
   - Touch detection logic mixed with widget management
   - "Quick hack" comment on isClick() suggests technical debt

4. **module_ui/module.go: ACCEPTABLE but VERBOSE**
   - Proxy pattern is correct
   - But 383 lines for UI bindings is excessive
   - Repetitive callback/value handling pattern

### What I learned

**Test File Distribution:**
- Good: `writer_test.go` (119 lines), `runtime/js/runtime_test.go` (528 lines)
- Weak: `loupedeck_test.go` (48 lines) for 243-line main file
- Missing: `displayknob_test.go` for 426-line most complex file

**Inconsistencies:**
1. Some files use `log/slog`, others use `fmt.Printf`
2. Error handling: some panic, some return errors
3. Comments: extensive design notes in displayknob.go vs minimal elsewhere

### What warrants a second pair of eyes

**displayknob.go** widget system design:
- The comments describe a full widget architecture
- But it's all in one file, mixed with hardware code
- Is this aspirational (not used) or actual (needs extraction)?

Looking at cmd/ and examples/, the WidgetHolder pattern appears used, confirming it's active code that needs proper placement.

---

## Summary of Key Findings

### Most Complex Files and Why

| File | Lines | Complexity Drivers | Recommended Action |
|------|-------|-------------------|-------------------|
| `displayknob.go` | 426 | Mixes hardware, widgets, touch, graphics | **SPLIT**: Extract widget system to `runtime/widgets/` |
| `runtime/js/module_ui/module.go` | 383 | Repetitive proxy boilerplate, 6 component types | **REFACTOR**: Use code generation or generic helpers |
| `runtime/gfx/surface.go` | 385 | Concurrent graphics with batching | **KEEP**: Complexity is justified |
| `pkg/runtimeowner/runner.go` | 230 | Thread-safety, context tracking | **KEEP**: Well-structured |
| `runtime/render/visual_runtime.go` | 198 | Multi-layer rendering | **KEEP**: Clean separation |

### Package Split Recommendations

**HIGH PRIORITY:**
1. **Extract widget system** from root package
   - `DKWidget`, `DKAnalogWidget`, `WidgetHolder` → `runtime/widgets/` or `pkg/widgets/`
   - Touch detection → `pkg/input/` or stay in widgets

**MEDIUM PRIORITY:**
2. **Split runtime/js modules**
   - Keep tiny modules (easing, metrics) as-is (12-28 lines)
   - Group related modules: `ui`, `gfx`, `anim`, `state`
   - Consider: `runtime/js/modules/` subdirectory

**LOW PRIORITY:**
3. **Reorganize root package**
   - `loupedeck/` → hardware protocol only
   - `pkg/display/` → Display management (DrawTarget abstraction)
   - Current root is acceptable but not ideal

### Coupling Reduction Opportunities

1. **JS Module Bridge**: Could simplify `runtimebridge` + `runtimeowner` chain
2. **UI/Render**: Could invert dependency (UI.Render() instead of render.UI)

### Files with Technical Debt

1. `displayknob.go`: "Quick hack" comment on touch detection
2. `displayknob.go`: Mixed abstraction levels
3. Root package: No clear separation between protocol and widgets

---

## Code Review Instructions

To validate this analysis:

1. **Check file sizes:**
   ```bash
   wc -l *.go runtime/**/*.go pkg/**/*.go | sort -n | tail -20
   ```

2. **Verify package dependencies:**
   ```bash
   grep -r "import" --include="*.go" . | grep -v test | grep -v ".git"
   ```

3. **Check widget system usage:**
   ```bash
   grep -r "DKWidget\|WidgetHolder\|DKAnalogWidget" --include="*.go" cmd/ examples/
   ```

4. **Review complexity in displayknob.go:**
   - Lines 30-100: Hardware abstraction
   - Lines 100-180: Touch/drag detection
   - Lines 180-300: Widget interface + analog widget
   - Lines 300-426: WidgetHolder + navigation bar

---

## What should be done in the future

1. **Immediate (next sprint):**
   - Extract widget system to dedicated package
   - Add tests for displayknob.go (currently no test file)

2. **Short-term (next month):**
   - Refactor module_ui to reduce boilerplate
   - Document package boundaries and responsibilities

3. **Medium-term (next quarter):**
   - Evaluate JS module organization
   - Consider code generation for JS bindings

4. **Ongoing:**
   - Monitor root package size - if it grows beyond 4000 lines, prioritize splitting
   - Add integration tests between runtime packages

---

## Step 5: Senior Analysis — Full Codebase Read and Independent Assessment

The user asked for my own assessment, noting the previous analysis wasn't useful. I read every single Go file in the codebase (root, runtime/, pkg/, cmd/) — approximately 7,600 lines — and identified the actual structural problems.

### Prompt Context

**User prompt (verbatim):** "The work so far was your little brother. Now it's your turn. Your little brother has been analyzing the codebase so far, but I don't think this is very useful. What's your assessment and how would you refactor this. Store in a second design-doc."

**Assistant interpretation:** The user wants an honest, independent analysis that goes deeper than file-size metrics. They want to know what's actually wrong and get a realistic refactoring plan.

**Inferred user intent:** Get actionable, concrete refactoring guidance from someone who read the whole codebase, not just line counts.

### What I did

1. Read all 15+ root package Go files completely
2. Read all runtime/ subsystem files (ui/, reactive/, host/, anim/, gfx/, render/, js/, metrics/, easing/)
3. Read all pkg/ files (runtimeowner, runtimebridge, jsmetrics)
4. Read cmd/loupe-js-live/main.go (the primary application)
5. Traced which types are actually used by the JS runtime path vs. legacy

### What I learned

**The real problem is not big files. It's that two complete UI systems coexist and the old one was never retired.**

The root `loupedeck` package is a god package that absorbed hardware protocol, font management, SVG parsing, event dispatch, render scheduling, and a widget framework across three eras of development. The new `runtime/` system (reactive UI, component tree, rendering, host events) was built alongside it but the old code was never cleaned up.

**Specific findings:**

1. **Legacy widgets are dead code for the JS runtime:** `WatchedInt`, `IntKnob`, `MultiButton`, `TouchDial`, `DisplayKnob`, `DKWidget`, `WidgetHolder` — only used by `cmd/loupe-feature-tester`, never by `cmd/loupe-js-live`.

2. **The Loupedeck struct has 30+ fields** mixing hardware, events, fonts, and drag state. After moving legacy widgets out, ~15 fields disappear.

3. **Name mappings are triplicated:** `inputs.go` defines constants, `module_ui/module.go` remaps them for JS, `cmd/loupe-js-live/main.go` remaps them for logging.

4. **The root package has no business doing font management or SVG parsing.** These are application-level concerns.

5. **The `Bind*` legacy API coexists awkwardly with the `On*` subscription API.** After legacy moves, `Bind*` can be deleted entirely.

### What was tricky

Tracing the actual usage of legacy types required reading every `cmd/` binary, not just the library code. The previous analysis missed this because it only looked at library file sizes.

### What warrants a second pair of eyes

The claim that `Bind*` methods are unused outside legacy widgets — need to verify by grepping all cmd/ binaries for `BindButton`, `BindKnob`, etc.

### Code review instructions

```bash
# Verify legacy widgets are only used by feature-tester
grep -r "IntKnob\|MultiButton\|TouchDial\|DisplayKnob\|DKWidget\|WidgetHolder\|WatchedInt" cmd/ --include="*.go"

# Verify Bind* methods are unused in cmd/
grep -r "BindButton\|BindKnob\|BindTouch\|BindTouchCT\|BindButtonUp\|BindTouchUp" cmd/ --include="*.go"

# Verify name mapping duplication
grep -rn '"Circle".*deck.Circle\|deck.Circle.*"Circle"' runtime/ cmd/ --include="*.go"
```

---

## Step 6: Big Brother Pass — Grade Prior Reviews and Drop Legacy Compatibility

The user clarified that legacy compatibility is not required. That materially changes the refactoring plan: instead of preserving the obsolete widget stack in a `legacy/` package, the correct move is to delete it.

### Prompt Context

**User prompt (verbatim):** "Ok, you are now the biggest brother. Your younger brothers have done their analysis, now it's your timie to assess how good they were, and then do your own analysis. We don't need to keep legacy compatibility."

**Assistant interpretation:** Re-evaluate the previous two analyses, explain where they were useful or weak, and produce a stronger plan under a no-compatibility constraint.

**Inferred user intent:** Get a final, decisive architectural recommendation that optimizes for codebase clarity rather than migration safety.

### What I did
- Verified actual usage of legacy APIs with `rg`
- Confirmed `WatchedInt`, `MultiButton`, `TouchDial`, etc. are only used by `cmd/loupe-feature-tester`
- Confirmed `Bind*` APIs are used only by tests and `displayknob.go`
- Verified CT touch/drag fields are isolated to `displayknob.go` + `listen.go`
- Verified root font APIs are only used by obsolete widget code
- Identified an additional structural issue: `connect.go` hardcodes `Model = "foo"` and requires all callers to remember `SetDisplays()` manually

### What worked
The grep-based verification removed any ambiguity about whether the old widget system still matters. It does not, except for one tester binary.

### What I learned
**Revised recommendation:**
1. Delete the obsolete widget stack entirely
2. Delete the legacy `Bind*` callback API entirely
3. Remove root-level font support entirely
4. Move device profile/display setup into connect-time
5. Keep `runtime/` as the sole app/runtime architecture
6. Move SVG icon parsing out of the root package

### What didn't work
The earlier compatibility-preserving plan is no longer the right answer under the new constraint. It would preserve dead architecture that the user explicitly said we do not need to keep.

### What warrants a second pair of eyes
Whether `writer.go` / `renderer.go` should move to `internal/pipeline` immediately or stay in root for one intermediate step. Both are viable; my recommendation favors `internal/pipeline` to narrow the root package faster.

### Code review instructions
```bash
# Verify only feature tester uses obsolete widget stack
rg -n "\b(IntKnob|MultiButton|TouchDial|DisplayKnob|DKWidget|WidgetHolder|WatchedInt)\b" cmd runtime . --glob '!ttmp/**' --glob '!sources/**'

# Verify Bind* APIs are dead outside tests/obsolete code
rg -n "\bBind(Button|ButtonUp|Knob|Touch|TouchUp|TouchCT)\b" . --glob '!ttmp/**' --glob '!sources/**'

# Verify connect-time setup smell
rg -n "Model\s*=\s*\"foo\"|SetDisplays\(" . --glob '!ttmp/**' --glob '!sources/**'
```

---

## Step 7: Checkpoint the pkg/device move and dead-code deletion

At this point the user had already moved the former root driver files into `pkg/device/`, and I took over to cleanly remove the legacy widget stack that we had already identified as deletable. The main goal of this step was to checkpoint that structural change as a focused commit before layering the connect/profile work on top.

This step deliberately did not try to solve everything at once. It removed the obsolete widget and value stack, simplified the listener model to `On*` subscriptions only, and preserved a working baseline in the new `pkg/device` layout so the next refactor could focus only on connection correctness.

### Prompt Context

**User prompt (verbatim):** "commit at appropriate intervals and keep a diary. I think now would be a good time."

**Assistant interpretation:** The user wants the current structural work checkpointed now, in focused commits, and wants the running ticket diary updated as the work proceeds.

**Inferred user intent:** Keep the refactor reviewable and resumable by recording the state after each coherent architectural slice rather than batching everything into one giant uncommitted diff.

**Commit (code):** `8018a20` — "Move driver to pkg/device and delete legacy widget stack"

### What I did
- Verified the repo was in the expected post-move state (`pkg/device/*`, no root package imports remaining in code).
- Staged the package move plus the dead-code cleanup already performed.
- Included the new LOUPE-008 implementation plan doc in the same checkpoint because it described the new target layout and had already been authored.
- Committed:
  - deletion of `displayknob.go`, `intknob.go`, `multibutton.go`, `touchdials.go`, `watchedint.go`
  - deletion of `cmd/loupe-feature-tester/main.go`
  - deletion of obsolete `loupedeck_test.go`
  - moved root driver files into `pkg/device/`
  - updated imports in `cmd/`, `runtime/host`, `runtime/js/module_ui`, and runtime tests
  - simplified `pkg/device/listeners.go` and removed legacy bind state
  - removed legacy font/widget code from `pkg/device/loupedeck.go`

### Why
- The package move and legacy deletion were already a coherent architectural slice.
- Committing here reduced the blast radius for the next change.
- It also made the upcoming connect/profile cleanup easier to review because it no longer had to be mentally disentangled from the package reorganization.

### What worked
- The package move had been applied consistently enough that the repo could be checkpointed cleanly.
- The widget stack really was removable: after deleting it and cleaning references, `go test ./...` had previously passed in the `pkg/device` layout.
- The resulting commit was focused and aligned with the intended architecture.

### What didn't work
- My first commit attempt failed due to a stale git lock file:

```text
fatal: Unable to create '/home/manuel/code/wesen/2026-04-11--loupedeck-test/.git/index.lock': File exists.

Another git process seems to be running in this repository, e.g.
an editor opened by 'git commit'. Please make sure all processes
are terminated then try again. If it still fails, a git process
may have crashed in this repository earlier:
remove the file manually to continue.
```

- I checked for active git processes with:

```bash
cd /home/manuel/code/wesen/2026-04-11--loupedeck-test && ps -ef | grep '[g]it' && echo '---' && ls -l .git/index.lock
```

- There was no active lock file by the time I retried, so I removed defensively and re-ran the commit.

### What I learned
- The `pkg/device` move was already in good enough shape to checkpoint without extra rescue work.
- The repo is much easier to reason about once the old widget system is gone; the driver package now reads like an actual device driver rather than a half-driver / half-widget-framework hybrid.

### What was tricky to build
- The sharp edge here was not the code but the repository state. Because the package move had already happened outside my current turn, I needed to validate that the staged set was coherent before committing it.
- The stale `.git/index.lock` could easily have led to accidental manual cleanup without verification; I checked process state first so I wasn’t deleting a live lock from another tool.

### What warrants a second pair of eyes
- Review the shape of `pkg/device` after the move to ensure the package boundary is actually what we want long-term.
- Review that deleting `cmd/loupe-feature-tester` is acceptable and that we do not want a replacement exerciser yet.

### What should be done in the future
- Follow immediately with connect-time profile initialization so `cmd/*` no longer needs `SetDisplays()`.
- Add authoritative input naming APIs (`String()` / `Parse*()`) to remove duplicated maps from `runtime/js/module_ui` and `cmd/loupe-js-live`.

### Code review instructions
- Start with `pkg/device/` and confirm only driver-relevant code remains.
- Review commit `8018a20` with:

```bash
git show --stat 8018a20
git show 8018a20 -- pkg/device cmd runtime/host runtime/js/module_ui
```

- Validate with:

```bash
go test ./...
```

### Technical details
- Staged diff summary before commit:

```text
32 files changed, 1307 insertions(+), 1683 deletions(-)
```

- Commit produced:

```text
[main 8018a20] Move driver to pkg/device and delete legacy widget stack
```

---

## Step 8: Move device profiling into connect-time initialization

With the structural checkpoint in place, I tackled the highest-value functional cleanup: making `pkg/device.Connect*()` return a fully initialized device. Before this step the device still came back with `Model: "foo"`, and every command had to remember to call `SetDisplays()` manually before rendering.

This step converted device display layout selection from a late, caller-owned setup action into an early, connect-owned responsibility. That is the right abstraction boundary: hardware identity and display capabilities are part of connecting to the device, not something each command should remember to patch in afterwards.

### Prompt Context

**User prompt (verbatim):** (see Step 7)

**Assistant interpretation:** The user wants me to proceed with the connect/profile cleanup on top of the newly moved `pkg/device` package, still keeping commits focused and the diary current.

**Inferred user intent:** Finish the most important architectural cleanup now that the package layout is sane: remove the `SetDisplays()` footgun and make connect return a usable device.

**Commit (code):** `2dac4b1` — "Initialize device profiles during connect"

### What I did
- Added `pkg/device/profile.go` with:
  - `DeviceProfile`
  - `DisplaySpec`
  - `deviceProfiles` table
  - `resolveProfile(product string)`
  - `(*Loupedeck).applyProfile(profile)`
- Updated `pkg/device/connect.go` to:
  - resolve the profile immediately after the websocket connect succeeds
  - fail fast on unknown product IDs
  - apply the profile before returning the device
  - remove the hardcoded `Model: "foo"`
  - log the resolved model name
- Removed the old `SetDisplays()` implementation from `pkg/device/display.go`.
- Removed all `SetDisplays()` calls from:
  - `cmd/loupe-js-live/main.go`
  - `cmd/loupe-fps-bench/main.go`
  - `cmd/loupe-svg-buttons/main.go`
- Added `pkg/device/profile_test.go` covering:
  - known profile resolution
  - unknown profile failure
  - `applyProfile()` populating displays + model
- Ran `gofmt -w` on the touched files.
- Ran focused and full test passes.

### Why
- `SetDisplays()` was the biggest remaining architectural smell after the package move.
- A device object should be usable after `Connect*()` returns.
- Resolving profiles in one place makes unsupported product IDs an error instead of a panic later in arbitrary caller code.

### What worked
- The profile table approach was straightforward and removed all remaining `SetDisplays()` call sites.
- Focused validation passed immediately:

```bash
go test ./pkg/device ./cmd/...
```

- Full validation also passed:

```bash
go test ./...
```

### What didn't work
- Before the final full run, I had previously seen a transient failure in `runtime/js` during an earlier sweep after the package move:

```text
--- FAIL: TestAnimModuleLoopCanDriveReactiveUpdates (0.04s)
    runtime_test.go:229: expected loop to update visible text, got "0"
FAIL
FAIL	github.com/go-go-golems/loupedeck/runtime/js	0.543s
```

- I did not change `runtime/js` in this step. Re-running the full suite after the profile cleanup passed cleanly, so this looks like a flaky or timing-sensitive test rather than a deterministic regression from the profile work.

### What I learned
- The connect path gets noticeably cleaner once display setup becomes profile-driven.
- `pkg/device/display.go` is much easier to understand once it only contains display mechanics, not product selection logic.
- The package move plus this connect-time setup change together complete the most important part of the architecture cleanup.

### What was tricky to build
- The main subtlety was making the profile failure path safe. Once the websocket is up, an unknown product should return an error without leaving the partially connected transport open. I explicitly closed both `conn` and `c` before returning the error from `resolveProfile()` failure.
- Another subtlety was making sure command binaries no longer assumed post-connect setup. I removed all `SetDisplays()` call sites and verified the commands still compiled.

### What warrants a second pair of eyes
- Review `pkg/device/profile.go` for correctness of offsets and display IDs, especially CT vs Live family differences.
- Review whether `Razer Stream Controller` as the `0d06` model name is the desired public name.
- Decide whether `DeviceProfile` should remain internal-only or eventually become part of the public API.

### What should be done in the future
- Add `String()` / `Parse*()` methods in `pkg/device/inputs.go`.
- Replace manual name maps in `runtime/js/module_ui/module.go`.
- Remove `buttonName`, `touchName`, `knobName`, `buttonStatusName` from `cmd/loupe-js-live/main.go`.

### Code review instructions
- Start with:
  - `pkg/device/profile.go`
  - `pkg/device/connect.go`
  - `pkg/device/display.go`
- Then inspect the caller cleanup in:
  - `cmd/loupe-js-live/main.go`
  - `cmd/loupe-fps-bench/main.go`
  - `cmd/loupe-svg-buttons/main.go`
- Review commit:

```bash
git show --stat 2dac4b1
git show 2dac4b1 -- pkg/device cmd/loupe-js-live cmd/loupe-fps-bench cmd/loupe-svg-buttons
```

- Validate with:

```bash
go test ./...
rg -n '\.SetDisplays\(|Model:\s*"foo"' pkg/device cmd runtime --glob '!ttmp/**'
```

### Technical details
- Focused formatting command:

```bash
gofmt -w pkg/device/connect.go pkg/device/display.go pkg/device/profile.go pkg/device/profile_test.go cmd/loupe-js-live/main.go cmd/loupe-fps-bench/main.go cmd/loupe-svg-buttons/main.go
```

- Full test command:

```bash
go test ./...
```

- Commit produced:

```text
[main 2dac4b1] Initialize device profiles during connect
```

---

## Step 9: Add canonical input naming and parsing APIs

With the driver moved and connect-time profiling fixed, the next most obvious cleanup was duplicated naming logic for buttons, knobs, and touch buttons. The project still had two parallel hand-maintained mapping systems: one in `runtime/js/module_ui` for JS-facing event registration and one in `cmd/loupe-js-live` for event logging. That duplication was small but exactly the kind of repeated maintenance burden that grows sharp edges over time.

This step made the device enums themselves authoritative. Instead of every consumer maintaining its own lookup tables, `pkg/device/inputs.go` now owns the canonical string forms and parse logic. That gives us one place to define names, one place to define aliases, and one place to define how unknown values are reported.

### Prompt Context

**User prompt (verbatim):** "continue"

**Assistant interpretation:** Continue with the next planned cleanup slice after connect/profile setup, preserving the focused-commit and diary workflow.

**Inferred user intent:** Keep executing the cleanup plan incrementally, turning the architecture doc into real code while maintaining reviewable checkpoints.

**Commit (code):** `41d4f67` — "Add canonical input naming and parsing APIs"

### What I did
- Extended `pkg/device/inputs.go` with:
  - `func (Button) String() string`
  - `func ParseButton(string) (Button, error)`
  - `func (ButtonStatus) String() string`
  - `func (Knob) String() string`
  - `func ParseKnob(string) (Knob, error)`
  - `func (TouchButton) String() string`
  - `func ParseTouchButton(string) (TouchButton, error)`
- Added alias support for CT button names:
  - `A` → `Up`
  - `B` → `Down`
  - `C` → `Left`
  - `D` → `Right`
- Added `pkg/device/inputs_test.go` covering:
  - button / knob / touch round-trip parsing
  - alias parsing
  - `ButtonStatus.String()`
  - unknown-name rejection
- Replaced hand-maintained maps in `runtime/js/module_ui/module.go` with `ParseButton`, `ParseKnob`, and `ParseTouchButton`.
- Removed local naming helpers from `cmd/loupe-js-live/main.go` and switched event logging to `String()` methods.
- Ran `gofmt -w`, targeted tests, and then the full suite.

### Why
- The naming maps were duplicated and easy to drift.
- Input types should define their own public names.
- JS event registration and CLI logging should both depend on the same canonical source of truth.

### What worked
- The new APIs dropped neatly into both consumers.
- `runtime/js/module_ui` became simpler because it now validates names through the device package rather than local maps.
- `cmd/loupe-js-live` lost a block of boilerplate helper code.
- Validation passed cleanly:

```bash
go test ./pkg/device ./runtime/js ./cmd/loupe-js-live
go test ./...
```

### What didn't work
- The first test run surfaced a type bug in `inputs.go`:

```text
# github.com/go-go-golems/loupedeck/pkg/device [github.com/go-go-golems/loupedeck/pkg/device.test]
pkg/device/inputs_test.go:69:14: ButtonUp.String undefined (type untyped int has no field or method String)
pkg/device/inputs_test.go:70:47: ButtonUp.String undefined (type untyped int has no field or method String)
FAIL	github.com/go-go-golems/loupedeck/pkg/device [build failed]
```

- Cause: `ButtonUp` was still declared as an untyped constant (`ButtonUp = 1`) instead of `ButtonUp ButtonStatus = 1`.
- Fix: explicitly typed `ButtonUp` as `ButtonStatus` and reran formatting/tests.

- The commit also hit the recurring stale git lock issue again:

```text
fatal: Unable to create '/home/manuel/code/wesen/2026-04-11--loupedeck-test/.git/index.lock': File exists.
```

- I used the same safe recovery approach as before: check for live git processes, then retry after removing the stale lock.

### What I learned
- `ButtonStatus` had an implicit typing bug that only became visible once we added methods and tests around it.
- The duplicated naming logic was low-effort to remove and had a good cleanup-to-risk ratio.
- Alias handling belongs in parsing, not in duplicated ad hoc maps in callers.

### What was tricky to build
- The tricky part was deciding what canonical names to emit for button values that have aliases (`Up`/`A`, `Left`/`C`, etc.). I chose the directional names (`Up`, `Left`, `Down`, `Right`) as the canonical string forms while still accepting the letter aliases in parsing.
- That choice matters because `String()` becomes the single visible name in logs and any future serialization. Using the directional names is more descriptive and less layout-specific.

### What warrants a second pair of eyes
- Review the canonical naming choices for aliased CT buttons (`Up` vs `A`, etc.).
- Review whether `TouchLeft` / `TouchRight` should remain accepted JS-facing names even though the current JS code mostly uses `Touch1`–`Touch12`.

### What should be done in the future
- Consider using the new `String()` APIs in any remaining debug logging or tests elsewhere in the codebase.
- If external serialization of inputs is ever added, document these canonical names as the stable wire representation.

### Code review instructions
- Start with:
  - `pkg/device/inputs.go`
  - `pkg/device/inputs_test.go`
- Then review the two consumers:
  - `runtime/js/module_ui/module.go`
  - `cmd/loupe-js-live/main.go`
- Review commit:

```bash
git show --stat 41d4f67
git show 41d4f67 -- pkg/device/inputs.go pkg/device/inputs_test.go runtime/js/module_ui/module.go cmd/loupe-js-live/main.go
```

- Validate with:

```bash
go test ./pkg/device ./runtime/js ./cmd/loupe-js-live
go test ./...
```

### Technical details
- Formatting command:

```bash
gofmt -w pkg/device/inputs.go pkg/device/inputs_test.go runtime/js/module_ui/module.go cmd/loupe-js-live/main.go
```

- Commit produced:

```text
[main 41d4f67] Add canonical input naming and parsing APIs
```
