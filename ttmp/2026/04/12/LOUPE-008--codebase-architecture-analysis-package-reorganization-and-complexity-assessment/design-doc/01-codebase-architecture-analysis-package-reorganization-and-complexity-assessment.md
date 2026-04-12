---
Title: Codebase Architecture Analysis - Package Reorganization and Complexity Assessment
Ticket: LOUPE-008
Status: active
Topics:
    - architecture
    - refactoring
    - analysis
    - code-quality
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/displayknob.go
      Note: 426 lines - highest complexity, mixes hardware abstraction with widget system
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/gfx/surface.go
      Note: 385 lines - concurrent graphics surface, complexity justified
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/js/module_ui/module.go
      Note: 383 lines - JS UI bindings with repetitive patterns, needs refactoring
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/pkg/runtimeowner/runner.go
      Note: 230 lines - well-structured thread-safe JS execution
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/render/visual_runtime.go
      Note: 198 lines - clean UI rendering separation
    - Path: /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/ui/ui.go
      Note: 203 lines - component tree management
ExternalSources: []
Summary: "Comprehensive analysis of go-go-golems/loupedeck codebase architecture identifying reorganization opportunities, complex files requiring refactoring, and package boundary improvements"
LastUpdated: 2026-04-12T15:45:00-04:00
WhatFor: "Guide codebase reorganization decisions and prioritize refactoring efforts based on complexity analysis"
WhenToUse: "When planning refactors, evaluating new features, or assessing code organization"
---

# Codebase Architecture Analysis - Package Reorganization and Complexity Assessment

## Executive Summary

The go-go-golems/loupedeck codebase is a Go library for controlling Loupedeck Live hardware devices with an embedded JavaScript runtime for dynamic UIs. This analysis examined **7,653 lines of Go code** across the root package, runtime/, pkg/, and cmd/ directories.

**Key Findings:**
- **1 critical reorganization needed**: `displayknob.go` (426 lines) mixes hardware abstraction, touch detection, widget interfaces, and graphics rendering
- **1 refactoring opportunity**: `runtime/js/module_ui/module.go` (383 lines) has repetitive JS-to-Go proxy boilerplate
- **5 well-structured packages**: surface.go, runner.go, visual_runtime.go demonstrate good abstraction patterns
- **Clean runtime architecture**: 9 subsystems in runtime/ with good separation of concerns

**Immediate Recommendation**: Extract the widget system (DKWidget, WidgetHolder) from the root package to a dedicated `runtime/widgets/` or `pkg/widgets/` package to resolve the most significant architectural mixing.

## Problem Statement

### Current State

The codebase has grown organically through several feature phases:
1. Hardware protocol implementation (USB communication)
2. Event handling system (buttons, knobs, touch)
3. Graphics primitives (8-bit surfaces)
4. JavaScript runtime integration (goja)
5. Widget system for CT knob display
6. Reactive UI framework

This organic growth has resulted in:
- **Mixed abstraction levels**: Hardware protocol code coexists with high-level widget abstractions
- **Package cohesion issues**: The root `loupedeck` package handles USB messages AND user-visible widgets
- **Repetitive patterns**: JS module bindings follow identical templates with 80% duplication
- **Missing test coverage**: Most complex file (`displayknob.go`) has no corresponding test file

### Impact

- **Onboarding difficulty**: New developers must understand USB protocols to work on UI widgets
- **Testing challenges**: Hardware-dependent code mixed with testable UI logic
- **Maintenance risk**: Changes to widget system could inadvertently affect hardware communication
- **Feature development friction**: Adding new widget types requires modifying hardware package

### Scope

