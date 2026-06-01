/**
 * Tile surface drawing example
 *
 * Demonstrates the full gfx surface drawing API on per-tile surfaces.
 * Each tile shows a different drawing primitive: rectangles, lines,
 * text, crosshatch, compositing, and custom patterns.
 *
 * Run with:
 *   go run ./cmd/loupedeck run ./examples/js/14-tile-surface-drawing.js --duration 30s
 */

__package__({
  name: 'tile-surface-drawing',
  short: 'Tile surface drawing primitives example'
});

function runScene() {
  const gfx = require("loupedeck/gfx");
  const ui = require("loupedeck/ui");
  const anim = require("loupedeck/anim");

  // Create 12 per-tile surfaces
  const surfaces = [];
  for (let i = 0; i < 12; i++) {
    surfaces.push(gfx.surface(90, 90));
  }

  ui.page("drawing", page => {
    for (let row = 0; row < 3; row++) {
      for (let col = 0; col < 4; col++) {
        const idx = row * 4 + col;
        page.tile(col, row, tile => {
          tile.surface(surfaces[idx]);
        });
      }
    }
  });

  // Tile 0: Fill rectangles
  surfaces[0].batch(() => {
    surfaces[0].clear(0);
    surfaces[0].fillRect(5, 5, 30, 30, 200);
    surfaces[0].fillRect(35, 5, 30, 30, 140);
    surfaces[0].fillRect(5, 35, 30, 30, 80);
    surfaces[0].fillRect(35, 35, 30, 30, 40);
    surfaces[0].text("RECT", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 1: Lines
  surfaces[1].batch(() => {
    surfaces[1].clear(0);
    // Diagonal lines
    for (let i = 0; i < 8; i++) {
      surfaces[1].line(i * 12, 0, 0, 70, 40 + i * 20);
      surfaces[1].line(89 - i * 12, 0, 89, 70, 40 + i * 20);
    }
    // Horizontal and vertical
    surfaces[1].line(0, 35, 89, 35, 100);
    surfaces[1].line(45, 0, 45, 70, 100);
    surfaces[1].text("LINES", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 2: Text rendering
  surfaces[2].batch(() => {
    surfaces[2].clear(0);
    surfaces[2].text("BIG", { x: 0, y: 10, width: 90, height: 20, center: true, brightness: 255 });
    surfaces[2].text("medium", { x: 0, y: 30, width: 90, height: 14, center: true, brightness: 160 });
    surfaces[2].text("small text", { x: 0, y: 48, width: 90, height: 14, center: true, brightness: 80 });
    surfaces[2].text("dim", { x: 0, y: 62, width: 90, height: 14, center: true, brightness: 40 });
    surfaces[2].text("TEXT", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 3: Crosshatch patterns
  surfaces[3].batch(() => {
    surfaces[3].clear(0);
    surfaces[3].crosshatch(0, 0, 44, 44, 2, 60);
    surfaces[3].crosshatch(46, 0, 44, 44, 4, 100);
    surfaces[3].crosshatch(0, 46, 44, 44, 3, 80);
    surfaces[3].crosshatch(46, 46, 44, 44, 6, 40);
    surfaces[3].text("HATCH", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 4: Pixel art (set/add)
  surfaces[4].batch(() => {
    surfaces[4].clear(0);
    // Draw a simple face with pixels
    // Eyes
    for (let dy = 0; dy < 4; dy++) {
      for (let dx = 0; dx < 4; dx++) {
        surfaces[4].set(25 + dx, 20 + dy, 220);
        surfaces[4].set(61 + dx, 20 + dy, 220);
      }
    }
    // Mouth (additive arc)
    for (let x = 25; x <= 65; x++) {
      const y = Math.round(50 + Math.sin((x - 25) / 40 * Math.PI) * 8);
      for (let dy = 0; dy < 2; dy++) {
        surfaces[4].add(x, y + dy, 150);
      }
    }
    // Nose dot
    surfaces[4].set(43, 36, 120);
    surfaces[4].set(44, 36, 120);
    surfaces[4].set(43, 37, 120);
    surfaces[4].set(44, 37, 120);
    surfaces[4].text("PIXEL", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 5: Animated wave (updates with anim loop)
  // (drawn in the animation loop below)

  // Tile 6: Composited layers
  surfaces[6].batch(() => {
    const overlay = gfx.surface(40, 40);
    overlay.fillRect(0, 0, 40, 40, 100);
    overlay.text("Hi", { x: 0, y: 12, width: 40, height: 14, center: true, brightness: 255 });

    surfaces[6].clear(0);
    surfaces[6].fillRect(5, 5, 80, 60, 30);
    surfaces[6].compositeAdd(overlay, 25, 20);
    surfaces[6].text("COMP", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 7: Non-centered vs centered text
  surfaces[7].batch(() => {
    surfaces[7].clear(0);
    // Left-aligned text
    surfaces[7].text("left", { x: 4, y: 8, width: 82, height: 14, center: false, brightness: 180 });
    // Centered text
    surfaces[7].text("center", { x: 0, y: 28, width: 90, height: 14, center: true, brightness: 180 });
    // Right-ish aligned by using a narrow width
    surfaces[7].text("right", { x: 44, y: 48, width: 42, height: 14, center: true, brightness: 180 });
    surfaces[7].text("ALIGN", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 8: Brightness gradient
  surfaces[8].batch(() => {
    surfaces[8].clear(0);
    for (let i = 0; i < 8; i++) {
      const brightness = Math.round((i + 1) / 8 * 255);
      surfaces[8].fillRect(4 + i * 10, 10, 9, 55, brightness);
    }
    surfaces[8].text("BRIGHT", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 9: Border patterns
  surfaces[9].batch(() => {
    surfaces[9].clear(0);
    // Outer border
    for (let i = 0; i < 3; i++) {
      surfaces[9].fillRect(i, i, 90 - 2*i, 1, 200);
      surfaces[9].fillRect(i, 89 - i, 90 - 2*i, 1, 200);
      surfaces[9].fillRect(i, i, 1, 90 - 2*i, 200);
      surfaces[9].fillRect(89 - i, i, 1, 90 - 2*i, 200);
    }
    // Corner dots
    surfaces[9].set(4, 4, 255);
    surfaces[9].set(85, 4, 255);
    surfaces[9].set(4, 85, 255);
    surfaces[9].set(85, 85, 255);
    surfaces[9].text("BORDER", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 10: Checkerboard
  surfaces[10].batch(() => {
    surfaces[10].clear(0);
    const size = 10;
    for (let row = 0; row < 9; row++) {
      for (let col = 0; col < 9; col++) {
        if ((row + col) % 2 === 0) {
          surfaces[10].fillRect(col * size, row * size, size, size, 120);
        }
      }
    }
    surfaces[10].text("CHECK", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
  });

  // Tile 11: Animated pulse (updated in animation loop)
  // (initial draw in the animation loop below)

  // Animation loop for tiles 5 (wave) and 11 (pulse)
  let frame = 0;
  anim.loop(50, () => {
    frame++;

    // Tile 5: Animated sine wave
    surfaces[5].batch(() => {
      surfaces[5].clear(0);
      const t = frame * 0.05;
      for (let x = 0; x < 90; x++) {
        const y1 = Math.round(35 + Math.sin(x * 0.08 + t) * 20);
        const y2 = Math.round(35 + Math.sin(x * 0.12 + t * 1.5) * 15);
        surfaces[5].add(x, y1, 180);
        surfaces[5].add(x, y1 + 1, 80);
        surfaces[5].add(x, y2, 60);
      }
      surfaces[5].text("WAVE", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
    });

    // Tile 11: Animated expanding rings
    surfaces[11].batch(() => {
      surfaces[11].clear(0);
      const t = frame * 0.03;
      const cx = 45, cy = 40;
      for (let r = 5; r <= 35; r += 6) {
        const phase = (t + r * 0.05) % 1;
        const radius = Math.round(phase * 40);
        const brightness = Math.round((1 - phase) * 200);
        if (brightness > 10) {
          for (let a = 0; a < Math.PI * 2; a += 0.08) {
            const px = Math.round(cx + Math.cos(a) * radius);
            const py = Math.round(cy + Math.sin(a) * radius);
            if (px >= 0 && px < 90 && py >= 0 && py < 90) {
              surfaces[11].add(px, py, brightness);
            }
          }
        }
      }
      surfaces[11].text("PULSE", { x: 0, y: 72, width: 90, height: 14, center: true, brightness: 100 });
    });
  });

  ui.show("drawing");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the tile surface drawing example'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
