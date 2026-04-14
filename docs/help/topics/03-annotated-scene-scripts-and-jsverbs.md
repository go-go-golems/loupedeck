---
Title: Annotated scene scripts with jsverbs and jsdoc
Slug: loupedeck-jsverbs-scenes
Short: Use annotated JavaScript scene files with __verb__, __section__, __doc__, and loupedeck's verb/doc commands.
Topics:
- loupedeck
- javascript
- goja
- documentation
Commands:
- loupedeck run
- loupedeck verbs
- loupedeck doc
Flags:
- script
- verb
- verb-config
- verb-values-json
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

Loupedeck scene scripts can now be authored in two styles:

1. **plain runtime scripts** that execute directly via `loupedeck run --script ...`
2. **annotated scripts** that declare verbs and docs using jsverbs/jsdoc metadata

Annotated scripts are useful when you want one file to expose named scene entrypoints, documented parameters, reusable configuration sections, and extractable reference docs.

## Minimal annotated pattern

```js
__package__({ name: "documented" });

__section__("display", {
  fields: {
    theme: { type: "choice", choices: ["dark", "light"], default: "dark" }
  }
});

__doc__("configureScene", {
  summary: "Configure the scene"
});

function configureScene(title, display, meta) {
  const ui = require("loupedeck/ui");
  ui.page("home", page => {
    page.tile(0, 0, tile => tile.text(title));
  });
  ui.show("home");
  return { title, theme: display.theme, rootDir: meta.rootDir };
}

__verb__("configureScene", {
  name: "configure",
  parents: ["documented"],
  sections: ["display"],
  fields: {
    title: { argument: true },
    display: { bind: "display" },
    meta: { bind: "context" }
  }
});

doc`---
symbol: configureScene
---
Long-form prose for the scene.`;
```

Important details:

- `__verb__` takes a **string** function name, not an identifier reference.
- metadata must stay declarative/static.
- raw `loupedeck run --script ...` remains available for plain non-annotated scripts.

## Run an annotated scene verb

Use the normal hardware run path with `--verb` when you want a specific annotated entrypoint:

```bash
loupedeck run \
  --script ./examples/js/12-documented-scene.js \
  --verb "documented configure" \
  --verb-values-json '{"default":{"title":"OPS"},"display":{"theme":"light","refreshRate":60}}'
```

You can also provide one or more config files:

```bash
loupedeck run \
  --script ./examples/js/12-documented-scene.js \
  --verb "documented configure" \
  --verb-config ./scene-values.yaml
```

## Discover available verbs

List explicit verbs discovered for a scene file or directory:

```bash
loupedeck verbs list --script ./examples/js/12-documented-scene.js
```

## Inspect generated help/flags

Render the generated Glazed/Cobra help for a verb:

```bash
loupedeck verbs help \
  --script ./examples/js/12-documented-scene.js \
  --verb "documented configure"
```

This is the easiest way to inspect the generated flags for section fields and bound parameters.

## Extract docs

Export jsdoc/jsdocex metadata as JSON:

```bash
loupedeck doc --script ./examples/js/12-documented-scene.js --format json
```

or Markdown:

```bash
loupedeck doc --script ./examples/js/12-documented-scene.js --format markdown
```

## Reference example

See:

- `examples/js/12-documented-scene.js`

That file is intended to be the canonical loupedeck example for:

- `__package__`
- `__section__`
- `__verb__`
- `__doc__`
- `doc\`...\``
- context binding
- section binding
