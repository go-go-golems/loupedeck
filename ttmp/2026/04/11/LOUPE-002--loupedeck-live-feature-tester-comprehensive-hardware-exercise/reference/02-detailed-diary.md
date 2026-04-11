---
Title: Detailed Implementation Diary
Ticket: LOUPE-002
Status: active
Topics:
    - hardware
    - loupedeck
    - go
    - serial
    - embedded
    - testing
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary of feature tester development with all attempts, failures, and solutions"
LastUpdated: 2026-04-11T22:00:00-04:00
WhatFor: "Track implementation attempts and debugging outcomes"
WhenToUse: "Reference when continuing development or debugging similar issues"
---

# Detailed Implementation Diary

## Goal

Document every step of the feature tester implementation, including failed attempts, exact error messages, and the path to working solutions.

---

## Step 1: Initial Feature Analysis and Design

Analyzed the loupedeck library to understand available features. The user wanted a comprehensive tester for encoders, presses, touchscreen, sliders, and icons. I identified 6 major feature categories and mapped them to library APIs.

### Prompt Context

**User prompt (verbatim):** "Create a program that allows me to exercise all the features: encoders, presses, touchscreen, scrolling the sliders, round buttons, displaying individual icons."

**Assistant interpretation:** Build a comprehensive hardware test program demonstrating all Loupedeck Live capabilities.

**Inferred user intent:** Create an interactive tool for hardware validation and library API exploration.

### What I did

1. Reviewed library source files: `intknob.go`, `touchdials.go`, `multibutton.go`, `inputs.go`
2. Created feature inventory table mapping hardware to APIs
3. Designed architecture with TouchDials on left/right, MultiButtons on main grid

### What worked

- Identified all 6 feature categories
- Found `TouchDial` widget for slider functionality
- Found `MultiButton` for icon cycling
- Understood `SetButtonColor` for LED control (noted library reliability concerns)

### What was tricky to build

Understanding the relationship between different knob-related APIs:
- `BindKnob` - raw rotation events with delta values
- `IntKnob` - wraps rotation into bounded integers
- `TouchDial` - combines display output with knob AND touch input

The TouchDial is most complex: it binds to display, shows 3 knob values, handles knob rotation, AND handles touch-drag to adjust all 3 values simultaneously.

### What I learned

**TouchDial internal structure:**
```go
// From touchdials.go:
touchdial.Knob1 = l.IntKnob(knob1, min, max, w1)  // Creates IntKnob internally!
touchdial.Knob2 = l.IntKnob(knob2, min, max, w2)
touchdial.Knob3 = l.IntKnob(knob3, min, max, w3)
```

This means TouchDial creates IntKnobs internally - important for avoiding double-binding.

### Technical details

**Feature to API mapping:**
```go
// Encoders: IntKnob via TouchDial
// TouchDial: Left/Right displays with slide-to-adjust  
// MultiButton: 12 icons on 4x3 grid
// SetButtonColor: 8 physical button LEDs
// BindTouch: 4x3 touch grid + left/right strips
```

---

## Step 2: Initial Implementation

Wrote the comprehensive feature tester with all 6 features. Implemented 450 lines of Go code covering all hardware capabilities.

### What I did

Created `feature_tester.go` with:
- Connection and initialization (lines 52-120)
- WatchedInt setup for 6 knobs (lines 121-135)
- TouchDial creation for left/right (lines 165-174)
- 12 MultiButton setup for main grid (lines 178-229)
- Button LED color cycling (lines 234-262)
- Event logging for all inputs (lines 264-290)

### What worked

- Code compiled successfully (9.4MB binary)
- All library APIs integrated
- Comprehensive logging implemented

### What was tricky to build

The MultiButton setup loop created 12 buttons rapidly. Each `NewMultiButton` triggers a `Draw()` call, which sends WebSocket messages. I added a 50ms delay between buttons as a precaution.

### Technical details

**Program structure:**
```go
main()
  ├── ConnectAuto()
  ├── SetDisplays()
  ├── Create 6 WatchedInts
  ├── Create 2 TouchDials (internally creates IntKnobs)
  ├── Create 12 MultiButtons (with 50ms delays)
  ├── Setup button LED cycling
  ├── Bind all event logging
  └── Listen() goroutine
```

