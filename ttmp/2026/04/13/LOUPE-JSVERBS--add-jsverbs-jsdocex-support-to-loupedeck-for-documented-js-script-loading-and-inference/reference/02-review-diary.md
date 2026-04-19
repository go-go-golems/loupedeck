---
Title: Review Diary - LOUPE-JSVERBS Design Review
Ticket: LOUPE-JSVERBS
Status: active
Topics:
    - loupedeck
    - jsverbs
    - jsdoc
    - review
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/scan.go
      Note: Tree-sitter scanner - confirmed engine-free, only runtime.go imports engine
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go
      Note: The ONLY jsverbs file that imports engine - critical for dependency strategy
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/testdata/jsverbs/basics.js
      Note: Actual fixture showing __verb__("greet", {...}) STRING syntax, NOT identifier
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/go.mod
      Note: Does NOT include go-go-goja as dependency - major gap in design
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/go.mod
      Note: Go 1.26.1 vs loupedeck Go 1.25.0; different goja versions
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/pkg/runtimeowner/runner.go
      Note: Duplicated from go-go-goja but diverged
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/runtimeowner/runner.go
      Note: Original; loupedeck copy has OwnerContext() method removed
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/runtime.go
      Note: engine.Runtime struct - nearly identical to loupedeck's js.Runtime
    - Path: /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/binding.go
      Note: Binding plan logic - missed by design doc, critical for adapter
ExternalSources: []
Summary: Critical review of the first-pass design doc for LOUPE-JSVERBS. Found 3 blocking issues (dependency strategy, factual API error, goja version mismatch), 3 structural issues (overengineering, underspecified adapter, duplicated packages), and 4 minor issues. Recommends revise-then-proceed.
LastUpdated: 2026-04-13T19:30:00-04:00
WhatFor: Document the design review findings so the next implementer knows exactly what to fix before coding.
WhenToUse: Read before starting implementation of LOUPE-JSVERBS to avoid the identified pitfalls.
---

# Review Diary: LOUPE-JSVERBS Design Review

## Goal

Review the first-pass design document for LOUPE-JSVERBS (jsverbs + jsdocex integration into loupedeck), cross-check every claim against actual source code, and produce a verdict: go ahead, revise, or redo.

## Step 1: Cross-Reference Design Claims Against Source Code

The design doc describes the jsverbs/jsdocex/loupedeck architecture and proposes an integration. I systematically verified each claim by reading the actual source files.

### Prompt Context

**User prompt (verbatim):** "We are taking over for your little brother who did the first design pass. Review their work and if it's ok to go ahead with."

**Assistant interpretation:** Critically review the LOUPE-JSVERBS design doc and diary. Verify factual accuracy, assess architectural soundness, identify missing analysis, and render a go/no-go verdict.

**Inferred user intent:** Don't waste time on a flawed design. Catch problems before coding starts.

### What I did

1. Re-read the full design doc (~25KB) end-to-end
2. Re-read the investigation diary (~20KB)
3. Cross-checked every code excerpt and API claim against actual source files:
   - `go-go-goja/pkg/jsverbs/{scan,model,command,runtime,binding}.go`
   - `go-go-goja/pkg/jsdoc/{extract,model,server,batch}/*.go`
   - `loupedeck/runtime/js/{runtime.go,env/env.go}`
   - `loupedeck/cmd/loupedeck/cmds/run/command.go`
4. Verified import chains, dependency graphs, and API signatures
5. Ran actual test fixtures to confirm the `__verb__` syntax
6. Compared `go.mod` files for version compatibility

### What worked in the design

- **Architecture description is accurate.** The 4-layer jsverbs breakdown (scan → model → command → runtime) is correct. The model types are described faithfully.
- **Gap analysis is solid.** The five gaps (no scan-time extraction, no type inference, no docs, no command registration, runtime mismatch) are real and correctly ordered by dependency.
- **The "adapter, don't use jsverbs runtime directly" instinct is right.** The two runtimes are genuinely different (loupedeck uses its own `runtimeowner` with event loop; jsverbs uses `engine.Runtime`). A bridge layer is the correct approach.
- **The diary is well-structured.** The chronological entries trace a logical investigation path.

---

## Verdict: REVISE THEN PROCEED

