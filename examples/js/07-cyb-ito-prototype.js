const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");

const TILE = 90;
const MAIN_W = 360;
const MAIN_H = 270;
const SIDE_W = 60;
const SIDE_H = 270;
const VISIBLE_TOP_INSET = 3;
const SHOW_SCAN_LAYER = false;
const SHOW_RIPPLE_LAYER = false;
const SHOW_HUD_LAYER = false;

const main = gfx.surface(MAIN_W, MAIN_H);
const mainScan = gfx.surface(MAIN_W, MAIN_H);
const mainRipple = gfx.surface(MAIN_W, MAIN_H);
const mainHUD = gfx.surface(MAIN_W, MAIN_H);
const left = gfx.surface(SIDE_W, SIDE_H);
const right = gfx.surface(SIDE_W, SIDE_H);

const phase = state.signal(0);
const scroll = state.signal(0);
const active = state.signal(0);
const lastEvent = state.signal("BOOT");
const ripple = state.signal(0);
const flash = state.signal(0);
let rippleHandle = null;
let flashHandle = null;

const sideText = ["EYE", "SPIN", "TEETH", "MELT", "HOLE", "FACE", "WORM", "NOISE", "WARP", "CRACK", "PULSE", "VOID", "TOUCH", "B1", "B2", "LIVE"];

const tiles = [
  { key: "EYE", short: "EYE", draw: drawEyeTile },
  { key: "SPIRAL", short: "SPIR", draw: drawSpiralTile },
  { key: "TEETH", short: "TEETH", draw: drawTeethTile },
  { key: "MELT", short: "MELT", draw: drawGenericTile },
  { key: "HOLE", short: "HOLE", draw: drawGenericTile },
  { key: "FACE", short: "FACE", draw: drawGenericTile },
  { key: "WORM", short: "WORM", draw: drawGenericTile },
  { key: "NOISE", short: "NOISE", draw: drawGenericTile },
  { key: "WARP", short: "WARP", draw: drawGenericTile },
  { key: "CRACK", short: "CRACK", draw: drawGenericTile },
  { key: "PULSE", short: "PULSE", draw: drawGenericTile },
  { key: "VOID", short: "VOID", draw: drawGenericTile },
];

function clamp(v, min, max) {
  return Math.max(min, Math.min(max, v));
}

function tileRect(idx) {
  const col = idx % 4;
  const row = Math.floor(idx / 4);
  return { col, row, x: col * TILE, y: row * TILE };
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
    if (thick > 2) addP(surface, px, py + 1, brt * fade * 0.3);
  }
}

function drawTileFrame(x, y, isActive) {
  const border = isActive ? 255 : 24;
  const glow = isActive ? 110 : 0;
  main.line(x, y, x + TILE - 1, y, border);
  main.line(x, y + TILE - 1, x + TILE - 1, y + TILE - 1, border);
  main.line(x, y, x, y + TILE - 1, border);
  main.line(x + TILE - 1, y, x + TILE - 1, y + TILE - 1, border);
  main.line(x + 1, y + 1, x + TILE - 2, y + 1, glow);
  main.line(x + 1, y + TILE - 2, x + TILE - 2, y + TILE - 2, glow);
  main.line(x + 1, y + 1, x + 1, y + TILE - 2, glow);
  main.line(x + TILE - 2, y + 1, x + TILE - 2, y + TILE - 2, glow);
  for (let i = 0; i < 5; i++) {
    addP(main, x + 1 + i, y + 1, isActive ? 150 : 14);
    addP(main, x + 1, y + 1 + i, isActive ? 150 : 14);
    addP(main, x + TILE - 2 - i, y + 1, isActive ? 150 : 14);
    addP(main, x + TILE - 2, y + 1 + i, isActive ? 150 : 14);
  }
}

function drawTileChrome(idx, x, y, isActive) {
  const titleY = y + VISIBLE_TOP_INSET + 2;
  const dividerY = y + VISIBLE_TOP_INSET + 13;
  drawText(main, String(idx + 1).padStart(2, "0"), x + 15, titleY, isActive ? 180 : 70, 22, 10);
  drawText(main, tiles[idx].short, x + 58, titleY, isActive ? 110 : 30, 48, 10);
  lineH(main, x + 2, x + TILE - 3, dividerY, isActive ? 30 : 6);
}

function drawGenericTile(idx, x, y, t, isActive) {
  const pulse = Math.floor((Math.sin(t * 2 + idx * 0.4) * 0.5 + 0.5) * 28);
  main.fillRect(x + 6, y + VISIBLE_TOP_INSET + 18, TILE - 12, TILE - 28, isActive ? 10 : 0);
  main.crosshatch(x + 10, y + VISIBLE_TOP_INSET + 22, TILE - 20, TILE - 34, isActive ? 5 : 7, isActive ? 28 + pulse : 10 + pulse);
  const cx = x + 45;
  const cy = y + 48;
  const r = isActive ? 18 : 10 + Math.floor(Math.sin(t * 1.4 + idx) * 4);
  main.line(cx - r, cy, cx + r, cy, isActive ? 230 : 70 + pulse);
  main.line(cx, cy - r, cx, cy + r, isActive ? 230 : 60 + pulse);
  main.fillRect(cx - (isActive ? 5 : 2), cy - (isActive ? 5 : 2), isActive ? 10 : 4, isActive ? 10 : 4, isActive ? 255 : 90 + pulse);
  if (isActive) {
    drawText(main, "ACTIVE", x + 45, y + 68, 255, 66, 12);
  }
}

