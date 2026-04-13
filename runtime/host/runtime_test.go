package host

import (
	"sync"
	"testing"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

type fakeSource struct {
	mu      sync.Mutex
	buttons map[device.Button][]device.ButtonFunc
	touches map[device.TouchButton][]device.TouchFunc
	knobs   map[device.Knob][]device.KnobFunc
}

func newFakeSource() *fakeSource {
	return &fakeSource{
		buttons: map[device.Button][]device.ButtonFunc{},
		touches: map[device.TouchButton][]device.TouchFunc{},
		knobs:   map[device.Knob][]device.KnobFunc{},
	}
}

type fakeSub struct {
	closeFn func()
}

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

func (f *fakeSource) OnTouch(touch device.TouchButton, fn device.TouchFunc) device.Subscription {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.touches[touch] = append(f.touches[touch], fn)
	idx := len(f.touches[touch]) - 1
	return &fakeSub{closeFn: func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if idx < len(f.touches[touch]) {
			f.touches[touch] = append(f.touches[touch][:idx], f.touches[touch][idx+1:]...)
		}
	}}
}

func (f *fakeSource) OnKnob(knob device.Knob, fn device.KnobFunc) device.Subscription {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.knobs[knob] = append(f.knobs[knob], fn)
	idx := len(f.knobs[knob]) - 1
	return &fakeSub{closeFn: func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if idx < len(f.knobs[knob]) {
			f.knobs[knob] = append(f.knobs[knob][:idx], f.knobs[knob][idx+1:]...)
		}
	}}
}

func (f *fakeSource) emitButton(button device.Button, status device.ButtonStatus) {
	f.mu.Lock()
	callbacks := append([]device.ButtonFunc(nil), f.buttons[button]...)
	f.mu.Unlock()
	for _, cb := range callbacks {
		cb(button, status)
	}
}

func (f *fakeSource) emitTouch(touch device.TouchButton, status device.ButtonStatus, x, y uint16) {
	f.mu.Lock()
	callbacks := append([]device.TouchFunc(nil), f.touches[touch]...)
	f.mu.Unlock()
	for _, cb := range callbacks {
		cb(touch, status, x, y)
	}
}

func (f *fakeSource) emitKnob(knob device.Knob, value int) {
	f.mu.Lock()
	callbacks := append([]device.KnobFunc(nil), f.knobs[knob]...)
	f.mu.Unlock()
	for _, cb := range callbacks {
		cb(knob, value)
	}
}

func TestAttachedEventSourceDeliversCallbacks(t *testing.T) {
	r := New(ui.New(nil))
	source := newFakeSource()
	r.Attach(source)

	buttonCalls := 0
	touchCalls := 0
	knobCalls := 0

	r.OnButton(device.Circle, func(device.Button, device.ButtonStatus) {
		buttonCalls++
	})
	r.OnTouch(device.Touch1, func(device.TouchButton, device.ButtonStatus, uint16, uint16) {
		touchCalls++
	})
	r.OnKnob(device.Knob1, func(device.Knob, int) {
		knobCalls++
	})

	source.emitButton(device.Circle, device.ButtonDown)
	source.emitTouch(device.Touch1, device.ButtonDown, 10, 20)
	source.emitKnob(device.Knob1, 3)

	if buttonCalls != 1 || touchCalls != 1 || knobCalls != 1 {
		t.Fatalf("unexpected callback counts button=%d touch=%d knob=%d", buttonCalls, touchCalls, knobCalls)
	}
}

func TestShowInvokesPageHook(t *testing.T) {
	uiRuntime := ui.New(nil)
	uiRuntime.AddPage("home")
	r := New(uiRuntime)

	called := 0
	r.OnShow("home", func(page string) {
		called++
		if page != "home" {
			t.Fatalf("unexpected page %q", page)
		}
	})

	if err := r.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected show hook to fire once, got %d", called)
	}
}

func TestTimersFireAndStop(t *testing.T) {
	r := New(nil)
	defer r.Close()

	timeoutCh := make(chan struct{}, 1)
	intervalCh := make(chan struct{}, 4)

	r.SetTimeout(10*time.Millisecond, func() {
		timeoutCh <- struct{}{}
	})
	interval := r.SetInterval(10*time.Millisecond, func() {
		intervalCh <- struct{}{}
	})

	select {
	case <-timeoutCh:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout did not fire")
	}

	for i := 0; i < 2; i++ {
		select {
		case <-intervalCh:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("interval did not fire enough times")
		}
	}

	interval.Stop()
	remaining := len(intervalCh)
	time.Sleep(40 * time.Millisecond)
	if len(intervalCh) > remaining {
		t.Fatal("interval continued firing after stop")
	}
}

func TestReplayActivePageMarksTilesDirtyWithoutRerunningShowHooks(t *testing.T) {
	uiRuntime := ui.New(nil)
	page := uiRuntime.AddPage("home")
	tile := page.AddTile(0, 0)
	tile.SetText("REC")
	r := New(uiRuntime)

	showCalls := 0
	r.OnShow("home", func(string) {
		showCalls++
	})
	if err := r.Show("home"); err != nil {
		t.Fatalf("show home: %v", err)
	}
	if showCalls != 1 {
		t.Fatalf("expected one show hook call, got %d", showCalls)
	}
	uiRuntime.ClearDirty()
	if tile.Dirty() {
		t.Fatal("expected tile to be clean after explicit clear")
	}

	if !r.ReplayActivePage() {
		t.Fatal("expected replay to report active page")
	}
	if !tile.Dirty() {
		t.Fatal("expected replay to mark active tile dirty")
	}
	if showCalls != 1 {
		t.Fatalf("expected replay not to rerun show hooks, got %d calls", showCalls)
	}
}
