package run

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	commoncmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/common"
	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
)

func TestPrepareRawScriptBootstrapRunsPlainScript(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "..", "..", "examples", "js", "01-hello.js")
	runtimeOptions, bootstrap, err := prepareRawScriptBootstrap(scriptPath)
	if err != nil {
		t.Fatalf("prepare raw bootstrap: %v", err)
	}
	rt, err := jsruntime.OpenRuntime(context.Background(), nil, runtimeOptions...)
	if err != nil {
		t.Fatalf("open runtime: %v", err)
	}
	defer func() { _ = rt.Close(context.Background()) }()
	if _, err := bootstrap(context.Background(), rt); err != nil {
		t.Fatalf("run raw bootstrap: %v", err)
	}
	page := rt.Env.UI.Page("hello")
	if page == nil {
		t.Fatal("expected hello page")
	}
	if tile := page.Tile(0, 0); tile == nil || tile.Text() != "HELLO" {
		t.Fatalf("expected HELLO tile, got %#v", tile)
	}
}

func TestNewCommandHelpShowsRuntimeFlagsWithoutStructuredOutputToggles(t *testing.T) {
	command, err := NewCommand()
	if err != nil {
		t.Fatalf("new command: %v", err)
	}
	cobraCommand, err := commoncmd.BuildRuntimeCobraCommand(command)
	if err != nil {
		t.Fatalf("build cobra command: %v", err)
	}
	var out bytes.Buffer
	cobraCommand.SetOut(&out)
	cobraCommand.SetErr(&out)
	cobraCommand.SetArgs([]string{"--help"})
	if err := cobraCommand.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}
	help := out.String()
	if !strings.Contains(help, "--duration string") {
		t.Fatalf("expected duration flag in help, got %q", help)
	}
	if !strings.Contains(help, "0s") {
		t.Fatalf("expected 0s default in help, got %q", help)
	}
	if strings.Contains(help, "with-glaze-output") || strings.Contains(help, "print-schema") {
		t.Fatalf("expected no structured-output flags in help, got %q", help)
	}
}

func TestPrepareRawScriptBootstrapResolvesShorthandPath(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "..", "..", "examples", "js", "02")
	runtimeOptions, bootstrap, err := prepareRawScriptBootstrap(scriptPath)
	if err != nil {
		t.Fatalf("prepare raw bootstrap with shorthand: %v", err)
	}
	rt, err := jsruntime.OpenRuntime(context.Background(), nil, runtimeOptions...)
	if err != nil {
		t.Fatalf("open runtime: %v", err)
	}
	defer func() { _ = rt.Close(context.Background()) }()
	if _, err := bootstrap(context.Background(), rt); err != nil {
		t.Fatalf("run raw bootstrap with shorthand: %v", err)
	}
	page := rt.Env.UI.Page("counter")
	if page == nil {
		t.Fatal("expected counter page")
	}
}
