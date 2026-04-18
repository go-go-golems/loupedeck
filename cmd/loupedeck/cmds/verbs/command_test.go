package verbs

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

func mustBootstrap(t *testing.T) Bootstrap {
	t.Helper()
	bootstrap, err := DiscoverBootstrap(nil)
	if err != nil {
		t.Fatalf("discover bootstrap: %v", err)
	}
	return bootstrap
}

func TestNewCommandShowsEmbeddedVerbHelp(t *testing.T) {
	cmd, err := NewCommand(mustBootstrap(t))
	if err != nil {
		t.Fatalf("new command: %v", err)
	}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"documented", "configure", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}
	help := out.String()
	if !strings.Contains(help, "--theme") {
		t.Fatalf("expected --theme in help output, got %q", help)
	}
	if !strings.Contains(help, "--device") {
		t.Fatalf("expected --device in help output, got %q", help)
	}
	if strings.Contains(help, "verbs list") || strings.Contains(help, "verbs help") {
		t.Fatalf("expected old inspection subcommands to be gone, got %q", help)
	}
}

func TestNewCommandInvokesDynamicVerbThroughCustomInvoker(t *testing.T) {
	repositories, err := scanRepositories(mustBootstrap(t))
	if err != nil {
		t.Fatalf("scan repositories: %v", err)
	}
	discovered, err := collectDiscoveredVerbs(repositories)
	if err != nil {
		t.Fatalf("collect discovered verbs: %v", err)
	}
	commands, err := buildCommands(discovered, func(repo scannedRepository, verb *jsverbs.VerbSpec, _ *cmds.CommandDescription) jsverbs.VerbInvoker {
		return func(ctx context.Context, _ *jsverbs.Registry, _ *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
			defaultValues, _ := parsedValues.Get("default")
			displayValues, _ := parsedValues.Get("display")
			sessionValues, _ := parsedValues.Get("loupedeck")
			title, _ := defaultValues.GetField("title")
			theme, _ := displayValues.GetField("theme")
			device, _ := sessionValues.GetField("device")
			return map[string]interface{}{
				"repository": repo.Repository.Name,
				"verb":       verb.FullPath(),
				"title":      title,
				"theme":      theme,
				"device":     device,
			}, nil
		}
	})
	if err != nil {
		t.Fatalf("build commands: %v", err)
	}
	var target cmds.Command
	for _, command := range commands {
		if command.Description().FullPath() == "documented/configure" {
			target = command
			break
		}
	}
	if target == nil {
		t.Fatal("expected documented configure command")
	}
	parsedValues, err := runner.ParseCommandValues(target, runner.WithValuesForSections(map[string]map[string]interface{}{
		"default": {
			"title": "OPS",
		},
		"display": {
			"theme": "light",
		},
		"loupedeck": {
			"device": "/dev/mock",
		},
	}))
	if err != nil {
		t.Fatalf("parse command values: %v", err)
	}
	gp := &captureProcessor{}
	glazeCommand, ok := target.(cmds.GlazeCommand)
	if !ok {
		t.Fatalf("expected glaze command, got %T", target)
	}
	if err := glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, gp); err != nil {
		t.Fatalf("run command: %v", err)
	}
	if len(gp.rows) != 1 {
		t.Fatalf("expected one row, got %#v", gp.rows)
	}
	row := rowToMap(gp.rows[0])
	if row["verb"] != "documented configure" || row["title"] != "OPS" || row["theme"] != "light" || row["device"] != "/dev/mock" {
		t.Fatalf("unexpected row %#v", row)
	}
}

type captureProcessor struct {
	rows []types.Row
}

func (c *captureProcessor) AddRow(_ context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}

func (c *captureProcessor) Close(context.Context) error {
	return nil
}

var _ middlewares.Processor = (*captureProcessor)(nil)

func rowToMap(row types.Row) map[string]interface{} {
	ret := map[string]interface{}{}
	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		ret[pair.Key] = pair.Value
	}
	return ret
}
