const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");
const present = require("loupedeck/present");
const sceneMetrics = require("loupedeck/scene-metrics").create("scene");

const TILE = 90;
const MAIN_W = 360;
const MAIN_H = 270;
const SIDE_W = 60;
const SIDE_H = 270;
const TOP = 3;
const LABEL_Y = 6;
const LABEL_H = 16;
const DIVIDER_Y = 22;
const KANJI_TILE_DX = -10;
const KANJI_TILE_DY = 7;
const KANJI_HUD_DX = -10;
const KANJI_HUD_DY = 7;
const SIDEBAR_KANJI_X = 8;
const SIDEBAR_KANJI_Y = 5;
const SIDEBAR_KANJI_W = 40;
const SIDEBAR_KANJI_H = 24;
const CJK_FONT_PATH = "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc";

function loadOptionalFont(path, opts) {
  try {
    return gfx.font(path, opts);
  } catch (_err) {
    return null;
  }
}

const KANJI_FONT = loadOptionalFont(CJK_FONT_PATH, { size: 18, dpi: 72, index: 0 });
const KANJI_FONT_SMALL = loadOptionalFont(CJK_FONT_PATH, { size: 14, dpi: 72, index: 0 });
const KANJI_FONT_SIDEBAR = loadOptionalFont(CJK_FONT_PATH, { size: 20, dpi: 72, index: 0 }) || KANJI_FONT;

const frame = gfx.surface(MAIN_W, MAIN_H);
const baseLayer = gfx.surface(MAIN_W, MAIN_H);
const chromeLayer = gfx.surface(MAIN_W, MAIN_H);
const sceneLayer = gfx.surface(MAIN_W, MAIN_H);
const fxLayer = gfx.surface(MAIN_W, MAIN_H);
const hudLayer = gfx.surface(MAIN_W, MAIN_H);
const accentLayer = gfx.surface(MAIN_W, MAIN_H);
const leftStrip = gfx.surface(SIDE_W, SIDE_H);
const rightStrip = gfx.surface(SIDE_W, SIDE_H);

const phase = state.signal(0);
const stripScroll = state.signal(0);
const frameCounter = state.signal(0);
const active = state.signal(0);
const lastEvent = state.signal("BOOT");
const touchRipple = state.signal(0);
const touchRippleOriginX = state.signal((MAIN_W / 2) | 0);
const touchRippleOriginY = state.signal((MAIN_H / 2) | 0);

let baseLayerReady = false;
let touchRippleHandle = null;

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

