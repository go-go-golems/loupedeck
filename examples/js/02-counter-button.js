const state = require("loupedeck/state");
const ui = require("loupedeck/ui");

const count = state.signal(0);

ui.page("counter", page => {
  page.tile(0, 0, tile => {
    tile.icon("circle");
    tile.text(() => `COUNT ${count.get()}`);
  });

  page.tile(1, 0, tile => {
    tile.text("CIRCLE");
  });

  page.tile(2, 0, tile => {
    tile.text("TO");
  });

  page.tile(3, 0, tile => {
    tile.text("INC");
  });
});

ui.onButton("Circle", () => {
  count.update(v => v + 1);
});

ui.show("counter");
