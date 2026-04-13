package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
)

type renderStatsWindow struct {
	FlushTicks      int
	NonEmptyFlushes int
	FlushedDisplays int
	FlushedTiles    int
	FlushedOps      int
	TotalRender     time.Duration
	MaxRender       time.Duration
}

func (w *renderStatsWindow) Record(dirtyDisplays, dirtyTiles, flushedOps int, elapsed time.Duration) {
	if w == nil {
		return
	}
	w.FlushTicks++
	if dirtyDisplays == 0 && dirtyTiles == 0 && flushedOps == 0 {
		return
	}
	w.NonEmptyFlushes++
	w.FlushedDisplays += dirtyDisplays
	w.FlushedTiles += dirtyTiles
	w.FlushedOps += flushedOps
	w.TotalRender += elapsed
	if elapsed > w.MaxRender {
		w.MaxRender = elapsed
	}
}

func (w renderStatsWindow) String() string {
	avgMs := 0.0
	if w.NonEmptyFlushes > 0 {
		avgMs = float64(w.TotalRender) / float64(time.Millisecond) / float64(w.NonEmptyFlushes)
	}
	return fmt.Sprintf("flush_ticks=%d non_empty_flushes=%d displays=%d tiles=%d ops=%d avg_render_ms=%.2f max_render_ms=%.2f",
		w.FlushTicks,
		w.NonEmptyFlushes,
		w.FlushedDisplays,
		w.FlushedTiles,
		w.FlushedOps,
		avgMs,
		float64(w.MaxRender)/float64(time.Millisecond),
	)
}

func diffWriterStats(a, b device.WriterStats) device.WriterStats {
	return device.WriterStats{
		QueuedCommands:    b.QueuedCommands - a.QueuedCommands,
		SentCommands:      b.SentCommands - a.SentCommands,
		SentMessages:      b.SentMessages - a.SentMessages,
		FailedCommands:    b.FailedCommands - a.FailedCommands,
		MaxQueueDepth:     maxInt(a.MaxQueueDepth, b.MaxQueueDepth),
		CurrentQueueDepth: b.CurrentQueueDepth,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatJSCounters(s metrics.Snapshot) string {
	keys := metrics.CounterKeys(s)
	if len(keys) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, s.Counters[key]))
	}
	return strings.Join(parts, ", ")
}

func formatJSTimings(s metrics.Snapshot) string {
	keys := metrics.TimingKeys(s)
	if len(keys) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		stat := s.Timings[key]
		avgMs := 0.0
		if stat.Count > 0 {
			avgMs = float64(stat.TotalNanos) / float64(time.Millisecond) / float64(stat.Count)
		}
		parts = append(parts, fmt.Sprintf("%s[count=%d avg_ms=%.2f max_ms=%.2f]", key, stat.Count, avgMs, float64(stat.MaxNanos)/float64(time.Millisecond)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "; ")
}

func filterTraceEvents(events []metrics.TraceEvent, goOnly bool) []metrics.TraceEvent {
	filtered := make([]metrics.TraceEvent, 0, len(events))
	for _, event := range events {
		isGo := strings.HasPrefix(event.Name, "go.")
		if goOnly == isGo {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func formatTraceFields(fields map[string]string) string {
	if len(fields) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, fields[key]))
	}
	return strings.Join(parts, ", ")
}
