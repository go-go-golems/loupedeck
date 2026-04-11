// feature_tester.go - Comprehensive Loupedeck Live Hardware Feature Tester
//
// This program exercises all major hardware features of the Loupedeck Live:
//   - 6 Knob encoders with value tracking (IntKnob)
//   - 2 TouchDial sliders on left/right displays (drag to adjust)
//   - 12 MultiButton icons on main display 4×3 grid (touch to cycle)
//   - 8 Physical buttons with LED color cycling (SetButtonColor)
//   - Comprehensive event logging for all inputs
//
// Hardware: Loupedeck Live with firmware 2.x
// Connection: USB serial (auto-detected)
//
// Controls:
//   - Turn knobs 1-6: Adjust individual values (shown on displays)
//   - Click knobs: Reset individual value to 0
//   - Touch-drag left/right display: Adjust all 3 knobs simultaneously
//   - Touch main display buttons (1-12): Cycle icon colors
//   - Press physical buttons: Cycle LED colors through rainbow
//   - Press CIRCLE button: Exit program
//
// Build: go build feature_tester.go
// Run: ./feature_tester

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

// Rainbow colors for button LED cycling
var rainbowColors = []color.RGBA{
	{255, 0, 0, 255},     // Red
	{255, 127, 0, 255},   // Orange
	{255, 255, 0, 255},   // Yellow
	{0, 255, 0, 255},     // Green
	{0, 0, 255, 255},     // Blue
	{75, 0, 130, 255},    // Indigo
	{148, 0, 211, 255},   // Violet
	{255, 255, 255, 255}, // White
}

// MultiButton color states (3 states per button)
var multiButtonColors = [][]color.Color{
	{color.RGBA{255, 200, 200, 255}, color.RGBA{255, 100, 100, 255}, color.RGBA{200, 0, 0, 255}},     // Red gradient
	{color.RGBA{255, 220, 200, 255}, color.RGBA{255, 150, 100, 255}, color.RGBA{200, 80, 0, 255}},     // Orange gradient
	{color.RGBA{255, 255, 200, 255}, color.RGBA{255, 255, 100, 255}, color.RGBA{200, 200, 0, 255}},     // Yellow gradient
	{color.RGBA{220, 255, 200, 255}, color.RGBA{150, 255, 100, 255}, color.RGBA{80, 200, 0, 255}},      // Lime gradient
	{color.RGBA{200, 255, 200, 255}, color.RGBA{100, 255, 100, 255}, color.RGBA{0, 200, 0, 255}},       // Green gradient
	{color.RGBA{200, 255, 220, 255}, color.RGBA{100, 255, 150, 255}, color.RGBA{0, 200, 100, 255}},    // Spring gradient
	{color.RGBA{200, 255, 255, 255}, color.RGBA{100, 255, 255, 255}, color.RGBA{0, 200, 200, 255}},     // Cyan gradient
	{color.RGBA{200, 220, 255, 255}, color.RGBA{100, 150, 255, 255}, color.RGBA{0, 100, 200, 255}},     // Azure gradient
	{color.RGBA{200, 200, 255, 255}, color.RGBA{100, 100, 255, 255}, color.RGBA{0, 0, 200, 255}},       // Blue gradient
	{color.RGBA{220, 200, 255, 255}, color.RGBA{150, 100, 255, 255}, color.RGBA{100, 0, 200, 255}},     // Violet gradient
	{color.RGBA{255, 200, 255, 255}, color.RGBA{255, 100, 255, 255}, color.RGBA{200, 0, 200, 255}},     // Magenta gradient
	{color.RGBA{255, 200, 220, 255}, color.RGBA{255, 100, 150, 255}, color.RGBA{200, 0, 80, 255}},      // Rose gradient
}

// Global state for button LED cycling
var buttonColorIndices = make(map[loupedeck.Button]int)
var physicalButtons = []loupedeck.Button{
	loupedeck.Circle,
	loupedeck.Button1,
	loupedeck.Button2,
	loupedeck.Button3,
	loupedeck.Button4,
	loupedeck.Button5,
	loupedeck.Button6,
	loupedeck.Button7,
}

