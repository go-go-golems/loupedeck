package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/go-go-golems/loupedeck/pkg/device"
)

type sweepResult struct {
	Name          string
	TargetFPS     float64
	Duration      time.Duration
	Completed     int
	AchievedFPS   float64
	WriterDelta   device.WriterStats
	RenderDelta   device.RenderStats
	ListenErr     error
	Stable        bool
	StabilityNote string
}

type buttonScenarioResult struct {
	Name           string
	Duration       time.Duration
	WriterDelta    device.WriterStats
	RenderDelta    device.RenderStats
	ListenErr      error
	TotalTargetFPS float64
	TotalActualFPS float64
	Stable         bool
	PerButton      []perButtonResult
}

type perButtonResult struct {
	Button    device.TouchButton
	TargetFPS float64
	Completed int
	ActualFPS float64
}

type buttonSpec struct {
	Button device.TouchButton
	X      int
	Y      int
}

var touchButtons = []buttonSpec{
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

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	writerOptions := device.WriterOptions{
		QueueSize:    4096,
		SendInterval: 0,
	}

	fmt.Println("Loupedeck Live FPS benchmark")
	fmt.Println("Mode: raw writer benchmark (render scheduler disabled, writer interval = 0)")
	fmt.Println("Display measured as full touchscreen main display = 360x270 for product 0004")
	fmt.Println()

	fullTargets := []float64{1, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 24, 28, 32, 36, 40}
	singleTargets := []float64{20, 40, 60, 80, 100, 120, 140, 160, 200, 240, 280, 320, 400}
	duration := 4 * time.Second

	fullFrames := precomputeFrames(32, 360, 270, 0)
	singleFrames := precomputeFrames(32, 90, 90, 7)

	fmt.Println("== Full-screen main display sweep ==")
	fullResults, err := runSingleRegionSweep(writerOptions, nil, fullTargets, duration, func(l *device.Loupedeck) *device.Display {
		l.SetDisplays()
		return l.GetDisplay("main")
	}, func(d *device.Display, frame int) {
		d.Draw(fullFrames[frame%len(fullFrames)], 0, 0)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "full-screen benchmark failed: %v\n", err)
		os.Exit(1)
	}
	printSweepResults(fullResults)
	fmt.Println()

	fmt.Println("== Single touch-button area sweep (90x90) ==")
	singleResults, err := runSingleRegionSweep(writerOptions, nil, singleTargets, duration, func(l *device.Loupedeck) *device.Display {
		l.SetDisplays()
		return l.GetDisplay("main")
	}, func(d *device.Display, frame int) {
		d.Draw(singleFrames[frame%len(singleFrames)], 0, 0)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "single-button benchmark failed: %v\n", err)
		os.Exit(1)
	}
	printSweepResults(singleResults)
	fmt.Println()

	fmt.Println("== 12-button mixed-framerate animation sweep ==")
	baseRates := []float64{4, 6, 8, 10, 12, 14, 16, 18, 20, 24, 28, 32}
	scales := []float64{0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0}
	buttonFrames := precomputeFrames(32, 90, 90, 19)
	buttonResults, err := runButtonSweep(writerOptions, nil, baseRates, scales, 6*time.Second, buttonFrames)
	if err != nil {
		fmt.Fprintf(os.Stderr, "button-bank benchmark failed: %v\n", err)
		os.Exit(1)
	}
	printButtonSweepResults(buttonResults)
}

