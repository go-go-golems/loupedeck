package gfx

import (
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

func TestSurfaceTextNewlineRendersMultipleLines(t *testing.T) {
	s := NewSurface(90, 90)
	s.Text("LINE1\nLINE2", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Brightness: 255,
		Face:       basicfont.Face7x13,
		Center:     true,
	})

	// Face7x13 has Height=13. Two lines with no gap:
	// Line 1: Y=0..12, Line 2: Y=13..25.
	// Verify that pixels appear both above and below the split point.
	face := basicfont.Face7x13
	lineH := face.Metrics().Height.Ceil() // 13
	separator := lineH // row 13

	topNonZero := 0
	botNonZero := 0
	for y := 0; y < separator; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				topNonZero++
			}
		}
	}
	for y := separator; y < separator+lineH; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				botNonZero++
			}
		}
	}

	if topNonZero == 0 {
		t.Fatal("expected first line to produce visible pixels above separator")
	}
	if botNonZero == 0 {
		t.Fatal("expected second line to produce visible pixels below separator")
	}
}

func TestSurfaceTextSingleLineUnchanged(t *testing.T) {
	s := NewSurface(80, 20)
	s.Text("HELLO", TextOptions{
		X:          0,
		Y:          0,
		Width:      80,
		Height:     20,
		Brightness: 180,
		Face:       basicfont.Face7x13,
		Center:     true,
	})
	nonZero := 0
	for y := 0; y < s.Height(); y++ {
		for x := 0; x < s.Width(); x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected single-line text drawing to produce visible pixels")
	}
}

func TestSurfaceTextThreeLines(t *testing.T) {
	s := NewSurface(90, 90)
	s.Text("A\nB\nC", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Height:     30,
		Brightness: 255,
		Face:       basicfont.Face7x13,
		Center:     true,
	})
	nonZero := 0
	for y := 0; y < 90; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected three-line text to produce visible pixels")
	}
}

func TestSurfaceTextLineGapSpacing(t *testing.T) {
	face := basicfont.Face7x13
	lineH := face.Metrics().Height.Ceil() // ~16 for Face7x13
	gap := 4

	s := NewSurface(90, 90)
	s.Text("TOP\nBOT", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Height:     30,
		Brightness: 255,
		Face:       face,
		Center:     true,
		LineGap:    gap,
	})

	// The second line should start at Y = lineH + gap = ~20.
	// Verify that there are pixels at row >= 20 (second line) and < 20 (first line).
	firstLinePixels := 0
	secondLinePixels := 0
	separator := lineH + gap
	for y := 0; y < separator; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				firstLinePixels++
			}
		}
	}
	for y := separator; y < separator+lineH; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				secondLinePixels++
			}
		}
	}
	if firstLinePixels == 0 {
		t.Fatal("expected first line pixels before separator")
	}
	if secondLinePixels == 0 {
		t.Fatalf("expected second line pixels at or after row %d (lineH=%d, gap=%d)", separator, lineH, gap)
	}
}

func TestSurfaceTextEmptyLinePreserved(t *testing.T) {
	s := NewSurface(90, 90)
	s.Text("A\n\nC", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Height:     30,
		Brightness: 255,
		Face:       basicfont.Face7x13,
		Center:     true,
	})
	nonZero := 0
	for y := 0; y < 90; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected A\\n\\nC to produce visible pixels")
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"hello", []string{"hello"}},
		{"a\nb", []string{"a", "b"}},
		{"a\nb\nc", []string{"a", "b", "c"}},
		{"\n", []string{"", ""}},
		{"a\n", []string{"a", ""}},
		{"\na", []string{"", "a"}},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitLines(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitLines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestSurfaceTextNilSurfaceNoPanic(t *testing.T) {
	var s *Surface
	s.Text("hello", TextOptions{}) // should not panic
	s.Text("a\nb", TextOptions{})  // multi-line on nil surface
}

func TestSurfaceTextEmptyStringNoOp(t *testing.T) {
	s := NewSurface(10, 10)
	s.Text("", TextOptions{})
	// Surface should remain all zeros
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if s.At(x, y) != 0 {
				t.Fatalf("expected empty string to be no-op, got non-zero at (%d,%d)", x, y)
			}
		}
	}
}

func TestSurfaceTextTrailingNewline(t *testing.T) {
	s := NewSurface(90, 90)
	s.Text("HELLO\n", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Height:     30,
		Brightness: 255,
		Face:       basicfont.Face7x13,
		Center:     true,
	})
	nonZero := 0
	for y := 0; y < 90; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected HELLO\\n to produce visible pixels")
	}
}

// --- Word wrapping tests ---

func TestWrapTextSingleWordFits(t *testing.T) {
	face := basicfont.Face7x13
	lines := wrapText("HELLO", face, 100)
	if len(lines) != 1 || lines[0] != "HELLO" {
		t.Fatalf("expected [HELLO], got %v", lines)
	}
}

