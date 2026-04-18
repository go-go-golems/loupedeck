---
Title: Annotated scene scripts with jsverbs and jsdoc
Slug: loupedeck-jsverbs-scenes
Short: Use annotated JavaScript scene files with __verb__, __section__, __doc__, and the dynamic `loupedeck verbs ...` command tree.
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
- verbs-repository
- duration
- device
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

Loupedeck scene scripts can now be authored in two distinct styles:

1. **plain runtime scripts** executed directly with `loupedeck run <file.js>`
2. **annotated scene scripts** exposed as first-class CLI commands under `loupedeck verbs ...`

That split is intentional.

- `run` is the plain-file runner.
- `verbs` is the annotated-scene runner.
- `doc` remains the extraction/export surface.

## Plain scripts versus annotated scripts

Use `run` for ordinary scene files:

```bash
loupedeck run ./examples/js/01-hello.js --duration 5s
```

Use `verbs` when the script declares explicit jsverbs metadata:

```bash
loupedeck verbs documented configure OPS --theme light --duration 5s
```

The old transitional `run --verb` / `verbs list` / `verbs help` flow has been removed. Annotated commands now live directly in the command tree.

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
- plain `run` remains filename-oriented and is not the annotated-scene UX.

## Built-in and external repositories

The `verbs` tree is built from repositories discovered before command registration.

### Built-in repository

Loupedeck always includes one embedded built-in repository. That means the documented example command is always available:

```bash
loupedeck verbs documented configure OPS
```

### External repositories

You can add more repositories in three ways:

1. app config
2. environment variable
3. repeated CLI flags

### App config

Supported app-config locations come from the standard Glazed config-plan app config sources:

- `/etc/loupedeck/config.yaml`
- `$XDG_CONFIG_HOME/loupedeck/config.yaml`
- `~/.loupedeck/config.yaml`

Repository config shape:

```yaml
verbs:
  repositories:
    - name: team-scenes
      path: ~/code/acme/loupedeck-scenes
    - name: local-scenes
      path: ~/.loupedeck/verbs
```

### Environment variable

```bash
export LOUPEDECK_VERB_REPOSITORIES=/path/to/repo-a:/path/to/repo-b
```

### CLI override

```bash
loupedeck --verbs-repository ./examples/js --verbs-repository ~/.loupedeck/verbs verbs documented configure OPS
```

Repository precedence is:

1. embedded built-in repository
2. app-config repositories
3. `LOUPEDECK_VERB_REPOSITORIES`
4. repeated `--verbs-repository`

Duplicate repository paths are deduped. Duplicate full jsverb command paths across repositories are a hard startup error.

## Running an annotated scene command

Once a repository is loaded, explicit verbs appear as normal nested commands:

```bash
loupedeck verbs documented configure OPS --theme light --refreshRate 60 --duration 5s
```

The generated command help comes from jsverbs metadata, while the session/device flags come from loupedeck:

- verb fields such as `title`, `theme`, and `refreshRate`
- session fields such as `--device`, `--duration`, `--flush-interval`, and `--queue-size`

Inspect help directly on the generated command:

```bash
loupedeck verbs documented configure --help
```

## Extract docs

Documentation extraction remains file-oriented:

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

That file is the canonical loupedeck example for:

- `__package__`
- `__section__`
- `__verb__`
- `__doc__`
- `doc\`...\``
- context binding
- section binding
- the built-in annotated verbs repository