The design has the right general shape but contains **3 blocking issues** that must be fixed before any code is written, plus several structural issues that will cause implementation pain if not addressed.

---

## Blocking Issues

### Blocker 1: Factual Error — `__verb__()` Syntax in Example Code

**The design's "Desired State" example is wrong.**

Design doc shows:
```javascript
__verb__(configureDashboard, {    // ← IDENTIFIER (wrong)
  name: "configure",
  ...
});
```

Actual jsverbs fixture `testdata/jsverbs/basics.js` uses:
```javascript
__verb__("greet", {               // ← STRING (correct)
  short: "Greet one person",
  ...
});
```

**Evidence:** In `scan.go:handleVerb()` → `namedObjectArgs()`, the first argument is parsed via `parseLiteralNode()`. An `identifier` node hits the default case and returns an error:

```go
// scan.go parseLiteralNode()
default:
    return nil, fmt.Errorf("unsupported metadata literal %q", node.Kind())
```

The function name is passed as a **string literal**, not an identifier reference. The design's example would fail at scan-time.

**Fix:** Change all `__verb__(funcName, {...})` examples in the design to `__verb__("funcName", {...})`.

**Severity:** This seems minor but it means the author didn't verify the core API against the test fixtures. It undermines confidence in other claims.

---

### Blocker 2: Dependency Strategy Entirely Missing

The design proposes importing `go-go-goja/pkg/jsverbs` and `go-go-goja/pkg/jsdoc` into loupedeck. **This is a major new dependency** and the design doesn't discuss it at all.

**Facts from code:**

| Aspect | Loupedeck | go-go-goja |
|--------|-----------|------------|
| Go version | `1.25.0` | `1.26.1` |
| goja | `v0.0.0-20260311135729` | `v0.0.0-20251103141225` (4 months older) |
| goja_nodejs | `v0.0.0-20260212` | `v0.0.0-20250409` (10 months older) |
| tree-sitter | NOT present | Required by jsverbs + jsdoc |
| go-go-goja | NOT a dependency | — |

**Impact:**
- Adding `go-go-goja` as a dependency pulls in `tree-sitter/go-tree-sitter` and `tree-sitter-javascript` (new transitive deps)
- Goja version mismatch may cause compilation or behavioral differences
- Go version mismatch: `go-go-goja` requires 1.26.1, loupedeck is on 1.25.0
- The `runtimeowner` and `runtimebridge` packages are **already duplicated** between the two repos with minor divergences — adding more shared code without a strategy compounds this

**Possible strategies (all unaddressed):**

1. **Add go-go-goja as a Go module dependency** (standard `go get`). Requires version alignment.
2. **Copy the scan/model/command/binding code** into loupedeck's `pkg/jsverbs/` (avoid the dependency entirely). The scan+model+command+binding layers have NO dependency on `engine` or go-go-goja-specific packages — they only import `glazed` and `tree-sitter`.
3. **Extract jsverbs-core into a shared library** package that both repos depend on.
4. **Use Go workspace** (`go.work`) to develop across both repos.

**The design must pick one.**

---

### Blocker 3: `runtime.go` Imports `engine` — Import Strategy Not Addressed

The design proposes using `pkg/jsverbs` as a package. But if loupedeck does `import "go-go-goja/pkg/jsverbs"`, it transitively imports `runtime.go`, which imports `engine`, which imports `modules/database`, `modules/exec`, `modules/fs`, `modules/timer` via blank imports.

**Evidence:**
```go
// runtime.go:15
"github.com/go-go-golems/go-go-goja/engine"

// engine/runtime.go imports
_ "github.com/go-go-golems/go-go-goja/modules/database"
_ "github.com/go-go-golems/go-go-goja/modules/exec"
_ "github.com/go-go-golems/go-go-goja/modules/fs"
_ "github.com/go-go-golems/go-go-goja/modules/timer"
```

Loupedeck doesn't need any of these. The design's adapter approach avoids *using* runtime.go, but the Go compiler doesn't care — it's in the same package, so it gets compiled.

**The design needs to either:**
- Propose copying scan/model/command/binding (not runtime.go) into loupedeck
- Propose splitting jsverbs into `jsverbs/core` and `jsverbs/runtime` packages upstream
- Accept the transitive dependency and explain why it's acceptable

---

