package run

import (
	"context"
	"path/filepath"
	"testing"

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
	if err := bootstrap(context.Background(), rt); err != nil {
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

func TestPrepareVerbBootstrapRunsAnnotatedVerb(t *testing.T) {
	opts := options{
		ScriptPath:     filepath.Join("..", "..", "..", "..", "examples", "js", "12-documented-scene.js"),
		Verb:           "documented configure",
		VerbValuesJSON: `{"default":{"title":"OPS"},"display":{"theme":"light","refreshRate":60}}`,
	}
	runtimeOptions, bootstrap, err := prepareVerbBootstrap(opts)
	if err != nil {
		t.Fatalf("prepare verb bootstrap: %v", err)
	}
	rt, err := jsruntime.OpenRuntime(context.Background(), nil, runtimeOptions...)
	if err != nil {
		t.Fatalf("open runtime: %v", err)
	}
	defer func() { _ = rt.Close(context.Background()) }()
	if err := bootstrap(context.Background(), rt); err != nil {
		t.Fatalf("run verb bootstrap: %v", err)
	}
	page := rt.Env.UI.Page("documented-home")
	if page == nil {
		t.Fatal("expected documented-home page")
	}
	if tile := page.Tile(0, 0); tile == nil || tile.Text() != "OPS" {
		t.Fatalf("expected OPS tile, got %#v", tile)
	}
}
