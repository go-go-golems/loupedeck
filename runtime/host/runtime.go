package host

import (
	"sync"

	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

type EventSource interface {
	OnButton(device.Button, device.ButtonFunc) device.Subscription
	OnTouch(device.TouchButton, device.TouchFunc) device.Subscription
	OnKnob(device.Knob, device.KnobFunc) device.Subscription
}

type Runtime struct {
	UI *ui.UI

	mu          sync.Mutex
	source      EventSource
	nextID      uint64
	closed      bool
	buttons     map[uint64]*buttonBinding
	touches     map[uint64]*touchBinding
	knobs       map[uint64]*knobBinding
	showHooks   map[string]map[uint64]func(string)
	managedTime map[*Timer]struct{}
}

func New(uiRuntime *ui.UI) *Runtime {
	if uiRuntime == nil {
		uiRuntime = ui.New(nil)
	}
	return &Runtime{
		UI:          uiRuntime,
		buttons:     map[uint64]*buttonBinding{},
		touches:     map[uint64]*touchBinding{},
		knobs:       map[uint64]*knobBinding{},
		showHooks:   map[string]map[uint64]func(string){},
		managedTime: map[*Timer]struct{}{},
	}
}

func (r *Runtime) next() uint64 {
	r.nextID++
	return r.nextID
}

func (r *Runtime) Attach(source EventSource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	for _, binding := range r.buttons {
		binding.closeSourceSub()
	}
	for _, binding := range r.touches {
		binding.closeSourceSub()
	}
	for _, binding := range r.knobs {
		binding.closeSourceSub()
	}
	r.source = source
	for _, binding := range r.buttons {
		binding.attach(source)
	}
	for _, binding := range r.touches {
		binding.attach(source)
	}
	for _, binding := range r.knobs {
		binding.attach(source)
	}
}

func (r *Runtime) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	for _, timer := range r.snapshotTimersLocked() {
		timer.Stop()
	}
	for _, binding := range r.buttons {
		binding.closeSourceSub()
	}
	for _, binding := range r.touches {
		binding.closeSourceSub()
	}
	for _, binding := range r.knobs {
		binding.closeSourceSub()
	}
	r.mu.Unlock()
}

func (r *Runtime) snapshotTimersLocked() []*Timer {
	ret := make([]*Timer, 0, len(r.managedTime))
	for timer := range r.managedTime {
		ret = append(ret, timer)
	}
	return ret
}

func (r *Runtime) addTimer(timer *Timer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.managedTime[timer] = struct{}{}
}

func (r *Runtime) removeTimer(timer *Timer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.managedTime, timer)
}