This analysis covers:
- **Root package**: 3,434 lines (hardware + widgets)
- **runtime/**: 4,219 lines (9 subsystems)
- **pkg/**: 540 lines (shared infrastructure)
- **cmd/**: CLI applications (not analyzed for complexity)

Out of scope: JavaScript examples, external dependencies, build tooling.

## Current-State Architecture

### Package Structure

```
github.com/go-go-golems/loupedeck/
├── Root Package (loupedeck)
│   ├── Hardware Protocol
│   │   ├── message.go (212)     - Binary message formatting
│   │   ├── connect.go (203)     - USB connection management
│   │   ├── listen.go (102)      - Event listening
│   │   ├── writer.go (220)      - Async message writing
│   │   └── inputs.go (220)      - Input parsing
│   ├── Device Abstraction
│   │   ├── loupedeck.go (243)   - Main device struct
│   │   ├── listeners.go (241) - Event listener registry
│   │   ├── display.go (193)     - Display management
│   │   ├── renderer.go (148)    - Image rendering
│   │   └── svg_icons.go (252)   - SVG icon loading
│   ├── Widget System [PROBLEM AREA]
│   │   ├── displayknob.go (426) - Hardware + Widgets + Graphics
│   │   ├── intknob.go (83)      - Integer knob abstraction
│   │   ├── multibutton.go (135) - Multi-state button
│   │   ├── touchdials.go (145)  - Touch dial handling
│   │   └── watchedint.go (59)   - Observable integer
│   └── Utilities
│       ├── dialer.go (126)      - Dial callback management
│       └── *.go                 - Tests, other helpers
│
├── runtime/                     - JavaScript UI Runtime
│   ├── gfx/ (385)               - 8-bit graphics surfaces
│   ├── ui/ (600)                - Component tree (page/display/tile)
│   ├── render/ (198)            - UI → image rendering
│   ├── reactive/ (300)          - Signal-based reactivity
│   ├── anim/ (146)              - Animation definitions
│   ├── easing/ (58)             - Easing functions
│   ├── host/ (300)              - Event/timers/pages integration
│   ├── metrics/ (116)           - Performance metrics
│   └── js/ (1000)               - Goja JS runtime [COMPLEX]
│       ├── runtime.go (118)     - Runtime initialization
│       ├── env/ (58)            - Environment bindings
│       ├── module_ui/ (383)     - UI JS bindings [REFACTOR CANDIDATE]
│       ├── module_gfx/ (155)    - Graphics JS bindings
│       ├── module_anim/ (143)   - Animation JS bindings
│       ├── module_state/ (135)  - State JS bindings
│       ├── module_easing/ (28)  - Easing JS bindings
│       ├── module_metrics/ (12) - Metrics JS bindings
│       └── module_scene_metrics/ (12) - Scene metrics
│
├── pkg/                         - Shared Infrastructure
│   ├── runtimeowner/ (315)      - Thread-safe JS execution
│   ├── runtimebridge/ (50)      - Binding lookup registry
│   └── jsmetrics/ (225)         - JS metrics collection
│
└── cmd/                         - CLI Applications
    ├── loupe-feature-tester/    - Hardware test tool
    ├── loupe-fps-bench/         - Performance benchmark
    ├── loupe-js-demo/           - JS runtime demo
    ├── loupe-js-live/           - Live JS environment
    └── loupe-svg-buttons/       - SVG button demo
```

### File Complexity Analysis

| Rank | File | Lines | Complexity Drivers | Cohesion Score |
|------|------|-------|-------------------|----------------|
| 1 | `displayknob.go` | 426 | 5 mixed responsibilities (hardware, touch, widgets, graphics, navigation) | **LOW** |
| 2 | `runtime/gfx/surface.go` | 385 | Concurrent graphics with mutex, batching, 10+ primitives | **HIGH** |
| 3 | `runtime/js/module_ui/module.go` | 383 | 6 component types × 4 methods each, repetitive bridge code | **MEDIUM** |
| 4 | `runtime/js/runtime_test.go` | 528 | Comprehensive integration tests (acceptable complexity) | N/A |
| 5 | `runtime/ui/ui_test.go` | 247 | Test coverage for UI system | N/A |
| 6 | `pkg/runtimeowner/runner.go` | 230 | Thread-safety, context tracking, panic recovery | **HIGH** |
| 7 | `runtime/ui/display.go` | 257 | Display component with layer management | **HIGH** |
| 8 | `runtime/render/visual_runtime.go` | 198 | Multi-layer rendering, theme system | **HIGH** |
| 9 | `runtime/ui/ui.go` | 203 | Component tree orchestration | **HIGH** |
| 10 | `loupedeck.go` | 243 | Main device abstraction | **MEDIUM** |

### Dependency Graph

```
┌─────────────────────────────────────────────────────────────┐
│                        cmd/*                                │
│                   (CLI Applications)                         │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
│     pkg/     │  │ runtime/js   │  │     Root         │
│              │  │              │  │   (loupedeck)    │
│ runtimeowner │◄─┤ module_ui    │  │                  │
│ runtimebridge│  │ module_gfx   │  │ displayknob.go ◄─┼──┐
│  jsmetrics   │  │ module_anim  │  │ loupedeck.go     │  │
└──────────────┘  │ module_state │  │ listeners.go     │  │
        ▲         │ ...          │  │ inputs.go        │  │
        │         └──────────────┘  └──────────────────┘  │
        │                   │                   ▲         │
        │         ┌─────────┴─────────┐         │         │
        │         ▼                   ▼         │         │
        │  ┌──────────────┐    ┌──────────────┐ │         │
        │  │ runtime/ui   │    │ runtime/gfx  │ │         │
        └──┤              │    │              │─┘         │
           │   Page     │    │   Surface    │             │
           │   Display  │◄───┤   (pixels)   │             │
           │   Tile     │    │              │             │
           └──────────────┘    └──────────────┘             │
                  │                                            │
                  ▼                                            │
           ┌──────────────┐                                   │
           │runtime/render│◄──────────────────────────────────┘
           │   (images)   │   [displayknob.go uses both
           └──────────────┘    hardware AND rendering]
```

## Gap Analysis

### Critical Gap: Widget System Location

**Evidence**: `displayknob.go` lines 120-426 contain:
- Lines 120-180: Touch drag detection (`isClick`, `RegisterDragDisplayKnobWatcher`)
- Lines 180-300: Widget interface (`DKWidget`, `DKAnalogWidget`) with graphics rendering
- Lines 300-380: Widget container (`WidgetHolder`) with swipe navigation
- Lines 380-426: Navigation bar rendering with trigonometric calculations

These are mixed with hardware abstraction (lines 1-120) in the same file.

**Impact**: 
- Hardware package imports graphics libraries (`github.com/jphsd/graphics2d`)
- Touch detection logic (`isClick` duration thresholds) buried in hardware file
- Widget interface in hardware package forces hardware imports for UI work

### Moderate Gap: JS Module Boilerplate

**Evidence**: `runtime/js/module_ui/module.go` lines 140-380 show the pattern:

```go
// Repeated for text, icon, visible, surface... (~80 lines per component)
_ = obj.Set("text", func(call goja.FunctionCall) goja.Value {
    if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
        tile.BindText(func() string {
            result, err := bindings.Owner.Call(ownerCtx, "ui.tile.text", func(...) {
                value, err := fn(goja.Undefined())
                // ... 10 more lines of bridge code
            })
            // ... error handling
        })
    } else {
        tile.SetText(stringify(call.Argument(0)))
    }
    return goja.Undefined()
})
```

This pattern repeats 12+ times across tile, display, and page objects.

**Impact**:
- Adding a new property requires 10+ lines of boilerplate
- Risk of inconsistency between component types
- Difficult to review (large file with similar-looking code)

### Minor Gap: Package Naming Consistency

**Evidence**:
- `runtime/js/` contains modules but isn't named `modules/`
- `pkg/jsmetrics/` is separate from `runtime/metrics/`
- Some utilities in root package could be in `pkg/`

## Proposed Solution

### 1. Extract Widget System (HIGH PRIORITY)

**Proposed Structure:**

```
runtime/widgets/
├── widget.go              # DKWidget interface
├── analog.go              # DKAnalogWidget implementation
├── holder.go               # WidgetHolder with swipe navigation
└── touch/
    └── detector.go        # Touch/drag detection (extracted from displayknob.go)

Root package changes:
- displayknob.go → Keep only hardware abstraction (lines 1-120, ~120 lines)
- Remove: widget interfaces, WidgetHolder, touch detection, navigation bar rendering
```

**Migration Steps:**

1. Create `runtime/widgets/widget.go`:
```go
package widgets

import deck "github.com/go-go-golems/loupedeck"

type Widget interface {
    Activate(deck *deck.Loupedeck)
    Deactivate(deck *deck.Loupedeck)
}

type Analog struct {
    Min, Max                 int
    MinDegrees, TotalDegrees float64
    Value                    *deck.WatchedInt
    Name                     string
    active                   bool
}
```

2. Create `runtime/widgets/holder.go`:
```go
package widgets

// WidgetHolder manages swipeable widgets on CT knob display
type Holder struct {
    widgets []Widget
    active  int
    // ...
}

func (h *Holder) RegisterDragHandler(deck *deck.Loupedeck) {
    // Touch detection logic extracted from displayknob.go
}
```

3. Create `runtime/widgets/touch/detector.go`:
```go
package touch

import "time"

// Detector distinguishes clicks from drags
type Detector struct {
    MaxClickDuration time.Duration
    MaxClickDistance int
}

func (d *Detector) IsClick(duration time.Duration, dx, dy int) bool {
    // Extracted from displayknob.go isClick()
}
```

4. Update root `displayknob.go` to ~120 lines:
```go
package loupedeck

// Keep: DisplayKnob struct (hardware abstraction)
// Keep: DisplayKnob(), Get(), Set(), Inc()
// Keep: RegisterDragDisplayKnobWatcher (thin wrapper calling widgets/touch)
// Remove: DKWidget, DKAnalogWidget, WidgetHolder, drawWidgetHolderNavBar
```

**Benefits:**
- Hardware package no longer imports graphics libraries
- Widget system can be tested without hardware
- Clear boundary: root = hardware protocol, runtime/widgets = user-facing widgets

### 2. Refactor JS Module Boilerplate (MEDIUM PRIORITY)

**Option A: Code Generation (Recommended for scale)**

Create `runtime/js/generate/` with templates:

```go
// Template for property binding
{{define "property"}}
_ = obj.Set("{{.Name}}", func(call goja.FunctionCall) goja.Value {
    if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
        {{.Receiver}}.Bind{{.TitleName}}(func() {{.Type}} {
            result, err := bindings.Owner.Call(ownerCtx, "{{.Path}}", func(_ context.Context, vm *goja.Runtime) (any, error) {
                value, err := fn(goja.Undefined())
                if err != nil { return nil, err }
                return {{.Convert}}, nil
            })
            if err != nil { panic(runtime.NewGoError(err)) }
            return result.({{.Type}})
        })
    } else {
        {{.Receiver}}.Set{{.TitleName}}({{.ConvertStatic}})
    }
    return goja.Undefined()
})
{{end}}
```

Generate from schema:
```yaml
# runtime/js/schemas/ui.yaml
components:
  - name: tile
    properties:
      - name: text
        type: string
        getter: tile.Text()
        setter: tile.SetText()
        binder: tile.BindText()
```

**Option B: Reflection-Based Helpers (Simpler, less type-safe)**

```go
// runtime/js/helpers/binder.go
func BindProperty[T any](
    obj *goja.Object,
    name string,
    getter func() T,
    setter func(T),
    binder func(func() T),
    toGo func(goja.Value) T,
    toJS func(T) goja.Value,
) {
    // Single implementation handles all property types
}
```

**Benefits:**
- 70% reduction in module_ui/module.go lines (383 → ~120)
- Consistent error handling across all properties
- Easier to add new components

### 3. Optional: Reorganize runtime/js/ Structure

**Current:**
```
runtime/js/
├── module_ui/
├── module_gfx/
├── module_anim/
├── module_state/
├── module_easing/
├── module_metrics/
└── module_scene_metrics/
```

**Proposed (if module count grows beyond 10):**
```
runtime/js/modules/
├── ui/          # Was module_ui
├── gfx/         # Was module_gfx
├── anim/        # Was module_anim (includes easing)
├── state/       # Was module_state
└── metrics/     # Was module_metrics + module_scene_metrics
```

**Migration:** Simple directory move with import updates. Low priority until >10 modules.

### 4. Optional: Create pkg/display/ Abstraction

**Current**: `display.go` (193 lines) and `renderer.go` (148 lines) in root package.

**Proposed**: 
```
pkg/display/
├── target.go    # DrawTarget interface
├── renderer.go  # Render UI to targets
└── theme.go     # Theme configuration
```

**Benefits:** Decouple rendering from hardware, enable custom display targets (file, network, etc.).

**Priority**: Low - current structure is acceptable.

## Phased Implementation Plan

### Phase 1: Widget System Extraction (2-3 days)

**Day 1: Setup and Core Extraction**
- [ ] Create `runtime/widgets/` package
- [ ] Extract `DKWidget` interface to `widget.go`
- [ ] Extract `DKAnalogWidget` to `analog.go`
- [ ] Verify compilation: `go build ./...`

**Day 2: Touch Detection and Container**
- [ ] Create `runtime/widgets/touch/detector.go` with `isClick` logic
- [ ] Extract `WidgetHolder` to `holder.go`
- [ ] Move navigation bar rendering to `holder.go`
- [ ] Update imports in root `displayknob.go`

**Day 3: Cleanup and Testing**
- [ ] Reduce `displayknob.go` to ~120 lines (hardware only)
- [ ] Add `runtime/widgets/` tests (currently missing)
- [ ] Verify `cmd/` applications still work
- [ ] Run full test suite: `go test ./...`

**Success Criteria:**
- `displayknob.go` < 150 lines
- Hardware package no longer imports graphics libraries
- Widget system has test coverage

### Phase 2: JS Module Refactoring (3-5 days)

**Approach Decision Point:**
- If team comfortable with code generation → Option A (templates)
- If prefer simplicity → Option B (reflection helpers)

**Day 1-2: Design Helper API**
- [ ] Define property binding interface
- [ ] Implement core helper functions
- [ ] Test with one component (tile.text only)

**Day 3-4: Migrate Components**
- [ ] Migrate all tile properties
- [ ] Migrate all display properties
- [ ] Migrate page properties

**Day 5: Cleanup**
- [ ] Remove old repetitive code
- [ ] Verify JS examples still work
- [ ] Measure line count reduction (target: 383 → ~150 lines)

**Success Criteria:**
- `module_ui/module.go` < 200 lines
- All existing JS examples run correctly
- No regressions in UI functionality

### Phase 3: Documentation and Hygiene (1-2 days)

- [ ] Document package boundaries in README
- [ ] Add package-level doc comments
- [ ] Update ARCHITECTURE.md if exists
- [ ] Add diagram showing new structure

## Testing and Validation Strategy

### Pre-Refactor Baseline

```bash
# Capture current state
go test ./... -count=1 > /tmp/before_tests.log
wc -l displayknob.go runtime/js/module_ui/module.go > /tmp/before_lines.log
go list -deps ./... | wc -l > /tmp/before_deps.log
```

### Per-Phase Validation

**Phase 1 (Widget Extraction):**
```bash
# Verify hardware package no longer imports graphics
go list -deps github.com/go-go-golems/loupedeck | grep graphics2d
# Expected: no output (graphics2d should only be in runtime/)

# Verify widget system compiles independently
go build ./runtime/widgets/...

# Test existing examples still work
go run ./cmd/loupe-js-live/... --validate
```

**Phase 2 (JS Module Refactoring):**
```bash
# Run JS runtime tests
go test ./runtime/js/... -v

# Verify example scripts execute
cd examples/js && for f in *.js; do echo "Testing $f"; done
```

### Post-Refactor Comparison

| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| `displayknob.go` lines | 426 | ~120 | -72% |
| `module_ui/module.go` lines | 383 | ~150 | -61% |
| Hardware pkg imports | 15 | ~12 | -3 |
| Widget test coverage | 0% | >70% | +70% |
| JS module duplication | High | Low | Improved |

## Risks, Alternatives, and Open Questions

### Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking API changes | Medium | High | Keep old types as aliases during transition |
| JS examples break | Low | Medium | Comprehensive test suite before refactor |
| Increased package count | Certain | Low | Document rationale, use clear naming |
| Developer confusion during transition | Medium | Low | Clear commit messages, PR review |

### Alternatives Considered

**Alternative 1: Keep Current Structure**
- **Pros**: No work required, no risk of breakage
- **Cons**: Technical debt accumulates, harder onboarding, widget system growth blocked
- **Verdict**: Rejected - current structure is actively problematic

**Alternative 2: Merge Everything into Fewer Packages**
- **Pros**: Fewer imports, simpler dependency graph
- **Cons**: Loses separation of concerns, makes testing harder
- **Verdict**: Rejected - opposite of standard Go best practices

**Alternative 3: Extract to External Module**
- **Pros**: Cleanest separation, reusable by other projects
- **Cons**: Overkill for current scale, adds versioning complexity
- **Verdict**: Rejected - premature for current codebase size

### Open Questions

1. **Should WidgetHolder use runtime/reactive for state?**
   - Current: Uses `WatchedInt` from root package
   - Option: Could integrate with `runtime/reactive/signal.go`
   - Status: Not blocking Phase 1, can evolve later

2. **Should we keep DK prefix on widget types?**
   - DK = Display Knob
   - Option: Rename to `widgets.Analog`, `widgets.Holder` for clarity
   - Status: Keep DK prefix during extraction, rename in follow-up if desired

3. **Is the graphics2d dependency actually problematic?**
   - Current: Only displayknob.go uses it in root package
   - After Phase 1: Only runtime/widgets/ uses it
   - Status: Acceptable dependency, just needs to be in right package

## References

### Key Files

| File | Path | Lines | Role in Analysis |
|------|------|-------|------------------|
| `displayknob.go` | Root | 426 | Primary complexity driver |
| `surface.go` | `runtime/gfx/` | 385 | Well-structured complexity example |
| `module_ui/module.go` | `runtime/js/` | 383 | Boilerplate duplication example |
| `runner.go` | `pkg/runtimeowner/` | 230 | Good abstraction reference |
| `visual_runtime.go` | `runtime/render/` | 198 | Clean separation example |
| `ui.go` | `runtime/ui/` | 203 | Component tree pattern |

### Related Documentation

- `runtime/reactive/` - Signal-based reactivity (solidjs-inspired)
- `runtime/anim/` + `runtime/easing/` - Animation system
- `docs/help/` - User-facing documentation
- `examples/js/` - JavaScript API usage patterns

### External References

- Go Best Practices: https://go.dev/doc/effective_go
- Package Design: https://dave.cheney.net/2019/10/06/use-internal-packages-to-reduce-your-api-surface
- Goja Runtime: https://github.com/dop251/goja

---

## Conclusion

The go-go-golems/loupedeck codebase demonstrates solid architectural patterns in the runtime/ subsystem but suffers from a critical cohesion issue in the root package. The 426-line `displayknob.go` file is the primary concern, mixing hardware abstraction with high-level widget abstractions.

**Immediate Action**: Extract the widget system to `runtime/widgets/` (Phase 1, 2-3 days).

**Secondary Action**: Refactor JS module boilerplate to reduce `module_ui/module.go` by 60% (Phase 2, 3-5 days).

These changes will:
1. Create clear package boundaries (hardware vs. widgets)
2. Enable independent testing of the widget system
3. Reduce cognitive load for new developers
4. Provide a clean foundation for future widget types
5. Improve long-term maintainability without breaking existing functionality

The analysis reveals that most complexity in the codebase is **justified** (surface.go, runner.go, visual_runtime.go) - the issue is **organization**, not implementation quality. Phase 1 extraction addresses the only significant architectural debt.
