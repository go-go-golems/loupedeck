__package__({
  name: 'cyb-ito-tile-port-first3',
  short: 'CYB ITO first three tiles example scene'
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

  const eye = gfx.surface(TILE, TILE);
  const spiral = gfx.surface(TILE, TILE);
  const teeth = gfx.surface(TILE, TILE);

  const phase = state.signal(0);
  const active = state.signal(0);
  const lastEvent = state.signal("BOOT");

  function addP(surface, x, y, v) {
    surface.add(x | 0, y | 0, Math.max(0, Math.min(255, v | 0)));
  }

  function lineH(surface, x1, x2, y, v) {
    for (let x = x1; x <= x2; x++) addP(surface, x, y, v);
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
    surface.line(0, 0, 89, 0, border);
    surface.line(0, 89, 89, 89, border);
    surface.line(0, 0, 0, 89, border);
    surface.line(89, 0, 89, 89, border);
    drawText(surface, label, 45, LABEL_Y, isActive ? 170 : 60, 72, LABEL_H);
    lineH(surface, 2, 87, DIVIDER_Y, isActive ? 26 : 6);
  }

  function renderEye(t, isActive) {
    frame(eye, "EYE", isActive);
    const cx = 45;
    const cy = TOP + 41;
    const brt = isActive ? 210 : 86;

    for (let a = 0; a < Math.PI * 2; a += 0.02) {
      const rx = 30 * Math.cos(a);
      const py = Math.sin(a) * 12;
      addP(eye, cx + rx, cy + py, brt);
      addP(eye, cx + rx * 1.04, cy + py * 1.1, brt * 0.4);
    }

    const irisR = 11 + Math.sin(t * 0.5) * 2;
    for (let a = 0; a < Math.PI * 2; a += 0.03) {
      addP(eye, cx + Math.cos(a) * irisR, cy + Math.sin(a) * irisR, brt);
    }

    const pupilR = 5 + Math.sin(t * 1.5) * 2;
    for (let dy = -pupilR; dy <= pupilR; dy++) {
      for (let dx = -pupilR; dx <= pupilR; dx++) {
        if (dx * dx + dy * dy <= pupilR * pupilR) addP(eye, cx + dx, cy + dy, brt * 0.9);
      }
    }

    addP(eye, cx - 2, cy - 2, 255);
    addP(eye, cx - 3, cy - 2, 255);
    addP(eye, cx - 2, cy - 3, 255);

    for (let i = 0; i < 8; i++) {
      const a = i * Math.PI / 4 + Math.sin(t * 0.3 + i) * 0.2;
      for (let r = irisR + 1; r < irisR + 8 + Math.sin(t + i) * 3; r++) {
        const wx = Math.sin(r * 0.5 + i) * 0.8;
        addP(eye, cx + Math.cos(a) * r + wx, cy + Math.sin(a) * r * 0.5, brt * 0.3 * (1 - r / 25));
      }
    }

    eye.crosshatch(4, TOP + 16, 18, 50, 3, isActive ? 20 : 8);
    eye.crosshatch(68, TOP + 16, 18, 50, 3, isActive ? 20 : 8);
  }

  function renderSpiral(t, isActive) {
    frame(spiral, "SPIR", isActive);
    const cx = 45;
    const cy = TOP + 43;
    const brt = isActive ? 190 : 76;
    drawSpiral(spiral, cx, cy, 6, 32, brt, 0.5, t, 2);
    drawSpiral(spiral, cx, cy, 4, 12, brt * 1.2, 0.85, t, 1);
    drawSpiral(spiral, 14, TOP + 20, 2, 6, brt * 0.3, 1.0, t, 1);
    drawSpiral(spiral, 76, TOP + 70, 2, 6, brt * 0.3, -0.7, t, 1);
    for (let r = 8; r < 34; r += 7) {
      const wobble = Math.sin(r * 0.5 + t) * 3;
      for (let a = 0; a < Math.PI * 2; a += 0.05) {
        const wr = r + Math.sin(a * 3 + t) * wobble;
        addP(spiral, cx + Math.cos(a) * wr, cy + Math.sin(a) * wr, brt * 0.12);
      }
    }
  }

  function renderTeeth(t, isActive) {
    frame(teeth, "TEETH", isActive);
    const cx = 45;
    const cy = TOP + 43;
    const brt = isActive ? 235 : 110;
    const gapOpen = 3 + Math.sin(t * 0.45) * 2;

    for (let a = -Math.PI; a < 0; a += 0.03) {
      const rx = 32;
      const ry = 8 + Math.sin(t * 0.7) * 2;
      addP(teeth, cx + Math.cos(a) * rx, cy - 6 + Math.sin(a) * ry, brt);
      addP(teeth, cx + Math.cos(a) * rx, cy + 6 - Math.sin(a) * ry, brt);
    }

    teeth.fillRect(12, cy - gapOpen + 1, 66, gapOpen * 2 - 1, 0);

    const teethW = 7;
    const teethCount = 8;
    for (let i = 0; i < teethCount; i++) {
      const tx = 9 + i * teethW + (i >= 4 ? 3 : 0);
      const th = 11 + Math.sin(i * 1.1 + t * 0.25) * 2;
      for (let dy = 0; dy < th; dy++) {
        const taper = 1 - dy / th * 0.25;
        const w = Math.max(3, (teethW * taper) | 0);
        for (let dx = 0; dx < w; dx++) {
          addP(teeth, tx + dx, cy - gapOpen - dy, brt * (0.65 + dy / th * 0.35));
        }
        addP(teeth, tx, cy - gapOpen - dy, brt);
        addP(teeth, tx + w - 1, cy - gapOpen - dy, brt);
      }
    }

    for (let i = 0; i < teethCount; i++) {
      const tx = 9 + i * teethW + (i >= 4 ? 3 : 0);
      const th = 9 + Math.sin(i * 0.8 + t * 0.3) * 2;
      for (let dy = 0; dy < th; dy++) {
        const taper = 1 - dy / th * 0.2;
        const w = Math.max(3, (teethW * taper) | 0);
        for (let dx = 0; dx < w; dx++) {
          addP(teeth, tx + dx, cy + gapOpen + dy, brt * (0.65 + dy / th * 0.35));
        }
        addP(teeth, tx, cy + gapOpen + dy, brt);
        addP(teeth, tx + w - 1, cy + gapOpen + dy, brt);
      }
    }

    teeth.crosshatch(6, cy - gapOpen - 18, 78, 6, 2, brt * 0.12);
    teeth.crosshatch(6, cy + gapOpen + 12, 78, 6, 2, brt * 0.12);
    lineH(teeth, 10, 80, cy - gapOpen - 1, brt * 0.18);
    lineH(teeth, 10, 80, cy + gapOpen + 1, brt * 0.18);
  }

  function renderAll() {
    const t = phase.get() * Math.PI * 2;
    const a = active.get();
    renderEye(t, a === 0);
    renderSpiral(t, a === 1);
    renderTeeth(t, a === 2);
  }

  function setActive(idx, why) {
    active.set(idx);
    lastEvent.set(why);
    renderAll();
  }

  ui.page("tiles-first3", page => {
    page.tile(0, 0, tile => tile.surface(eye));
    page.tile(1, 0, tile => tile.surface(spiral));
    page.tile(2, 0, tile => tile.surface(teeth));
    page.tile(3, 0, tile => tile.text(() => lastEvent.get()));
    page.tile(0, 1, tile => tile.text("TOUCH1"));
    page.tile(1, 1, tile => tile.text("TOUCH2"));
    page.tile(2, 1, tile => tile.text("TOUCH3"));
    page.tile(3, 1, tile => tile.text("B1/B2"));
  });

  ui.onTouch("Touch1", () => setActive(0, "T1"));
  ui.onTouch("Touch2", () => setActive(1, "T2"));
  ui.onTouch("Touch3", () => setActive(2, "T3"));
  ui.onButton("Button1", () => setActive((active.get() + 2) % 3, "B1"));
  ui.onButton("Button2", () => setActive((active.get() + 1) % 3, "B2"));

  renderAll();
  anim.loop(1200, t => {
    phase.set(t);
    renderAll();
  });

  ui.show("tiles-first3");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the CYB ITO first three tiles example scene'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
