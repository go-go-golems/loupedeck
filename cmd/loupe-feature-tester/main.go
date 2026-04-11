package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"time"

	loupedeck "github.com/go-go-golems/loupedeck"
)

var rainbowColors = []color.RGBA{
	{255, 0, 0, 255},
	{255, 127, 0, 255},
	{255, 255, 0, 255},
	{0, 255, 0, 255},
	{0, 0, 255, 255},
	{75, 0, 130, 255},
	{148, 0, 211, 255},
	{255, 255, 255, 255},
}

var multiButtonColors = [][]color.Color{
	{color.RGBA{255, 200, 200, 255}, color.RGBA{255, 100, 100, 255}, color.RGBA{200, 0, 0, 255}},
	{color.RGBA{255, 220, 200, 255}, color.RGBA{255, 150, 100, 255}, color.RGBA{200, 80, 0, 255}},
	{color.RGBA{255, 255, 200, 255}, color.RGBA{255, 255, 100, 255}, color.RGBA{200, 200, 0, 255}},
	{color.RGBA{220, 255, 200, 255}, color.RGBA{150, 255, 100, 255}, color.RGBA{80, 200, 0, 255}},
	{color.RGBA{200, 255, 200, 255}, color.RGBA{100, 255, 100, 255}, color.RGBA{0, 200, 0, 255}},
	{color.RGBA{200, 255, 220, 255}, color.RGBA{100, 255, 150, 255}, color.RGBA{0, 200, 100, 255}},
	{color.RGBA{200, 255, 255, 255}, color.RGBA{100, 255, 255, 255}, color.RGBA{0, 200, 200, 255}},
	{color.RGBA{200, 220, 255, 255}, color.RGBA{100, 150, 255, 255}, color.RGBA{0, 100, 200, 255}},
	{color.RGBA{200, 200, 255, 255}, color.RGBA{100, 100, 255, 255}, color.RGBA{0, 0, 200, 255}},
	{color.RGBA{220, 200, 255, 255}, color.RGBA{150, 100, 255, 255}, color.RGBA{100, 0, 200, 255}},
	{color.RGBA{255, 200, 255, 255}, color.RGBA{255, 100, 255, 255}, color.RGBA{200, 0, 200, 255}},
	{color.RGBA{255, 200, 220, 255}, color.RGBA{255, 100, 150, 255}, color.RGBA{200, 0, 80, 255}},
}

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
	slog.Info("Starting Loupedeck feature tester (root package version)")

	writerOptions := loupedeck.WriterOptions{
		QueueSize:    128,
		SendInterval: 40 * time.Millisecond,
	}

	l, err := loupedeck.ConnectAutoWithOptions(writerOptions)
	if err != nil {
		slog.Error("Failed to connect", "error", err)
		fmt.Fprintf(os.Stderr, "connection failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := l.Close(); err != nil {
			slog.Warn("Close failed", "error", err)
		}
	}()

	l.SetDisplays()

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- l.Listen()
	}()

	leftDisplay := l.GetDisplay("left")
	mainDisplay := l.GetDisplay("main")
	rightDisplay := l.GetDisplay("right")
	if leftDisplay == nil || mainDisplay == nil || rightDisplay == nil {
		fmt.Fprintln(os.Stderr, "missing expected Loupedeck Live displays")
		os.Exit(1)
	}

	knobValues := make([]*loupedeck.WatchedInt, 6)
	for i := range knobValues {
		knobValues[i] = loupedeck.NewWatchedInt(128)
	}

	for i := range knobValues {
		knobNum := i + 1
		knobValues[i].AddWatcher(func(kn int) func(int) {
			return func(v int) {
				slog.Info(fmt.Sprintf("[KNOB %d]", kn), "value", v)
			}
		}(knobNum))
	}

	_ = l.NewTouchDial(leftDisplay, knobValues[0], knobValues[1], knobValues[2], 0, 255)
	_ = l.NewTouchDial(rightDisplay, knobValues[3], knobValues[4], knobValues[5], 0, 255)

	touchButtons := []loupedeck.TouchButton{
		loupedeck.Touch1, loupedeck.Touch2, loupedeck.Touch3, loupedeck.Touch4,
		loupedeck.Touch5, loupedeck.Touch6, loupedeck.Touch7, loupedeck.Touch8,
		loupedeck.Touch9, loupedeck.Touch10, loupedeck.Touch11, loupedeck.Touch12,
	}
	multiButtons := map[loupedeck.TouchButton]*loupedeck.MultiButton{}

	for i, touchButton := range touchButtons {
		btnNum := i + 1
		colors := multiButtonColors[i]
		stateValue := loupedeck.NewWatchedInt(0)
		icon := createIcon(90, 90, fmt.Sprintf("%d", btnNum), color.Black, colors[0])
		multiBtn := l.NewMultiButton(stateValue, touchButton, icon, 0)
		for state := 1; state <= 2; state++ {
			multiBtn.Add(createIcon(90, 90, fmt.Sprintf("%d", btnNum), color.Black, colors[state]), state)
		}
		multiButtons[touchButton] = multiBtn

		stateValue.AddWatcher(func(btn int) func(int) {
			return func(s int) {
				slog.Info(fmt.Sprintf("[MULTI ] Touch%d", btn), "state", s)
			}
		}(btnNum))

		bx, by := touchButtonCoordinates(touchButton)
		flashColor := rainbowColors[i%len(rainbowColors)]
		l.OnTouch(touchButton, func(btn int, buttonX, buttonY int, fc color.RGBA) func(loupedeck.TouchButton, loupedeck.ButtonStatus, uint16, uint16) {
			return func(b loupedeck.TouchButton, s loupedeck.ButtonStatus, x, y uint16) {
				slog.Info(fmt.Sprintf("[TOUCH ] Touch%d", btn), "status", "PRESSED", "x", x, "y", y)
				flash := image.NewRGBA(image.Rect(0, 0, 90, 90))
				draw.Draw(flash, flash.Bounds(), &image.Uniform{fc}, image.Point{}, draw.Src)
				mainDisplay.Draw(flash, buttonX, buttonY)
			}
		}(btnNum, bx, by, flashColor))
		l.OnTouchUp(touchButton, func(btn int, mb *loupedeck.MultiButton) func(loupedeck.TouchButton, loupedeck.ButtonStatus, uint16, uint16) {
			return func(b loupedeck.TouchButton, s loupedeck.ButtonStatus, x, y uint16) {
				slog.Info(fmt.Sprintf("[TOUCH ] Touch%d", btn), "status", "RELEASED")
				mb.Draw()
			}
		}(btnNum, multiBtn))
	}

	knobIDs := []loupedeck.Knob{
		loupedeck.Knob1, loupedeck.Knob2, loupedeck.Knob3,
		loupedeck.Knob4, loupedeck.Knob5, loupedeck.Knob6,
	}
	for i, knobID := range knobIDs {
		knobNum := i + 1
		l.OnKnob(knobID, func(kn int) func(loupedeck.Knob, int) {
			return func(k loupedeck.Knob, delta int) {
				direction := "→"
				if delta < 0 {
					direction = "←"
				}
				slog.Info(fmt.Sprintf("[KNOB %d]", kn), "delta", delta, "direction", direction, "raw_event", true)
			}
		}(knobNum))
	}

	for _, btn := range physicalButtons {
		buttonColorIndices[btn] = 0
		button := btn
		btnName := buttonName(button)
		l.OnButton(button, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
			idx := buttonColorIndices[button]
			if err := l.SetButtonColor(button, rainbowColors[idx]); err != nil {
				slog.Warn("SetButtonColor failed", "button", btnName, "error", err)
			}
			slog.Info(fmt.Sprintf("[BUTTON] %s", btnName), "status", "PRESSED", "color", colorName(idx))
			buttonColorIndices[button] = (idx + 1) % len(rainbowColors)
		})
		l.OnButtonUp(button, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
			slog.Info(fmt.Sprintf("[BUTTON] %s", btnName), "status", "RELEASED")
		})
	}

	exitCh := make(chan struct{}, 1)
	l.OnButton(loupedeck.Circle, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
		slog.Info("[EXIT  ] CIRCLE button pressed - exiting...")
		select {
		case exitCh <- struct{}{}:
		default:
		}
	})

	slog.Info("Feature tester ready", "writer_stats", l.WriterStats())

	select {
	case err := <-listenErrCh:
		if err != nil {
			slog.Error("Listen exited", "error", err)
		}
	case <-exitCh:
		slog.Info("Exiting via Circle button")
	case <-time.After(5 * time.Minute):
		slog.Info("Exiting via timeout")
	}

	off := color.RGBA{0, 0, 0, 255}
	for _, btn := range physicalButtons {
		if err := l.SetButtonColor(btn, off); err != nil {
			slog.Warn("Failed to reset button color", "button", btn, "error", err)
		}
	}

	slog.Info("Final writer stats", "stats", l.WriterStats())
}