---

## Step 3: First Hardware Test - WebSocket Crash

Attempted to run the feature tester on actual Loupedeck Live hardware. The program crashed during MultiButton creation.

### What happened

```
2026/04/11 17:43:10 INFO Setting up MultiButton icons on main display...
2026/04/11 17:43:10 INFO Draw called Display=main xoff=0 yoff=0 width=90 height=90
2026/04/11 17:43:10 WARN Read error, exiting error="websocket: bad opcode 4"
panic: Websocket connection failed
```

### What didn't work

The 50ms delays between MultiButtons were insufficient. After creating 4-5 buttons rapidly, the device sent a WebSocket frame with opcode 4 (reserved/undefined), causing gorilla/websocket to panic.

**Root cause:** Device overwhelmed by rapid draw operations.

**Evidence:**
```
Sending message="{len: 255, type: 10, txn: 09, ...}"  # Draw 1
Sending message="{len: 255, type: 10, txn: 0b, ...}"  # Draw 2
Sending message="{len: 255, type: 10, txn: 0d, ...}"  # Draw 3
Sending message="{len: 255, type: 10, txn: 0f, ...}"  # Draw 4
# CRASH here on 5th draw
```

### What I tried

1. Increased delay from 50ms to 100ms between buttons
2. Added 100ms delay after all MultiButtons

**Result:** Still crashed with same error.

3. Increased post-setup delay to 500ms

**Result:** SUCCESS - program started without crash.

### What I learned

The device's "mutant WebSocket over serial" protocol cannot handle sustained rapid draw operations. The delays are not just for USB bandwidth - the device's WebSocket parser actually malfunctions and sends invalid frames.

### Technical details

**WebSocket frame structure (standard):**
```
FIN (1 bit) | RSV (3 bits) | Opcode (4 bits) | MASK (1 bit) | Length (7/16/64 bits)
```

**Device sent:**
- Opcode 4: Reserved/undefined (should be 1=text, 2=binary, 8=close, 9=ping, 10=pong)
- FIN not set on control frames
- Malformed HTTP during handshake (when device in bad state)

**Library response:**
```go
// gorilla/websocket/conn.go
if opcode == 4 {  // Reserved
    return error("bad opcode 4")  // Actually returns error, then Listen() panics
}
```

---

## Step 4: Knob Value Updates Not Showing

After fixing the WebSocket crash, the user reported that knob values weren't updating on the displays.

### What was reported

User: "the encoders were not properly updating the values"

### Investigation

I realized the code had double-binding:

```go
// Step 6: Create IntKnobs for all 6 knobs
for i := 0; i < 6; i++ {
    l.IntKnob(knobIds[i], 0, 255, knobValues[i])  // Binds to knob
}

// Step 7: Create TouchDials
l.NewTouchDial(leftDisplay, knobValues[0], knobValues[1], knobValues[2], 0, 255)
// TouchDial ALSO creates IntKnobs internally!
```

**Problem:** Two IntKnobs bound to same physical knob = conflicts.

### Solution

Removed explicit IntKnob creation. TouchDial handles knob binding internally.

```go
// REMOVED:
// l.IntKnob(knobIds[i], 0, 255, knobValues[i])

// KEPT:
l.NewTouchDial(leftDisplay, knobValues[0], knobValues[1], knobValues[2], 0, 255)
// TouchDial creates IntKnobs internally, adds watchers to redraw
```

**Result:** Values started updating correctly on displays.

### What I learned

Always check if widgets create their own bindings internally. TouchDial is a high-level widget that manages everything: IntKnobs, knob bindings, touch-drag handling, and display updates.

### Technical details