func main() {
	slog.Info("╔════════════════════════════════════════════════════════════╗")
	slog.Info("║     Loupedeck Live - Comprehensive Feature Tester          ║")
	slog.Info("╚════════════════════════════════════════════════════════════╝")
	slog.Info("")
	slog.Info("Features tested:")
	slog.Info("  • 6 Knob encoders with value tracking")
	slog.Info("  • 2 TouchDial sliders (drag left/right displays)")
	slog.Info("  • 12 MultiButton icons (4×3 grid, touch to cycle)")
	slog.Info("  • 8 Physical buttons with LED color cycling")
	slog.Info("  • Comprehensive event logging")
	slog.Info("")
	slog.Info("Controls:")
	slog.Info("  • Turn knobs: Adjust values")
	slog.Info("  • Click knobs: Reset to 0")
	slog.Info("  • Drag displays: Adjust 3 knobs at once")
	slog.Info("  • Touch grid: Cycle icon colors")
	slog.Info("  • CIRCLE button: Exit")
	slog.Info("")

	// Step 1: Connect to device
	slog.Info("Connecting to Loupedeck...")
	l, err := loupedeck.ConnectAuto()
	if err != nil {
		slog.Error("Failed to connect", "error", err)
		fmt.Fprintf(os.Stderr, "\nConnection failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Ensure device is connected via USB and no other software is using it.\n")
		os.Exit(1)
	}
	defer l.Close()

	slog.Info("Connected", "model", l.Model, "product", l.Product)

	// Step 2: Configure displays
	l.SetDisplays()

	// Step 3: Start event listener
	go l.Listen()

	// Step 4: Get display references
	leftDisplay := l.GetDisplay("left")
	mainDisplay := l.GetDisplay("main")
	rightDisplay := l.GetDisplay("right")

	if leftDisplay == nil || mainDisplay == nil || rightDisplay == nil {
		slog.Error("Failed to get displays")
		os.Exit(1)
	}

	slog.Info("Displays ready",
		"left", fmt.Sprintf("%dx%d", leftDisplay.Width(), leftDisplay.Height()),
		"main", fmt.Sprintf("%dx%d", mainDisplay.Width(), mainDisplay.Height()),
		"right", fmt.Sprintf("%dx%d", rightDisplay.Width(), rightDisplay.Height()),
	)

	// Step 5: Create WatchedInt values for all 6 knobs (initial value 128, range 0-255)
	slog.Info("Creating knob values...")
	knobValues := make([]*loupedeck.WatchedInt, 6)
	for i := 0; i < 6; i++ {
		knobValues[i] = loupedeck.NewWatchedInt(128)
	}

	// Step 6: Create IntKnobs for all 6 knobs (individual control)
	slog.Info("Setting up knob encoders...")
	knobIds := []loupedeck.Knob{loupedeck.Knob1, loupedeck.Knob2, loupedeck.Knob3,
		loupedeck.Knob4, loupedeck.Knob5, loupedeck.Knob6}
	
	for i := 0; i < 6; i++ {
		knobNum := i + 1
		watchedInt := knobValues[i]
		
		// Create IntKnob with min=0, max=255
		l.IntKnob(knobIds[i], 0, 255, watchedInt)
		
		// Add watcher to log changes
		watchedInt.AddWatcher(func(kn int) func(int) {
			return func(v int) {
				slog.Info(fmt.Sprintf("[KNOB %d]", kn), "value", v)
			}
		}(knobNum))
	}

	// Step 7: Create TouchDials for left and right displays (sliders)
	slog.Info("Setting up TouchDial sliders...")
	
	// Left display: Knobs 1-3
	_ = l.NewTouchDial(leftDisplay, knobValues[0], knobValues[1], knobValues[2], 0, 255)
	slog.Info("TouchDial LEFT active (Knobs 1-3)")
	
	// Right display: Knobs 4-6
	_ = l.NewTouchDial(rightDisplay, knobValues[3], knobValues[4], knobValues[5], 0, 255)
	slog.Info("TouchDial RIGHT active (Knobs 4-6)")

	// Step 8: Create MultiButtons for main display 4×3 grid
	slog.Info("Setting up MultiButton icons on main display...")
	
	touchButtons := []loupedeck.TouchButton{
		loupedeck.Touch1, loupedeck.Touch2, loupedeck.Touch3, loupedeck.Touch4,
		loupedeck.Touch5, loupedeck.Touch6, loupedeck.Touch7, loupedeck.Touch8,
		loupedeck.Touch9, loupedeck.Touch10, loupedeck.Touch11, loupedeck.Touch12,
	}
	
	for i := 0; i < 12; i++ {
		btnNum := i + 1
		colors := multiButtonColors[i]
		
		// Create WatchedInt for this button's state (0, 1, or 2)
		stateValue := loupedeck.NewWatchedInt(0)
		
		// Create initial icon (state 0)
		icon := createIcon(90, 90, fmt.Sprintf("%d", btnNum), color.Black, colors[0])
		
		// Create MultiButton
		multiBtn := l.NewMultiButton(stateValue, touchButtons[i], icon, 0)
		
		// Add states 1 and 2
		for state := 1; state <= 2; state++ {
			icon := createIcon(90, 90, fmt.Sprintf("%d", btnNum), color.Black, colors[state])
			multiBtn.Add(icon, state)
		}
		
		// Watch for state changes and log
		stateValue.AddWatcher(func(btn int) func(int) {
			return func(s int) {
				slog.Info(fmt.Sprintf("[MULTI ] Touch%d", btn), "state", s)
			}
		}(btnNum))
		
		// Bind touch for logging
		l.BindTouch(touchButtons[i], func(btn int) func(loupedeck.TouchButton, loupedeck.ButtonStatus, uint16, uint16) {
			return func(b loupedeck.TouchButton, s loupedeck.ButtonStatus, x, y uint16) {
				if s == loupedeck.ButtonDown {
					slog.Info(fmt.Sprintf("[TOUCH ] Touch%d", btn), "status", "PRESSED", "x", x, "y", y)
				} else {
					slog.Info(fmt.Sprintf("[TOUCH ] Touch%d", btn), "status", "RELEASED")
				}
			}
		}(btnNum))
	}

	slog.Info("MultiButtons ready (12 icons)")

	// Step 9: Setup physical buttons with LED color cycling
	slog.Info("Setting up physical button LEDs...")
	
	// Initialize button color indices
	for _, btn := range physicalButtons {
		buttonColorIndices[btn] = 0
	}
	
	// Bind all physical buttons
	for _, btn := range physicalButtons {
		button := btn // capture for closure
		btnName := buttonName(button)
		
		l.BindButton(button, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
			if s == loupedeck.ButtonDown {
				// Get current color index
				idx := buttonColorIndices[button]
				color := rainbowColors[idx]
				
				// Set LED color
				if err := l.SetButtonColor(button, color); err != nil {
					slog.Warn("Failed to set button color", "button", btnName, "error", err)
				}
				
				slog.Info(fmt.Sprintf("[BUTTON] %s", btnName), 
					"status", "PRESSED", 
					"color", colorName(idx))
				
				// Advance to next color
				buttonColorIndices[button] = (idx + 1) % len(rainbowColors)
			}
		})
		
		// Also bind release for logging
		l.BindButtonUp(button, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
			slog.Info(fmt.Sprintf("[BUTTON] %s", btnName), "status", "RELEASED")
		})
	}

	// Step 10: Bind knob rotation for logging
	slog.Info("Setting up knob rotation logging...")
	
	for i := 0; i < 6; i++ {
		knobNum := i + 1
		knobId := knobIds[i]
		
		l.BindKnob(knobId, func(k loupedeck.Knob, delta int) {
			// delta is signed: positive = right turn, negative = left turn
			direction := "→"
			if delta < 0 {
				direction = "←"
			}
			slog.Info(fmt.Sprintf("[KNOB %d]", knobNum), 
				"delta", delta, 
				"direction", direction)
		})
	}

	// Step 11: Bind CIRCLE button for exit
	exitChan := make(chan bool, 1)
	
	l.BindButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		if s == loupedeck.ButtonDown {
			slog.Info("[EXIT  ] CIRCLE button pressed - exiting...")
			exitChan <- true
		}
	})

	// Step 12: All setup complete
	slog.Info("")
	slog.Info("╔════════════════════════════════════════════════════════════╗")
	slog.Info("║              Feature Tester Ready!                        ║")
	slog.Info("╠════════════════════════════════════════════════════════════╣")
	slog.Info("║  Left/Right displays: Drag to adjust 3 knobs             ║")
	slog.Info("║  Main display: Touch buttons 1-12 to cycle icons          ║")
	slog.Info("║  Physical buttons: Press to cycle LED colors               ║")
	slog.Info("║  CIRCLE button: Press to exit                              ║")
	slog.Info("╚════════════════════════════════════════════════════════════╝")
	slog.Info("")

	// Step 13: Wait for exit
	select {
	case <-exitChan:
		slog.Info("Exiting via CIRCLE button")
	case <-time.After(5 * time.Minute):
		slog.Info("Exiting via 5-minute timeout")
	}

	// Cleanup: Reset all button LEDs to off
	slog.Info("Resetting button LEDs...")
	offColor := color.RGBA{0, 0, 0, 255}
	for _, btn := range physicalButtons {
		l.SetButtonColor(btn, offColor)
	}
	
	slog.Info("Goodbye!")
}

// createIcon creates a simple icon with text and colored background
func createIcon(width, height int, text string, fg, bg color.Color) image.Image {
	im := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	
	// Draw border
	borderColor := color.RGBA{64, 64, 64, 255}
	for x := 0; x < width; x++ {
		im.Set(x, 0, borderColor)
		im.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		im.Set(0, y, borderColor)
		im.Set(width-1, y, borderColor)
	}
	
	return im
}

// buttonName returns a human-readable name for a button
func buttonName(b loupedeck.Button) string {
	names := map[loupedeck.Button]string{
		loupedeck.Circle:  "Circle",
		loupedeck.Button1: "Button1",
		loupedeck.Button2: "Button2",
		loupedeck.Button3: "Button3",
		loupedeck.Button4: "Button4",
		loupedeck.Button5: "Button5",
		loupedeck.Button6: "Button6",
		loupedeck.Button7: "Button7",
	}
	if name, ok := names[b]; ok {
		return name
	}
	return fmt.Sprintf("Button%d", b)
}

// colorName returns a human-readable name for a rainbow color index
func colorName(idx int) string {
	names := []string{
		"Red", "Orange", "Yellow", "Green",
		"Blue", "Indigo", "Violet", "White",
	}
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return "Unknown"
}
