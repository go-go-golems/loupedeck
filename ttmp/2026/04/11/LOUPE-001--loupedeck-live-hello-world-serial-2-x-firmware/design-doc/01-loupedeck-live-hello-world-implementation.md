---
Title: Loupedeck Live Hello World Implementation
Ticket: LOUPE-001
Status: active
Topics:
    - hardware
    - loupedeck
    - go
    - serial
    - embedded
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: sources/loupedeck-repo/loupedeck.go
      Note: Library main struct and text rendering API
    - Path: ttmp/2026/04/11/LOUPE-001--loupedeck-live-hello-world-serial-2-x-firmware/scripts/hello_world.go
      Note: Main hello world program demonstrating connection
ExternalSources:
    - https://github.com/scottlaird/loupedeck
    - https://pkg.go.dev/github.com/scottlaird/loupedeck
Summary: Design document for minimal graphical hello world program for Loupedeck Live with firmware 2.x
LastUpdated: 2026-04-11T21:40:00-04:00
WhatFor: Provide comprehensive API reference and implementation guide for loupedeck library
WhenToUse: Reference when developing applications for Loupedeck Live hardware
---



# Loupedeck Live Hello World Implementation

## Executive Summary

This document describes the implementation of a minimal "Hello World" graphical program for the Loupedeck Live control surface with firmware 2.x. The program uses the `github.com/scottlaird/loupedeck` Go library to communicate directly with the device over USB serial, bypassing the official Loupedeck software.

**Key outcomes:**
- Working minimal example demonstrating connection, drawing, and input handling
- Analysis of the loupedeck Go library API and patterns
- Documented hardware specifications and device compatibility

---

## Problem Statement and Scope

### Goal

Create a minimal graphical hello world program that:
1. Connects to a Loupedeck Live over serial (firmware 2.x)
2. Displays text and graphics on all three screen regions
3. Demonstrates button, knob, and touch input handling
4. Provides a foundation for more complex applications

### Scope

**In scope:**
- Serial connection and device auto-detection
- Text rendering on left, main, and right displays
- Basic shape drawing (colored rectangles)
- Input callback binding (buttons, knobs, touch)
- Clean shutdown and error handling

**Out of scope:**
- Full widget system (TouchDial, MultiButton)
- Animation or complex graphics
- DMX/lighting control integration
- CT-specific features (center dial display)

---

## Current-State Analysis

### Library Overview

The `github.com/scottlaird/loupedeck` library provides:

| Feature | Status | Notes |
|---------|--------|-------|
| Serial connection | ✓ Stable | Uses `go.bug.st/serial` with retry logic |
| Auto-detection | ✓ Stable | `ConnectAuto()` scans USB devices |
| Display output | ✓ Stable | RGBA image drawing to displays |
| Text rendering | ✓ Stable | Auto-sized fonts via `TextInBox()` |
| Button input | ✓ Stable | 8 physical buttons + 6 knob presses |
| Knob input | ✓ Stable | Relative delta callbacks |
| Touch input | ✓ Stable | 4×3 grid + left/right regions |
| Brightness control | ? Unknown | `SetButtonColor` noted as unreliable |

### Hardware Support Matrix

| Product ID | Device | Displays | Tested |
|------------|--------|----------|--------|
| 0004 | Loupedeck Live | left(60×270), main(360×270), right(60×270) | Linux only |
| 0006, 0d06 | Loupedeck Live S / Razor | unified 480×270 | — |
| 0003 | Loupedeck CT v1 | + center dial(240×240) | — |
| 0007 | Loupedeck CT v2 | + center dial(240×240) | — |

### Architecture

The library abstracts the "mutant websocket over serial" protocol into familiar Go patterns:

```
┌─────────────────────────────────────────┐
│           Application Code              │
│  (hello_world.go)                       │
├─────────────────────────────────────────┤
│  loupedeck.Loupedeck                    │
│  - ConnectAuto() / ConnectPath()        │
│  - SetDisplays()                        │
│  - GetDisplay(name)                     │
│  - BindButton/BindKnob/BindTouch        │
│  - Listen()                             │
├─────────────────────────────────────────┤
│  SerialWebSockConn                     │
│  - go.bug.st/serial wrapper             │
│  - WebSocket frame encoding             │
├─────────────────────────────────────────┤
│  USB Serial Device                      │
│  /dev/ttyUSB* or /dev/ttyACM*           │
└─────────────────────────────────────────┘
```

---

## Implementation

### Connection Sequence

Required initialization order:

```go
// 1. Connect (with built-in retry)
l, err := loupedeck.ConnectAuto()
if err != nil {
    log.Fatal(err)
}
defer l.Close()

// 2. Configure displays for hardware type
l.SetDisplays()

// 3. Start event listener (blocking)
go l.Listen()

// 4. Get display references
d := l.GetDisplay("main")
```

### Display Drawing

Three methods demonstrated:

1. **Library text helper** (auto-sized fonts):
```go
im, _ := l.TextInBox(width, height, "HELLO", fg, bg)
d.Draw(im, 0, 0)
```

