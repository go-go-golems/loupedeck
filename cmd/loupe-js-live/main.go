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
	mainDisplay := deckConn.GetDisplay("main")
	if mainDisplay == nil {
		fmt.Fprintln(os.Stderr, "missing main display")
		os.Exit(1)
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- deckConn.Listen()
	}()

	env := envpkg.Ensure(nil)
	env.Host.Attach(deckConn)
	rt := jsruntime.NewRuntime(env)
	defer rt.Close(nil)
	if _, err := rt.RunString(rt.Context(), string(script)); err != nil {
		fmt.Fprintf(os.Stderr, "run script: %v\n", err)
		os.Exit(1)
	}

	renderer := render.New(rt.Env.UI, mainDisplay)
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
			clearMain(mainDisplay)
			return
		case <-sigCh:
			clearMain(mainDisplay)
			return
		case <-exitCh:
			clearMain(mainDisplay)
			return
		case <-timeout:
			clearMain(mainDisplay)
			return
		}
	}
}

func clearMain(display *loupedeck.Display) {
	if display == nil {
		return
	}
	im := image.NewRGBA(image.Rect(0, 0, display.Width(), display.Height()))
	draw.Draw(im, im.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	display.Draw(im, 0, 0)
	time.Sleep(100 * time.Millisecond)
}
