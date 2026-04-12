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
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary of codebase analysis investigating package structure, file complexity, and reorganization opportunities for the go-go-golems/loupedeck project"
LastUpdated: 2026-04-12T15:35:00-04:00
WhatFor: "Document the investigation process, findings, and analysis methodology for codebase reorganization recommendations"
WhenToUse: "When reviewing analysis conclusions or extending the investigation"
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
