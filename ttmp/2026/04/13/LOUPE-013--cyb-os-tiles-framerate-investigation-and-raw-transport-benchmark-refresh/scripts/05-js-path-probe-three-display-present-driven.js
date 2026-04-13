const ui = require("loupedeck/ui");
const gfx = require("loupedeck/gfx");
const present = require("loupedeck/present");

const MAIN_W = 360;
const MAIN_H = 270;
const SIDE_W = 60;
const SIDE_H = 270;

const left = gfx.surface(SIDE_W, SIDE_H);
const main = gfx.surface(MAIN_W, MAIN_H);
const right = gfx.surface(SIDE_W, SIDE_H);

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

function drawLeft() {
  left.batch(() => {
    left.clear(0);
    for (let y = 0; y < SIDE_H; y += 20) left.fillRect(0, y, SIDE_W, 1, 12);
    const y = (frame * 7) % (SIDE_H + 20) - 20;
    left.fillRect(0, y, SIDE_W, 20, 140);
  });
}

function drawRight() {
  right.batch(() => {
    right.clear(0);
    for (let y = 10; y < SIDE_H; y += 24) right.fillRect(0, y, SIDE_W, 1, 12);
    const y = (frame * 5 + 40) % (SIDE_H + 20) - 20;
    right.fillRect(0, y, SIDE_W, 20, 140);
  });
}

ui.page("js-path-probe-three-display-present-driven", page => {
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

ui.show("js-path-probe-three-display-present-driven");
present.onFrame(() => {
  drawLeft();
  drawMain();
  drawRight();
  frame++;
  present.invalidate("next");
});
present.invalidate("initial");
