/**
 * LOUPE-016 Probe Script: Per-Tile Invalidation Benchmark
 *
 * Compares two approaches to updating a single tile:
 *
 *   Path A (Retained Tile): Uses tile.text(() => ...) with reactive binding.
 *     Only the dirty tile is re-rendered and sent to hardware.
 *
 *   Path B (Surface Redraw): Uses display.surface() + present.onFrame().
 *     The ENTIRE display is re-rendered every frame, even if only one tile changed.
 *
 * Run with:
 *   go run ./cmd/loupedeck run ./ttmp/2026/05/28/LOUPE-016-.../scripts/02-per-tile-invalidation-probe.js --duration 10s
 */

__package__({
  name: 'per-tile-invalidation-probe',
  short: 'LOUPE-016 probe: per-tile vs full-page invalidation'
});

function runScene() {
  const ui = require("loupedeck/ui");
  const state = require("loupedeck/state");
  const anim = require("loupedeck/anim");
  const metrics = require("loupedeck/metrics");
  const gfx = require("loupedeck/gfx");

  const seconds = state.signal(0);

  // --- Path A: Retained Tile Mode (per-tile invalidation) ---
  // When seconds changes, only the tile at (3,2) is marked dirty.
  // The renderer only re-renders and sends that one 90x90 tile.

  ui.page("tile-perf", page => {
    page.tile(0, 0, tile => { tile.text("RETAINED"); });
    page.tile(1, 0, tile => { tile.text("MODE"); });

    // This is the reactive tile — only this one redraws when seconds changes
    page.tile(2, 0, tile => {
      tile.text(() => `SEC:${String(seconds.get()).padStart(2, "0")}`);
    });

    page.tile(3, 0, tile => { tile.text("STATIC"); });

    page.tile(0, 1, tile => { tile.text("A"); });
    page.tile(1, 1, tile => { tile.text("B"); });
    page.tile(2, 1, tile => { tile.text("C"); });
    page.tile(3, 1, tile => { tile.text("D"); });

    page.tile(0, 2, tile => { tile.text("E"); });
    page.tile(1, 2, tile => { tile.text("F"); });
    page.tile(2, 2, tile => { tile.text("G"); });
    page.tile(3, 2, tile => { tile.text("H"); });
  });

  // Animate seconds to simulate a clock
  anim.loop(1000, () => {
    const d = new Date();
    seconds.set(d.getSeconds());
  });

  ui.show("tile-perf");

  // NOTE: To test Path B (surface mode), you would need to:
  // 1. Create a gfx.surface(360, 270)
  // 2. Assign it to the main display via page.display("main", d => d.surface(mainSurface))
  // 3. Use present.onFrame() to redraw the entire surface every frame
  // 4. Use present.invalidate() on every change
  //
  // In Path B, even if only one tile's content changes,
  // renderer.Flush() re-renders the entire 360x270 display.
  // This is the performance problem we want to fix.
}

__verb__("runScene", {
  name: "run",
  short: "Run the per-tile invalidation probe"
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
