# loupedeck xgoja command-provider smoke

This example builds a generated xgoja binary that mounts `loupedeck.scenes` as
`loupe` and runs an annotated scene verb without hardware.

The verb creates a tiny Express app, exposes `/state` and `/deal`, switches its
scene state when `/deal` is posted, and writes `dist/web-scene/scene-report.md`.

Run:

```bash
make smoke
```
