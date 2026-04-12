package ui

import (
	"testing"

	"github.com/go-go-golems/loupedeck/runtime/gfx"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

func TestShowUnknownPageReturnsError(t *testing.T) {
	ui := New(nil)
	if err := ui.Show("missing"); err == nil {
		t.Fatal("expected unknown page show to fail")
	}
}

func TestPageActivationAndDirtyTileFiltering(t *testing.T) {
	rt := reactive.NewRuntime()
	ui := New(rt)

	home := ui.AddPage("home")
	home00 := home.AddTile(0, 0)
	home00.SetText("home")
	home11 := home.AddTile(1, 1)
	home11.SetText("other")

	alt := ui.AddPage("alt")
	alt00 := alt.AddTile(0, 0)
	alt00.SetText("alt")

	if err := ui.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}

	dirty := ui.DirtyTiles()
	if len(dirty) != 2 {
		t.Fatalf("expected 2 dirty home tiles, got %d", len(dirty))
	}
	if dirty[0] != home00 || dirty[1] != home11 {
		t.Fatalf("unexpected dirty tile order/filtering: %#v", dirty)
	}

	ui.ClearDirty()
	alt00.SetText("alt-updated")
	dirty = ui.DirtyTiles()
	if len(dirty) != 0 {
		t.Fatalf("expected hidden-page dirty tiles to be filtered out, got %d", len(dirty))
	}

	if err := ui.Show("alt"); err != nil {
		t.Fatalf("show alt: %v", err)
	}
	dirty = ui.DirtyTiles()
	if len(dirty) != 1 || dirty[0] != alt00 {
		t.Fatalf("expected alt tile to become dirty/visible after show, got %#v", dirty)
	}
}

func TestTileStaticProperties(t *testing.T) {
	ui := New(nil)
	page := ui.AddPage("home")
	tile := page.AddTile(2, 1)

	tile.SetText("REC")
	tile.SetIcon("record")
	tile.SetVisible(false)

	if tile.Text() != "REC" {
		t.Fatalf("expected text REC, got %q", tile.Text())
	}
	if tile.Icon() != "record" {
		t.Fatalf("expected icon record, got %q", tile.Icon())
	}
	if tile.Visible() {
		t.Fatal("expected tile to be invisible")
	}
	if !tile.Dirty() {
		t.Fatal("expected tile to be dirty after static property changes")
	}
}

func TestReactiveTileBindingsUpdateProperties(t *testing.T) {
	rt := reactive.NewRuntime()
	ui := New(rt)
	page := ui.AddPage("home")
	tile := page.AddTile(0, 0)

	label := reactive.NewSignal(rt, "IDLE")
	icon := reactive.NewSignal(rt, "pause")
	visible := reactive.NewSignal(rt, true)

	tile.BindText(func() string { return label.Get() })
	tile.BindIcon(func() string { return icon.Get() })
	tile.BindVisible(func() bool { return visible.Get() })

	if tile.Text() != "IDLE" || tile.Icon() != "pause" || !tile.Visible() {
		t.Fatalf("unexpected initial bound tile state text=%q icon=%q visible=%v", tile.Text(), tile.Icon(), tile.Visible())
	}
	if !tile.Dirty() {
		t.Fatal("expected initial binding run to mark tile dirty")
	}

	ui.ClearDirty()
	rt.Batch(func() {
		label.Set("REC")
		icon.Set("record")
		visible.Set(false)
	})

	if tile.Text() != "REC" {
		t.Fatalf("expected updated text REC, got %q", tile.Text())
	}
	if tile.Icon() != "record" {
		t.Fatalf("expected updated icon record, got %q", tile.Icon())
	}
	if tile.Visible() {
		t.Fatal("expected updated tile to be invisible")
	}
	if !tile.Dirty() {
		t.Fatal("expected reactive updates to mark tile dirty")
	}
}

