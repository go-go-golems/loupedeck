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
	"strings"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
	xdraw "golang.org/x/image/draw"
)

const defaultLibraryPath = "/home/manuel/code/wesen/2026-04-11--loupedeck-test/ttmp/2026/04/11/LOUPE-004--animated-svg-icon-buttons-for-loupedeck-live/sources/local/macos1-icon-library.html"

type buttonSpec struct {
	Button device.TouchButton
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

type controlAction int

const (
	actionExit controlAction = iota
	actionPrevBank
	actionNextBank
	actionToggleCycle
)

type preparedIcon struct {
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

type animatedIcon struct {
	Spec buttonSpec
	preparedIcon
}

var grid = []buttonSpec{
	{Button: device.Touch1, X: 0, Y: 0},
	{Button: device.Touch2, X: 90, Y: 0},
	{Button: device.Touch3, X: 180, Y: 0},
	{Button: device.Touch4, X: 270, Y: 0},
	{Button: device.Touch5, X: 0, Y: 90},
	{Button: device.Touch6, X: 90, Y: 90},
	{Button: device.Touch7, X: 180, Y: 90},
	{Button: device.Touch8, X: 270, Y: 90},
	{Button: device.Touch9, X: 0, Y: 180},
	{Button: device.Touch10, X: 90, Y: 180},
	{Button: device.Touch11, X: 180, Y: 180},
	{Button: device.Touch12, X: 270, Y: 180},
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
	pageEvery := flag.Duration("page-every", 0, "auto-cycle bank interval (0 disables automatic paging)")
	offset := flag.Int("offset", 0, "starting icon offset within the selected icon list")
	iconsFlag := flag.String("icons", "", "comma-separated icon names to use, in order")
	flag.Parse()

	if *fps <= 0 {
		fmt.Fprintln(os.Stderr, "fps must be > 0")
		os.Exit(1)
	}

	lib, err := device.LoadSVGIconLibrary(*libraryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load library: %v\n", err)
		os.Exit(1)
	}
	selected, err := resolveIconIndexes(lib, *iconsFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "select icons: %v\n", err)
		os.Exit(1)
	}
	if len(selected) == 0 {
		fmt.Fprintln(os.Stderr, "no icons selected")
		os.Exit(1)
	}
	selected = rotateIndexes(selected, *offset)

	prepared, err := buildPreparedIcons(lib, selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prepare animated icons: %v\n", err)
		os.Exit(1)
	}
	banks := totalBanks(len(prepared), len(grid))

	writerOptions := device.WriterOptions{QueueSize: 1024, SendInterval: 0}
	deck, err := device.ConnectAutoWithWriterAndRenderOptions(writerOptions, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := deck.Close(); err != nil {
			slog.Warn("close failed", "error", err)
		}
	}()

	mainDisplay := deck.GetDisplay("main")
	if mainDisplay == nil {
		fmt.Fprintln(os.Stderr, "missing main display")
		os.Exit(1)
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- deck.Listen()
	}()

	controlCh := make(chan controlAction, 16)
	sendControl := func(action controlAction) {
		select {
		case controlCh <- action:
		default:
		}
	}

	deck.OnButton(device.Circle, func(b device.Button, s device.ButtonStatus) {
		sendControl(actionExit)
	})
	deck.OnButton(device.Button1, func(b device.Button, s device.ButtonStatus) {
		sendControl(actionPrevBank)
	})
	deck.OnButton(device.Button2, func(b device.Button, s device.ButtonStatus) {
		sendControl(actionNextBank)
	})
	deck.OnButton(device.Button3, func(b device.Button, s device.ButtonStatus) {
		sendControl(actionToggleCycle)
	})
	deck.OnTouchUp(device.Touch1, func(t device.TouchButton, s device.ButtonStatus, x, y uint16) {
		sendControl(actionPrevBank)
	})
	deck.OnTouchUp(device.Touch12, func(t device.TouchButton, s device.ButtonStatus, x, y uint16) {
		sendControl(actionNextBank)
	})
	deck.OnTouchUp(device.Touch6, func(t device.TouchButton, s device.ButtonStatus, x, y uint16) {
		sendControl(actionToggleCycle)
	})

	fmt.Printf("Starting animated SVG button demo library=%s selected_icons=%d banks=%d fps=%.2f controls=[Button1/Touch1 prev, Button2/Touch12 next, Button3/Touch6 toggle-cycle, Circle exit]\n",
		*libraryPath,
		len(prepared),
		banks,
		*fps,
	)

	currentBank := 0
	cyclingEnabled := *pageEvery > 0 && banks > 1
	icons := makeBank(prepared, currentBank)
	announceBank(currentBank, banks, icons, cyclingEnabled, *pageEvery)

	animationTicker := time.NewTicker(time.Duration(float64(time.Second) / *fps))
	defer animationTicker.Stop()

	var pageTicker *time.Ticker
	if *pageEvery > 0 {
		pageTicker = time.NewTicker(*pageEvery)
		defer pageTicker.Stop()
	}

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
		case action := <-controlCh:
			nowElapsed := time.Since(started).Seconds()
			switch action {
			case actionExit:
				clearMainDisplay(mainDisplay)
				return
			case actionPrevBank:
				if banks > 1 {
					currentBank = mod(currentBank-1, banks)
					icons = makeBank(prepared, currentBank)
					announceBank(currentBank, banks, icons, cyclingEnabled, *pageEvery)
					renderAll(mainDisplay, icons, nowElapsed)
				}
			case actionNextBank:
				if banks > 1 {
					currentBank = mod(currentBank+1, banks)
					icons = makeBank(prepared, currentBank)
					announceBank(currentBank, banks, icons, cyclingEnabled, *pageEvery)
					renderAll(mainDisplay, icons, nowElapsed)
				}
			case actionToggleCycle:
				if pageTicker != nil && banks > 1 {
					cyclingEnabled = !cyclingEnabled
					announceBank(currentBank, banks, icons, cyclingEnabled, *pageEvery)
				}
			}
		case <-animationTicker.C:
			elapsed := time.Since(started).Seconds()
			renderAll(mainDisplay, icons, elapsed)
			if *duration > 0 && time.Since(started) >= *duration {
				clearMainDisplay(mainDisplay)
				return
			}
		case <-pageTick(pageTicker):
			if cyclingEnabled && banks > 1 {
				currentBank = mod(currentBank+1, banks)
				icons = makeBank(prepared, currentBank)
				announceBank(currentBank, banks, icons, cyclingEnabled, *pageEvery)
				renderAll(mainDisplay, icons, time.Since(started).Seconds())
			}
		}
	}
}

