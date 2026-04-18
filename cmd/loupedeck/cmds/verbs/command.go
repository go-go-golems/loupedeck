package verbs

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	runcmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/run"
	"github.com/spf13/cobra"
)

type invokerFactory func(repo scannedRepository, verb *jsverbs.VerbSpec, verbDescription *cmds.CommandDescription) jsverbs.VerbInvoker

type glazeCommandWrapper struct {
	base cmds.GlazeCommand
	desc *cmds.CommandDescription
}

func (c *glazeCommandWrapper) Description() *cmds.CommandDescription {
	return c.desc
}

func (c *glazeCommandWrapper) ToYAML(w io.Writer) error {
	return c.desc.ToYAML(w)
}

func (c *glazeCommandWrapper) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	return c.base.RunIntoGlazeProcessor(ctx, parsedValues, gp)
}

type writerCommandWrapper struct {
	base cmds.WriterCommand
	desc *cmds.CommandDescription
}

func (c *writerCommandWrapper) Description() *cmds.CommandDescription {
	return c.desc
}

func (c *writerCommandWrapper) ToYAML(w io.Writer) error {
	return c.desc.ToYAML(w)
}

func (c *writerCommandWrapper) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	return c.base.RunIntoWriter(ctx, parsedValues, w)
}

func NewCommand(bootstrap Bootstrap) (*cobra.Command, error) {
	return newCommandWithInvokerFactory(bootstrap, liveSceneInvokerFactory)
}

func newCommandWithInvokerFactory(bootstrap Bootstrap, invokers invokerFactory) (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "verbs",
		Short: "Run annotated loupedeck scene verbs",
	}

	repositories, err := scanRepositories(bootstrap)
	if err != nil {
		return nil, err
	}
	discovered, err := collectDiscoveredVerbs(repositories)
	if err != nil {
		return nil, err
	}
	commands, err := buildCommands(discovered, invokers)
	if err != nil {
		return nil, err
	}
	if err := glazedcli.AddCommandsToRootCommand(root, commands, nil,
		glazedcli.WithDualMode(true),
		glazedcli.WithGlazeToggleFlag("with-glaze-output"),
		glazedcli.WithParserConfig(glazedcli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   glazedcli.CobraCommandDefaultMiddlewares,
		}),
	); err != nil {
		return nil, err
	}
	return root, nil
}

func buildCommands(discovered []discoveredVerb, invokers invokerFactory) ([]cmds.Command, error) {
	commands := make([]cmds.Command, 0, len(discovered))
	for _, discoveredVerb := range discovered {
		verbDescription, err := discoveredVerb.Repository.Registry.CommandDescriptionForVerb(discoveredVerb.Verb)
		if err != nil {
			return nil, err
		}
		upstreamCommand, err := discoveredVerb.Repository.Registry.CommandForVerbWithInvoker(discoveredVerb.Verb, invokers(discoveredVerb.Repository, discoveredVerb.Verb, verbDescription))
		if err != nil {
			return nil, err
		}
		augmentedDescription, err := augmentDescription(upstreamCommand.Description())
		if err != nil {
			return nil, err
		}
		wrapped, err := wrapCommandWithDescription(upstreamCommand, augmentedDescription)
		if err != nil {
			return nil, err
		}
		commands = append(commands, wrapped)
	}
	return commands, nil
}

func augmentDescription(description *cmds.CommandDescription) (*cmds.CommandDescription, error) {
	if description == nil {
		return nil, fmt.Errorf("command description is nil")
	}
	ret := description.Clone(true)
	commonSections, err := runcmd.CommonSections()
	if err != nil {
		return nil, err
	}
	ret.SetSections(commonSections...)
	return ret, nil
}

func wrapCommandWithDescription(command cmds.Command, description *cmds.CommandDescription) (cmds.Command, error) {
	switch c := command.(type) {
	case cmds.GlazeCommand:
		return &glazeCommandWrapper{base: c, desc: description}, nil
	case cmds.WriterCommand:
		return &writerCommandWrapper{base: c, desc: description}, nil
	default:
		return nil, fmt.Errorf("unsupported command type %T", command)
	}
}

func liveSceneInvokerFactory(repo scannedRepository, verb *jsverbs.VerbSpec, verbDescription *cmds.CommandDescription) jsverbs.VerbInvoker {
	identity := runcmd.SceneIdentity{ScriptPath: verbSourceLabel(repo, verb), Verb: verb.FullPath()}
	runtimeOptions := repo.runtimeOptions()
	return func(ctx context.Context, _ *jsverbs.Registry, _ *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
		sessionOptions, err := runcmd.DecodeSessionOptions(parsedValues)
		if err != nil {
			return nil, err
		}
		verbValues := subsetValuesForDescription(parsedValues, verbDescription)
		return runcmd.RunAnnotatedVerbScene(ctx, identity, sessionOptions, runtimeOptions, repo.Registry, verb, verbValues)
	}
}

func subsetValuesForDescription(parsedValues *values.Values, description *cmds.CommandDescription) *values.Values {
	if parsedValues == nil || description == nil || description.Schema == nil {
		return values.New()
	}
	ret := values.New()
	description.Schema.ForEach(func(slug string, _ schema.Section) {
		if sectionValues, ok := parsedValues.Get(slug); ok {
			ret.Set(slug, sectionValues.Clone())
		}
	})
	return ret
}

func verbSourceLabel(repo scannedRepository, verb *jsverbs.VerbSpec) string {
	if verb == nil || verb.File == nil {
		return repo.Repository.Name
	}
	if verb.File.AbsPath != "" {
		return verb.File.AbsPath
	}
	return fmt.Sprintf("%s:%s", repo.Repository.Name, filepath.ToSlash(verb.File.RelPath))
}