## Structural Issues (Non-Blocking but Important)

### Issue 4: `ScriptRegistry` Is Overengineered for Phase 1

The design proposes a `ScriptRegistry` with multi-script loading, conflict resolution, and command flattening. But loupedeck's current `run` command loads **one script at a time**. Starting with a `ScriptLoader` that handles single scripts would be simpler and sufficient.

**Recommendation:** Replace `ScriptRegistry` with a simpler `ScriptLoader` that scans one file and creates commands. Add multi-script management later.

### Issue 5: Phase 3 (RuntimeAdapter) Is Critically Underspecified

Phase 3 is the hardest part — bridging jsverbs execution into loupedeck's runtime — but gets the least detail. The key challenges aren't addressed:

1. **How to inject the overlay into loupedeck's `require.Registry`:** The current registry uses filesystem loading. jsverbs uses a custom `require.WithLoader()`. The design says "use the existing registry" and "inject the overlay" but doesn't explain how to do both simultaneously.

2. **Promise resolution:** jsverbs uses polling (5ms sleep loop). The design says "use event loop instead" but doesn't show how. The event loop's `RunOnLoop` runs callbacks on the loop goroutine, not the caller's goroutine — synchronization with the Owner.Call pattern is non-trivial.

3. **Argument marshaling:** The design references `buildArguments()` from jsverbs but doesn't discuss that it depends on `buildVerbBindingPlan()` from `binding.go`, which is a complex 150-line function. This must be reused as-is.

### Issue 6: Duplicated Packages Not Addressed

Both repos have `pkg/runtimeowner/` and `pkg/runtimebridge/` with slight divergences. Adding more shared code (jsverbs, jsdoc) without addressing this creates maintenance burden.

### Issue 7: `binding.go` Missed Entirely

The design never mentions `binding.go` (the 150-line binding plan resolution logic). This is a critical file — it determines how parameters map to JS function arguments. Any adapter must reuse it.

---

## Minor Issues

- **The doc server (Phase 5) is premature.** Nice-to-have but not core to the jsverbs integration. Should be explicitly marked optional/deferrable.
- **The "alternatives" section misses the copy option.** Copying scan/model/command/binding into loupedeck (strategy 2 above) avoids all dependency issues and is the simplest path.
- **The API design confuses jsdoc and jsverbs doc templates.** Both packages handle `doc\`...\`` but differently. The design should clarify which to use when.
- **Line count estimates are guesses.** "250 lines for adapter" is probably 400+ given the binding plan complexity.

---

## What Should Be Done

### Before writing any code:

1. **Fix the `__verb__` syntax error** in the design doc's examples (use strings, not identifiers)
2. **Choose a dependency strategy.** My recommendation: copy `scan.go`, `model.go`, `command.go`, `binding.go` into `loupedeck/pkg/jsverbs/`. This avoids all transitive dependency issues and the `engine` import problem. These four files are self-contained — they only depend on `glazed` (already in loupedeck) and `tree-sitter` (new but small and justified).
3. **Remove `ScriptRegistry`.** Replace with simpler single-script `ScriptLoader`.
4. **Expand Phase 3** with concrete code for: overlay injection into `require.Registry`, promise resolution via event loop, and binding plan reuse.
5. **Add `binding.go` to the architecture section.** It's as important as scan/model/command.
6. **Decide on the jsdoc integration separately.** jsdoc extraction is orthogonal to jsverbs command execution. Treat it as a separate, later phase.

### Revised phase order:

1. **Copy + adapt scanning** (scan.go, model.go, binding.go into loupedeck/pkg/jsverbs/)
2. **Single-script command loading** (ScriptLoader, not ScriptRegistry)
3. **Runtime adapter** (source overlay + argument marshaling + promise handling)
4. **CLI integration** (extend `run` command with --verb flag)
5. *(deferred)* jsdoc extraction
6. *(deferred)* doc server
7. *(deferred)* multi-script registry

---

## Step 2: Direction Change — Standardize on go-go-goja Runtime Instead of Copying Internals

After the initial review, the user explicitly chose the opposite trade-off: do not copy jsverbs internals into loupedeck, and do not keep the local runtime stack. Instead, import `go-go-goja` directly and replace the copied runtime infrastructure.

