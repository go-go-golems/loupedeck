package loupedeck

import "testing"

func newTestLoupedeck() *Loupedeck {
	return &Loupedeck{
		buttonBindings:    make(map[Button]ButtonFunc),
		buttonUpBindings:  make(map[Button]ButtonFunc),
		knobBindings:      make(map[Knob]KnobFunc),
		touchBindings:     make(map[TouchButton]TouchFunc),
		touchUpBindings:   make(map[TouchButton]TouchFunc),
		buttonListeners:   make(map[Button]map[uint64]ButtonFunc),
		buttonUpListeners: make(map[Button]map[uint64]ButtonFunc),
		knobListeners:     make(map[Knob]map[uint64]KnobFunc),
		touchListeners:    make(map[TouchButton]map[uint64]TouchFunc),
		touchUpListeners:  make(map[TouchButton]map[uint64]TouchFunc),
	}
}

func TestOnButtonSupportsMultipleListenersAndCleanup(t *testing.T) {
	l := newTestLoupedeck()

	calls := []string{}
	sub1 := l.OnButton(Circle, func(Button, ButtonStatus) {
		calls = append(calls, "listener-1")
	})
	sub2 := l.OnButton(Circle, func(Button, ButtonStatus) {
		calls = append(calls, "listener-2")
	})

	if !l.dispatchButton(Circle, ButtonDown) {
		t.Fatalf("expected dispatchButton to report handled event")
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 listeners to fire, got %d (%v)", len(calls), calls)
	}

	if err := sub1.Close(); err != nil {
		t.Fatalf("close sub1: %v", err)
	}
	calls = nil

	if !l.dispatchButton(Circle, ButtonDown) {
		t.Fatalf("expected dispatchButton to report handled event after unsubscribe")
	}
	if len(calls) != 1 || calls[0] != "listener-2" {
		t.Fatalf("expected only listener-2 after unsubscribe, got %v", calls)
	}

	if err := sub2.Close(); err != nil {
		t.Fatalf("close sub2: %v", err)
	}
	if l.dispatchButton(Circle, ButtonDown) {
		t.Fatalf("expected dispatchButton to report unhandled event after all unsubscribed")
	}
}

func TestBindAndOnKnobCoexist(t *testing.T) {
	l := newTestLoupedeck()

	calls := []string{}
	l.BindKnob(Knob1, func(Knob, int) {
		calls = append(calls, "bind")
	})
	l.OnKnob(Knob1, func(Knob, int) {
		calls = append(calls, "on")
	})

	if !l.dispatchKnob(Knob1, 1) {
		t.Fatalf("expected dispatchKnob to report handled event")
	}
	if len(calls) != 2 {
		t.Fatalf("expected both bind and listener to fire, got %v", calls)
	}
}

func TestOnTouchUpCleanup(t *testing.T) {
	l := newTestLoupedeck()

	count := 0
	sub := l.OnTouchUp(Touch1, func(TouchButton, ButtonStatus, uint16, uint16) {
		count++
	})

	if !l.dispatchTouch(Touch1, ButtonUp, 10, 20) {
		t.Fatalf("expected touch up to be handled")
	}
	if count != 1 {
		t.Fatalf("expected one touch-up callback, got %d", count)
	}

	if err := sub.Close(); err != nil {
		t.Fatalf("close touch subscription: %v", err)
	}
	if l.dispatchTouch(Touch1, ButtonUp, 10, 20) {
		t.Fatalf("expected touch up to be unhandled after unsubscribe")
	}
}
