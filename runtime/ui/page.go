package ui

import "fmt"

type TileCoord struct {
	Col int
	Row int
}

type Page struct {
	ui    *UI
	Name  string
	tiles map[TileCoord]*Tile
}

func (p *Page) AddTile(col, row int) *Tile {
	if col < 0 || col >= MainDisplayColumns || row < 0 || row >= MainDisplayRows {
		panic(fmt.Sprintf("ui: tile coordinate out of range col=%d row=%d", col, row))
	}
	coord := TileCoord{Col: col, Row: row}
	if tile, ok := p.tiles[coord]; ok {
		return tile
	}
	tile := &Tile{
		page:    p,
		Col:     col,
		Row:     row,
		visible: true,
	}
	p.tiles[coord] = tile
	return tile
}

func (p *Page) Tile(col, row int) *Tile {
	return p.tiles[TileCoord{Col: col, Row: row}]
}

func (p *Page) Tiles() []*Tile {
	ret := make([]*Tile, 0, len(p.tiles))
	for _, tile := range p.tiles {
		ret = append(ret, tile)
	}
	return ret
}