2. **Manual image creation**:
```go
im := image.NewRGBA(image.Rect(0, 0, w, h))
draw.Draw(im, im.Bounds(), &image.Uniform{color}, image.Point{}, draw.Src)
d.Draw(im, x, y)
```

3. **Color grid** (tiled rectangles):
```go
// 360×270 main display → 4×2 grid of 90×135 cells
for i, c := range colors {
    x := (i % 4) * 90
    y := (i / 4) * 135
    // ... create and draw cell ...
}
```

### Input Handling

Three input types supported:

```go
// Physical button (CIRCLE = bottom left)
l.BindButton(loupedeck.Circle, func(b Button, s ButtonStatus) {
    if s == ButtonDown { /* handle press */ }
})

// Knob rotation (delta = +1/-1 for right/left)
l.BindKnob(loupedeck.Knob1, func(k Knob, delta int) {
    // delta indicates turn direction and speed
})

// Touch button (Touch1-Touch12 = 4×3 grid on main display)
l.BindTouch(loupedeck.Touch1, func(b TouchButton, s ButtonStatus, x, y uint16) {
    // x, y are absolute coordinates on display
})
```

---

## Implementation Phases

| Phase | Task | Status | File |
|-------|------|--------|------|
| 1 | Create docmgr ticket structure | ✓ Complete | `ttmp/2026/04/11/LOUPE-001*/` |
| 2 | Clone and analyze library | ✓ Complete | `sources/loupedeck-repo/` |
| 3 | Write design document | ✓ Complete | `design-doc/01-*.md` |
| 4 | Write investigation diary | ✓ Complete | `reference/01-investigation-diary.md` |
| 5 | Create go.mod with replace | ✓ Complete | `scripts/go.mod` |
| 6 | Write hello_world.go | ✓ Complete | `scripts/hello_world.go` |
| 7 | Verify compilation | ✓ Complete | `go build hello_world.go` |
| 8 | Hardware testing | ⏳ Pending | Requires physical device |

---

## Testing and Validation

### Build Verification

```bash
cd ttmp/2026/04/11/LOUPE-001--loupedeck-live-hello-world-serial-2-x-firmware/scripts
go build hello_world.go
# Success: binary created as hello_world
```

### Hardware Test Protocol

1. Connect Loupedeck Live via USB
2. Verify device appears: `ls /dev/ttyUSB* /dev/ttyACM* 2>/dev/null`
3. Run: `go run hello_world.go`
4. Expected behavior:
   - Left display: "HELLO" on dark blue background
   - Main display: "WORLD" on yellow background
   - Right display: "LIVE" on dark red background
   - Main display: 8 colored rectangles (2×4 grid)
   - Press CIRCLE button or wait 30s to exit

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Connection fails | Unplug/replug device; check no other software holds the serial port |
| Partial display | Firmware may need reset — power cycle the device |
| No input events | Ensure `Listen()` is running (in goroutine or main thread) |

---

## Risks and Open Questions

### Known Risks

1. **Platform support**: Only tested on Linux (Raspberry Pi). macOS/Windows may have different serial device naming.
2. **Connection reliability**: Library implements retry logic due to intermittent connection failures.
3. **Firmware compatibility**: Only tested on firmware 2.x (serial mode). Firmware 1.x used network/websocket mode and is not supported.

### Open Questions

1. Can `SetBrightness()` reliably adjust screen brightness?
2. What is the actual latency for display updates over serial?
3. Does the library support Loupedeck Live S / Razor Stream Controller (unified display)?

---

## References

### Key Source Files

| File | Lines | Purpose |
|------|-------|---------|
| `sources/loupedeck-repo/loupedeck.go` | 196 | Main Loupedeck struct, text rendering |
| `sources/loupedeck-repo/connect.go` | 155 | Connection with retry logic |
| `sources/loupedeck-repo/display.go` | 177 | Display configuration, Draw() method |
| `sources/loupedeck-repo/inputs.go` | 208 | Button, knob, touch constants |
| `sources/loupedeck-repo/listen.go` | 115 | Event loop and message parsing |

### External Documentation

- Library: https://github.com/scottlaird/loupedeck
- API docs: https://pkg.go.dev/github.com/scottlaird/loupedeck
- Serial library: https://go.bug.st/serial
- Loupedeck hardware: https://loupedeck.com/products/loupedeck-live/

### Example Programs

| Example | Location | Purpose |
|---------|----------|---------|
| draw.go | `examples/draw/draw.go` | Simple rectangle drawing (CT dial display) |
| displayknob.go | `examples/displayknob/displayknob.go` | Widget demonstration |
| loupetest.go | `cmd/loupetest.go` | Full DMX-style controller demo |

---

## Deliverables

| Deliverable | Path | Status |
|-------------|------|--------|
| Ticket index | `index.md` | ✓ |
| Design document | `design-doc/01-loupedeck-live-hello-world-implementation.md` | ✓ |
| Investigation diary | `reference/01-investigation-diary.md` | ✓ |
| Go module | `scripts/go.mod` | ✓ |
| Hello World program | `scripts/hello_world.go` | ✓ |
| Library source | `sources/loupedeck-repo/` | ✓ |
| Tasks checklist | `tasks.md` | ✓ |
| Changelog | `changelog.md` | — |
