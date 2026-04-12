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
- Added color-tinted display-layer support for overlays in the runtime (`75a3c02` — `Add red accent layer and touch spiral ripple`).
- Updated the full-page scene so the selected tile is highlighted through a red accent layer rather than forcing the entire base frame into a color path.
- Added a large touch-triggered spiral ripple overlay that spans the whole screen while preserving the presenter-driven single-frame flush model.
- Added an interactive ticket-local run script for user verification: `scripts/05-run-red-ripple-scene-interactive.sh`.
- Tuned the selected-tile accent so the actual mini-app/art remains visible while being tinted red instead of being partially covered by a red overlay (`a78d513` — `Tune red accent rendering and fullscreen touch ripple`).
- Tuned the touch ripple so it uses the touch location and reads as a true fullscreen spiral/ring effect rather than a mostly tile-local accent.
- Corrected the touch-to-main-display coordinate mapping after user feedback showed the ripple origin was shifted by one or more tiles. The touch callback already supplied coordinates relative to the whole display, so the scene now subtracts the `60px` left-strip offset instead of adding tile-local offsets on top (`7d074db` — `Fix touch ripple origin coordinate mapping`).