func buildPreparedIcons(lib *device.SVGIconLibrary, order []int) ([]preparedIcon, error) {
	icons := make([]preparedIcon, 0, len(order))
	for position, idx := range order {
		base, err := lib.Icons[idx].Rasterize(64)
		if err != nil {
			return nil, fmt.Errorf("rasterize %q: %w", lib.Icons[idx].Name, err)
		}
		cropped := cropVisible(base)
		icons = append(icons, preparedIcon{
			Name:      lib.Icons[idx].Name,
			Sprite:    cropped,
			Mode:      animationMode(position % 4),
			Speed:     0.75 + float64(position%5)*0.17,
			Phase:     float64(position) * 0.47,
			InnerBox:  62 + (position % 3),
			BaseScale: 0.92 + float64(position%4)*0.025,
			Invert:    position%5 == 0,
			Border:    borderPalette[position%len(borderPalette)],
		})
	}
	return icons, nil
}

func resolveIconIndexes(lib *device.SVGIconLibrary, iconsFlag string) ([]int, error) {
	if strings.TrimSpace(iconsFlag) == "" {
		indexes := make([]int, len(lib.Icons))
		for i := range lib.Icons {
			indexes[i] = i
		}
		return indexes, nil
	}
	lookup := map[string]int{}
	for i, icon := range lib.Icons {
		lookup[strings.ToLower(strings.TrimSpace(icon.Name))] = i
	}
	parts := strings.Split(iconsFlag, ",")
	indexes := make([]int, 0, len(parts))
	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		idx, ok := lookup[name]
		if !ok {
			return nil, fmt.Errorf("icon %q not found in library", part)
		}
		indexes = append(indexes, idx)
	}
	if len(indexes) == 0 {
		return nil, fmt.Errorf("icons flag did not resolve to any icons")
	}
	return indexes, nil
}

