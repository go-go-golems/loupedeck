package loupedeck

import "sync"

// Subscription represents a registered event listener that can be removed.
type Subscription interface {
	Close() error
}

type listenerSubscription struct {
	once    sync.Once
	closeFn func()
}

func (s *listenerSubscription) Close() error {
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

func (l *Loupedeck) nextListenerID() uint64 {
	l.listenerMutex.Lock()
	defer l.listenerMutex.Unlock()
	l.listenerID++
	return l.listenerID
}

func (l *Loupedeck) ensureListenerMapsLocked() {
	if l.buttonBindings == nil {
		l.buttonBindings = make(map[Button]ButtonFunc)
	}
	if l.buttonUpBindings == nil {
		l.buttonUpBindings = make(map[Button]ButtonFunc)
	}
	if l.knobBindings == nil {
		l.knobBindings = make(map[Knob]KnobFunc)
	}
	if l.touchBindings == nil {
		l.touchBindings = make(map[TouchButton]TouchFunc)
	}
	if l.touchUpBindings == nil {
		l.touchUpBindings = make(map[TouchButton]TouchFunc)
	}
	if l.buttonListeners == nil {
		l.buttonListeners = make(map[Button]map[uint64]ButtonFunc)
	}
	if l.buttonUpListeners == nil {
		l.buttonUpListeners = make(map[Button]map[uint64]ButtonFunc)
	}
	if l.knobListeners == nil {
		l.knobListeners = make(map[Knob]map[uint64]KnobFunc)
	}
	if l.touchListeners == nil {
		l.touchListeners = make(map[TouchButton]map[uint64]TouchFunc)
	}
	if l.touchUpListeners == nil {
		l.touchUpListeners = make(map[TouchButton]map[uint64]TouchFunc)
	}
}

func (l *Loupedeck) OnButton(b Button, f ButtonFunc) Subscription {
	id := l.nextListenerID()
	l.listenerMutex.Lock()
	defer l.listenerMutex.Unlock()
	l.ensureListenerMapsLocked()
	if l.buttonListeners[b] == nil {
		l.buttonListeners[b] = map[uint64]ButtonFunc{}
	}
	l.buttonListeners[b][id] = f
	return &listenerSubscription{closeFn: func() {
		l.listenerMutex.Lock()
		defer l.listenerMutex.Unlock()
		delete(l.buttonListeners[b], id)
		if len(l.buttonListeners[b]) == 0 {
			delete(l.buttonListeners, b)
		}
	}}
}

func (l *Loupedeck) OnButtonUp(b Button, f ButtonFunc) Subscription {
	id := l.nextListenerID()
	l.listenerMutex.Lock()
	defer l.listenerMutex.Unlock()
	l.ensureListenerMapsLocked()
	if l.buttonUpListeners[b] == nil {
		l.buttonUpListeners[b] = map[uint64]ButtonFunc{}
	}
	l.buttonUpListeners[b][id] = f
	return &listenerSubscription{closeFn: func() {
		l.listenerMutex.Lock()
		defer l.listenerMutex.Unlock()
		delete(l.buttonUpListeners[b], id)
		if len(l.buttonUpListeners[b]) == 0 {
			delete(l.buttonUpListeners, b)
		}
	}}
}

func (l *Loupedeck) OnKnob(k Knob, f KnobFunc) Subscription {
	id := l.nextListenerID()
	l.listenerMutex.Lock()
	defer l.listenerMutex.Unlock()
	l.ensureListenerMapsLocked()
	if l.knobListeners[k] == nil {
		l.knobListeners[k] = map[uint64]KnobFunc{}
	}
	l.knobListeners[k][id] = f
	return &listenerSubscription{closeFn: func() {
		l.listenerMutex.Lock()
		defer l.listenerMutex.Unlock()
		delete(l.knobListeners[k], id)
		if len(l.knobListeners[k]) == 0 {
			delete(l.knobListeners, k)
		}
	}}
}

func (l *Loupedeck) OnTouch(b TouchButton, f TouchFunc) Subscription {
	id := l.nextListenerID()
	l.listenerMutex.Lock()
	defer l.listenerMutex.Unlock()
	l.ensureListenerMapsLocked()
	if l.touchListeners[b] == nil {
		l.touchListeners[b] = map[uint64]TouchFunc{}
	}
	l.touchListeners[b][id] = f
	return &listenerSubscription{closeFn: func() {
		l.listenerMutex.Lock()
		defer l.listenerMutex.Unlock()
		delete(l.touchListeners[b], id)
		if len(l.touchListeners[b]) == 0 {
			delete(l.touchListeners, b)
		}
	}}
}

func (l *Loupedeck) OnTouchUp(b TouchButton, f TouchFunc) Subscription {
	id := l.nextListenerID()
	l.listenerMutex.Lock()
	defer l.listenerMutex.Unlock()
	l.ensureListenerMapsLocked()
	if l.touchUpListeners[b] == nil {
		l.touchUpListeners[b] = map[uint64]TouchFunc{}
	}
	l.touchUpListeners[b][id] = f
	return &listenerSubscription{closeFn: func() {
		l.listenerMutex.Lock()
		defer l.listenerMutex.Unlock()
		delete(l.touchUpListeners[b], id)
		if len(l.touchUpListeners[b]) == 0 {
			delete(l.touchUpListeners, b)
		}
	}}
}

func (l *Loupedeck) dispatchButton(button Button, status ButtonStatus) bool {
	var primary ButtonFunc
	listeners := make([]ButtonFunc, 0)

	l.listenerMutex.RLock()
	if status == ButtonDown {
		primary = l.buttonBindings[button]
		for _, fn := range l.buttonListeners[button] {
			listeners = append(listeners, fn)
		}
	} else {
		primary = l.buttonUpBindings[button]
		for _, fn := range l.buttonUpListeners[button] {
			listeners = append(listeners, fn)
		}
	}
	l.listenerMutex.RUnlock()

	called := false
	if primary != nil {
		called = true
		primary(button, status)
	}
	for _, fn := range listeners {
		called = true
		fn(button, status)
	}
	return called
}

func (l *Loupedeck) dispatchKnob(knob Knob, value int) bool {
	l.listenerMutex.RLock()
	primary := l.knobBindings[knob]
	listeners := make([]KnobFunc, 0, len(l.knobListeners[knob]))
	for _, fn := range l.knobListeners[knob] {
		listeners = append(listeners, fn)
	}
	l.listenerMutex.RUnlock()

	called := false
	if primary != nil {
		called = true
		primary(knob, value)
	}
	for _, fn := range listeners {
		called = true
		fn(knob, value)
	}
	return called
}

func (l *Loupedeck) dispatchTouch(button TouchButton, status ButtonStatus, x, y uint16) bool {
	var primary TouchFunc
	listeners := make([]TouchFunc, 0)

	l.listenerMutex.RLock()
	if status == ButtonDown {
		primary = l.touchBindings[button]
		for _, fn := range l.touchListeners[button] {
			listeners = append(listeners, fn)
		}
	} else {
		primary = l.touchUpBindings[button]
		for _, fn := range l.touchUpListeners[button] {
			listeners = append(listeners, fn)
		}
	}
	l.listenerMutex.RUnlock()

	called := false
	if primary != nil {
		called = true
		primary(button, status, x, y)
	}
	for _, fn := range listeners {
		called = true
		fn(button, status, x, y)
	}
	return called
}
