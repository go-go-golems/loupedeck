package verbs

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

func TestListCommandShowsDocumentedVerb(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--script", examplePath(t)})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute list: %v", err)
	}
	if !strings.Contains(out.String(), "documented configure") {
		t.Fatalf("expected documented configure in output, got %q", out.String())
	}
}

func TestHelpCommandShowsVerbFlags(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"help", "--script", examplePath(t), "--verb", "documented configure"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}
	help := out.String()
	if !strings.Contains(help, "--theme") {
		t.Fatalf("expected --theme in help output, got %q", help)
	}
	if !strings.Contains(help, "--refreshRate") {
		t.Fatalf("expected --refreshRate in help output, got %q", help)
	}
}
