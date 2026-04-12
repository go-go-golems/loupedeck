package render

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/go-go-golems/loupedeck/runtime/ui"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	TileWidth         = 90
	TileHeight        = 90
	MainDisplayWidth  = 360
	MainDisplayHeight = 270
	SideDisplayWidth  = 60
	SideDisplayHeight = 270
)

type DrawTarget interface {
	Draw(im image.Image, xoff, yoff int)
}

type Renderer struct {
	UI      *ui.UI
	Targets map[string]DrawTarget
	Theme   Theme
}

type Theme struct {
	Background color.Color
	Foreground color.Color
	Accent     color.Color
}

func DefaultTheme() Theme {
	return Theme{
		Background: color.RGBA{0x0f, 0x11, 0x17, 0xff},
		Foreground: color.RGBA{0xf5, 0xf7, 0xfa, 0xff},
		Accent:     color.RGBA{0x52, 0xd1, 0xff, 0xff},
	}
}

func TileRect(col, row int) image.Rectangle {
	return image.Rect(col*TileWidth, row*TileHeight, (col+1)*TileWidth, (row+1)*TileHeight)
}

func New(uiRuntime *ui.UI, target DrawTarget) *Renderer {
	return NewWithDisplays(uiRuntime, map[string]DrawTarget{
		ui.DisplayMain: target,
	})
}

func NewWithDisplays(uiRuntime *ui.UI, targets map[string]DrawTarget) *Renderer {
	return &Renderer{
		UI:      uiRuntime,
		Targets: targets,
		Theme:   DefaultTheme(),
	}
}

func (r *Renderer) Flush() int {
	if r == nil || r.UI == nil || len(r.Targets) == 0 {
		return 0
	}
	flushed := 0

	displays := r.UI.DirtyDisplays()
	for _, display := range displays {
		target := r.Targets[display.Name]
		if target == nil {
			continue
		}
		target.Draw(r.renderDisplay(display), 0, 0)
		flushed++
	}
	r.UI.ClearDirtyDisplays(displays)

	tiles := r.UI.DirtyTiles()
	for _, tile := range tiles {
		target := r.Targets[ui.DisplayMain]
		if target == nil {
			continue
		}
		rect := TileRect(tile.Col, tile.Row)
		target.Draw(r.renderTile(tile), rect.Min.X, rect.Min.Y)
		flushed++
	}
	r.UI.ClearDirtyTiles(tiles)

	return flushed
}

func (r *Renderer) renderDisplay(display *ui.Display) image.Image {
	bounds := displayBounds(display)
	im := image.NewRGBA(bounds)
	draw.Draw(im, im.Bounds(), &image.Uniform{r.Theme.Background}, image.Point{}, draw.Src)
	if display == nil || !display.Visible() {
		return im
	}

	accentHeight := 8
	if im.Bounds().Dy() < accentHeight {
		accentHeight = im.Bounds().Dy()
	}
	accent := image.Rect(0, 0, im.Bounds().Dx(), accentHeight)
	draw.Draw(im, accent, &image.Uniform{r.Theme.Accent}, image.Point{}, draw.Src)

	if icon := display.Icon(); icon != "" {
		drawCenteredLabel(im, icon, im.Bounds().Dy()/2-12, r.Theme.Foreground)
	}
	if text := display.Text(); text != "" {
		drawCenteredLabel(im, text, im.Bounds().Dy()/2+12, r.Theme.Foreground)
	}
	return im
}

func (r *Renderer) renderTile(tile *ui.Tile) image.Image {
	im := image.NewRGBA(image.Rect(0, 0, TileWidth, TileHeight))
	draw.Draw(im, im.Bounds(), &image.Uniform{r.Theme.Background}, image.Point{}, draw.Src)
	if tile == nil || !tile.Visible() {
		return im
	}

	accent := image.Rect(0, 0, TileWidth, 8)
	draw.Draw(im, accent, &image.Uniform{r.Theme.Accent}, image.Point{}, draw.Src)

	if icon := tile.Icon(); icon != "" {
		drawCenteredLabel(im, icon, 24, r.Theme.Foreground)
	}
	if text := tile.Text(); text != "" {
		drawCenteredLabel(im, text, 58, r.Theme.Foreground)
	}
	return im
}

func displayBounds(display *ui.Display) image.Rectangle {
	if display == nil {
		return image.Rect(0, 0, MainDisplayWidth, MainDisplayHeight)
	}
	switch display.Name {
	case ui.DisplayLeft, ui.DisplayRight:
		return image.Rect(0, 0, SideDisplayWidth, SideDisplayHeight)
	case ui.DisplayMain:
		fallthrough
	default:
		return image.Rect(0, 0, MainDisplayWidth, MainDisplayHeight)
	}
}

func drawCenteredLabel(dst draw.Image, text string, baseline int, fg color.Color) {
	if text == "" {
		return
	}
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{fg},
		Face: face,
	}
	width := d.MeasureString(text).Round()
	x := (dst.Bounds().Dx() - width) / 2
	if x < 0 {
		x = 0
	}
	d.Dot = fixed.P(x, baseline)
	d.DrawString(text)
}
