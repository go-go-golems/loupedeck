package ui

import "github.com/go-go-golems/loupedeck/runtime/reactive"

type Tile struct {
	page *Page

	Col int
	Row int

	text    string
	icon    string
	visible bool
	dirty   bool

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
	t.textSub = t.page.ui.Reactive.Watch(func() {
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
	t.iconSub = t.page.ui.Reactive.Watch(func() {
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
	t.visibleSub = t.page.ui.Reactive.Watch(func() {
		t.SetVisible(fn())
	})
}

func (t *Tile) markDirty() {
	if t.dirty {
		return
	}
	t.dirty = true
	t.page.ui.markDirty(t)
}
