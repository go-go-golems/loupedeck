---
Title: Investigation Diary - jsverbs and jsdocex Analysis for Loupedeck
Ticket: LOUPE-JSVERBS
Status: active
Topics:
    - loupedeck
    - jsverbs
    - jsdoc
    - goja
    - investigation
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/scan.go
      Note: Tree-sitter scanning for __verb__, __section__, __package__ extraction (822 lines)
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/model.go
      Note: VerbSpec, Registry, SectionSpec data models with shared section support
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go
      Note: goja runtime invocation with source overlay injection and promise handling
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/command.go
      Note: Glazed command wrappers for JS verbs with output mode support
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsdoc/extract/extract.go
      Note: Tree-sitter extraction for __doc__, __example__, doc`...` patterns (700+ lines)
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsdoc/model/model.go
      Note: SymbolDoc, Example, Package documentation models
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsdoc/model/store.go
      Note: DocStore with ByPackage, BySymbol, ByExample, ByConcept indexes
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsdoc/server/server.go
      Note: HTTP server with SSE for documentation browsing
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/runtime.go
      Note: Loupedeck JS runtime with goja VM, event loop, and native modules
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/env/env.go
      Note: Environment bundling Reactive, UI, Host, Anim, Present, Metrics
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/run/command.go
      Note: Current run command with direct script loading via os.ReadFile
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/examples/js/01-hello.js
      Note: Basic example using ui.page() and ui.show() (22 lines)
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/examples/js/07-cyb-ito-prototype.js
      Note: Complex example with 423 lines of tile configuration
ExternalSources: []
Summary: Chronological investigation of jsverbs and jsdocex subsystems from go-go-goja, examining their architecture, integration patterns, and applicability to loupedeck's JS runtime.
LastUpdated: 2026-04-13T18:00:00-04:00
WhatFor: Document the step-by-step investigation process, what was learned, what patterns were discovered, and how they apply to the loupedeck integration.
WhenToUse: Reference when implementing the integration, debugging issues, or onboarding engineers to the jsverbs/jsdocex subsystems.
---

# Investigation Diary: jsverbs and jsdocex Analysis for Loupedeck

## Entry 1: Initial Exploration (2026-04-13 16:30)

### What I Did

Started by locating the relevant repositories:
```bash
find /home/manuel -type d -name "go-go-goja" 2>/dev/null | head -5
```

Found multiple worktrees:
- `/home/manuel/code/wesen/go-go-golems/go-go-goja` (main)
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja` (corporate)
- Several dated worktrees

Also located loupedeck:
```bash
find /home/manuel -type d -name "loupedeck*" 2>/dev/null | head -10
```

Key finding: `/home/manuel/code/wesen/corporate-headquarters/loupedeck` is the main worktree.

### What I Learned

The go-go-goja package structure:
```
pkg/
├── jsverbs/      # Command discovery from JS files
│   ├── scan.go   # Tree-sitter parsing (822 lines)
│   ├── model.go  # Data models
│   ├── command.go # Glazed integration
│   ├── runtime.go # goja invocation
│   └── binding.go # Parameter binding plans
├── jsdoc/        # Documentation extraction
│   ├── extract/  # Tree-sitter extraction
│   ├── model/    # Doc models and store
│   ├── server/   # HTTP doc browser
│   └── batch/    # Batch processing
```

### What Was Tricky

Multiple worktrees exist - had to identify the canonical one. The corporate-headquarters version has the full git history and is the active workspace.

## Entry 2: Deep Dive into jsverbs (2026-04-13 16:45)

### What I Did

Read the core jsverbs files:
1. `command.go` - Glazed command wrapper (11525 bytes)
2. `runtime.go` - Runtime invocation (7885 bytes)
3. `model.go` - Data models (7404 bytes)
4. `scan.go` - Tree-sitter scanning (continues at offset 200)

### What I Learned

**Key Pattern: Verb Registration via Sentinel Functions**

JS scripts use special functions that are no-ops at runtime but parsed at scan-time:

```javascript
__package__({ name: "test", title: "Test Package" });
__section__("filters", { fields: { ... } });
__verb__(myFunction, { name: "my-cmd", useSections: ["filters"] });
```

**Architecture Layers:**

1. **Scan Layer** (`scan.go`):
   - Tree-sitter JavaScript parser
   - Extracts 4 sentinel types: `__package__`, `__section__`, `__verb__`, and public functions
   - Builds `Registry` with `[]*FileSpec` and verb index

2. **Model Layer** (`model.go`):
   - `Registry` holds all scanned files and shared sections
   - `FileSpec` per file: functions, sections, package metadata
   - `VerbSpec` per command: name, parents, sections, binding plan
   - `SectionSpec` for parameter groups

3. **Command Layer** (`command.go`):
   - `Command` implements `cmds.GlazeCommand`
   - `WriterCommand` implements `cmds.WriterCommand`
   - Builds Glazed field definitions from VerbSpec
   - Output modes: `glaze` (tabular) or `text`

4. **Runtime Layer** (`runtime.go`):
   - Creates goja runtime with `engine.NewBuilder()`
   - Injects source overlay: `__glazedVerbRegistry` capture
   - Marshals arguments via binding plan
   - Handles promises via polling (5ms sleep loop)

**Parameter Binding System:**

Four binding modes discovered in `runtime.go`:
- `BindingModeAll`: Pass all values as map
- `BindingModeContext`: Pass rich context with verb info, sections
- `BindingModeSection`: Pass specific section
- `BindingModePositional`: Pass individual fields

This is sophisticated - allows flexible argument passing patterns.

**Source Overlay Injection:**

Critical technique in `runtime.go` lines 103-134:
```go
func (r *Registry) injectOverlay(moduleKey string, file *FileSpec, source string) string {
    // Adds prelude with __glazedVerbRegistry initialization
    // Adds suffix capturing all exported functions
    // Result: module can be required and functions retrieved
}
```

This is how jsverbs gets function references without polluting the global scope.

### What Was Tricky

The `scan.go` file is 822 lines - had to read it in chunks. The tree-sitter AST traversal is complex with nested recursions for:
- `expression_statement` → `call_expression`
- `function_declaration`
- `lexical_declaration` / `variable_declaration`
- `export_statement`

Understanding when a function is captured as a verb vs. just a public function required reading the `finalizeVerbs()` method carefully.

## Entry 3: Understanding jsdocex (2026-04-13 17:00)

### What I Did

Read jsdoc package:
1. `extract/extract.go` - Tree-sitter extraction (700+ lines)
2. `model/model.go` - Documentation models
3. `model/store.go` - DocStore with indexes
4. `server/server.go` - HTTP API and SSE

### What I Learned

**Documentation Extraction Patterns:**

Four sentinel patterns extracted:
1. `__package__({...})` - Package metadata
2. `__doc__("name", {...})` or `__doc__({name: "...", ...})` - Symbol documentation
3. `__example__({...})` - Executable examples
4. `doc\`...\`` - Template literal with frontmatter and prose

