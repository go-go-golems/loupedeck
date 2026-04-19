__package__({
  name: 'knob-meter',
  short: 'Knob meter example scene'
});

function runScene() {
  const state = require("loupedeck/state");
  const ui = require("loupedeck/ui");

  const level = state.signal(50);

  function clamp(v, min, max) {
    return Math.max(min, Math.min(max, v));
  }

  ui.page("knob-meter", page => {
    page.tile(0, 0, tile => {
      tile.text("KNOB1");
    });

    page.tile(1, 0, tile => {
      tile.text(() => `LVL ${Math.round(level.get())}`);
    });

    page.tile(2, 0, tile => {
      tile.text(() => `${Math.round(level.get())}%`);
    });

    page.tile(3, 0, tile => {
      tile.text("TURN");
    });
  });

  ui.onKnob("Knob1", event => {
    level.update(v => clamp(v + event.value, 0, 100));
  });

  ui.show("knob-meter");
}

__verb__("runScene", {
  name: "run",
  short: 'Run the knob meter example scene'
});

if (typeof globalThis.__glazedVerbRegistry === "undefined") {
  runScene();
}
