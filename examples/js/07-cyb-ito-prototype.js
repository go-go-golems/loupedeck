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
const left = gfx.surface(SIDE_W, SIDE_H);
const right = gfx.surface(SIDE_W, SIDE_H);

const phase = state.signal(0);
const scroll = state.signal(0);
const active = state.signal(-1);

const titles = ["眼", "渦", "歯", "溶", "穴", "狂", "蟲", "砂", "歪", "裂", "脈", "闇"];
const subs = ["EYE", "SPIRAL", "TEETH", "MELT", "HOLE", "FACE", "WORM", "NOISE", "WARP", "CRACK", "PULSE", "VOID"];
const kanji = "呪螺旋恐怖闇影穴裂溶歪狂蝕腐朽這寄生喰渦巻沈黙叫骸".split("");

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
  const base = isActive ? 24 : 10;
  const border = isActive ? 120 : 32;
  const pulse = Math.floor((Math.sin(t * Math.PI * 2 + idx * 0.4) * 0.5 + 0.5) * 28);

  main.fillRect(x, y, TILE, TILE, base);
  main.line(x, y, x + TILE - 1, y, border);
  main.line(x, y + TILE - 1, x + TILE - 1, y + TILE - 1, border);
  main.line(x, y, x, y + TILE - 1, border);
  main.line(x + TILE - 1, y, x + TILE - 1, y + TILE - 1, border);

  main.line(x + 2, y + 13, x + TILE - 3, y + 13, isActive ? 40 : 18);
  main.crosshatch(x + 6, y + 20, TILE - 12, TILE - 28, isActive ? 3 : 4, 12 + pulse);

  const cx = x + 45;
  const cy = y + 48;
  const r = 10 + Math.floor(Math.sin(t * Math.PI * 2 * 0.7 + idx) * 4);
  main.line(cx - r, cy, cx + r, cy, 70 + pulse);
  main.line(cx, cy - r, cx, cy + r, 60 + pulse);
  main.fillRect(cx - 2, cy - 2, 4, 4, isActive ? 180 : 90 + pulse);

  const scanY = y + 20 + ((Math.floor(t * 80) + idx * 5) % (TILE - 28));
  main.fillRect(x + 3, scanY, TILE - 6, 1, isActive ? 150 : 60);

  main.text(titles[idx], { x: x + 4, y: y + 2, width: 22, height: 12, brightness: isActive ? 190 : 80 });
  main.text(subs[idx], { x: x + 40, y: y + 3, width: 46, height: 10, brightness: isActive ? 110 : 40, center: true });
}

function renderMain() {
  main.clear(0);
  const t = phase.get();
  const activeIdx = active.get();
  for (let i = 0; i < 12; i++) {
    drawTile(i, t, activeIdx);
  }
}

function renderLeft() {
  left.clear(0);
  const t = phase.get();
  for (let seg = 0; seg < 12; seg++) {
    const sy = 4 + seg * 22;
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
    const ci = (i + Math.floor(off / 20)) % kanji.length;
    right.text(kanji[ci], { x: 12, y, width: 32, height: 18, brightness: 50 + (i % 3) * 20, center: true });
  }
}

function renderAll() {
  renderMain();
  renderLeft();
  renderRight();
}

ui.page("cyb-ito-proto", page => {
  page.display("left", display => {
    display.surface(left);
  });
  page.display("main", display => {
    display.surface(main);
  });
  page.display("right", display => {
    display.surface(right);
  });
});

function activate(idx) {
  active.set(clamp(idx, 0, 11));
  renderAll();
}

[
  "Touch1", "Touch2", "Touch3", "Touch4",
  "Touch5", "Touch6", "Touch7", "Touch8",
  "Touch9", "Touch10", "Touch11", "Touch12",
].forEach((name, idx) => {
  ui.onTouch(name, () => activate(idx));
});

ui.onButton("Button1", () => activate((active.get() + 11 + 12) % 12));
ui.onButton("Button2", () => activate((active.get() + 1) % 12));

renderAll();
anim.loop(1200, t => {
  phase.set(t);
  scroll.set((scroll.get() + 1) % (SIDE_H + 20));
  renderAll();
});

ui.show("cyb-ito-proto");
