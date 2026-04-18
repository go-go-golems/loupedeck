package js

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/runtime/gfx"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"golang.org/x/image/font/gofont/goregular"
)

type fakeSource struct {
	mu      sync.Mutex
	buttons map[device.Button][]device.ButtonFunc
}

func newFakeSource() *fakeSource {
	return &fakeSource{buttons: map[device.Button][]device.ButtonFunc{}}
}

type fakeSub struct{ closeFn func() }

func (s *fakeSub) Close() error {
	if s.closeFn != nil {
		s.closeFn()
		s.closeFn = nil
	}
	return nil
}

func (f *fakeSource) OnButton(button device.Button, fn device.ButtonFunc) device.Subscription {
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

func (f *fakeSource) OnTouch(device.TouchButton, device.TouchFunc) device.Subscription {
	return &fakeSub{}
}

func (f *fakeSource) OnKnob(device.Knob, device.KnobFunc) device.Subscription {
	return &fakeSub{}
}

func (f *fakeSource) emitButton(button device.Button, status device.ButtonStatus) {
	f.mu.Lock()
	callbacks := append([]device.ButtonFunc(nil), f.buttons[button]...)
	f.mu.Unlock()
	for _, cb := range callbacks {
		cb(button, status)
	}
}

func TestRequireStateAndUIBuildReactivePage(t *testing.T) {
	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()
	env := rt.Env

	bindings, ok := runtimebridge.Lookup(rt.VM)
	if !ok {
		t.Fatal("expected runtime bridge bindings to be registered")
	}
	if bindings.Owner == nil || bindings.Context == nil || bindings.Loop == nil {
		t.Fatal("expected owner/context/loop bindings to be populated")
	}
	lookupEnv, ok := envpkg.Lookup(rt.VM)
	if !ok || lookupEnv != env {
		t.Fatal("expected environment to be available through env lookup")
	}

	_, err := rt.RunString(context.Background(), `
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
	defer func() { _ = rt.Close(context.Background()) }()
	env = rt.Env

	_, err := rt.RunString(context.Background(), `
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

	source.emitButton(device.Circle, device.ButtonDown)
	tile := env.UI.Page("home").Tile(0, 0)
	waitForText(t, tile, "ARMED")
}

func TestPresenterAutoFlushesForPlainReactiveScripts(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	rt := NewRuntime(env)
	defer func() { _ = rt.Close(context.Background()) }()
	env = rt.Env

	flushes := make(chan struct{}, 4)
	env.Present.SetFlushFunc(func() (int, error) {
		env.UI.ClearDirty()
		flushes <- struct{}{}
		return 1, nil
	})
	env.Present.Start(rt.Context())
	defer env.Present.Close()

	_, err := rt.RunString(context.Background(), `
		const state = require("loupedeck/state");
		const ui = require("loupedeck/ui");
		const count = state.signal(0);
		ui.page("counter", page => {
		  page.tile(0, 0, tile => {
		    tile.text(() => String(count.get()));
		  });
		});
		ui.onButton("Circle", () => {
		  count.update(v => v + 1);
		});
		ui.show("counter");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	select {
	case <-flushes:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial auto flush")
	}

	source.emitButton(device.Circle, device.ButtonDown)

	select {
	case <-flushes:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for button-triggered auto flush")
	}
}

func TestAnimModuleCanDriveSignalTweenFromButtonEvent(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	env.Anim.FrameInterval = 5 * time.Millisecond
	rt := NewRuntime(env)
	defer func() { _ = rt.Close(context.Background()) }()
	env = rt.Env

	_, err := rt.RunString(context.Background(), `
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

	source.emitButton(device.Circle, device.ButtonDown)
	time.Sleep(80 * time.Millisecond)
	if got := env.UI.Page("home").Tile(0, 0).Text(); got != "9" {
		t.Fatalf("expected tweened value 9, got %q", got)
	}
}

func TestAnimModuleLoopCanDriveReactiveUpdates(t *testing.T) {
	env := envpkg.Ensure(nil)
	env.Anim.FrameInterval = 5 * time.Millisecond
	rt := NewRuntime(env)
	defer func() { _ = rt.Close(context.Background()) }()
	env = rt.Env

	_, err := rt.RunString(context.Background(), `
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

	deadline := time.Now().Add(200 * time.Millisecond)
	for {
		if got := env.UI.Page("home").Tile(0, 0).Text(); got != "0" && got != "" {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
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
	defer func() { _ = rt.Close(context.Background()) }()

	_, err := rt.RunString(context.Background(), `
		const gfx = require("loupedeck/gfx");
		const base = gfx.surface(16, 16);
		base.batch(() => {
		  base.clear(0);
		  base.set(1, 1, 30);
		  base.add(1, 1, 20);
		  base.line(0, 0, 15, 0, 50);
		  base.crosshatch(0, 0, 8, 8, 2, 20);
		  base.text("EYE", { x: 0, y: 0, width: 16, height: 16, brightness: 120, center: true });
		  const overlay = gfx.surface(4, 4);
		  overlay.fillRect(0, 0, 4, 4, 100);
		  base.compositeAdd(overlay, 2, 2);
		});
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

func TestGfxModuleCanLoadFontHandleAndUseItForText(t *testing.T) {
	fontPath := filepath.Join(t.TempDir(), "Go-Regular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o644); err != nil {
		t.Fatalf("write font: %v", err)
	}

	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()

	_, err := rt.RunString(context.Background(), fmt.Sprintf(`
		const gfx = require("loupedeck/gfx");
		const font = gfx.font(%q, { size: 12, dpi: 72 });
		const s = gfx.surface(32, 16);
		s.text("A", { x: 0, y: 0, width: 32, height: 16, brightness: 255, center: true, font });
		globalThis.__gfxFont = font;
		globalThis.__gfxSurface = s;
	`, fontPath))
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	fontObj := rt.VM.Get("__gfxFont").ToObject(rt.VM)
	pathFn, ok := goja.AssertFunction(fontObj.Get("path"))
	if !ok {
		t.Fatal("expected gfx font to expose path()")
	}
	pathValue, err := pathFn(fontObj)
	if err != nil {
		t.Fatalf("call path(): %v", err)
	}
	if pathValue.String() != fontPath {
		t.Fatalf("expected font path %q, got %q", fontPath, pathValue.String())
	}

	surfaceObj := rt.VM.Get("__gfxSurface").ToObject(rt.VM)
	exported := surfaceObj.Get("__surface").Export()
	surface, ok := exported.(*gfx.Surface)
	if !ok || surface == nil {
		t.Fatal("expected exported gfx surface")
	}
	found := false
	for y := 0; y < surface.Height() && !found; y++ {
		for x := 0; x < surface.Width(); x++ {
			if surface.At(x, y) != 0 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected rendered text pixels from loaded font")
	}
}

func TestGfxModuleCanRenderKanjiFromCollectionFontWhenAvailable(t *testing.T) {
	const cjkCollection = "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc"
	if _, err := os.Stat(cjkCollection); err != nil {
		t.Skipf("CJK collection not available: %v", err)
	}

	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()

	_, err := rt.RunString(context.Background(), fmt.Sprintf(`
		const gfx = require("loupedeck/gfx");
		const font = gfx.font(%q, { size: 18, dpi: 72, index: 0 });
		const s = gfx.surface(32, 24);
		s.text("渦", { x: 0, y: 0, width: 32, height: 24, brightness: 255, center: true, font });
		globalThis.__kanjiSurface = s;
	`, cjkCollection))
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	surfaceObj := rt.VM.Get("__kanjiSurface").ToObject(rt.VM)
	exported := surfaceObj.Get("__surface").Export()
	surface, ok := exported.(*gfx.Surface)
	if !ok || surface == nil {
		t.Fatal("expected exported kanji surface")
	}
	found := false
	for y := 0; y < surface.Height() && !found; y++ {
		for x := 0; x < surface.Width(); x++ {
			if surface.At(x, y) != 0 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected non-zero kanji glyph pixels from collection font")
	}
}

func TestDisplayCanOwnGfxSurface(t *testing.T) {
	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()
	env := rt.Env

	_, err := rt.RunString(context.Background(), `
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
	defer func() { _ = rt.Close(context.Background()) }()
	env := rt.Env

	_, err := rt.RunString(context.Background(), `
		const ui = require("loupedeck/ui");
		const gfx = require("loupedeck/gfx");
		const base = gfx.surface(360, 270);
		const overlay = gfx.surface(360, 270);
		base.fillRect(0, 0, 5, 5, 80);
		overlay.fillRect(10, 10, 5, 5, 160);
		ui.page("home", page => {
		  page.display("main", display => {
		    display.surface(base);
		    display.layer("overlay", overlay, { r: 255, g: 32, b: 32 });
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
	fg := main.Layer("overlay").Foreground()
	if fg == nil {
		t.Fatal("expected overlay layer to retain foreground tint")
	}
	rv, gv, bv, _ := fg.RGBA()
	if rv <= gv || rv <= bv {
		t.Fatalf("expected red foreground tint, got r=%d g=%d b=%d", rv, gv, bv)
	}
}

func TestTileCanOwnGfxSurface(t *testing.T) {
	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()
	env := rt.Env

	_, err := rt.RunString(context.Background(), `
		const ui = require("loupedeck/ui");
		const gfx = require("loupedeck/gfx");
		const s = gfx.surface(90, 90);
		s.fillRect(0, 0, 5, 5, 120);
		ui.page("home", page => {
		  page.tile(0, 0, tile => {
		    tile.surface(s);
		  });
		});
		ui.show("home");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	tile := env.UI.Page("home").Tile(0, 0)
	if tile == nil || tile.Surface() == nil {
		t.Fatal("expected tile to own a gfx surface")
	}
	if got := tile.Surface().At(0, 0); got == 0 {
		t.Fatal("expected tile surface content to remain attached")
	}
}

func TestMetricsModuleRecordsCountersTimingsAndTrace(t *testing.T) {
	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()

	_, err := rt.RunString(context.Background(), `
		const metrics = require("loupedeck/metrics");
		metrics.inc("scene.frames", 2);
		metrics.observeMillis("scene.manual", 1.5);
		metrics.trace("renderAll.begin", { reason: "loop", active: 0 });
		metrics.time("scene.renderAll", () => {
		  for (let i = 0; i < 1000; i++) {}
		});
		metrics.trace("renderAll.end", { reason: "loop" });
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	snap := rt.Env.Metrics.Snapshot()
	if got := snap.Counters["scene.frames"]; got != 2 {
		t.Fatalf("expected scene.frames=2, got %d", got)
	}
	if got := snap.Timings["scene.manual"].Count; got != 1 {
		t.Fatalf("expected scene.manual count 1, got %d", got)
	}
	if got := snap.Timings["scene.renderAll"].Count; got != 1 {
		t.Fatalf("expected scene.renderAll count 1, got %d", got)
	}
	if len(snap.Trace) != 2 {
		t.Fatalf("expected 2 trace events, got %d", len(snap.Trace))
	}
	if got := snap.Trace[0].Name; got != "renderAll.begin" {
		t.Fatalf("expected first trace event renderAll.begin, got %q", got)
	}
	if got := snap.Trace[0].Fields["reason"]; got != "loop" {
		t.Fatalf("expected first trace reason loop, got %q", got)
	}
	if got := snap.Trace[0].Fields["active"]; got != "0" {
		t.Fatalf("expected first trace active 0, got %q", got)
	}
	if got := snap.Trace[1].Name; got != "renderAll.end" {
		t.Fatalf("expected second trace event renderAll.end, got %q", got)
	}
}

func TestSceneMetricsModuleProvidesReusableHelpers(t *testing.T) {
	rt := NewRuntime(nil)
	defer func() { _ = rt.Close(context.Background()) }()

	_, err := rt.RunString(context.Background(), `
		const sceneMetrics = require("loupedeck/scene-metrics").create("demo");
		sceneMetrics.recordLoopTick();
		sceneMetrics.recordActivation("T3");
		sceneMetrics.trace("renderAll.begin", { reason: "loop", active: 2 });
		sceneMetrics.recordRebuild("loop", () => {
		  sceneMetrics.timeTile("SPIRAL", () => {
		    for (let i = 0; i < 1000; i++) {}
		  });
		});
		sceneMetrics.trace("renderAll.end", { reason: "loop" });
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	snap := rt.Env.Metrics.Snapshot()
	if got := snap.Counters["demo.loopTicks"]; got != 1 {
		t.Fatalf("expected demo.loopTicks=1, got %d", got)
	}
	if got := snap.Counters["demo.activations"]; got != 1 {
		t.Fatalf("expected demo.activations=1, got %d", got)
	}
	if got := snap.Counters["demo.activations.touch"]; got != 1 {
		t.Fatalf("expected demo.activations.touch=1, got %d", got)
	}
	if got := snap.Counters["demo.renderAll.calls"]; got != 1 {
		t.Fatalf("expected demo.renderAll.calls=1, got %d", got)
	}
	if got := snap.Counters["demo.renderAll.reason.loop"]; got != 1 {
		t.Fatalf("expected demo.renderAll.reason.loop=1, got %d", got)
	}
	if got := snap.Timings["demo.renderAll"].Count; got != 1 {
		t.Fatalf("expected demo.renderAll timing count 1, got %d", got)
	}
	if got := snap.Timings["demo.tile.SPIRAL"].Count; got != 1 {
		t.Fatalf("expected demo.tile.SPIRAL timing count 1, got %d", got)
	}
	if len(snap.Trace) != 2 {
		t.Fatalf("expected 2 trace events, got %d", len(snap.Trace))
	}
	if got := snap.Trace[0].Name; got != "demo.renderAll.begin" {
		t.Fatalf("expected first scene trace event demo.renderAll.begin, got %q", got)
	}
	if got := snap.Trace[0].Fields["active"]; got != "2" {
		t.Fatalf("expected first scene trace active 2, got %q", got)
	}
	if got := snap.Trace[1].Name; got != "demo.renderAll.end" {
		t.Fatalf("expected second scene trace event demo.renderAll.end, got %q", got)
	}
}

func TestPresentModuleRegistersFrameCallbackAndInvalidates(t *testing.T) {
	env := envpkg.Ensure(nil)
	rt := NewRuntime(env)
	defer func() { _ = rt.Close(context.Background()) }()
	done := make(chan string, 1)
	env.Present.SetFlushFunc(func() (int, error) {
		done <- "flushed"
		return 1, nil
	})
	env.Present.Start(rt.Context())

	_, err := rt.RunString(context.Background(), `
		const present = require("loupedeck/present");
		const metrics = require("loupedeck/metrics");
		present.onFrame(reason => {
		  metrics.trace("present.frame", { reason });
		});
		present.invalidate("initial");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for presenter flush")
	}

	snap := rt.Env.Metrics.Snapshot()
	if len(snap.Trace) != 1 {
		t.Fatalf("expected one trace event from frame callback, got %d", len(snap.Trace))
	}
	if got := snap.Trace[0].Name; got != "present.frame" {
		t.Fatalf("expected frame trace event, got %q", got)
	}
	if got := snap.Trace[0].Fields["reason"]; got != "initial" {
		t.Fatalf("expected frame reason initial, got %q", got)
	}
}

func TestPresentModuleCoalescesInvalidationsToLatestReason(t *testing.T) {
	env := envpkg.Ensure(nil)
	rt := NewRuntime(env)
	defer func() { _ = rt.Close(context.Background()) }()
	firstFlushStarted := make(chan struct{}, 1)
	releaseFirstFlush := make(chan struct{})
	secondFlushDone := make(chan struct{}, 1)
	var flushCount int32
	env.Present.SetFlushFunc(func() (int, error) {
		count := atomic.AddInt32(&flushCount, 1)
		if count == 1 {
			firstFlushStarted <- struct{}{}
			<-releaseFirstFlush
		}
		if count == 2 {
			secondFlushDone <- struct{}{}
		}
		return 1, nil
	})
	env.Present.Start(rt.Context())

	_, err := rt.RunString(context.Background(), `
		const present = require("loupedeck/present");
		const metrics = require("loupedeck/metrics");
		present.onFrame(reason => {
		  metrics.trace("present.frame", { reason });
		});
		present.invalidate("initial");
	`)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	select {
	case <-firstFlushStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first flush start")
	}

	_, err = rt.RunString(context.Background(), `
		present.invalidate("loop-1");
		present.invalidate("loop-2");
		present.invalidate("loop-3");
	`)
	if err != nil {
		t.Fatalf("run script 2: %v", err)
	}
	close(releaseFirstFlush)
	select {
	case <-secondFlushDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second flush done")
	}

	snap := rt.Env.Metrics.Snapshot()
	if len(snap.Trace) != 2 {
		t.Fatalf("expected two frame trace events, got %d", len(snap.Trace))
	}
	if got := snap.Trace[0].Fields["reason"]; got != "initial" {
		t.Fatalf("expected first frame reason initial, got %q", got)
	}
	if got := snap.Trace[1].Fields["reason"]; got != "loop-3" {
		t.Fatalf("expected second frame reason loop-3, got %q", got)
	}
}

func TestCloseRemovesRuntimeBridgeBindings(t *testing.T) {
	rt := NewRuntime(nil)
	if _, ok := runtimebridge.Lookup(rt.VM); !ok {
		t.Fatal("expected bindings before close")
	}
	if err := rt.Close(context.Background()); err != nil {
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
	defer func() { _ = rt.Close(context.Background()) }()
	env = rt.Env

	_, err := rt.RunString(context.Background(), `
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
		go source.emitButton(device.Circle, device.ButtonDown)
	}
	waitForText(t, env.UI.Page("home").Tile(0, 0), "25")
}

func TestButtonEventAfterCloseDoesNotApplyMutation(t *testing.T) {
	source := newFakeSource()
	env := envpkg.Ensure(nil)
	env.Host.Attach(source)
	rt := NewRuntime(env)
	env = rt.Env

	_, err := rt.RunString(context.Background(), `
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
	if err := rt.Close(context.Background()); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	source.emitButton(device.Circle, device.ButtonDown)
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
