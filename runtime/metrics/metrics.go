package metrics

import (
	"sort"
	"sync"
	"time"
)

const DefaultTraceLimit = 500

type TimingStats struct {
	Count      uint64
	TotalNanos uint64
	MaxNanos   uint64
}

type TraceEvent struct {
	Seq           uint64
	TimeUnixNanos int64
	Name          string
	Fields        map[string]string
}

type Snapshot struct {
	Counters map[string]int64
	Timings  map[string]TimingStats
	Trace    []TraceEvent
}

type Collector struct {
	mu         sync.Mutex
	counters   map[string]int64
	timings    map[string]TimingStats
	trace      []TraceEvent
	nextSeq    uint64
	traceLimit int
}

func New() *Collector {
	return NewWithTraceLimit(DefaultTraceLimit)
}

func NewWithTraceLimit(limit int) *Collector {
	if limit < 0 {
		limit = 0
	}
	return &Collector{
		counters:   map[string]int64{},
		timings:    map[string]TimingStats{},
		trace:      make([]TraceEvent, 0, limit),
		traceLimit: limit,
	}
}

func (c *Collector) Inc(name string, delta int64) {
	if c == nil || name == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name] += delta
}

func (c *Collector) ObserveDuration(name string, d time.Duration) {
	if c == nil || name == "" {
		return
	}
	if d < 0 {
		d = 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	stat := c.timings[name]
	stat.Count++
	stat.TotalNanos += uint64(d)
	if uint64(d) > stat.MaxNanos {
		stat.MaxNanos = uint64(d)
	}
	c.timings[name] = stat
}

func (c *Collector) ObserveMillis(name string, ms float64) {
	if ms < 0 {
		ms = 0
	}
	c.ObserveDuration(name, time.Duration(ms*float64(time.Millisecond)))
}

func (c *Collector) Trace(name string, fields map[string]string) {
	if c == nil || name == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextSeq++
	event := TraceEvent{
		Seq:           c.nextSeq,
		TimeUnixNanos: time.Now().UnixNano(),
		Name:          name,
		Fields:        copyFields(fields),
	}
	if c.traceLimit <= 0 {
		return
	}
	if len(c.trace) < c.traceLimit {
		c.trace = append(c.trace, event)
		return
	}
	copy(c.trace, c.trace[1:])
	c.trace[len(c.trace)-1] = event
}

func (c *Collector) Snapshot() Snapshot {
	if c == nil {
		return Snapshot{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return snapshotLocked(c.counters, c.timings, c.trace)
}

func (c *Collector) SnapshotAndReset() Snapshot {
	if c == nil {
		return Snapshot{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	s := snapshotLocked(c.counters, c.timings, c.trace)
	c.counters = map[string]int64{}
	c.timings = map[string]TimingStats{}
	c.trace = make([]TraceEvent, 0, c.traceLimit)
	return s
}

func snapshotLocked(counters map[string]int64, timings map[string]TimingStats, trace []TraceEvent) Snapshot {
	counterCopy := make(map[string]int64, len(counters))
	for k, v := range counters {
		counterCopy[k] = v
	}
	timingCopy := make(map[string]TimingStats, len(timings))
	for k, v := range timings {
		timingCopy[k] = v
	}
	traceCopy := make([]TraceEvent, len(trace))
	for i, event := range trace {
		traceCopy[i] = TraceEvent{
			Seq:           event.Seq,
			TimeUnixNanos: event.TimeUnixNanos,
			Name:          event.Name,
			Fields:        copyFields(event.Fields),
		}
	}
	return Snapshot{Counters: counterCopy, Timings: timingCopy, Trace: traceCopy}
}

func copyFields(fields map[string]string) map[string]string {
	if len(fields) == 0 {
		return nil
	}
	out := make(map[string]string, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return out
}

func CounterKeys(s Snapshot) []string {
	keys := make([]string, 0, len(s.Counters))
	for k := range s.Counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TimingKeys(s Snapshot) []string {
	keys := make([]string, 0, len(s.Timings))
	for k := range s.Timings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
