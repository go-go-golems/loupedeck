package gfx

import (
	"image"
	"image/color"
	"image/draw"
)

type Surface struct {
	width  int
	height int
	pixels []uint8

	nextListenerID uint64
	listeners      map[uint64]func()
}

func NewSurface(width, height int) *Surface {
	if width <= 0 || height <= 0 {
		panic("gfx: surface dimensions must be positive")
	}
	return &Surface{
		width:  width,
		height: height,
		pixels: make([]uint8, width*height),
	}
}

func (s *Surface) Width() int {
	if s == nil {
		return 0
	}
	return s.width
}

func (s *Surface) Height() int {
	if s == nil {
		return 0
	}
	return s.height
}

func (s *Surface) Bounds() image.Rectangle {
	if s == nil {
		return image.Rectangle{}
	}
	return image.Rect(0, 0, s.width, s.height)
}

func (s *Surface) Clear(v uint8) {
	if s == nil {
		return
	}
	for i := range s.pixels {
		s.pixels[i] = v
	}
	s.notifyChanged()
}

func (s *Surface) At(x, y int) uint8 {
	if s == nil || !s.inBounds(x, y) {
		return 0
	}
	return s.pixels[y*s.width+x]
}

func (s *Surface) Set(x, y int, v uint8) {
	if s == nil || !s.inBounds(x, y) {
		return
	}
	s.pixels[y*s.width+x] = v
}

func (s *Surface) Add(x, y int, v uint8) {
	if s == nil || !s.inBounds(x, y) {
		return
	}
	i := y*s.width + x
	s.pixels[i] = saturatingAdd(s.pixels[i], v)
}

func (s *Surface) FillRect(x, y, width, height int, v uint8) {
	if s == nil || width <= 0 || height <= 0 {
		return
	}
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			s.Set(px, py, v)
		}
	}
	s.notifyChanged()
}

func (s *Surface) Line(x1, y1, x2, y2 int, v uint8) {
	if s == nil {
		return
	}
	dx := abs(x2 - x1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	dy := -abs(y2 - y1)
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx + dy
	for {
		s.Add(x1, y1, v)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
	s.notifyChanged()
}

func (s *Surface) Crosshatch(x, y, width, height, density int, v uint8) {
	if s == nil || width <= 0 || height <= 0 {
		return
	}
	if density <= 0 {
		density = 1
	}
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if (px+py)%density == 0 {
				s.Add(px, py, v)
			}
			if density < 4 && (px-py)%density == 0 {
				s.Add(px, py, v/2)
			}
		}
	}
	s.notifyChanged()
}

func (s *Surface) CompositeAdd(src *Surface, xoff, yoff int) {
	if s == nil || src == nil {
		return
	}
	for y := 0; y < src.height; y++ {
		for x := 0; x < src.width; x++ {
			v := src.At(x, y)
			if v == 0 {
				continue
			}
			s.Add(xoff+x, yoff+y, v)
		}
	}
	s.notifyChanged()
}

func (s *Surface) OnChange(fn func()) Subscription {
	if s == nil || fn == nil {
		return &surfaceSubscription{}
	}
	if s.listeners == nil {
		s.listeners = map[uint64]func(){}
	}
	s.nextListenerID++
	id := s.nextListenerID
	s.listeners[id] = fn
	return &surfaceSubscription{closeFn: func() {
		delete(s.listeners, id)
	}}
}

func (s *Surface) ToRGBA(fg, bg color.Color) *image.RGBA {
	if s == nil {
		return image.NewRGBA(image.Rectangle{})
	}
	im := image.NewRGBA(s.Bounds())
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	fr, fgc, fb, _ := fg.RGBA()
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			v := s.At(x, y)
			if v == 0 {
				continue
			}
			im.SetRGBA(x, y, color.RGBA{
				R: uint8(fr >> 8),
				G: uint8(fgc >> 8),
				B: uint8(fb >> 8),
				A: v,
			})
		}
	}
	return im
}

func (s *Surface) inBounds(x, y int) bool {
	return x >= 0 && x < s.width && y >= 0 && y < s.height
}

type Subscription interface {
	Close() error
}

type surfaceSubscription struct {
	closeFn func()
}

func (s *surfaceSubscription) Close() error {
	if s == nil || s.closeFn == nil {
		return nil
	}
	s.closeFn()
	s.closeFn = nil
	return nil
}

func (s *Surface) notifyChanged() {
	if s == nil || len(s.listeners) == 0 {
		return
	}
	for _, fn := range s.listeners {
		fn()
	}
}

func saturatingAdd(a, b uint8) uint8 {
	sum := int(a) + int(b)
	if sum > 255 {
		return 255
	}
	return uint8(sum)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
