package env

import (
	"fmt"
	"image/color"

	"github.com/go-go-golems/loupedeck/pkg/device"
)

// LoupedeckDeviceControl adapts a physical device connection to the JavaScript
// hardware-control surface. Keeping this adapter in env avoids exposing the raw
// *device.Loupedeck to JavaScript modules and makes mock controls easy in tests.
type LoupedeckDeviceControl struct {
	Deck *device.Loupedeck
}

func (c *LoupedeckDeviceControl) SetButtonColor(button device.Button, col color.RGBA) error {
	if c == nil || c.Deck == nil {
		return fmt.Errorf("loupedeck hardware not connected")
	}
	return c.Deck.SetButtonColor(button, col)
}

func (c *LoupedeckDeviceControl) SetBrightness(brightness int) error {
	if c == nil || c.Deck == nil {
		return fmt.Errorf("loupedeck hardware not connected")
	}
	return c.Deck.SetBrightness(brightness)
}