func touchButtonCoordinates(b loupedeck.TouchButton) (int, int) {
	switch b {
	case loupedeck.Touch1:
		return 0, 0
	case loupedeck.Touch2:
		return 90, 0
	case loupedeck.Touch3:
		return 180, 0
	case loupedeck.Touch4:
		return 270, 0
	case loupedeck.Touch5:
		return 0, 90
	case loupedeck.Touch6:
		return 90, 90
	case loupedeck.Touch7:
		return 180, 90
	case loupedeck.Touch8:
		return 270, 90
	case loupedeck.Touch9:
		return 0, 180
	case loupedeck.Touch10:
		return 90, 180
	case loupedeck.Touch11:
		return 180, 180
	case loupedeck.Touch12:
		return 270, 180
	default:
		return 0, 0
	}
}

func createIcon(width, height int, text string, fg, bg color.Color) image.Image {
	im := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	borderColor := color.RGBA{64, 64, 64, 255}
	for x := 0; x < width; x++ {
		im.Set(x, 0, borderColor)
		im.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		im.Set(0, y, borderColor)
		im.Set(width-1, y, borderColor)
	}
	_ = text
	_ = fg
	return im
}

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

func colorName(idx int) string {
	names := []string{"Red", "Orange", "Yellow", "Green", "Blue", "Indigo", "Violet", "White"}
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return "Unknown"
}
