const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const present = require("loupedeck/present");

const MAIN_W = 360;
const MAIN_H = 270;
const main = gfx.surface(MAIN_W, MAIN_H);

let frame = 0;

function drawMain() {
  main.batch(() => {
    main.clear(0);
    for (let y = 0; y < MAIN_H; y += 18) main.fillRect(0, y, MAIN_W, 1, 10);
    for (let x = 0; x < MAIN_W; x += 24) main.fillRect(x, 0, 1, MAIN_H, 8);
    const barX = (frame * 9) % (MAIN_W + 24) - 24;
    main.fillRect(barX, 0, 24, MAIN_H, 180);
    for (let i = 0; i < 6; i++) {
      const x = (frame * (4 + i) + i * 41) % (MAIN_W + 14) - 14;
      const y = (i * 42 + ((frame * (2 + i)) % 20)) % MAIN_H;
      main.fillRect(x, y, 14, 14, 60 + i * 18);
    }
  });
}

ui.page("js-path-probe-main-only-present-driven", page => {
  page.display("main", display => {
    display.surface(main);
  });
});

ui.show("js-path-probe-main-only-present-driven");
present.onFrame(() => {
  drawMain();
  frame++;
  present.invalidate("next");
});
present.invalidate("initial");
