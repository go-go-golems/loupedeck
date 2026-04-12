package host

import (
	"sync"
	"time"
)

type Timer struct {
	once   sync.Once
	stopFn func()
}

func (t *Timer) Stop() {
	if t == nil {
		return
	}
	t.once.Do(func() {
		if t.stopFn != nil {
			t.stopFn()
		}
	})
}

func (r *Runtime) SetTimeout(delay time.Duration, fn func()) *Timer {
	timer := &Timer{}
	inner := time.AfterFunc(delay, func() {
		defer r.removeTimer(timer)
		fn()
	})
	timer.stopFn = func() {
		inner.Stop()
		r.removeTimer(timer)
	}
	r.addTimer(timer)
	return timer
}

func (r *Runtime) SetInterval(interval time.Duration, fn func()) *Timer {
	if interval <= 0 {
		interval = time.Millisecond
	}
	stopCh := make(chan struct{})
	ticker := time.NewTicker(interval)
	timer := &Timer{}
	go func() {
		defer func() {
			ticker.Stop()
			r.removeTimer(timer)
		}()
		for {
			select {
			case <-ticker.C:
				fn()
			case <-stopCh:
				return
			}
		}
	}()
	timer.stopFn = func() {
		close(stopCh)
	}
	r.addTimer(timer)
	return timer
}
