# Tasks

## Phase 0: Design and ticket setup ✅
- [x] Research codebase (xgoja, gojahttp, loupedeck, database modules)
- [x] Create ticket LDCK-API-001
- [x] Write architecture and REST API design document
- [x] Write investigation diary
- [x] Validate docs with `docmgr doctor`
- [x] Upload to reMarkable

## Phase 1: Bootstrap xgoja buildspec ✅
- [x] Create `loupedeck-server/` directory with `xgoja.yaml`
- [x] Validate buildspec with `xgoja doctor`
- [x] Verify express, db, and loupedeck modules can be selected by runtime profile

## Phase 2: Recover xgoja server lifecycle ✅
- [x] Identify why built-in xgoja `run` is short-lived and unsuitable for HTTP services
- [x] Add `loupedeck-server/pkg/xgoja/serverprovider` CommandSetProvider
- [x] Add long-running `serve` command that loads `server.js` then waits for SIGINT/SIGTERM
- [x] Update `xgoja.yaml` commandProviders stanza
- [x] Build generated xgoja binary with `xgoja build`
- [x] Run API smoke tests against generated xgoja `serve` command (28/28 passing)

## Phase 3: REST API script and polling ✅
- [x] Write `server.js` with express routes for info, pages, display draw, brightness, buttons, events
- [x] Add SQLite `events` table
- [x] Wire `ui.onButton/onKnob/onTouch` callbacks to insert into event table
- [x] Implement `GET /api/v1/events?since=&limit=&type=`
- [x] Implement `DELETE /api/v1/events` acknowledge endpoint

## Phase 4: Clean hardware JS API (`loupedeck/hw`) ✅
- [x] 4a. Add `DeviceControl` interface to `runtime/js/env` using typed `device.Button` + `color.RGBA`
- [x] 4b. Add `LoupedeckDeviceControl` adapter for `*device.Loupedeck`
- [x] 4c. Add new `runtime/js/module_hw` module exporting `setBrightness` and `setButtonColor`
- [x] 4d. Add validation for brightness, button names, color objects, and `#rrggbb` strings
- [x] 4e. Register `loupedeck/hw` in `runtime/js/provider/provider.go`
- [x] 4f. Wire `environment.DeviceControl` in the existing xgoja hardware capability after hardware connects
- [x] 4g. Add unit tests for `loupedeck/hw` no-hardware, invalid input, and mock hardware success
- [x] 4h. Update `loupedeck-server/xgoja.yaml` runtime profile to include `loupedeck/hw`
- [x] 4i. Update `server.js` brightness/button endpoints to call `hw` and return 503 when unavailable
- [x] 4j. Rebuild generated binary and re-run API smoke tests (27/27 passing; hardware writes correctly return 503 without hardware)

## Phase 5: Hardware operator tests 🚧
- [x] 5a. Add hardware operator smoke script for visible display and LED confirmation
- [x] 5b. Test `PUT /api/v1/buttons/:name/color` on real hardware at HTTP/device-write level (operator visual confirmation still requested)
- [x] 5c. Test `PUT /api/v1/brightness` on real hardware at HTTP/device-write level (operator visual confirmation still requested)
- [ ] 5d. Test page/display rendering on real hardware with render diagnostics and operator confirmation
- [ ] 5e. Test button/knob/touch events end-to-end (hardware → SQLite → polling) with operator input

## Phase 6: Cleanup and documentation
- [ ] 6a. Decide whether to delete, build-tag, or archive manual `loupedeck-server/main.go` spike
- [ ] 6b. Update design/code-review docs with final `loupedeck/hw` API
- [ ] 6c. Re-upload updated docs to reMarkable

## DEFERRED (until go-go-goja gets a fetch module)
- [ ] Webhook registration and delivery
- [ ] HMAC signature headers

## Commits
- `ed576d5` — feat: standalone loupedeck-server binary with REST API
- `cf57f16` — chore: add .gitignore, remove db from tracking
- `2b226e4` — feat: add debug logging for hardware event callbacks
- `f6542ad` — feat: add xgoja long-running server command provider
- `3bf52bd` — feat: add loupedeck hardware JS module
- `faff68b` — feat: wire loupedeck hardware module into server API
- `609eacf` — docs: record loupedeck hw API implementation
- docs: add hardware operator validation script (latest docs commit; see `git log`)
