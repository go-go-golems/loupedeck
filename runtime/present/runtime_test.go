package present

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInvalidateCoalescesWhileFlushBusy(t *testing.T) {
	rt := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	renders := []string{}
	flushes := int32(0)
	firstFlushStarted := make(chan struct{}, 1)
	releaseFirstFlush := make(chan struct{})
	secondFlushDone := make(chan struct{}, 1)

	rt.SetRenderFunc(func(reason string) error {
		mu.Lock()
		renders = append(renders, reason)
		mu.Unlock()
		return nil
	})
	rt.SetFlushFunc(func() (int, error) {
		count := atomic.AddInt32(&flushes, 1)
		if count == 1 {
			firstFlushStarted <- struct{}{}
			<-releaseFirstFlush
		}
		if count == 2 {
			secondFlushDone <- struct{}{}
		}
		return 1, nil
	})

	rt.Start(ctx)
	rt.Invalidate("initial")
	<-firstFlushStarted
	rt.Invalidate("loop-1")
	rt.Invalidate("loop-2")
	rt.Invalidate("loop-3")
	close(releaseFirstFlush)

	select {
	case <-secondFlushDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second flush")
	}

	mu.Lock()
	defer mu.Unlock()
	want := []string{"initial", "loop-3"}
	if !reflect.DeepEqual(renders, want) {
		t.Fatalf("unexpected render reasons: got %v want %v", renders, want)
	}
}

func TestInvalidateBeforeCallbacksIsPresentedOnceCallbacksInstalled(t *testing.T) {
	rt := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rt.Start(ctx)
	rt.Invalidate("initial")

	done := make(chan struct{}, 1)
	rt.SetRenderFunc(func(reason string) error {
		if reason != "initial" {
			t.Fatalf("unexpected reason %q", reason)
		}
		return nil
	})
	rt.SetFlushFunc(func() (int, error) {
		done <- struct{}{}
		return 1, nil
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for deferred present")
	}
}

func TestPresentationRunsSerially(t *testing.T) {
	rt := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var inCallback int32
	flushDone := make(chan struct{}, 2)
	rt.SetRenderFunc(func(reason string) error {
		if got := atomic.AddInt32(&inCallback, 1); got != 1 {
			t.Fatalf("expected render serialization, got inCallback=%d", got)
		}
		defer atomic.AddInt32(&inCallback, -1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	rt.SetFlushFunc(func() (int, error) {
		if got := atomic.LoadInt32(&inCallback); got != 0 {
			t.Fatalf("expected flush after render completed, got inCallback=%d", got)
		}
		flushDone <- struct{}{}
		return 1, nil
	})

	rt.Start(ctx)
	rt.Invalidate("a")
	rt.Invalidate("b")

	select {
	case <-flushDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for flush")
	}
}

func TestCloseStopsFurtherWork(t *testing.T) {
	rt := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var renders int32
	rt.SetRenderFunc(func(reason string) error {
		atomic.AddInt32(&renders, 1)
		return nil
	})
	rt.SetFlushFunc(func() (int, error) { return 1, nil })
	rt.Start(ctx)
	rt.Close()
	rt.Invalidate("after-close")
	time.Sleep(20 * time.Millisecond)
	if got := atomic.LoadInt32(&renders); got != 0 {
		t.Fatalf("expected no renders after close, got %d", got)
	}
}