func runSingleRegionSweep(
	writerOptions device.WriterOptions,
	renderOptions *device.RenderOptions,
	targets []float64,
	duration time.Duration,
	displayFn func(*device.Loupedeck) *device.Display,
	drawFn func(*device.Display, int),
) ([]sweepResult, error) {
	l, err := device.ConnectAutoWithWriterAndRenderOptions(writerOptions, renderOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = l.Close()
	}()

	display := displayFn(l)
	if display == nil {
		return nil, fmt.Errorf("missing expected display")
	}

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- l.Listen()
	}()

	results := make([]sweepResult, 0, len(targets))
	for _, target := range targets {
		result := runSingleRegionTarget(l, display, duration, target, listenErrCh, drawFn)
		results = append(results, result)
		fmt.Printf("target=%5.1f achieved=%6.2f stable=%-5v sent=%d failed=%d note=%s\n",
			result.TargetFPS,
			result.AchievedFPS,
			result.Stable,
			result.WriterDelta.SentCommands,
			result.WriterDelta.FailedCommands,
			result.StabilityNote,
		)
		if result.ListenErr != nil || result.WriterDelta.FailedCommands > 0 {
			break
		}
	}

	clearDisplay(display)
	return results, nil
}

func runSingleRegionTarget(
	l *device.Loupedeck,
	display *device.Display,
	duration time.Duration,
	targetFPS float64,
	listenErrCh <-chan error,
	drawFn func(*device.Display, int),
) sweepResult {
	startWriter := l.WriterStats()
	startRender := l.RenderStats()
	start := time.Now()
	deadline := start.Add(duration)
	interval := time.Duration(float64(time.Second) / targetFPS)
	next := start
	completed := 0
	frame := 0
	var listenErr error

	for time.Now().Before(deadline) {
		select {
		case err := <-listenErrCh:
			listenErr = err
			goto done
		default:
		}

		drawFn(display, frame)
		completed++
		frame++
		next = next.Add(interval)
		if sleep := time.Until(next); sleep > 0 {
			time.Sleep(sleep)
		}
	}

done:
	elapsed := time.Since(start)
	writerDelta := diffWriterStats(startWriter, l.WriterStats())
	renderDelta := diffRenderStats(startRender, l.RenderStats())
	achieved := float64(completed) / elapsed.Seconds()
	stable, note := evaluateStability(targetFPS, achieved, writerDelta, listenErr)

	return sweepResult{
		TargetFPS:     targetFPS,
		Duration:      elapsed,
		Completed:     completed,
		AchievedFPS:   achieved,
		WriterDelta:   writerDelta,
		RenderDelta:   renderDelta,
		ListenErr:     listenErr,
		Stable:        stable,
		StabilityNote: note,
	}
}

func runButtonSweep(
	writerOptions device.WriterOptions,
	renderOptions *device.RenderOptions,
	baseRates []float64,
	scales []float64,
	duration time.Duration,
	frames []image.Image,
) ([]buttonScenarioResult, error) {
	results := make([]buttonScenarioResult, 0, len(scales))
	for _, scale := range scales {
		l, err := device.ConnectAutoWithWriterAndRenderOptions(writerOptions, renderOptions)
		if err != nil {
			return results, err
		}
		l.SetDisplays()
		mainDisplay := l.GetDisplay("main")
		if mainDisplay == nil {
			_ = l.Close()
			return results, fmt.Errorf("missing main display")
		}

		listenErrCh := make(chan error, 1)
		go func() {
			listenErrCh <- l.Listen()
		}()

		targets := make([]float64, len(baseRates))
		for i, rate := range baseRates {
			targets[i] = rate * scale
		}
		result := runButtonScenario(l, mainDisplay, targets, duration, frames, listenErrCh)
		results = append(results, result)
		printButtonScenarioResult(result, scale)

		clearDisplay(mainDisplay)
		_ = l.Close()
		time.Sleep(750 * time.Millisecond)

		if result.ListenErr != nil || result.WriterDelta.FailedCommands > 0 {
			break
		}
	}
	return results, nil
}

