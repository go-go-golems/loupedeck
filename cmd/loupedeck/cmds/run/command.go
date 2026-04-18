package run

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/pkg/scriptmeta"
	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
)

type Command struct {
	*cmds.CommandDescription
}

type settings_ struct {
	ScriptPath string `glazed:"script"`
}

type options struct {
	ScriptPath string
	Session    SessionOptions
}

type commandResult struct {
	ScriptPath     string
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
	commonSections, err := CommonSections()
	if err != nil {
		return nil, err
	}
	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Arguments",
		schema.WithFields(
			fields.New("script", fields.TypeString, fields.WithIsArgument(true), fields.WithHelp("Path to the JavaScript file to execute")),
		),
	)
	if err != nil {
		return nil, err
	}

	sections := append([]schema.Section{defaultSection}, commonSections...)
	desc := cmds.NewCommandDescription(
		"run",
		cmds.WithShort("Run a plain Loupedeck Live JavaScript file on hardware"),
		cmds.WithLong(`Execute a plain JavaScript scene file against a real Loupedeck device.

Examples:
  loupedeck run ./examples/js/01-hello.js --duration 5s
  loupedeck run ./examples/js/11-cyb-os-tiles.js --send-interval 0ms --flush-interval 20ms
  loupedeck run ./examples/js/01-hello.js --with-glaze-output --output json`),
		cmds.WithSections(sections...),
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
	opts, err := decodeOptions(vals)
	if err != nil {
		return nil, err
	}
	if err := run(ctx, opts); err != nil {
		return nil, err
	}
	devicePath := opts.Session.DevicePath
	if devicePath == "" {
		devicePath = "auto"
	}
	return &commandResult{
		ScriptPath:     opts.ScriptPath,
		DevicePath:     devicePath,
		Duration:       opts.Session.Duration,
		SendInterval:   opts.Session.SendInterval,
		FlushInterval:  opts.Session.FlushInterval,
		QueueSize:      opts.Session.QueueSize,
		ExitOnCircle:   opts.Session.ExitOnCircle,
		TraceLimit:     opts.Session.TraceLimit,
		RequestedStats: opts.Session.LogRenderStats || opts.Session.LogWriterStats || opts.Session.LogJSStats || opts.Session.LogJSTrace || opts.Session.LogGoTrace,
		Status:         "ok",
	}, nil
}

func decodeOptions(vals *values.Values) (options, error) {
	s := &settings_{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return options{}, err
	}
	if s.ScriptPath == "" {
		return options{}, fmt.Errorf("missing script path")
	}
	session, err := DecodeSessionOptions(vals)
	if err != nil {
		return options{}, err
	}
	return options{ScriptPath: s.ScriptPath, Session: session}, nil
}

func parseDurationFlag(name, value string) (time.Duration, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse --%s: %w", name, err)
	}
	return d, nil
}

func run(ctx context.Context, opts options) error {
	return runRawScriptScene(ctx, opts)
}

func prepareRawScriptBootstrap(scriptPath string) ([]engine.Option, RuntimeBootstrap, error) {
	target, err := scriptmeta.ResolveTarget(scriptPath)
	if err != nil {
		return nil, nil, err
	}
	if target.EntryFile == "" {
		return nil, nil, fmt.Errorf("raw script execution requires a JavaScript file, got directory %s", target.Path)
	}
	script, err := os.ReadFile(target.EntryFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read script: %w", err)
	}
	runtimeOptions, err := scriptmeta.EngineOptionsForTarget(target, nil)
	if err != nil {
		return nil, nil, err
	}
	bootstrap := func(runCtx context.Context, rt *jsruntime.Runtime) (any, error) {
		_, err := rt.RunString(runCtx, string(script))
		if err != nil {
			return nil, fmt.Errorf("run script: %w", err)
		}
		return nil, nil
	}
	return runtimeOptions, bootstrap, nil
}

func runRawScriptScene(ctx context.Context, opts options) error {
	runtimeOptions, bootstrap, err := prepareRawScriptBootstrap(opts.ScriptPath)
	if err != nil {
		return err
	}
	_, err = RunSceneSession(ctx, SceneIdentity{ScriptPath: opts.ScriptPath}, opts.Session, runtimeOptions, bootstrap)
	return err
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
