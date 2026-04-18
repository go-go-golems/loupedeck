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
