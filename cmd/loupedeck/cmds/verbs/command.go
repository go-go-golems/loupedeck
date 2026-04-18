package verbs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	runcmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/run"
	"github.com/spf13/cobra"
)

type invokerFactory func(repo scannedRepository, verb *jsverbs.VerbSpec, verbDescription *cmds.CommandDescription) jsverbs.VerbInvoker

type runtimeCommandWrapper struct {
	desc       *cmds.CommandDescription
	outputMode string
	execute    func(context.Context, *values.Values) (any, error)
}

func (c *runtimeCommandWrapper) Description() *cmds.CommandDescription {
	return c.desc
}

func (c *runtimeCommandWrapper) ToYAML(w io.Writer) error {
	return c.desc.ToYAML(w)
}

func (c *runtimeCommandWrapper) Run(ctx context.Context, parsedValues *values.Values) error {
	result, err := c.execute(ctx, parsedValues)
	if err != nil {
		return err
	}
	return printRuntimeCommandResult(os.Stdout, c.outputMode, result)
}

var _ cmds.BareCommand = (*runtimeCommandWrapper)(nil)

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
		glazedcli.WithParserConfig(glazedcli.CobraParserConfig{
			SkipCommandSettingsSection: true,
		}),
	); err != nil {
		return nil, err
	}
	return root, nil
}

func buildCommands(discovered []discoveredVerb, invokers invokerFactory) ([]cmds.Command, error) {
	commands := make([]cmds.Command, 0, len(discovered))
	for _, discoveredVerb := range discovered {
		repo := discoveredVerb.Repository
		verb := discoveredVerb.Verb
		verbDescription, err := repo.Registry.CommandDescriptionForVerb(verb)
		if err != nil {
			return nil, err
		}
		augmentedDescription, err := augmentDescription(verbDescription)
		if err != nil {
			return nil, err
		}
		invoker := invokers(repo, verb, verbDescription)
		commands = append(commands, &runtimeCommandWrapper{
			desc:       augmentedDescription,
			outputMode: verb.OutputMode,
			execute: func(ctx context.Context, parsedValues *values.Values) (any, error) {
				return invoker(ctx, repo.Registry, verb, parsedValues)
			},
		})
	}
	return commands, nil
}

func augmentDescription(description *cmds.CommandDescription) (*cmds.CommandDescription, error) {
	if description == nil {
		return nil, fmt.Errorf("command description is nil")
	}
	ret := description.Clone(true)
	runtimeSections, err := runcmd.RuntimeSections()
	if err != nil {
		return nil, err
	}
	ret.SetSections(runtimeSections...)
	return ret, nil
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

func printRuntimeCommandResult(w io.Writer, outputMode string, result any) error {
	if w == nil || result == nil {
		return nil
	}
	switch outputMode {
	case jsverbs.OutputModeText:
		switch v := result.(type) {
		case nil:
			return nil
		case string:
			_, err := io.WriteString(w, v)
			if err != nil {
				return err
			}
			if !strings.HasSuffix(v, "\n") {
				_, err = io.WriteString(w, "\n")
			}
			return err
		case []byte:
			if _, err := w.Write(v); err != nil {
				return err
			}
			if len(v) == 0 || v[len(v)-1] != '\n' {
				_, err := io.WriteString(w, "\n")
				return err
			}
			return nil
		default:
			_, err := fmt.Fprintf(w, "%v\n", result)
			return err
		}
	default:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
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
