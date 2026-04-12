package ui

import (
	"fmt"
	"sort"

	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

const (
	MainDisplayColumns = 4
	MainDisplayRows    = 3
)

type UI struct {
	Reactive   *reactive.Runtime
	pages      map[string]*Page
	activePage *Page
	dirtyTiles map[*Tile]struct{}
}

func New(rt *reactive.Runtime) *UI {
	if rt == nil {
		rt = reactive.NewRuntime()
	}
	return &UI{
		Reactive:   rt,
		pages:      map[string]*Page{},
		dirtyTiles: map[*Tile]struct{}{},
	}
}

func (u *UI) AddPage(name string) *Page {
	if page, ok := u.pages[name]; ok {
		return page
	}
	page := &Page{
		ui:    u,
		Name:  name,
		tiles: map[TileCoord]*Tile{},
	}
	u.pages[name] = page
	return page
}

func (u *UI) Page(name string) *Page {
	return u.pages[name]
}

func (u *UI) ActivePage() *Page {
	return u.activePage
}

func (u *UI) Show(name string) error {
	page, ok := u.pages[name]
	if !ok {
		return fmt.Errorf("ui: unknown page %q", name)
	}
	u.activePage = page
	u.invalidatePage(page)
	return nil
}

func (u *UI) InvalidateActivePage() bool {
	if u.activePage == nil {
		return false
	}
	u.invalidatePage(u.activePage)
	return true
}

func (u *UI) DirtyTiles() []*Tile {
	if u.activePage == nil || len(u.dirtyTiles) == 0 {
		return nil
	}
	ret := make([]*Tile, 0, len(u.dirtyTiles))
	for tile := range u.dirtyTiles {
		if tile.page == u.activePage {
			ret = append(ret, tile)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Row != ret[j].Row {
			return ret[i].Row < ret[j].Row
		}
		return ret[i].Col < ret[j].Col
	})
	return ret
}

func (u *UI) ClearDirty() {
	tiles := make([]*Tile, 0, len(u.dirtyTiles))
	for tile := range u.dirtyTiles {
		tiles = append(tiles, tile)
	}
	u.ClearDirtyTiles(tiles)
}

func (u *UI) ClearDirtyTiles(tiles []*Tile) {
	for _, tile := range tiles {
		if tile == nil {
			continue
		}
		tile.dirty = false
		delete(u.dirtyTiles, tile)
	}
}

func (u *UI) markDirty(tile *Tile) {
	u.dirtyTiles[tile] = struct{}{}
}

func (u *UI) invalidatePage(page *Page) {
	for _, tile := range page.tiles {
		tile.markDirty()
	}
}
