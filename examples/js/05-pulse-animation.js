const state = require("loupedeck/state");
const ui = require("loupedeck/ui");
const anim = require("loupedeck/anim");
const easing = require("loupedeck/easing");

const pulse = state.signal(0);

ui.page("pulse", page => {
  page.tile(0, 0, tile => {
    tile.text("PULSE");
  });
  page.tile(1, 0, tile => {
    tile.text(() => `${Math.round(easing.inOutCubic(pulse.get()) * 100)}%`);
  });
  page.tile(2, 0, tile => {
    tile.text("LOOP");
  });
  page.tile(3, 0, tile => {
    tile.text("RUN");
  });
});

anim.loop(1200, t => {
  pulse.set(t);
});

ui.show("pulse");
