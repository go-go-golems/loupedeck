package device

import (
	"image"
	"image/color"
	"sync"
	"testing"
	"time"
)

type fakeWSConn struct {
	mu        sync.Mutex
	writes    [][]byte
	writeTime []time.Time
}

func (f *fakeWSConn) ReadMessage() (int, []byte, error) {
	return 0, nil, nil
}

func (f *fakeWSConn) WriteMessage(messageType int, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	dup := append([]byte(nil), data...)
	f.writes = append(f.writes, dup)
	f.writeTime = append(f.writeTime, time.Now())
	return nil
}

func (f *fakeWSConn) Close() error { return nil }

func (f *fakeWSConn) snapshot() ([][]byte, []time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	writes := make([][]byte, len(f.writes))
	for i, w := range f.writes {
		writes[i] = append([]byte(nil), w...)
	}
	times := append([]time.Time(nil), f.writeTime...)
	return writes, times
}

func newWriterTestLoupedeck(interval time.Duration) (*Loupedeck, *fakeWSConn) {
	conn := &fakeWSConn{}
	l := newTestLoupedeck()
	l.conn = conn
	l.writer = newOutboundWriter(conn, WriterOptions{QueueSize: 16, SendInterval: interval})
	l.transactionCallbacks = map[byte]transactionCallback{}
	return l, conn
}

func TestWriterPreservesSendOrder(t *testing.T) {
	l, conn := newWriterTestLoupedeck(0)
	defer l.writer.Close()

	m1 := l.NewMessage(SetBrightness, []byte{9})
	m2 := l.NewMessage(SetColor, []byte{byte(Circle), 1, 2, 3})

	if err := l.Send(m1); err != nil {
		t.Fatalf("send m1: %v", err)
	}
	if err := l.Send(m2); err != nil {
		t.Fatalf("send m2: %v", err)
	}

	writes, _ := conn.snapshot()
	if len(writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(writes))
	}
	if MessageType(writes[0][1]) != SetBrightness {
		t.Fatalf("expected first message type %v, got %v", SetBrightness, MessageType(writes[0][1]))
	}
	if MessageType(writes[1][1]) != SetColor {
		t.Fatalf("expected second message type %v, got %v", SetColor, MessageType(writes[1][1]))
	}
}

func TestWriterAppliesConfiguredInterval(t *testing.T) {
	interval := 20 * time.Millisecond
	l, conn := newWriterTestLoupedeck(interval)
	defer l.writer.Close()

	if err := l.Send(l.NewMessage(SetBrightness, []byte{9})); err != nil {
		t.Fatalf("send 1: %v", err)
	}
	if err := l.Send(l.NewMessage(SetBrightness, []byte{8})); err != nil {
		t.Fatalf("send 2: %v", err)
	}

	_, times := conn.snapshot()
	if len(times) != 2 {
		t.Fatalf("expected 2 write timestamps, got %d", len(times))
	}
	if delta := times[1].Sub(times[0]); delta < interval-(5*time.Millisecond) {
		t.Fatalf("expected pacing interval >= %v, got %v", interval-(5*time.Millisecond), delta)
	}
}

func TestDisplayDrawUsesWriterForFramebufferThenDraw(t *testing.T) {
	l, conn := newWriterTestLoupedeck(0)
	defer l.writer.Close()

	d := &Display{loupedeck: l, id: 'A', Name: "main"}
	im := image.NewRGBA(image.Rect(0, 0, 1, 1))
	im.Set(0, 0, color.RGBA{255, 0, 0, 255})

	d.Draw(im, 0, 0)

	writes, _ := conn.snapshot()
	if len(writes) != 2 {
		t.Fatalf("expected framebuffer+draw writes, got %d", len(writes))
	}
	if MessageType(writes[0][1]) != WriteFramebuff {
		t.Fatalf("expected first display message to be WriteFramebuff, got %v", MessageType(writes[0][1]))
	}
	if MessageType(writes[1][1]) != Draw {
		t.Fatalf("expected second display message to be Draw, got %v", MessageType(writes[1][1]))
	}
}
