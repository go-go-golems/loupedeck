package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	stdraw "image/draw"
	"log/slog"
	"math"
	"os"
	"time"

	loupedeck "github.com/go-go-golems/loupedeck"
	xdraw "golang.org/x/image/draw"
)

const defaultLibraryPath = "/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/sources/local/macos1-icon-library.html"

type buttonSpec struct {
	Button loupedeck.TouchButton
	X      int
	Y      int
}

type animationMode int

const (
	animPulse animationMode = iota
	animBob
	animSlide
	animBlink
)

type animatedIcon struct {
	Spec      buttonSpec
	Name      string
	Sprite    *image.RGBA
	Mode      animationMode
	Speed     float64
	Phase     float64
	InnerBox  int
	BaseScale float64
	Invert    bool
	Border    color.RGBA
}

var grid = []buttonSpec{
	{Button: loupedeck.Touch1, X: 0, Y: 0},
	{Button: loupedeck.Touch2, X: 90, Y: 0},
	{Button: loupedeck.Touch3, X: 180, Y: 0},
	{Button: loupedeck.Touch4, X: 270, Y: 0},
	{Button: loupedeck.Touch5, X: 0, Y: 90},
	{Button: loupedeck.Touch6, X: 90, Y: 90},
	{Button: loupedeck.Touch7, X: 180, Y: 90},
	{Button: loupedeck.Touch8, X: 270, Y: 90},
	{Button: loupedeck.Touch9, X: 0, Y: 180},
	{Button: loupedeck.Touch10, X: 90, Y: 180},
	{Button: loupedeck.Touch11, X: 180, Y: 180},
	{Button: loupedeck.Touch12, X: 270, Y: 180},
}

var borderPalette = []color.RGBA{
	{0x1a, 0x1a, 0x1a, 0xff},
	{0x55, 0x55, 0x55, 0xff},
	{0x88, 0x88, 0x80, 0xff},
	{0x1a, 0x1a, 0x1a, 0xff},
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	libraryPath := flag.String("library", defaultLibraryPath, "path to imported icon-library HTML")
	fps := flag.Float64("fps", 12, "target animation fps")
	duration := flag.Duration("duration", 0, "optional max runtime (0 = run until Circle)")
	flag.Parse()

	if *fps <= 0 {
		fmt.Fprintln(os.Stderr, "fps must be > 0")
		os.Exit(1)
	}

	lib, err := loupedeck.LoadSVGIconLibrary(*libraryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load library: %v\n", err)
		os.Exit(1)
	}
	if len(lib.Icons) < len(grid) {
		fmt.Fprintf(os.Stderr, "need at least %d icons, got %d\n", len(grid), len(lib.Icons))
		os.Exit(1)
	}

	writerOptions := loupedeck.WriterOptions{QueueSize: 1024, SendInterval: 0}
	deck, err := loupedeck.ConnectAutoWithWriterAndRenderOptions(writerOptions, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := deck.Close(); err != nil {
			slog.Warn("close failed", "error", err)
		}
	}()

	deck.SetDisplays()
	mainDisplay := deck.GetDisplay("main")
	if mainDisplay == nil {
		fmt.Fprintln(os.Stderr, "missing main display")
		os.Exit(1)
	}

	icons, err := buildAnimatedIcons(lib)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prepare animated icons: %v\n", err)
		os.Exit(1)
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- deck.Listen()
	}()

	exitCh := make(chan struct{}, 1)
	deck.OnButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		slog.Info("Circle exit requested")
		select {
		case exitCh <- struct{}{}:
		default:
		}
	})

	fmt.Printf("Starting animated SVG button demo library=%s icon_count=%d fps=%.2f\n", *libraryPath, len(lib.Icons), *fps)

	ticker := time.NewTicker(time.Duration(float64(time.Second) / *fps))
	defer ticker.Stop()
	started := time.Now()

	renderAll(mainDisplay, icons, 0)

	for {
		select {
		case err := <-listenErrCh:
			if err != nil {
				slog.Error("listen exited", "error", err)
			}
			clearMainDisplay(mainDisplay)
			return
		case <-exitCh:
			clearMainDisplay(mainDisplay)
			return
		case <-ticker.C:
			elapsed := time.Since(started).Seconds()
			renderAll(mainDisplay, icons, elapsed)
			if *duration > 0 && time.Since(started) >= *duration {
				clearMainDisplay(mainDisplay)
				return
			}
		}
	}
}

func buildAnimatedIcons(lib *loupedeck.SVGIconLibrary) ([]animatedIcon, error) {
	icons := make([]animatedIcon, 0, len(grid))
	for i, spec := range grid {
		base, err := lib.Icons[i].Rasterize(64)
		if err != nil {
			return nil, fmt.Errorf("rasterize %q: %w", lib.Icons[i].Name, err)
		}
		cropped := cropVisible(base)
		icons = append(icons, animatedIcon{
			Spec:      spec,
			Name:      lib.Icons[i].Name,
			Sprite:    cropped,
			Mode:      animationMode(i % 4),
			Speed:     0.75 + float64(i%5)*0.17,
			Phase:     float64(i) * 0.47,
			InnerBox:  62 + (i % 3),
			BaseScale: 0.92 + float64(i%4)*0.025,
			Invert:    i%5 == 0,
			Border:    borderPalette[i%len(borderPalette)],
		})
	}
	return icons, nil
}

