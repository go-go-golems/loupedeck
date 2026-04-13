package main

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
)

func clearDisplays(displays map[string]*device.Display) {
	for _, display := range displays {
		if display == nil {
			continue
		}
		im := image.NewRGBA(image.Rect(0, 0, display.Width(), display.Height()))
		draw.Draw(im, im.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
		display.Draw(im, 0, 0)
	}
	time.Sleep(100 * time.Millisecond)
}
