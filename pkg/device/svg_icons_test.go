package device

import (
	"strings"
	"testing"
)

const sampleSVGHTML = `
<!DOCTYPE html>
<html>
<head>
<style>
:root {
  --white: #f0f0e8;
  --black: #1a1a1a;
}
</style>
</head>
<body>
<svg width="0" height="0" style="position:absolute">
  <defs>
    <pattern id="dither50" width="2" height="2" patternUnits="userSpaceOnUse"></pattern>
    <pattern id="dither25" width="4" height="4" patternUnits="userSpaceOnUse"></pattern>
  </defs>
</svg>
<div class="icon-cell">
  <svg viewBox="0 0 48 48">
    <rect x="4" y="4" width="40" height="40" fill="var(--white)" stroke="var(--black)" stroke-width="2"/>
  </svg>
  <div class="icon-label">Finder</div>
</div>
<div class="icon-cell">
  <svg viewBox="0 0 48 48">
    <rect x="8" y="8" width="20" height="20" class="dither-50" style="animation: blink1 1s step-end infinite; transform-origin: 24px 24px"/>
  </svg>
  <div class="icon-label">Trash</div>
</div>
</body>
</html>
`

func TestParseSVGIconLibraryHTML(t *testing.T) {
	lib, err := parseSVGIconLibraryHTML("sample.html", sampleSVGHTML)
	if err != nil {
		t.Fatalf("parseSVGIconLibraryHTML: %v", err)
	}
	if got, want := len(lib.Icons), 2; got != want {
		t.Fatalf("icon count = %d, want %d", got, want)
	}
	if lib.Variables["--white"] != "#f0f0e8" {
		t.Fatalf("missing white variable: %#v", lib.Variables)
	}
	finder := lib.Icons[0]
	trash := lib.Icons[1]
	if !strings.Contains(finder.SVG, `fill="#f0f0e8"`) {
		t.Fatalf("finder SVG did not inline white fill: %s", finder.SVG)
	}
	if !strings.Contains(finder.SVG, `stroke="#1a1a1a"`) {
		t.Fatalf("finder SVG did not inline black stroke: %s", finder.SVG)
	}
	if !strings.Contains(trash.SVG, `fill="url(#dither50)"`) {
		t.Fatalf("trash SVG did not expand dither fill: %s", trash.SVG)
	}
	if !strings.Contains(trash.SVG, `<defs>`) {
		t.Fatalf("trash SVG did not inject defs: %s", trash.SVG)
	}
	if strings.Contains(trash.SVG, `animation:`) {
		t.Fatalf("trash SVG still contains animation style: %s", trash.SVG)
	}
}

func TestSVGIconRasterizeAndVisibleBounds(t *testing.T) {
	lib, err := parseSVGIconLibraryHTML("sample.html", sampleSVGHTML)
	if err != nil {
		t.Fatalf("parseSVGIconLibraryHTML: %v", err)
	}
	img, err := lib.Icons[0].Rasterize(48)
	if err != nil {
		t.Fatalf("Rasterize: %v", err)
	}
	bounds := visibleBounds(img)
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("visible bounds empty: %v", bounds)
	}
	cropped := cropImage(img, bounds)
	if cropped.Bounds().Dx() > 48 || cropped.Bounds().Dy() > 48 {
		t.Fatalf("cropped bounds unexpectedly large: %v", cropped.Bounds())
	}
	fit := fitRect(cropped.Bounds(), 70, 50)
	if fit.Dx() > 70 || fit.Dy() > 50 {
		t.Fatalf("fit rect exceeds target: %v", fit)
	}
}