func TestDisplayActivationAndDirtyFiltering(t *testing.T) {
	rt := reactive.NewRuntime()
	ui := New(rt)

	home := ui.AddPage("home")
	left := home.AddDisplay(DisplayLeft)
	left.SetText("LEFT")
	right := home.AddDisplay(DisplayRight)
	right.SetText("RIGHT")

	alt := ui.AddPage("alt")
	altLeft := alt.AddDisplay(DisplayLeft)
	altLeft.SetText("ALT")

	if err := ui.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	dirty := ui.DirtyDisplays()
	if len(dirty) != 2 || dirty[0] != left || dirty[1] != right {
		t.Fatalf("unexpected home dirty displays: %#v", dirty)
	}

	ui.ClearDirty()
	altLeft.SetText("ALT2")
	dirty = ui.DirtyDisplays()
	if len(dirty) != 0 {
		t.Fatalf("expected hidden-page dirty displays to be filtered out, got %d", len(dirty))
	}

	if err := ui.Show("alt"); err != nil {
		t.Fatalf("show alt: %v", err)
	}
	dirty = ui.DirtyDisplays()
	if len(dirty) != 1 || dirty[0] != altLeft {
		t.Fatalf("expected alt display to become dirty/visible after show, got %#v", dirty)
	}
}

func TestDisplayMainTileCompatibility(t *testing.T) {
	ui := New(nil)
	page := ui.AddPage("home")
	main := page.Display(DisplayMain)
	if main == nil {
		t.Fatal("expected implicit main display")
	}
	tileViaPage := page.AddTile(2, 1)
	tileViaDisplay := main.Tile(2, 1)
	if tileViaPage == nil || tileViaDisplay == nil || tileViaPage != tileViaDisplay {
		t.Fatal("expected page tile API to delegate to main display")
	}
}

func TestDisplaySurfaceMutationMarksDisplayDirty(t *testing.T) {
	ui := New(nil)
	page := ui.AddPage("home")
	left := page.AddDisplay(DisplayLeft)
	surface := gfx.NewSurface(8, 8)
	left.SetSurface(surface)
	if err := ui.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	ui.ClearDirty()
	surface.FillRect(0, 0, 4, 4, 100)
	dirty := ui.DirtyDisplays()
	if len(dirty) != 1 || dirty[0] != left {
		t.Fatalf("expected surface mutation to mark left display dirty, got %#v", dirty)
	}
}

func TestDisplayLayerMutationMarksDisplayDirty(t *testing.T) {
	ui := New(nil)
	page := ui.AddPage("home")
	main := page.AddDisplay(DisplayMain)
	overlay := gfx.NewSurface(16, 16)
	main.SetLayer("overlay", overlay)
	if err := ui.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	ui.ClearDirty()
	overlay.FillRect(0, 0, 4, 4, 100)
	dirty := ui.DirtyDisplays()
	if len(dirty) != 1 || dirty[0] != main {
		t.Fatalf("expected layer mutation to mark main display dirty, got %#v", dirty)
	}
}

func TestDisplayLayerOrderIsStable(t *testing.T) {
	ui := New(nil)
	page := ui.AddPage("home")
	main := page.AddDisplay(DisplayMain)
	main.SetLayer("basefx", gfx.NewSurface(4, 4))
	main.SetLayer("scan", gfx.NewSurface(4, 4))
	main.SetLayer("scan", gfx.NewSurface(4, 4))
	main.SetLayer("ripple", gfx.NewSurface(4, 4))
	layers := main.Layers()
	if len(layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(layers))
	}
	if layers[0].Name != "basefx" || layers[1].Name != "scan" || layers[2].Name != "ripple" {
		t.Fatalf("unexpected layer order: %#v %#v %#v", layers[0].Name, layers[1].Name, layers[2].Name)
	}
	main.SetLayer("scan", nil)
	layers = main.Layers()
	if len(layers) != 2 || layers[0].Name != "basefx" || layers[1].Name != "ripple" {
		t.Fatalf("unexpected layer order after removal: %#v", layers)
	}
}

func TestTileSurfaceMutationMarksTileDirty(t *testing.T) {
	ui := New(nil)
	page := ui.AddPage("home")
	tile := page.AddTile(0, 0)
	surface := gfx.NewSurface(8, 8)
	tile.SetSurface(surface)
	if err := ui.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	ui.ClearDirty()
	surface.FillRect(0, 0, 4, 4, 100)
	dirty := ui.DirtyTiles()
	if len(dirty) != 1 || dirty[0] != tile {
		t.Fatalf("expected tile surface mutation to mark tile dirty, got %#v", dirty)
	}
}
