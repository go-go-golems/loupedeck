package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
)

type options struct {
	ScriptPath      string
	DevicePath      string
	Duration        time.Duration
	QueueSize       int
	SendInterval    time.Duration
	FlushInterval   time.Duration
	LogEvents       bool
	LogRenderStats  bool
	LogWriterStats  bool
	LogJSStats      bool
	LogJSTrace      bool
	LogGoTrace      bool
	TraceLimit      int
	TraceDumpOnExit bool
	StatsInterval   time.Duration
	ExitOnCircle    bool
}

func parseOptions() (options, error) {
	scriptPath := flag.String("script", "", "Path to a JS file")
	devicePath := flag.String("device", "", "Optional serial device path (default: auto-detect)")
	duration := flag.Duration("duration", 15*time.Second, "How long to run before exiting; 0 means run until interrupted")
	queueSize := flag.Int("queue-size", 256, "Writer queue size")
	sendInterval := flag.Duration("send-interval", 35*time.Millisecond, "Writer pacing interval")
	flushInterval := flag.Duration("flush-interval", device.DefaultRenderOptions.FlushInterval, "Retained render scheduler flush interval")
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
		return options{}, errors.New("missing --script")
	}
	if extras := flag.Args(); len(extras) > 0 {
		return options{}, fmt.Errorf("unexpected positional arguments after flags: %q\nhint: did you mean to pass a flag like --send-interval instead of a bare argument?", extras)
	}
	if *flushInterval <= 0 {
		return options{}, fmt.Errorf("--flush-interval must be > 0, got %s", *flushInterval)
	}

	return options{
		ScriptPath:      *scriptPath,
		DevicePath:      *devicePath,
		Duration:        *duration,
		QueueSize:       *queueSize,
		SendInterval:    *sendInterval,
		FlushInterval:   *flushInterval,
		LogEvents:       *logEvents,
		LogRenderStats:  *logRenderStats,
		LogWriterStats:  *logWriterStats,
		LogJSStats:      *logJSStats,
		LogJSTrace:      *logJSTrace,
		LogGoTrace:      *logGoTrace,
		TraceLimit:      *traceLimit,
		TraceDumpOnExit: *traceDumpOnExit,
		StatsInterval:   *statsInterval,
		ExitOnCircle:    *exitOnCircle,
	}, nil
}
