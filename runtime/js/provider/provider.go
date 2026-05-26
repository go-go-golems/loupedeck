package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerutil"
	runcmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/run"
	verbscmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/verbs"
	"github.com/go-go-golems/loupedeck/pkg/device"
	"github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/js/module_anim"
	"github.com/go-go-golems/loupedeck/runtime/js/module_easing"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
	"github.com/go-go-golems/loupedeck/runtime/js/module_present"
	"github.com/go-go-golems/loupedeck/runtime/js/module_state"
	"github.com/go-go-golems/loupedeck/runtime/js/module_ui"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
	"github.com/go-go-golems/loupedeck/runtime/render"
)

const PackageID = "loupedeck"

type scenesCommandProviderConfig struct {
	IncludeRun   *bool    `json:"includeRun,omitempty"`
	Repositories []string `json:"repositories,omitempty"`
}

func Register(registry *providerapi.Registry) error {
	hardware := newHardwareCapability()
	return registry.Package(PackageID,
		moduleEntry(module_anim.ModuleName, "Animation loop and timeline helpers for Loupedeck scenes.", module_anim.Loader),
		moduleEntry(module_easing.ModuleName, "Easing functions for loupedeck animation curves.", module_easing.Loader),
		moduleEntry(module_gfx.ModuleName, "Offscreen drawing surfaces, colors, text, and font helpers.", module_gfx.Loader),
		moduleEntry(module_present.ModuleName, "Presentation helpers for Loupedeck visual scene runtimes.", module_present.Loader),
		moduleEntry(module_state.ModuleName, "Reactive state primitives for Loupedeck scenes.", module_state.Loader),
		moduleEntry(module_ui.ModuleName, "Retained Loupedeck UI pages, tiles, displays, and hardware events.", module_ui.Loader),
		providerapi.WithPackageCapability(hardware),
		providerapi.CommandSetProvider{
			Name:         "scenes",
			DefaultMount: "loupedeck",
			Description:  "Run Loupedeck JavaScript scenes and annotated scene verbs",
			New:          newScenesCommandSet,
		},
	)
}

func newScenesCommandSet(ctx providerapi.CommandSetContext) (*providerapi.CommandSet, error) {
	cfg := scenesCommandProviderConfig{}
	if len(ctx.Config) > 0 {
		if err := json.Unmarshal(ctx.Config, &cfg); err != nil {
			return nil, fmt.Errorf("decode loupedeck scenes command provider config: %w", err)
		}
	}
	sections, err := providerutil.CollectConfigSections(ctx.SelectedModules, providerapi.SectionContext{
		CommandProviderID: ctx.Name,
		RuntimeProfile:    ctx.RuntimeProfile,
	}, map[string]string{schema.DefaultSlug: "loupedeck scene command schema"})
	if err != nil {
		return nil, err
	}

	commands := []cmds.Command{}
	if cfg.IncludeRun == nil || *cfg.IncludeRun {
		runCommand, err := runcmd.NewCommand()
		if err != nil {
			return nil, fmt.Errorf("build loupedeck run command: %w", err)
		}
		commands = append(commands, runCommand)
	}
	verbArgs := make([]string, 0, len(cfg.Repositories)*2)
	for _, repo := range cfg.Repositories {
		verbArgs = append(verbArgs, "--"+verbscmd.VerbRepositoryFlag, repo)
	}
	bootstrap, err := verbscmd.DiscoverBootstrap(verbArgs)
	if err != nil {
		return nil, fmt.Errorf("discover loupedeck verb repositories: %w", err)
	}
	var verbCommands []cmds.Command
	if ctx.RuntimeFactory != nil {
		verbCommands, err = verbscmd.NewCommandsWithInvokerFactory(bootstrap, xgojaSceneInvokerFactory(ctx))
	} else {
		verbCommands, err = verbscmd.NewCommands(bootstrap)
	}
	if err != nil {
		return nil, fmt.Errorf("build loupedeck verb commands: %w", err)
	}
	commands = append(commands, verbCommands...)
	appendSections(commands, sections)
	return &providerapi.CommandSet{
		Commands: commands,
		ParserConfig: &cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
		},
	}, nil
}