**Template Literal Parsing:**

The `handleDocTemplate()` function in `extract.go` handles:
```javascript
doc`
---
symbol: myFunction
---
# Documentation
Prose content here...
`;
```

This attaches long-form Markdown to symbols after the `__doc__` metadata.

**DocStore Indexing:**

The store maintains four indexes:
- `ByPackage`: name → Package
- `BySymbol`: name → SymbolDoc
- `ByExample`: id → Example
- `ByConcept`: concept → []symbol names

This enables fast lookups for the doc browser.

**Server Capabilities:**

The server provides:
- `/api/store` - Full doc store dump
- `/api/package/<name>` - Package details
- `/api/symbol/<name>` - Symbol with examples
- `/api/example/<id>` - Example details
- `/api/search?q=...` - Full-text search
- `/events` - SSE for live reload

This is a complete documentation browsing system.

**JS-to-JSON Conversion:**

The `convertJSToJSON()` function (350+ lines) converts JS object literals to valid JSON:
- Handles single quotes → double quotes
- Handles unquoted keys
- Strips comments (// and /* */)
- Handles trailing commas
- Supports template literals as strings

This is necessary because the metadata is JS syntax, not JSON.

### What Was Tricky

The extraction handles many edge cases:
- `__doc__("name", {...})` vs `__doc__({name: "...", ...})`
- Finding `doc\`\`` calls inside `call_expression` nodes
- Tree-sitter AST structure varies between JS constructs

Had to read the `handleCallExpression()` method multiple times to understand the dispatch logic.

## Entry 4: Loupedeck Runtime Analysis (2026-04-13 17:15)

### What I Did

Read loupedeck's JS runtime:
1. `runtime/js/runtime.go` - Main runtime
2. `runtime/js/env/env.go` - Environment
3. `cmd/loupedeck/cmds/run/command.go` - Run command
4. Examples in `examples/js/`

### What I Learned

**Loupedeck Runtime Structure:**

```go
type Runtime struct {
    VM    *goja.Runtime
    Loop  *eventloop.EventLoop
    Owner runtimeowner.Runner
    Env   *envpkg.Environment
}
```

Key differences from jsverbs:
- Uses `eventloop.EventLoop` for async (not engine.Runtime)
- Uses `runtimeowner.Runner` for thread-safe execution
- Has `require.Registry` for module loading

**Environment Composition:**

```go
type Environment struct {
    Reactive *reactive.Runtime  // Signals/effects
    UI       *ui.UI            // Tile/page management
    Host     *host.Runtime     // Device connection
    Anim     *anim.Runtime     // Animation system
    Present  *present.Runtime  // Rendering
    Metrics  *metrics.Collector // Performance metrics
}
```

This is a sophisticated runtime with many subsystems.

**Current Script Loading:**

In `cmd/loupedeck/cmds/run/command.go`:
```go
script, err := os.ReadFile(opts.ScriptPath)
rt := jsruntime.NewRuntime(env)
if _, err := rt.RunString(rt.Context(), string(script)); err != nil {
    return fmt.Errorf("run script: %w", err)
}
```

Very simple - just reads file and runs it. No metadata extraction.

**Module System:**

Native modules registered:
- `loupedeck/ui` - Page and tile configuration
- `loupedeck/state` - Signal-based reactive state
- `loupedeck/anim` - Animations
- `loupedeck/gfx` - Graphics primitives
- `loupedeck/easing` - Easing functions
- `loupedeck/present` - Presentation logic
- `loupedeck/metrics` - Performance tracking

Example script pattern:
```javascript
const ui = require("loupedeck/ui");
const state = require("loupedeck/state");

ui.page("home", page => {
    page.tile(0, 0, tile => tile.text("HELLO"));
});

ui.show("home");
```

**Example Complexity:**

Line counts show evolution:
- `01-hello.js`: 22 lines (basic)
- `03-knob-meter.js`: 32 lines (state + events)
- `06-page-switcher.js`: 39 lines (pages + buttons)
- `07-cyb-ito-prototype.js`: 423 lines (complex dashboard)
- `11-cyb-os-tiles.js`: 596 lines (most recent)

The scripts are getting more complex - need for structure/documentation is growing.

### What Was Tricky

The runtime initialization order matters:
1. Create VM
2. Create event loop
3. Register native modules
4. Enable registry on VM
5. Create owner/runner
6. Set up bridge bindings

Need to inject verb registry at the right point (after step 3, before step 6).

## Entry 5: Integration Point Analysis (2026-04-13 17:30)

### What I Did

Compared jsverbs and loupedeck runtimes to identify integration strategies.

### What I Learned

**Runtime Difference Summary:**

| Aspect | jsverbs | loupedeck |
|--------|---------|-----------|
| VM | goja.Runtime | goja.Runtime |
| Async | engine.Runtime (wrapper) | eventloop.EventLoop |
| Execution | Direct Call() | runtimeowner.Runner.Call() |
| Modules | require.Registry | require.Registry |
| Source Loading | Custom loader | Filesystem |

**Integration Strategy:**

Cannot use jsverbs runtime directly because:
1. Loupedeck needs its specific native modules
2. Event loop integration differs
3. Thread ownership model differs

**Solution: Adapter Pattern**

Create `RuntimeAdapter` that:
1. Uses loupedeck's existing `jsruntime.Runtime`
2. Injects verb registry overlay into loupedeck's `require.Registry`
3. Marshals arguments same as jsverbs
4. Uses `runtimeowner.Runner.Call()` for execution

**Key Implementation:**

Port the source overlay injection from jsverbs:
```go
// From jsverbs runtime.go
func injectOverlay(source, moduleKey string, functions []string) string {
    prelude := strings.Join([]string{
        `globalThis.__glazedVerbRegistry = globalThis.__glazedVerbRegistry || {};`,
        `globalThis.__package__ = globalThis.__package__ || function() {};`,
        `globalThis.__section__ = globalThis.__section__ || function() {};`,
        `globalThis.__verb__ = globalThis.__verb__ || function() {};`,
        `globalThis.doc = globalThis.doc || function() { return ""; };`,
        "",
    }, "\n")
    
    suffix := // ... function capture
    
    return injectPrelude(source, prelude) + suffix
}
```

This makes the sentinel functions no-ops at runtime but available for scanning.

**Promise Handling:**

jsverbs uses polling (5ms sleep loop) for promises. Loupedeck should use the event loop:
```go
// Better approach for loupedeck
loop.RunOnLoop(func() {
    // Check promise state
    // Resolve/reject via channel
})
```

### What Was Tricky

Understanding how to inject the overlay into loupedeck's require.Registry required tracing:
1. How jsverbs does it: custom loader passed to `require.WithLoader()`
2. How loupedeck sets up registry: `registry.Enable(vm)`

The injection needs to happen at the require level, wrapping the existing filesystem loader.

## Entry 6: Doc Browser Integration (2026-04-13 17:40)

### What I Did

Examined `jsdoc/server/server.go` for documentation serving capabilities.

### What I Learned

**Server Architecture:**

```go
type Server struct {
    store *model.DocStore
    dir   string
    host  string
    port  int
    clients map[chan string]struct{}  // SSE clients
}
```

**API Endpoints:**

| Endpoint | Purpose |
|----------|---------|
| GET /api/store | Full doc store |
| GET /api/package/{name} | Package details |
| GET /api/symbol/{name} | Symbol with examples |
| GET /api/example/{id} | Example details |
| GET /api/search?q=... | Search |
| GET /api/batch/extract | Batch extraction |
| GET /api/batch/export | Export (multiple formats) |
| GET /events | SSE stream |
| GET / | Single-page app UI |

**Live Reload:**

Uses filesystem watcher (`watch.Watcher`) to:
1. Detect file changes
2. Re-parse affected files
3. Update DocStore
4. Broadcast SSE "reload" event

**Single-Page App:**

The `uiHTML` constant (not examined) provides a browser UI for:
- Browsing packages
- Viewing symbols with params/returns/prose
- Running examples
- Searching

### What Was Tricky

The batch endpoints suggest the server can handle multiple files. For loupedeck, we might want:
1. Single-script mode (one file)
2. Directory mode (all .js files)
3. Library mode (npm-style package)

Need to decide which mode to support initially.

## Entry 7: Skill File Review (2026-04-13 17:45)

### What I Did

Read the ticket-research-docmgr-remarkable skill file to understand documentation requirements.

### What I Learned

**Workflow Steps:**
1. Initialize ticket workspace
2. Gather evidence before writing conclusions
3. Write primary analysis document (this is where the design doc goes)
4. Maintain chronological investigation diary (this file)
5. Update ticket bookkeeping (relate files, changelog)
6. Validate doc quality with doctor
7. Upload to reMarkable

**Documentation Quality Requirements:**

- Executive summary for quick scanning
- Evidence-based claims with file anchors
- Concrete API references and pseudocode
- Phased implementation plan with file-level guidance
- Risk assessment and alternatives

### What Was Tricky

The skill emphasizes "file-backed evidence" - every claim must reference concrete files. This diary format requires careful tracking of what was examined.

## What Worked

1. **Repository Discovery:** Using `find` with multiple patterns found all relevant repos
2. **Structured Reading:** Reading files in dependency order (model → scan → runtime → command) helped build understanding
3. **Line-Anchored Notes:** Recording specific line numbers (e.g., `scan.go:103-134`) enables precise references
4. **Comparison Tables:** Comparing jsverbs vs loupedeck runtime structures clarified integration points

## What Didn't Work

1. **Multiple Worktrees:** Had to determine which go-go-goja was canonical - initially confusing
2. **Large Files:** `scan.go` at 822 lines required multiple read calls with offsets
3. **Missing Context:** Some Go code references unexported symbols - had to infer from usage

## What Was Tricky to Build

Not applicable - this is analysis phase, no code written yet.

## Code Review Instructions

When reviewing this analysis:

1. **Verify File References:**
   ```bash
   ls -la /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/
   wc -l /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/*.go
   ```

2. **Check Excerpts:**
   ```bash
   sed -n '103,134p' /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go
   ```

3. **Verify Examples:**
   ```bash
   wc -l /home/manuel/code/wesen/corporate-headquarters/loupedeck/examples/js/*.js
   ```

## Summary of Key Findings

### jsverbs Architecture
- 4-layer design: scan → model → command → runtime
- Sentinel functions (`__verb__`, `__section__`, `__package__`) for metadata
- Source overlay injection for function capture
- Four parameter binding modes (all, context, section, positional)
- Glazed integration for CLI generation

### jsdocex Architecture
- Tree-sitter extraction for 4 sentinel patterns
- DocStore with 4 indexes (ByPackage, BySymbol, ByExample, ByConcept)
- Template literal support for long-form docs
- HTTP server with SSE for live reload
- Batch processing for multiple files

### Loupedeck Runtime
- goja + eventloop + runtimeowner.Runner
- 7 native modules (ui, state, anim, gfx, easing, present, metrics)
- Simple script loading (ReadFile + RunString)
- Complex environment with Reactive, UI, Host subsystems

### Integration Strategy
- Cannot use jsverbs runtime directly (different async model)
- Need adapter layer using loupedeck's existing runtime
- Port source overlay injection to loupedeck's require.Registry
- Use runtimeowner.Runner.Call() for thread-safe execution
- Add ScriptRegistry for multi-script management

## Document History

- **2026-04-13 16:30:** Started exploration, located repositories
- **2026-04-13 16:45:** Deep dive into jsverbs (scan, model, command, runtime)
- **2026-04-13 17:00:** Examined jsdocex (extract, model, store, server)
- **2026-04-13 17:15:** Analyzed loupedeck runtime (runtime, env, run command)
- **2026-04-13 17:30:** Compared runtimes, designed adapter pattern
- **2026-04-13 17:40:** Reviewed doc server capabilities
- **2026-04-13 17:45:** Read skill documentation, understood deliverable requirements
- **2026-04-13 18:00:** Completed diary, ready for document generation