That decision changes the correct target architecture materially. The earlier review was still useful because it identified the sharp edges, but the updated plan now treats those edges as upstream integration work rather than reasons to avoid the dependency.

### Prompt Context

**User prompt (verbatim):** "No, let's actually switch to importing go-go-goja and also use it for the runtime, removing the one we copied over. Let's update the analysis/design/implementation plan with a better conceived and correct plan."

**Assistant interpretation:** Replace the previous dependency-avoidance recommendation with a revised design that embraces go-go-goja as the canonical runtime stack and updates the ticket docs accordingly.

**Inferred user intent:** Use this ticket to converge the runtime architecture, not merely to bolt jsverbs onto the existing loupedeck runtime.

### What I did

1. Read `go-go-goja/engine/{factory,runtime,module_roots}.go` to understand the real runtime composition API.
2. Verified that `engine.FactoryBuilder` already supports exactly the hooks loupedeck needs:
   - runtime module registrars,
   - runtime initializers,
   - require options,
   - module-root derivation from script path.
3. Re-read loupedeck native modules and confirmed they only need:
   - owner/context/loop from runtime bridge,
   - environment lookup by VM,
   - metrics lookup by VM.
4. Confirmed that upstream `runtimebridge` is intentionally generic and does **not** carry loupedeck-specific values.
5. Replaced the design doc with a new runtime-convergence-first plan.

### What worked

- The `engine` package turned out to be a better fit than the first pass assumed.
- `RuntimeModuleRegistrar` is the correct abstraction for loupedeck's module registration + environment seeding.
- Preserving `env.Lookup(vm)` while changing its implementation avoids churn in native modules.

### What didn't work

- The original idea of relying on jsverbs' default `Commands()` execution path still does not work for loupedeck, even after deciding to import go-go-goja directly. It still creates and closes a runtime per invocation.

### What I learned

- The real architectural mismatch is not “loupedeck runtime vs go-go-goja runtime.” The real mismatch is **long-lived host session vs ephemeral jsverbs invocation lifecycle**.
- Therefore, the corrected plan must include upstream jsverbs APIs for host-owned runtime invocation.

### What was tricky to build

The subtle part was realizing that “use go-go-goja for runtime” still does **not** mean “use jsverbs runtime.go as-is.” `pkg/jsverbs/runtime.go` is built for short-lived CLI commands, not persistent scene sessions. The revised design had to separate:

- using go-go-goja `engine.Runtime` as the canonical runtime,
- from using jsverbs' current default per-call runtime lifecycle.

### What warrants a second pair of eyes

- The proposed upstream jsverbs API shape (`CommandDescriptionForVerb`, `RequireLoader`, `InvokeInRuntime`, optionally `CommandsWithInvoker`) should be reviewed before implementation because it affects package ergonomics beyond loupedeck.
- The exact dependency/toolchain alignment strategy in `go.mod` still needs careful execution.

### What should be done in the future

1. Implement the runtime convergence phases first.
2. Upstream the jsverbs host-runtime APIs.
3. Only then wire loupedeck `run --verb` and jsdoc surfaces.

### Code review instructions

Verify the new direction with:

```bash
# Engine composition API
sed -n '1,220p' /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/factory.go

# Runtime-scoped registration hook
sed -n '1,120p' /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/runtime_modules.go

# Script-root require helper
sed -n '1,200p' /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/module_roots.go

# Current ephemeral jsverbs invocation path
sed -n '1,140p' /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go

# Loupedeck env lookup dependency on local bridge
sed -n '1,120p' /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/env/env.go
```

## Code Review Instructions

When verifying this review:

```bash
# Verify __verb__ syntax (string, not identifier)
cat /home/manuel/code/wesen/go-go-golems/go-go-goja/testdata/jsverbs/basics.js | grep "__verb__"

# Verify only runtime.go imports engine
grep -l "engine" /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/*.go

# Verify goja version mismatch
grep "goja " /home/manuel/code/wesen/corporate-headquarters/loupedeck/go.mod
grep "goja " /home/manuel/code/wesen/go-go-golems/go-go-goja/go.mod

# Verify duplicated packages
diff -rq /home/manuel/code/wesen/corporate-headquarters/loupedeck/pkg/runtimeowner/ /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/runtimeowner/

# Verify engine's transitive imports
head -20 /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/runtime.go
```
