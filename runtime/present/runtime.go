package present

import (
	"context"
	"sync"
)

type RenderFunc func(reason string) error

type FlushFunc func() (int, error)

type Runtime struct {
	mu          sync.Mutex
	render      RenderFunc
	flush       FlushFunc
	dirty       bool
	dirtyReason string
	wakeCh      chan struct{}
	closeCh     chan struct{}
	closeOnce   sync.Once
	startOnce   sync.Once
	doneCh      chan struct{}
	closed      bool
}

func New() *Runtime {
	return &Runtime{
		render:  func(string) error { return nil },
		wakeCh:  make(chan struct{}, 1),
		closeCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

func (r *Runtime) SetRenderFunc(fn RenderFunc) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.render = fn
	r.mu.Unlock()
	r.signalWake()
}

func (r *Runtime) SetFlushFunc(fn FlushFunc) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.flush = fn
	r.mu.Unlock()
	r.signalWake()
}

func (r *Runtime) Invalidate(reason string) {
	if r == nil {
		return
	}
	if reason == "" {
		reason = "unknown"
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.dirty = true
	r.dirtyReason = reason
	r.mu.Unlock()
	r.signalWake()
}

func (r *Runtime) Start(ctx context.Context) {
	if r == nil {
		return
	}
	r.startOnce.Do(func() {
		go r.loop(ctx)
	})
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		close(r.closeCh)
		r.mu.Unlock()
	})
	<-r.doneCh
}

func (r *Runtime) loop(ctx context.Context) {
	defer close(r.doneCh)
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.closeCh:
			return
		case <-r.wakeCh:
		}
		for {
			render, flush, reason, ok := r.nextWork()
			if !ok {
				break
			}
			if render != nil {
				if err := render(reason); err != nil {
					continue
				}
			}
			if flush != nil {
				_, _ = flush()
			}
		}
	}
}

func (r *Runtime) nextWork() (RenderFunc, FlushFunc, string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || !r.dirty || r.render == nil || r.flush == nil {
		return nil, nil, "", false
	}
	reason := r.dirtyReason
	r.dirty = false
	r.dirtyReason = ""
	return r.render, r.flush, reason, true
}

func (r *Runtime) signalWake() {
	if r == nil {
		return
	}
	select {
	case r.wakeCh <- struct{}{}:
	default:
	}
}
