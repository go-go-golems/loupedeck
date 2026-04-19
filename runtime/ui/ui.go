package ui

import (
	"fmt"
	"sort"
	"sync"

	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

const (
	MainDisplayColumns = 4
	MainDisplayRows    = 3
)

type UI struct {
	Reactive      *reactive.Runtime
	mu            sync.Mutex
	pages         map[string]*Page
	activePage    *Page
	dirtyTiles    map[*Tile]struct{}
	dirtyDisplays map[*Display]struct{}
	onDirty       func()
}

func New(rt *reactive.Runtime) *UI {
	if rt == nil {
		rt = reactive.NewRuntime()
	}
	return &UI{
		Reactive:      rt,
		pages:         map[string]*Page{},
		dirtyTiles:    map[*Tile]struct{}{},
		dirtyDisplays: map[*Display]struct{}{},
	}
}

func (u *UI) AddPage(name string) *Page {
	if page, ok := u.pages[name]; ok {
		return page
	}
	page := &Page{
		ui:       u,
		Name:     name,
		displays: map[string]*Display{},
	}
	page.AddDisplay(DisplayMain)
	u.pages[name] = page
	return page
}

func (u *UI) Page(name string) *Page {
	return u.pages[name]
}

func (u *UI) ActivePage() *Page {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.activePage
}

func (u *UI) Show(name string) error {
	u.mu.Lock()
	page, ok := u.pages[name]
	if !ok {
		u.mu.Unlock()
		return fmt.Errorf("ui: unknown page %q", name)
	}
	u.activePage = page
	u.mu.Unlock()
	u.invalidatePage(page)
	return nil
}

func (u *UI) InvalidateActivePage() bool {
	u.mu.Lock()
	page := u.activePage
	u.mu.Unlock()
	if page == nil {
		return false
	}
	u.invalidatePage(page)
	return true
}

func (u *UI) DirtyTiles() []*Tile {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.activePage == nil || len(u.dirtyTiles) == 0 {
		return nil
	}
	ret := make([]*Tile, 0, len(u.dirtyTiles))
	for tile := range u.dirtyTiles {
		if tile.display.page == u.activePage {
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

func (u *UI) DirtyDisplays() []*Display {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.activePage == nil || len(u.dirtyDisplays) == 0 {
		return nil
	}
	ret := make([]*Display, 0, len(u.dirtyDisplays))
	for display := range u.dirtyDisplays {
		if display.page == u.activePage {
			ret = append(ret, display)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return displayOrder(ret[i].Name) < displayOrder(ret[j].Name)
	})
	return ret
}

func (u *UI) ClearDirty() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for tile := range u.dirtyTiles {
		tile.dirty = false
		delete(u.dirtyTiles, tile)
	}
	for display := range u.dirtyDisplays {
		display.dirty = false
		delete(u.dirtyDisplays, display)
	}
}

func (u *UI) ClearDirtyTiles(tiles []*Tile) {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, tile := range tiles {
		if tile == nil {
			continue
		}
		tile.dirty = false
		delete(u.dirtyTiles, tile)
	}
}

func (u *UI) ClearDirtyDisplays(displays []*Display) {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, display := range displays {
		if display == nil {
			continue
		}
		display.dirty = false
		delete(u.dirtyDisplays, display)
	}
}

func (u *UI) SetDirtyHandler(fn func()) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.onDirty = fn
}

func (u *UI) markDirtyTile(tile *Tile) {
	var onDirty func()
	u.mu.Lock()
	if tile == nil || tile.dirty {
		u.mu.Unlock()
		return
	}
	tile.dirty = true
	u.dirtyTiles[tile] = struct{}{}
	onDirty = u.onDirty
	u.mu.Unlock()
	if onDirty != nil {
		onDirty()
	}
}

func (u *UI) markDirtyDisplay(display *Display) {
	var onDirty func()
	u.mu.Lock()
	if display == nil || display.dirty {
		u.mu.Unlock()
		return
	}
	display.dirty = true
	u.dirtyDisplays[display] = struct{}{}
	onDirty = u.onDirty
	u.mu.Unlock()
	if onDirty != nil {
		onDirty()
	}
}

func (u *UI) invalidatePage(page *Page) {
	for _, display := range page.displays {
		if display.Name != DisplayMain || display.Configured() {
			display.markDirty()
		}
	}
	for _, tile := range page.Tiles() {
		tile.markDirty()
	}
}

func displayOrder(name string) int {
	switch name {
	case DisplayLeft:
		return 0
	case DisplayMain:
		return 1
	case DisplayRight:
		return 2
	default:
		return 99
	}
}
