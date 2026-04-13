const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const anim = require("loupedeck/anim");
const present = require("loupedeck/present");

const BS = 90;
const COLS = 4;
const ROWS = 3;
const MAIN_W = COLS * BS;
const MAIN_H = ROWS * BS;
const SIDE_W = 60;
const SIDE_H = MAIN_H;
const FULL_W = SIDE_W + MAIN_W + SIDE_W;
const CJK_FONT_PATH = "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc";

function loadOptionalFont(path, opts) {
  try {
    return gfx.font(path, opts);
  } catch (_err) {
    return null;
  }
}

const JP_8 = loadOptionalFont(CJK_FONT_PATH, { size: 8, dpi: 72, index: 0 });
const JP_10 = loadOptionalFont(CJK_FONT_PATH, { size: 10, dpi: 72, index: 0 });
const JP_11 = loadOptionalFont(CJK_FONT_PATH, { size: 11, dpi: 72, index: 0 });
const JP_14 = loadOptionalFont(CJK_FONT_PATH, { size: 14, dpi: 72, index: 0 });
const JP_18 = loadOptionalFont(CJK_FONT_PATH, { size: 18, dpi: 72, index: 0 });

const left = gfx.surface(SIDE_W, SIDE_H);
const main = gfx.surface(MAIN_W, MAIN_H);
const right = gfx.surface(SIDE_W, SIDE_H);

const loopPhase = state.signal(0);
let frame = 0;
let scrollOff = 0;
const ripples = [];

const kanjiStream = "電脳空間仮想現実神経接続量子演算暗号解読超伝導体結晶化人工知能深層学習機械翻訳自律制御情報処理回路接続侵入防止走査起動".split("");