func rotateIndexes(indexes []int, offset int) []int {
	if len(indexes) == 0 {
		return nil
	}
	offset = mod(offset, len(indexes))
	rotated := make([]int, 0, len(indexes))
	rotated = append(rotated, indexes[offset:]...)
	rotated = append(rotated, indexes[:offset]...)
	return rotated
}

func totalBanks(total, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}

func makeBank(prepared []preparedIcon, bank int) []animatedIcon {
	icons := make([]animatedIcon, 0, len(grid))
	start := bank * len(grid)
	for slot, spec := range grid {
		idx := start + slot
		if idx >= len(prepared) {
			icons = append(icons, animatedIcon{Spec: spec, preparedIcon: blankPreparedIcon(slot)})
			continue
		}
		icons = append(icons, animatedIcon{Spec: spec, preparedIcon: prepared[idx]})
	}
	return icons
}

func blankPreparedIcon(slot int) preparedIcon {
	return preparedIcon{
		Name:      "",
		Sprite:    nil,
		Mode:      animationMode(slot % 4),
		Speed:     1,
		Phase:     float64(slot) * 0.25,
		InnerBox:  60,
		BaseScale: 1,
		Invert:    false,
		Border:    borderPalette[slot%len(borderPalette)],
	}
}

func announceBank(bank, banks int, icons []animatedIcon, cycling bool, every time.Duration) {
	first, last := bankSummaryNames(icons)
	status := "off"
	if cycling {
		status = fmt.Sprintf("on(%s)", every)
	}
	fmt.Printf("Bank %d/%d auto-cycle=%s first=%q last=%q\n", bank+1, banks, status, first, last)
}

func bankSummaryNames(icons []animatedIcon) (string, string) {
	first, last := "", ""
	for _, icon := range icons {
		if icon.Name == "" {
			continue
		}
		if first == "" {
			first = icon.Name
		}
		last = icon.Name
	}
	if first == "" {
		first = "(empty)"
		last = "(empty)"
	}
	return first, last
}

func mod(v, n int) int {
	if n == 0 {
		return 0
	}
	v %= n
	if v < 0 {
		v += n
	}
	return v
}

func pageTick(t *time.Ticker) <-chan time.Time {
	if t == nil {
		return nil
	}
	return t.C
}

func renderAll(display *device.Display, icons []animatedIcon, elapsed float64) {
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

	if icon.Sprite == nil {
		drawPlaceholder(im, fg, icon.Spec.Button)
		return im
	}

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

func drawPlaceholder(dst *image.RGBA, fg color.RGBA, button device.TouchButton) {
	midY := dst.Bounds().Dy() / 2
	for x := 24; x < dst.Bounds().Dx()-24; x++ {
		dst.Set(x, midY, fg)
		dst.Set(x, midY+1, fg)
	}
	//exhaustive:ignore placeholder art intentionally distinguishes only a few exemplar buttons.
	switch button {
	case device.Touch1:
		for i := 0; i < 10; i++ {
			dst.Set(24+i, midY-i/2, fg)
			dst.Set(24+i, midY+i/2, fg)
		}
	case device.Touch12:
		for i := 0; i < 10; i++ {
			dst.Set(dst.Bounds().Dx()-25-i, midY-i/2, fg)
			dst.Set(dst.Bounds().Dx()-25-i, midY+i/2, fg)
		}
	case device.Touch6:
		for i := 0; i < 10; i++ {
			dst.Set(dst.Bounds().Dx()/2-6+i, midY-8, fg)
			dst.Set(dst.Bounds().Dx()/2-6+i, midY+8, fg)
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

func clearMainDisplay(display *device.Display) {
	im := image.NewRGBA(image.Rect(0, 0, display.Width(), display.Height()))
	stdraw.Draw(im, im.Bounds(), &image.Uniform{color.Black}, image.Point{}, stdraw.Src)
	display.Draw(im, 0, 0)
	time.Sleep(150 * time.Millisecond)
}
