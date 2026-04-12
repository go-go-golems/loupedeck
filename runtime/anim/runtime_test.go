package anim

import (
	"math"
	"testing"
	"time"

	"github.com/go-go-golems/loupedeck/runtime/easing"
	"github.com/go-go-golems/loupedeck/runtime/host"
)

func TestTweenFloat64ReachesTarget(t *testing.T) {
	rt := New(host.New(nil))
	rt.FrameInterval = 5 * time.Millisecond
	defer rt.Host.Close()

	value := 0.0
	h := rt.TweenFloat64(
		func() float64 { return value },
		func(v float64) { value = v },
		10,
		25*time.Millisecond,
		easing.Linear,
	)
	defer h.Stop()
	time.Sleep(80 * time.Millisecond)

	if math.Abs(value-10) > 0.2 {
		t.Fatalf("expected tween to reach ~10, got %f", value)
	}
}

func TestLoopAdvancesPhase(t *testing.T) {
	rt := New(host.New(nil))
	rt.FrameInterval = 5 * time.Millisecond
	defer rt.Host.Close()

	phases := make(chan float64, 8)
	h := rt.Loop(20*time.Millisecond, func(v float64) {
		phases <- v
	})
	defer h.Stop()

	seen := 0
	deadline := time.After(100 * time.Millisecond)
	for seen < 3 {
		select {
		case <-phases:
			seen++
		case <-deadline:
			t.Fatal("loop did not advance enough phases")
		}
	}
}

func TestTimelineRunsSequentialTweens(t *testing.T) {
	rt := New(host.New(nil))
	rt.FrameInterval = 5 * time.Millisecond
	defer rt.Host.Close()

	value := 0.0
	timeline := rt.Timeline().
		To(func() float64 { return value }, func(v float64) { value = v }, 5, 20*time.Millisecond, easing.Linear).
		To(func() float64 { return value }, func(v float64) { value = v }, 9, 20*time.Millisecond, easing.Linear)

	h := timeline.Play()
	defer h.Stop()
	time.Sleep(80 * time.Millisecond)

	if math.Abs(value-9) > 0.2 {
		t.Fatalf("expected timeline to finish near 9, got %f", value)
	}
}
