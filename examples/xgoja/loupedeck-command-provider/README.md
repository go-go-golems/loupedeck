# loupedeck xgoja command-provider hardware + web demo

This example builds a generated xgoja binary that mounts `loupedeck.scenes` as
`loupe` and runs an annotated scene verb that can drive **both**:

- a browser UI served by the xgoja `express` module; and
- a real Loupedeck device through `require("loupedeck/ui")` and
  `require("loupedeck/state")`.

The web page and the hardware page share the same JavaScript state. Pressing the
web `/deal` button updates the Loupedeck tiles/displays; pressing hardware
`Button1` or `Touch1` updates the web state.

## Hardware run

Build the generated binary and run the interactive scene:

```bash
make hardware
```

Equivalent direct command:

```bash
make build
./dist/loupedeck-command-provider loupe web-scene-switcher web-scene-switcher \
  --deck-enabled=true \
  --http-listen 127.0.0.1:8791 \
  --out ./dist/web-scene-hardware
```

Then open:

```text
http://127.0.0.1:8791/
```

Hardware controls:

- `Button1`: deal the page
- `Button2`: reset to waiting
- `Touch1`: deal the page

The command stays alive by default (`--wait-ms 0`). Stop it with `Ctrl-C`.

If auto-detect does not find the device, pass the serial path:

```bash
./dist/loupedeck-command-provider loupe web-scene-switcher web-scene-switcher \
  --deck-device /dev/ttyACM0 \
  --http-listen 127.0.0.1:8791
```

## Hardware test target

`make test-hardware` keeps hardware enabled, posts to `/deal`, exits after the
web action, and asserts that marker/report files were written. Use this when a
real Loupedeck is connected:

```bash
make test-hardware
```

If auto-detect is not enough:

```bash
make test-hardware DECK_DEVICE=/dev/ttyACM0
```

## Headless smoke

`make smoke` keeps hardware disabled so CI/local validation can still run without
a connected device:

```bash
make smoke
```

The smoke posts to `/deal`, exits with `--exit-on-deal=true`, and asserts that
`dealt.txt` plus `scene-report.md` were written.
