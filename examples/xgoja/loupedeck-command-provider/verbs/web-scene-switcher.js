__section__("switcher", {
  title: "Web scene switcher",
  fields: {
    out: { type: "string", default: "./dist/web-scene", help: "Output directory for the scene report" },
    wait_ms: { type: "int", default: 0, help: "How long to wait for /deal before exiting; 0 keeps the hardware/web UI running" },
    exit_on_deal: { type: "bool", default: false, help: "Exit after /deal is posted (useful for smoke tests)" },
  },
});

async function webSceneSwitcher(switcher) {
  const express = require("express");
  const fs = require("fs");
  const timer = require("timer");
  const state = require("loupedeck/state");
  const ui = require("loupedeck/ui");
  const easing = require("loupedeck/easing");
  const gfx = require("loupedeck/gfx");

  const outDir = switcher.out || switcher.outDir || "./dist/web-scene";
  const waitMs = Number(switcher.wait_ms ?? switcher.waitMs ?? switcher["wait-ms"] ?? 0);
  const exitOnDeal = Boolean(switcher.exit_on_deal ?? switcher.exitOnDeal ?? switcher["exit-on-deal"] ?? false);

  fs.mkdirSync(outDir, { recursive: true });
  const baseOut = outDir.replace(/\/$/, "");
  const reportPath = `${baseOut}/scene-report.md`;
  const markerPath = `${baseOut}/dealt.txt`;
  const logPath = `${baseOut}/scene-debug.log`;
  function log(message, details) {
    const line = `[${new Date().toISOString()}] ${message}${details ? " " + JSON.stringify(details) : ""}`;
    console.log(line);
    fs.appendFileSync(logPath, line + "\n", "utf8");
  }
  log("webSceneSwitcher starting", { outDir, waitMs, exitOnDeal });

  const scene = state.signal("waiting");
  const dealt = state.signal(false);
  const lastEvent = state.signal("open http://127.0.0.1:8791/");
  const dealCount = state.signal(0);
  const pulse = state.signal(0);
  const surface = gfx.surface(24, 24);

  function repaint(reason) {
    if (typeof ui.invalidate === "function") {
      ui.invalidate(reason || "scene-state-change");
    }
  }

  function setScene(next, event) {
    log("setScene", { next, event });
    scene.set(next);
    lastEvent.set(event || next);
  }

  function deal(source) {
    log("deal", { source });
    dealt.set(true);
    dealCount.update(v => v + 1);
    setScene("dealt", `${source} dealt page`);
    fs.writeFileSync(markerPath, `dealt by ${source}\n`, "utf8");
    repaint("deal");
  }

  function reset(source) {
    log("reset", { source });
    dealt.set(false);
    setScene("waiting", `${source} reset scene`);
    repaint("reset");
  }

  log("creating hardware UI page");
  ui.page("web-switcher", page => {
    log("configuring tile 0,0");
    page.tile(0, 0, tile => {
      tile.text(() => scene.get() === "dealt" ? "DEALT" : "WAIT");
    });
    log("configuring tile 1,0");
    page.tile(1, 0, tile => {
      tile.text(() => `COUNT ${dealCount.get()}`);
    });
    log("configuring tile 2,0");
    page.tile(2, 0, tile => {
      tile.text("WEB");
    });
    log("configuring tile 3,0");
    page.tile(3, 0, tile => {
      tile.text(() => `PULSE ${Math.round(easing.inOutQuad(pulse.get()) * 100)}%`);
    });
    log("configuring tile 0,1");
    page.tile(0, 1, tile => {
      tile.text("B1 DEAL");
    });
    log("configuring tile 1,1");
    page.tile(1, 1, tile => {
      tile.text("B2 RESET");
    });
    log("configuring tile 2,1");
    page.tile(2, 1, tile => {
      tile.text("TOUCH DEAL");
    });
    log("configuring tile 3,1");
    page.tile(3, 1, tile => {
      tile.text(() => dealt.get() ? "OK" : "POST");
    });
    log("configuring left display");
    page.display("left", display => {
      display.text(() => `Scene: ${scene.get()}`);
    });
    log("configuring right display");
    page.display("right", display => {
      display.text(() => lastEvent.get());
    });
  });
  log("hardware UI page configured");

  log("registering hardware handlers");
  ui.onButton("Button1", () => deal("hardware Button1"));
  ui.onButton("Button2", () => reset("hardware Button2"));
  ui.onTouch("Touch1", () => deal("hardware Touch1"));
  ui.show("web-switcher");
  log("hardware UI page shown");

  log("registering web routes");
  const app = express.app();
  app.get("/", (req, res) => {
    res.type("html").send(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Loupedeck web scene switcher</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 2rem; background: #111827; color: #f9fafb; }
    .card { max-width: 42rem; padding: 1.5rem; border-radius: 1rem; background: #1f2937; box-shadow: 0 10px 30px #0008; }
    button { font-size: 1.1rem; padding: .75rem 1rem; margin-right: .5rem; border: 0; border-radius: .5rem; cursor: pointer; }
    .deal { background: #10b981; color: #052e1b; }
    .reset { background: #f59e0b; color: #3b2500; }
    code { color: #93c5fd; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Loupedeck + web UI</h1>
    <p>The web page and the hardware are backed by the same JS state.</p>
    <p><strong>Scene:</strong> <span id="scene">${scene.get()}</span></p>
    <p><strong>Dealt:</strong> <span id="dealt">${dealt.get()}</span></p>
    <p><strong>Deal count:</strong> <span id="count">${dealCount.get()}</span></p>
    <p><strong>Last event:</strong> <span id="event">${lastEvent.get()}</span></p>
    <button class="deal" id="deal-button" type="button">Deal with this page</button>
    <button class="reset" id="reset-button" type="button">Reset</button>
    <p>Hardware shortcuts: <code>Button1</code> deals, <code>Button2</code> resets, <code>Touch1</code> deals.</p>
  </div>
  <script>
    async function refresh() {
      const s = await fetch('/state').then(r => r.json());
      document.getElementById('scene').textContent = s.scene;
      document.getElementById('dealt').textContent = s.dealt;
      document.getElementById('count').textContent = s.dealCount;
      document.getElementById('event').textContent = s.lastEvent;
    }
    async function postAction(path) {
      await fetch(path, { method: 'POST' });
      await refresh();
    }
    document.getElementById('deal-button').addEventListener('click', () => postAction('/deal'));
    document.getElementById('reset-button').addEventListener('click', () => postAction('/reset'));
    setInterval(refresh, 500);
  </script>
</body>
</html>`);
  });
  app.get("/state", (req, res) => res.json({
    scene: scene.get(),
    dealt: dealt.get(),
    dealCount: dealCount.get(),
    lastEvent: lastEvent.get(),
    eased: easing.inOutQuad(pulse.get()),
    hasSurface: !!surface,
  }));
  app.post("/deal", (req, res) => {
    deal("web /deal");
    res.json({ ok: true, scene: scene.get(), dealt: dealt.get(), dealCount: dealCount.get() });
  });
  app.post("/reset", (req, res) => {
    reset("web /reset");
    res.json({ ok: true, scene: scene.get(), dealt: dealt.get(), dealCount: dealCount.get() });
  });

  log("web routes registered; entering main loop");
  const start = Date.now();
  while (true) {
    pulse.set(((Date.now() - start) % 2000) / 2000);
    repaint("pulse");
    if (dealt.get() && exitOnDeal) {
      break;
    }
    if (waitMs > 0 && Date.now() - start >= waitMs) {
      break;
    }
    await timer.sleep(50);
  }

  log("main loop exiting", { scene: scene.get(), dealt: dealt.get(), dealCount: dealCount.get() });
  fs.writeFileSync(reportPath, [
    "# Loupedeck Web Scene Smoke",
    "",
    `- Final scene: ${scene.get()}`,
    `- Dealt: ${dealt.get()}`,
    `- Deal count: ${dealCount.get()}`,
    `- Last event: ${lastEvent.get()}`,
    `- Surface allocated: ${!!surface}`,
    "",
  ].join("\n"), "utf8");

  return { scene: scene.get(), dealt: dealt.get(), deal_count: dealCount.get(), report_path: reportPath, marker_path: markerPath };
}

__verb__("webSceneSwitcher", {
  name: "web-scene-switcher",
  short: "Open a web page and drive the real Loupedeck UI from the same JS state",
  sections: ["switcher"],
  fields: { switcher: { bind: "switcher" } },
});
