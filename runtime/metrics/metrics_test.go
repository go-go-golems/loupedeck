package metrics

import (
	"testing"
	"time"
)

func TestCollectorSnapshotAndReset(t *testing.T) {
	c := New()
	c.Inc("scene.frames", 2)
	c.ObserveDuration("scene.render", 5*time.Millisecond)
	c.ObserveDuration("scene.render", 7*time.Millisecond)
	c.Trace("renderAll.begin", map[string]string{"reason": "loop"})

	s := c.SnapshotAndReset()
	if got := s.Counters["scene.frames"]; got != 2 {
		t.Fatalf("expected scene.frames=2, got %d", got)
	}
	stat := s.Timings["scene.render"]
	if stat.Count != 2 {
		t.Fatalf("expected timing count 2, got %d", stat.Count)
	}
	if stat.MaxNanos != uint64(7*time.Millisecond) {
		t.Fatalf("expected max 7ms, got %d", stat.MaxNanos)
	}
	if len(s.Trace) != 1 {
		t.Fatalf("expected one trace event, got %d", len(s.Trace))
	}
	if s.Trace[0].Name != "renderAll.begin" || s.Trace[0].Fields["reason"] != "loop" {
		t.Fatalf("unexpected trace event: %+v", s.Trace[0])
	}
	if next := c.Snapshot(); len(next.Counters) != 0 || len(next.Timings) != 0 || len(next.Trace) != 0 {
		t.Fatalf("expected reset collector, got %+v", next)
	}
}

func TestCollectorTracePreservesOrderAndSequence(t *testing.T) {
	c := NewWithTraceLimit(10)
	c.Trace("loop.tick", map[string]string{"phase": "0.1"})
	c.Trace("renderAll.begin", map[string]string{"reason": "loop"})
	c.Trace("renderAll.end", map[string]string{"reason": "loop"})

	s := c.Snapshot()
	if len(s.Trace) != 3 {
		t.Fatalf("expected 3 trace events, got %d", len(s.Trace))
	}
	for i, want := range []string{"loop.tick", "renderAll.begin", "renderAll.end"} {
		if got := s.Trace[i].Name; got != want {
			t.Fatalf("trace[%d] name=%q, want %q", i, got, want)
		}
		if gotSeq := s.Trace[i].Seq; gotSeq != uint64(i+1) {
			t.Fatalf("trace[%d] seq=%d, want %d", i, gotSeq, i+1)
		}
	}
}

func TestCollectorTraceHonorsBoundedBuffer(t *testing.T) {
	c := NewWithTraceLimit(2)
	c.Trace("a", nil)
	c.Trace("b", nil)
	c.Trace("c", map[string]string{"k": "v"})

	s := c.Snapshot()
	if len(s.Trace) != 2 {
		t.Fatalf("expected bounded trace length 2, got %d", len(s.Trace))
	}
	if s.Trace[0].Name != "b" || s.Trace[1].Name != "c" {
		t.Fatalf("unexpected bounded trace contents: %+v", s.Trace)
	}
	if s.Trace[0].Seq != 2 || s.Trace[1].Seq != 3 {
		t.Fatalf("unexpected trace sequence numbers: %+v", s.Trace)
	}
}

func TestCollectorSnapshotCopiesTraceFields(t *testing.T) {
	c := NewWithTraceLimit(2)
	fields := map[string]string{"reason": "loop"}
	c.Trace("renderAll.begin", fields)
	fields["reason"] = "mutated"

	s := c.Snapshot()
	if got := s.Trace[0].Fields["reason"]; got != "loop" {
		t.Fatalf("expected snapshot to preserve original field value, got %q", got)
	}
	s.Trace[0].Fields["reason"] = "changed-again"
	again := c.Snapshot()
	if got := again.Trace[0].Fields["reason"]; got != "loop" {
		t.Fatalf("expected collector trace fields to be copied defensively, got %q", got)
	}
}
