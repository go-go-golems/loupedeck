/**
 * Per-tile draw() clock example
 *
 * Demonstrates the tile.draw(fn) and tile.invalidate() APIs.
 * Unlike example 13 which creates shared surfaces and assigns them,
 * this example uses tile.draw() which auto-creates a per-tile surface
 * and passes it to the drawing function. tile.invalidate() is used
 * to mark tiles dirty for re-rendering on each animation frame.
 *
 * Run with:
 *   go run ./cmd/loupedeck run ./examples/js/15-tile-draw-clock.js --duration 30s
 */

__package__({
  name: 'tile-draw-clock',
  short: 'tile.draw() and tile.invalidate() clock example'
});

function runScene() {
  const ui = require("loupedeck/ui");
  const anim = require("loupedeck/anim");

  // Store tile references for invalidation
  const tiles = {};

  ui.page("clock", page => {
    // Tile (0,0): Hours — large centered text via tile.draw()
    page.tile(0, 0, tile => {
      tile.draw(s => {
        s.clear(0);
        s.fillRect(0, 0, 90, 8, 100);  // accent bar
        s.text("HH", { x: 0, y: 20, width: 90, height: 20, center: true, brightness: 180 });
      });
      tiles.hh = tile;
    });

    // Tile (1,0): Colon separator
    page.tile(1, 0, tile => {
      tile.draw(s => {
        s.clear(0);
        s.text(":", { x: 0, y: 30, width: 90, height: 20, center: true, brightness: 120 });
      });
      tiles.colon = tile;
    });

    // Tile (2,0): Minutes — large centered text via tile.draw()
    page.tile(2, 0, tile => {
      tile.draw(s => {
        s.clear(0);
        s.fillRect(0, 0, 90, 8, 100);
        s.text("MM", { x: 0, y: 20, width: 90, height: 20, center: true, brightness: 180 });
      });
      tiles.mm = tile;
    });

    // Tile (3,0): Seconds — animated with tile.draw() on each frame
    page.tile(3, 0, tile => {
      tile.draw(s => {
        s.clear(0);
        s.fillRect(0, 0, 90, 8, 60);   // dim accent
        s.text("SS", { x: 0, y: 20, width: 90, height: 20, center: true, brightness: 140 });
      });
      tiles.ss = tile;
    });

    // Row 1: Date and day tiles
    page.tile(0, 1, tile => {
      tile.draw(s => {
        s.clear(0);
        s.fillRect(0, 0, 90, 4, 40);
        s.text("DATE", { x: 0, y: 8, width: 90, height: 14, center: true, brightness: 80 });
      });
      tiles.date = tile;
    });

    page.tile(1, 1, tile => {
      tile.draw(s => {
        s.clear(0);
        s.fillRect(0, 0, 90, 4, 40);
        s.text("DAY", { x: 0, y: 8, width: 90, height: 14, center: true, brightness: 80 });
      });
      tiles.day = tile;
    });

    // Row 1 remaining tiles: decorative
    for (let col = 2; col < 4; col++) {
      page.tile(col, 1, tile => {
        tile.draw(s => {
          s.clear(10);
          s.crosshatch(0, 0, 90, 90, 4, 25);
        });
      });
    }

    // Row 2: Animated progress bar for seconds
    page.tile(0, 2, tile => {
      tile.draw(s => {
        s.clear(0);
        s.fillRect(0, 0, 90, 4, 40);
        s.text("SEC%", { x: 0, y: 8, width: 90, height: 14, center: true, brightness: 80 });
      });
      tiles.secBar = tile;
    });

    // Remaining row 2 and row 3 tiles: dim filler
    for (let row = 2; row < 3; row++) {
      for (let col = 1; col < 4; col++) {
        page.tile(col, row, tile => {
          tile.draw(s => {
            s.clear(5);
          });
        });
      }
    }
    for (let row = 3; row < 3; row++) {
      // No row 3 tiles in this example
    }
  });

  // Animate: update tiles every second using tile.draw() + tile.invalidate()
  let lastSec = -1;
  anim.loop(1, () => {
    const now = new Date();
    const sec = now.getSeconds();

    // Only redraw when the second changes
    if (sec === lastSec) return;
    lastSec = sec;

    const hh = String(now.getHours()).padStart(2, '0');
    const mm = String(now.getMinutes()).padStart(2, '0');
    const ss = String(sec).padStart(2, '0');

    // Redraw HH tile
    tiles.hh.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 8, 100);
      s.text(hh, { x: 0, y: 20, width: 90, height: 20, center: true, brightness: 200 });
    });

    // Redraw MM tile
    tiles.mm.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 8, 100);
      s.text(mm, { x: 0, y: 20, width: 90, height: 20, center: true, brightness: 200 });
    });

    // Redraw SS tile
    tiles.ss.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 8, 60);
      s.text(ss, { x: 0, y: 20, width: 90, height: 20, center: true, brightness: 140 });
    });

    // Redraw date tile
    const dateStr = now.toLocaleDateString();
    tiles.date.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 4, 40);
      s.text(dateStr, { x: 0, y: 8, width: 90, height: 14, center: true, brightness: 80 });
    });

    // Redraw day tile
    const days = ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'];
    tiles.day.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 4, 40);
      s.text(days[now.getDay()], { x: 0, y: 8, width: 90, height: 14, center: true, brightness: 80 });
    });

    // Redraw seconds progress bar
    const pct = sec / 60;
    tiles.secBar.draw(s => {
      s.clear(0);
      s.fillRect(0, 0, 90, 4, 40);
      s.text('SEC%', { x: 0, y: 8, width: 90, height: 14, center: true, brightness: 80 });
      const barW = Math.round(82 * pct);
      s.fillRect(4, 30, barW, 6, 120);
    });
  });
}

module.exports = { runScene };
