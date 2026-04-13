package js

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExampleScriptsBoot(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "examples", "js", "*.js"))
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected example scripts")
	}

	for _, path := range files {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			script, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read script: %v", err)
			}
			rt := NewRuntime(nil)
			defer func() { _ = rt.Close(context.Background()) }()
			if _, err := rt.RunString(context.Background(), string(script)); err != nil {
				t.Fatalf("run script %s: %v", filepath.Base(path), err)
			}
			time.Sleep(20 * time.Millisecond)
		})
	}
}
