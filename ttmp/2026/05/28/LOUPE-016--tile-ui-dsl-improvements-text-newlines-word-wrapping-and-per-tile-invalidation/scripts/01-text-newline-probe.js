/**
 * LOUPE-016 Probe Script: Text Newline and Wrapping Investigation
 *
 * This script tests the current behavior of text rendering in the
 * loupedeck tile UI DSL. It exercises:
 *
 *   1. Multi-line text (newline characters in tile.text())
 *   2. Long text that exceeds tile width (wrapping behavior)
 *   3. Per-tile invalidation via retained tile mode
 *
 * Run with:
 *   go run ./cmd/loupedeck run ./ttmp/2026/05/28/LOUPE-016-.../scripts/01-text-newline-probe.js --duration 10s
 */

__package__({
  name: 'text-newline-probe',
  short: 'LOUPE-016 probe: text newline and wrapping behavior'
});

function runScene() {
  const ui = require("loupedeck/ui");
  const state = require("loupedeck/state");

  // Test 1: Multi-line text with \n characters
  // Expected: text renders on multiple lines
  // Current: \n is ignored/corrupted, all text on one line

  const counter = state.signal(0);

  ui.page("text-probe", page => {
    // Tile (0,0): Static multi-line text
    page.tile(0, 0, tile => {
      tile.text("LINE1\nLINE2");
    });

    // Tile (1,0): Reactive multi-line text
    page.tile(1, 0, tile => {
      tile.text(() => `TICK\n${String(counter.get()).padStart(2, "0")}`);
    });

    // Tile (2,0): Long text that should wrap
    page.tile(2, 0, tile => {
      tile.text("THIS IS A LONG STRING THAT OVERFLOWS");
    });

    // Tile (3,0): Short text for comparison
    page.tile(3, 0, tile => {
      tile.text(() => `C:${counter.get()}`);
    });

    // Tile (0,1): Three lines
    page.tile(0, 1, tile => {
      tile.text("A\nB\nC");
    });

    // Tile (1,1): Single line (control)
    page.tile(1, 1, tile => {
      tile.text("CONTROL");
    });

    // Tile (2,1): Very long word
    page.tile(2, 1, tile => {
      tile.text("SUPERLONGWORDTHATWONTWRAP");
    });

    // Tile (3,1): Counter to observe per-tile invalidation
    page.tile(3, 1, tile => {
      tile.text(() => `SEC:${counter.get() % 60}`);
    });
  });

  ui.onButton("Button1", () => {
    counter.update(v => v + 1);
  });

  ui.show("text-probe");
}

__verb__("runScene", {
  name: "run",
  short: "Run the text newline probe"
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
