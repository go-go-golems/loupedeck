package gfx

import (
	"image"
	"image/color"
	"image/draw"
	"sync"
)

type Surface struct {
	width  int
	height int
	pixels []uint8

	mu             sync.Mutex
	cond           *sync.Cond
	batchDepth     int
	changedInBatch bool
	nextListenerID uint64
	listeners      map[uint64]func()
}

func NewSurface(width, height int) *Surface {
	if width <= 0 || height <= 0 {
		panic("gfx: surface dimensions must be positive")
	}
	s := &Surface{
		width:  width,
		height: height,
		pixels: make([]uint8, width*height),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
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

func (s *Surface) Batch(fn func()) {
	if s == nil || fn == nil {
		return
	}
	s.mu.Lock()
	s.batchDepth++
	s.mu.Unlock()

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		fn()
	}()

	listeners := s.endBatch()
	notifyListeners(listeners)
	if recovered != nil {
		panic(recovered)
	}
}

func (s *Surface) Clear(v uint8) {
	if s == nil {
		return
	}
	s.mu.Lock()
	for i := range s.pixels {
		s.pixels[i] = v
	}
	listeners := s.markChangedLocked()
	s.mu.Unlock()
	notifyListeners(listeners)
}

func (s *Surface) At(x, y int) uint8 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitForStableLocked()
	if !s.inBounds(x, y) {
		return 0
	}
	return s.pixels[y*s.width+x]
}

func (s *Surface) Set(x, y int, v uint8) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.inBounds(x, y) {
		s.setLocked(x, y, v)
	}
	s.mu.Unlock()
}

func (s *Surface) Add(x, y int, v uint8) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.inBounds(x, y) {
		s.addLocked(x, y, v)
	}
	s.mu.Unlock()
}

func (s *Surface) FillRect(x, y, width, height int, v uint8) {
	if s == nil || width <= 0 || height <= 0 {
		return
	}
	s.mu.Lock()
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if s.inBounds(px, py) {
				s.setLocked(px, py, v)
			}
		}
	}
	listeners := s.markChangedLocked()
	s.mu.Unlock()
	notifyListeners(listeners)
}

func (s *Surface) Line(x1, y1, x2, y2 int, v uint8) {
	if s == nil {
		return
	}
	s.mu.Lock()
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
		if s.inBounds(x1, y1) {
			s.addLocked(x1, y1, v)
		}
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
	listeners := s.markChangedLocked()
	s.mu.Unlock()
	notifyListeners(listeners)
}

func (s *Surface) Crosshatch(x, y, width, height, density int, v uint8) {
	if s == nil || width <= 0 || height <= 0 {
		return
	}
	if density <= 0 {
		density = 1
	}
	s.mu.Lock()
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if (px+py)%density == 0 && s.inBounds(px, py) {
				s.addLocked(px, py, v)
			}
			if density < 4 && (px-py)%density == 0 && s.inBounds(px, py) {
				s.addLocked(px, py, v/2)
			}
		}
	}
	listeners := s.markChangedLocked()
	s.mu.Unlock()
	notifyListeners(listeners)
}

func (s *Surface) CompositeAdd(src *Surface, xoff, yoff int) {
	if s == nil || src == nil {
		return
	}
	srcPixels, srcWidth, srcHeight := src.snapshotPixels()
	s.mu.Lock()
	for y := 0; y < srcHeight; y++ {
		for x := 0; x < srcWidth; x++ {
			v := srcPixels[y*srcWidth+x]
			if v == 0 {
				continue
			}
			if s.inBounds(xoff+x, yoff+y) {
				s.addLocked(xoff+x, yoff+y, v)
			}
		}
	}
	listeners := s.markChangedLocked()
	s.mu.Unlock()
	notifyListeners(listeners)
}

func (s *Surface) OnChange(fn func()) Subscription {
	if s == nil || fn == nil {
		return &surfaceSubscription{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listeners == nil {
		s.listeners = map[uint64]func(){}
	}
	s.nextListenerID++
	id := s.nextListenerID
	s.listeners[id] = fn
	return &surfaceSubscription{closeFn: func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.listeners, id)
	}}
}

func (s *Surface) ToRGBA(fg, bg color.Color) *image.RGBA {
	if s == nil {
		return image.NewRGBA(image.Rectangle{})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitForStableLocked()
	im := image.NewRGBA(s.Bounds())
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	fr, fgc, fb, _ := fg.RGBA()
	br, bgc, bb, _ := bg.RGBA()
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			v := uint32(s.pixels[y*s.width+x])
			if v == 0 {
				continue
			}
			inv := 255 - v
			r := (uint32(fr>>8)*v + uint32(br>>8)*inv) / 255
			g := (uint32(fgc>>8)*v + uint32(bgc>>8)*inv) / 255
			b := (uint32(fb>>8)*v + uint32(bb>>8)*inv) / 255
			im.SetRGBA(x, y, color.RGBA{
				R: clampUint32ToUint8(r),
				G: clampUint32ToUint8(g),
				B: clampUint32ToUint8(b),
				A: 0xff,
			})
		}
	}
	return im
}

func (s *Surface) setLocked(x, y int, v uint8) {
	s.pixels[y*s.width+x] = v
}

func (s *Surface) addLocked(x, y int, v uint8) {
	i := y*s.width + x
	s.pixels[i] = saturatingAdd(s.pixels[i], v)
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

func (s *Surface) endBatch() []func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.batchDepth == 0 {
		return nil
	}
	s.batchDepth--
	if s.batchDepth > 0 {
		return nil
	}
	listeners := []func(){}
	if s.changedInBatch {
		listeners = s.listenerSnapshotLocked()
		s.changedInBatch = false
	}
	s.cond.Broadcast()
	return listeners
}

func (s *Surface) markChangedLocked() []func() {
	if s.batchDepth > 0 {
		s.changedInBatch = true
		return nil
	}
	return s.listenerSnapshotLocked()
}

func (s *Surface) waitForStableLocked() {
	for s.batchDepth > 0 {
		s.cond.Wait()
	}
}

func (s *Surface) listenerSnapshotLocked() []func() {
	if len(s.listeners) == 0 {
		return nil
	}
	listeners := make([]func(), 0, len(s.listeners))
	for _, fn := range s.listeners {
		listeners = append(listeners, fn)
	}
	return listeners
}

func (s *Surface) snapshotPixels() ([]uint8, int, int) {
	if s == nil {
		return nil, 0, 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitForStableLocked()
	pixels := append([]uint8(nil), s.pixels...)
	return pixels, s.width, s.height
}

func notifyListeners(listeners []func()) {
	for _, fn := range listeners {
		if fn != nil {
			fn()
		}
	}
}

func saturatingAdd(a, b uint8) uint8 {
	sum := int(a) + int(b)
	return clampIntToUint8(sum)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
