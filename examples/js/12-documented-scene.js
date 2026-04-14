__package__({
  name: "documented",
  short: "Annotated loupedeck scene example"
});

__section__("display", {
  title: "Display settings",
  description: "Theme and refresh configuration for the demo scene",
  fields: {
    theme: {
      type: "choice",
      choices: ["dark", "light"],
      default: "dark",
      help: "Theme name stored in the configured scene"
    },
    refreshRate: {
      type: "integer",
      default: 30,
      help: "Requested refresh rate in Hz"
    }
  }
});

__doc__("configureScene", {
  summary: "Configure the documented demo scene",
  params: [
    { name: "title", type: "string", description: "Title shown on the main tile" },
    { name: "display", type: "object", description: "Display settings section values" },
    { name: "meta", type: "object", description: "Invocation metadata including rootDir" }
  ],
  returns: {
    type: "object",
    description: "Summary of the configured scene"
  }
});

__example__({
  id: "documented.configure",
  title: "Configure the documented scene",
  symbols: ["configureScene"],
  concepts: ["scene-setup", "section-binding"]
});

function configureScene(title, display, meta) {
  const ui = require("loupedeck/ui");

  ui.page("documented-home", page => {
    page.tile(0, 0, tile => {
      tile.text(title);
    });
    page.tile(1, 0, tile => {
      tile.text(`THEME ${String(display.theme || "dark").toUpperCase()}`);
    });
    page.tile(2, 0, tile => {
      tile.text(`HZ ${display.refreshRate || 30}`);
    });
    page.tile(3, 0, tile => {
      tile.text("DOC");
    });
  });

  ui.show("documented-home");

  return {
    page: "documented-home",
    title,
    theme: display.theme,
    refreshRate: display.refreshRate,
    rootDir: meta.rootDir
  };
}

__verb__("configureScene", {
  name: "configure",
  parents: ["documented"],
  sections: ["display"],
  fields: {
    title: {
      argument: true,
      help: "Title to show on the main tile"
    },
    display: {
      bind: "display"
    },
    meta: {
      bind: "context"
    }
  }
});

doc`---
symbol: configureScene
---
# Documented scene configuration

This example shows the intended loupedeck authoring pattern for annotated scenes:

- file-level metadata via \`__package__\`
- reusable settings via \`__section__\`
- executable entrypoints via \`__verb__("configureScene", ...)\`
- symbol docs via \`__doc__\`
- long-form prose via tagged \`doc\`...\`\`
`;
