__package__({
  name: 'touch-feedback',
  short: 'Touch feedback example scene'
});

function runScene() {
  const state = require("loupedeck/state");
  const ui = require("loupedeck/ui");

  const last = state.signal("NONE");

  ui.page("touch", page => {
    page.tile(0, 0, tile => {
      tile.text("TOUCH1");
    });

    page.tile(3, 0, tile => {
      tile.text(() => last.get());
    });

    page.tile(1, 1, tile => {
      tile.text("TOUCH6");
    });

    page.tile(3, 2, tile => {
      tile.text("TOUCH12");
    });
  });

  ui.onTouch("Touch1", () => last.set("T1"));
  ui.onTouch("Touch6", () => last.set("T6"));
  ui.onTouch("Touch12", () => last.set("T12"));

  ui.show("touch");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the touch feedback example scene'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
