# Changelog

## 2026-04-11

Created the LOUPE-005 ticket and wrote the initial goja/JavaScript API brainstorming package, including a deep design document, a multi-approach example-script reference, and a continuity diary.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md — Main design brainstorm for the future JS runtime and API shapes
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md — Multi-scenario example scripts spanning several design styles
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/02-implementation-diary.md — Chronological record for this ticket

## 2026-04-11

Added an intern-oriented textbook for the preferred reactive runtime and a detailed implementation plan that breaks the work into pure-Go runtime phases, retained-UI phases, goja adapter phases, tests, acceptance criteria, and PR-sized milestones.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Conceptual textbook explaining signals, mutation semantics, retained UI, animation, and host/runtime responsibilities for a new intern
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Detailed phased build plan and milestone map for implementing the preferred reactive runtime

## 2026-04-11

Validated the expanded LOUPE-005 ticket docs with `docmgr doctor`, committed the new textbook and implementation-plan package to git, and uploaded the full intern-oriented bundle to the reMarkable under the existing LOUPE-005 folder.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/tasks.md — Updated to mark validation, commit, and reMarkable upload complete
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/02-implementation-diary.md — Chronological continuity record for the expanded documentation and delivery work
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md — Included in the uploaded reMarkable bundle
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Included in the uploaded reMarkable bundle
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md — Included in the uploaded reMarkable bundle
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/reference/01-javascript-api-example-scripts.md — Included in the uploaded reMarkable bundle

## 2026-04-11

Implemented milestone A of the reactive runtime as a pure-Go `runtime/reactive` package with signals, computed values, batching, eager watch/effect support, dependency tracking, cycle/reentrancy protection, and a focused unit-test suite. The implementation intentionally stayed goja-free so the semantic core could be validated in isolation before any JS bindings are added.

### Related Files

- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/runtime.go — Runtime coordination for batching, collector scoping, pending effect queues, and flush behavior
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/graph.go — Dependency graph primitives, dependent/source tracking, and default equality helpers
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/signal.go — Generic signal implementation with `Get`, `Set`, and `Update`
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/computed.go — Generic computed implementation with lazy reevaluation and dirty propagation
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/effect.go — Eager watch/effect implementation plus stop/unsubscribe support
- /home/manuel/code/wesen/2026-04-11--loupedeck-test/runtime/reactive/runtime_test.go — Unit tests covering equality no-ops, invalidation chains, diamond graphs, batching, stop behavior, and panic guards
