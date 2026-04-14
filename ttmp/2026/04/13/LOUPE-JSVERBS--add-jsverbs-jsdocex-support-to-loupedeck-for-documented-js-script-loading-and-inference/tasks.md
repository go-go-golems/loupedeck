# Tasks

## TODO

### 0. Baseline and dependency convergence

- [x] Add `github.com/go-go-golems/go-go-goja` as a direct dependency in `loupedeck/go.mod`
- [x] Align Go/toolchain requirements between `loupedeck` and `go-go-goja`
- [x] Resolve and document selected `goja` and `goja_nodejs` versions after convergence
- [x] Ensure tree-sitter dependencies required by `pkg/jsverbs` and `pkg/jsdoc` are present and build cleanly in the shared workspace
- [x] Run baseline tests in both repos and record any pre-existing failures before functional changes begin

### 1. Standardize loupedeck on go-go-goja runtime infrastructure

- [x] Create a loupedeck runtime registrar using `go-go-goja/engine.RuntimeModuleRegistrar`
- [x] Register all loupedeck native JS modules through that registrar
- [x] Seed loupedeck runtime-scoped host state from the registrar during runtime setup
- [x] Rename `Environment` to `LoupeDeckEnvironment` so the host-specific runtime state is explicit
- [x] Preserve the public `env.Lookup(vm)` API while reimplementing it without loupedeck-local `runtimebridge.Values`
- [x] Rework `pkg/jsmetrics.Lookup(vm)` to derive the collector from the loupedeck environment lookup instead of bridge values
- [x] Switch loupedeck JS modules to use upstream `go-go-goja/pkg/runtimebridge`
- [x] Switch loupedeck JS modules to use upstream `go-go-goja/pkg/runtimeowner`
- [x] Replace `runtime/js/runtime.go` with an engine-based helper or remove it entirely if no helper is needed
- [x] Remove loupedeck-local `pkg/runtimebridge` after migration is complete
- [x] Remove loupedeck-local `pkg/runtimeowner` after migration is complete

### 2. Keep current loupedeck run behavior working on the shared runtime

- [ ] Migrate `cmd/loupedeck/cmds/run/command.go` to create and own an `engine.Runtime`
- [ ] Preserve current raw `--script` execution behavior during the migration
- [ ] Preserve current renderer / present lifecycle and exit semantics
- [ ] Verify existing event callbacks still work correctly after runtime migration
- [ ] Verify runtime shutdown still cleans up device, renderer, and runtime resources in the correct order

### 3. Extend go-go-goja jsverbs for host-owned long-lived runtimes

- [ ] Add an exported jsverbs API for building command descriptions per verb without forcing the default runtime invocation path
- [ ] Add an exported jsverbs API for obtaining the scanned-source require loader / overlay loader
- [ ] Add an exported jsverbs API for invoking a verb inside an already-live caller-owned `engine.Runtime`
- [ ] Keep existing jsverbs convenience APIs working for current upstream callers
- [ ] Add upstream tests proving jsverbs invocation can reuse a live runtime without closing it
- [ ] Add upstream tests proving a runtime remains usable after jsverbs invocation completes

### 4. Integrate jsverbs into loupedeck scene execution

- [ ] Scan target script or script root with jsverbs when `--verb` is requested
- [ ] Compose the engine runtime with both loupedeck runtime registration and jsverbs scanned-source loading
- [ ] Use script-path-derived module roots so local `require("./...")` continues to work
- [ ] Add `--verb` support to the loupedeck run path
- [ ] If verbs are present, invoke the selected verb inside the already-live runtime
- [ ] Keep compatibility mode for plain scripts with no jsverbs metadata
- [ ] Verify a verb can configure a scene and leave the runtime alive for later callbacks and reactive updates
- [ ] Verify Glazed help/flags for `--verb` execution reflect jsverbs metadata accurately

### 5. Integrate jsdoc/jsdocex extraction

- [ ] Add script or directory scanning for jsdoc metadata using `pkg/jsdoc`
- [ ] Build a `DocStore` from loupedeck scene scripts
- [ ] Add a `loupedeck doc` CLI surface for extracted documentation
- [ ] Support at least `json` and `markdown` output modes for docs
- [ ] Ensure docs and verb scanning use the same script root and source set where appropriate

### 6. Add annotated reference examples

- [ ] Add one fully annotated loupedeck example script using `__package__`, `__section__`, `__verb__`, `__doc__`, and `doc\`...\``
- [ ] Ensure the example demonstrates correct `__verb__("functionName", {...})` string syntax
- [ ] Ensure the example covers at least one section binding and one context binding
- [ ] Ensure the example is usable both as a runnable scene and as a jsdoc extraction fixture

### 7. Testing and validation

- [ ] Update or replace runtime tests so they validate the engine-based runtime path instead of the removed local runtime wrapper
- [ ] Add loupedeck integration tests for `run --script ... --verb ...`
- [ ] Add tests for compatibility mode on plain non-jsverbs scripts
- [ ] Add tests for jsdoc extraction from the annotated reference example
- [ ] Run targeted tests in both repos after each major migration milestone
- [ ] Run full test suites in both repos before final review

### 8. Cleanup and documentation

- [ ] Remove stale comments and docs that reference the old loupedeck-local runtime ownership stack
- [ ] Update loupedeck help/docs to describe jsverbs-enabled scene scripts and the new `--verb` flow
- [ ] Update ticket docs with any design changes discovered during implementation
- [ ] Record final migration notes about removed duplicated runtime infrastructure
- [ ] Capture any follow-up tickets for optional doc server support, advanced multi-script support, or broader scripting ergonomics

## Done

- [x] Create the ticket workspace and primary docs
- [x] Analyze `pkg/jsverbs`, `pkg/jsdoc`, and current loupedeck runtime structure
- [x] Review the first-pass design and identify correctness / architecture issues
- [x] Replace the first-pass design with the revised runtime-convergence-first plan
- [x] Upload the revised analysis/design bundle to reMarkable
