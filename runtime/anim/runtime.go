package anim

import (
	"sync"
	"time"

	"github.com/go-go-golems/loupedeck/runtime/easing"
	"github.com/go-go-golems/loupedeck/runtime/host"
)

type Runtime struct {
	Host          *host.Runtime
	FrameInterval time.Duration
}

type Handle struct {
	once   sync.Once
	stopFn func()
}

type TweenStep struct {
	Get      func() float64
	Set      func(float64)
	To       float64
	Duration time.Duration
	Ease     easing.Func
}

type Timeline struct {
	rt    *Runtime
	steps []TweenStep
}

func New(hostRuntime *host.Runtime) *Runtime {
	return &Runtime{
		Host:          hostRuntime,
		FrameInterval: 16 * time.Millisecond,
	}
}

func (h *Handle) Stop() {
	if h == nil {
		return
	}
	h.once.Do(func() {
		if h.stopFn != nil {
			h.stopFn()
		}
	})
}

func (r *Runtime) TweenFloat64(get func() float64, set func(float64), to float64, duration time.Duration, ease easing.Func) *Handle {
	return r.tweenFloat64(get, set, to, duration, ease, nil)
}

func (r *Runtime) tweenFloat64(get func() float64, set func(float64), to float64, duration time.Duration, ease easing.Func, onDone func()) *Handle {
	if ease == nil {
		ease = easing.Linear
	}
	if duration <= 0 {
		set(to)
		if onDone != nil {
			onDone()
		}
		return &Handle{}
	}
	start := get()
	startTime := time.Now()
	handle := &Handle{}
	step := func() {
		elapsed := time.Since(startTime)
		if elapsed >= duration {
			set(to)
			handle.Stop()
			if onDone != nil {
				onDone()
			}
			return
		}
		progress := float64(elapsed) / float64(duration)
		eased := ease(progress)
		set(start + (to-start)*eased)
	}
	step()
	interval := r.Host.SetInterval(r.FrameInterval, step)
	handle.stopFn = func() {
		interval.Stop()
	}
	return handle
}

func (r *Runtime) Loop(duration time.Duration, fn func(float64)) *Handle {
	if duration <= 0 {
		duration = time.Second
	}
	start := time.Now()
	handle := &Handle{}
	step := func() {
		elapsed := time.Since(start)
		phase := float64(elapsed%duration) / float64(duration)
		fn(phase)
	}
	step()
	interval := r.Host.SetInterval(r.FrameInterval, step)
	handle.stopFn = func() {
		interval.Stop()
	}
	return handle
}

func (r *Runtime) Timeline() *Timeline {
	return &Timeline{rt: r}
}

func (t *Timeline) To(get func() float64, set func(float64), to float64, duration time.Duration, ease easing.Func) *Timeline {
	t.steps = append(t.steps, TweenStep{Get: get, Set: set, To: to, Duration: duration, Ease: ease})
	return t
}

func (t *Timeline) Play() *Handle {
	handle := &Handle{}
	if len(t.steps) == 0 {
		return handle
	}
	var current *Handle
	var stopMu sync.Mutex
	playStep := func(index int) {}
	playStep = func(index int) {
		if index >= len(t.steps) {
			return
		}
		step := t.steps[index]
		current = t.rt.tweenFloat64(step.Get, step.Set, step.To, step.Duration, step.Ease, func() {
			playStep(index + 1)
		})
	}
	handle.stopFn = func() {
		stopMu.Lock()
		defer stopMu.Unlock()
		if current != nil {
			current.Stop()
		}
	}
	playStep(0)
	return handle
}