func TestWrapTextSingleWordTooWide(t *testing.T) {
	face := basicfont.Face7x13
	// Even if the word is wider than wrapWidth, it should appear on its own line
	lines := wrapText("SUPERLONGWORD", face, 20)
	if len(lines) != 1 || lines[0] != "SUPERLONGWORD" {
		t.Fatalf("expected [SUPERLONGWORD], got %v", lines)
	}
}

func TestWrapTextTwoWordsWrap(t *testing.T) {
	face := basicfont.Face7x13
	d := &font.Drawer{Face: face}
	combined := d.MeasureString("AA BB").Round()
	single := d.MeasureString("AA").Round()

	// Set wrapWidth so "AA BB" doesn't fit but "AA" does
	wrapWidth := single + 10
	if wrapWidth >= combined {
		t.Skip("wrapWidth too large for this test")
	}

	lines := wrapText("AA BB", face, wrapWidth)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping to produce 2+ lines, got %v", lines)
	}
	if lines[0] != "AA" {
		t.Fatalf("expected first line 'AA', got %q", lines[0])
	}
}

func TestWrapTextEmptyString(t *testing.T) {
	face := basicfont.Face7x13
	lines := wrapText("", face, 80)
	if len(lines) != 1 || lines[0] != "" {
		t.Fatalf("expected empty string line, got %v", lines)
	}
}

func TestWrapTextZeroWidthNoWrap(t *testing.T) {
	face := basicfont.Face7x13
	lines := wrapText("hello world", face, 0)
	if len(lines) != 1 || lines[0] != "hello world" {
		t.Fatalf("expected no wrapping with width=0, got %v", lines)
	}
}

func TestSurfaceTextWrapWidthRendersMultipleLines(t *testing.T) {
	s := NewSurface(90, 90)
	// "VERY LONG SENTENCE" with wrapWidth=50 wraps to 3 lines: VERY, LONG, SENTENCE
	s.Text("VERY LONG SENTENCE", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Brightness: 255,
		Face:       basicfont.Face7x13,
		WrapWidth:  50,
	})

	// Face7x13 lineH=13, 3 lines span rows ~0–39.
	// Verify pixels appear at three distinct vertical positions.
	face := basicfont.Face7x13
	lineH := face.Metrics().Height.Ceil() // 13

	line1NonZero := 0
	line2NonZero := 0
	line3NonZero := 0
	for x := 0; x < 90; x++ {
		for y := 0; y < lineH; y++ {
			if s.At(x, y) != 0 {
				line1NonZero++
			}
		}
		for y := lineH; y < 2*lineH; y++ {
			if s.At(x, y) != 0 {
				line2NonZero++
			}
		}
		for y := 2 * lineH; y < 3*lineH; y++ {
			if s.At(x, y) != 0 {
				line3NonZero++
			}
		}
	}
	if line1NonZero == 0 {
		t.Fatal("expected first wrapped line to produce pixels")
	}
	if line2NonZero == 0 {
		t.Fatal("expected second wrapped line to produce pixels")
	}
	if line3NonZero == 0 {
		t.Fatal("expected third wrapped line to produce pixels")
	}
}

func TestSurfaceTextWrapWithNewlines(t *testing.T) {
	s := NewSurface(90, 90)
	// Text with explicit newlines AND wrapping
	s.Text("AB CD\nEF GH", TextOptions{
		X:          0,
		Y:          0,
		Width:      90,
		Brightness: 255,
		Face:       basicfont.Face7x13,
		WrapWidth:  30, // narrow enough to wrap each pair
	})
	nonZero := 0
	for y := 0; y < 90; y++ {
		for x := 0; x < 90; x++ {
			if s.At(x, y) != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("expected wrapped+newlined text to produce visible pixels")
	}
}

func TestExpandTextLinesNoWrap(t *testing.T) {
	lines := expandTextLines("hello\nworld", TextOptions{Face: basicfont.Face7x13})
	if len(lines) != 2 || lines[0] != "hello" || lines[1] != "world" {
		t.Fatalf("expected [hello, world], got %v", lines)
	}
}

func TestExpandTextLinesWithWrap(t *testing.T) {
	face := basicfont.Face7x13
	d := &font.Drawer{Face: face}
	combined := d.MeasureString("AA BB").Round()
	single := d.MeasureString("AA").Round()
	wrapWidth := single + 10
	if wrapWidth >= combined {
		t.Skip("wrapWidth too large for this test")
	}

	lines := expandTextLines("AA BB", TextOptions{Face: face, WrapWidth: wrapWidth})
	if len(lines) < 2 {
		t.Fatalf("expected wrapping to produce 2+ lines, got %v", lines)
	}
}