func renderAll(display *loupedeck.Display, icons []animatedIcon, elapsed float64) {
	for _, icon := range icons {
		display.Draw(renderFrame(icon, elapsed), icon.Spec.X, icon.Spec.Y)
	}
}

func renderFrame(icon animatedIcon, elapsed float64) image.Image {
	const tile = 90
	bg := color.RGBA{0xf0, 0xf0, 0xe8, 0xff}
	fg := color.RGBA{0x1a, 0x1a, 0x1a, 0xff}
	im := image.NewRGBA(image.Rect(0, 0, tile, tile))
	stdraw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.Point{}, stdraw.Src)
	drawMacFrame(im, fg, icon.Border, elapsed, icon)

	phase := elapsed*icon.Speed + icon.Phase
	scale := icon.BaseScale
	xOffset := 0
	yOffset := 0
	invert := false

	switch icon.Mode {
	case animPulse:
		scale *= 0.92 + 0.12*(0.5+0.5*math.Sin(phase*2*math.Pi))
	case animBob:
		scale *= 0.98
		yOffset = int(math.Round(4 * math.Sin(phase*2*math.Pi)))
	case animSlide:
		scale *= 0.96
		xOffset = int(math.Round(4 * math.Sin(phase*2*math.Pi)))
	case animBlink:
		scale *= 1.02
		invert = math.Sin(phase*2*math.Pi) > 0.55
	}
	if icon.Invert && math.Sin(phase*math.Pi) > 0.8 {
		invert = !invert
	}

	inner := icon.InnerBox
	targetSide := int(math.Round(float64(inner) * scale))
	if targetSide < 28 {
		targetSide = 28
	}
	if targetSide > 72 {
		targetSide = 72
	}

	fit := fitTarget(icon.Sprite.Bounds(), targetSide, targetSide)
	x0 := (tile-fit.Dx())/2 + xOffset
	y0 := (tile-fit.Dy())/2 + yOffset
	y0 -= 2
	dst := image.Rect(x0, y0, x0+fit.Dx(), y0+fit.Dy())
	xdraw.NearestNeighbor.Scale(im, dst, maybeInvert(icon.Sprite, invert, fg, bg), icon.Sprite.Bounds(), stdraw.Over, nil)
	return im
}

func drawMacFrame(dst *image.RGBA, fg, accent color.RGBA, elapsed float64, icon animatedIcon) {
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y += 4 {
		for x := b.Min.X; x < b.Max.X; x += 4 {
			if ((x/4)+(y/4))%2 == 0 {
				dst.Set(x, y, color.RGBA{0xd8, 0xd8, 0xd0, 0xff})
			}
		}
	}
	for x := 0; x < b.Dx(); x++ {
		dst.Set(x, 0, fg)
		dst.Set(x, 1, fg)
		dst.Set(x, b.Dy()-1, fg)
		dst.Set(x, b.Dy()-2, fg)
	}
	for y := 0; y < b.Dy(); y++ {
		dst.Set(0, y, fg)
		dst.Set(1, y, fg)
		dst.Set(b.Dx()-1, y, fg)
		dst.Set(b.Dx()-2, y, fg)
	}
	barY := 8 + int(math.Round(1.5*math.Sin((elapsed+icon.Phase)*math.Pi)))
	for x := 10; x < b.Dx()-10; x += 6 {
		for yy := 0; yy < 2; yy++ {
			dst.Set(x, barY+yy, accent)
			dst.Set(x+1, barY+yy, accent)
			dst.Set(x+2, barY+yy, accent)
		}
	}
}

func cropVisible(img *image.RGBA) *image.RGBA {
	bounds := alphaBounds(img)
	if bounds.Empty() {
		return img
	}
	cropped := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	stdraw.Draw(cropped, cropped.Bounds(), img, bounds.Min, stdraw.Src)
	return cropped
}

func alphaBounds(img image.Image) image.Rectangle {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+1 > maxX {
				maxX = x + 1
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}
	if !found {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX, maxY)
}

func fitTarget(src image.Rectangle, maxW, maxH int) image.Rectangle {
	sw, sh := src.Dx(), src.Dy()
	if sw <= 0 || sh <= 0 {
		return image.Rect(0, 0, 1, 1)
	}
	scale := math.Min(float64(maxW)/float64(sw), float64(maxH)/float64(sh))
	if scale <= 0 {
		scale = 1
	}
	w := int(math.Round(float64(sw) * scale))
	h := int(math.Round(float64(sh) * scale))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return image.Rect(0, 0, w, h)
}

func maybeInvert(src *image.RGBA, invert bool, fg, bg color.RGBA) image.Image {
	if !invert {
		return src
	}
	dst := image.NewRGBA(src.Bounds())
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			if r>>8 < 128 && g>>8 < 128 && b>>8 < 128 {
				dst.Set(x, y, bg)
			} else {
				dst.Set(x, y, fg)
			}
		}
	}
	return dst
}

func clearMainDisplay(display *loupedeck.Display) {
	im := image.NewRGBA(image.Rect(0, 0, display.Width(), display.Height()))
	stdraw.Draw(im, im.Bounds(), &image.Uniform{color.Black}, image.Point{}, stdraw.Src)
	display.Draw(im, 0, 0)
	time.Sleep(150 * time.Millisecond)
}
