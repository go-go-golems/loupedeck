package host

func (r *Runtime) OnShow(page string, fn func(string)) Subscription {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.next()
	if r.showHooks[page] == nil {
		r.showHooks[page] = map[uint64]func(string){}
	}
	r.showHooks[page][id] = fn
	return &eventSubscription{closeFn: func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if hooks, ok := r.showHooks[page]; ok {
			delete(hooks, id)
			if len(hooks) == 0 {
				delete(r.showHooks, page)
			}
		}
	}}
}

func (r *Runtime) Show(page string) error {
	if err := r.UI.Show(page); err != nil {
		return err
	}
	r.mu.Lock()
	hooks := make([]func(string), 0, len(r.showHooks[page]))
	for _, hook := range r.showHooks[page] {
		hooks = append(hooks, hook)
	}
	r.mu.Unlock()
	for _, hook := range hooks {
		hook(page)
	}
	return nil
}

// ReplayActivePage marks the currently active retained page dirty again so a
// renderer can redraw it after reconnect. It deliberately does not rerun page
// show hooks; replay should restore visuals, not re-trigger page-entry logic.
func (r *Runtime) ReplayActivePage() bool {
	return r.UI.InvalidateActivePage()
}
