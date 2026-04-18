package verbs

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
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
	if !strings.Contains(help, "0s") {
		t.Fatalf("expected 0s default in help output, got %q", help)
	}
	if strings.Contains(help, "with-glaze-output") || strings.Contains(help, "print-schema") {
		t.Fatalf("expected no structured-output toggles in help, got %q", help)
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
	captured := map[string]interface{}{}
	commands, err := buildCommands(discovered, func(repo scannedRepository, verb *jsverbs.VerbSpec, _ *cmds.CommandDescription) jsverbs.VerbInvoker {
		return func(ctx context.Context, _ *jsverbs.Registry, _ *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
			defaultValues, _ := parsedValues.Get("default")
			displayValues, _ := parsedValues.Get("display")
			sessionValues, _ := parsedValues.Get("loupedeck")
			title, _ := defaultValues.GetField("title")
			theme, _ := displayValues.GetField("theme")
			device, _ := sessionValues.GetField("device")
			duration, _ := sessionValues.GetField("duration")
			captured["repository"] = repo.Repository.Name
			captured["verb"] = verb.FullPath()
			captured["title"] = title
			captured["theme"] = theme
			captured["device"] = device
			captured["duration"] = duration
			return nil, nil
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
			"device":   "/dev/mock",
			"duration": "0s",
		},
	}))
	if err != nil {
		t.Fatalf("parse command values: %v", err)
	}
	bareCommand, ok := target.(cmds.BareCommand)
	if !ok {
		t.Fatalf("expected bare command, got %T", target)
	}
	if err := bareCommand.Run(context.Background(), parsedValues); err != nil {
		t.Fatalf("run command: %v", err)
	}
	if captured["verb"] != "documented configure" || captured["title"] != "OPS" || captured["theme"] != "light" || captured["device"] != "/dev/mock" || captured["duration"] != "0s" {
		t.Fatalf("unexpected capture %#v", captured)
	}
}
