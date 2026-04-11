package loupedeck

import (
	"fmt"
	"html"
	"image"
	"image/draw"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type SVGIcon struct {
	Name string
	SVG  string
}

type SVGIconLibrary struct {
	Path      string
	Defs      string
	Variables map[string]string
	Icons     []SVGIcon
}

var (
	svgRootVariablesRe = regexp.MustCompile(`(?s):root\s*\{(.*?)\}`)
	svgVariableRe      = regexp.MustCompile(`(--[a-zA-Z0-9_-]+)\s*:\s*([^;]+);`)
	svgDefsRe          = regexp.MustCompile(`(?s)<svg[^>]*>\s*(<defs>.*?</defs>)\s*</svg>`)
	svgIconCellRe      = regexp.MustCompile(`(?s)<div class="icon-cell">\s*(<svg.*?</svg>)\s*<div class="icon-label">(.*?)</div>\s*</div>`)
	svgOpenTagRe       = regexp.MustCompile(`(?s)<svg\b([^>]*)>`)
	svgStyleAttrRe     = regexp.MustCompile(`style="([^"]*)"`)
	svgTagRe           = regexp.MustCompile(`<[^>]+>`)
)

func LoadSVGIconLibrary(path string) (*SVGIconLibrary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return loadSVGIconLibraryFromReader(path, f)
}

func loadSVGIconLibraryFromReader(path string, r io.Reader) (*SVGIconLibrary, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return parseSVGIconLibraryHTML(path, string(data))
}

func parseSVGIconLibraryHTML(path, source string) (*SVGIconLibrary, error) {
	vars := extractSVGVariables(source)
	defs := extractSVGDefs(source)
	matches := svgIconCellRe.FindAllStringSubmatch(source, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no icon cells found in %s", path)
	}

	icons := make([]SVGIcon, 0, len(matches))
	for _, m := range matches {
		rawSVG := m[1]
		label := normalizeSVGLabel(m[2])
		if label == "" {
			continue
		}
		icons = append(icons, SVGIcon{
			Name: label,
			SVG:  normalizeSVGFragment(rawSVG, defs, vars),
		})
	}
	if len(icons) == 0 {
		return nil, fmt.Errorf("no usable icons found in %s", path)
	}

	sort.Slice(icons, func(i, j int) bool {
		return icons[i].Name < icons[j].Name
	})

	return &SVGIconLibrary{
		Path:      path,
		Defs:      defs,
		Variables: vars,
		Icons:     icons,
	}, nil
}

func extractSVGVariables(source string) map[string]string {
	vars := map[string]string{}
	root := svgRootVariablesRe.FindStringSubmatch(source)
	if len(root) < 2 {
		return vars
	}
	for _, m := range svgVariableRe.FindAllStringSubmatch(root[1], -1) {
		vars[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
	}
	return vars
}

func extractSVGDefs(source string) string {
	m := svgDefsRe.FindStringSubmatch(source)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func normalizeSVGLabel(label string) string {
	label = html.UnescapeString(label)
	label = svgTagRe.ReplaceAllString(label, " ")
	label = strings.Join(strings.Fields(label), " ")
	return strings.TrimSpace(label)
}

func normalizeSVGFragment(rawSVG, defs string, vars map[string]string) string {
	svg := rawSVG
	for name, value := range vars {
		svg = strings.ReplaceAll(svg, "var("+name+")", value)
	}
	svg = strings.ReplaceAll(svg, `class="dither-50"`, `fill="url(#dither50)"`)
	svg = strings.ReplaceAll(svg, `class="dither-25"`, `fill="url(#dither25)"`)
	svg = stripSVGAnimationStyles(svg)
	if defs != "" && strings.Contains(svg, `url(#dither`) && !strings.Contains(svg, `<defs>`) {
		svg = svgOpenTagRe.ReplaceAllString(svg, `<svg$1 xmlns="http://www.w3.org/2000/svg">`+defs)
	} else if !strings.Contains(svg, `xmlns=`) {
		svg = svgOpenTagRe.ReplaceAllString(svg, `<svg$1 xmlns="http://www.w3.org/2000/svg">`)
	}
	return svg
}

func stripSVGAnimationStyles(svg string) string {
	return svgStyleAttrRe.ReplaceAllStringFunc(svg, func(attr string) string {
		m := svgStyleAttrRe.FindStringSubmatch(attr)
		if len(m) < 2 {
			return attr
		}
		decls := strings.Split(m[1], ";")
		kept := make([]string, 0, len(decls))
		for _, decl := range decls {
			decl = strings.TrimSpace(decl)
			if decl == "" {
				continue
			}
			parts := strings.SplitN(decl, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(strings.ToLower(parts[0]))
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(key, "animation") || key == "transform-origin" {
				continue
			}
			kept = append(kept, key+`: `+value)
		}
		if len(kept) == 0 {
			return ""
		}
		return `style="` + strings.Join(kept, "; ") + `"`
	})
}

func (icon SVGIcon) Rasterize(size int) (*image.RGBA, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid raster size %d", size)
	}
	i, err := oksvg.ReadIconStream(strings.NewReader(icon.SVG))
	if err != nil {
		return nil, fmt.Errorf("read SVG icon %q: %w", icon.Name, err)
	}
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	dasher := rasterx.NewDasher(size, size, scanner)
	i.SetTarget(0, 0, float64(size), float64(size))
	i.Draw(dasher, 1.0)
	return img, nil
}

func visibleBounds(img image.Image) image.Rectangle {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+1 > maxX {
				maxX = x + 1
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}
	if !found {
		return image.Rect(0, 0, 1, 1)
	}
	return image.Rect(minX, minY, maxX, maxY)
}

func cropImage(img image.Image, rect image.Rectangle) *image.RGBA {
	if rect.Empty() {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst
}

func fitRect(src image.Rectangle, maxWidth, maxHeight int) image.Rectangle {
	if maxWidth <= 0 || maxHeight <= 0 {
		return image.Rect(0, 0, 1, 1)
	}
	sw, sh := src.Dx(), src.Dy()
	if sw <= 0 || sh <= 0 {
		return image.Rect(0, 0, 1, 1)
	}
	scaleX := float64(maxWidth) / float64(sw)
	scaleY := float64(maxHeight) / float64(sh)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	if scale <= 0 {
		scale = 1
	}
	w := int(float64(sw) * scale)
	h := int(float64(sh) * scale)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return image.Rect(0, 0, w, h)
}
