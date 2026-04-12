package reactive

import "fmt"

type Subscription interface {
	Stop()
}

type Effect struct {
	rt         *Runtime
	fn         func()
	active     bool
	dirty      bool
	queued     bool
	evaluating bool
	deps       dependencySet
}

func (e *Effect) Stop() {
	if !e.active {
		return
	}
	e.active = false
	e.dirty = false
	if e.queued {
		delete(e.rt.pendingEffects, e)
		e.queued = false
	}
	e.deps.clear(e)
}

func (e *Effect) markDirty() {
	if !e.active {
		return
	}
	e.dirty = true
	e.rt.enqueueEffect(e)
}

func (e *Effect) trackDependency(source dependencySource) {
	e.deps.track(e, source)
}

func (e *Effect) run() {
	if e.evaluating {
		panic(fmt.Sprintf("reactive: reentrant effect execution for %T", e))
	}
	e.evaluating = true
	e.deps.clear(e)
	defer func() {
		e.evaluating = false
	}()

	e.rt.withCollector(e, func() {
		e.fn()
	})
	e.dirty = false
}
