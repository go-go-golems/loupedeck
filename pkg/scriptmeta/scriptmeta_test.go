package scriptmeta

import (
	"context"
	"path/filepath"
	"testing"
)

func examplePath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "examples", "js", "12-documented-scene.js")
}

func TestResolveTargetAcceptsScriptShorthand(t *testing.T) {
	target, err := ResolveTarget(filepath.Join("..", "..", "examples", "js", "02"))
	if err != nil {
		t.Fatalf("resolve target shorthand: %v", err)
	}
	if got := filepath.Base(target.EntryFile); got != "02-counter-button.js" {
		t.Fatalf("entry file = %q", got)
	}
}

func TestScanVerbRegistryFindsDocumentedVerb(t *testing.T) {
	target, registry, err := ScanVerbRegistry(examplePath(t))
	if err != nil {
		t.Fatalf("scan verb registry: %v", err)
	}
	verb, err := FindVerb(target, registry, "documented configure")
	if err != nil {
		t.Fatalf("find verb: %v", err)
	}
	if verb.FullPath() != "documented configure" {
		t.Fatalf("verb full path = %q", verb.FullPath())
	}
}

func TestBuildDocStoreExtractsAnnotatedExample(t *testing.T) {
	_, store, err := BuildDocStore(context.Background(), examplePath(t))
	if err != nil {
		t.Fatalf("build doc store: %v", err)
	}
	sym, ok := store.BySymbol["configureScene"]
	if !ok {
		t.Fatal("expected configureScene symbol docs")
	}
	if sym.Summary == "" {
		t.Fatal("expected summary for configureScene")
	}
	if sym.Prose == "" {
		t.Fatal("expected prose for configureScene")
	}
}

func TestExportDocStoreMarkdown(t *testing.T) {
	_, store, err := BuildDocStore(context.Background(), examplePath(t))
	if err != nil {
		t.Fatalf("build doc store: %v", err)
	}
	output, err := ExportDocStore(context.Background(), store, "markdown")
	if err != nil {
		t.Fatalf("export markdown: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected markdown output")
	}
	if string(output) == "" {
		t.Fatal("expected non-empty markdown output")
	}
}
