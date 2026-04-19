package gfx

import (
	"image"
	"image/color"
	"image/draw"

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
}

func (s *Surface) Text(text string, opts TextOptions) {
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
