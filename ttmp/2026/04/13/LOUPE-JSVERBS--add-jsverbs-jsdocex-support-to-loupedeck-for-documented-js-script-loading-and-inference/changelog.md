# Changelog

## 2026-04-13

- Initial workspace created


## 2026-04-13

Completed architecture analysis of jsverbs and jsdocex subsystems from go-go-goja. Examined scan layer, model layer, command layer, runtime layer, and documentation extraction. Analyzed loupedeck runtime for integration points.

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsdoc/extract/extract.go — 700-line documentation extractor
- /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/scan.go — 822-line tree-sitter scanner


## 2026-04-13

Design review completed: 3 blocking issues found (wrong __verb__ syntax, missing dependency strategy, engine transitive import), 3 structural issues (overengineered ScriptRegistry, underspecified adapter, duplicated packages). Verdict: REVISE THEN PROCEED.

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go — Only file importing engine - critical for dependency strategy
- /home/manuel/code/wesen/go-go-golems/go-go-goja/testdata/jsverbs/basics.js — Proof that __verb__ uses string args


## 2026-04-13

Replaced the first-pass design with a runtime-convergence-first plan: standardize on go-go-goja engine/runtime packages, remove loupedeck-local runtime copies, preserve env lookup via a loupedeck-specific VM bridge, and require upstream jsverbs host-runtime APIs for long-lived scene invocation.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/runtime.go — Local runtime wrapper now targeted for removal after migration
- /home/manuel/code/wesen/go-go-golems/go-go-goja/engine/factory.go — Canonical runtime composition API adopted by the revised plan
- /home/manuel/code/wesen/go-go-golems/go-go-goja/pkg/jsverbs/runtime.go — Current ephemeral invocation path that must gain host-runtime hooks


## 2026-04-13

Replaced placeholder tasks with a concrete unified implementation checklist across go-go-goja and loupedeck, organized by runtime convergence, jsverbs host-runtime APIs, loupedeck integration, jsdoc extraction, testing, and cleanup.

### Related Files

- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/ttmp/2026/04/13/LOUPE-JSVERBS--add-jsverbs-jsdocex-support-to-loupedeck-for-documented-js-script-loading-and-inference/tasks.md — Concrete task checklist for the shared workspace implementation plan


## 2026-04-14

Completed Phase 0 and Phase 1 implementation: added direct go-go-goja workspace dependency, aligned loupedeck to Go 1.26.1, upstreamed OwnerContext compatibility, replaced the local runtime stack with engine-based runtime registration, renamed Environment to LoupeDeckEnvironment, and removed copied runtimebridge/runtimeowner packages.

### Related Files

- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/runtimeowner/runner.go — Upstream OwnerContext compatibility patch for loupedeck modules
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/go.mod — Direct workspace dependency and toolchain alignment
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/runtime/js/env/env.go — LoupeDeckEnvironment rename and env lookup migration
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/runtime/js/registrar.go — Engine registrar for loupedeck runtime composition


## 2026-04-14

Completed the full ticket: added upstream jsverbs host-runtime APIs, integrated verb-aware scene execution and jsdoc extraction into loupedeck, added verbs/doc commands, shipped an annotated reference scene, updated tests, and refreshed help/docs to match the new runtime model.

### Related Files

- /home/manuel/workspaces/2026-04-13/js-loupedeck/go-go-goja/pkg/jsverbs/runtime.go — Exported InvokeInRuntime and RequireLoader for host-owned runtimes
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/doc/command.go — Doc extraction CLI
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/run/command.go — Run command now supports verb-aware scene bootstrapping
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/cmd/loupedeck/cmds/verbs/command.go — Verb listing and metadata-accurate help surfaces
- /home/manuel/workspaces/2026-04-13/js-loupedeck/loupedeck/examples/js/12-documented-scene.js — Annotated reference example

