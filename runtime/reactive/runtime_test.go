package reactive

import "testing"

func TestSignalSetEqualValueDoesNotRerunWatchers(t *testing.T) {
	rt := NewRuntime()
	s := NewSignal(rt, 1)

	runs := 0
	seen := 0
	rt.Watch(func() {
		runs++
		seen = s.Get()
	})

	if runs != 1 {
		t.Fatalf("expected initial watch run once, got %d", runs)
	}
	if seen != 1 {
		t.Fatalf("expected initial seen value 1, got %d", seen)
	}

	s.Set(1)
	if runs != 1 {
		t.Fatalf("expected equal set to avoid rerun, got %d runs", runs)
	}

	s.Update(func(v int) int { return v + 1 })
	if runs != 2 {
		t.Fatalf("expected update to rerun watcher, got %d runs", runs)
	}
	if seen != 2 {
		t.Fatalf("expected updated seen value 2, got %d", seen)
	}
}

func TestComputedInvalidationChain(t *testing.T) {
	rt := NewRuntime()
	base := NewSignal(rt, 2)

	doubleRuns := 0
	double := NewComputed(rt, func() int {
		doubleRuns++
		return base.Get() * 2
	})

	labelRuns := 0
	label := NewComputed(rt, func() int {
		labelRuns++
		return double.Get() + 1
	})

	watchRuns := 0
	seen := 0
	rt.Watch(func() {
		watchRuns++
		seen = label.Get()
	})

	if doubleRuns != 1 || labelRuns != 1 || watchRuns != 1 {
		t.Fatalf("expected initial chain to evaluate once each, got double=%d label=%d watch=%d", doubleRuns, labelRuns, watchRuns)
	}
	if seen != 5 {
		t.Fatalf("expected initial seen value 5, got %d", seen)
	}

	base.Set(5)
	if doubleRuns != 2 || labelRuns != 2 || watchRuns != 2 {
		t.Fatalf("expected chain rerun once each after update, got double=%d label=%d watch=%d", doubleRuns, labelRuns, watchRuns)
	}
	if seen != 11 {
		t.Fatalf("expected updated seen value 11, got %d", seen)
	}
}

func TestDiamondDependencyGraphDoesNotDoubleEvaluateDownstreamComputed(t *testing.T) {
	rt := NewRuntime()
	base := NewSignal(rt, 2)

	leftRuns := 0
	left := NewComputed(rt, func() int {
		leftRuns++
		return base.Get() + 1
	})

	rightRuns := 0
	right := NewComputed(rt, func() int {
		rightRuns++
		return base.Get() * 2
	})

	totalRuns := 0
	total := NewComputed(rt, func() int {
		totalRuns++
		return left.Get() + right.Get()
	})

	watchRuns := 0
	seen := 0
	rt.Watch(func() {
		watchRuns++
		seen = total.Get()
	})

	base.Set(3)

	if leftRuns != 2 || rightRuns != 2 || totalRuns != 2 {
		t.Fatalf("expected one downstream reevaluation per computed, got left=%d right=%d total=%d", leftRuns, rightRuns, totalRuns)
	}
	if watchRuns != 2 {
		t.Fatalf("expected watcher to run twice total, got %d", watchRuns)
	}
	if seen != 10 {
		t.Fatalf("expected seen value 10, got %d", seen)
	}
}

func TestBatchDefersWatcherFlushUntilOuterBatchCompletes(t *testing.T) {
	rt := NewRuntime()
	a := NewSignal(rt, 0)
	b := NewSignal(rt, 0)

	runs := 0
	seen := 0
	rt.Watch(func() {
		runs++
		seen = a.Get() + b.Get()
	})

	rt.Batch(func() {
		a.Set(1)
		b.Set(2)
		a.Set(3)
	})

	if runs != 2 {
		t.Fatalf("expected a single rerun after first batch, got %d", runs)
	}
	if seen != 5 {
		t.Fatalf("expected batch result 5, got %d", seen)
	}

	rt.Batch(func() {
		a.Set(4)
		rt.Batch(func() {
			b.Set(5)
			a.Set(6)
		})
	})

	if runs != 3 {
		t.Fatalf("expected nested batch to add one rerun, got %d", runs)
	}
	if seen != 11 {
		t.Fatalf("expected nested batch result 11, got %d", seen)
	}
}

func TestComputedCyclePanics(t *testing.T) {
	rt := NewRuntime()
	var c *Computed[int]
	c = NewComputed(rt, func() int {
		return c.Get()
	})

	defer func() {
		if recover() == nil {
			t.Fatal("expected cyclic computed evaluation to panic")
		}
	}()

	_ = c.Get()
}

func TestEffectReentrancyPanics(t *testing.T) {
	rt := NewRuntime()
	effect := &Effect{rt: rt, active: true, dirty: true}
	effect.fn = func() {
		effect.run()
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected reentrant effect execution to panic")
		}
	}()

	effect.run()
}

func TestStoppingWatchDetachesDependencies(t *testing.T) {
	rt := NewRuntime()
	s := NewSignal(rt, 0)
	runs := 0
	effect := rt.Watch(func() {
		runs++
		_ = s.Get()
	})

	effect.Stop()
	s.Set(1)

	if runs != 1 {
		t.Fatalf("expected stopped watcher not to rerun, got %d runs", runs)
	}
}