function clamp(v, min, max) {
  return Math.max(min, Math.min(max, v));
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

function drawRect(surface, x, y, w, h, v) {
  lineH(surface, x, x + w - 1, y, v);
  lineH(surface, x, x + w - 1, y + h - 1, v);
  lineV(surface, x, y, y + h - 1, v);
  lineV(surface, x + w - 1, y, y + h - 1, v);
}

function drawText(surface, text, x, y, brightness, width, height, font) {
  const opts = { x, y, width, height, brightness, center: true };
  if (font) opts.font = font;
  surface.text(text, opts);
}

function circleOutline(surface, cx, cy, r, brightness, skipTopGap) {
  let x = r;
  let y = 0;
  let err = 1 - r;
  while (x >= y) {
    if (!(skipTopGap && y < 6 && Math.abs(x) < 4)) {
      addP(surface, cx + x, cy + y, brightness);
      addP(surface, cx - x, cy + y, brightness);
      addP(surface, cx + x, cy - y, brightness);
      addP(surface, cx - x, cy - y, brightness);
    }
    addP(surface, cx + y, cy + x, brightness);
    addP(surface, cx - y, cy + x, brightness);
    if (!(skipTopGap && x < 6 && Math.abs(y) < 4)) {
      addP(surface, cx + y, cy - x, brightness);
      addP(surface, cx - y, cy - x, brightness);
    }
    y++;
    if (err < 0) err += 2 * y + 1;
    else {
      x--;
      err += 2 * (y - x) + 1;
    }
  }
}

const tiles = [
  {
    title: "CLK", sub: "時計",
    draw(surface, ox, oy, t, active) {
      const d = new Date();
      const h = String(d.getHours()).padStart(2, "0");
      const m = String(d.getMinutes()).padStart(2, "0");
      const s = String(d.getSeconds()).padStart(2, "0");
      drawText(surface, h + ":" + m, ox + 8, oy + 22, active ? 220 : 90, 74, 24);
      drawText(surface, s, ox + 32, oy + 48, active ? 140 : 40, 26, 14);
      if (d.getSeconds() % 2 === 0) {
        for (let dy = 0; dy < 3; dy++) for (let dx = 0; dx < 3; dx++) addP(surface, ox + 24 + dx, oy + 50 + dy, active ? 180 : 60);
      }
      const days = ["日", "月", "火", "水", "木", "金", "土"];
      drawText(surface, days[d.getDay()] + "曜日", ox + 18, oy + 66, active ? 100 : 25, 54, 14, JP_11);
    },
  },
  {
    title: "CPU", sub: "処理",
    draw(surface, ox, oy, t, active) {
      for (let i = 0; i < 8; i++) {
        const level = Math.sin(t * 2 + i * 0.8) * 0.3 + Math.sin(t * 0.7 + i * 1.5) * 0.3 + 0.5;
        const bh = (level * 40) | 0;
        const brt = active ? 160 : ((20 + level * 50) | 0);
        const bx = ox + 8 + i * 10;
        for (let y = 0; y < bh; y++) {
          const py = oy + 68 - y;
          if ((bx + py) % 2 === 0) addP(surface, bx, py, brt);
          if ((bx + 1 + py) % 2 === 0) addP(surface, bx + 1, py, brt);
        }
        addP(surface, bx, oy + 68 - bh, brt + 20);
        addP(surface, bx + 1, oy + 68 - bh, brt + 20);
      }
      const pct = (((Math.sin(t * 1.3) * 30 + 50) | 0)) + "%";
      drawText(surface, pct, ox + 22, oy + 74, active ? 160 : 35, 46, 12);
    },
  },
  {
    title: "WAVE", sub: "波形",
    draw(surface, ox, oy, t, active) {
      const w = 76, h = 50, bx = ox + 7, by = oy + 18;
      drawRect(surface, bx, by, w, h, active ? 40 : 12);
      for (let x = bx; x < bx + w; x += 12) lineV(surface, x, by, by + h, 4);
      for (let y = by; y < by + h; y += 10) lineH(surface, bx, bx + w, y, 4);
      const brt = active ? 200 : 70;
      for (let x = 0; x < w; x++) {
        const val = Math.sin(x * 0.12 + t * 3) * 0.35 + Math.sin(x * 0.05 + t * 1.7) * 0.25 + Math.sin(x * 0.3 + t * 5) * 0.1;
        const py = (by + h / 2 + val * h * 0.4) | 0;
        addP(surface, bx + x, py, brt);
        addP(surface, bx + x, py + 1, (brt * 0.4) | 0);
        addP(surface, bx + x, py + 8, (brt * 0.15) | 0);
      }
    },
  },
  {
    title: "NET", sub: "神経", nodes: null,
    draw(surface, ox, oy, _t, active) {
      if (!this.nodes) {
        this.nodes = [];
        this.edges = [];
        const layers = [3, 5, 4, 2];
        let id = 0;
        layers.forEach((cnt, li) => {
          for (let i = 0; i < cnt; i++) {
            this.nodes.push({ x: 12 + li * 22, y: 44 - cnt * 7 + i * 14 + 7, layer: li, pulse: 0, id: id++ });
          }
        });
        this.nodes.forEach(n => {
          this.nodes.filter(m => m.layer === n.layer + 1).forEach(m => {
            this.edges.push({ a: n.id, b: m.id, fire: 0 });
          });
        });
      }
      if (frame % 20 === 0 || active) {
        const e = this.edges[(Math.random() * this.edges.length) | 0];
        e.fire = 1;
        this.nodes[e.b].pulse = 1;
      }
      this.edges.forEach(e => {
        const a = this.nodes[e.a], b = this.nodes[e.b];
        const brt = e.fire > 0 ? (active ? 180 : 80) : (active ? 18 : 6);
        const dx = b.x - a.x, dy = b.y - a.y;
        const steps = Math.max(Math.abs(dx), Math.abs(dy));
        for (let s = 0; s <= steps; s++) addP(surface, ox + a.x + (dx * s / steps) | 0, oy + a.y + (dy * s / steps) | 0, brt);
        if (e.fire > 0) e.fire -= 0.04;
      });
      this.nodes.forEach(n => {
        const brt = n.pulse > 0 ? (active ? 255 : 120) : (active ? 40 : 15);
        const s = n.pulse > 0 ? 2 : 1;
        for (let dy = -s; dy <= s; dy++) for (let dx = -s; dx <= s; dx++) addP(surface, ox + n.x + dx, oy + n.y + dy, brt);
        if (n.pulse > 0) n.pulse -= 0.03;
      });
    },
  },
  {
    title: "MEM", sub: "記憶", grid: null,
    draw(surface, ox, oy, _t, active) {
      if (!this.grid) {
        this.grid = new Uint8Array(64);
        for (let i = 0; i < 64; i++) this.grid[i] = Math.random() > 0.5 ? 1 : 0;
      }
      if (frame % 15 === 0) {
        const i = (Math.random() * 64) | 0;
        this.grid[i] = this.grid[i] ? 0 : 1;
      }
      const cs = 8, gap = 1, bx = ox + 9, by = oy + 18;
      for (let r = 0; r < 8; r++) for (let cl = 0; cl < 8; cl++) {
        const on = this.grid[r * 8 + cl];
        const brt = on ? (active ? 160 : 45) : (active ? 12 : 4);
        const px = bx + cl * (cs + gap), py = by + r * (cs + gap);
        for (let dy = 0; dy < cs; dy++) for (let dx = 0; dx < cs; dx++) {
          if (on || dx === 0 || dy === 0 || dx === cs - 1 || dy === cs - 1) addP(surface, px + dx, py + dy, brt);
          else addP(surface, px + dx, py + dy, 2);
        }
      }
      let used = 0;
      for (let i = 0; i < 64; i++) used += this.grid[i];
      drawText(surface, (((used / 64) * 100) | 0) + "%", ox + 28, oy + 76, active ? 120 : 30, 34, 12);
    },
  },
  {
    title: "TERM", sub: "端末", lines: ["> boot_seq OK", "> neural.init", "> 全回路接続", "> scan: 7node", "> ready_"],
    draw(surface, ox, oy, t, active) {
      surface.fillRect(ox + 4, oy + 16, 82, 66, active ? 6 : 2);
      if (frame % 60 === 0) {
        const msgs = ["> ping OK", "> 走査完了", "> mem: 64K", "> net: UP", "> 暗号解読中", "> node_7 ACK", "> データ受信", "> sys nominal"];
        this.lines.push(msgs[(Math.random() * msgs.length) | 0]);
        if (this.lines.length > 8) this.lines.shift();
      }
      const brt = active ? 180 : 55;
      this.lines.forEach((line, i) => {
        const ly = oy + 20 + i * 8;
        if (ly > oy + 78) return;
        const font = /[一-龯ぁ-んァ-ン]/.test(line) ? JP_8 : null;
        drawText(surface, line, ox + 4, ly, brt - (8 - i) * 3, 82, 9, font);
      });
      if (Math.sin(t * 6) > 0) {
        for (let dy = 0; dy < 7; dy++) for (let dx = 0; dx < 5; dx++) addP(surface, ox + 8 + dx, oy + 20 + this.lines.length * 8 + dy - 8, active ? 180 : 50);
      }
    },
  },
  {
    title: "RADAR", sub: "探知",
    draw(surface, ox, oy, t, active) {
      const cx = ox + 45, cy = oy + 44, r = 28;
      for (let ri = 10; ri <= r; ri += 9) circleOutline(surface, cx, cy, ri, active ? 20 : 7, false);
      lineH(surface, cx - r, cx + r, cy, active ? 15 : 5);
      lineV(surface, cx, cy - r, cy + r, active ? 15 : 5);
      const angle = t * 1.5;
      const brt = active ? 200 : 80;
      for (let i = 0; i < r; i++) {
        const sx = (Math.cos(angle) * i) | 0, sy = (Math.sin(angle) * i) | 0;
        addP(surface, cx + sx, cy + sy, brt);
        for (let tr = 1; tr < 6; tr++) {
          const ta = angle - tr * 0.08;
          const tsx = (Math.cos(ta) * i) | 0, tsy = (Math.sin(ta) * i) | 0;
          addP(surface, cx + tsx, cy + tsy, (brt * 0.6 / tr) | 0);
        }
      }
      [[12, -8], [-6, 15], [-15, -5], [8, 12]].forEach(([bx, by]) => {
        const bAngle = Math.atan2(by, bx);
        let diff = ((angle - bAngle) % (Math.PI * 2) + Math.PI * 2) % (Math.PI * 2);
        if (diff < 0.5) {
          const bb = ((1 - diff / 0.5) * 160) | 0;
          addP(surface, cx + bx, cy + by, bb);
          addP(surface, cx + bx + 1, cy + by, bb);
          addP(surface, cx + bx, cy + by + 1, bb);
        } else {
          addP(surface, cx + bx, cy + by, active ? 25 : 8);
        }
      });
    },
  },
  {
    title: "LOG", sub: "記録",
    draw(surface, ox, oy, _t, active) {
      const logKanji = "警告注意正常異常起動停止接続切断送信受信解析完了".split("");
      for (let i = 0; i < 5; i++) {
        const yOff = ((frame * 0.3 + i * 16) % 80) - 10;
        const ci = ((frame * 0.02) | 0 + i) % logKanji.length;
        const fade = 1 - Math.abs(yOff - 30) / 40;
        if (fade <= 0) continue;
        const brt = ((active ? 140 : 40) * Math.max(0, fade)) | 0;
        drawText(surface, logKanji[ci], ox + 12, oy + 16 + (yOff | 0), brt, 20, 20, JP_18);
        const dotBrt = ((active ? 100 : 25) * fade) | 0;
        addP(surface, ox + 40, oy + 24 + (yOff | 0), dotBrt);
        addP(surface, ox + 41, oy + 24 + (yOff | 0), dotBrt);
        const hex = ((ci * 17 + frame) & 0xFF).toString(16).toUpperCase().padStart(2, "0");
        drawText(surface, "0x" + hex, ox + 46, oy + 20 + (yOff | 0), (brt * 0.5) | 0, 36, 9);
      }
    },
  },
  {
    title: "SIG", sub: "信号",
    draw(surface, ox, oy, t, active) {
      for (let i = 0; i < 6; i++) {
        const maxH = 12 + i * 7;
        const level = Math.sin(t + i * 0.4) * 0.3 + 0.7;
        const h = (maxH * level) | 0;
        const brt = active ? 160 : ((30 + level * 40) | 0);
        const bx = ox + 12 + i * 12;
        for (let y = 0; y < h; y++) {
          const py = oy + 70 - y;
          for (let dx = 0; dx < 6; dx++) addP(surface, bx + dx, py, y < h - 1 ? brt : brt + 40);
        }
      }
      const db = ((Math.sin(t * 1.5) * 20 - 40) | 0);
      drawText(surface, db + "dB", ox + 22, oy + 76, active ? 140 : 30, 46, 10);
    },
  },
  {
    title: "DISK", sub: "記憶", angle: 0,
    draw(surface, ox, oy, t, active) {
      const cx = ox + 45, cy = oy + 42;
      circleOutline(surface, cx, cy, 22, active ? 80 : 25, false);
      circleOutline(surface, cx, cy, 8, active ? 80 : 25, false);
      this.angle += active ? 0.12 : 0.03;
      for (let s = 0; s < 4; s++) {
        const sa = this.angle + s * Math.PI / 2;
        for (let i = 10; i < 22; i++) {
          const sx = (Math.cos(sa) * i) | 0, sy = (Math.sin(sa) * i) | 0;
          addP(surface, cx + sx, cy + sy, active ? 120 : 35);
        }
      }
      for (let dy = -2; dy <= 2; dy++) for (let dx = -2; dx <= 2; dx++) addP(surface, cx + dx, cy + dy, active ? 200 : 50);
      if (Math.sin(t * 8) > 0.5) drawText(surface, "R/W", ox + 30, oy + 72, active ? 160 : 40, 30, 9);
    },
  },
  {
    title: "PWR", sub: "電源",
    draw(surface, ox, oy, t, active) {
      const cx = ox + 45, cy = oy + 38;
      circleOutline(surface, cx, cy, 18, active ? 180 : 50, true);
      lineV(surface, cx, cy - 20, cy, active ? 180 : 50);
      lineV(surface, cx + 1, cy - 20, cy, active ? 180 : 50);
      const pulse = Math.sin(t * 3) * 0.5 + 0.5;
      const statusBrt = active ? 200 : ((pulse * 60 + 20) | 0);
      drawText(surface, active ? "ONLINE" : "STANDBY", ox + 14, oy + 66, statusBrt, 62, 10);
      const barW = (pulse * 60) | 0;
      for (let i = 0; i < barW; i++) {
        addP(surface, ox + 15 + i, oy + 78, (statusBrt * 0.6) | 0);
        addP(surface, ox + 15 + i, oy + 79, (statusBrt * 0.3) | 0);
      }
    },
  },
  {
    title: "行列", sub: "MATRIX",
    draw(surface, ox, oy, _t, active) {
      const chars = "01電脳空仮現実".split("");
      for (let col = 0; col < 6; col++) {
        const speed = 0.8 + col * 0.3;
        const offset = (frame * speed * 0.05 + col * 7) % 12;
        for (let row = 0; row < 5; row++) {
          const yp = (oy + 16 + row * 14 - ((offset * 14) % 14)) | 0;
          if (yp < oy + 14 || yp > oy + 80) continue;
          const ci = (row + col + ((frame * speed * 0.02) | 0)) % chars.length;
          const dist = Math.abs(yp - (oy + 48)) / (BS / 2);
          const fade = Math.max(0, 1 - dist * 1.2);
          const brt = ((active ? 160 : 40) * fade) | 0;
          if (brt < 3) continue;
          const ch = chars[ci];
          const font = /[一-龯ぁ-んァ-ン]/.test(ch) ? JP_11 : null;
          drawText(surface, ch, ox + 8 + col * 13, yp, brt, 13, 13, font);
        }
      }
    },
  },
];

const tileState = tiles.map((_, i) => ({
  idx: i,
  col: i % COLS,
  row: (i / COLS) | 0,
  gx: (i % COLS) * BS,
  gy: ((i / COLS) | 0) * BS,
  flash: 0,
  active: false,
  scanY: 0,
  scanning: false,
}));

function addRipple(globalX, globalY, maxRadius) {
  ripples.push({ x: globalX, y: globalY, radius: 0, maxRadius });
}

function activateTile(idx) {
  const ts = tileState[idx];
  if (!ts) return;
  ts.flash = 1;
  ts.active = true;
  ts.scanning = true;
  ts.scanY = 0;
}

function setAllInactive() {
  tileState.forEach(ts => { ts.active = false; });
}

function updateAnimationState() {
  frame++;
  scrollOff += 0.5;
  tileState.forEach(ts => {
    ts.flash *= 0.88;
    if (ts.flash < 0.01) ts.flash = 0;
    if (ts.scanning) {
      ts.scanY += 4;
      if (ts.scanY > BS) {
        ts.scanning = false;
        ts.scanY = 0;
      }
    }
  });
  for (let i = ripples.length - 1; i >= 0; i--) {
    const rp = ripples[i];
    rp.radius += 5;
    if (rp.radius > rp.maxRadius) ripples.splice(i, 1);
  }
}

function drawMainRipples() {
  for (let i = 0; i < ripples.length; i++) {
    const rp = ripples[i];
    const cx = rp.x - SIDE_W;
    const cy = rp.y;
    const r = rp.radius | 0;
    const thick = Math.max(1, Math.min(3, 5 - ((rp.radius / 55) | 0)));
    const fade = Math.max(0, 1 - rp.radius / (rp.maxRadius * 0.55));
    const brt = (255 * fade) | 0;
    if (brt < 3) continue;
    const inner = Math.max(0, r - thick);
    const r2 = r * r;
    const ri2 = inner * inner;
    for (let y = Math.max(0, (cy - r) | 0); y <= Math.min(MAIN_H - 1, (cy + r) | 0); y++) {
      const dy = y - cy;
      const dy2 = dy * dy;
      if (dy2 > r2) continue;
      const xO = Math.sqrt(r2 - dy2);
      const xI = dy2 < ri2 ? Math.sqrt(ri2 - dy2) : 0;
      for (let x = Math.max(0, (cx - xO) | 0); x <= Math.min(MAIN_W - 1, (cx - xI) | 0); x++) addP(main, x, y, brt);
      for (let x = Math.max(0, (cx + xI) | 0); x <= Math.min(MAIN_W - 1, (cx + xO) | 0); x++) addP(main, x, y, brt);
    }
    if (rp.radius > 16) {
      circleOutline(main, cx, cy, (rp.radius - 14) | 0, (brt * 0.2) | 0, false);
    }
  }
}

function drawSideRipples(surface, xOffset, width) {
  for (let i = 0; i < ripples.length; i++) {
    const rp = ripples[i];
    const cx = rp.x - xOffset;
    const cy = rp.y;
    const r = rp.radius | 0;
    const fade = Math.max(0, 1 - rp.radius / (rp.maxRadius * 0.55));
    const brt = (160 * fade) | 0;
    if (brt < 3) continue;
    for (let a = 0; a < Math.PI * 2; a += 0.05) {
      addP(surface, cx + Math.cos(a) * r, cy + Math.sin(a) * r, brt);
    }
  }
}

function applyScanlines(surface, width, height) {
  for (let y = 0; y < height; y += 3) {
    for (let x = 0; x < width; x++) addP(surface, x, y, -3);
  }
}

function renderMain() {
  main.batch(() => {
    main.clear(0);
    const t = frame * 0.03;
    tileState.forEach((ts, idx) => {
      const ox = ts.gx, oy = ts.gy;
      const pulse = Math.sin(t + idx * 0.5) * 0.5 + 0.5;
      const border = ts.active ? 60 : ((6 + pulse * 8) | 0);
      drawRect(main, ox, oy, BS, BS, border);
      const cm = ts.active ? 140 : 20;
      for (let i = 0; i < 6; i++) {
        addP(main, ox + 1 + i, oy + 1, cm); addP(main, ox + 1, oy + 1 + i, cm);
        addP(main, ox + BS - 2 - i, oy + 1, cm); addP(main, ox + BS - 2, oy + 1 + i, cm);
        addP(main, ox + 1 + i, oy + BS - 2, cm); addP(main, ox + 1, oy + BS - 2 - i, cm);
        addP(main, ox + BS - 2 - i, oy + BS - 2, cm); addP(main, ox + BS - 2, oy + BS - 2 - i, cm);
      }
      const tile = tiles[idx];
      const subFont = /[一-龯ぁ-んァ-ン]/.test(tile.sub) ? JP_8 : null;
      drawText(main, tile.title, ox + 4, oy + 3, ts.active ? 180 : 35, 30, 9);
      drawText(main, tile.sub, ox + 56, oy + 3, ts.active ? 120 : 18, 30, 9, subFont);
      lineH(main, ox + 2, ox + BS - 3, oy + 13, ts.active ? 40 : 8);
      tile.draw(main, ox, oy, t, ts.active || ts.flash > 0.1);
      if (ts.scanning) {
        const sy = (oy + ts.scanY) | 0;
        if (sy > 0 && sy < MAIN_H) {
          lineH(main, ox + 1, ox + BS - 2, sy, 180);
          lineH(main, ox + 1, ox + BS - 2, sy + 1, 60);
        }
      }
      if (ts.flash > 0) {
        const fb = (ts.flash * 60) | 0;
        main.fillRect(ox + 1, oy + 1, BS - 2, BS - 2, fb);
      }
    });
    drawMainRipples();
    applyScanlines(main, MAIN_W, MAIN_H);
  });
}

function renderLeft() {
  left.batch(() => {
    left.clear(0);
    lineV(left, SIDE_W - 1, 0, SIDE_H - 1, 15);
    const t = frame * 0.03;
    for (let seg = 0; seg < 12; seg++) {
      const sy = 4 + seg * 22, sh = 18;
      const level = Math.sin(t * 1.8 + seg * 0.6) * 0.5 + 0.5;
      const fillH = (level * sh) | 0;
      const brt = (15 + level * 55) | 0;
      for (let y = 0; y < fillH; y++) {
        const py = sy + sh - 1 - y;
        for (let x = 4; x < SIDE_W - 4; x++) if ((x + py) % 2 === 0) addP(left, x, py, brt);
      }
    }
    tileState.forEach(ts => {
      if (ts.flash > 0.05) {
        const py = (ts.gy + BS / 2) | 0;
        const b = (ts.flash * 200) | 0;
        for (let dy = -1; dy <= 1; dy++) for (let dx = -1; dx <= 1; dx++) addP(left, SIDE_W - 6 + dx, py + dy, b);
      }
    });
    drawSideRipples(left, 0, SIDE_W);
    applyScanlines(left, SIDE_W, SIDE_H);
  });
}

function renderRight() {
  right.batch(() => {
    right.clear(0);
    lineV(right, 0, 0, SIDE_H - 1, 15);
    for (let i = 0; i < 16; i++) {
      const yp = ((i * 20 - (scrollOff % 20) + SIDE_H) % SIDE_H) - 20;
      if (yp < -20 || yp > SIDE_H) continue;
      const ci = (i + Math.floor(scrollOff / 20)) % kanjiStream.length;
      const dist = Math.abs(yp + 10 - SIDE_H / 2) / (SIDE_H / 2);
      const fade = Math.max(0, 1 - dist * 1.3);
      const brt = (fade * 40 + 5) | 0;
      drawText(right, kanjiStream[ci], 6, yp | 0, brt, 30, 18, JP_14);
    }
    tileState.forEach(ts => {
      if (ts.flash > 0.05) {
        const py = (ts.gy + BS / 2) | 0;
        const b = (ts.flash * 200) | 0;
        for (let dy = -1; dy <= 1; dy++) for (let dx = -1; dx <= 1; dx++) addP(right, 5 + dx, py + dy, b);
      }
    });
    drawSideRipples(right, SIDE_W + MAIN_W, SIDE_W);
    applyScanlines(right, SIDE_W, SIDE_H);
  });
}

function renderAll() {
  renderMain();
  renderLeft();
  renderRight();
}

ui.page("cyb-os-tiles", page => {
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

for (let i = 1; i <= 12; i++) {
  const idx = i - 1;
  ui.onTouch(`Touch${i}`, event => {
    if (event.status === "down") {
      activateTile(idx);
      addRipple(event.x, event.y, Math.max(FULL_W, SIDE_H) * 1.1);
    } else if (event.status === "up") {
      setAllInactive();
    }
    present.invalidate(`touch-${i}-${event.status}`);
  });
}

ui.show("cyb-os-tiles");
present.onFrame(() => {
  renderAll();
});
present.invalidate("initial");
anim.loop(2000, t => {
  loopPhase.set(t);
  updateAnimationState();
  present.invalidate("loop");
});
