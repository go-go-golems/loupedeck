package gfx

import (
	"image/color"
	"testing"
	"time"

	"golang.org/x/image/font/basicfont"
)

func TestSurfaceClearAndAddSaturates(t *testing.T) {
	s := NewSurface(4, 4)
	s.Clear(10)
	if got := s.At(1, 1); got != 10 {
		t.Fatalf("expected clear value 10, got %d", got)
	}
	s.Add(1, 1, 250)
	if got := s.At(1, 1); got != 255 {
		t.Fatalf("expected saturating add to clamp at 255, got %d", got)
	}
}

func TestSurfaceLineTouchesEndpoints(t *testing.T) {
	s := NewSurface(10, 10)
	s.Line(1, 2, 8, 2, 90)
	if got := s.At(1, 2); got == 0 {
		t.Fatal("expected line to draw first endpoint")
	}
	if got := s.At(8, 2); got == 0 {
		t.Fatal("expected line to draw second endpoint")
	}
}

func TestSurfaceCrosshatchMarksPixels(t *testing.T) {
	s := NewSurface(10, 10)
	s.Crosshatch(0, 0, 10, 10, 2, 40)
	nonZero := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected crosshatch to mark some pixels")
	}
}

func TestSurfaceCompositeAddAddsSourcePixels(t *testing.T) {
	dst := NewSurface(8, 8)
	src := NewSurface(4, 4)
	src.FillRect(1, 1, 2, 2, 100)
	dst.CompositeAdd(src, 2, 2)
	if got := dst.At(3, 3); got != 100 {
		t.Fatalf("expected composited value 100 at translated position, got %d", got)
	}
}

func TestSurfaceTextDrawsVisiblePixels(t *testing.T) {
	s := NewSurface(80, 20)
	s.Text("EYE", TextOptions{
		X:          0,
		Y:          0,
		Width:      80,
		Height:     20,
		Brightness: 180,
		Face:       basicfont.Face7x13,
		Center:     true,
	})
	nonZero := 0
	for y := 0; y < s.Height(); y++ {
		for x := 0; x < s.Width(); x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected text drawing to produce visible pixels")
	}
}

func TestSurfaceTextHandlesSmallHeightsWithoutPanicking(t *testing.T) {
	s := NewSurface(32, 32)
	s.Text("1", TextOptions{
		X:          4,
		Y:          2,
		Width:      22,
		Height:     12,
		Brightness: 80,
		Face:       basicfont.Face7x13,
	})
}

func TestSurfaceBatchCoalescesChangeNotifications(t *testing.T) {
	s := NewSurface(16, 16)
	count := 0
	sub := s.OnChange(func() {
		count++
	})
	defer func() { _ = sub.Close() }()

	s.Batch(func() {
		s.Clear(1)
		s.FillRect(0, 0, 4, 4, 10)
		s.Line(0, 0, 15, 15, 20)
		s.Crosshatch(0, 0, 8, 8, 2, 5)
	})

	if count != 1 {
		t.Fatalf("expected one coalesced notification, got %d", count)
	}
}

func TestSurfaceToRGBAWaitsForBatchCompletion(t *testing.T) {
	s := NewSurface(8, 8)
	started := make(chan struct{})
	release := make(chan struct{})
	readDone := make(chan struct{})

	go func() {
		s.Batch(func() {
			s.FillRect(0, 0, 2, 2, 100)
			close(started)
			<-release
			s.FillRect(2, 0, 2, 2, 200)
		})
	}()

	<-started
	go func() {
		im := s.ToRGBA(color.White, color.Black)
		if got := im.RGBAAt(3, 0).R; got == 0 {
			t.Errorf("expected second half of batched surface to be visible after batch, got %d", got)
		}
		close(readDone)
	}()

	select {
	case <-readDone:
		t.Fatal("expected ToRGBA to wait until the batch completed")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case <-readDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for ToRGBA after releasing the batch")
	}
}
