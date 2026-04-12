package ui

import (
	"fmt"
	"image/color"

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
	layers     map[string]*DisplayLayer
	layerOrder []string

	textSub    reactive.Subscription
	iconSub    reactive.Subscription
	visibleSub reactive.Subscription

	tiles map[TileCoord]*Tile
}

type LayerOptions struct {
	Foreground color.Color
}

type DisplayLayer struct {
	Name       string
	surface    *gfx.Surface
	foreground color.Color
	surfaceSub gfx.Subscription
}

func (l *DisplayLayer) Surface() *gfx.Surface {
	if l == nil {
		return nil
	}
	return l.surface
}

func (l *DisplayLayer) Foreground() color.Color {
	if l == nil {
		return nil
	}
	return l.foreground
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

func (d *Display) SetLayer(name string, surface *gfx.Surface) {
	d.SetLayerWithOptions(name, surface, LayerOptions{})
}

func (d *Display) SetLayerWithOptions(name string, surface *gfx.Surface, opts LayerOptions) {
	if name == "" {
		panic("ui: display layer name must not be empty")
	}
	d.configured = true
	if d.layers == nil {
		d.layers = map[string]*DisplayLayer{}
	}
	if layer, ok := d.layers[name]; ok {
		if layer.surfaceSub != nil {
			_ = layer.surfaceSub.Close()
			layer.surfaceSub = nil
		}
		if surface == nil {
			delete(d.layers, name)
			for i, existing := range d.layerOrder {
				if existing == name {
					d.layerOrder = append(d.layerOrder[:i], d.layerOrder[i+1:]...)
					break
				}
			}
			d.markDirty()
			return
		}
		layer.surface = surface
		layer.foreground = opts.Foreground
		layer.surfaceSub = surface.OnChange(func() {
			d.markDirty()
		})
		d.markDirty()
		return
	}
	if surface == nil {
		return
	}
	layer := &DisplayLayer{Name: name, surface: surface, foreground: opts.Foreground}
	layer.surfaceSub = surface.OnChange(func() {
		d.markDirty()
	})
	d.layers[name] = layer
	d.layerOrder = append(d.layerOrder, name)
	d.markDirty()
}

func (d *Display) Layer(name string) *DisplayLayer {
	if d == nil || d.layers == nil {
		return nil
	}
	return d.layers[name]
}

func (d *Display) Layers() []*DisplayLayer {
	if d == nil || len(d.layerOrder) == 0 {
		return nil
	}
	ret := make([]*DisplayLayer, 0, len(d.layerOrder))
	for _, name := range d.layerOrder {
		if layer := d.layers[name]; layer != nil {
			ret = append(ret, layer)
		}
	}
	return ret
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
	d.page.ui.markDirtyDisplay(d)
}