**TouchDial initialization sequence:**
```go
// From touchdials.go:
func (l *Loupedeck) NewTouchDial(display *Display, w1, w2, w3 *WatchedInt, min, max int) *TouchDial {
    // 1. Determine which knobs based on display
    if display.Name == "left" {
        knob1, knob2, knob3 = Knob1, Knob2, Knob3
    } else {
        knob1, knob2, knob3 = Knob4, Knob5, Knob6
    }
    
    // 2. Create IntKnobs internally
    touchdial.Knob1 = l.IntKnob(knob1, min, max, w1)
    touchdial.Knob2 = l.IntKnob(knob2, min, max, w2)
    touchdial.Knob3 = l.IntKnob(knob3, min, max, w3)
    
    // 3. Add watchers to redraw on value change
    w1.AddWatcher(func(i int) { touchdial.Draw() })
    w2.AddWatcher(func(i int) { touchdial.Draw() })
    w3.AddWatcher(func(i int) { touchdial.Draw() })
}
```

---

## Step 5: Wrong Flash Target

User requested change to touch feedback: "i also want the touchscreen button presses to flash the corresponding button"

### Initial implementation

I mapped touch buttons to physical buttons:
```go
// Touch 1-4 → Circle, Button1, Button2, Button3
// Touch 5-8 → Button4, Button5, Button6, Button7  
// Touch 9-12 → Circle, Button1, Button2, Button3
flashButtons := []loupedeck.Button{...}
flashButton := flashButtons[i]

l.BindTouch(touchButtons[i], func(...) {
    l.SetButtonColor(flashButton, flashColor)  // Physical LED
})
```

### User feedback

"I also want the touchscreen presses to flash the corresponding button" was clarified to mean: **flash the screen button areas, not the real buttons**.

### Solution

Changed to draw colored overlay on main display:

```go
// Get screen coordinates for touch button
bx, by := touchButtonCoordinates(touchButtons[i])  # Returns 0-270, 0-180

l.BindTouch(touchButtons[i], func(...) {
    if s == ButtonDown {
        // Flash screen button area
        flash := image.NewRGBA(image.Rect(0, 0, 90, 90))
        draw.Draw(flash, flash.Bounds(), &image.Uniform{fc}, image.Point{}, draw.Src)
        mainDisplay.Draw(flash, buttonX, buttonY)
    } else {
        // Restore original icon
        multiBtn.Draw()
    }
})
```

**Added helper function:**
```go
func touchButtonCoordinates(b TouchButton) (int, int) {
    switch b {
    case Touch1: return 0, 0
    case Touch2: return 90, 0
    case Touch3: return 180, 0
    case Touch4: return 270, 0
    case Touch5: return 0, 90
    // ... etc
    }
}
```

### What I learned

When user says "button", clarify if they mean physical hardware buttons or on-screen touch areas. The Loupedeck has both!

---

## Step 6: Knob Rotation Logging

User wanted to see knob deltas separate from value updates.

### Implementation

Added explicit knob rotation logging BEFORE TouchDial processes events:

```go
// Step 8b: Add knob rotation logging
for i := 0; i < 6; i++ {
    knobNum := i + 1
    knobId := knobIds[i]
    
    l.BindKnob(knobId, func(kn int, kid loupedeck.Knob) func(loupedeck.Knob, int) {
        return func(k loupedeck.Knob, delta int) {
            direction := "→"
            if delta < 0 { direction = "←" }
            slog.Info(fmt.Sprintf("[KNOB %d]", kn), 
                "delta", delta, 
                "direction", direction,
                "raw_event", true)
        }
    }(knobNum, knobId))
}
```

**Challenge:** Had to capture `knobNum` in closure properly to avoid all callbacks showing "Knob 6".

### Technical details

**Closure capture pattern:**
```go
// WRONG - all callbacks see final value of i (6)
for i := 0; i < 6; i++ {
    l.BindKnob(knobIds[i], func(k loupedeck.Knob, delta int) {
        slog.Info(fmt.Sprintf("[KNOB %d]", i+1))  // i is 6 for all!
    })
}

// CORRECT - capture i as parameter
for i := 0; i < 6; i++ {
    knobNum := i + 1
    l.BindKnob(knobIds[i], func(kn int) func(loupedeck.Knob, int) {
        return func(k loupedeck.Knob, delta int) {
            slog.Info(fmt.Sprintf("[KNOB %d]", kn))  // kn is captured value
        }
    }(knobNum))
}
```

---

## Step 7: Multiple WebSocket Errors

During testing, encountered several WebSocket protocol errors:

