---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: go-go-goja/modules/database/database.go
      Note: SQLite database module with query/exec/configure
    - Path: go-go-goja/modules/express/express.go
      Note: Express-style JS API for route registration
    - Path: go-go-goja/pkg/gojahttp/host.go
      Note: gojahttp Host.ServeHTTP that routes HTTP requests into goja runtime
    - Path: go-go-goja/pkg/xgoja/providers/host/host.go
      Note: Guarded host modules (fs
    - Path: go-go-goja/pkg/xgoja/providers/http/http.go
      Note: HTTP capability that starts net/http server and wires express module
    - Path: loupedeck-server/main.go
      Note: Custom Go binary implementation
    - Path: loupedeck-server/server.js
      Note: REST API express routes and event listeners
    - Path: loupedeck/pkg/device/connect.go
      Note: Hardware connection via serial-over-USB
    - Path: loupedeck/pkg/device/loupedeck.go
      Note: Core Loupedeck struct with SetBrightness/SetButtonColor
    - Path: loupedeck/runtime/js/module_ui/module.go
      Note: UI module exposing pages/tiles/events to JS
    - Path: loupedeck/runtime/js/provider/provider.go
      Note: Loupedeck provider with hardware capability and module registration
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Architecture and REST API Design

## Executive Summary

This document describes the design and implementation of a standalone xgoja binary called `loupedeck-server` that exposes the Loupedeck Live hardware as a REST HTTP API. A coding agent (or any HTTP client) can use this API to control what appears on the Loupedeck's displays, set button colors, and subscribe to hardware events (button presses, knob turns, touches). The binary combines three existing xgoja provider packages — `loupedeck` (hardware UI modules), `go-go-goja-http` (express HTTP server), and `go-go-goja-host` (guarded database/filesystem/exec) — into a single generated program whose runtime profile wires an express-based REST API alongside the Loupedeck rendering pipeline.

The design supports two interaction patterns:

- **Polling**: the agent calls `GET /api/v1/events` to drain queued hardware events since the last poll.
- **Webhooks**: the agent registers a callback URL via `POST /api/v1/webhooks`, and the server POSTs hardware events to that URL as they arrive.

This document is written for an intern who is new to the codebase. It explains every layer — from the physical hardware protocol to the JavaScript module bindings to the generated Go binary — and provides file-anchored evidence, pseudocode, and concrete API contracts for every endpoint.

---

## Problem Statement

A coding agent running in a terminal or cloud environment has no way to interact with a Loupedeck Live device on the developer's desk. The agent cannot show status indicators, receive button press notifications, or trigger visual feedback on the device. Currently, the only way to drive the Loupedeck is through a JavaScript runtime that runs on the same machine as the device and connects via serial-over-USB.

The goal is to bridge this gap by running a small HTTP server on the machine that is physically connected to the Loupedeck, and exposing a REST API that any network client — including a remote coding agent — can call to control the device and receive events.

---

## Current-State Architecture

### The Loupedeck Hardware Stack

The Loupedeck Live is a USB control surface with:

- **3 displays**: left (60×270), main/center (360×270), right (60×270). The newer firmware exposes a unified 480×270 display on some models.
- **6 knobs** (Knob1–Knob6) with press detection (KnobPress1–6).
- **8 buttons** (Circle, Button1–Button7) below the display, with RGB LED control.
- **12 touch regions** (Touch1–12) on the center display, plus side touch areas (TouchLeft, TouchRight).

The device connects via USB and appears as a serial port. The Go driver (`loupedeck/pkg/device`) opens the serial port, wraps it in a WebSocket-like protocol, and provides high-level APIs for drawing images on displays, setting button colors, and registering event listeners.

Key files in the hardware stack:

| File | Role |
|------|------|
| `loupedeck/pkg/device/connect.go` | Auto-detect serial port, establish WebSocket connection |
| `loupedeck/pkg/device/loupedeck.go` | Core `Loupedeck` struct: displays, listeners, brightness, button colors |
| `loupedeck/pkg/device/display.go` | `Display.Draw(image, xoff, yoff)` — writes RGB565 framebuffer |
| `loupedeck/pkg/device/inputs.go` | Constants for Button, Knob, TouchButton; ParseButton/ParseKnob helpers |
| `loupedeck/pkg/device/listeners.go` | `OnButton`, `OnKnob`, `OnTouch` subscription API |
| `loupedeck/pkg/device/writer.go` | Outbound message queue with pacing and coalescing |
| `loupedeck/pkg/device/renderer.go` | Render scheduler that batches draw calls by display region |
| `loupedeck/pkg/device/message.go` | Wire protocol message types and serialization |

### The JavaScript Runtime Stack (xgoja)

The JavaScript runtime is built with xgoja, a build system that generates a Go binary from a declarative YAML spec. The generated binary embeds a goja (ECMAScript) engine and exposes Go-backed `require()` modules to JavaScript code.

The architecture has two separate decisions:

1. **Build-time package selection**: which Go provider packages are compiled into the binary.
2. **Runtime profile selection**: which compiled-in modules are available to a given command invocation.

Key concepts:

| Concept | Description |
|---------|-------------|
| **Provider package** | A Go package that exports a `Register(*providerapi.Registry)` function, contributing named modules and optional JavaScript verb sources |
| **Module** | A JavaScript `require()` target backed by Go code (e.g., `require("loupedeck/ui")`) |
| **Runtime profile** | A named set of module selections, each with a `package`, `name`, and `as` (the require alias) |
| **Buildspec** (`xgoja.yaml`) | The declarative YAML that names the binary, selects providers, defines profiles, and enables commands |
| **ConfigSectionCapability** | A provider extension that adds CLI flags (e.g., `--deck-enabled`, `--http-listen`) parsed from the runtime profile |
| **RuntimeInitializerCapability** | A provider extension that runs setup code (e.g., connecting to hardware, starting HTTP server) after the runtime is created |

Key provider packages we will use:

| Package ID | Go import | What it provides |
|------------|-----------|-------------------|
| `loupedeck` | `github.com/go-go-golems/loupedeck/pkg/xgoja/provider` | Modules: `loupedeck/ui`, `loupedeck/gfx`, `loupedeck/anim`, `loupedeck/state`, `loupedeck/present`, etc. Hardware capability that connects to the real device. |
| `go-go-goja-http` | `github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/http` | Module: `express`. HTTP capability that starts a Go HTTP server and routes requests through goja. |
| `go-go-goja-host` | `github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host` | Guarded modules: `fs`, `exec`, `database`, `db`. Require explicit `config.allow=true`. |

### The gojahttp Express Layer

The `gojahttp` package (`go-go-goja/pkg/gojahttp`) is the bridge between Go's `net/http` and JavaScript express-style handlers.

The flow for an HTTP request:

```
HTTP Request → gojahttp.Host.ServeHTTP()
  → Registry.Match(method, path)   // find registered JS handler
  → owner.Call("http-handler", fn)  // enter goja runtime on the owner goroutine
    → JS handler(req, res)          // execute JavaScript handler
    → res.json(data) / res.send()   // write response
  ← HTTP Response
```

The express module (`go-go-goja/modules/express/express.go`) exposes:

```javascript
const express = require("express")
const app = express()

app.get("/path", (req, res) => {
  res.json({ hello: "world" })
})
```

Under the hood, `app.get(pattern, handler)` calls `host.Register("GET", pattern, handler)`, which adds the route to the `gojahttp.Registry`. When the HTTP server receives a request, it matches the route and invokes the handler inside the goja runtime's serialized owner loop.

### The Loupedeck UI Module

The `loupedeck/ui` module (`loupedeck/runtime/js/module_ui/module.go`) exposes the retained UI model to JavaScript:

```javascript
const ui = require("loupedeck/ui")

ui.onButton("Button1", (event) => {
  // event.name, event.status ("down" / "up")
})

ui.onKnob("Knob1", (event) => {
  // event.name, event.value (+1 / -1)
})

ui.onTouch("Touch1", (event) => {
  // event.name, event.status, event.x, event.y
})

const page = ui.page("main", (page) => {
  const tile = page.tile(0, 0, (tile) => {
    tile.text("Hello")
    tile.icon("check")
  })
})

ui.show("main")
ui.invalidate("data-changed")
```

The `LoupeDeckEnvironment` (`loupedeck/runtime/js/env/env.go`) is the central state holder:

- `Reactive`: reactive signal/effect runtime
- `UI`: retained UI tree (pages, tiles, displays)
- `Host`: event source binding layer (buttons, knobs, touches)
- `Anim`: animation timeline
- `Present`: render scheduler that flushes UI tree to hardware displays
- `Metrics`: performance counters

The hardware capability (`loupedeck/runtime/js/provider/provider.go`) in `InitRuntimeFromSections` does the following:

1. Parses `--deck-enabled`, `--deck-device`, `--deck-queue-size`, `--deck-send-interval`, `--deck-flush-interval` flags.
2. Calls `device.ConnectAuto()` (or `ConnectPath`) to open the serial device.
3. Attaches the device as an `EventSource` to the `Host` runtime.
4. Starts the `Present` flush loop that invalidates dirty UI regions onto the hardware displays.
5. Registers a closer that blanks displays and closes the connection on shutdown.

---

## Gap Analysis

The current system has all the pieces needed to control the Loupedeck from JavaScript, but there is no way to expose that control surface over the network. Specifically:

1. **No REST API**: The express module can serve HTTP, but no JavaScript code currently registers routes that expose the Loupedeck's state and events as a REST API.
2. **No event queue**: Hardware events (button, knob, touch) are dispatched to JavaScript callbacks in real-time, but there is no persisted queue that a polling client could drain.
3. **No webhook mechanism**: There is no way to register a URL and have the server POST events to it.
4. **No persisted state**: The database module (`go-go-goja/modules/database`) can create SQLite databases, but no schema exists for storing event queues or webhook registrations.
5. **No generated binary spec**: No `xgoja.yaml` combines the loupedeck, http, and host providers into a single build.

---

## Proposed Solution

### Overview

Create a new xgoja buildspec (`xgoja.yaml`) that combines the loupedeck, http, and host providers, plus a JavaScript application file (`server.js`) that:

1. Creates an express app with a REST API under `/api/v1/`.
2. Uses the `database` module to create a SQLite-backed event queue and webhook registry.
3. Uses the `loupedeck/ui` module to listen for hardware events and persist them to the queue.
4. Exposes endpoints for display control, button colors, brightness, event polling, and webhook management.
5. Optionally serves a simple dashboard on `/` for visual debugging.

### System Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                      Coding Agent (remote)                       │
│                                                                  │
│   POST /api/v1/webhooks        GET /api/v1/events?since=123      │
│   POST /api/v1/displays/main   PUT /api/v1/buttons/Button1/color │
└────────────────────────┬─────────────────────────────────────────┘
                         │ HTTP
                         ▼
┌──────────────────────────────────────────────────────────────────┐
│                    loupedeck-server (xgoja binary)               │
│                                                                  │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐    │
│  │  express    │  │  loupedeck/  │  │  database (SQLite)    │    │
│  │  (gojahttp) │  │  ui,gfx,anim │  │                      │    │
│  │             │  │  state,pres  │  │  events_queue table   │    │
│  │  REST API   │  │              │  │  webhooks table       │    │
│  │  routes     │  │  hardware    │  │  pages table          │    │
│  └──────┬──────┘  │  callbacks   │  └──────────┬───────────┘    │
│         │         └──────┬───────┘             │                 │
│         │                │                     │                 │
│         ▼                ▼                     ▼                 │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │              Goja Runtime (serialized owner loop)        │    │
│  │                                                          │    │
│  │   require("express")   require("loupedeck/ui")          │    │
│  │   require("database")  require("loupedeck/gfx")         │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │              Hardware Layer (device package)             │    │
│  │                                                          │    │
│  │   Serial-over-USB → WebSocket protocol → Loupedeck Live  │    │
│  │   Displays (RGB565)  Buttons (RGB LED)  Knobs/Touch     │    │
│  └──────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

### REST API Design

All endpoints are under `/api/v1/`. The API follows REST conventions: resources are nouns, actions are HTTP methods, and responses are JSON.

#### Connection & Device Info

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/info` | Device info: model, version, serial, display dimensions, connected status |

**Response**:
```json
{
  "connected": true,
  "model": "Loupedeck Live",
  "version": "1.0.0",
  "serialNo": "LD12345",
  "displays": {
    "left":  { "width": 60,  "height": 270 },
    "main":  { "width": 360, "height": 270 },
    "right": { "width": 60,  "height": 270 }
  },
  "buttons": ["Circle","Button1","Button2","Button3","Button4","Button5","Button6","Button7"],
  "knobs": ["Knob1","Knob2","Knob3","Knob4","Knob5","Knob6"],
  "touches": ["Touch1","Touch2","Touch3","Touch4","Touch5","Touch6","Touch7","Touch8","Touch9","Touch10","Touch11","Touch12"]
}
```

#### Brightness

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/brightness` | Get current brightness |
| `PUT` | `/api/v1/brightness` | Set brightness (0–10) |

**PUT body**: `{ "value": 9 }`

#### Button Colors

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/buttons` | List all buttons with current colors |
| `GET` | `/api/v1/buttons/:name` | Get one button's current color |
| `PUT` | `/api/v1/buttons/:name/color` | Set button LED color |

**PUT body**: `{ "r": 255, "g": 0, "b": 128 }`

**GET /api/v1/buttons response**:
```json
{
  "buttons": {
    "Circle":  { "r": 0, "g": 255, "b": 0 },
    "Button1": { "r": 255, "g": 0, "b": 0 }
  }
}
```

#### Display Control

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/displays` | List available displays with dimensions |
| `POST` | `/api/v1/displays/:name/draw` | Draw a tile/page to a display region |

The draw endpoint accepts a JSON payload that describes what to render. The server translates this into tile/page operations using the retained UI model.

**POST body** (tile grid mode — the most common for agent use):
```json
{
  "layout": "grid",
  "tiles": [
    { "col": 0, "row": 0, "text": "BUILD", "icon": "hammer", "bg": "#333333", "fg": "#00ff00" },
    { "col": 1, "row": 0, "text": "TEST",  "icon": "check",  "bg": "#333333", "fg": "#ffff00" },
    { "col": 2, "row": 0, "text": "DEPLOY","icon": "rocket", "bg": "#333333", "fg": "#00ff00" },
    { "col": 3, "row": 0, "text": "OK ✓",  "bg": "#003300", "fg": "#00ff00" },
    { "col": 0, "row": 1, "text": "agent: thinking...", "bg": "#000033", "fg": "#6666ff", "wrap": true },
    { "col": 0, "row": 2, "text": "", "bg": "#000000" }
  ],
  "display": "main"
}
```

**POST body** (raw pixel mode — for custom graphics):
```json
{
  "layout": "raw",
  "display": "main",
  "x": 0,
  "y": 0,
  "width": 360,
  "height": 270,
  "pixels": "<base64-encoded RGBA>"
}
```

**POST body** (page show mode — switch to an existing page):
```json
{
  "layout": "show",
  "page": "status"
}
```

#### Page Management

The retained UI model supports named pages. Agents can create pages and switch between them.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/pages` | List all pages |
| `POST` | `/api/v1/pages` | Create a new page with tiles |
| `GET` | `/api/v1/pages/:name` | Get page details (tiles) |
| `DELETE` | `/api/v1/pages/:name` | Delete a page |
| `POST` | `/api/v1/pages/show` | Show a page on the display |

**POST /api/v1/pages body**:
```json
{
  "name": "agent-status",
  "tiles": [
    { "col": 0, "row": 0, "text": "Thinking...", "icon": "spinner", "fg": "#6666ff" },
    { "col": 1, "row": 0, "text": "Step 3/10", "fg": "#ffffff" }
  ]
}
```

**POST /api/v1/pages/show body**: `{ "name": "agent-status" }`

#### Tile Updates (Partial)

For efficient updates without recreating the whole page:

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/api/v1/pages/:name/tiles/:col/:row` | Update a specific tile |

**PUT body**:
```json
{
  "text": "Step 4/10",
  "icon": "check",
  "fg": "#00ff00"
}
```

#### Events (Polling)

Hardware events are persisted in a SQLite table. The agent polls to drain events since a cursor.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/events` | Get events since a cursor |
| `DELETE` | `/api/v1/events` | Delete events up to a cursor (acknowledge) |

**GET query parameters**:

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `since` | int | 0 | Return events with id > since |
| `limit` | int | 100 | Maximum events to return |
| `type` | string | "" | Filter by type: `button`, `knob`, `touch` |

**Response**:
```json
{
  "cursor": 42,
  "events": [
    { "id": 38, "type": "button", "name": "Button1", "status": "down", "timestamp": 1717100000 },
    { "id": 39, "type": "button", "name": "Button1", "status": "up",   "timestamp": 1717100001 },
    { "id": 40, "type": "knob",   "name": "Knob2",   "value": 1,       "timestamp": 1717100002 },
    { "id": 41, "type": "touch",  "name": "Touch3",   "status": "down", "x": 150, "y": 90, "timestamp": 1717100003 }
  ]
}
```

**DELETE body**: `{ "until": 42 }` — deletes events with id <= 42.

#### Webhooks

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/webhooks` | List registered webhooks |
| `POST` | `/api/v1/webhooks` | Register a new webhook |
| `DELETE` | `/api/v1/webhooks/:id` | Delete a webhook |

**POST body**:
```json
{
  "url": "https://my-agent.example.com/loupedeck-events",
  "secret": "shared-hmac-secret",
  "events": ["button", "knob", "touch"],
  "method": "POST",
  "headers": { "Authorization": "Bearer token123" }
}
```

When a hardware event occurs, the server:

1. Persists the event to the SQLite queue (for polling).
2. For each registered webhook whose `events` filter matches, POSTs the event as JSON to the webhook URL.
3. The POST includes an `X-Loupedeck-Signature` header with `HMAC-SHA256(secret, body)`.
4. Failed webhook deliveries are retried with exponential backoff (up to 3 attempts).

**Webhook delivery body** (POSTed to the registered URL):
```json
{
  "event": { "id": 38, "type": "button", "name": "Button1", "status": "down", "timestamp": 1717100000 },
  "device": { "model": "Loupedeck Live", "serial": "LD12345" }
}
```

---

## xgoja Buildspec

The generated binary is defined by an `xgoja.yaml` buildspec:

```yaml
name: loupedeck-server
go:
  version: "1.26"
  module: github.com/go-go-golems/loupedeck-server

target:
  kind: xgoja
  output: dist/loupedeck-server

packages:
  - id: loupedeck
    import: github.com/go-go-golems/loupedeck/pkg/xgoja/provider
    register: Register
    replace: ../loupedeck

  - id: go-go-goja-http
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/http
    register: Register

  - id: go-go-goja-host
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host
    register: Register

  - id: go-go-goja-core
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/core
    register: Register

runtimes:
  server:
    modules:
      # Express HTTP server
      - package: go-go-goja-http
        name: express
        as: express

      # Loupedeck hardware modules
      - package: loupedeck
        name: loupedeck/ui
        as: loupedeck/ui
      - package: loupedeck
        name: loupedeck/gfx
        as: loupedeck/gfx
      - package: loupedeck
        name: loupedeck/state
        as: loupedeck/state
      - package: loupedeck
        name: loupedeck/present
        as: loupedeck/present
      - package: loupedeck
        name: loupedeck/anim
        as: loupedeck/anim

      # Database for event queue and webhook storage
      - package: go-go-goja-host
        name: db
        as: db
        config:
          allowConfigure: true

      # Path and YAML helpers
      - package: go-go-goja-core
        name: path
        as: path
      - package: go-go-goja-core
        name: yaml
        as: yaml

commands:
  run:
    enabled: true
    runtime: server
    name: serve
  eval:
    enabled: true
    runtime: server
    name: eval
  repl:
    enabled: true
    runtime: server
    name: repl

jsverbs:
  - id: server-verbs
    path: ./verbs
    embed: true
```

Running the server:

```bash
# Build
xgoja build -f xgoja.yaml --xgoja-replace /path/to/go-go-goja --keep-work

# Run with hardware + HTTP
./dist/loupedeck-server serve server.js \
  --deck-enabled \
  --deck-device "" \
  --http-enabled \
  --http-listen 0.0.0.0:9876

# Run without hardware (mock mode, for development)
./dist/loupedeck-server serve server.js \
  --deck-enabled=false \
  --http-enabled \
  --http-listen 127.0.0.1:9876
```

---

## JavaScript Application: server.js

The `server.js` file is the heart of the system. It:

1. Initializes the SQLite database (event queue + webhook tables).
2. Registers express routes for the REST API.
3. Subscribes to hardware events via `loupedeck/ui`.
4. Starts the webhook delivery loop.

### Pseudocode

```javascript
const express = require("express")
const ui = require("loupedeck/ui")
const db = require("db")

// ── 1. Database Setup ──────────────────────────────────────────

db.configure("sqlite3", "loupedeck_events.db")

db.exec(`
  CREATE TABLE IF NOT EXISTS events (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    type      TEXT NOT NULL,           -- 'button' | 'knob' | 'touch'
    name      TEXT NOT NULL,           -- 'Button1', 'Knob3', 'Touch5'
    status    TEXT,                    -- 'down' | 'up' (button/touch)
    value     INTEGER,                -- knob delta (+1/-1)
    x         INTEGER,                -- touch x coord
    y         INTEGER,                -- touch y coord
    timestamp INTEGER NOT NULL
  )
`)

db.exec(`
  CREATE TABLE IF NOT EXISTS webhooks (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    url     TEXT NOT NULL,
    secret  TEXT DEFAULT '',
    events  TEXT DEFAULT '[]',         -- JSON array of event types
    method  TEXT DEFAULT 'POST',
    headers TEXT DEFAULT '{}',         -- JSON object
    active  INTEGER DEFAULT 1
  )
`)

// ── 2. Express App ─────────────────────────────────────────────

const app = express()

// ── Helper: parse JSON body ────────────────────────────────────

function jsonBody(req) {
  if (typeof req.body === "object") return req.body
  try { return JSON.parse(req.rawBody) } catch { return null }
}

// ── Device Info ────────────────────────────────────────────────

app.get("/api/v1/info", (req, res) => {
  // The environment is shared; we read from ui module state
  res.json({
    connected: true,  // updated by hardware capability
    model: "Loupedeck Live",
    displays: {
      left:  { width: 60,  height: 270 },
      main:  { width: 360, height: 270 },
      right: { width: 60,  height: 270 }
    },
    buttons: ["Circle","Button1","Button2","Button3","Button4","Button5","Button6","Button7"],
    knobs: ["Knob1","Knob2","Knob3","Knob4","Knob5","Knob6"],
    touches: ["Touch1","Touch2","Touch3","Touch4","Touch5","Touch6","Touch7","Touch8","Touch9","Touch10","Touch11","Touch12"]
  })
})

// ── Brightness ─────────────────────────────────────────────────

app.get("/api/v1/brightness", (req, res) => {
  // Brightness is tracked in app state
  res.json({ value: currentBrightness })
})

app.put("/api/v1/brightness", (req, res) => {
  const body = jsonBody(req)
  if (!body || typeof body.value !== "number") {
    res.status(400).json({ error: "missing or invalid 'value'" })
    return
  }
  currentBrightness = body.value
  // The hardware SetBrightness is called through the ui module's
  // environment → deckConn.SetBrightness(value)
  // We expose this through a dedicated page/display call
  res.json({ value: currentBrightness })
})

// ── Buttons ────────────────────────────────────────────────────

app.get("/api/v1/buttons", (req, res) => {
  res.json({ buttons: buttonColors })
})

app.get("/api/v1/buttons/:name", (req, res) => {
  const color = buttonColors[req.params.name]
  if (!color) {
    res.status(404).json({ error: `unknown button ${req.params.name}` })
    return
  }
  res.json({ name: req.params.name, ...color })
})

app.put("/api/v1/buttons/:name/color", (req, res) => {
  const body = jsonBody(req)
  if (!body || typeof body.r !== "number") {
    res.status(400).json({ error: "missing r,g,b" })
    return
  }
  const name = req.params.name
  buttonColors[name] = { r: body.r, g: body.g, b: body.b }
  // Call hardware: deckConn.SetButtonColor(button, color.RGBA{R,G,B})
  // This needs a small Go bridge or an extended ui module API
  res.json({ name, ...buttonColors[name] })
})

// ── Pages ──────────────────────────────────────────────────────

const pages = {}  // name → { page, tiles }

app.get("/api/v1/pages", (req, res) => {
  res.json({ pages: Object.keys(pages) })
})

app.post("/api/v1/pages", (req, res) => {
  const body = jsonBody(req)
  const page = ui.page(body.name, (p) => {
    for (const t of body.tiles) {
      p.tile(t.col, t.row, (tile) => {
        if (t.text)  tile.text(t.text)
        if (t.icon)  tile.icon(t.icon)
      })
    }
  })
  pages[body.name] = { page, tiles: body.tiles }
  res.json({ name: body.name, tiles: body.tiles.length })
})

app.get("/api/v1/pages/:name", (req, res) => {
  const p = pages[req.params.name]
  if (!p) {
    res.status(404).json({ error: `page ${req.params.name} not found` })
    return
  }
  res.json({ name: req.params.name, tiles: p.tiles })
})

app.delete("/api/v1/pages/:name", (req, res) => {
  delete pages[req.params.name]
  res.json({ deleted: req.params.name })
})

app.post("/api/v1/pages/show", (req, res) => {
  const body = jsonBody(req)
  ui.show(body.name)
  res.json({ showing: body.name })
})

// ── Display Draw ───────────────────────────────────────────────

app.post("/api/v1/displays/:name/draw", (req, res) => {
  const body = jsonBody(req)
  if (body.layout === "grid") {
    // Create a temporary page, add tiles, show it
    const pageName = `_draw_${Date.now()}`
    ui.page(pageName, (p) => {
      for (const t of body.tiles) {
        p.tile(t.col, t.row, (tile) => {
          if (t.text)  tile.text(t.text)
          if (t.icon)  tile.icon(t.icon)
        })
      }
    })
    ui.show(pageName)
    res.json({ ok: true, page: pageName })
  } else if (body.layout === "show") {
    ui.show(body.page)
    res.json({ ok: true, showing: body.page })
  } else {
    res.status(400).json({ error: `unsupported layout: ${body.layout}` })
  }
})

// ── Events (Polling) ───────────────────────────────────────────

app.get("/api/v1/events", (req, res) => {
  const since = parseInt(req.query.since || "0")
  const limit = parseInt(req.query.limit || "100")
  const type  = req.query.type || ""

  let sql = "SELECT * FROM events WHERE id > ?"
  const args = [since]
  if (type) {
    sql += " AND type = ?"
    args.push(type)
  }
  sql += " ORDER BY id ASC LIMIT ?"
  args.push(limit)

  const rows = db.query(sql, ...args)
  const cursor = rows.length > 0 ? rows[rows.length - 1].id : since
  res.json({ cursor, events: rows })
})

app.delete("/api/v1/events", (req, res) => {
  const body = jsonBody(req)
  if (!body || typeof body.until !== "number") {
    res.status(400).json({ error: "missing 'until'" })
    return
  }
  db.exec("DELETE FROM events WHERE id <= ?", body.until)
  res.json({ deleted_until: body.until })
})

// ── Webhooks ───────────────────────────────────────────────────

app.get("/api/v1/webhooks", (req, res) => {
  const rows = db.query("SELECT * FROM webhooks WHERE active = 1")
  res.json({ webhooks: rows })
})

app.post("/api/v1/webhooks", (req, res) => {
  const body = jsonBody(req)
  if (!body || !body.url) {
    res.status(400).json({ error: "missing 'url'" })
    return
  }
  const events = JSON.stringify(body.events || ["button","knob","touch"])
  const headers = JSON.stringify(body.headers || {})
  db.exec(
    "INSERT INTO webhooks (url, secret, events, method, headers) VALUES (?, ?, ?, ?, ?)",
    body.url, body.secret || "", events, body.method || "POST", headers
  )
  const rows = db.query("SELECT last_insert_rowid() as id")
  res.json({ id: rows[0].id, url: body.url })
})

app.delete("/api/v1/webhooks/:id", (req, res) => {
  const id = parseInt(req.params.id)
  db.exec("UPDATE webhooks SET active = 0 WHERE id = ?", id)
  res.json({ deleted: id })
})

// ── 3. Hardware Event Listeners ────────────────────────────────

const ALL_BUTTONS = ["Circle","Button1","Button2","Button3","Button4","Button5","Button6","Button7"]
const ALL_KNOBS   = ["Knob1","Knob2","Knob3","Knob4","Knob5","Knob6"]
const ALL_TOUCHES = ["Touch1","Touch2","Touch3","Touch4","Touch5","Touch6","Touch7","Touch8","Touch9","Touch10","Touch11","Touch12"]

const buttonColors = {}
ALL_BUTTONS.forEach(name => { buttonColors[name] = { r: 0, g: 0, b: 0 } })

let currentBrightness = 9

// Record a hardware event to SQLite and dispatch webhooks
function recordEvent(event) {
  const ts = Math.floor(Date.now() / 1000)
  db.exec(
    "INSERT INTO events (type, name, status, value, x, y, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)",
    event.type, event.name,
    event.status || null,
    event.value || null,
    event.x || null,
    event.y || null,
    ts
  )
  dispatchWebhooks(event)
}

// Subscribe to all buttons
ALL_BUTTONS.forEach(name => {
  ui.onButton(name, (event) => {
    recordEvent({ type: "button", name, status: event.status })
  })
})

// Subscribe to all knobs
ALL_KNOBS.forEach(name => {
  ui.onKnob(name, (event) => {
    recordEvent({ type: "knob", name, value: event.value })
  })
})

// Subscribe to all touch regions
ALL_TOUCHES.forEach(name => {
  ui.onTouch(name, (event) => {
    recordEvent({ type: "touch", name, status: event.status, x: event.x, y: event.y })
  })
})

// ── 4. Webhook Delivery ────────────────────────────────────────

function dispatchWebhooks(event) {
  const webhooks = db.query("SELECT * FROM webhooks WHERE active = 1")
  for (const wh of webhooks) {
    let filter = ["button","knob","touch"]
    try { filter = JSON.parse(wh.events) } catch {}
    if (filter.length > 0 && !filter.includes(event.type)) continue

    // Fire-and-forget HTTP POST (asynchronous, retried)
    deliverWebhook(wh, event)
  }
}

function deliverWebhook(wh, event) {
  const payload = JSON.stringify({
    event: {
      type: event.type,
      name: event.name,
      status: event.status,
      value: event.value,
      x: event.x,
      y: event.y,
      timestamp: Math.floor(Date.now() / 1000)
    }
  })
  // NOTE: Actual HTTP delivery requires the exec module or a dedicated
  // Go-backed fetch/curl helper. For MVP, we log and use exec("curl", ...).
  // A cleaner approach would add a fetch() native module.
  const headers = (() => {
    try { return JSON.parse(wh.headers) } catch { return {} }
  })()
  // This is simplified; production would use a proper fetch module
  console.log(`[webhook] → ${wh.url} : ${payload}`)
}
```

### Key Design Decisions in server.js

1. **SQLite for event queue**: The `database` module provides a built-in SQLite engine. Using `AUTOINCREMENT` gives us a natural cursor for polling. The `since` parameter is an event ID, not a timestamp, which avoids clock-skew issues.

2. **Express for routing**: The `express` module's `app.get/post/put/delete` API maps directly to our REST endpoints. The gojahttp `Host.ServeHTTP` method routes each request through the goja runtime's serialized owner loop, ensuring thread safety.

3. **Page abstraction for display control**: Rather than exposing raw pixel drawing over HTTP, we use the retained UI model (pages/tiles). This gives us efficient partial updates via `tile.text()` / `tile.icon()` and automatic re-rendering via the `Present` flush loop.

4. **Webhook delivery limitation**: The current module set does not include a `fetch()` or HTTP client module. For the MVP, webhook delivery can use the `exec` module to call `curl`, or a dedicated native `fetch` module should be added as a follow-up. The design document notes this gap explicitly.

---

## Implementation Plan

### Phase 1: Bootstrap the xgoja Binary (Day 1)

**Goal**: Get a generated binary that starts an HTTP server and connects to the Loupedeck.

1. Create the `xgoja.yaml` buildspec in a new directory `loupedeck-server/`.
2. Create `server.js` with the `/api/v1/info` endpoint (hard-coded, no database yet).
3. Run `xgoja doctor -f xgoja.yaml` and fix validation errors.
4. Run `xgoja build -f xgoja.yaml --xgoja-replace /path/to/go-go-goja --keep-work`.
5. Test: `./dist/loupedeck-server serve server.js --deck-enabled --http-listen :9876`.
6. Verify: `curl http://localhost:9876/api/v1/info` returns JSON.

**Files to create**:
- `loupedeck-server/xgoja.yaml`
- `loupedeck-server/server.js`

**Files to reference**:
- `go-go-goja/pkg/xgoja/providers/http/http.go` — HTTP capability and express loader
- `loupedeck/pkg/xgoja/provider/provider.go` — Loupedeck provider registration
- `loupedeck/runtime/js/provider/provider.go` — Hardware capability and runtime initializer
- `go-go-goja/pkg/xgoja/providers/host/host.go` — Guarded database module

### Phase 2: Database + Event Queue + Polling (Day 2)

**Goal**: Persist hardware events in SQLite and expose the polling API.

1. Add `database` module to the runtime profile with `allowConfigure: true`.
2. Create the `events` table in `server.js`.
3. Wire `ui.onButton/onKnob/onTouch` callbacks to insert into the table.
4. Implement `GET /api/v1/events` with `since`, `limit`, `type` query params.
5. Implement `DELETE /api/v1/events` with `until` body param.
6. Test: press a button, then `curl http://localhost:9876/api/v1/events?since=0`.

**Files to modify**:
- `loupedeck-server/server.js`

**Reference files**:
- `go-go-goja/modules/database/database.go` — query/exec API signature
- `loupedeck/runtime/js/module_ui/module.go` — onButton/onKnob/onTouch callback shapes

### Phase 3: Display Control + Button Colors (Day 3)

**Goal**: Full control over displays and buttons via the REST API.

1. Implement page management endpoints (`POST /api/v1/pages`, `POST /api/v1/pages/show`).
2. Implement tile update endpoint (`PUT /api/v1/pages/:name/tiles/:col/:row`).
3. Implement display draw endpoint (`POST /api/v1/displays/:name/draw`).
4. Implement button color endpoint (`PUT /api/v1/buttons/:name/color`).
5. Implement brightness endpoint (`PUT /api/v1/brightness`).

**Challenge**: Button colors and brightness are hardware-level operations (`deckConn.SetButtonColor`, `deckConn.SetBrightness`) that are not currently exposed through the `loupedeck/ui` module. The `loupedeck/ui` module focuses on the retained UI model (pages/tiles/displays). We need one of:

- **Option A**: Extend the `loupedeck/ui` module to expose `setButtonColor(name, r, g, b)` and `setBrightness(value)` functions.
- **Option B**: Create a new `loupedeck/hw` module that exposes raw hardware control.
- **Option C**: Use the `exec` module to call a separate CLI tool for these operations.

**Recommendation**: Option A is the cleanest. Add two exports to `module_ui`:

```javascript
// In loupedeck/ui module additions:
exports.setButtonColor = function(name, r, g, b) { ... }
exports.setBrightness = function(value) { ... }
```

These would delegate to `env.Host` → `deckConn.SetButtonColor` / `deckConn.SetBrightness`.

### Phase 4: Webhook Delivery (Day 4)

**Goal**: Register webhook URLs and deliver events via HTTP POST.

1. Create the `webhooks` table in `server.js`.
2. Implement `POST /api/v1/webhooks`, `GET /api/v1/webhooks`, `DELETE /api/v1/webhooks/:id`.
3. Implement webhook delivery: on each hardware event, query active webhooks, filter by event type, POST to each URL.

**Challenge**: JavaScript in goja has no built-in `fetch()`. Options:

- **Option A**: Use `require("exec").run("curl", ["-s", "-X", "POST", ...])` — simple but slow (spawns a process per delivery).
- **Option B**: Add a `fetch` native module to go-go-goja that wraps Go's `net/http` client — clean, efficient, reusable.
- **Option C**: Use a Go-backed webhook delivery goroutine that reads from a channel — most robust, but requires Go code changes.

**Recommendation**: Option B for the MVP, with a follow-up to Option C for production reliability. The `fetch` module would expose:

```javascript
const fetch = require("fetch")
const response = fetch("https://example.com/hook", {
  method: "POST",
  headers: { "Content-Type": "application/json", "X-Signature": sig },
  body: JSON.stringify(payload)
})
```

For the initial implementation, we can use Option A (curl via exec) to unblock testing.

### Phase 5: Polish + Testing (Day 5)

1. Add input validation and error handling to all endpoints.
2. Add CORS headers for browser-based dashboard access.
3. Create a simple HTML dashboard served on `/` that shows device info and recent events.
4. Write integration tests that start the server, make HTTP requests, and verify responses.
5. Write a mock mode that works without real hardware (`--deck-enabled=false`).
6. Add graceful shutdown (SIGINT handler via runtime closer).

---

## API Reference (Complete)

### `GET /api/v1/info`

Returns device information and available controls.

**Response 200**:
```json
{
  "connected": true,
  "model": "Loupedeck Live",
  "version": "1.0.0",
  "serialNo": "LD12345",
  "displays": {
    "left":  { "width": 60,  "height": 270 },
    "main":  { "width": 360, "height": 270 },
    "right": { "width": 60,  "height": 270 }
  },
  "buttons": ["Circle","Button1","Button2","Button3","Button4","Button5","Button6","Button7"],
  "knobs": ["Knob1","Knob2","Knob3","Knob4","Knob5","Knob6"],
  "touches": ["Touch1","Touch2","Touch3","Touch4","Touch5","Touch6","Touch7","Touch8","Touch9","Touch10","Touch11","Touch12"]
}
```

### `GET /api/v1/brightness`

Returns current brightness level.

**Response 200**: `{ "value": 9 }`

### `PUT /api/v1/brightness`

Sets brightness level (0–10).

**Request body**: `{ "value": 9 }`
**Response 200**: `{ "value": 9 }`
**Response 400**: `{ "error": "missing or invalid 'value'" }`

### `GET /api/v1/buttons`

Lists all buttons with current LED colors.

**Response 200**:
```json
{
  "buttons": {
    "Circle":  { "r": 0, "g": 255, "b": 0 },
    "Button1": { "r": 255, "g": 0, "b": 0 }
  }
}
```

### `GET /api/v1/buttons/:name`

Gets one button's LED color.

**Response 200**: `{ "name": "Circle", "r": 0, "g": 255, "b": 0 }`
**Response 404**: `{ "error": "unknown button Foo" }`

### `PUT /api/v1/buttons/:name/color`

Sets a button's LED color.

**Request body**: `{ "r": 255, "g": 0, "b": 128 }`
**Response 200**: `{ "name": "Button1", "r": 255, "g": 0, "b": 128 }`
**Response 400**: `{ "error": "missing r,g,b" }`
**Response 404**: `{ "error": "unknown button Foo" }`

### `GET /api/v1/displays`

Lists available displays.

**Response 200**:
```json
{
  "displays": {
    "left":  { "width": 60,  "height": 270 },
    "main":  { "width": 360, "height": 270 },
    "right": { "width": 60,  "height": 270 }
  }
}
```

### `POST /api/v1/displays/:name/draw`

Draws content to a display.

**Request body (grid)**:
```json
{
  "layout": "grid",
  "display": "main",
  "tiles": [
    { "col": 0, "row": 0, "text": "Hello", "icon": "check", "fg": "#00ff00" }
  ]
}
```

**Response 200**: `{ "ok": true, "page": "_draw_1717100000" }`
**Response 400**: `{ "error": "unsupported layout: raw" }`

### `GET /api/v1/pages`

Lists all named pages.

**Response 200**: `{ "pages": ["status", "agent-status"] }`

### `POST /api/v1/pages`

Creates a new page with tiles.

**Request body**:
```json
{
  "name": "status",
  "tiles": [
    { "col": 0, "row": 0, "text": "OK", "icon": "check", "fg": "#00ff00" }
  ]
}
```

**Response 200**: `{ "name": "status", "tiles": 1 }`
**Response 400**: `{ "error": "missing 'name'" }`

### `GET /api/v1/pages/:name`

Gets page details.

**Response 200**: `{ "name": "status", "tiles": [...] }`
**Response 404**: `{ "error": "page status not found" }`

### `DELETE /api/v1/pages/:name`

Deletes a page.

**Response 200**: `{ "deleted": "status" }`

### `POST /api/v1/pages/show`

Shows a page on the device.

**Request body**: `{ "name": "status" }`
**Response 200**: `{ "showing": "status" }`

### `PUT /api/v1/pages/:name/tiles/:col/:row`

Updates a specific tile on a page (partial update).

**Request body**:
```json
{ "text": "BUILDING", "fg": "#ff6600" }
```

**Response 200**: `{ "ok": true }`
**Response 404**: `{ "error": "page status not found" }`

### `GET /api/v1/events`

Polls hardware events since a cursor.

**Query params**: `since` (int, default 0), `limit` (int, default 100), `type` (string, optional filter)

**Response 200**:
```json
{
  "cursor": 42,
  "events": [
    { "id": 38, "type": "button", "name": "Button1", "status": "down", "timestamp": 1717100000 }
  ]
}
```

### `DELETE /api/v1/events`

Acknowledges events up to a cursor (removes them).

**Request body**: `{ "until": 42 }`
**Response 200**: `{ "deleted_until": 42 }`

### `GET /api/v1/webhooks`

Lists active webhooks.

**Response 200**:
```json
{
  "webhooks": [
    { "id": 1, "url": "https://example.com/hook", "events": ["button","knob"], "active": 1 }
  ]
}
```

### `POST /api/v1/webhooks`

Registers a new webhook.

**Request body**:
```json
{
  "url": "https://example.com/hook",
  "secret": "hmac-key",
  "events": ["button"],
  "method": "POST",
  "headers": { "Authorization": "Bearer token" }
}
```

**Response 200**: `{ "id": 2, "url": "https://example.com/hook" }`

### `DELETE /api/v1/webhooks/:id`

Deactivates a webhook.

**Response 200**: `{ "deleted": 2 }`

---

## Database Schema

### `events` table

```sql
CREATE TABLE IF NOT EXISTS events (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  type      TEXT NOT NULL,           -- 'button' | 'knob' | 'touch'
  name      TEXT NOT NULL,           -- 'Button1', 'Knob3', 'Touch5'
  status    TEXT,                    -- 'down' | 'up' (button/touch)
  value     INTEGER,                -- knob delta (+1/-1)
  x         INTEGER,                -- touch x coordinate
  y         INTEGER,                -- touch y coordinate
  timestamp INTEGER NOT NULL        -- Unix epoch seconds
);
CREATE INDEX IF NOT EXISTS idx_events_id ON events(id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
```

### `webhooks` table

```sql
CREATE TABLE IF NOT EXISTS webhooks (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  url     TEXT NOT NULL,
  secret  TEXT DEFAULT '',
  events  TEXT DEFAULT '[]',         -- JSON array of event type filters
  method  TEXT DEFAULT 'POST',
  headers TEXT DEFAULT '{}',         -- JSON object of extra HTTP headers
  active  INTEGER DEFAULT 1          -- soft delete flag
);
```

---

## Risks, Alternatives, and Open Questions

### Risks

1. **gojahttp serialization bottleneck**: All HTTP request handlers run through the goja runtime's serialized owner loop. Under high concurrency (many simultaneous HTTP requests), this could become a bottleneck. Mitigation: the Loupedeck use case is low-throughput (a few events per second at most), so this is unlikely to be a problem in practice.

2. **SQLite write concurrency**: The database module uses `database/sql` with the `sqlite3` driver, which supports one writer at a time. Since all writes come from the single goja owner goroutine, this is safe, but polling clients reading while events are being inserted could see contention. Mitigation: use WAL mode for better read concurrency.

3. **Webhook delivery reliability**: Using `exec("curl", ...)` for webhook delivery is fragile (process spawn overhead, no connection pooling, no retry with backoff). This is acceptable for MVP but must be replaced with a native `fetch` module or a Go-backed delivery goroutine for production.

4. **Missing hardware APIs in `loupedeck/ui`**: The current `loupedeck/ui` module does not expose `setButtonColor` or `setBrightness`. These must be added to the Go module (`loupedeck/runtime/js/module_ui/module.go`) by accessing `env.Host` → `deckConn`.

### Alternatives Considered

1. **Pure Go HTTP server (no JavaScript)**: We could write a standalone Go binary that directly uses `device.Loupedeck` and `net/http` without the JavaScript layer. This would be simpler and faster, but would lose the ability to use the retained UI model (pages/tiles/displays) and the express/gojahttp ecosystem. We chose the xgoja approach because it composes existing modules and allows the server logic to be extended in JavaScript without recompiling.

2. **WebSocket instead of REST**: A WebSocket API would provide real-time event streaming without polling. This is a natural follow-up but adds complexity (connection management, reconnection, multiplexing). The REST + polling approach is simpler to implement and consume from any HTTP client.

3. **Server-Sent Events (SSE)**: SSE provides real-time events over HTTP without the full WebSocket protocol. This is a good middle ground and could be added as a `GET /api/v1/events/stream` endpoint in a follow-up.

### Open Questions

1. **Should `fetch` be a dedicated xgoja module?** Yes — a `fetch` module wrapping Go's `net/http` client would be useful beyond this project. It should be added to `go-go-goja/modules/` with proper async/Promise support.

2. **Should the server support multiple Loupedeck devices?** The current hardware capability connects to one device. Multi-device support would require extending the provider to manage multiple `Loupedeck` instances. This is out of scope for the MVP.

3. **How should the server handle hardware disconnects?** The `device.Loupedeck` serial connection can drop. The server should detect this (the `Listen()` goroutine exits), set `connected: false` in the info endpoint, and attempt reconnection. This is a follow-up.

4. **Should event IDs be globally monotonic across restarts?** SQLite `AUTOINCREMENT` resets if the database is deleted. For persistent cursors across restarts, we could use a separate `event_counter` table with a monotonically increasing value. This is a follow-up.

---

## Key File References

| File | Relevance |
|------|-----------|
| `go-go-goja/pkg/xgoja/providers/http/http.go` | HTTP capability: starts `net/http.Server`, wires express to gojahttp |
| `go-go-goja/pkg/gojahttp/host.go` | `Host.ServeHTTP`: routes HTTP → JS handlers via runtime owner |
| `go-go-goja/pkg/gojahttp/route_registry.go` | Route matching: method + pattern with `:param` support |
| `go-go-goja/pkg/gojahttp/request_response.go` | `RequestDTO`, `Response` objects passed to JS handlers |
| `go-go-goja/modules/express/express.go` | Express module: `app.get/post/put/delete/static` API |
| `go-go-goja/modules/database/database.go` | Database module: `configure`, `query`, `exec`, `close` |
| `go-go-goja/pkg/xgoja/providers/host/host.go` | Guarded host modules (fs, exec, database) with config guards |
| `go-go-goja/pkg/xgoja/providers/core/core.go` | Core modules (path, yaml, crypto, etc.) |
| `loupedeck/pkg/xgoja/provider/provider.go` | Thin wrapper: delegates to `loupedeck/runtime/js/provider` |
| `loupedeck/runtime/js/provider/provider.go` | Loupedeck provider: registers 8 modules + hardware capability + scenes command set |
| `loupedeck/runtime/js/module_ui/module.go` | UI module: `page`, `tile`, `display`, `onButton`, `onKnob`, `onTouch`, `show`, `invalidate` |
| `loupedeck/runtime/js/env/env.go` | `LoupeDeckEnvironment`: central state holder (UI, Host, Anim, Present, Metrics) |
| `loupedeck/runtime/host/runtime.go` | Host runtime: event source binding, Attach/Close lifecycle |
| `loupedeck/runtime/host/events.go` | OnButton/OnKnob/OnTouch subscription management |
| `loupedeck/pkg/device/connect.go` | Hardware connection: auto-detect serial, WebSocket handshake |
| `loupedeck/pkg/device/loupedeck.go` | Core `Loupedeck` struct: SetBrightness, SetButtonColor, Close |
| `loupedeck/pkg/device/display.go` | Display drawing: RGB565 framebuffer, WriteFramebuff+Draw commands |
| `loupedeck/pkg/device/inputs.go` | Input constants: Button, Knob, TouchButton, ParseButton/ParseKnob |
| `loupedeck/pkg/device/listeners.go` | Event listener API: OnButton/OnKnob/OnTouch with Subscription |
| `loupedeck/pkg/device/profile.go` | Device profiles: display dimensions per model |
| `go-go-goja/pkg/xgoja/providerapi/` | Provider API: Registry, Module, Capability interfaces |
| `go-go-goja/engine/` | Runtime engine: owner loop, module loading, lifecycle |

---

## How It All Fits Together: Request Lifecycle

When a coding agent calls `POST /api/v1/displays/main/draw` with a grid of tiles, here is the exact sequence:

```
1. Agent sends HTTP POST to loupedeck-server:9876

2. Go net/http.Server receives the request
   → gojahttp.Host.ServeHTTP(w, r)

3. Host matches route: POST /api/v1/displays/:name/draw
   → found the handler registered by app.post("/api/v1/displays/:name/draw", ...)

4. Host creates RequestDTO from the HTTP request
   → { method: "POST", path: "/api/v1/displays/main/draw",
       params: { name: "main" }, body: { layout: "grid", tiles: [...] } }

5. Host calls owner.Call("http-handler", fn)
   → enters goja runtime on the serialized owner goroutine

6. JavaScript handler executes:
   const body = jsonBody(req)   // parses req.body
   ui.page(pageName, (p) => {  // creates a retained UI page
     for (const t of body.tiles) {
       p.tile(t.col, t.row, (tile) => {
         tile.text(t.text)      // sets tile text binding
         tile.icon(t.icon)      // sets tile icon binding
       })
     }
   })
   ui.show(pageName)           // makes the page active

7. Present flush loop (running on a separate goroutine):
   → detects dirty page (ui.invalidate was called internally)
   → calls renderer.Flush()
   → iterates dirty display regions
   → for each region: renders retained UI tree to RGB565 image
   → calls display.Draw(image, xoff, yoff)
   → outbound writer sends WriteFramebuff + Draw messages over serial
   → Loupedeck hardware updates its display

8. Response returns to agent:
   res.json({ ok: true, page: pageName })

9. Agent receives: { "ok": true, "page": "_draw_1717100000" }
```

When a button is pressed on the Loupedeck:

```
1. Loupedeck hardware sends button event over serial

2. device.Loupedeck.Listen() goroutine reads the message
   → dispatches to registered button listener

3. host.Runtime.OnButton callback fires:
   → runtimeServices.PostWithLifetimeContext("ui.onButton.callback", fn)
   → schedules fn on the goja owner goroutine

4. JavaScript callback executes:
   ui.onButton("Button1", (event) => {
     recordEvent({ type: "button", name: "Button1", status: event.status })
   })

5. recordEvent inserts into SQLite:
   db.exec("INSERT INTO events (type, name, status, value, x, y, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)", ...)

6. recordEvent dispatches webhooks:
   - Queries active webhooks from SQLite
   - Filters by event type
   - POSTs event payload to each webhook URL (via exec/fetch)

7. Agent polls events:
   GET /api/v1/events?since=0
   → db.query("SELECT * FROM events WHERE id > 0 ORDER BY id ASC LIMIT 100")
   → returns { cursor: 42, events: [...] }
```

---

## Testing Strategy

### Unit Tests

1. **Provider validation**: Run `xgoja doctor -f xgoja.yaml` to catch spec errors.
2. **Module smoke test**: Start the binary with `eval` and call each `require()` module:
   ```bash
   ./dist/loupedeck-server eval 'const express = require("express"); console.log(typeof express.app)'
   ./dist/loupedeck-server eval 'const ui = require("loupedeck/ui"); console.log(typeof ui.page)'
   ./dist/loupedeck-server eval 'const db = require("db"); db.configure("sqlite3", ":memory:"); console.log("ok")'
   ```

### Integration Tests

1. **Mock mode test**: Start server without hardware (`--deck-enabled=false`), make HTTP requests, verify responses.
2. **Hardware test**: Start with hardware, press a button, poll events, verify the event appears.
3. **Webhook test**: Register a webhook, press a button, verify the webhook URL was called.

### Manual Validation

```bash
# Start server
./dist/loupedeck-server serve server.js --deck-enabled --http-listen :9876

# In another terminal:
# Device info
curl http://localhost:9876/api/v1/info

# Set brightness
curl -X PUT http://localhost:9876/api/v1/brightness -d '{"value":5}'

# Create a page
curl -X POST http://localhost:9876/api/v1/pages \
  -d '{"name":"test","tiles":[{"col":0,"row":0,"text":"Hello"}]}'

# Show the page
curl -X POST http://localhost:9876/api/v1/pages/show -d '{"name":"test"}'

# Poll events (press a button first)
curl 'http://localhost:9876/api/v1/events?since=0'

# Register a webhook
curl -X POST http://localhost:9876/api/v1/webhooks \
  -d '{"url":"https://webhook.site/abc123","events":["button"]}'
```
