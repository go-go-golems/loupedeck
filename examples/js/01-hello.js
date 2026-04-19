__package__({
  name: 'hello',
  short: 'Hello example scene'
});

function runScene() {
  const ui = require("loupedeck/ui");

  ui.page("hello", page => {
    page.tile(0, 0, tile => {
      tile.icon("hello");
      tile.text("HELLO");
    });

    page.tile(1, 0, tile => {
      tile.text("LOUPE");
    });

    page.tile(2, 0, tile => {
      tile.text("DECK");
    });

    page.tile(3, 0, tile => {
      tile.text("JS");
    });
  });

  ui.show("hello");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the hello example scene'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