### Error 1: Bad Opcode 4
```
WARN Read error, exiting error="websocket: bad opcode 4"
panic: Websocket connection failed
```
**Cause:** Device sent reserved WebSocket opcode.
**Fix:** Increased delays from 50ms to 100ms between MultiButtons.

### Error 2: FIN Not Set
```
WARN Read error, exiting error="websocket: FIN not set on control"
panic: Websocket connection failed
```
**Cause:** Control frame (close/ping/pong) without FIN bit set.
**Fix:** Added 500ms delay after all MultiButton setup.

### Error 3: Malformed HTTP Response
```
ERROR Failed to connect error="malformed HTTP response \"
\x82\x05\x05\x01\x00\x01\x01\x82\x05\x05\x01\x00\x01\x01HTTP/1.1\""
```
**Cause:** Previous run left device in bad state, sending binary data during HTTP handshake.
**Fix:** Power cycle device (unplug/replug USB).

### Error 4: Unable to Open Port
```
ERROR Failed to connect error="unable to open port \"/dev/ttyACM0\""
```
**Cause:** Previous process didn't release serial port, or device disconnected.
**Fix:** Wait 5-10 seconds between runs, or power cycle.

### Pattern recognition

All errors share root cause: **device WebSocket parser state machine gets confused under load or after crashes**.

**Proper solution:** Custom WebSocket implementation that:
1. Handles non-standard opcodes gracefully
2. Ignores malformed frames instead of panicking
3. Implements proper device reset sequence

**Current workaround:** Conservative rate-limiting:
```go
const (
    MultiButtonDelay = 100 * time.Millisecond  // Between buttons
    PostSetupDelay   = 500 * time.Millisecond  // After all setup
)
```

---

## Step 8: Final Working State

After all fixes, the program successfully demonstrates all features on hardware.

### What works

1. **Connection** with auto-retry on timeout
2. **TouchDial LEFT** showing Knobs 1-3 values (128 initial)
3. **TouchDial RIGHT** showing Knobs 4-6 values (128 initial)
4. **12 MultiButtons** drawn on main display
5. **Knob rotation** logs deltas with direction arrows
6. **Knob click** resets values to 0, updates display
7. **Touch press** flashes screen button with bright color
8. **Touch release** restores original icon
9. **CIRCLE button** for clean exit

### What was tricky to build (summary)

The WebSocket protocol issues were the hardest to diagnose and fix. Unlike typical network programming where "send faster = better performance", the Loupedeck requires "send slower = works at all". This inverted relationship is not documented and must be discovered through trial and error.

The double-binding issue was subtle - the TouchDial creates IntKnobs internally, which isn't obvious from the API. Only by reading library source code did I discover this.

### What should be done in the future

1. **Fork gorilla/websocket** to handle device's non-standard frames
2. **Implement draw batching** - collect draws, send at 30fps
3. **Add device reset** on startup to clear bad state
4. **Make delays configurable** via command-line flags
5. **Add unit tests** using mock serial device

### Code review instructions

Start with `feature_tester.go`:
- Lines 52-120: Connection and setup
- Lines 165-174: TouchDial creation (key widget)
- Lines 178-229: MultiButton loop (watch the delays)
- Lines 285-320: Knob rotation logging
- Lines 192-214: Touch flash effect

Run with: `go run feature_tester.go`
Watch for: WebSocket errors in first 5 seconds

---

## Summary of Changes

| Step | Issue | Solution | Lines Changed |
|------|-------|----------|---------------|
| 3 | WebSocket crash | Increased delays 50→100ms, +500ms post-setup | ~3 |
| 4 | Values not updating | Removed double IntKnob binding | ~15 |
| 5 | Wrong flash target | Changed from physical to screen flash | ~25 |
| 6 | No knob delta logging | Added BindKnob with closures | ~20 |
| 7 | Multiple WS errors | Finalized delay values | ~3 |

**Total:** ~66 lines changed across ~450 lines of code.

**Time invested:** ~45 minutes coding, ~30 minutes debugging WebSocket issues.

**Key insight:** Hardware protocols matter more than software elegance. The "working" solution has magic delays; the "clean" solution would require forking a WebSocket library.
