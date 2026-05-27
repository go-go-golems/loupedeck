const easing = require("loupedeck/easing");
const gfx = require("loupedeck/gfx");
if (easing.linear(0.5) !== 0.5) throw new Error("bad linear easing");
if (typeof gfx.surface !== "function") throw new Error("missing gfx.surface");
const surface = gfx.surface(8, 8);
if (!surface) throw new Error("missing surface instance");
console.log("loupedeck provider ok");
