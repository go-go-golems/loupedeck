package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
	"github.com/go-go-golems/loupedeck/runtime/render"
)

func main() {
	scriptPath := flag.String("script", "", "Path to a JS file")
	devicePath := flag.String("device", "", "Optional serial device path (default: auto-detect)")
	duration := flag.Duration("duration", 15*time.Second, "How long to run before exiting; 0 means run until interrupted")
	queueSize := flag.Int("queue-size", 256, "Writer queue size")
	sendInterval := flag.Duration("send-interval", 35*time.Millisecond, "Writer pacing interval")
	logEvents := flag.Bool("log-events", false, "Log high-level button/touch/knob events")
	logRenderStats := flag.Bool("log-render-stats", false, "Log retained renderer flush statistics")
	logWriterStats := flag.Bool("log-writer-stats", false, "Log writer queue/send statistics")
	logJSStats := flag.Bool("log-js-stats", false, "Log JS-side metrics recorded through loupedeck/metrics")
	logJSTrace := flag.Bool("log-js-trace", false, "Log JS-side ordered trace events recorded through loupedeck/metrics")
	logGoTrace := flag.Bool("log-go-trace", false, "Log Go-side ordered trace events around flush activity")
	traceLimit := flag.Int("trace-limit", metrics.DefaultTraceLimit, "Maximum number of ordered trace events to retain between dumps")
	traceDumpOnExit := flag.Bool("trace-dump-on-exit", true, "Dump any remaining trace events and JS stats once more before exit")
	statsInterval := flag.Duration("stats-interval", time.Second, "Interval for periodic stats logging")
	exitOnCircle := flag.Bool("exit-on-circle", true, "Exit when the Circle button is pressed")
	flag.Parse()

	if *scriptPath == "" {
		fmt.Fprintln(os.Stderr, "missing --script")
		os.Exit(2)
	}
	if extras := flag.Args(); len(extras) > 0 {
		fmt.Fprintf(os.Stderr, "unexpected positional arguments after flags: %q\n", extras)
		fmt.Fprintln(os.Stderr, "hint: did you mean to pass a flag like --send-interval instead of a bare argument?")
		os.Exit(2)
	}
	script, err := os.ReadFile(*scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read script: %v\n", err)
		os.Exit(1)
	}

	writerOptions := device.WriterOptions{QueueSize: *queueSize, SendInterval: *sendInterval}
	var deckConn *device.Loupedeck
	if *devicePath == "" {
		deckConn, err = device.ConnectAutoWithOptions(writerOptions)
	} else {
		deckConn, err = device.ConnectPathWithOptions(*devicePath, writerOptions)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("closing loupedeck connection")
		if err := deckConn.Close(); err != nil {
			slog.Warn("close failed", "error", err)
		}
	}()
	displays := map[string]*device.Display{
		"left":  deckConn.GetDisplay("left"),
		"main":  deckConn.GetDisplay("main"),
		"right": deckConn.GetDisplay("right"),
	}
	if displays["main"] == nil {
		fmt.Fprintln(os.Stderr, "missing main display")
		os.Exit(1)
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- deckConn.Listen()
	}()

	env := envpkg.Ensure(&envpkg.Environment{Metrics: metrics.NewWithTraceLimit(*traceLimit)})
	env.Host.Attach(deckConn)
	if *logEvents {
		registerEventLogging(env)
	}
	rt := jsruntime.NewRuntime(env)
	defer rt.Close(nil)
	if _, err := rt.RunString(rt.Context(), string(script)); err != nil {
		fmt.Fprintf(os.Stderr, "run script: %v\n", err)
		os.Exit(1)
	}

	renderer := render.NewWithDisplays(rt.Env.UI, map[string]render.DrawTarget{
		"left":  displays["left"],
		"main":  displays["main"],
		"right": displays["right"],
	})
	renderer.Theme = render.Theme{
		Background: color.Black,
		Foreground: color.White,
		Accent:     color.White,
	}

	exitCh := make(chan struct{}, 1)
	if *exitOnCircle {
		rt.Env.Host.OnButton(device.Circle, func(device.Button, device.ButtonStatus) {
			select {
			case exitCh <- struct{}{}:
			default:
			}
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var statsTicker *time.Ticker
	var statsTick <-chan time.Time
	if *statsInterval > 0 && (*logRenderStats || *logWriterStats || *logJSStats || *logJSTrace || *logGoTrace) {
		statsTicker = time.NewTicker(*statsInterval)
		defer statsTicker.Stop()
		statsTick = statsTicker.C
	}
	var timeout <-chan time.Time
	if *duration > 0 {
		timer := time.NewTimer(*duration)
		defer timer.Stop()
		timeout = timer.C
	}

	renderWindow := renderStatsWindow{}
	var renderWindowMu sync.Mutex
	lastWriterStats := deckConn.WriterStats()
	dumpMetricsWindow := func(label string) {
		if !*logJSStats && !*logJSTrace && !*logGoTrace {
			return
		}
		snap := rt.Env.Metrics.SnapshotAndReset()
		if *logJSStats {
			slog.Info("js stats", "script", *scriptPath, "label", label, "counters", formatJSCounters(snap), "timings", formatJSTimings(snap))
		}
		if *logJSTrace {
			for _, event := range filterTraceEvents(snap.Trace, false) {
				slog.Info("js trace", "script", *scriptPath, "label", label, "seq", event.Seq, "event", event.Name, "fields", formatTraceFields(event.Fields))
			}
		}
		if *logGoTrace {
			for _, event := range filterTraceEvents(snap.Trace, true) {
				slog.Info("go trace", "script", *scriptPath, "label", label, "seq", event.Seq, "event", event.Name, "fields", formatTraceFields(event.Fields))
			}
		}
	}
	rt.Env.Present.SetFlushFunc(func() (int, error) {
		dirtyDisplays := len(rt.Env.UI.DirtyDisplays())
		dirtyTiles := len(rt.Env.UI.DirtyTiles())
		if *logGoTrace {
			rt.Env.Metrics.Trace("go.flush.begin", map[string]string{"dirtyDisplays": fmt.Sprintf("%d", dirtyDisplays), "dirtyTiles": fmt.Sprintf("%d", dirtyTiles)})
		}
		start := time.Now()
		flushed := renderer.Flush()
		elapsed := time.Since(start)
		if *logGoTrace {
			rt.Env.Metrics.Trace("go.flush.end", map[string]string{"ops": fmt.Sprintf("%d", flushed), "elapsedMs": fmt.Sprintf("%.2f", float64(elapsed)/float64(time.Millisecond))})
		}
		if *logRenderStats {
			renderWindowMu.Lock()
			renderWindow.Record(dirtyDisplays, dirtyTiles, flushed, elapsed)
			renderWindowMu.Unlock()
		}
		return flushed, nil
	})
	rt.Env.Present.Start(rt.Context())
	defer rt.Env.Present.Close()

	slog.Info("Loupedeck JS live runner started", "script", *scriptPath, "duration", *duration, "send_interval", *sendInterval, "log_render_stats", *logRenderStats, "log_writer_stats", *logWriterStats, "log_js_stats", *logJSStats, "log_js_trace", *logJSTrace, "log_go_trace", *logGoTrace, "trace_limit", *traceLimit)
	exitRunner := func(reason string, attrs ...any) {
		logAttrs := []any{"reason", reason, "script", *scriptPath}
		logAttrs = append(logAttrs, attrs...)
		slog.Info("Loupedeck JS live runner exiting", logAttrs...)
		if *traceDumpOnExit {
			dumpMetricsWindow("final")
		}
		clearDisplays(displays)
	}
	for {
		select {
		case <-statsTick:
			if *logRenderStats {
				renderWindowMu.Lock()
				slog.Info("render stats", "script", *scriptPath, "stats", renderWindow.String())
				renderWindow = renderStatsWindow{}
				renderWindowMu.Unlock()
			}
			if *logWriterStats {
				current := deckConn.WriterStats()
				delta := diffWriterStats(lastWriterStats, current)
				slog.Info("writer stats", "script", *scriptPath, "delta", delta, "current", current)
				lastWriterStats = current
			}
			dumpMetricsWindow("interval")
		case err := <-listenErrCh:
			if err != nil {
				fmt.Fprintf(os.Stderr, "listen: %v\n", err)
				exitRunner("listen-error", "error", err)
			} else {
				exitRunner("listen-stopped")
			}
			return
		case sig := <-sigCh:
			exitRunner("signal", "signal", sig.String())
			return
		case <-exitCh:
			exitRunner("circle-button")
			return
		case <-timeout:
			exitRunner("timeout", "duration", *duration)
			return
		}
	}
}

func clearDisplays(displays map[string]*device.Display) {
	for _, display := range displays {
		if display == nil {
			continue
		}
		im := image.NewRGBA(image.Rect(0, 0, display.Width(), display.Height()))
		draw.Draw(im, im.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
		display.Draw(im, 0, 0)
	}
	time.Sleep(100 * time.Millisecond)
}

func registerEventLogging(env *envpkg.Environment) {
	if env == nil {
		return
	}
	for _, button := range []device.Button{
		device.Circle,
		device.Button1,
		device.Button2,
		device.Button3,
		device.Button4,
		device.Button5,
		device.Button6,
		device.Button7,
	} {
		button := button
		env.Host.OnButton(button, func(b device.Button, s device.ButtonStatus) {
			slog.Info("button event", "button", b.String(), "status", s.String())
		})
	}
	for _, touch := range []device.TouchButton{
		device.Touch1,
		device.Touch2,
		device.Touch3,
		device.Touch4,
		device.Touch5,
		device.Touch6,
		device.Touch7,
		device.Touch8,
		device.Touch9,
		device.Touch10,
		device.Touch11,
		device.Touch12,
	} {
		touch := touch
		env.Host.OnTouch(touch, func(t device.TouchButton, s device.ButtonStatus, x, y uint16) {
			slog.Info("touch event", "touch", t.String(), "status", s.String(), "x", x, "y", y)
		})
	}
	for _, knob := range []device.Knob{
		device.Knob1,
		device.Knob2,
		device.Knob3,
		device.Knob4,
		device.Knob5,
		device.Knob6,
	} {
		knob := knob
		env.Host.OnKnob(knob, func(k device.Knob, value int) {
			slog.Info("knob event", "knob", k.String(), "value", value)
		})
	}
}

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
