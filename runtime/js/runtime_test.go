package js

import (
	"sync"
	"testing"
	"time"

	"github.com/dop251/goja"
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
	rt := NewRuntime(nil)
	defer rt.Close(nil)
	env := rt.Env

	_, err := rt.RunString(nil, `
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
	rt := NewRuntime(env)
	defer rt.Close(nil)
	env = rt.Env

	_, err := rt.RunString(nil, `
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

func TestAnimModuleCanDriveSignalTweenFromButtonEvent(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	env.Anim.FrameInterval = 5 * time.Millisecond
	rt := NewRuntime(env)
	defer rt.Close(nil)
	env = rt.Env

	_, err := rt.RunString(nil, `
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const anim = require("loupedeck/anim");
		const easing = require("loupedeck/easing");
		const level = state.signal(0);
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.text(() => String(Math.round(level.get())));
		  });
		});
		ui.onButton("Circle", () => {
		  anim.to(level, 9, 30, easing.linear);
		});
		ui.show("home");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	source.emitButton(deck.Circle, deck.ButtonDown)
	time.Sleep(80 * time.Millisecond)
	if got := env.UI.Page("home").Tile(0, 0).Text(); got != "9" {
		t.Fatalf("expected tweened value 9, got %q", got)
	}
}

func TestAnimModuleLoopCanDriveReactiveUpdates(t *testing.T) {
	env := envpkg.Ensure(nil)
	env.Anim.FrameInterval = 5 * time.Millisecond
	rt := NewRuntime(env)
	defer rt.Close(nil)
	env = rt.Env

	_, err := rt.RunString(nil, `
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const anim = require("loupedeck/anim");
		const phase = state.signal(0);
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.text(() => String(Math.round(phase.get() * 10)));
		  });
		});
		const handle = anim.loop(20, t => {
		  phase.set(t);
		});
		ui.show("home");
		globalThis.__loopHandle = handle;
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	time.Sleep(40 * time.Millisecond)
	loopHandle := rt.VM.Get("__loopHandle").ToObject(rt.VM)
	stop, ok := goja.AssertFunction(loopHandle.Get("stop"))
	if !ok {
		t.Fatal("expected loop handle to expose stop()")
	}
	if _, err := stop(loopHandle); err != nil {
		t.Fatalf("stop loop: %v", err)
	}
	if got := env.UI.Page("home").Tile(0, 0).Text(); got == "0" || got == "" {
		t.Fatalf("expected loop to update visible text, got %q", got)
	}
}