const tileChromeLabels = [
  "EYE",
  "SPIR",
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

const tileKanji = [
  "眼",
  "渦",
  "歯",
  "溶",
  "穴",
  "狂",
  "蟲",
  "砂",
  "歪",
  "裂",
  "脈",
  "闇",
];

const horrorKanji = "呪螺旋恐怖闇影穴裂溶歪狂蝕腐朽這寄生喰渦巻沈黙叫骸".split("");

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

function drawText(surface, text, x, y, brightness, width, height, font) {
  const opts = { x, y, width, height, brightness, center: true };
  if (font) opts.font = font;
  surface.text(text, opts);
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

function drip(surface, x, startY, len, brt, seed) {
  const wobble = Math.sin(seed * 7.7 + frameCounter.get() * 0.02) * 2;
  for (let i = 0; i < len; i++) {
    const fade = 1 - i / len;
    const wx = (x + Math.sin(i * 0.15 + wobble) * 1.5) | 0;
    addP(surface, wx, startY + i, brt * fade);
    if (i > len - 5) {
      addP(surface, wx - 1, startY + i, brt * fade * 0.6);
      addP(surface, wx + 1, startY + i, brt * fade * 0.6);
      if (i > len - 3) addP(surface, wx, startY + i + 1, brt * fade * 0.4);
    }
  }
}

function drawTileFrame(surface, x, y, isActive) {
  const border = isActive ? 255 : 24;
  const glow = isActive ? 80 : 0;
  surface.fillRect(x, y, TILE, TILE, isActive ? 10 : 0);
  surface.line(x, y, x + TILE - 1, y, border);
  surface.line(x, y + TILE - 1, x + TILE - 1, y + TILE - 1, border);
  surface.line(x, y, x, y + TILE - 1, border);
  surface.line(x + TILE - 1, y, x + TILE - 1, y + TILE - 1, border);
  if (glow > 0) {
    surface.line(x + 1, y + 1, x + TILE - 2, y + 1, glow);
    surface.line(x + 1, y + TILE - 2, x + TILE - 2, y + TILE - 2, glow);
    surface.line(x + 1, y + 1, x + 1, y + TILE - 2, glow);
    surface.line(x + TILE - 2, y + 1, x + TILE - 2, y + TILE - 2, glow);
  }
}

function drawTileChrome(surface, idx, x, y, isActive) {
  drawText(surface, tileKanji[idx], x + 17 + KANJI_TILE_DX, y + LABEL_Y - 2 + KANJI_TILE_DY, isActive ? 235 : 110, 24, 20, KANJI_FONT);
  drawText(surface, tileChromeLabels[idx], x + 46, y + LABEL_Y + 1, isActive ? 140 : 55, 40, 12);
  lineH(surface, x + 2, x + TILE - 3, y + DIVIDER_Y, isActive ? 28 : 8);
}

function drawEyeTile(surface, x, y, t, isActive) {
  const oy = y + TOP;
  const cx = x + 45;
  const cy = oy + 41;
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

  surface.crosshatch(x + 4, oy + 16, 18, 50, 3, isActive ? 20 : 8);
  surface.crosshatch(x + 68, oy + 16, 18, 50, 3, isActive ? 20 : 8);
}

function drawSpiralTile(surface, x, y, t, isActive) {
  const oy = y + TOP;
  const cx = x + 45;
  const cy = oy + 43;
  const brt = isActive ? 190 : 76;
  drawSpiral(surface, cx, cy, 6, 32, brt, 0.5, t, 2);
  drawSpiral(surface, cx, cy, 4, 12, brt * 1.2, 0.85, t, 1);
  drawSpiral(surface, x + 14, oy + 20, 2, 6, brt * 0.3, 1.0, t, 1);
  drawSpiral(surface, x + 76, oy + 70, 2, 6, brt * 0.3, -0.7, t, 1);
  for (let r = 8; r < 34; r += 7) {
    const wobble = Math.sin(r * 0.5 + t) * 3;
    for (let a = 0; a < Math.PI * 2; a += 0.05) {
      const wr = r + Math.sin(a * 3 + t) * wobble;
      addP(surface, cx + Math.cos(a) * wr, cy + Math.sin(a) * wr, brt * 0.12);
    }
  }
}

function drawTeethTile(surface, x, y, t, isActive) {
  const oy = y + TOP;
  const cx = x + 45;
  const cy = oy + 43;
  const brt = isActive ? 235 : 110;
  const gapOpen = 3 + Math.sin(t * 0.45) * 2;

  for (let a = -Math.PI; a < 0; a += 0.03) {
    const rx = 32;
    const ry = 8 + Math.sin(t * 0.7) * 2;
    addP(surface, cx + Math.cos(a) * rx, cy - 6 + Math.sin(a) * ry, brt);
    addP(surface, cx + Math.cos(a) * rx, cy + 6 - Math.sin(a) * ry, brt);
  }

  surface.fillRect(x + 12, cy - gapOpen + 1, 66, gapOpen * 2 - 1, 0);

  const teethW = 7;
  const teethCount = 8;
  for (let i = 0; i < teethCount; i++) {
    const tx = x + 9 + i * teethW + (i >= 4 ? 3 : 0);
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

  surface.crosshatch(x + 6, cy - gapOpen - 18, 78, 6, 2, brt * 0.12);
  surface.crosshatch(x + 6, cy + gapOpen + 12, 78, 6, 2, brt * 0.12);
}

function drawMeltTile(surface, x, y, t, isActive) {
  const brt = isActive ? 220 : 90;
  const top = y + 30;
  surface.fillRect(x + 10, top, 70, 10, brt * 0.2);
  for (let i = 0; i < 8; i++) {
    const px = x + 12 + i * 9;
    const length = 14 + ((Math.sin(t * 2 + i * 0.9) * 0.5 + 0.5) * 28) | 0;
    const width = i % 3 === 0 ? 4 : 2;
    for (let w = 0; w < width; w++) {
      lineV(surface, px + w, top + 7, top + 7 + length, brt * (0.7 + w * 0.1));
    }
    fillDisk(surface, px + (width >> 1), top + 7 + length + 2, 3 + (i % 2), brt * 0.6);
  }
}

function drawHoleTile(surface, x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 48;
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
    surface.line(cx, cy, cx + Math.cos(a) * 34, cy + Math.sin(a) * 26, brt * 0.12);
  }
  fillDisk(surface, cx, cy, 9, 0);
}

function drawFaceTile(surface, x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 46;
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
}

function drawWormTile(surface, x, y, t, isActive) {
  const brt = isActive ? 225 : 92;
  const baseY = y + 50;
  for (let i = 0; i < 9; i++) {
    const px = x + 10 + i * 8;
    const py = baseY + Math.sin(t * 2.2 + i * 0.6) * 14 + Math.sin(i * 0.2) * 5;
    const r = 3 + ((8 - i) * 0.2);
    fillDisk(surface, px, py | 0, r | 0, brt * (1 - i / 12 * 0.3));
    if (i === 0) {
      addP(surface, px + 2, py - 1, 255);
      addP(surface, px + 2, py + 1, 255);
    }
  }
}

function drawNoiseTile(surface, x, y, t, isActive) {
  const brt = isActive ? 190 : 70;
  for (let py = y + 30; py < y + 84; py += 2) {
    for (let px = x + 6; px < x + 84; px += 2) {
      const n = hash(px, py, (t * 37) | 0);
      if (n > 0.72) addP(surface, px, py, brt);
      else if (n > 0.6) addP(surface, px, py, brt * 0.4);
    }
  }
}

function drawWarpTile(surface, x, y, t, isActive) {
  const brt = isActive ? 210 : 82;
  const cx = x + 45;
  const cy = y + 48;
  for (let gx = x + 10; gx <= x + 80; gx += 8) {
    for (let py = y + 28; py <= y + 84; py++) {
      const dx = Math.sin((py - y) * 0.11 + t * 2 + gx * 0.1) * 6;
      addP(surface, gx + dx, py, brt * 0.42);
    }
  }
  for (let gy = y + 30; gy <= y + 82; gy += 8) {
    for (let px = x + 8; px <= x + 82; px++) {
      const dy = Math.cos(px * 0.11 + t * 2.3 + gy * 0.09) * 5;
      addP(surface, px, gy + dy, brt * 0.35);
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

function drawCrackTile(surface, x, y, t, isActive) {
  const brt = isActive ? 240 : 96;
  const cx = x + 44;
  const cy = y + 40;
  branchCrack(surface, cx, cy, 1.7 + Math.sin(t) * 0.1, 28, brt, 4);
  branchCrack(surface, cx, cy, 0.7 + Math.sin(t * 0.6) * 0.1, 22, brt * 0.7, 3);
  branchCrack(surface, cx, cy, 2.6, 18, brt * 0.65, 3);
}

function drawPulseTile(surface, x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 48;
  const brt = isActive ? 220 : 84;
  for (let r = 6; r < 28; r += 5) {
    const pulse = 1 + Math.sin(t * 4 - r * 0.4) * 0.18;
    for (let a = 0; a < Math.PI * 2; a += 0.05) {
      addP(surface, cx + Math.cos(a) * r * pulse, cy + Math.sin(a) * r * pulse, brt * (1 - r / 30));
    }
  }
  const baseY = y + 74;
  lineH(surface, x + 8, x + 24, baseY, brt * 0.5);
  surface.line(x + 24, baseY, x + 34, baseY - 10, brt);
  surface.line(x + 34, baseY - 10, x + 40, baseY + 6, brt);
  surface.line(x + 40, baseY + 6, x + 48, baseY - 18, brt);
  surface.line(x + 48, baseY - 18, x + 57, baseY, brt);
  lineH(surface, x + 57, x + 82, baseY, brt * 0.5);
}

function drawVoidTile(surface, x, y, t, isActive) {
  const cx = x + 45;
  const cy = y + 48;
  const brt = isActive ? 180 : 60;
  for (let i = 0; i < 60; i++) {
    const px = x + 8 + ((hash(i, 3, 1) * 74) | 0);
    const py = y + 28 + ((hash(i, 7, 2) * 54) | 0);
    const tw = hash(px, py, (t * 20) | 0);
    if (tw > 0.82) addP(surface, px, py, 220 + tw * 30);
  }
  for (let a = 0; a < Math.PI * 2; a += 0.05) {
    const rr = 24 + Math.sin(a * 7 + t * 3) * 3;
    addP(surface, cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, brt * 0.4);
  }
  fillDisk(surface, cx, cy, 10 + (Math.sin(t * 2) * 2) | 0, 0);
}

function drawTileArt(surface, idx, x, y, t, isActive) {
  return sceneMetrics.timeTile(tileLabels[idx], () => {
    switch (idx) {
      case 0: return drawEyeTile(surface, x, y, t, isActive);
      case 1: return drawSpiralTile(surface, x, y, t, isActive);
      case 2: return drawTeethTile(surface, x, y, t, isActive);
      case 3: return drawMeltTile(surface, x, y, t, isActive);
      case 4: return drawHoleTile(surface, x, y, t, isActive);
      case 5: return drawFaceTile(surface, x, y, t, isActive);
      case 6: return drawWormTile(surface, x, y, t, isActive);
      case 7: return drawNoiseTile(surface, x, y, t, isActive);
      case 8: return drawWarpTile(surface, x, y, t, isActive);
      case 9: return drawCrackTile(surface, x, y, t, isActive);
      case 10: return drawPulseTile(surface, x, y, t, isActive);
      case 11: return drawVoidTile(surface, x, y, t, isActive);
    }
  });
}

function rebuildBaseLayer() {
  baseLayer.batch(() => {
    baseLayer.clear(0);
    for (let y = 0; y < MAIN_H; y += 18) {
      lineH(baseLayer, 0, MAIN_W - 1, y, 2);
    }
    for (let i = 0; i < 12; i++) {
      const { x, y } = tileRect(i);
      baseLayer.fillRect(x + 3, y + 3, TILE - 6, TILE - 6, 3);
      baseLayer.fillRect(x + 8, y + 28, TILE - 16, TILE - 18, 2);
      if ((i % 2) === 0) {
        baseLayer.crosshatch(x + 6, y + 28, 18, 46, 4, 5);
      } else {
        baseLayer.crosshatch(x + 66, y + 28, 18, 46, 4, 5);
      }
    }
  });
  baseLayerReady = true;
}

function renderChromeLayer(activeIdx) {
  chromeLayer.batch(() => {
    chromeLayer.clear(0);
    for (let i = 0; i < 12; i++) {
      if (i === activeIdx) continue;
      const { x, y } = tileRect(i);
      drawTileFrame(chromeLayer, x, y, false);
      drawTileChrome(chromeLayer, i, x, y, false);
    }
  });
}

function renderSceneLayer(t, activeIdx) {
  sceneLayer.batch(() => {
    sceneLayer.clear(0);
    for (let i = 0; i < 12; i++) {
      if (i === activeIdx) continue;
      const { x, y } = tileRect(i);
      drawTileArt(sceneLayer, i, x, y, t, false);
    }
  });
}

function drawScanlines(surface, tick) {
  const phaseOffset = ((tick * 12) | 0) % 6;
  for (let y = phaseOffset; y < MAIN_H; y += 6) {
    lineH(surface, 0, MAIN_W - 1, y, 8);
    if (y + 1 < MAIN_H) {
      lineH(surface, 0, MAIN_W - 1, y + 1, 3);
    }
  }
}

function drawFrameNoise(surface, tick) {
  const seed = (tick * 97) | 0;
  for (let y = 0; y < MAIN_H; y += 3) {
    for (let x = (y + seed) % 5; x < MAIN_W; x += 5) {
      const n = hash(x, y, seed);
      if (n > 0.86) addP(surface, x, y, 22);
      else if (n > 0.76) addP(surface, x, y, 10);
    }
  }
}

function drawActiveSweep(surface, idx, tick, brightness) {
  const { x, y } = tileRect(idx);
  const sweepX = x + 6 + (((Math.sin(tick * Math.PI * 2) * 0.5 + 0.5) * (TILE - 12)) | 0);
  for (let py = y + 24; py < y + TILE - 6; py++) {
    const fade = 1 - Math.abs(py - (y + 54)) / 40;
    addP(surface, sweepX, py, brightness * fade);
    addP(surface, sweepX + 1, py, brightness * 0.35 * fade);
  }
}

function drawActiveRipple(surface, idx, tick, brightness) {
  const { x, y } = tileRect(idx);
  const cx = x + 45;
  const cy = y + 48;
  const pulse = tick * Math.PI * 2;
  for (let ring = 0; ring < 2; ring++) {
    const baseR = 18 + ring * 11;
    const radius = baseR + Math.sin(pulse * 1.3 - ring * 0.8) * 3;
    for (let a = 0; a < Math.PI * 2; a += 0.08) {
      addP(surface, cx + Math.cos(a) * radius, cy + Math.sin(a) * radius, brightness + ring * 10);
    }
  }
}

function renderFXLayer(tick) {
  fxLayer.batch(() => {
    fxLayer.clear(0);
    drawScanlines(fxLayer, tick);
    drawFrameNoise(fxLayer, tick);
  });
}

function retrigger(signal, targetValue, durationMs, previousHandle) {
  if (previousHandle && previousHandle.stop) {
    previousHandle.stop();
  }
  signal.set(targetValue);
  return anim.to(signal, 0, durationMs);
}

function triggerTouchRipple(idx, touchX, touchY) {
  const { x, y } = tileRect(idx);
  const ox = typeof touchX === "number" ? clamp(touchX - 60, 0, MAIN_W - 1) : x + 45;
  const oy = typeof touchY === "number" ? clamp(touchY, 0, MAIN_H - 1) : y + 48;
  touchRippleOriginX.set(ox);
  touchRippleOriginY.set(oy);
  touchRippleHandle = retrigger(touchRipple, 1, 1200, touchRippleHandle);
}

function drawSelectedTileAccent(surface, idx, tick, t) {
  const { x, y } = tileRect(idx);
  drawTileFrame(surface, x, y, true);
  drawTileChrome(surface, idx, x, y, true);
  drawTileArt(surface, idx, x, y, t, true);
  drawActiveSweep(surface, idx, tick, 120);
  drawActiveRipple(surface, idx, tick, 48);
}

function drawFullscreenSpiralRipple(surface, tick) {
  const amount = touchRipple.get();
  if (amount <= 0.001) {
    return;
  }
  const cx = touchRippleOriginX.get();
  const cy = touchRippleOriginY.get();
  const maxR = Math.sqrt(MAIN_W * MAIN_W + MAIN_H * MAIN_H);
  const progress = 1 - amount;
  const front = 18 + Math.pow(progress, 0.78) * (maxR + 40);
  const pulse = tick * Math.PI * 2;
  const arms = 5;
  const armSteps = 520;

  for (let arm = 0; arm < arms; arm++) {
    const armOffset = arm * (Math.PI * 2 / arms) + pulse * 0.6;
    for (let i = 0; i < armSteps; i++) {
      const p = i / (armSteps - 1);
      const r = p * front;
      const swirl = p * Math.PI * 18 + progress * 8;
      const angle = armOffset + swirl;
      const brightness = (1 - p * 0.5) * amount * 120;
      const px = cx + Math.cos(angle) * r;
      const py = cy + Math.sin(angle) * r;
      addP(surface, px, py, brightness);
      addP(surface, px + 1, py, brightness * 0.45);
      addP(surface, px, py + 1, brightness * 0.25);
    }
  }

  for (let ring = 0; ring < 3; ring++) {
    const ringFront = front - ring * 22;
    if (ringFront <= 0) continue;
    for (let a = 0; a < Math.PI * 2; a += 0.018) {
      const wobble = Math.sin(a * (6 + ring * 2) + pulse * (2.5 + ring * 0.4)) * (10 + ring * 4) * amount;
      const rr = ringFront + wobble;
      const brightness = amount * (85 - ring * 18);
      addP(surface, cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, brightness);
      addP(surface, cx + Math.cos(a) * (rr - 8), cy + Math.sin(a) * (rr - 8), brightness * 0.35);
    }
  }
}

function renderAccentLayer(tick, activeIdx, t) {
  accentLayer.batch(() => {
    accentLayer.clear(0);
    drawSelectedTileAccent(accentLayer, activeIdx, tick, t);
    drawFullscreenSpiralRipple(accentLayer, tick);
  });
}

function renderLeftStrip(tick, activeIdx) {
  leftStrip.batch(() => {
    leftStrip.clear(0);
    lineV(leftStrip, SIDE_W - 1, 0, SIDE_H - 1, 12);
    for (let seg = 0; seg < 12; seg++) {
      const sy = 4 + seg * 22;
      const sh = 18;
      const level = Math.sin(tick * Math.PI * 2 * 1.2 + seg * 0.8) * 0.4 + 0.5;
      const fillH = (level * sh) | 0;
      const brt = (12 + level * 45) | 0;
      for (let yy = 0; yy < fillH; yy++) {
        const py = sy + sh - 1 - yy;
        for (let x = 8; x < SIDE_W - 8; x++) {
          if ((x + py) % 2 === 0) leftStrip.set(x, py, brt);
        }
      }
      if (fillH > 5) {
        drip(leftStrip, (SIDE_W / 2) | 0, sy + sh, (8 + level * 10) | 0, (brt * 0.5) | 0, seg * 3.7);
      }
    }

    const activeY = tileRect(activeIdx).y + (TILE / 2) | 0;
    const pip = (16 + touchRipple.get() * 120) | 0;
    if (pip > 18) {
      for (let dy = -1; dy <= 1; dy++) {
        for (let dx = -1; dx <= 1; dx++) {
          addP(leftStrip, SIDE_W - 6 + dx, activeY + dy, pip);
        }
      }
    }
  });
}

function renderRightStrip(activeIdx) {
  rightStrip.batch(() => {
    rightStrip.clear(0);
    lineV(rightStrip, 0, 0, SIDE_H - 1, 12);
    const off = stripScroll.get();
    for (let i = 0; i < 16; i++) {
      const y = ((i * 20 - (off % 20) + SIDE_H) % SIDE_H) - 20;
      if (y < -20 || y > SIDE_H) continue;
      const ci = (i + Math.floor(off / 20)) % horrorKanji.length;
      const dist = Math.abs(y + 10 - SIDE_H / 2) / (SIDE_H / 2);
      const fade = Math.max(0, 1 - dist * 1.3);
      const brt = (fade * 120 + 36) | 0;
      drawText(rightStrip, horrorKanji[ci], SIDEBAR_KANJI_X, y + SIDEBAR_KANJI_Y, brt, SIDEBAR_KANJI_W, SIDEBAR_KANJI_H, KANJI_FONT_SIDEBAR);
    }

    const activeY = tileRect(activeIdx).y + (TILE / 2) | 0;
    const pip = (16 + touchRipple.get() * 120) | 0;
    if (pip > 18) {
      for (let dy = -1; dy <= 1; dy++) {
        for (let dx = -1; dx <= 1; dx++) {
          addP(rightStrip, 5 + dx, activeY + dy, pip);
        }
      }
    }
  });
}

function renderHUDLayer(activeIdx, eventText) {
  const { x, y } = tileRect(activeIdx);
  hudLayer.batch(() => {
    hudLayer.clear(0);
    hudLayer.fillRect(x + 28, y + 68, 34, 10, 16);
    drawText(hudLayer, eventText, 315, 248, 120, 42, 14);
    drawText(hudLayer, tileKanji[activeIdx], 282 + KANJI_HUD_DX, 229 + KANJI_HUD_DY, 150, 24, 20, KANJI_FONT_SMALL || KANJI_FONT);
    drawText(hudLayer, tileLabels[activeIdx], 324, 232, 70, 54, 12);
  });
}

function composeFrame() {
  frame.batch(() => {
    frame.clear(0);
    frame.compositeAdd(baseLayer, 0, 0);
    frame.compositeAdd(sceneLayer, 0, 0);
    frame.compositeAdd(chromeLayer, 0, 0);
    frame.compositeAdd(fxLayer, 0, 0);
    frame.compositeAdd(hudLayer, 0, 0);
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
    if (!baseLayerReady) {
      rebuildBaseLayer();
    }
    const tick = phase.get();
    const t = tick * Math.PI * 2;
    const activeIdx = active.get();
    const eventText = lastEvent.get();

    renderChromeLayer(activeIdx);
    renderSceneLayer(t, activeIdx);
    renderFXLayer(tick);
    renderAccentLayer(tick, activeIdx, t);
    renderLeftStrip(tick, activeIdx);
    renderRightStrip(activeIdx);
    renderHUDLayer(activeIdx, eventText);
    composeFrame();
  });
  sceneMetrics.trace("renderAll.end", {
    reason: resolvedReason,
    active: String(active.get()),
    lastEvent: lastEvent.get(),
  });
  return result;
}

function setActive(idx, why, isTouch, localX, localY) {
  sceneMetrics.trace("setActive", { idx: String(idx), why });
  sceneMetrics.recordActivation(why);
  active.set(idx);
  lastEvent.set(why);
  if (isTouch) {
    triggerTouchRipple(idx, localX, localY);
  }
  present.invalidate(why);
}

ui.page("full-page-all12", page => {
  page.display("left", display => {
    display.surface(leftStrip);
  });
  page.display("main", display => {
    display.surface(frame);
    display.layer("accent", accentLayer, { r: 255, g: 32, b: 32 });
  });
  page.display("right", display => {
    display.surface(rightStrip);
  });
});

for (let i = 1; i <= 12; i++) {
  const idx = i - 1;
  ui.onTouch(`Touch${i}`, event => setActive(idx, `T${i}`, true, event.x, event.y));
}
ui.onButton("Button1", () => setActive((active.get() + 11) % 12, "B1", false));
ui.onButton("Button2", () => setActive((active.get() + 1) % 12, "B2", false));

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
  frameCounter.set(frameCounter.get() + 1);
  stripScroll.set((stripScroll.get() + 0.4) % (horrorKanji.length * 20));
  present.invalidate("loop");
});