function drawEyeTile(_idx, x, y, t, isActive) {
  const oy = y + VISIBLE_TOP_INSET;
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
  for (let dy = -pupilR; dy <= pupilR; dy++) {
    for (let dx = -pupilR; dx <= pupilR; dx++) {
      if (dx * dx + dy * dy <= pupilR * pupilR) {
        addP(main, cx + dx, cy + dy, brt * 0.9);
      }
    }
  }

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

  for (let i = 0; i < 3; i++) {
    for (let a = -0.8; a < 0.8; a += 0.04) {
      addP(main, cx + Math.cos(a + Math.PI) * 28, cy + Math.sin(a + Math.PI) * 12 + 4 + i * 3, brt * 0.2);
    }
  }
}

function drawSpiralTile(_idx, x, y, t, isActive) {
  const oy = y + VISIBLE_TOP_INSET;
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

function drawTeethTile(_idx, x, y, t, isActive) {
  const oy = y + VISIBLE_TOP_INSET;
  const cx = x + 45;
  const cy = oy + 43;
  const brt = isActive ? 235 : 110;
  const gapOpen = 3 + Math.sin(t * 0.45) * 2;

  for (let a = -Math.PI; a < 0; a += 0.03) {
    const rx = 32;
    const ry = 8 + Math.sin(t * 0.7) * 2;
    addP(main, cx + Math.cos(a) * rx, cy - 6 + Math.sin(a) * ry, brt);
    addP(main, cx + Math.cos(a) * rx, cy + 6 + -Math.sin(a) * ry, brt);
  }

  main.fillRect(x + 12, cy - gapOpen + 1, 66, gapOpen * 2 - 1, 0);

  const teethW = 7;
  const teethCount = 8;
  for (let i = 0; i < teethCount; i++) {
    const tx = x + 9 + i * teethW + (i >= 4 ? 3 : 0);
    const th = 11 + Math.sin(i * 1.1 + t * 0.25) * 2;
    for (let dy = 0; dy < th; dy++) {
      const taper = 1 - dy / th * 0.25;
      const w = Math.max(3, (teethW * taper) | 0);
      for (let dx = 0; dx < w; dx++) {
        addP(main, tx + dx, cy - gapOpen - dy, brt * (0.65 + dy / th * 0.35));
      }
      addP(main, tx, cy - gapOpen - dy, brt);
      addP(main, tx + w - 1, cy - gapOpen - dy, brt);
    }
  }

  for (let i = 0; i < teethCount; i++) {
    const tx = x + 9 + i * teethW + (i >= 4 ? 3 : 0);
    const th = 9 + Math.sin(i * 0.8 + t * 0.3) * 2;
    for (let dy = 0; dy < th; dy++) {
      const taper = 1 - dy / th * 0.2;
      const w = Math.max(3, (teethW * taper) | 0);
      for (let dx = 0; dx < w; dx++) {
        addP(main, tx + dx, cy + gapOpen + dy, brt * (0.65 + dy / th * 0.35));
      }
      addP(main, tx, cy + gapOpen + dy, brt);
      addP(main, tx + w - 1, cy + gapOpen + dy, brt);
    }
  }

  main.crosshatch(x + 6, cy - gapOpen - 18, 78, 6, 2, brt * 0.12);
  main.crosshatch(x + 6, cy + gapOpen + 12, 78, 6, 2, brt * 0.12);
  lineH(main, x + 10, x + 80, cy - gapOpen - 1, brt * 0.18);
  lineH(main, x + 10, x + 80, cy + gapOpen + 1, brt * 0.18);
}

function renderMain() {
  main.clear(0);
  const t = phase.get() * Math.PI * 2;
  const activeIdx = active.get();
  for (let i = 0; i < 12; i++) {
    const { x, y } = tileRect(i);
    const isActive = i === activeIdx;
    main.fillRect(x, y, TILE, TILE, isActive ? 32 : 6);
    drawTileFrame(x, y, isActive);
    drawTileChrome(i, x, y, isActive);
    tiles[i].draw(i, x, y, t, isActive);
  }
}

function renderMainScan() {
  mainScan.clear(0);
  const t = phase.get();
  const activeIdx = active.get();
  if (!SHOW_SCAN_LAYER) return;
  for (let i = 0; i < 12; i++) {
    const { x, y } = tileRect(i);
    const localY = 20 + VISIBLE_TOP_INSET + ((Math.floor(t * 80) + i * 5) % (TILE - 28));
    mainScan.fillRect(x + 8, y + localY, TILE - 16, 1, i === activeIdx ? 80 : 14);
  }
  const sweepY = Math.floor(t * (MAIN_H - 6));
  mainScan.fillRect(0, sweepY, MAIN_W, 1, 10);
  const f = flash.get();
  if (f > 0) {
    const { x, y } = tileRect(activeIdx);
    const b = Math.floor(20 + f * 70);
    mainScan.fillRect(x + 3, y + 3, TILE - 6, 1, b);
    mainScan.fillRect(x + 3, y + TILE - 4, TILE - 6, 1, b);
  }
}

function renderMainRipple() {
  mainRipple.clear(0);
  if (!SHOW_RIPPLE_LAYER) return;
  const activeIdx = active.get();
  const p = ripple.get();
  if (p <= 0) return;
  const { x, y } = tileRect(activeIdx);
  const cx = x + 45;
  const cy = y + 45;
  const radius = 10 + Math.floor((1 - p) * 34);
  const bright = Math.floor(80 + p * 150);
  drawSpiral(mainRipple, cx, cy, 3, radius, bright, 0.7, phase.get() * Math.PI * 2, 1);
  mainRipple.line(cx - radius, cy, cx + radius, cy, bright);
  mainRipple.line(cx, cy - radius, cx, cy + radius, bright);
}

function renderMainHUD() {
  mainHUD.clear(0);
  if (!SHOW_HUD_LAYER) return;
  const activeIdx = active.get();
  const row = Math.floor(activeIdx / 4) + 1;
  const col = (activeIdx % 4) + 1;
  mainHUD.fillRect(92, 116, 176, 38, 0);
  mainHUD.line(92, 116, 267, 116, 255);
  mainHUD.line(92, 153, 267, 153, 255);
  mainHUD.line(92, 116, 92, 153, 255);
  mainHUD.line(267, 116, 267, 153, 255);
  drawText(mainHUD, `SEL ${tiles[activeIdx].key} R${row}C${col}`, 180, 120, 255, 162, 12);
  drawText(mainHUD, lastEvent.get(), 180, 134, 190, 162, 12);
}

function renderLeft() {
  left.clear(0);
  const t = phase.get() * Math.PI * 2;
  drawText(left, "B1", 30, 6, 220, 40, 12);
  drawText(left, "B2", 30, 20, 220, 40, 12);
  for (let seg = 0; seg < 10; seg++) {
    const sy = 40 + seg * 22;
    const sh = 18;
    const level = Math.sin(t * 1.2 + seg * 0.8) * 0.4 + 0.5;
    const fillH = Math.floor(level * sh);
    const brt = Math.floor(20 + level * 80);
    left.fillRect(4, sy, SIDE_W - 8, sh, 6);
    if (fillH > 0) {
      left.fillRect(8, sy + sh - fillH, SIDE_W - 16, fillH, brt);
    }
    left.line(2, sy + sh, SIDE_W - 3, sy + sh, 12);
  }
}

function renderRight() {
  right.clear(0);
  const off = scroll.get();
  for (let i = 0; i < 16; i++) {
    const y = ((i * 20 - (off % 20) + SIDE_H) % SIDE_H) - 20;
    if (y < -18 || y > SIDE_H) continue;
    const ci = (i + Math.floor(off / 20)) % sideText.length;
    right.text(sideText[ci], { x: 6, y, width: 48, height: 12, brightness: 50 + (i % 3) * 20, center: true });
  }
}

function renderAll() {
  renderMain();
  renderMainScan();
  renderMainRipple();
  renderMainHUD();
  renderLeft();
  renderRight();
}

ui.page("cyb-ito-proto", page => {
  page.display("left", display => {
    display.surface(left);
  });
  page.display("main", display => {
    display.surface(main);
    if (SHOW_SCAN_LAYER) display.layer("scan", mainScan);
    if (SHOW_RIPPLE_LAYER) display.layer("ripple", mainRipple);
    if (SHOW_HUD_LAYER) display.layer("hud", mainHUD);
  });
  page.display("right", display => {
    display.surface(right);
  });
});

function retrigger(signal, targetValue, durationMs, previousHandle) {
  if (previousHandle && previousHandle.stop) {
    previousHandle.stop();
  }
  signal.set(targetValue);
  return anim.to(signal, 0, durationMs);
}

function activate(idx, reason) {
  const next = clamp(idx, 0, 11);
  active.set(next);
  rippleHandle = retrigger(ripple, 1, 220, rippleHandle);
  flashHandle = retrigger(flash, 1, 140, flashHandle);
  lastEvent.set(`${reason} -> ${tiles[next].key}`);
  renderAll();
}

[
  "Touch1", "Touch2", "Touch3", "Touch4",
  "Touch5", "Touch6", "Touch7", "Touch8",
  "Touch9", "Touch10", "Touch11", "Touch12",
].forEach((name, idx) => {
  ui.onTouch(name, () => activate(idx, name));
});

ui.onButton("Button1", () => activate((active.get() + 11 + 12) % 12, "B1"));
ui.onButton("Button2", () => activate((active.get() + 1) % 12, "B2"));

activate(0, "BOOT");
anim.loop(1200, t => {
  phase.set(t);
  scroll.set((scroll.get() + 1) % (SIDE_H + 20));
  renderAll();
});

ui.show("cyb-ito-proto");
