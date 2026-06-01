/**
 * Per-tile clock example
 *
 * Demonstrates per-tile surfaces for independent tile content.
 * Each clock tile uses its own 90×90 surface. When a tile's
 * surface changes, only that tile is re-rendered and sent
 * to the hardware — 12× less data than a full-display redraw.
 *
 * Run with:
 *   go run ./cmd/loupedeck run ./examples/js/13-per-tile-clock.js --duration 30s
 */

__package__({
  name: 'per-tile-clock',
  short: 'Per-tile surface clock example'
});

function runScene() {
  const gfx = require("loupedeck/gfx");
  const ui = require("loupedeck/ui");
  const anim = require("loupedeck/anim");

  // Create per-tile surfaces — one per tile we want to draw into
  const clockSurface = gfx.surface(90, 90);
  const dateSurface = gfx.surface(90, 90);
  const secondsSurface = gfx.surface(90, 90);
  const meterSurface = gfx.surface(90, 90);

  ui.page("clock", page => {
    // Tile (0,0): Clock face showing HH:MM
    page.tile(0, 0, tile => {
      tile.surface(clockSurface);
    });

    // Tile (1,0): Date display
    page.tile(1, 0, tile => {
      tile.surface(dateSurface);
    });

    // Tile (2,0): Seconds meter
    page.tile(2, 0, tile => {
      tile.surface(secondsSurface);
    });

    // Tile (3,0): Second progress bar
    page.tile(3, 0, tile => {
      tile.surface(meterSurface);
    });

    // Other tiles use simple retained text
    page.tile(0, 1, tile => { tile.text("CLOCK"); });
    page.tile(1, 1, tile => { tile.text("TILES"); });
    page.tile(2, 1, tile => { tile.text("PER"); });
    page.tile(3, 1, tile => { tile.text("TILE"); });

    page.tile(0, 2, tile => { tile.text("12x"); });
    page.tile(1, 2, tile => { tile.text("LESS"); });
    page.tile(2, 2, tile => { tile.text("DATA"); });
    page.tile(3, 2, tile => { tile.text(":)"); });
  });

  let lastSecond = -1;

  // Update per-tile surfaces every 100ms
  anim.loop(100, () => {
    const d = new Date();
    const h = String(d.getHours()).padStart(2, "0");
    const m = String(d.getMinutes()).padStart(2, "0");
    const s = d.getSeconds();
    const ms = d.getMilliseconds();

    // Only do full redraws when the second changes
    if (s !== lastSecond) {
      lastSecond = s;

      // Clock surface: HH:MM
      clockSurface.batch(() => {
        clockSurface.clear(0);
        // Accent bar
        clockSurface.fillRect(0, 0, 90, 4, 80);
        // Time text
        clockSurface.text(`${h}:${m}`, {
          x: 0, y: 22, width: 90, height: 24,
          center: true, brightness: 220
        });
        // Seconds small text
        clockSurface.text(String(s).padStart(2, "0"), {
          x: 0, y: 50, width: 90, height: 14,
          center: true, brightness: 90
        });
        // Blinking colon
        if (s % 2 === 0) {
          clockSurface.fillRect(40, 28, 10, 2, 180);
        }
      });

      // Date surface
      const days = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
      const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun",
                      "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
      dateSurface.batch(() => {
        dateSurface.clear(0);
        dateSurface.fillRect(0, 0, 90, 4, 60);
        dateSurface.text(`${d.getDate()}`, {
          x: 0, y: 16, width: 90, height: 24,
          center: true, brightness: 200
        });
        dateSurface.text(days[d.getDay()], {
          x: 0, y: 44, width: 90, height: 14,
          center: true, brightness: 120
        });
        dateSurface.text(months[d.getMonth()], {
          x: 0, y: 62, width: 90, height: 14,
          center: true, brightness: 80
        });
      });
    }

    // Seconds meter: updates every 100ms for smooth animation
    secondsSurface.batch(() => {
      secondsSurface.clear(0);
      const progress = (s + ms / 1000) / 60;
      const barH = Math.round(progress * 70);
      // Background frame
      secondsSurface.fillRect(14, 8, 62, 74, 15);
      // Fill bar
      secondsSurface.fillRect(14, 78 - barH, 62, barH, 150);
      // Top highlight
      if (barH > 0) {
        secondsSurface.fillRect(14, 78 - barH, 62, 2, 255);
      }
    });

    // Progress bar surface
    meterSurface.batch(() => {
      meterSurface.clear(0);
      const progress = (s + ms / 1000) / 60;
      const barW = Math.round(progress * 82);
      // Background
      meterSurface.fillRect(4, 38, 82, 14, 20);
      // Fill
      if (barW > 0) {
        meterSurface.fillRect(4, 38, barW, 14, 160);
      }
      // Percentage text
      meterSurface.text(`${Math.round(progress * 100)}%`, {
        x: 0, y: 56, width: 90, height: 14,
        center: true, brightness: 100
      });
      meterSurface.text("MINUTE", {
        x: 0, y: 14, width: 90, height: 14,
        center: true, brightness: 60
      });
    });
  });

  ui.show("clock");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the per-tile clock example'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
