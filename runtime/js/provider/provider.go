package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerutil"
	runcmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/run"
	verbscmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/verbs"
	"github.com/go-go-golems/loupedeck/runtime/js/module_easing"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
)

const PackageID = "loupedeck"

type scenesCommandProviderConfig struct {
	IncludeRun   *bool    `json:"includeRun,omitempty"`
	Repositories []string `json:"repositories,omitempty"`
}

func Register(registry *providerapi.Registry) error {
	return registry.Package(PackageID,
		moduleEntry(module_easing.ModuleName, "Easing functions for loupedeck animation curves.", module_easing.Loader),
		moduleEntry(module_gfx.ModuleName, "Offscreen drawing surfaces, colors, text, and font helpers.", module_gfx.Loader),
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
		ParserConfig: &glazedcli.CobraParserConfig{
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

var _ providerapi.RuntimeHandle = runtimeHandle{}
