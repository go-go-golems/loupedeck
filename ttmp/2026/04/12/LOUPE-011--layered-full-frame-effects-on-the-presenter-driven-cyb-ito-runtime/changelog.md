# Changelog

## 2026-04-12

- Created ticket `LOUPE-011` for layered full-frame effects on the presenter-driven cyb-ito runtime.
- Added the main implementation plan describing the correct software-layered / single-full-frame-flush architecture.
- Added the implementation diary.
- Added the first detailed phased task breakdown.
- Implemented the first layered full-page code slice in `/home/manuel/code/wesen/2026-04-11--loupedeck-test/examples/js/10-cyb-ito-full-page-all12.js` (commit `4b44402` — `Add layered full-frame compositor to cyb-ito scene`).
- Split the scene into internal software layers (`baseLayer`, `chromeLayer`, `sceneLayer`, `fxLayer`, `hudLayer`, final `frame`) while keeping one presenter-driven full-page display surface.
- Added the first FX pass: scanlines, sparse grain/noise, active sweep, and active ripple overlays.
- Validated the slice with `go test ./...` and a hardware smoke run using `--send-interval 0ms`.
- Archived reproducibility scripts under `scripts/`.