func runButtonScenario(
	l *device.Loupedeck,
	display *device.Display,
	targetRates []float64,
	duration time.Duration,
	frames []image.Image,
	listenErrCh <-chan error,
) buttonScenarioResult {
	startWriter := l.WriterStats()
	startRender := l.RenderStats()
	start := time.Now()
	deadline := start.Add(duration)

	counts := make([]int, len(touchButtons))
	var countsMu sync.Mutex
	listenStop := make(chan struct{})
	var listenErr error
	go func() {
		select {
		case err := <-listenErrCh:
			listenErr = err
		case <-listenStop:
		}
	}()

	var wg sync.WaitGroup
	for i, spec := range touchButtons {
		target := targetRates[i]
		if target <= 0 {
			continue
		}
		wg.Add(1)
		go func(index int, spec buttonSpec, target float64) {
			defer wg.Done()
			interval := time.Duration(float64(time.Second) / target)
			next := start
			frame := index * 3
			for {
				if time.Now().After(deadline) {
					return
				}
				display.Draw(frames[frame%len(frames)], spec.X, spec.Y)
				countsMu.Lock()
				counts[index]++
				countsMu.Unlock()
				frame++
				next = next.Add(interval)
				if sleep := time.Until(next); sleep > 0 {
					time.Sleep(sleep)
				}
			}
		}(i, spec, target)
	}
	wg.Wait()
	close(listenStop)

	elapsed := time.Since(start)
	writerDelta := diffWriterStats(startWriter, l.WriterStats())
	renderDelta := diffRenderStats(startRender, l.RenderStats())

	perButton := make([]perButtonResult, 0, len(touchButtons))
	totalTarget := 0.0
	totalActual := 0.0
	allStable := listenErr == nil && writerDelta.FailedCommands == 0
	for i, spec := range touchButtons {
		actual := float64(counts[i]) / elapsed.Seconds()
		perButton = append(perButton, perButtonResult{
			Button:    spec.Button,
			TargetFPS: targetRates[i],
			Completed: counts[i],
			ActualFPS: actual,
		})
		totalTarget += targetRates[i]
		totalActual += actual
		if targetRates[i] > 0 && actual < targetRates[i]*0.9 {
			allStable = false
		}
	}

	sort.Slice(perButton, func(i, j int) bool {
		return perButton[i].Button < perButton[j].Button
	})

	return buttonScenarioResult{
		Duration:       elapsed,
		WriterDelta:    writerDelta,
		RenderDelta:    renderDelta,
		ListenErr:      listenErr,
		TotalTargetFPS: totalTarget,
		TotalActualFPS: totalActual,
		Stable:         allStable,
		PerButton:      perButton,
	}
}

func evaluateStability(target, achieved float64, writer device.WriterStats, listenErr error) (bool, string) {
	switch {
	case listenErr != nil:
		return false, fmt.Sprintf("listen err: %v", listenErr)
	case writer.FailedCommands > 0:
		return false, fmt.Sprintf("failed commands: %d", writer.FailedCommands)
	case achieved < target*0.95:
		return false, fmt.Sprintf("fell behind target by %.1f%%", 100*(1-achieved/target))
	default:
		return true, "ok"
	}
}

func diffWriterStats(a, b device.WriterStats) device.WriterStats {
	return device.WriterStats{
		QueuedCommands: b.QueuedCommands - a.QueuedCommands,
		SentCommands:   b.SentCommands - a.SentCommands,
		SentMessages:   b.SentMessages - a.SentMessages,
		FailedCommands: b.FailedCommands - a.FailedCommands,
		MaxQueueDepth:  maxInt(a.MaxQueueDepth, b.MaxQueueDepth),
	}
}

