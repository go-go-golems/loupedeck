package doc

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func examplePath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "..", "..", "examples", "js", "12-documented-scene.js")
}

func TestDocCommandOutputsMarkdown(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--script", examplePath(t), "--format", "markdown"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute doc command: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "Documented scene configuration") {
		t.Fatalf("expected prose heading in markdown output, got %q", text)
	}
	if !strings.Contains(text, "Symbol: configureScene") {
		t.Fatalf("expected symbol section in markdown output, got %q", text)
	}
}
