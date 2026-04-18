package scriptmeta

import (
	"context"
	"os"
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

func TestFindVerbRestrictsExplicitLookupToEntryFile(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "one.js")
	fileB := filepath.Join(dir, "two.js")
	if err := os.WriteFile(fileA, []byte(`
__package__({ name: "one" });
function runScene() { return { ok: true }; }
__verb__("runScene", { name: "run" });
`), 0o644); err != nil {
		t.Fatalf("write fileA: %v", err)
	}
	if err := os.WriteFile(fileB, []byte(`
__package__({ name: "two" });
function runOther() { return { ok: true }; }
__verb__("runOther", { name: "run" });
`), 0o644); err != nil {
		t.Fatalf("write fileB: %v", err)
	}

	target, registry, err := ScanVerbRegistry(fileA)
	if err != nil {
		t.Fatalf("scan verb registry: %v", err)
	}
	verb, err := FindVerb(target, registry, "one run")
	if err != nil {
		t.Fatalf("find own file verb: %v", err)
	}
	if verb.FullPath() != "one run" {
		t.Fatalf("verb full path = %q", verb.FullPath())
	}
	if _, err := FindVerb(target, registry, "two run"); err == nil {
		t.Fatal("expected lookup for other file verb to fail")
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

func TestBuildDocStoreRestrictsFileTargetsToSelectedFile(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "alpha.js")
	fileB := filepath.Join(dir, "beta.js")
	if err := os.WriteFile(fileA, []byte("function alpha() { return 1; }\n__doc__(\"alpha\", { summary: \"Alpha summary\" });\ndoc`---\nsymbol: alpha\n---\nAlpha prose\n`;\n"), 0o644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.WriteFile(fileB, []byte("function beta() { return 1; }\n__doc__(\"beta\", { summary: \"Beta summary\" });\ndoc`---\nsymbol: beta\n---\nBeta prose\n`;\n"), 0o644); err != nil {
		t.Fatalf("write beta: %v", err)
	}

	_, store, err := BuildDocStore(context.Background(), fileA)
	if err != nil {
		t.Fatalf("build doc store: %v", err)
	}
	if _, ok := store.BySymbol["alpha"]; !ok {
		t.Fatal("expected alpha docs")
	}
	if _, ok := store.BySymbol["beta"]; ok {
		t.Fatal("did not expect beta docs when targeting alpha.js")
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
