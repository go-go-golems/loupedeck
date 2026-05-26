package provider

import (
	"encoding/json"
	"fmt"

	"github.com/dop251/goja_nodejs/require"
	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
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
	verbCommands, err := verbscmd.NewCommands(bootstrap)
	if err != nil {
		return nil, fmt.Errorf("build loupedeck verb commands: %w", err)
	}
	commands = append(commands, verbCommands...)
	return &providerapi.CommandSet{
		Commands: commands,
		ParserConfig: &glazedcli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
		},
	}, nil
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
