package main

import (
	"context"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-go-golems/loupedeck/runtime/js"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/render"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

type captureTarget struct {
	mu sync.Mutex
	im image.Image
}

func cloneImage(src image.Image) image.Image {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func (t *captureTarget) Draw(im image.Image, _xoff, _yoff int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.im = cloneImage(im)
}

func (t *captureTarget) Image() image.Image {
	t.mu.Lock()
	defer t.mu.Unlock()
	return cloneImage(t.im)
}

func main() {
	scriptPath := flag.String("script", "", "path to JS scene")
	outPath := flag.String("out", "scene-preview.png", "output PNG path")
	wait := flag.Duration("wait", 300*time.Millisecond, "time to wait for scene rendering")
	flag.Parse()

	if *scriptPath == "" {
		panic("--script is required")
	}

	scriptBytes, err := os.ReadFile(*scriptPath)
	if err != nil {
		panic(err)
	}

	env := envpkg.Ensure(nil)
	rt := js.NewRuntime(env)
	defer rt.Close(context.Background())

	leftTarget := &captureTarget{}
	mainTarget := &captureTarget{}
	rightTarget := &captureTarget{}
	r := render.NewWithDisplays(env.UI, map[string]render.DrawTarget{
		ui.DisplayLeft:  leftTarget,
		ui.DisplayMain:  mainTarget,
		ui.DisplayRight: rightTarget,
	})
	r.Theme = render.Theme{
		Background: color.Black,
		Foreground: color.White,
		Accent:     color.White,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	env.Present.SetFlushFunc(func() (int, error) {
		return r.Flush(), nil
	})
	env.Present.Start(ctx)
	defer env.Present.Close()

	if _, err := rt.RunString(nil, string(scriptBytes)); err != nil {
		panic(err)
	}

	time.Sleep(*wait)

	canvas := image.NewRGBA(image.Rect(0, 0, render.SideDisplayWidth+render.MainDisplayWidth+render.SideDisplayWidth, render.MainDisplayHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	if im := leftTarget.Image(); im != nil {
		draw.Draw(canvas, image.Rect(0, 0, render.SideDisplayWidth, render.SideDisplayHeight), im, im.Bounds().Min, draw.Src)
	}
	if im := mainTarget.Image(); im != nil {
		draw.Draw(canvas, image.Rect(render.SideDisplayWidth, 0, render.SideDisplayWidth+render.MainDisplayWidth, render.MainDisplayHeight), im, im.Bounds().Min, draw.Src)
	}
	if im := rightTarget.Image(); im != nil {
		draw.Draw(canvas, image.Rect(render.SideDisplayWidth+render.MainDisplayWidth, 0, render.SideDisplayWidth+render.MainDisplayWidth+render.SideDisplayWidth, render.SideDisplayHeight), im, im.Bounds().Min, draw.Src)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, canvas); err != nil {
		panic(err)
	}
}