func xgojaSceneInvokerFactory(providerCtx providerapi.CommandSetContext) verbscmd.InvokerFactory {
	return func(repo verbscmd.ScannedRepository, _ *jsverbs.VerbSpec, _ *cmds.CommandDescription) jsverbs.VerbInvoker {
		return func(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
			if providerCtx.RuntimeFactory == nil {
				return nil, fmt.Errorf("xgoja runtime factory is nil")
			}
			profile := strings.TrimSpace(providerCtx.RuntimeProfile)
			if profile == "" {
				return nil, fmt.Errorf("xgoja runtime profile is empty")
			}
			if registry == nil {
				registry = repo.Registry
			}
			if registry == nil {
				return nil, fmt.Errorf("jsverbs registry is nil")
			}
			opts := []require.Option{require.WithLoader(registry.RequireLoader())}
			if !repo.Repository.Embedded && strings.TrimSpace(repo.Repository.RootDir) != "" {
				folders := []string{repo.Repository.RootDir, filepath.Join(repo.Repository.RootDir, "node_modules")}
				parent := filepath.Dir(repo.Repository.RootDir)
				if parent != repo.Repository.RootDir {
					folders = append(folders, parent, filepath.Join(parent, "node_modules"))
				}
				opts = append(opts, require.WithGlobalFolders(folders...))
			}
			rt, err := providerCtx.RuntimeFactory.NewRuntime(ctx, profile, opts...)
			if err != nil {
				return nil, err
			}
			defer func() { _ = rt.Close(context.Background()) }()
			if err := providerutil.InitRuntimeFromSections(ctx, parsedValues, runtimeHandle{rt: rt}, providerCtx.SelectedModules); err != nil {
				return nil, err
			}
			return registry.InvokeInRuntime(ctx, rt, verb, parsedValues)
		}
	}
}

func appendSections(commands []cmds.Command, sections []schema.Section) {
	if len(sections) == 0 {
		return
	}
	for _, command := range commands {
		if command == nil || command.Description() == nil {
			continue
		}
		for _, section := range sections {
			command.Description().SetSections(section)
		}
	}
}

type runtimeHandle struct {
	rt *engine.Runtime
}

func (h runtimeHandle) Runtime() *goja.Runtime {
	if h.rt == nil {
		return nil
	}
	return h.rt.VM
}

func (h runtimeHandle) Close(ctx context.Context) error {
	if h.rt == nil {
		return nil
	}
	return h.rt.Close(ctx)
}

func (h runtimeHandle) AddCloser(fn func(context.Context) error) error {
	if h.rt == nil {
		return fmt.Errorf("runtime is nil")
	}
	return h.rt.AddCloser(fn)
}

func moduleEntry(name, description string, loader func() require.ModuleLoader) providerapi.Module {
	return providerapi.Module{
		Name:        name,
		DefaultAs:   name,
		Description: description,
		New: func(providerapi.ModuleContext) (require.ModuleLoader, error) {
			return loader(), nil
		},
	}
}

type hardwareCapability struct{}

type hardwareSettings struct {
	Enabled       bool   `glazed:"enabled"`
	DevicePath    string `glazed:"device"`
	QueueSize     int    `glazed:"queue-size"`
	SendInterval  string `glazed:"send-interval"`
	FlushInterval string `glazed:"flush-interval"`
}

func newHardwareCapability() *hardwareCapability { return &hardwareCapability{} }

func (c *hardwareCapability) CapabilityID() string { return "loupedeck.hardware" }

