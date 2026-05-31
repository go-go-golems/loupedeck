# Tasks

## Phase 1: Bootstrap the xgoja binary ✅
- [x] 1a. Create `loupedeck-server/` directory with `xgoja.yaml`
- [x] 1b. Validate buildspec with `xgoja doctor` — 18/18 checks pass
- [x] 1c. Build with `xgoja build` — binary at `dist/loupedeck-server`
- [x] 1d. Test: express, db, loupedeck/ui modules load correctly

## Phase 2: Custom Go binary + REST API ✅
- [x] 2a. Write custom `main.go` that creates engine.Runtime directly
- [x] 2b. Add SIGINT blocking so HTTP server stays alive
- [x] 2c. Wire gojahttp.Host, loupedeck/ui, database modules
- [x] 2d. Write `server.js` with all REST endpoints
- [x] 2e. Test: 28/28 smoke tests pass (no hardware)

## Phase 3: SQLite event queue + polling ✅
- [x] 3a. Add `database` module to runtime (db with allowConfigure:true)
- [x] 3b. Create `events` table in `server.js`
- [x] 3c. Wire `ui.onButton/onKnob/onTouch` → insert into events table
- [x] 3d. Implement `GET /api/v1/events?since=&limit=&type=`
- [x] 3e. Implement `DELETE /api/v1/events` (acknowledge)
- [x] 3f. Add debug logging to event callbacks

## Phase 4: Display control endpoints ✅
- [x] 4a. Implement page management (`POST /api/v1/pages`, `GET`, `DELETE`)
- [x] 4b. Implement `POST /api/v1/pages/show`
- [x] 4c. Implement display draw (`POST /api/v1/displays/:name/draw`)

## Phase 5: Extend loupedeck/ui module for hardware control
- [ ] 5a. Add `setButtonColor(name, r, g, b)` to `module_ui`
- [ ] 5b. Add `setBrightness(value)` to `module_ui`
- [ ] 5c. Rebuild binary, wire button color/brightness endpoints

## Phase 6: Smoke tests + tmux ✅
- [x] 6a. Write `01-smoke-test.sh` — 28/28 passing
- [x] 6b. Write `02-start-server.sh` — tmux launcher
- [x] 6c. Write `03-event-poll-demo.sh` — polling workflow demo
- [x] 6d. Write `04-stop-server.sh` — tmux stopper

## Phase 7: Hardware event verification
- [ ] 7a. Test event capture with real hardware (button press → SQLite → polling)
- [ ] 7b. Verify knob and touch event delivery
- [ ] 7c. Test display rendering on real hardware

## DEFERRED (until go-go-goja gets a fetch module)
- [ ] Webhook registration and delivery
- [ ] HMAC signature headers

## Commits
- `ed576d5` — feat: standalone loupedeck-server binary with REST API
- `cf57f16` — chore: add .gitignore, remove db from tracking
- `2b226e4` — feat: add debug logging for hardware event callbacks
