package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	loupedeck "github.com/go-go-golems/loupedeck"
	jsruntime "github.com/go-go-golems/loupedeck/runtime/js"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/render"
)

func main() {
	scriptPath := flag.String("script", "", "Path to a JS file")
	devicePath := flag.String("device", "", "Optional serial device path (default: auto-detect)")
	duration := flag.Duration("duration", 15*time.Second, "How long to run before exiting; 0 means run until interrupted")
	flushInterval := flag.Duration("flush-interval", 16*time.Millisecond, "How often to flush retained UI to the device")
	queueSize := flag.Int("queue-size", 256, "Writer queue size")
	sendInterval := flag.Duration("send-interval", 35*time.Millisecond, "Writer pacing interval")
	logEvents := flag.Bool("log-events", false, "Log high-level button/touch/knob events")
	exitOnCircle := flag.Bool("exit-on-circle", true, "Exit when the Circle button is pressed")
	flag.Parse()

	if *scriptPath == "" {
		fmt.Fprintln(os.Stderr, "missing --script")
		os.Exit(2)
	}
	script, err := os.ReadFile(*scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read script: %v\n", err)
		os.Exit(1)
	}

	writerOptions := loupedeck.WriterOptions{QueueSize: *queueSize, SendInterval: *sendInterval}
	var deckConn *loupedeck.Loupedeck
	if *devicePath == "" {
		deckConn, err = loupedeck.ConnectAutoWithOptions(writerOptions)
	} else {
		deckConn, err = loupedeck.ConnectPathWithOptions(*devicePath, writerOptions)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := deckConn.Close(); err != nil {
			slog.Warn("close failed", "error", err)
		}
	}()
	deckConn.SetDisplays()
	displays := map[string]*loupedeck.Display{
		"left":  deckConn.GetDisplay("left"),
		"main":  deckConn.GetDisplay("main"),
		"right": deckConn.GetDisplay("right"),
	}
	if displays["main"] == nil {
		fmt.Fprintln(os.Stderr, "missing main display")
		os.Exit(1)
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- deckConn.Listen()
	}()

	env := envpkg.Ensure(nil)
	env.Host.Attach(deckConn)
	if *logEvents {
		registerEventLogging(env)
	}
	rt := jsruntime.NewRuntime(env)
	defer rt.Close(nil)
	if _, err := rt.RunString(rt.Context(), string(script)); err != nil {
		fmt.Fprintf(os.Stderr, "run script: %v\n", err)
		os.Exit(1)
	}

	renderer := render.NewWithDisplays(rt.Env.UI, map[string]render.DrawTarget{
		"left":  displays["left"],
		"main":  displays["main"],
		"right": displays["right"],
	})
	renderer.Flush()

	exitCh := make(chan struct{}, 1)
	if *exitOnCircle {
		rt.Env.Host.OnButton(loupedeck.Circle, func(loupedeck.Button, loupedeck.ButtonStatus) {
			select {
			case exitCh <- struct{}{}:
			default:
			}
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(*flushInterval)
	defer ticker.Stop()
	var timeout <-chan time.Time
	if *duration > 0 {
		timer := time.NewTimer(*duration)
		defer timer.Stop()
		timeout = timer.C
	}

	slog.Info("Loupedeck JS live runner started", "script", *scriptPath, "duration", *duration, "flush_interval", *flushInterval)
	for {
		select {
		case <-ticker.C:
			renderer.Flush()
		case err := <-listenErrCh:
			if err != nil {
				fmt.Fprintf(os.Stderr, "listen: %v\n", err)
			}
			clearDisplays(displays)
			return
		case <-sigCh:
			clearDisplays(displays)
			return
		case <-exitCh:
			clearDisplays(displays)
			return
		case <-timeout:
			clearDisplays(displays)
			return
		}
	}
}

func clearDisplays(displays map[string]*loupedeck.Display) {
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

func registerEventLogging(env *envpkg.Environment) {
	if env == nil {
		return
	}
	for _, button := range []loupedeck.Button{
		loupedeck.Circle,
		loupedeck.Button1,
		loupedeck.Button2,
		loupedeck.Button3,
		loupedeck.Button4,
		loupedeck.Button5,
		loupedeck.Button6,
		loupedeck.Button7,
	} {
		button := button
		env.Host.OnButton(button, func(b loupedeck.Button, s loupedeck.ButtonStatus) {
			slog.Info("button event", "button", buttonName(b), "status", buttonStatusName(s))
		})
	}
	for _, touch := range []loupedeck.TouchButton{
		loupedeck.Touch1,
		loupedeck.Touch2,
		loupedeck.Touch3,
		loupedeck.Touch4,
		loupedeck.Touch5,
		loupedeck.Touch6,
		loupedeck.Touch7,
		loupedeck.Touch8,
		loupedeck.Touch9,
		loupedeck.Touch10,
		loupedeck.Touch11,
		loupedeck.Touch12,
	} {
		touch := touch
		env.Host.OnTouch(touch, func(t loupedeck.TouchButton, s loupedeck.ButtonStatus, x, y uint16) {
			slog.Info("touch event", "touch", touchName(t), "status", buttonStatusName(s), "x", x, "y", y)
		})
	}
	for _, knob := range []loupedeck.Knob{
		loupedeck.Knob1,
		loupedeck.Knob2,
		loupedeck.Knob3,
		loupedeck.Knob4,
		loupedeck.Knob5,
		loupedeck.Knob6,
	} {
		knob := knob
		env.Host.OnKnob(knob, func(k loupedeck.Knob, value int) {
			slog.Info("knob event", "knob", knobName(k), "value", value)
		})
	}
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

func touchName(t loupedeck.TouchButton) string {
	names := map[loupedeck.TouchButton]string{
		loupedeck.Touch1:  "Touch1",
		loupedeck.Touch2:  "Touch2",
		loupedeck.Touch3:  "Touch3",
		loupedeck.Touch4:  "Touch4",
		loupedeck.Touch5:  "Touch5",
		loupedeck.Touch6:  "Touch6",
		loupedeck.Touch7:  "Touch7",
		loupedeck.Touch8:  "Touch8",
		loupedeck.Touch9:  "Touch9",
		loupedeck.Touch10: "Touch10",
		loupedeck.Touch11: "Touch11",
		loupedeck.Touch12: "Touch12",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return fmt.Sprintf("Touch%d", t)
}

func knobName(k loupedeck.Knob) string {
	names := map[loupedeck.Knob]string{
		loupedeck.Knob1: "Knob1",
		loupedeck.Knob2: "Knob2",
		loupedeck.Knob3: "Knob3",
		loupedeck.Knob4: "Knob4",
		loupedeck.Knob5: "Knob5",
		loupedeck.Knob6: "Knob6",
	}
	if name, ok := names[k]; ok {
		return name
	}
	return fmt.Sprintf("Knob%d", k)
}

func buttonStatusName(s loupedeck.ButtonStatus) string {
	switch s {
	case loupedeck.ButtonDown:
		return "down"
	case loupedeck.ButtonUp:
		return "up"
	default:
		return fmt.Sprintf("status(%d)", s)
	}
}