func (c *hardwareCapability) ConfigSections(providerapi.SectionContext) ([]schema.Section, error) {
	section, err := schema.NewSection(
		"loupedeck-hardware",
		"Loupedeck hardware",
		schema.WithDescription("Real hardware/session settings for xgoja Loupedeck UI modules"),
		schema.WithPrefix("deck-"),
		schema.WithFields(
			fields.New("enabled", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Connect to real Loupedeck hardware and render UI pages")),
			fields.New("device", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional serial device path (default: auto-detect)")),
			fields.New("queue-size", fields.TypeInteger, fields.WithDefault(256), fields.WithHelp("Writer queue size")),
			fields.New("send-interval", fields.TypeString, fields.WithDefault("35ms"), fields.WithHelp("Writer pacing interval")),
			fields.New("flush-interval", fields.TypeString, fields.WithDefault(device.DefaultRenderOptions.FlushInterval.String()), fields.WithHelp("Retained render scheduler flush interval")),
		),
	)
	if err != nil {
		return nil, err
	}
	return []schema.Section{section}, nil
}

func (c *hardwareCapability) InitRuntimeFromSections(ctx context.Context, vals *values.Values, handle providerapi.RuntimeHandle) error {
	if handle == nil || handle.Runtime() == nil {
		return fmt.Errorf("loupedeck hardware runtime handle is nil")
	}
	settings := hardwareSettings{Enabled: true, QueueSize: 256, SendInterval: "35ms", FlushInterval: device.DefaultRenderOptions.FlushInterval.String()}
	if vals != nil {
		if err := vals.DecodeSectionInto("loupedeck-hardware", &settings); err != nil {
			return err
		}
	}

	environment := env.Ensure(&env.LoupeDeckEnvironment{Metrics: metrics.New()})
	env.Store(handle.Runtime(), environment)

	closers := []func(context.Context) error{
		func(context.Context) error {
			env.Delete(handle.Runtime())
			return nil
		},
	}

	if settings.Enabled {
		deckConn, displays, err := connectHardware(settings)
		if err != nil {
			for _, closer := range closers {
				_ = closer(ctx)
			}
			return err
		}
		environment.Host.Attach(deckConn)
		listenErrCh := make(chan error, 1)
		go func() { listenErrCh <- deckConn.Listen() }()
		go func() {
			if err := <-listenErrCh; err != nil {
				fmt.Printf("loupedeck listen failed: %v\n", err)
			}
		}()

		renderer := render.NewWithDisplays(environment.UI, map[string]render.DrawTarget{
			"left":  displays["left"],
			"main":  displays["main"],
			"right": displays["right"],
		})
		renderer.Theme = render.Theme{Background: color.Black, Foreground: color.White, Accent: color.White}
		environment.Present.SetFlushFunc(func() (int, error) {
			return renderer.Flush(), nil
		})
		environment.Present.Start(ctx)

		closers = append([]func(context.Context) error{
			func(context.Context) error {
				environment.Present.Close()
				clearDisplays(displays)
				return deckConn.Close()
			},
		}, closers...)
	}

	if closerRegistry, ok := handle.(providerapi.RuntimeCloserRegistry); ok {
		return closerRegistry.AddCloser(func(ctx context.Context) error {
			var ret error
			for _, closer := range closers {
				if err := closer(ctx); err != nil && ret == nil {
					ret = err
				}
			}
			return ret
		})
	}
	return nil
}

func connectHardware(settings hardwareSettings) (*device.Loupedeck, map[string]*device.Display, error) {
	sendInterval, err := time.ParseDuration(settings.SendInterval)
	if err != nil {
		return nil, nil, fmt.Errorf("parse --deck-send-interval: %w", err)
	}
	flushInterval, err := time.ParseDuration(settings.FlushInterval)
	if err != nil {
		return nil, nil, fmt.Errorf("parse --deck-flush-interval: %w", err)
	}
	if flushInterval <= 0 {
		return nil, nil, fmt.Errorf("--deck-flush-interval must be > 0, got %s", flushInterval)
	}
	writerOptions := device.WriterOptions{QueueSize: settings.QueueSize, SendInterval: sendInterval}
	renderOptions := device.DefaultRenderOptions
	renderOptions.FlushInterval = flushInterval
	var deckConn *device.Loupedeck
	if strings.TrimSpace(settings.DevicePath) == "" {
		var err error
		deckConn, err = device.ConnectAutoWithWriterAndRenderOptions(writerOptions, &renderOptions)
		if err != nil {
			return nil, nil, fmt.Errorf("connect loupedeck: %w", err)
		}
	} else {
		var err error
		deckConn, err = device.ConnectPathWithWriterAndRenderOptions(settings.DevicePath, writerOptions, &renderOptions)
		if err != nil {
			return nil, nil, fmt.Errorf("connect loupedeck %s: %w", settings.DevicePath, err)
		}
	}
	displays := map[string]*device.Display{
		"left":  deckConn.GetDisplay("left"),
		"main":  deckConn.GetDisplay("main"),
		"right": deckConn.GetDisplay("right"),
	}
	if displays["main"] == nil {
		_ = deckConn.Close()
		return nil, nil, fmt.Errorf("missing main display")
	}
	return deckConn, displays, nil
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

var _ providerapi.RuntimeHandle = runtimeHandle{}
var _ providerapi.ConfigSectionCapability = (*hardwareCapability)(nil)
var _ providerapi.RuntimeInitializerCapability = (*hardwareCapability)(nil)
