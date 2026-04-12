package reactive

// Runtime coordinates dependency tracking, batched mutation flushes, and eager
// watcher execution for the reactive graph. It is intentionally single-threaded
// for now; host/runtime code should mutate it from one scheduling goroutine.
type Runtime struct {
	current        dependencyCollector
	batchDepth     int
	flushing       bool
	pendingEffects map[*Effect]struct{}
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func NewSignal[T any](r *Runtime, initial T) *Signal[T] {
	return NewSignalWithEqual(r, initial, defaultEqual[T])
}

func NewSignalWithEqual[T any](r *Runtime, initial T, equal func(a, b T) bool) *Signal[T] {
	if equal == nil {
		equal = defaultEqual[T]
	}
	return &Signal[T]{
		rt:    r,
		value: initial,
		equal: equal,
	}
}

func NewComputed[T any](r *Runtime, fn func() T) *Computed[T] {
	return &Computed[T]{
		rt:    r,
		fn:    fn,
		dirty: true,
	}
}

func (r *Runtime) Watch(fn func()) *Effect {
	e := &Effect{
		rt:     r,
		fn:     fn,
		active: true,
		dirty:  true,
	}
	r.enqueueEffect(e)
	r.maybeFlush()
	return e
}

func (r *Runtime) Batch(fn func()) {
	r.batchDepth++
	defer func() {
		r.batchDepth--
		if r.batchDepth == 0 {
			r.Flush()
		}
	}()
	fn()
}

func (r *Runtime) Flush() {
	if r.flushing {
		return
	}
	r.flushing = true
	defer func() {
		r.flushing = false
	}()

	for len(r.pendingEffects) > 0 {
		effects := make([]*Effect, 0, len(r.pendingEffects))
		for effect := range r.pendingEffects {
			effects = append(effects, effect)
		}
		for _, effect := range effects {
			delete(r.pendingEffects, effect)
			effect.queued = false
		}
		for _, effect := range effects {
			if effect.active && effect.dirty {
				effect.run()
			}
		}
	}
}

func (r *Runtime) maybeFlush() {
	if r.batchDepth > 0 || r.flushing {
		return
	}
	r.Flush()
}

func (r *Runtime) trackDependency(source dependencySource) {
	if r.current == nil {
		return
	}
	r.current.trackDependency(source)
}

func (r *Runtime) withCollector(collector dependencyCollector, fn func()) {
	prev := r.current
	r.current = collector
	defer func() {
		r.current = prev
	}()
	fn()
}

func (r *Runtime) enqueueEffect(effect *Effect) {
	if !effect.active {
		return
	}
	if effect.queued {
		return
	}
	if r.pendingEffects == nil {
		r.pendingEffects = map[*Effect]struct{}{}
	}
	r.pendingEffects[effect] = struct{}{}
	effect.queued = true
}
