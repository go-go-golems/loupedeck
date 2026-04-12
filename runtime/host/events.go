package host

import (
	"sync"

	deck "github.com/go-go-golems/loupedeck"
)

type Subscription interface {
	Close() error
}

type eventSubscription struct {
	once    sync.Once
	closeFn func()
}

func (s *eventSubscription) Close() error {
	if s == nil {
		return nil
	}
	s.once.Do(func() {
		if s.closeFn != nil {
			s.closeFn()
		}
	})
	return nil
}

type buttonBinding struct {
	button deck.Button
	fn     deck.ButtonFunc
	sub    deck.Subscription
}

type touchBinding struct {
	touch deck.TouchButton
	fn    deck.TouchFunc
	sub   deck.Subscription
}

type knobBinding struct {
	knob deck.Knob
	fn   deck.KnobFunc
	sub  deck.Subscription
}

func (b *buttonBinding) attach(source EventSource) {
	if source == nil || b.sub != nil {
		return
	}
	b.sub = source.OnButton(b.button, b.fn)
}

func (b *buttonBinding) closeSourceSub() {
	if b.sub != nil {
		_ = b.sub.Close()
		b.sub = nil
	}
}

func (b *touchBinding) attach(source EventSource) {
	if source == nil || b.sub != nil {
		return
	}
	b.sub = source.OnTouch(b.touch, b.fn)
}

func (b *touchBinding) closeSourceSub() {
	if b.sub != nil {
		_ = b.sub.Close()
		b.sub = nil
	}
}

func (b *knobBinding) attach(source EventSource) {
	if source == nil || b.sub != nil {
		return
	}
	b.sub = source.OnKnob(b.knob, b.fn)
}

func (b *knobBinding) closeSourceSub() {
	if b.sub != nil {
		_ = b.sub.Close()
		b.sub = nil
	}
}

func (r *Runtime) OnButton(button deck.Button, fn deck.ButtonFunc) Subscription {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.next()
	binding := &buttonBinding{button: button, fn: fn}
	binding.attach(r.source)
	r.buttons[id] = binding
	return &eventSubscription{closeFn: func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if binding, ok := r.buttons[id]; ok {
			binding.closeSourceSub()
			delete(r.buttons, id)
		}
	}}
}

func (r *Runtime) OnTouch(touch deck.TouchButton, fn deck.TouchFunc) Subscription {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.next()
	binding := &touchBinding{touch: touch, fn: fn}
	binding.attach(r.source)
	r.touches[id] = binding
	return &eventSubscription{closeFn: func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if binding, ok := r.touches[id]; ok {
			binding.closeSourceSub()
			delete(r.touches, id)
		}
	}}
}

func (r *Runtime) OnKnob(knob deck.Knob, fn deck.KnobFunc) Subscription {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.next()
	binding := &knobBinding{knob: knob, fn: fn}
	binding.attach(r.source)
	r.knobs[id] = binding
	return &eventSubscription{closeFn: func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if binding, ok := r.knobs[id]; ok {
			binding.closeSourceSub()
			delete(r.knobs, id)
		}
	}}
}