func diffRenderStats(a, b device.RenderStats) device.RenderStats {
	return device.RenderStats{
		Invalidations:         b.Invalidations - a.Invalidations,
		CoalescedReplacements: b.CoalescedReplacements - a.CoalescedReplacements,
		FlushedCommands:       b.FlushedCommands - a.FlushedCommands,
		MaxPendingRegionCount: maxInt(a.MaxPendingRegionCount, b.MaxPendingRegionCount),
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func printSweepResults(results []sweepResult) {
	if len(results) == 0 {
		fmt.Println("no results")
		return
	}
	bestStable := 0.0
	bestAchieved := 0.0
	for _, r := range results {
		if r.Stable && r.TargetFPS > bestStable {
			bestStable = r.TargetFPS
		}
		if r.AchievedFPS > bestAchieved {
			bestAchieved = r.AchievedFPS
		}
	}
	fmt.Printf("summary: max stable target fps=%.1f, peak achieved fps=%.2f\n", bestStable, bestAchieved)
}

func printButtonScenarioResult(result buttonScenarioResult, scale float64) {
	fmt.Printf("scale=%0.3f total-target=%6.2f total-actual=%6.2f stable=%v sent=%d failed=%d\n",
		scale,
		result.TotalTargetFPS,
		result.TotalActualFPS,
		result.Stable,
		result.WriterDelta.SentCommands,
		result.WriterDelta.FailedCommands,
	)
	for _, pb := range result.PerButton {
		fmt.Printf("  %-7v target=%5.2f actual=%5.2f completed=%d\n", pb.Button, pb.TargetFPS, pb.ActualFPS, pb.Completed)
	}
	if result.ListenErr != nil {
		fmt.Printf("  listen error: %v\n", result.ListenErr)
	}
}

func printButtonSweepResults(results []buttonScenarioResult) {
	if len(results) == 0 {
		fmt.Println("no button-bank results")
		return
	}
	bestStable := -1.0
	var best buttonScenarioResult
	for _, r := range results {
		if r.Stable && r.TotalTargetFPS > bestStable {
			bestStable = r.TotalTargetFPS
			best = r
		}
	}
	if bestStable < 0 {
		fmt.Println("summary: no fully stable mixed-framerate scenario found")
		return
	}
	fmt.Printf("summary: best stable mixed-framerate total target fps=%0.2f total actual fps=%0.2f\n", best.TotalTargetFPS, best.TotalActualFPS)
}

func clearDisplay(display *device.Display) {
	im := image.NewRGBA(image.Rect(0, 0, display.Width(), display.Height()))
	draw.Draw(im, im.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	display.Draw(im, 0, 0)
	time.Sleep(150 * time.Millisecond)
}

func precomputeFrames(count, width, height, seed int) []image.Image {
	frames := make([]image.Image, 0, count)
	for frame := 0; frame < count; frame++ {
		frames = append(frames, makePatternFrame(width, height, frame, count, seed))
	}
	return frames
}

func makePatternFrame(width, height, frame, total, seed int) image.Image {
	im := image.NewRGBA(image.Rect(0, 0, width, height))
	phase := float64((frame*7 + seed*11) % total)
	cx := float64(width) / 2
	cy := float64(height) / 2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			radius := math.Hypot(dx, dy)
			stripe := int((float64(x+y)/12)+phase) % 2
			ring := int(radius/10+phase/2) % 2
			wave := math.Sin((float64(x)+phase*8)/18.0) + math.Cos((float64(y)-phase*5)/14.0)
			var c color.RGBA
			switch {
			case stripe == 0 && ring == 0:
				c = color.RGBA{R: clampByte(40 + int(wave*20) + (seed*17)%80), G: clampByte(100 + x%120), B: clampByte(180 + y%70), A: 255}
			case stripe == 0:
				c = color.RGBA{R: clampByte(180 + (x+seed*13)%70), G: clampByte(40 + int(radius)%110), B: clampByte(70 + (y*2)%100), A: 255}
			case ring == 0:
				c = color.RGBA{R: clampByte(60 + (y+seed*9)%90), G: clampByte(180 + int(wave*18) + (x % 40)), B: clampByte(60 + (x+y)%80), A: 255}
			default:
				c = color.RGBA{R: clampByte(210 - (x % 90)), G: clampByte(70 + (seed*23)%90), B: clampByte(120 + int(radius)%100), A: 255}
			}
			if (x/15+frame)%5 == 0 || (y/15+frame)%7 == 0 {
				c.R = clampByte(int(c.R) + 25)
				c.G = clampByte(int(c.G) + 25)
				c.B = clampByte(int(c.B) + 25)
			}
			im.Set(x, y, c)
		}
	}
	return im
}

func clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
