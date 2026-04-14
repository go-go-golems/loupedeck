package js

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/loupedeck/pkg/scriptmeta"
)

func TestInvokeAnnotatedVerbIntoLiveRuntime(t *testing.T) {
	example := filepath.Join("..", "..", "examples", "js", "12-documented-scene.js")
	target, registry, err := scriptmeta.ScanVerbRegistry(example)
	if err != nil {
		t.Fatalf("scan registry: %v", err)
	}
	verb, err := scriptmeta.FindVerb(target, registry, "documented configure")
	if err != nil {
		t.Fatalf("find verb: %v", err)
	}
	desc, err := registry.CommandDescriptionForVerb(verb)
	if err != nil {
		t.Fatalf("command description: %v", err)
	}
	parsedValues, err := scriptmeta.ParseVerbValues(desc, nil, `{"default":{"title":"OPS"},"display":{"theme":"light","refreshRate":60}}`)
	if err != nil {
		t.Fatalf("parse values: %v", err)
	}
	runtimeOptions, err := scriptmeta.EngineOptionsForTarget(target, registry)
	if err != nil {
		t.Fatalf("runtime options: %v", err)
	}
	rt, err := OpenRuntime(context.Background(), nil, runtimeOptions...)
	if err != nil {
		t.Fatalf("open runtime: %v", err)
	}
	defer func() { _ = rt.Close(context.Background()) }()

	result, err := registry.InvokeInRuntime(context.Background(), rt.Runtime, verb, parsedValues)
	if err != nil {
		t.Fatalf("invoke in runtime: %v", err)
	}
	row, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if row["title"] != "OPS" {
		t.Fatalf("title = %#v", row["title"])
	}
	page := rt.Env.UI.Page("documented-home")
	if page == nil {
		t.Fatal("expected documented-home page")
	}
	tile := page.Tile(0, 0)
	if tile == nil || tile.Text() != "OPS" {
		t.Fatalf("expected tile text OPS, got %#v", tile)
	}

	_, err = rt.RunString(context.Background(), `globalThis.__jsverbsStillAlive = 42;`)
	if err != nil {
		t.Fatalf("runtime should remain usable after verb invoke: %v", err)
	}
}
