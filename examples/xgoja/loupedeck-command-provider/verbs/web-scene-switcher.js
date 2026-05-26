__section__("switcher", {
  title: "Web scene switcher",
  fields: {
    out: { type: "string", default: "./dist/web-scene", help: "Output directory for the scene report" },
    wait_ms: { type: "int", default: 5000, help: "How long to wait for /deal before failing" },
  },
});

async function webSceneSwitcher(switcher) {
  const express = require("express");
  const fs = require("fs");
  const timer = require("timer");
  const easing = require("loupedeck/easing");
  const gfx = require("loupedeck/gfx");

  fs.mkdirSync(switcher.out, { recursive: true });
  const reportPath = `${switcher.out.replace(/\/$/, "")}/scene-report.md`;
  const markerPath = `${switcher.out.replace(/\/$/, "")}/dealt.txt`;

  const surface = gfx.surface(24, 24);
  const eased = easing.inOutQuad(0.5);
  let scene = "waiting";
  let dealt = false;

  const app = express.app();
  app.get("/", (req, res) => {
    res.type("html").send(`<!doctype html>
<html><body>
  <h1>Loupedeck web scene smoke</h1>
  <p id="scene">scene: ${scene}</p>
  <form method="post" action="/deal"><button>Deal with this page</button></form>
</body></html>`);
  });
  app.get("/state", (req, res) => res.json({ scene, dealt, eased, hasSurface: !!surface }));
  app.post("/deal", (req, res) => {
    scene = "dealt";
    dealt = true;
    fs.writeFileSync(markerPath, "dealt\n", "utf8");
    res.json({ ok: true, scene });
  });

  const deadline = Date.now() + Number(switcher.wait_ms || 5000);
  while (!dealt && Date.now() < deadline) {
    await timer.sleep(50);
  }
  if (!dealt) {
    throw new Error("timed out waiting for /deal");
  }

  fs.writeFileSync(reportPath, [
    "# Loupedeck Web Scene Smoke",
    "",
    `- Final scene: ${scene}`,
    `- Eased midpoint: ${eased}`,
    `- Surface allocated: ${!!surface}`,
    "",
  ].join("\n"), "utf8");

  return { scene, dealt, report_path: reportPath, marker_path: markerPath };
}

__verb__("webSceneSwitcher", {
  name: "web-scene-switcher",
  short: "Open an Express page and switch scenes when /deal is posted",
  sections: ["switcher"],
  fields: { switcher: { bind: "switcher" } },
});
