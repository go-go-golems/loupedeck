package ui

import (
	"github.com/go-go-golems/loupedeck/runtime/gfx"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

const (
	TileSurfaceWidth  = 90
	TileSurfaceHeight = 90
)

type Tile struct {
	display *Display

	Col int
	Row int

	text       string
	icon       string
	visible    bool
	wrap       bool
	dirty      bool
	surface    *gfx.Surface
	surfaceSub gfx.Subscription

	textSub    reactive.Subscription
	iconSub    reactive.Subscription
	visibleSub reactive.Subscription
}

func (t *Tile) Text() string {
	return t.text
}

func (t *Tile) Icon() string {
	return t.icon
}

func (t *Tile) Visible() bool {
	return t.visible
}

func (t *Tile) Wrap() bool {
	return t.wrap
}

func (t *Tile) SetWrap(value bool) {
	if t.wrap == value {
		return
	}
	t.wrap = value
	t.markDirty()
}

func (t *Tile) Surface() *gfx.Surface {
	return t.surface
}

func (t *Tile) Dirty() bool {
	return t.dirty
}

func (t *Tile) SetText(value string) {
	if t.text == value {
		return
	}
	t.text = value
	t.markDirty()
}

func (t *Tile) BindText(fn func() string) {
	if t.textSub != nil {
		t.textSub.Stop()
	}
	t.textSub = t.display.page.ui.Reactive.Watch(func() {
		t.SetText(fn())
	})
}

func (t *Tile) SetIcon(value string) {
	if t.icon == value {
		return
	}
	t.icon = value
	t.markDirty()
}

func (t *Tile) BindIcon(fn func() string) {
	if t.iconSub != nil {
		t.iconSub.Stop()
	}
	t.iconSub = t.display.page.ui.Reactive.Watch(func() {
		t.SetIcon(fn())
	})
}

func (t *Tile) SetVisible(value bool) {
	if t.visible == value {
		return
	}
	t.visible = value
	t.markDirty()
}

func (t *Tile) BindVisible(fn func() bool) {
	if t.visibleSub != nil {
		t.visibleSub.Stop()
	}
	t.visibleSub = t.display.page.ui.Reactive.Watch(func() {
		t.SetVisible(fn())
	})
}

func (t *Tile) SetSurface(surface *gfx.Surface) {
	if t.surfaceSub != nil {
		_ = t.surfaceSub.Close()
		t.surfaceSub = nil
	}
	t.surface = surface
	if surface != nil {
		t.surfaceSub = surface.OnChange(func() {
			t.markDirty()
		})
	}
	t.markDirty()
}

// Draw creates a per-tile surface (90×90) if one doesn't already exist,
// passes it to fn for drawing, and marks the tile as dirty.
// This is a convenience method for the common pattern of creating
// a surface, drawing on it, and assigning it to the tile.
func (t *Tile) Draw(fn func(*gfx.Surface)) {
	if t.surface == nil {
		t.SetSurface(gfx.NewSurface(TileSurfaceWidth, TileSurfaceHeight))
	}
	fn(t.surface)
}

// Invalidate marks the tile as dirty, causing it to be re-rendered
// on the next flush. This is useful when the tile's surface has been
// modified externally (e.g., through a reference to the surface).
func (t *Tile) Invalidate() {
	t.markDirty()
}

func (t *Tile) markDirty() {
	t.display.page.ui.markDirtyTile(t)
}
