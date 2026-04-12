package js

import (
	"sync"
	"testing"
	"time"

	"github.com/dop251/goja"
	deck "github.com/go-go-golems/loupedeck"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
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

	bindings, ok := runtimebridge.Lookup(rt.VM)
	if !ok {
		t.Fatal("expected runtime bridge bindings to be registered")
	}
	if bindings.Owner == nil || bindings.Context == nil || bindings.Loop == nil {
		t.Fatal("expected owner/context/loop bindings to be populated")
	}
	if bindings.Values["environment"] != env {
		t.Fatal("expected environment to be available through runtime bindings")
	}

	_, err := rt.RunString(nil, `
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const mode = state.signal("IDLE");
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.icon("record");
		    tile.text(() => mode.get());
		  });
		  page.display("left", display => {
		    display.text("LEFT");
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
	left := env.UI.Page("home").Display("left")
	if left == nil || left.Text() != "LEFT" {
		t.Fatalf("expected left display text LEFT, got %#v", left)
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
	waitForText(t, tile, "ARMED")
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

func TestGfxModuleCanBuildAndCompositeSurface(t *testing.T) {
	rt := NewRuntime(nil)
	defer rt.Close(nil)

	_, err := rt.RunString(nil, `
		const gfx = require("loupedeck/gfx");
		const base = gfx.surface(16, 16);
		base.clear(0);
		base.set(1, 1, 30);
		base.add(1, 1, 20);
		base.line(0, 0, 15, 0, 50);
		base.crosshatch(0, 0, 8, 8, 2, 20);
		base.text("EYE", { x: 0, y: 0, width: 16, height: 16, brightness: 120, center: true });
		const overlay = gfx.surface(4, 4);
		overlay.fillRect(0, 0, 4, 4, 100);
		base.compositeAdd(overlay, 2, 2);
		globalThis.__gfxBase = base;
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	base := rt.VM.Get("__gfxBase").ToObject(rt.VM)
	at, ok := goja.AssertFunction(base.Get("at"))
	if !ok {
		t.Fatal("expected gfx surface to expose at()")
	}
	v, err := at(base, rt.VM.ToValue(2), rt.VM.ToValue(2))
	if err != nil {
		t.Fatalf("call at(): %v", err)
	}
	if got := int(v.ToInteger()); got == 0 {
		t.Fatal("expected composited/textured surface to have non-zero brightness at sampled point")
	}
	widthFn, ok := goja.AssertFunction(base.Get("width"))
	if !ok {
		t.Fatal("expected gfx surface to expose width()")
	}
	width, err := widthFn(base)
	if err != nil {
		t.Fatalf("call width(): %v", err)
	}
	if width.ToInteger() != 16 {
		t.Fatalf("expected width 16, got %d", width.ToInteger())
	}
}

func TestDisplayCanOwnGfxSurface(t *testing.T) {
	rt := NewRuntime(nil)
	defer rt.Close(nil)
	env := rt.Env

	_, err := rt.RunString(nil, `
		const ui = require("loupedeck/ui");
		const gfx = require("loupedeck/gfx");
		const s = gfx.surface(60, 270);
		s.fillRect(0, 0, 5, 5, 120);
		ui.page("home", page => {
		  page.display("left", display => {
		    display.surface(s);
		  });
		});
		ui.show("home");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	left := env.UI.Page("home").Display("left")
	if left == nil || left.Surface() == nil {
		t.Fatal("expected left display to own a gfx surface")
	}
	if got := left.Surface().At(0, 0); got == 0 {
		t.Fatal("expected surface content to remain attached to display")
	}
}

func TestDisplayCanOwnNamedGfxLayer(t *testing.T) {
	rt := NewRuntime(nil)
	defer rt.Close(nil)
	env := rt.Env

	_, err := rt.RunString(nil, `
		const ui = require("loupedeck/ui");
		const gfx = require("loupedeck/gfx");
		const base = gfx.surface(360, 270);
		const overlay = gfx.surface(360, 270);
		base.fillRect(0, 0, 5, 5, 80);
		overlay.fillRect(10, 10, 5, 5, 160);
		ui.page("home", page => {
		  page.display("main", display => {
		    display.surface(base);
		    display.layer("overlay", overlay);
		  });
		});
		ui.show("home");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	main := env.UI.Page("home").Display("main")
	if main == nil || main.Layer("overlay") == nil || main.Layer("overlay").Surface() == nil {
		t.Fatal("expected main display to own a named overlay layer")
	}
	if got := main.Layer("overlay").Surface().At(10, 10); got == 0 {
		t.Fatal("expected overlay layer content to remain attached to display")
	}
}

func TestCloseRemovesRuntimeBridgeBindings(t *testing.T) {
	rt := NewRuntime(nil)
	if _, ok := runtimebridge.Lookup(rt.VM); !ok {
		t.Fatal("expected bindings before close")
	}
	if err := rt.Close(nil); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	if _, ok := runtimebridge.Lookup(rt.VM); ok {
		t.Fatal("expected bindings to be removed on close")
	}
}

func TestConcurrentButtonCallbacksSerializeOntoOwnerThread(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	rt := NewRuntime(env)
	defer rt.Close(nil)
	env = rt.Env

	_, err := rt.RunString(nil, `
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const count = state.signal(0);
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.text(() => String(count.get()));
		  });
		});
		ui.onButton("Circle", () => {
		  count.update(v => v + 1);
		});
		ui.show("home");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	for i := 0; i < 25; i++ {
		go source.emitButton(deck.Circle, deck.ButtonDown)
	}
	waitForText(t, env.UI.Page("home").Tile(0, 0), "25")
}

func TestButtonEventAfterCloseDoesNotApplyMutation(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	rt := NewRuntime(env)
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
	if err := rt.Close(nil); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	source.emitButton(deck.Circle, deck.ButtonDown)
	time.Sleep(30 * time.Millisecond)
	if got := env.UI.Page("home").Tile(0, 0).Text(); got != "IDLE" {
		t.Fatalf("expected no mutation after close, got %q", got)
	}
}

func waitForText(t *testing.T, tile interface{ Text() string }, want string) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := tile.Text(); got == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for tile text %q, got %q", want, tile.Text())
}
