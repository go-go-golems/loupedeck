package run

import (
	"context"
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

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/pkg/scriptmeta"
	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
	"github.com/go-go-golems/loupedeck/runtime/render"
)

type Command struct {
	*cmds.CommandDescription
}

type settings_ struct {
	ScriptPath      string   `glazed:"script"`
	Verb            string   `glazed:"verb"`
	VerbConfig      []string `glazed:"verb-config"`
	VerbValuesJSON  string   `glazed:"verb-values-json"`
	DevicePath      string   `glazed:"device"`
	Duration        string   `glazed:"duration"`
	QueueSize       int      `glazed:"queue-size"`
	SendInterval    string   `glazed:"send-interval"`
	FlushInterval   string   `glazed:"flush-interval"`
	LogEvents       bool     `glazed:"log-events"`
	LogRenderStats  bool     `glazed:"log-render-stats"`
	LogWriterStats  bool     `glazed:"log-writer-stats"`
	LogJSStats      bool     `glazed:"log-js-stats"`
	LogJSTrace      bool     `glazed:"log-js-trace"`
	LogGoTrace      bool     `glazed:"log-go-trace"`
	TraceLimit      int      `glazed:"trace-limit"`
	TraceDumpOnExit bool     `glazed:"trace-dump-on-exit"`
	StatsInterval   string   `glazed:"stats-interval"`
	ExitOnCircle    bool     `glazed:"exit-on-circle"`
}

