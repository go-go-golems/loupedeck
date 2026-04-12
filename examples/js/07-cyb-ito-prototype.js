const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");

const TILE = 90;
const MAIN_W = 360;
const MAIN_H = 270;
const SIDE_W = 60;
const SIDE_H = 270;

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

const titles = ["01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"];
const subs = ["EYE", "SPIN", "TEETH", "MELT", "HOLE", "FACE", "WORM", "NOISE", "WARP", "CRACK", "PULSE", "VOID"];
const sideText = ["EYE", "SPIN", "TEETH", "MELT", "HOLE", "FACE", "WORM", "NOISE", "WARP", "CRACK", "PULSE", "VOID", "TOUCH", "B1", "B2", "LIVE"];

function clamp(v, min, max) {
  return Math.max(min, Math.min(max, v));
}

function tileRect(idx) {
  const col = idx % 4;
  const row = Math.floor(idx / 4);
  return { col, row, x: col * TILE, y: row * TILE };
}

function drawTile(idx, t, activeIdx) {
  const { x, y } = tileRect(idx);
  const isActive = idx === activeIdx;
  const pulse = Math.floor((Math.sin(t * Math.PI * 2 + idx * 0.4) * 0.5 + 0.5) * 28);

  main.fillRect(x, y, TILE, TILE, isActive ? 72 : 8);
  main.fillRect(x + 6, y + 6, TILE - 12, TILE - 12, isActive ? 8 : 0);
  main.line(x, y, x + TILE - 1, y, isActive ? 255 : 24);
  main.line(x, y + TILE - 1, x + TILE - 1, y + TILE - 1, isActive ? 255 : 24);
  main.line(x, y, x, y + TILE - 1, isActive ? 255 : 24);
  main.line(x + TILE - 1, y, x + TILE - 1, y + TILE - 1, isActive ? 255 : 24);
  main.line(x + 1, y + 1, x + TILE - 2, y + 1, isActive ? 180 : 0);
  main.line(x + 1, y + TILE - 2, x + TILE - 2, y + TILE - 2, isActive ? 180 : 0);
  main.line(x + 1, y + 1, x + 1, y + TILE - 2, isActive ? 180 : 0);
  main.line(x + TILE - 2, y + 1, x + TILE - 2, y + TILE - 2, isActive ? 180 : 0);

  main.crosshatch(x + 10, y + 18, TILE - 20, TILE - 34, isActive ? 5 : 7, isActive ? 28 + pulse : 10 + pulse);

  const cx = x + 45;
  const cy = y + 48;
  const r = isActive ? 18 : 10 + Math.floor(Math.sin(t * Math.PI * 2 * 0.7 + idx) * 4);
  main.line(cx - r, cy, cx + r, cy, isActive ? 230 : 70 + pulse);
  main.line(cx, cy - r, cx, cy + r, isActive ? 230 : 60 + pulse);
  main.fillRect(cx - (isActive ? 5 : 2), cy - (isActive ? 5 : 2), isActive ? 10 : 4, isActive ? 10 : 4, isActive ? 255 : 90 + pulse);

  main.text(titles[idx], { x: x + 6, y: y + 4, width: 24, height: 12, brightness: isActive ? 255 : 90, center: true });
  main.text(subs[idx], { x: x + 28, y: y + 4, width: 54, height: 12, brightness: isActive ? 220 : 60, center: true });
  if (isActive) {
    main.text("ACTIVE", { x: x + 12, y: y + 66, width: 66, height: 12, brightness: 255, center: true });
  }
}

function renderMain() {
  main.clear(0);
  const t = phase.get();
  const activeIdx = active.get();
  for (let i = 0; i < 12; i++) {
    drawTile(i, t, activeIdx);
  }
}

function renderMainScan() {
  mainScan.clear(0);
  const t = phase.get();
  const activeIdx = active.get();
  for (let i = 0; i < 12; i++) {
    const { x, y } = tileRect(i);
    const localY = 20 + ((Math.floor(t * 80) + i * 5) % (TILE - 28));
    mainScan.fillRect(x + 8, y + localY, TILE - 16, 2, i === activeIdx ? 180 : 40);
  }
  const sweepY = Math.floor(t * (MAIN_H - 6));
  mainScan.fillRect(0, sweepY, MAIN_W, 2, 24);
  const f = flash.get();
  if (f > 0) {
    const { x, y } = tileRect(activeIdx);
    const b = Math.floor(70 + f * 185);
    mainScan.fillRect(x + 3, y + 3, TILE - 6, 2, b);
    mainScan.fillRect(x + 3, y + TILE - 5, TILE - 6, 2, b);
    mainScan.fillRect(x + 3, y + 3, 2, TILE - 6, b);
    mainScan.fillRect(x + TILE - 5, y + 3, 2, TILE - 6, b);
  }
}

function renderMainRipple() {
  mainRipple.clear(0);
  const activeIdx = active.get();
  const p = ripple.get();
  if (p <= 0) return;
  const { x, y } = tileRect(activeIdx);
  const cx = x + 45;
  const cy = y + 45;
  const radius = 10 + Math.floor((1 - p) * 34);
  const bright = Math.floor(80 + p * 150);
  mainRipple.line(cx - radius, cy, cx + radius, cy, bright);
  mainRipple.line(cx, cy - radius, cx, cy + radius, bright);
  mainRipple.line(cx - radius, cy - radius, cx + radius, cy + radius, Math.floor(bright * 0.5));
  mainRipple.line(cx - radius, cy + radius, cx + radius, cy - radius, Math.floor(bright * 0.5));
  mainRipple.fillRect(x + 4, y + 4, TILE - 8, 3, Math.floor(bright * 0.6));
}

function renderMainHUD() {
  mainHUD.clear(0);
  const activeIdx = active.get();
  const row = Math.floor(activeIdx / 4) + 1;
  const col = (activeIdx % 4) + 1;
  mainHUD.fillRect(96, 118, 168, 34, 0);
  mainHUD.line(96, 118, 263, 118, 255);
  mainHUD.line(96, 151, 263, 151, 255);
  mainHUD.line(96, 118, 96, 151, 255);
  mainHUD.line(263, 118, 263, 151, 255);
  mainHUD.text(`SEL ${titles[activeIdx]} R${row}C${col}`, { x: 102, y: 121, width: 156, height: 12, brightness: 255, center: true });
  mainHUD.text(lastEvent.get(), { x: 102, y: 135, width: 156, height: 12, brightness: 190, center: true });
  mainHUD.text("TOUCH/B1/B2", { x: 108, y: 152, width: 144, height: 12, brightness: 120, center: true });
}

function renderLeft() {
  left.clear(0);
  const t = phase.get();
  left.text("B1", { x: 8, y: 2, width: 44, height: 12, brightness: 220, center: true });
  left.text("B2", { x: 8, y: 16, width: 44, height: 12, brightness: 220, center: true });
  for (let seg = 0; seg < 10; seg++) {
    const sy = 40 + seg * 22;
    const sh = 18;
    const level = Math.sin(t * Math.PI * 2 * 1.2 + seg * 0.8) * 0.4 + 0.5;
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
    display.layer("scan", mainScan);
    display.layer("ripple", mainRipple);
    display.layer("hud", mainHUD);
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
  lastEvent.set(`${reason} -> ${titles[next]}`);
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
