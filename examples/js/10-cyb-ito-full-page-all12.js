const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");
const present = require("loupedeck/present");
const sceneMetrics = require("loupedeck/scene-metrics").create("scene");

const TILE = 90;
const MAIN_W = 360;
const MAIN_H = 270;
const TOP = 3;
const LABEL_Y = 6;
const LABEL_H = 16;
const DIVIDER_Y = 22;

const main = gfx.surface(MAIN_W, MAIN_H);

const phase = state.signal(0);
const active = state.signal(0);
const lastEvent = state.signal("BOOT");

const tileLabels = [
  "EYE",
  "SPIRAL",
  "TEETH",
  "MELT",
  "HOLE",
  "FACE",
  "WORM",
  "NOISE",
  "WARP",
  "CRACK",
  "PULSE",
  "VOID",
];

function fract(v) {
  return v - Math.floor(v);
}

function hash(x, y, seed) {
  return fract(Math.sin(x * 12.9898 + y * 78.233 + seed * 37.719) * 43758.5453);
}

function clamp(v, min, max) {
  return Math.max(min, Math.min(max, v));
}

function tileRect(idx) {
  const col = idx % 4;
  const row = Math.floor(idx / 4);
  return { col, row, x: col * TILE, y: row * TILE };
}

function addP(surface, x, y, v) {
  surface.add(x | 0, y | 0, clamp(v | 0, 0, 255));
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

function drawTileFrame(x, y, isActive) {
  const border = isActive ? 255 : 24;
  const glow = isActive ? 80 : 0;
  main.fillRect(x, y, TILE, TILE, isActive ? 10 : 0);
  main.line(x, y, x + TILE - 1, y, border);
  main.line(x, y + TILE - 1, x + TILE - 1, y + TILE - 1, border);
  main.line(x, y, x, y + TILE - 1, border);
  main.line(x + TILE - 1, y, x + TILE - 1, y + TILE - 1, border);
  if (glow > 0) {
    main.line(x + 1, y + 1, x + TILE - 2, y + 1, glow);
    main.line(x + 1, y + TILE - 2, x + TILE - 2, y + TILE - 2, glow);
    main.line(x + 1, y + 1, x + 1, y + TILE - 2, glow);
    main.line(x + TILE - 2, y + 1, x + TILE - 2, y + TILE - 2, glow);
  }
}

function drawTileChrome(idx, x, y, isActive) {
  drawText(main, tileLabels[idx], x + 45, y + LABEL_Y, isActive ? 170 : 65, 74, LABEL_H);
  lineH(main, x + 2, x + TILE - 3, y + DIVIDER_Y, isActive ? 28 : 8);
}

function drawEyeTile(x, y, t, isActive) {
  const oy = y + TOP;
  const cx = x + 45;
  const cy = oy + 41;
  const brt = isActive ? 210 : 86;

  for (let a = 0; a < Math.PI * 2; a += 0.02) {
    const rx = 30 * Math.cos(a);
    const py = Math.sin(a) * 12;
    addP(main, cx + rx, cy + py, brt);
    addP(main, cx + rx * 1.04, cy + py * 1.1, brt * 0.4);
  }

  const irisR = 11 + Math.sin(t * 0.5) * 2;
  for (let a = 0; a < Math.PI * 2; a += 0.03) {
    addP(main, cx + Math.cos(a) * irisR, cy + Math.sin(a) * irisR, brt);
  }

  const pupilR = 5 + Math.sin(t * 1.5) * 2;
  fillDisk(main, cx, cy, pupilR | 0, brt * 0.9);

  addP(main, cx - 2, cy - 2, 255);
  addP(main, cx - 3, cy - 2, 255);
  addP(main, cx - 2, cy - 3, 255);

  for (let i = 0; i < 8; i++) {
    const a = i * Math.PI / 4 + Math.sin(t * 0.3 + i) * 0.2;
    for (let r = irisR + 1; r < irisR + 8 + Math.sin(t + i) * 3; r++) {
      const wx = Math.sin(r * 0.5 + i) * 0.8;
      addP(main, cx + Math.cos(a) * r + wx, cy + Math.sin(a) * r * 0.5, brt * 0.3 * (1 - r / 25));
    }
  }

  main.crosshatch(x + 4, oy + 16, 18, 50, 3, isActive ? 20 : 8);
  main.crosshatch(x + 68, oy + 16, 18, 50, 3, isActive ? 20 : 8);
}

function drawSpiralTile(x, y, t, isActive) {
  const oy = y + TOP;
  const cx = x + 45;
  const cy = oy + 43;
  const brt = isActive ? 190 : 76;
  drawSpiral(main, cx, cy, 6, 32, brt, 0.5, t, 2);
  drawSpiral(main, cx, cy, 4, 12, brt * 1.2, 0.85, t, 1);
  drawSpiral(main, x + 14, oy + 20, 2, 6, brt * 0.3, 1.0, t, 1);
  drawSpiral(main, x + 76, oy + 70, 2, 6, brt * 0.3, -0.7, t, 1);
  for (let r = 8; r < 34; r += 7) {
    const wobble = Math.sin(r * 0.5 + t) * 3;
    for (let a = 0; a < Math.PI * 2; a += 0.05) {
      const wr = r + Math.sin(a * 3 + t) * wobble;
      addP(main, cx + Math.cos(a) * wr, cy + Math.sin(a) * wr, brt * 0.12);
    }
  }
}

function drawTeethTile(x, y, t, isActive) {
  const oy = y + TOP;
  const cx = x + 45;
  const cy = oy + 43;
  const brt = isActive ? 235 : 110;
  const gapOpen = 3 + Math.sin(t * 0.45) * 2;

  for (let a = -Math.PI; a < 0; a += 0.03) {
    const rx = 32;
    const ry = 8 + Math.sin(t * 0.7) * 2;
    addP(main, cx + Math.cos(a) * rx, cy - 6 + Math.sin(a) * ry, brt);
    addP(main, cx + Math.cos(a) * rx, cy + 6 - Math.sin(a) * ry, brt);
  }

  main.fillRect(x + 12, cy - gapOpen + 1, 66, gapOpen * 2 - 1, 0);

  const teethW = 7;
  const teethCount = 8;
  for (let i = 0; i < teethCount; i++) {
    const tx = x + 9 + i * teethW + (i >= 4 ? 3 : 0);
    const up = 11 + Math.sin(i * 1.1 + t * 0.25) * 2;
    const down = 9 + Math.sin(i * 0.8 + t * 0.3) * 2;
    for (let dy = 0; dy < up; dy++) {
      const taper = 1 - dy / up * 0.25;
      const w = Math.max(3, (teethW * taper) | 0);
      for (let dx = 0; dx < w; dx++) addP(main, tx + dx, cy - gapOpen - dy, brt * (0.65 + dy / up * 0.35));
    }
    for (let dy = 0; dy < down; dy++) {
      const taper = 1 - dy / down * 0.2;
      const w = Math.max(3, (teethW * taper) | 0);
      for (let dx = 0; dx < w; dx++) addP(main, tx + dx, cy + gapOpen + dy, brt * (0.65 + dy / down * 0.35));
    }
  }

  main.crosshatch(x + 6, cy - gapOpen - 18, 78, 6, 2, brt * 0.12);
  main.crosshatch(x + 6, cy + gapOpen + 12, 78, 6, 2, brt * 0.12);
}

function drawMeltTile(x, y, t, isActive) {
  const brt = isActive ? 220 : 90;
  const top = y + 30;
  main.fillRect(x + 10, top, 70, 10, brt * 0.2);
  for (let i = 0; i < 8; i++) {
    const px = x + 12 + i * 9;
    const length = 14 + ((Math.sin(t * 2 + i * 0.9) * 0.5 + 0.5) * 28) | 0;
    const width = i % 3 === 0 ? 4 : 2;
    for (let w = 0; w < width; w++) {
      lineV(main, px + w, top + 7, top + 7 + length, brt * (0.7 + w * 0.1));
    }
    fillDisk(main, px + (width >> 1), top + 7 + length + 2, 3 + (i % 2), brt * 0.6);
  }
}

function drawHoleTile(x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 48;
  const brt = isActive ? 220 : 85;
  for (let r = 8; r < 30; r += 4) {
    for (let a = 0; a < Math.PI * 2; a += 0.04) {
      const wobble = Math.sin(a * 5 + t * 2 + r) * 2;
      const rr = r + wobble;
      addP(main, cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, brt * (1 - r / 34));
    }
  }
  for (let i = 0; i < 11; i++) {
    const a = i / 11 * Math.PI * 2 + t;
    main.line(cx, cy, cx + Math.cos(a) * 34, cy + Math.sin(a) * 26, brt * 0.12);
  }
  fillDisk(main, cx, cy, 9, 0);
}

function drawFaceTile(x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 46;
  const brt = isActive ? 215 : 88;
  for (let a = 0; a < Math.PI * 2; a += 0.04) {
    const rx = 24 + Math.sin(a * 3 + t) * 2;
    const ry = 28 + Math.cos(a * 2 + t * 0.8) * 2;
    addP(main, cx + Math.cos(a) * rx, cy + Math.sin(a) * ry, brt * 0.65);
  }
  fillDisk(main, cx - 10, cy - 6, 4, brt);
  fillDisk(main, cx + 12, cy - 4, 5, brt);
  main.line(cx - 2, cy, cx + 2, cy + 8, brt * 0.7);
  for (let a = 0.3; a < Math.PI - 0.2; a += 0.04) {
    addP(main, cx + Math.cos(a) * 15, cy + 16 + Math.sin(a) * 8, brt);
  }
}

function drawWormTile(x, y, t, isActive) {
  const brt = isActive ? 225 : 92;
  const baseY = y + 50;
  for (let i = 0; i < 9; i++) {
    const px = x + 10 + i * 8;
    const py = baseY + Math.sin(t * 2.2 + i * 0.6) * 14 + Math.sin(i * 0.2) * 5;
    const r = 3 + ((8 - i) * 0.2);
    fillDisk(main, px, py | 0, r | 0, brt * (1 - i / 12 * 0.3));
    if (i === 0) {
      addP(main, px + 2, py - 1, 255);
      addP(main, px + 2, py + 1, 255);
    }
  }
}

function drawNoiseTile(x, y, t, isActive) {
  const brt = isActive ? 190 : 70;
  for (let py = y + 30; py < y + 84; py += 2) {
    for (let px = x + 6; px < x + 84; px += 2) {
      const n = hash(px, py, (t * 37) | 0);
      if (n > 0.72) addP(main, px, py, brt);
      else if (n > 0.6) addP(main, px, py, brt * 0.4);
    }
  }
}

function drawWarpTile(x, y, t, isActive) {
  const brt = isActive ? 210 : 82;
  const cx = x + 45;
  const cy = y + 48;
  for (let gx = x + 10; gx <= x + 80; gx += 8) {
    for (let py = y + 28; py <= y + 84; py++) {
      const dx = Math.sin((py - y) * 0.11 + t * 2 + gx * 0.1) * 6;
      addP(main, gx + dx, py, brt * 0.42);
    }
  }
  for (let gy = y + 30; gy <= y + 82; gy += 8) {
    for (let px = x + 8; px <= x + 82; px++) {
      const dy = Math.cos(px * 0.11 + t * 2.3 + gy * 0.09) * 5;
      addP(main, px, gy + dy, brt * 0.35);
    }
  }
  fillDisk(main, cx, cy, 4, brt);
}

function branchCrack(x1, y1, angle, length, brt, depth) {
  if (depth <= 0 || length < 4) return;
  const x2 = x1 + Math.cos(angle) * length;
  const y2 = y1 + Math.sin(angle) * length;
  main.line(x1, y1, x2, y2, brt);
  branchCrack(x2, y2, angle - 0.5, length * 0.55, brt * 0.75, depth - 1);
  branchCrack(x2, y2, angle + 0.4, length * 0.45, brt * 0.6, depth - 1);
}

function drawCrackTile(x, y, t, isActive) {
  const brt = isActive ? 240 : 96;
  const cx = x + 44;
  const cy = y + 40;
  branchCrack(cx, cy, 1.7 + Math.sin(t) * 0.1, 28, brt, 4);
  branchCrack(cx, cy, 0.7 + Math.sin(t * 0.6) * 0.1, 22, brt * 0.7, 3);
  branchCrack(cx, cy, 2.6, 18, brt * 0.65, 3);
}

function drawPulseTile(x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 48;
  const brt = isActive ? 220 : 84;
  for (let r = 6; r < 28; r += 5) {
    const pulse = 1 + Math.sin(t * 4 - r * 0.4) * 0.18;
    for (let a = 0; a < Math.PI * 2; a += 0.05) {
      addP(main, cx + Math.cos(a) * r * pulse, cy + Math.sin(a) * r * pulse, brt * (1 - r / 30));
    }
  }
  const baseY = y + 74;
  lineH(main, x + 8, x + 24, baseY, brt * 0.5);
  main.line(x + 24, baseY, x + 34, baseY - 10, brt);
  main.line(x + 34, baseY - 10, x + 40, baseY + 6, brt);
  main.line(x + 40, baseY + 6, x + 48, baseY - 18, brt);
  main.line(x + 48, baseY - 18, x + 57, baseY, brt);
  lineH(main, x + 57, x + 82, baseY, brt * 0.5);
}

function drawVoidTile(x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 48;
  const brt = isActive ? 180 : 60;
  for (let i = 0; i < 60; i++) {
    const px = x + 8 + ((hash(i, 3, 1) * 74) | 0);
    const py = y + 28 + ((hash(i, 7, 2) * 54) | 0);
    const tw = hash(px, py, (t * 20) | 0);
    if (tw > 0.82) addP(main, px, py, 220 + tw * 30);
  }
  for (let a = 0; a < Math.PI * 2; a += 0.05) {
    const rr = 24 + Math.sin(a * 7 + t * 3) * 3;
    addP(main, cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, brt * 0.4);
  }
  fillDisk(main, cx, cy, 10 + (Math.sin(t * 2) * 2) | 0, 0);
}

function drawTile(idx, x, y, t, isActive) {
  return sceneMetrics.timeTile(tileLabels[idx], () => {
    drawTileFrame(x, y, isActive);
    drawTileChrome(idx, x, y, isActive);
    switch (idx) {
      case 0: return drawEyeTile(x, y, t, isActive);
      case 1: return drawSpiralTile(x, y, t, isActive);
      case 2: return drawTeethTile(x, y, t, isActive);
      case 3: return drawMeltTile(x, y, t, isActive);
      case 4: return drawHoleTile(x, y, t, isActive);
      case 5: return drawFaceTile(x, y, t, isActive);
      case 6: return drawWormTile(x, y, t, isActive);
      case 7: return drawNoiseTile(x, y, t, isActive);
      case 8: return drawWarpTile(x, y, t, isActive);
      case 9: return drawCrackTile(x, y, t, isActive);
      case 10: return drawPulseTile(x, y, t, isActive);
      case 11: return drawVoidTile(x, y, t, isActive);
    }
  });
}

function renderAll(reason) {
  const resolvedReason = reason || "unknown";
  sceneMetrics.trace("renderAll.begin", {
    reason: resolvedReason,
    active: String(active.get()),
    lastEvent: lastEvent.get(),
  });
  const result = sceneMetrics.recordRebuild(resolvedReason, () => {
    main.batch(() => {
      main.clear(0);
      const t = phase.get() * Math.PI * 2;
      const a = active.get();
      for (let i = 0; i < 12; i++) {
        const { x, y } = tileRect(i);
        drawTile(i, x, y, t, i === a);
      }
      drawText(main, lastEvent.get(), 315, 248, 120, 42, 14);
    });
  });
  sceneMetrics.trace("renderAll.end", {
    reason: resolvedReason,
    active: String(active.get()),
    lastEvent: lastEvent.get(),
  });
  return result;
}

function setActive(idx, why) {
  sceneMetrics.trace("setActive", { idx: String(idx), why });
  sceneMetrics.recordActivation(why);
  active.set(idx);
  lastEvent.set(why);
  present.invalidate(why);
}

ui.page("full-page-all12", page => {
  page.display("main", display => {
    display.surface(main);
  });
});

for (let i = 1; i <= 12; i++) {
  const idx = i - 1;
  ui.onTouch(`Touch${i}`, () => setActive(idx, `T${i}`));
}
ui.onButton("Button1", () => setActive((active.get() + 11) % 12, "B1"));
ui.onButton("Button2", () => setActive((active.get() + 1) % 12, "B2"));

present.onFrame(reason => {
  renderAll(reason || "present");
});

ui.show("full-page-all12");
present.invalidate("initial");
anim.loop(1400, t => {
  sceneMetrics.trace("loop.tick", {
    phase: String(t),
    active: String(active.get()),
  });
  sceneMetrics.recordLoopTick();
  phase.set(t);
  present.invalidate("loop");
});
