package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/render"
)

type pngTarget struct {
	outDir string
	count  int
}

func (p *pngTarget) Draw(im image.Image, xoff, yoff int) {
	p.count++
	name := filepath.Join(p.outDir, fmt.Sprintf("tile-%03d-%03d.png", xoff, yoff))
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, im); err != nil {
		panic(err)
	}
}

func main() {
	scriptPath := flag.String("script", "", "Path to a JS file that uses require(\"loupedeck/state\") and require(\"loupedeck/ui\")")
	outDir := flag.String("out-dir", "./out/loupe-js-demo", "Directory for rendered tile PNGs")
	flag.Parse()

	if *scriptPath == "" {
		fmt.Fprintln(os.Stderr, "missing --script")
		os.Exit(2)
	}
	script, err := os.ReadFile(*scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read script: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir out-dir: %v\n", err)
		os.Exit(1)
	}

	env := envpkg.Ensure(nil)
	rt := jsruntime.NewRuntime(env)
	defer func() { _ = rt.Close(context.Background()) }()
	if _, err := rt.RunString(context.Background(), string(script)); err != nil {
		fmt.Fprintf(os.Stderr, "run script: %v\n", err)
		os.Exit(1)
	}

	target := &pngTarget{outDir: *outDir}
	r := render.New(rt.Env.UI, target)
	count := r.Flush()
	fmt.Printf("Rendered %d dirty tiles into %s\n", count, *outDir)
}
