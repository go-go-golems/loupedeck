const ui = require("loupedeck/ui");

ui.page("home", page => {
  page.tile(0, 0, tile => tile.text("HOME"));
  page.tile(1, 0, tile => tile.text("B1 -> ALT"));
  page.tile(2, 0, tile => tile.text("B2 -> INFO"));
  page.tile(3, 0, tile => tile.text("CIRCLE EXIT"));
});

ui.page("alt", page => {
  page.tile(0, 0, tile => tile.text("ALT"));
  page.tile(1, 0, tile => tile.text("B1 -> HOME"));
  page.tile(2, 0, tile => tile.text("B2 -> INFO"));
  page.tile(3, 0, tile => tile.text("PAGE 2"));
});

ui.page("info", page => {
  page.tile(0, 0, tile => tile.text("INFO"));
  page.tile(1, 0, tile => tile.text("B1 -> HOME"));
  page.tile(2, 0, tile => tile.text("B2 -> ALT"));
  page.tile(3, 0, tile => tile.text("PAGE 3"));
});

ui.onButton("Button1", () => {
  const current = globalThis.__page || "home";
  const next = current === "home" ? "alt" : "home";
  globalThis.__page = next;
  ui.show(next);
});

ui.onButton("Button2", () => {
  const current = globalThis.__page || "home";
  const next = current === "info" ? "alt" : "info";
  globalThis.__page = next;
  ui.show(next);
});

globalThis.__page = "home";
ui.show("home");
