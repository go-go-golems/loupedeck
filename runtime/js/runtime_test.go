package js

import (
	"sync"
	"testing"

	deck "github.com/go-go-golems/loupedeck"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

type fakeSource struct {
	mu      sync.Mutex
	buttons map[deck.Button][]deck.ButtonFunc
}

func newFakeSource() *fakeSource {
	return &fakeSource{buttons: map[deck.Button][]deck.ButtonFunc{}}
}

type fakeSub struct{ closeFn func() }

func (s *fakeSub) Close() error {
	if s.closeFn != nil {
		s.closeFn()
		s.closeFn = nil
	}
	return nil
}

func (f *fakeSource) OnButton(button deck.Button, fn deck.ButtonFunc) deck.Subscription {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buttons[button] = append(f.buttons[button], fn)
	idx := len(f.buttons[button]) - 1
	return &fakeSub{closeFn: func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if idx < len(f.buttons[button]) {
			f.buttons[button] = append(f.buttons[button][:idx], f.buttons[button][idx+1:]...)
		}
	}}
}

func (f *fakeSource) OnTouch(deck.TouchButton, deck.TouchFunc) deck.Subscription {
	return &fakeSub{}
}

func (f *fakeSource) OnKnob(deck.Knob, deck.KnobFunc) deck.Subscription {
	return &fakeSub{}
}

func (f *fakeSource) emitButton(button deck.Button, status deck.ButtonStatus) {
	f.mu.Lock()
	callbacks := append([]deck.ButtonFunc(nil), f.buttons[button]...)
	f.mu.Unlock()
	for _, cb := range callbacks {
		cb(button, status)
	}
}

func TestRequireStateAndUIBuildReactivePage(t *testing.T) {
	env := envpkg.Ensure(nil)
	vm, env := NewRuntime(env)

	_, err := vm.RunString(`
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const mode = state.signal("IDLE");
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.icon("record");
		    tile.text(() => mode.get());
		  });
		});
		ui.show("home");
		mode.set("REC");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	tile := env.UI.Page("home").Tile(0, 0)
	if tile == nil {
		t.Fatal("expected tile to exist")
	}
	if tile.Icon() != "record" {
		t.Fatalf("expected icon record, got %q", tile.Icon())
	}
	if tile.Text() != "REC" {
		t.Fatalf("expected text REC, got %q", tile.Text())
	}
	if env.UI.ActivePage() == nil || env.UI.ActivePage().Name != "home" {
		t.Fatal("expected home to be active page")
	}
}

func TestButtonCallbackCanMutateSignalFromJS(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	vm, env := NewRuntime(env)

	_, err := vm.RunString(`
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const mode = state.signal("IDLE");
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.text(() => mode.get());
		  });
		});
		ui.onButton("Circle", () => {
		  mode.set("ARMED");
		});
		ui.show("home");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	source.emitButton(deck.Circle, deck.ButtonDown)
	tile := env.UI.Page("home").Tile(0, 0)
	if tile.Text() != "ARMED" {
		t.Fatalf("expected text ARMED after button event, got %q", tile.Text())
	}
}
