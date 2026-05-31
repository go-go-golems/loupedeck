package gfx

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type TextOptions struct {
	X          int
	Y          int
	Width      int
	Height     int
	Brightness uint8
	Face       font.Face
	Center     bool
	LineGap    int // extra vertical pixels between lines when text contains \n
	WrapWidth  int // if > 0, wrap text to this pixel width (measured with the font face)
}

func (s *Surface) Text(text string, opts TextOptions) {
	if s == nil || text == "" {
		return
	}

	// Expand text: first apply word wrapping, then split on newlines.
	lines := expandTextLines(text, opts)
	if len(lines) == 1 {
		s.renderLine(lines[0], opts)
		return
	}

	// Multi-line rendering: compute per-line height and lay out vertically.
	face := opts.Face
	if face == nil {
		face = basicfont.Face7x13
	}
	lineH := face.Metrics().Height.Ceil()
	gap := opts.LineGap
	if gap < 0 {
		gap = 0
	}

	for i, line := range lines {
		lineOpts := opts
		lineOpts.Y = opts.Y + i*(lineH+gap)
		lineOpts.Height = lineH + 4 // per-line alpha mask height
		lineOpts.WrapWidth = 0     // already expanded, don't re-wrap
		if line == "" {
			continue
		}
		s.renderLine(line, lineOpts)
	}
}

// expandTextLines expands text by first applying word wrapping (if WrapWidth is set),
// then splitting on newlines. The result is a flat list of single lines ready for
// renderLine().
func expandTextLines(text string, opts TextOptions) []string {
	face := opts.Face
	if face == nil {
		face = basicfont.Face7x13
	}

	// Split on explicit newlines first.
	paragraphs := splitLines(text)

	if opts.WrapWidth <= 0 {
		return paragraphs
	}

	// Apply word wrapping within each paragraph.
	var result []string
	for _, para := range paragraphs {
		if para == "" {
			result = append(result, "")
			continue
		}
		wrapped := wrapText(para, face, opts.WrapWidth)
		result = append(result, wrapped...)
	}
	return result
}

// wrapText wraps a single paragraph of text to fit within wrapWidth pixels,
// measured using the given font face. Returns one or more lines.
func wrapText(text string, face font.Face, wrapWidth int) []string {
	if text == "" || wrapWidth <= 0 {
		return []string{text}
	}

	d := &font.Drawer{Face: face}
	var lines []string
	var current strings.Builder

	words := strings.Fields(text)
	for i, word := range words {
		if i == 0 {
			current.WriteString(word)
			continue
		}
		candidate := current.String() + " " + word
		if d.MeasureString(candidate).Round() <= wrapWidth {
			current.WriteString(" ")
			current.WriteString(word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	if len(lines) == 0 {
		return []string{text}
	}
	return lines
}

// splitLines splits text on \n, preserving empty lines so that
// blank lines still take up vertical space in multi-line layout.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

// renderLine draws a single line of text onto the surface.
func (s *Surface) renderLine(text string, opts TextOptions) {
	if s == nil || text == "" {
		return
	}
	face := opts.Face
	if face == nil {
		face = basicfont.Face7x13
	}
	w := opts.Width
	if w <= 0 {
		w = s.width
	}
	h := opts.Height
	if h <= 0 {
		h = face.Metrics().Height.Ceil() + 4
	}
	brightness := opts.Brightness
	if brightness == 0 {
		brightness = 255
	}

	alpha := image.NewAlpha(image.Rect(0, 0, w, h))
	draw.Draw(alpha, alpha.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
	d := &font.Drawer{
		Dst:  alpha,
		Src:  image.White,
		Face: face,
	}
	baseline := h / 2
	if baseline < face.Metrics().Ascent.Ceil() {
		baseline = face.Metrics().Ascent.Ceil()
	}
	maxBaseline := h - face.Metrics().Descent.Ceil()
	if maxBaseline < 0 {
		maxBaseline = 0
	}
	if baseline > maxBaseline {
		baseline = maxBaseline
	}
	x := 0
	if opts.Center {
		x = (w - d.MeasureString(text).Round()) / 2
		if x < 0 {
			x = 0
		}
	}
	d.Dot = fixed.P(x, baseline)
	d.DrawString(text)

	s.mu.Lock()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			a := alpha.AlphaAt(x, y).A
			if a == 0 {
				continue
			}
			px := opts.X + x
			py := opts.Y + y
			if !s.inBounds(px, py) {
				continue
			}
			v := scaleUint8(a, brightness)
			s.addLocked(px, py, v)
		}
	}
	listeners := s.markChangedLocked()
	s.mu.Unlock()
	notifyListeners(listeners)
}
