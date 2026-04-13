package main

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
	"github.com/go-go-golems/loupedeck/runtime/render"
)

func run(opts options) error {
	script, err := os.ReadFile(opts.ScriptPath)
	if err != nil {
		return fmt.Errorf("read script: %w", err)
	}

	writerOptions := device.WriterOptions{QueueSize: opts.QueueSize, SendInterval: opts.SendInterval}
	var deckConn *device.Loupedeck
	if opts.DevicePath == "" {
		deckConn, err = device.ConnectAutoWithOptions(writerOptions)
	} else {
		deckConn, err = device.ConnectPathWithOptions(opts.DevicePath, writerOptions)
	}
	if err != nil {
		return fmt.Errorf("connect: %w", err)
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
		return fmt.Errorf("missing main display")
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- deckConn.Listen()
	}()

	env := envpkg.Ensure(&envpkg.Environment{Metrics: metrics.NewWithTraceLimit(opts.TraceLimit)})
	env.Host.Attach(deckConn)
	if opts.LogEvents {
		registerEventLogging(env)
	}

	rt := jsruntime.NewRuntime(env)
	defer rt.Close(nil)
	if _, err := rt.RunString(rt.Context(), string(script)); err != nil {
		return fmt.Errorf("run script: %w", err)
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
	if opts.ExitOnCircle {
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
	if opts.StatsInterval > 0 && (opts.LogRenderStats || opts.LogWriterStats || opts.LogJSStats || opts.LogJSTrace || opts.LogGoTrace) {
		statsTicker = time.NewTicker(opts.StatsInterval)
		defer statsTicker.Stop()
		statsTick = statsTicker.C
	}

	var timeout <-chan time.Time
	if opts.Duration > 0 {
		timer := time.NewTimer(opts.Duration)
		defer timer.Stop()
		timeout = timer.C
	}

	renderWindow := renderStatsWindow{}
	var renderWindowMu sync.Mutex
	lastWriterStats := deckConn.WriterStats()
	dumpMetricsWindow := func(label string) {
		if !opts.LogJSStats && !opts.LogJSTrace && !opts.LogGoTrace {
			return
		}
		snap := rt.Env.Metrics.SnapshotAndReset()
		if opts.LogJSStats {
			slog.Info("js stats", "script", opts.ScriptPath, "label", label, "counters", formatJSCounters(snap), "timings", formatJSTimings(snap))
		}
		if opts.LogJSTrace {
			for _, event := range filterTraceEvents(snap.Trace, false) {
				slog.Info("js trace", "script", opts.ScriptPath, "label", label, "seq", event.Seq, "event", event.Name, "fields", formatTraceFields(event.Fields))
			}
		}
		if opts.LogGoTrace {
			for _, event := range filterTraceEvents(snap.Trace, true) {
				slog.Info("go trace", "script", opts.ScriptPath, "label", label, "seq", event.Seq, "event", event.Name, "fields", formatTraceFields(event.Fields))
			}
		}
	}

	rt.Env.Present.SetFlushFunc(func() (int, error) {
		dirtyDisplays := len(rt.Env.UI.DirtyDisplays())
		dirtyTiles := len(rt.Env.UI.DirtyTiles())
		if opts.LogGoTrace {
			rt.Env.Metrics.Trace("go.flush.begin", map[string]string{"dirtyDisplays": fmt.Sprintf("%d", dirtyDisplays), "dirtyTiles": fmt.Sprintf("%d", dirtyTiles)})
		}
		start := time.Now()
		flushed := renderer.Flush()
		elapsed := time.Since(start)
		if opts.LogGoTrace {
			rt.Env.Metrics.Trace("go.flush.end", map[string]string{"ops": fmt.Sprintf("%d", flushed), "elapsedMs": fmt.Sprintf("%.2f", float64(elapsed)/float64(time.Millisecond))})
		}
		if opts.LogRenderStats {
			renderWindowMu.Lock()
			renderWindow.Record(dirtyDisplays, dirtyTiles, flushed, elapsed)
			renderWindowMu.Unlock()
		}
		return flushed, nil
	})
	rt.Env.Present.Start(rt.Context())
	defer rt.Env.Present.Close()

	slog.Info("Loupedeck JS live runner started", "script", opts.ScriptPath, "duration", opts.Duration, "send_interval", opts.SendInterval, "log_render_stats", opts.LogRenderStats, "log_writer_stats", opts.LogWriterStats, "log_js_stats", opts.LogJSStats, "log_js_trace", opts.LogJSTrace, "log_go_trace", opts.LogGoTrace, "trace_limit", opts.TraceLimit)
	exitRunner := func(reason string, attrs ...any) error {
		logAttrs := []any{"reason", reason, "script", opts.ScriptPath}
		logAttrs = append(logAttrs, attrs...)
		slog.Info("Loupedeck JS live runner exiting", logAttrs...)
		if opts.TraceDumpOnExit {
			dumpMetricsWindow("final")
		}
		clearDisplays(displays)
		return nil
	}

	for {
		select {
		case <-statsTick:
			if opts.LogRenderStats {
				renderWindowMu.Lock()
				slog.Info("render stats", "script", opts.ScriptPath, "stats", renderWindow.String())
				renderWindow = renderStatsWindow{}
				renderWindowMu.Unlock()
			}
			if opts.LogWriterStats {
				current := deckConn.WriterStats()
				delta := diffWriterStats(lastWriterStats, current)
				slog.Info("writer stats", "script", opts.ScriptPath, "delta", delta, "current", current)
				lastWriterStats = current
			}
			dumpMetricsWindow("interval")
		case err := <-listenErrCh:
			if err != nil {
				_ = exitRunner("listen-error", "error", err)
				return fmt.Errorf("listen: %w", err)
			}
			_ = exitRunner("listen-stopped")
			return nil
		case sig := <-sigCh:
			_ = exitRunner("signal", "signal", sig.String())
			return nil
		case <-exitCh:
			_ = exitRunner("circle-button")
			return nil
		case <-timeout:
			_ = exitRunner("timeout", "duration", opts.Duration)
			return nil
		}
	}
}
