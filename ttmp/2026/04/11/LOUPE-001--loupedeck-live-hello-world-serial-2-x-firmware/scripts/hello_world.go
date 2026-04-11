// hello_world.go - Minimal graphical hello world for Loupedeck Live
//
// This program demonstrates basic graphical output to a Loupedeck Live
// device with firmware 2.x over serial connection.
//
// Hardware tested: Loupedeck Live (product ID 0004)
// Displays: left (60x270), main (360x270), right (60x270)
//
// Requirements:
//   - Loupedeck Live with firmware 2.x
//   - USB connection (appears as serial device)
//   - Linux/Mac/Windows with USB serial support
//
// Build: go build hello_world.go
// Run: ./hello_world
// Or: go run hello_world.go

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"time"

	"github.com/scottlaird/loupedeck"
)

func main() {
	slog.Info("Loupedeck Hello World - starting...")

	// Step 1: Connect to the Loupedeck device
	// ConnectAuto finds the first available Loupedeck USB device
	slog.Info("Connecting to Loupedeck...")
	l, err := loupedeck.ConnectAuto()
	if err != nil {
		slog.Error("Failed to connect to Loupedeck", "error", err)
		fmt.Fprintf(os.Stderr, "Connection failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Tips:\n")
		fmt.Fprintf(os.Stderr, "  - Ensure the Loupedeck is connected via USB\n")
		fmt.Fprintf(os.Stderr, "  - Check that no other software is using the device\n")
		fmt.Fprintf(os.Stderr, "  - Try unplugging and reconnecting the device\n")
		os.Exit(1)
	}
	defer l.Close()

	slog.Info("Connected successfully", "model", l.Model, "product", l.Product, "version", l.Version)

	// Step 2: Configure displays based on hardware ID
	// This must be called before accessing displays
	l.SetDisplays()

	// Step 3: Start the event listener in a goroutine
	// Listen() is blocking, so we run it in the background
	go l.Listen()

	// Step 4: Get references to the displays
	mainDisplay := l.GetDisplay("main")
	leftDisplay := l.GetDisplay("left")
	rightDisplay := l.GetDisplay("right")

	if mainDisplay == nil || leftDisplay == nil || rightDisplay == nil {
		slog.Error("Failed to get display references")
		os.Exit(1)
	}

	slog.Info("Displays configured",
		"main", fmt.Sprintf("%dx%d", mainDisplay.Width(), mainDisplay.Height()),
		"left", fmt.Sprintf("%dx%d", leftDisplay.Width(), leftDisplay.Height()),
		"right", fmt.Sprintf("%dx%d", rightDisplay.Width(), rightDisplay.Height()),
	)

	// Step 5: Draw "HELLO" on the left display
	slog.Info("Drawing 'HELLO' on left display...")
	if err := drawTextToDisplay(l, leftDisplay, "HELLO", color.White, color.RGBA{0, 0, 128, 255}); err != nil {
		slog.Error("Failed to draw to left display", "error", err)
	}
	time.Sleep(1 * time.Second)

	// Step 6: Draw "WORLD" on the main display (center)
	slog.Info("Drawing 'WORLD' on main display...")
	if err := drawTextToDisplay(l, mainDisplay, "WORLD", color.Black, color.RGBA{255, 255, 0, 255}); err != nil {
		slog.Error("Failed to draw to main display", "error", err)
	}
	time.Sleep(1 * time.Second)

	// Step 7: Draw "LIVE" on the right display
	slog.Info("Drawing 'LIVE' on right display...")
	if err := drawTextToDisplay(l, rightDisplay, "LIVE", color.White, color.RGBA{128, 0, 0, 255}); err != nil {
		slog.Error("Failed to draw to right display", "error", err)
	}
	time.Sleep(1 * time.Second)

	// Step 8: Draw colored rectangles on main display (grid pattern)
	slog.Info("Drawing colored rectangles on main display...")
	drawColorGrid(mainDisplay)
	time.Sleep(2 * time.Second)

	// Step 9: Demonstrate button binding
	slog.Info("Setting up button callback (press CIRCLE button to exit)...")
	exitChan := make(chan bool, 1)
	l.BindButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		if s == loupedeck.ButtonDown {
			slog.Info("Circle button pressed - exiting")
			exitChan <- true
		}
	})

	// Step 10: Demonstrate knob binding
	slog.Info("Setting up knob callback (turn any left knob to see output)...")
	l.BindKnob(loupedeck.Knob1, func(k loupedeck.Knob, delta int) {
		slog.Info("Knob turned", "knob", k, "delta", delta)
	})
	l.BindKnob(loupedeck.Knob2, func(k loupedeck.Knob, delta int) {
		slog.Info("Knob turned", "knob", k, "delta", delta)
	})
	l.BindKnob(loupedeck.Knob3, func(k loupedeck.Knob, delta int) {
		slog.Info("Knob turned", "knob", k, "delta", delta)
	})

	// Step 11: Demonstrate touch binding
	slog.Info("Setting up touch callback (touch the main display)...")
	l.BindTouch(loupedeck.Touch1, func(b loupedeck.TouchButton, s loupedeck.ButtonStatus, x, y uint16) {
		slog.Info("Touch event", "button", b, "status", s, "x", x, "y", y)
	})

	// Wait for exit signal or timeout
	slog.Info("Hello World complete! Waiting for CIRCLE button press (or 30s timeout)...")
	select {
	case <-exitChan:
		slog.Info("Exiting via button press")
	case <-time.After(30 * time.Second):
		slog.Info("Exiting via timeout")
	}

	slog.Info("Goodbye!")
}

// drawTextToDisplay creates a text image and draws it to a display
func drawTextToDisplay(l *loupedeck.Loupedeck, d *loupedeck.Display, text string, fg, bg color.Color) error {
	width := d.Width()
	height := d.Height()

	// Create text image using library helper
	im, err := l.TextInBox(width, height, text, fg, bg)
	if err != nil {
		return fmt.Errorf("failed to create text image: %w", err)
	}

	// Draw to display at offset 0,0 (full display)
	d.Draw(im, 0, 0)
	return nil
}

// drawColorGrid draws a grid of colored rectangles on the display
func drawColorGrid(d *loupedeck.Display) {
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},     // Red
		color.RGBA{0, 255, 0, 255},     // Green
		color.RGBA{0, 0, 255, 255},     // Blue
		color.RGBA{255, 255, 0, 255},   // Yellow
		color.RGBA{255, 0, 255, 255},   // Magenta
		color.RGBA{0, 255, 255, 255},   // Cyan
		color.RGBA{255, 255, 255, 255}, // White
		color.RGBA{128, 128, 128, 255}, // Gray
	}

	// Main display is 360x270, divide into 4x2 grid of 90x135 rectangles
	cellWidth := 90
	cellHeight := 135

	for i, c := range colors {
		x := (i % 4) * cellWidth
		y := (i / 4) * cellHeight

		im := image.NewRGBA(image.Rect(0, 0, cellWidth-2, cellHeight-2))
		draw.Draw(im, im.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)

		d.Draw(im, x+1, y+1)
		time.Sleep(100 * time.Millisecond)
	}
}
