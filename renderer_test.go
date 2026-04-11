package loupedeck

import (
	"image"
	"image/color"
	"testing"
	"time"

	"maze.io/x/pixel/pixelcolor"
)

func newRenderTestLoupedeck(flushInterval time.Duration) (*Loupedeck, *fakeWSConn) {
	conn := &fakeWSConn{}
	l := newTestLoupedeck()
	l.conn = conn
	l.writer = newOutboundWriter(conn, WriterOptions{QueueSize: 16, SendInterval: 0})
	l.renderer = newRenderScheduler(l.writer, RenderOptions{FlushInterval: flushInterval, QueueSize: 16})
	l.transactionCallbacks = map[byte]transactionCallback{}
	return l, conn
}

func TestRendererCoalescesRepeatedRegionInvalidations(t *testing.T) {
	l, conn := newRenderTestLoupedeck(15 * time.Millisecond)
	defer l.renderer.Close()
	defer l.writer.Close()

	d := &Display{loupedeck: l, id: 'A', Name: "main"}

	im1 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	im1.Set(0, 0, color.RGBA{255, 0, 0, 255})
	im2 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	im2.Set(0, 0, color.RGBA{0, 255, 0, 255})

	d.Draw(im1, 0, 0)
	d.Draw(im2, 0, 0)

	time.Sleep(60 * time.Millisecond)

	writes, _ := conn.snapshot()
	if len(writes) != 2 {
		t.Fatalf("expected one coalesced framebuffer+draw pair, got %d writes", len(writes))
	}
	if MessageType(writes[0][1]) != WriteFramebuff {
		t.Fatalf("expected first write to be framebuffer, got %v", MessageType(writes[0][1]))
	}
	if MessageType(writes[1][1]) != Draw {
		t.Fatalf("expected second write to be draw, got %v", MessageType(writes[1][1]))
	}

	expectedPixel := pixelcolor.ToRGB565(color.RGBA{0, 255, 0, 255})
	low := byte(expectedPixel & 0xff)
	high := byte(expectedPixel >> 8)
	fb := writes[0]
	if gotLow, gotHigh := fb[len(fb)-2], fb[len(fb)-1]; gotLow != low || gotHigh != high {
		t.Fatalf("expected coalesced framebuffer to use last image pixel bytes [%d %d], got [%d %d]", low, high, gotLow, gotHigh)
	}

	stats := l.RenderStats()
	if stats.Invalidations != 2 {
		t.Fatalf("expected 2 invalidations, got %+v", stats)
	}
	if stats.CoalescedReplacements < 1 {
		t.Fatalf("expected at least one coalesced replacement, got %+v", stats)
	}
	if stats.FlushedCommands != 1 {
		t.Fatalf("expected 1 flushed command, got %+v", stats)
	}
}
