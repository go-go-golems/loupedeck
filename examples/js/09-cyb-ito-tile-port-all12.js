__package__({
  name: 'cyb-ito-tile-port-all12',
  short: 'CYB ITO all twelve tile-port example scene'
});

function runScene() {
  const state = require("loupedeck/state");
  const ui = require("loupedeck/ui");
  const gfx = require("loupedeck/gfx");
  const anim = require("loupedeck/anim");

  const TILE = 90;
  const TOP = 3;
  const LABEL_Y = 6;
  const LABEL_H = 16;
  const DIVIDER_Y = 22;
  const ART_Y = 28 + TOP;

  const phase = state.signal(0);
  const active = state.signal(0);
  const lastEvent = state.signal("BOOT");

  function fract(v) {
    return v - Math.floor(v);
  }

  function hash(x, y, seed) {
    return fract(Math.sin(x * 12.9898 + y * 78.233 + seed * 37.719) * 43758.5453);
  }

  function addP(surface, x, y, v) {
    surface.add(x | 0, y | 0, Math.max(0, Math.min(255, v | 0)));
  }

  function setP(surface, x, y, v) {
    surface.set(x | 0, y | 0, Math.max(0, Math.min(255, v | 0)));
  }

  function lineH(surface, x1, x2, y, v) {
    for (let x = x1; x <= x2; x++) addP(surface, x, y, v);
  }

  function lineV(surface, x, y1, y2, v) {
    for (let y = y1; y <= y2; y++) addP(surface, x, y, v);
  }

  function fillDisk(surface, cx, cy, r, v) {
    for (let dy = -r; dy <= r; dy++) {
      for (let dx = -r; dx <= r; dx++) {
        if (dx * dx + dy * dy <= r * r) addP(surface, cx + dx, cy + dy, v);
      }
    }
  }

  function drawText(surface, text, x, y, brightness, width, height) {
    surface.text(text, { x, y, width, height, brightness, center: true });
  }

  function drawSpiral(surface, cx, cy, turns, size, brt, speed, t, thick) {
    const steps = turns * 120;
    for (let i = 0; i < steps; i++) {
      const angle = i * 0.05 + t * speed;
      const r = i * size / steps;
      const px = (cx + Math.cos(angle) * r) | 0;
      const py = (cy + Math.sin(angle) * r) | 0;
      const fade = 1 - i / steps * 0.3;
      addP(surface, px, py, brt * fade);
      if (thick > 1) addP(surface, px + 1, py, brt * fade * 0.5);
    }
  }

  function frame(surface, label, isActive) {
    surface.clear(isActive ? 10 : 0);
    const border = isActive ? 255 : 24;
    const accent = isActive ? 170 : 70;
    surface.line(0, 0, 89, 0, border);
    surface.line(0, 89, 89, 89, border);
    surface.line(0, 0, 0, 89, border);
    surface.line(89, 0, 89, 89, border);
    drawText(surface, label, 45, LABEL_Y, accent, 74, LABEL_H);
    lineH(surface, 2, 87, DIVIDER_Y, isActive ? 28 : 8);
  }

  function renderEye(surface, t, isActive) {
    frame(surface, "EYE", isActive);
    const cx = 45;
    const cy = ART_Y + 18;
    const brt = isActive ? 210 : 86;

    for (let a = 0; a < Math.PI * 2; a += 0.02) {
      const rx = 30 * Math.cos(a);
      const py = Math.sin(a) * 12;
      addP(surface, cx + rx, cy + py, brt);
      addP(surface, cx + rx * 1.04, cy + py * 1.1, brt * 0.4);
    }

    const irisR = 11 + Math.sin(t * 0.5) * 2;
    for (let a = 0; a < Math.PI * 2; a += 0.03) {
      addP(surface, cx + Math.cos(a) * irisR, cy + Math.sin(a) * irisR, brt);
    }

    const pupilR = 5 + Math.sin(t * 1.5) * 2;
    fillDisk(surface, cx, cy, pupilR | 0, brt * 0.9);

    addP(surface, cx - 2, cy - 2, 255);
    addP(surface, cx - 3, cy - 2, 255);
    addP(surface, cx - 2, cy - 3, 255);

    for (let i = 0; i < 8; i++) {
      const a = i * Math.PI / 4 + Math.sin(t * 0.3 + i) * 0.2;
      for (let r = irisR + 1; r < irisR + 8 + Math.sin(t + i) * 3; r++) {
        const wx = Math.sin(r * 0.5 + i) * 0.8;
        addP(surface, cx + Math.cos(a) * r + wx, cy + Math.sin(a) * r * 0.5, brt * 0.3 * (1 - r / 25));
      }
    }

    surface.crosshatch(4, ART_Y - 6, 18, 50, 3, isActive ? 20 : 8);
    surface.crosshatch(68, ART_Y - 6, 18, 50, 3, isActive ? 20 : 8);
  }

  function renderSpiral(surface, t, isActive) {
    frame(surface, "SPIRAL", isActive);
    const cx = 45;
    const cy = ART_Y + 20;
    const brt = isActive ? 190 : 76;
    drawSpiral(surface, cx, cy, 6, 32, brt, 0.5, t, 2);
    drawSpiral(surface, cx, cy, 4, 12, brt * 1.2, 0.85, t, 1);
    drawSpiral(surface, 14, ART_Y - 3, 2, 6, brt * 0.3, 1.0, t, 1);
    drawSpiral(surface, 76, ART_Y + 47, 2, 6, brt * 0.3, -0.7, t, 1);
    for (let r = 8; r < 34; r += 7) {
      const wobble = Math.sin(r * 0.5 + t) * 3;
      for (let a = 0; a < Math.PI * 2; a += 0.05) {
        const wr = r + Math.sin(a * 3 + t) * wobble;
        addP(surface, cx + Math.cos(a) * wr, cy + Math.sin(a) * wr, brt * 0.12);
      }
    }
  }

  function renderTeeth(surface, t, isActive) {
    frame(surface, "TEETH", isActive);
    const cx = 45;
    const cy = ART_Y + 20;
    const brt = isActive ? 235 : 110;
    const gapOpen = 3 + Math.sin(t * 0.45) * 2;

    for (let a = -Math.PI; a < 0; a += 0.03) {
      const rx = 32;
      const ry = 8 + Math.sin(t * 0.7) * 2;
      addP(surface, cx + Math.cos(a) * rx, cy - 6 + Math.sin(a) * ry, brt);
      addP(surface, cx + Math.cos(a) * rx, cy + 6 - Math.sin(a) * ry, brt);
    }

    surface.fillRect(12, cy - gapOpen + 1, 66, gapOpen * 2 - 1, 0);

    const teethW = 7;
    for (let i = 0; i < 8; i++) {
      const tx = 9 + i * teethW + (i >= 4 ? 3 : 0);
      const up = 11 + Math.sin(i * 1.1 + t * 0.25) * 2;
      const down = 9 + Math.sin(i * 0.8 + t * 0.3) * 2;
      for (let dy = 0; dy < up; dy++) {
        const taper = 1 - dy / up * 0.25;
        const w = Math.max(3, (teethW * taper) | 0);
        for (let dx = 0; dx < w; dx++) addP(surface, tx + dx, cy - gapOpen - dy, brt * (0.65 + dy / up * 0.35));
      }
      for (let dy = 0; dy < down; dy++) {
        const taper = 1 - dy / down * 0.2;
        const w = Math.max(3, (teethW * taper) | 0);
        for (let dx = 0; dx < w; dx++) addP(surface, tx + dx, cy + gapOpen + dy, brt * (0.65 + dy / down * 0.35));
      }
    }

    surface.crosshatch(6, cy - gapOpen - 18, 78, 6, 2, brt * 0.12);
    surface.crosshatch(6, cy + gapOpen + 12, 78, 6, 2, brt * 0.12);
  }

  function renderMelt(surface, t, isActive) {
    frame(surface, "MELT", isActive);
    const brt = isActive ? 220 : 90;
    surface.fillRect(10, ART_Y - 2, 70, 10, brt * 0.2);
    for (let i = 0; i < 8; i++) {
      const x = 12 + i * 9;
      const length = 14 + ((Math.sin(t * 2 + i * 0.9) * 0.5 + 0.5) * 28) | 0;
      const width = i % 3 === 0 ? 4 : 2;
      for (let w = 0; w < width; w++) {
        lineV(surface, x + w, ART_Y + 5, ART_Y + 5 + length, brt * (0.7 + w * 0.1));
      }
      fillDisk(surface, x + (width >> 1), ART_Y + 5 + length + 2, 3 + (i % 2), brt * 0.6);
    }
    surface.crosshatch(8, ART_Y + 10, 74, 42, 3, isActive ? 18 : 8);
    lineH(surface, 10, 80, ART_Y + 48, brt * 0.3);
  }

  function renderHole(surface, t, isActive) {
    frame(surface, "HOLE", isActive);
    const cx = 45;
    const cy = ART_Y + 20;
    const brt = isActive ? 220 : 85;
    for (let r = 8; r < 30; r += 4) {
      for (let a = 0; a < Math.PI * 2; a += 0.04) {
        const wobble = Math.sin(a * 5 + t * 2 + r) * 2;
        const rr = r + wobble;
        addP(surface, cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, brt * (1 - r / 34));
      }
    }
    for (let i = 0; i < 11; i++) {
      const a = i / 11 * Math.PI * 2 + t;
      lineH(surface, cx - 1, cx + 1, cy, 0);
      surface.line(cx, cy, cx + Math.cos(a) * 34, cy + Math.sin(a) * 26, brt * 0.12);
    }
    fillDisk(surface, cx, cy, 9, 0);
  }

  function renderFace(surface, t, isActive) {
    frame(surface, "FACE", isActive);
    const cx = 45;
    const cy = ART_Y + 18;
    const brt = isActive ? 215 : 88;
    for (let a = 0; a < Math.PI * 2; a += 0.04) {
      const rx = 24 + Math.sin(a * 3 + t) * 2;
      const ry = 28 + Math.cos(a * 2 + t * 0.8) * 2;
      addP(surface, cx + Math.cos(a) * rx, cy + Math.sin(a) * ry, brt * 0.65);
    }
    fillDisk(surface, cx - 10, cy - 6, 4, brt);
    fillDisk(surface, cx + 12, cy - 4, 5, brt);
    surface.line(cx - 2, cy, cx + 2, cy + 8, brt * 0.7);
    for (let a = 0.3; a < Math.PI - 0.2; a += 0.04) {
      addP(surface, cx + Math.cos(a) * 15, cy + 16 + Math.sin(a) * 8, brt);
    }
    surface.crosshatch(18, ART_Y + 6, 54, 40, 4, isActive ? 16 : 7);
  }

  function renderWorm(surface, t, isActive) {
    frame(surface, "WORM", isActive);
    const brt = isActive ? 225 : 92;
    const baseY = ART_Y + 20;
    for (let i = 0; i < 9; i++) {
      const x = 10 + i * 8;
      const y = baseY + Math.sin(t * 2.2 + i * 0.6) * 14 + Math.sin(i * 0.2) * 5;
      const r = 3 + ((8 - i) * 0.2);
      fillDisk(surface, x, y | 0, r | 0, brt * (1 - i / 12 * 0.3));
      if (i === 0) {
        addP(surface, x + 2, y - 1, 255);
        addP(surface, x + 2, y + 1, 255);
      }
    }
    for (let i = 0; i < 6; i++) {
      const x = 26 + i * 8;
      const y = baseY + Math.sin(t * 2.2 + i * 0.6 + 1.2) * 14;
      lineV(surface, x, y + 5, y + 10, brt * 0.25);
    }
  }

  function renderNoise(surface, t, isActive) {
    frame(surface, "NOISE", isActive);
    const brt = isActive ? 190 : 70;
    for (let y = ART_Y - 2; y < 84; y += 2) {
      for (let x = 6; x < 84; x += 2) {
        const n = hash(x, y, (t * 37) | 0);
        if (n > 0.72) addP(surface, x, y, brt);
        else if (n > 0.6) addP(surface, x, y, brt * 0.4);
      }
    }
    for (let y = ART_Y + 6; y < 82; y += 7) {
      lineH(surface, 10, 80, y, brt * 0.07);
    }
  }

  function renderWarp(surface, t, isActive) {
    frame(surface, "WARP", isActive);
    const brt = isActive ? 210 : 82;
    const cx = 45;
    const cy = ART_Y + 20;
    for (let gx = 10; gx <= 80; gx += 8) {
      for (let y = ART_Y - 2; y <= 84; y++) {
        const dx = Math.sin((y - ART_Y) * 0.11 + t * 2 + gx * 0.1) * 6;
        addP(surface, gx + dx, y, brt * 0.42);
      }
    }
    for (let gy = ART_Y + 2; gy <= 82; gy += 8) {
      for (let x = 8; x <= 82; x++) {
        const dy = Math.cos(x * 0.11 + t * 2.3 + gy * 0.09) * 5;
        addP(surface, x, gy + dy, brt * 0.35);
      }
    }
    fillDisk(surface, cx, cy, 4, brt);
  }

  function branchCrack(surface, x1, y1, angle, length, brt, depth) {
    if (depth <= 0 || length < 4) return;
    const x2 = x1 + Math.cos(angle) * length;
    const y2 = y1 + Math.sin(angle) * length;
    surface.line(x1, y1, x2, y2, brt);
    branchCrack(surface, x2, y2, angle - 0.5, length * 0.55, brt * 0.75, depth - 1);
    branchCrack(surface, x2, y2, angle + 0.4, length * 0.45, brt * 0.6, depth - 1);
  }

  function renderCrack(surface, t, isActive) {
    frame(surface, "CRACK", isActive);
    const brt = isActive ? 240 : 96;
    const cx = 44;
    const cy = ART_Y + 12;
    branchCrack(surface, cx, cy, 1.7 + Math.sin(t) * 0.1, 28, brt, 4);
    branchCrack(surface, cx, cy, 0.7 + Math.sin(t * 0.6) * 0.1, 22, brt * 0.7, 3);
    branchCrack(surface, cx, cy, 2.6, 18, brt * 0.65, 3);
    surface.crosshatch(18, ART_Y + 22, 48, 26, 5, isActive ? 10 : 4);
  }

  function renderPulse(surface, t, isActive) {
    frame(surface, "PULSE", isActive);
    const cx = 45;
    const cy = ART_Y + 22;
    const brt = isActive ? 220 : 84;
    for (let r = 6; r < 28; r += 5) {
      const pulse = 1 + Math.sin(t * 4 - r * 0.4) * 0.18;
      for (let a = 0; a < Math.PI * 2; a += 0.05) {
        addP(surface, cx + Math.cos(a) * r * pulse, cy + Math.sin(a) * r * pulse, brt * (1 - r / 30));
      }
    }
    const baseY = 74;
    lineH(surface, 8, 24, baseY, brt * 0.5);
    surface.line(24, baseY, 34, baseY - 10, brt);
    surface.line(34, baseY - 10, 40, baseY + 6, brt);
    surface.line(40, baseY + 6, 48, baseY - 18, brt);
    surface.line(48, baseY - 18, 57, baseY, brt);
    lineH(surface, 57, 82, baseY, brt * 0.5);
  }

  function renderVoid(surface, t, isActive) {
    frame(surface, "VOID", isActive);
    const cx = 45;
    const cy = ART_Y + 20;
    const brt = isActive ? 180 : 60;
    for (let i = 0; i < 60; i++) {
      const x = 8 + ((hash(i, 3, 1) * 74) | 0);
      const y = ART_Y - 2 + ((hash(i, 7, 2) * 54) | 0);
      const tw = hash(x, y, (t * 20) | 0);
      if (tw > 0.82) addP(surface, x, y, 220 + tw * 30);
    }
    for (let a = 0; a < Math.PI * 2; a += 0.05) {
      const rr = 24 + Math.sin(a * 7 + t * 3) * 3;
      addP(surface, cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, brt * 0.4);
    }
    fillDisk(surface, cx, cy, 10 + (Math.sin(t * 2) * 2) | 0, 0);
    surface.crosshatch(14, ART_Y + 8, 62, 40, 6, isActive ? 10 : 4);
  }

  const tiles = [
    { key: "EYE", surface: gfx.surface(TILE, TILE), draw: renderEye },
    { key: "SPIRAL", surface: gfx.surface(TILE, TILE), draw: renderSpiral },
    { key: "TEETH", surface: gfx.surface(TILE, TILE), draw: renderTeeth },
    { key: "MELT", surface: gfx.surface(TILE, TILE), draw: renderMelt },
    { key: "HOLE", surface: gfx.surface(TILE, TILE), draw: renderHole },
    { key: "FACE", surface: gfx.surface(TILE, TILE), draw: renderFace },
    { key: "WORM", surface: gfx.surface(TILE, TILE), draw: renderWorm },
    { key: "NOISE", surface: gfx.surface(TILE, TILE), draw: renderNoise },
    { key: "WARP", surface: gfx.surface(TILE, TILE), draw: renderWarp },
    { key: "CRACK", surface: gfx.surface(TILE, TILE), draw: renderCrack },
    { key: "PULSE", surface: gfx.surface(TILE, TILE), draw: renderPulse },
    { key: "VOID", surface: gfx.surface(TILE, TILE), draw: renderVoid },
  ];

  function renderAll() {
    const t = phase.get() * Math.PI * 2;
    const a = active.get();
    for (let i = 0; i < tiles.length; i++) {
      tiles[i].draw(tiles[i].surface, t, a === i);
    }
  }

  function setActive(idx, why) {
    active.set(idx);
    lastEvent.set(why);
    renderAll();
  }

  ui.page("tiles-all12", page => {
    for (let i = 0; i < tiles.length; i++) {
      const col = i % 4;
      const row = (i / 4) | 0;
      const surface = tiles[i].surface;
      page.tile(col, row, tile => {
        tile.surface(surface);
      });
    }
  });

  for (let i = 1; i <= 12; i++) {
    const idx = i - 1;
    ui.onTouch(`Touch${i}`, () => setActive(idx, `T${i}`));
  }
  ui.onButton("Button1", () => setActive((active.get() + 11) % 12, "B1"));
  ui.onButton("Button2", () => setActive((active.get() + 1) % 12, "B2"));

  renderAll();
  anim.loop(1400, t => {
    phase.set(t);
    renderAll();
  });

  ui.show("tiles-all12");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the CYB ITO all twelve tile-port example scene'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