type options struct {
	ScriptPath      string
	Verb            string
	VerbConfig      []string
	VerbValuesJSON  string
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

type commandResult struct {
	ScriptPath     string
	Verb           string
	DevicePath     string
	Duration       time.Duration
	SendInterval   time.Duration
	FlushInterval  time.Duration
	QueueSize      int
	ExitOnCircle   bool
	TraceLimit     int
	RequestedStats bool
	Status         string
}

var _ cmds.BareCommand = (*Command)(nil)
var _ cmds.GlazeCommand = (*Command)(nil)

func NewCommand() (*Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"run",
		cmds.WithShort("Run a Loupedeck Live JavaScript scene on hardware"),
		cmds.WithLong(`Execute a JavaScript scene against a real Loupedeck device.

Examples:
  loupedeck run --script ./examples/js/01-hello.js --duration 5s
  loupedeck run --script ./examples/js/11-cyb-os-tiles.js --send-interval 0ms --flush-interval 20ms
  loupedeck run --script ./examples/js/12-documented-scene.js --verb configure --verb-values-json '{"default":{"title":"OPS"}}'
  loupedeck run --script ./examples/js/01-hello.js --with-glaze-output --output json`),
		cmds.WithFlags(
			fields.New("script", fields.TypeString, fields.WithIsArgument(true), fields.WithHelp("Path to the JavaScript file or scene directory to execute")),
			fields.New("verb", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional jsverbs verb to invoke inside the live runtime instead of raw script execution")),
			fields.New("verb-config", fields.TypeStringList, fields.WithDefault([]string{}), fields.WithHelp("Optional config files used to resolve jsverbs field values")),
			fields.New("verb-values-json", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional JSON object or section map used to resolve jsverbs field values")),
			fields.New("device", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional serial device path (default: auto-detect)")),
			fields.New("duration", fields.TypeString, fields.WithDefault("15s"), fields.WithHelp("How long to run before exiting; use 0 to run until interrupted")),
			fields.New("queue-size", fields.TypeInteger, fields.WithDefault(256), fields.WithHelp("Writer queue size")),
			fields.New("send-interval", fields.TypeString, fields.WithDefault("35ms"), fields.WithHelp("Writer pacing interval")),
			fields.New("flush-interval", fields.TypeString, fields.WithDefault(device.DefaultRenderOptions.FlushInterval.String()), fields.WithHelp("Retained render scheduler flush interval")),
			fields.New("log-events", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log high-level button/touch/knob events")),
			fields.New("log-render-stats", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log retained renderer flush statistics")),
			fields.New("log-writer-stats", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log writer queue/send statistics")),
			fields.New("log-js-stats", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log JS-side metrics recorded through loupedeck/metrics")),
			fields.New("log-js-trace", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log JS-side ordered trace events recorded through loupedeck/metrics")),
			fields.New("log-go-trace", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log Go-side ordered trace events around flush activity")),
			fields.New("trace-limit", fields.TypeInteger, fields.WithDefault(metrics.DefaultTraceLimit), fields.WithHelp("Maximum number of ordered trace events to retain between dumps")),
			fields.New("trace-dump-on-exit", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Dump any remaining trace events and JS stats once more before exit")),
			fields.New("stats-interval", fields.TypeString, fields.WithDefault("1s"), fields.WithHelp("Interval for periodic stats logging")),
			fields.New("exit-on-circle", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Exit when the Circle button is pressed")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)

	return &Command{CommandDescription: desc}, nil
}

func (c *Command) Run(ctx context.Context, vals *values.Values) error {
	_, err := c.execute(ctx, vals)
	return err
}

func (c *Command) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	result, err := c.execute(ctx, vals)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("script", result.ScriptPath),
		types.MRP("verb", result.Verb),
		types.MRP("device", result.DevicePath),
		types.MRP("duration", result.Duration.String()),
		types.MRP("send_interval", result.SendInterval.String()),
		types.MRP("flush_interval", result.FlushInterval.String()),
		types.MRP("queue_size", result.QueueSize),
		types.MRP("exit_on_circle", result.ExitOnCircle),
		types.MRP("trace_limit", result.TraceLimit),
		types.MRP("requested_stats", result.RequestedStats),
		types.MRP("status", result.Status),
	))
}

func (c *Command) execute(ctx context.Context, vals *values.Values) (*commandResult, error) {
	s := &settings_{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return nil, err
	}
	opts, err := decodeOptions(s)
	if err != nil {
		return nil, err
	}
	if err := run(ctx, opts); err != nil {
		return nil, err
	}
	devicePath := opts.DevicePath
	if devicePath == "" {
		devicePath = "auto"
	}
	return &commandResult{
		ScriptPath:     opts.ScriptPath,
		Verb:           opts.Verb,
		DevicePath:     devicePath,
		Duration:       opts.Duration,
		SendInterval:   opts.SendInterval,
		FlushInterval:  opts.FlushInterval,
		QueueSize:      opts.QueueSize,
		ExitOnCircle:   opts.ExitOnCircle,
		TraceLimit:     opts.TraceLimit,
		RequestedStats: opts.LogRenderStats || opts.LogWriterStats || opts.LogJSStats || opts.LogJSTrace || opts.LogGoTrace,
		Status:         "ok",
	}, nil
}

func decodeOptions(s *settings_) (options, error) {
	if s.ScriptPath == "" {
		return options{}, fmt.Errorf("missing --script")
	}
	duration, err := parseDurationFlag("duration", s.Duration)
	if err != nil {
		return options{}, err
	}
	sendInterval, err := parseDurationFlag("send-interval", s.SendInterval)
	if err != nil {
		return options{}, err
	}
	flushInterval, err := parseDurationFlag("flush-interval", s.FlushInterval)
	if err != nil {
		return options{}, err
	}
	if flushInterval <= 0 {
		return options{}, fmt.Errorf("--flush-interval must be > 0, got %s", flushInterval)
	}
	statsInterval, err := parseDurationFlag("stats-interval", s.StatsInterval)
	if err != nil {
		return options{}, err
	}
	return options{
		ScriptPath:      s.ScriptPath,
		Verb:            s.Verb,
		VerbConfig:      append([]string{}, s.VerbConfig...),
		VerbValuesJSON:  s.VerbValuesJSON,
		DevicePath:      s.DevicePath,
		Duration:        duration,
		QueueSize:       s.QueueSize,
		SendInterval:    sendInterval,
		FlushInterval:   flushInterval,
		LogEvents:       s.LogEvents,
		LogRenderStats:  s.LogRenderStats,
		LogWriterStats:  s.LogWriterStats,
		LogJSStats:      s.LogJSStats,
		LogJSTrace:      s.LogJSTrace,
		LogGoTrace:      s.LogGoTrace,
		TraceLimit:      s.TraceLimit,
		TraceDumpOnExit: s.TraceDumpOnExit,
		StatsInterval:   statsInterval,
		ExitOnCircle:    s.ExitOnCircle,
	}, nil
}

func parseDurationFlag(name, value string) (time.Duration, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse --%s: %w", name, err)
	}
	return d, nil
}

type runtimeBootstrap func(context.Context, *jsruntime.Runtime) error

func run(ctx context.Context, opts options) error {
	if strings.TrimSpace(opts.Verb) != "" {
		return runVerbScene(ctx, opts)
	}
	return runRawScriptScene(ctx, opts)
}

func prepareRawScriptBootstrap(scriptPath string) ([]engine.Option, runtimeBootstrap, error) {
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read script: %w", err)
	}
	target, err := scriptmeta.ResolveTarget(scriptPath)
	if err != nil {
		return nil, nil, err
	}
	runtimeOptions, err := scriptmeta.EngineOptionsForTarget(target, nil)
	if err != nil {
		return nil, nil, err
	}
	bootstrap := func(runCtx context.Context, rt *jsruntime.Runtime) error {
		_, err := rt.RunString(runCtx, string(script))
		if err != nil {
			return fmt.Errorf("run script: %w", err)
		}
		return nil
	}
	return runtimeOptions, bootstrap, nil
}

func runRawScriptScene(ctx context.Context, opts options) error {
	runtimeOptions, bootstrap, err := prepareRawScriptBootstrap(opts.ScriptPath)
	if err != nil {
		return err
	}
	return runSceneSession(ctx, opts, runtimeOptions, bootstrap)
}

func prepareVerbBootstrap(opts options) ([]engine.Option, runtimeBootstrap, error) {
	target, registry, err := scriptmeta.ScanVerbRegistry(opts.ScriptPath)
	if err != nil {
		return nil, nil, fmt.Errorf("scan verbs: %w", err)
	}
	verb, err := scriptmeta.FindVerb(target, registry, opts.Verb)
	if err != nil {
		return nil, nil, err
	}
	desc, err := registry.CommandDescriptionForVerb(verb)
	if err != nil {
		return nil, nil, err
	}
	parsedValues, err := scriptmeta.ParseVerbValues(desc, opts.VerbConfig, opts.VerbValuesJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("parse verb values: %w", err)
	}
	runtimeOptions, err := scriptmeta.EngineOptionsForTarget(target, registry)
	if err != nil {
		return nil, nil, err
	}
	bootstrap := func(runCtx context.Context, rt *jsruntime.Runtime) error {
		_, err := registry.InvokeInRuntime(runCtx, rt.Runtime, verb, parsedValues)
		if err != nil {
			return fmt.Errorf("invoke verb %s: %w", verb.FullPath(), err)
		}
		return nil
	}
	return runtimeOptions, bootstrap, nil
}

func runVerbScene(ctx context.Context, opts options) error {
	runtimeOptions, bootstrap, err := prepareVerbBootstrap(opts)
	if err != nil {
		return err
	}
	return runSceneSession(ctx, opts, runtimeOptions, bootstrap)
}

func runSceneSession(ctx context.Context, opts options, runtimeOptions []engine.Option, bootstrap runtimeBootstrap) error {
	writerOptions := device.WriterOptions{QueueSize: opts.QueueSize, SendInterval: opts.SendInterval}
	renderOptions := device.DefaultRenderOptions
	renderOptions.FlushInterval = opts.FlushInterval

	var deckConn *device.Loupedeck
	var err error
	if opts.DevicePath == "" {
		deckConn, err = device.ConnectAutoWithWriterAndRenderOptions(writerOptions, &renderOptions)
	} else {
		deckConn, err = device.ConnectPathWithWriterAndRenderOptions(opts.DevicePath, writerOptions, &renderOptions)
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

	env := envpkg.Ensure(&envpkg.LoupeDeckEnvironment{Metrics: metrics.NewWithTraceLimit(opts.TraceLimit)})
	env.Host.Attach(deckConn)
	if opts.LogEvents {
		registerEventLogging(env)
	}

	rt, err := jsruntime.OpenRuntime(context.Background(), env, runtimeOptions...)
	if err != nil {
		return err
	}
	defer func() { _ = rt.Close(context.Background()) }()
	if err := bootstrap(rt.Context(), rt); err != nil {
		return err
	}

	renderer := render.NewWithDisplays(rt.Env.UI, map[string]render.DrawTarget{
		"left":  displays["left"],
		"main":  displays["main"],
		"right": displays["right"],
	})
	renderer.Theme = render.Theme{Background: color.Black, Foreground: color.White, Accent: color.White}

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
			slog.Info("js stats", "script", opts.ScriptPath, "verb", opts.Verb, "label", label, "counters", formatJSCounters(snap), "timings", formatJSTimings(snap))
		}
		if opts.LogJSTrace {
			for _, event := range filterTraceEvents(snap.Trace, false) {
				slog.Info("js trace", "script", opts.ScriptPath, "verb", opts.Verb, "label", label, "seq", event.Seq, "event", event.Name, "fields", formatTraceFields(event.Fields))
			}
		}
		if opts.LogGoTrace {
			for _, event := range filterTraceEvents(snap.Trace, true) {
				slog.Info("go trace", "script", opts.ScriptPath, "verb", opts.Verb, "label", label, "seq", event.Seq, "event", event.Name, "fields", formatTraceFields(event.Fields))
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

	slog.Info("Loupedeck JS live runner started", "script", opts.ScriptPath, "verb", opts.Verb, "duration", opts.Duration, "send_interval", opts.SendInterval, "flush_interval", opts.FlushInterval, "log_render_stats", opts.LogRenderStats, "log_writer_stats", opts.LogWriterStats, "log_js_stats", opts.LogJSStats, "log_js_trace", opts.LogJSTrace, "log_go_trace", opts.LogGoTrace, "trace_limit", opts.TraceLimit)
	exitRunner := func(reason string, attrs ...any) error {
		logAttrs := []any{"reason", reason, "script", opts.ScriptPath, "verb", opts.Verb}
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
		case <-ctx.Done():
			_ = exitRunner("context-cancelled")
			return ctx.Err()
		case <-statsTick:
			if opts.LogRenderStats {
				renderWindowMu.Lock()
				slog.Info("render stats", "script", opts.ScriptPath, "verb", opts.Verb, "stats", renderWindow.String())
				renderWindow = renderStatsWindow{}
				renderWindowMu.Unlock()
			}
			if opts.LogWriterStats {
				current := deckConn.WriterStats()
				delta := diffWriterStats(lastWriterStats, current)
				slog.Info("writer stats", "script", opts.ScriptPath, "verb", opts.Verb, "delta", delta, "current", current)
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

func registerEventLogging(env *envpkg.LoupeDeckEnvironment) {
	if env == nil {
		return
	}
	for _, button := range []device.Button{device.Circle, device.Button1, device.Button2, device.Button3, device.Button4, device.Button5, device.Button6, device.Button7} {
		button := button
		env.Host.OnButton(button, func(b device.Button, s device.ButtonStatus) {
			slog.Info("button event", "button", b.String(), "status", s.String())
		})
	}
	for _, touch := range []device.TouchButton{device.Touch1, device.Touch2, device.Touch3, device.Touch4, device.Touch5, device.Touch6, device.Touch7, device.Touch8, device.Touch9, device.Touch10, device.Touch11, device.Touch12} {
		touch := touch
		env.Host.OnTouch(touch, func(t device.TouchButton, s device.ButtonStatus, x, y uint16) {
			slog.Info("touch event", "touch", t.String(), "status", s.String(), "x", x, "y", y)
		})
	}
	for _, knob := range []device.Knob{device.Knob1, device.Knob2, device.Knob3, device.Knob4, device.Knob5, device.Knob6} {
		knob := knob
		env.Host.OnKnob(knob, func(k device.Knob, value int) {
			slog.Info("knob event", "knob", k.String(), "value", value)
		})
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
		w.FlushTicks, w.NonEmptyFlushes, w.FlushedDisplays, w.FlushedTiles, w.FlushedOps, avgMs, float64(w.MaxRender)/float64(time.Millisecond))
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
