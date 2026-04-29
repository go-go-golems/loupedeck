# Changelog

## 2026-04-22

- Initial workspace created


## 2026-04-22

Fix: purge stale serial buffer after port open to prevent 'malformed HTTP response' and 'bad opcode' websocket errors

### Related Files

- /home/manuel/workspaces/2026-04-22/fix-loupedeck-serial/loupedeck/pkg/device/dialer.go — Added ResetInputBuffer() calls after serial.Open in ConnectSerialAuto and ConnectSerialPath

