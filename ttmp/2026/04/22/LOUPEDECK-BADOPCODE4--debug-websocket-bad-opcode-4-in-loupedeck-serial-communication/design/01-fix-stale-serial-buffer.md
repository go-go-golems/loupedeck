---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: pkg/device/dialer.go
      Note: ResetInputBuffer calls added after serial.Open to clear stale websocket frames
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Fix: Stale Serial Buffer Causing Websocket Framing Errors

## Problem

The Loupedeck software intermittently failed with two related errors:

1. **"malformed HTTP response"** during websocket handshake — the device sent binary websocket frames (`\x82...`) instead of an HTTP 101 response
2. **"websocket: bad opcode 4"** during normal operation — the websocket frame parser read an invalid opcode

Both errors were symptoms of the same root cause.

## Root Cause

The Loupedeck uses **websocket framing over USB serial**. The gorilla/websocket library communicates through a custom `net.Conn` implementation (`SerialWebSockConn`) that wraps a `go.bug.st/serial.Port`.

**USB serial devices have no disconnect signal.** When the client closes the serial port (timeout, crash, normal exit), the Loupedeck device cannot detect this. It continues sending websocket-framed protocol data into its TX buffer.

When the client reopens the serial port:
- The OS serial driver's read buffer contains **stale websocket binary frames** from the previous session
- The new connection starts a websocket handshake, expecting `HTTP/1.1 101 Switching Protocols`
- It instead reads a binary frame starting with `\x82` (FIN=1, opcode=2), causing "malformed HTTP response"
- If partial frames are buffered, the parser gets out of sync and interprets payload bytes as frame headers, causing "bad opcode 4"

## Evidence

Past Pi transcripts (session `658a1b75-c2ef-4693-8e5c-e02f4c344288`) show this error occurring consistently across multiple days:

```
malformed HTTP response "\x82\x05\x05\x00\x00\a\x01"
malformed HTTP response "\x82\t\tM\x00\x00\x01-\x00\xa1"
malformed HTTP version "\x82\x05\x05\x01\x00\x01\x01HTTP/1.1"
```

The `\x82` byte is a valid websocket BinaryMessage frame header. The device was correctly using websocket framing — but the client was in handshake mode, expecting HTTP.

## Solution

Call `ResetInputBuffer()` on the serial port immediately after opening it. This purges the OS driver's read buffer of any stale data from previous sessions.

### Code Change

**File:** `pkg/device/dialer.go`

**Location 1:** `ConnectSerialAuto()`
```go
p, err := serial.Open(port.Name, &serial.Mode{})
if err != nil {
    return nil, fmt.Errorf("unable to open port %q", port.Name)
}
// Purge any stale data from previous sessions.
if err := p.ResetInputBuffer(); err != nil {
    slog.Warn("Unable to reset serial input buffer", "port", port.Name, "err", err)
}
```

**Location 2:** `ConnectSerialPath()`
```go
p, err := serial.Open(serialPath, &serial.Mode{})
if err != nil {
    return nil, fmt.Errorf("unable to open serial device %q", serialPath)
}
// Purge any stale data from previous sessions.
if err := p.ResetInputBuffer(); err != nil {
    slog.Warn("Unable to reset serial input buffer", "path", serialPath, "err", err)
}
```

## Why This Works

- `go.bug.st/serial.Port` provides `ResetInputBuffer()` which discards all data in the driver's receive buffer
- This clears stale websocket frames from previous connections before the new handshake begins
- The device responds to the new HTTP upgrade request with a clean 101 response
- If `ResetInputBuffer()` fails, we log a warning and continue — the connection may still succeed

## Validation

Before fix: "malformed HTTP response" occurred on ~50% of connection attempts (based on transcript analysis).

After fix: 5 consecutive runs with zero errors:

```bash
for i in $(seq 1 5); do
  GOWORK=off go run ./cmd/loupedeck verbs counter-button run \
    --exit-on-circle=false --duration 2s 2>&1 | grep -E "malformed|bad opcode"
done
# No output — all connections succeeded
```

## Future Work

1. **Explicit serial mode configuration** — the current `&serial.Mode{}` uses all-zero defaults. Verify and set the Loupedeck's expected baud rate explicitly.
2. **Investigate the consistent 2-second timeout** — the first handshake attempt always times out. The device may need a specific post-open delay or initialization sequence.
3. **Connection-state metrics** — track handshake success/failure rates, retry counts, and stale-buffer hits.
4. **Cross-platform validation** — verify `ResetInputBuffer()` behavior is consistent on Windows and macOS.
