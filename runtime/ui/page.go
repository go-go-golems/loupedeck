package ui

import "fmt"

type TileCoord struct {
	Col int
	Row int
}

type Page struct {
	ui       *UI
	Name     string
	displays map[string]*Display
}

func (p *Page) AddDisplay(name string) *Display {
	if !ValidDisplayName(name) {
		panic(fmt.Sprintf("ui: unknown display %q", name))
	}
	if display, ok := p.displays[name]; ok {
		return display
	}
	display := &Display{
		page:    p,
		Name:    name,
		visible: true,
	}
	if name == DisplayMain {
		display.tiles = map[TileCoord]*Tile{}
	}
	p.displays[name] = display
	return display
}

func (p *Page) Display(name string) *Display {
	return p.displays[name]
}

func (p *Page) Displays() []*Display {
	ret := make([]*Display, 0, len(p.displays))
	for _, display := range p.displays {
		ret = append(ret, display)
	}
	return ret
}

func (p *Page) AddTile(col, row int) *Tile {
	return p.AddDisplay(DisplayMain).AddTile(col, row)
}

func (p *Page) Tile(col, row int) *Tile {
	main := p.Display(DisplayMain)
	if main == nil {
		return nil
	}
	return main.Tile(col, row)
}

func (p *Page) Tiles() []*Tile {
	main := p.Display(DisplayMain)
	if main == nil {
		return nil
	}
	return main.Tiles()
}
