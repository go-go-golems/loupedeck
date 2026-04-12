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
	TileWidth  = 90
	TileHeight = 90
)

type DrawTarget interface {
	Draw(im image.Image, xoff, yoff int)
}

type Renderer struct {
	UI     *ui.UI
	Target DrawTarget
	Theme  Theme
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
	return &Renderer{
		UI:     uiRuntime,
		Target: target,
		Theme:  DefaultTheme(),
	}
}

func (r *Renderer) Flush() int {
	if r == nil || r.UI == nil || r.Target == nil {
		return 0
	}
	tiles := r.UI.DirtyTiles()
	if len(tiles) == 0 {
		return 0
	}
	for _, tile := range tiles {
		rect := TileRect(tile.Col, tile.Row)
		im := r.renderTile(tile)
		r.Target.Draw(im, rect.Min.X, rect.Min.Y)
	}
	r.UI.ClearDirtyTiles(tiles)
	return len(tiles)
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
