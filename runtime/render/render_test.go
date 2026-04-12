package render

import (
	"image"
	"testing"

	"github.com/go-go-golems/loupedeck/runtime/gfx"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

type drawCall struct {
	x  int
	y  int
	im image.Image
}

type fakeTarget struct {
	calls []drawCall
}

func (f *fakeTarget) Draw(im image.Image, xoff, yoff int) {
	f.calls = append(f.calls, drawCall{x: xoff, y: yoff, im: im})
}

func TestTileRectMapsToMainDisplayGrid(t *testing.T) {
	rect := TileRect(3, 2)
	if rect.Min.X != 270 || rect.Min.Y != 180 || rect.Dx() != 90 || rect.Dy() != 90 {
		t.Fatalf("unexpected tile rect: %+v", rect)
	}
}

func TestFlushDrawsDirtyActiveTilesAtTileCoordinates(t *testing.T) {
	rt := reactive.NewRuntime()
	uiRuntime := ui.New(rt)
	page := uiRuntime.AddPage("home")
	t00 := page.AddTile(0, 0)
	t00.SetText("HOME")
	t32 := page.AddTile(3, 2)
	t32.SetText("END")
	if err := uiRuntime.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}

	target := &fakeTarget{}
	r := New(uiRuntime, target)
	flushed := r.Flush()

	if flushed != 2 {
		t.Fatalf("expected 2 flushed tiles, got %d", flushed)
	}
	if len(target.calls) != 2 {
		t.Fatalf("expected 2 draw calls, got %d", len(target.calls))
	}
	if target.calls[0].x != 0 || target.calls[0].y != 0 {
		t.Fatalf("unexpected first tile coords: (%d,%d)", target.calls[0].x, target.calls[0].y)
	}
	if target.calls[1].x != 270 || target.calls[1].y != 180 {
		t.Fatalf("unexpected second tile coords: (%d,%d)", target.calls[1].x, target.calls[1].y)
	}
	if t00.Dirty() || t32.Dirty() {
		t.Fatal("expected flushed tiles to be marked clean")
	}
}

func TestFlushPreservesHiddenPageDirtyTiles(t *testing.T) {
	rt := reactive.NewRuntime()
	uiRuntime := ui.New(rt)

	home := uiRuntime.AddPage("home")
	homeTile := home.AddTile(0, 0)
	homeTile.SetText("HOME")

	alt := uiRuntime.AddPage("alt")
	altTile := alt.AddTile(1, 0)
	altTile.SetText("ALT")

	if err := uiRuntime.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	uiRuntime.ClearDirty()
	altTile.SetText("ALT2")
	homeTile.SetText("HOME2")

	target := &fakeTarget{}
	r := New(uiRuntime, target)
	flushed := r.Flush()
	if flushed != 1 {
		t.Fatalf("expected only active page tile flush, got %d", flushed)
	}
	if !altTile.Dirty() {
		t.Fatal("expected hidden-page tile to remain dirty after active-page flush")
	}

	if err := uiRuntime.Show("alt"); err != nil {
		t.Fatalf("show alt: %v", err)
	}
	flushed = r.Flush()
	if flushed != 1 {
		t.Fatalf("expected alt page tile flush after page switch, got %d", flushed)
	}
	if altTile.Dirty() {
		t.Fatal("expected alt tile to be clean after its page flushes")
	}
}

func TestFlushDrawsDirtySideDisplays(t *testing.T) {
	rt := reactive.NewRuntime()
	uiRuntime := ui.New(rt)
	page := uiRuntime.AddPage("home")
	left := page.AddDisplay(ui.DisplayLeft)
	left.SetText("LEFT")
	right := page.AddDisplay(ui.DisplayRight)
	right.SetText("RIGHT")
	if err := uiRuntime.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}

	leftTarget := &fakeTarget{}
	mainTarget := &fakeTarget{}
	rightTarget := &fakeTarget{}
	r := NewWithDisplays(uiRuntime, map[string]DrawTarget{
		ui.DisplayLeft:  leftTarget,
		ui.DisplayMain:  mainTarget,
		ui.DisplayRight: rightTarget,
	})
	flushed := r.Flush()
	if flushed != 2 {
		t.Fatalf("expected 2 flushed side displays, got %d", flushed)
	}
	if len(leftTarget.calls) != 1 || len(rightTarget.calls) != 1 {
		t.Fatalf("expected one draw call per side display, got left=%d right=%d", len(leftTarget.calls), len(rightTarget.calls))
	}
	if len(mainTarget.calls) != 0 {
		t.Fatalf("expected no main draw calls, got %d", len(mainTarget.calls))
	}
	if leftTarget.calls[0].im.Bounds().Dx() != SideDisplayWidth || rightTarget.calls[0].im.Bounds().Dx() != SideDisplayWidth {
		t.Fatal("expected side display renders to use side display dimensions")
	}
}

func TestFlushDrawsDisplaySurfaceContent(t *testing.T) {
	rt := reactive.NewRuntime()
	uiRuntime := ui.New(rt)
	page := uiRuntime.AddPage("home")
	left := page.AddDisplay(ui.DisplayLeft)
	surface := gfx.NewSurface(SideDisplayWidth, SideDisplayHeight)
	surface.FillRect(0, 0, 4, 4, 120)
	left.SetSurface(surface)
	if err := uiRuntime.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	leftTarget := &fakeTarget{}
	r := NewWithDisplays(uiRuntime, map[string]DrawTarget{ui.DisplayLeft: leftTarget})
	flushed := r.Flush()
	if flushed != 1 {
		t.Fatalf("expected one flushed display surface, got %d", flushed)
	}
	if len(leftTarget.calls) != 1 {
		t.Fatalf("expected one left display draw call, got %d", len(leftTarget.calls))
	}
	rv, gv, bv, av := leftTarget.calls[0].im.At(0, 0).RGBA()
	if rv == 0 && gv == 0 && bv == 0 && av == 0 {
		t.Fatal("expected rendered surface pixel to be non-zero")
	}
}
