package ui

import (
	"fmt"

	"github.com/go-go-golems/loupedeck/runtime/gfx"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

const (
	DisplayLeft  = "left"
	DisplayMain  = "main"
	DisplayRight = "right"
)

func ValidDisplayName(name string) bool {
	switch name {
	case DisplayLeft, DisplayMain, DisplayRight:
		return true
	default:
		return false
	}
}

type Display struct {
	page *Page

	Name string

	text       string
	icon       string
	visible    bool
	dirty      bool
	configured bool
	surface    *gfx.Surface
	surfaceSub gfx.Subscription

	textSub    reactive.Subscription
	iconSub    reactive.Subscription
	visibleSub reactive.Subscription

	tiles map[TileCoord]*Tile
}

func (d *Display) Text() string {
	return d.text
}

func (d *Display) Icon() string {
	return d.icon
}

func (d *Display) Visible() bool {
	return d.visible
}

func (d *Display) Dirty() bool {
	return d.dirty
}

func (d *Display) Configured() bool {
	return d.configured
}

func (d *Display) Surface() *gfx.Surface {
	return d.surface
}

func (d *Display) SetText(value string) {
	d.configured = true
	if d.text == value {
		return
	}
	d.text = value
	d.markDirty()
}

func (d *Display) BindText(fn func() string) {
	d.configured = true
	if d.textSub != nil {
		d.textSub.Stop()
	}
	d.textSub = d.page.ui.Reactive.Watch(func() {
		d.SetText(fn())
	})
}

func (d *Display) SetIcon(value string) {
	d.configured = true
	if d.icon == value {
		return
	}
	d.icon = value
	d.markDirty()
}

func (d *Display) BindIcon(fn func() string) {
	d.configured = true
	if d.iconSub != nil {
		d.iconSub.Stop()
	}
	d.iconSub = d.page.ui.Reactive.Watch(func() {
		d.SetIcon(fn())
	})
}

func (d *Display) SetVisible(value bool) {
	d.configured = true
	if d.visible == value {
		return
	}
	d.visible = value
	d.markDirty()
}

func (d *Display) BindVisible(fn func() bool) {
	d.configured = true
	if d.visibleSub != nil {
		d.visibleSub.Stop()
	}
	d.visibleSub = d.page.ui.Reactive.Watch(func() {
		d.SetVisible(fn())
	})
}

func (d *Display) SetSurface(surface *gfx.Surface) {
	d.configured = true
	if d.surfaceSub != nil {
		_ = d.surfaceSub.Close()
		d.surfaceSub = nil
	}
	d.surface = surface
	if surface != nil {
		d.surfaceSub = surface.OnChange(func() {
			d.markDirty()
		})
	}
	d.markDirty()
}

func (d *Display) AddTile(col, row int) *Tile {
	if d.Name != DisplayMain {
		panic(fmt.Sprintf("ui: display %q does not support tiles", d.Name))
	}
	if col < 0 || col >= MainDisplayColumns || row < 0 || row >= MainDisplayRows {
		panic(fmt.Sprintf("ui: tile coordinate out of range col=%d row=%d", col, row))
	}
	coord := TileCoord{Col: col, Row: row}
	if tile, ok := d.tiles[coord]; ok {
		return tile
	}
	tile := &Tile{
		display: d,
		Col:     col,
		Row:     row,
		visible: true,
	}
	d.tiles[coord] = tile
	return tile
}

func (d *Display) Tile(col, row int) *Tile {
	if d.tiles == nil {
		return nil
	}
	return d.tiles[TileCoord{Col: col, Row: row}]
}

func (d *Display) Tiles() []*Tile {
	ret := make([]*Tile, 0, len(d.tiles))
	for _, tile := range d.tiles {
		ret = append(ret, tile)
	}
	return ret
}

func (d *Display) markDirty() {
	if d.dirty {
		return
	}
	d.dirty = true
	d.page.ui.markDirtyDisplay(d)
}
