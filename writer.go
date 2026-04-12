package loupedeck

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// WriterOptions control the B-lite outbound writer behavior.
type WriterOptions struct {
	QueueSize    int
	SendInterval time.Duration
}

// DefaultWriterOptions are conservative enough for the current hardware experiments.
var DefaultWriterOptions = WriterOptions{
	QueueSize:    128,
	SendInterval: 25 * time.Millisecond,
}

// WriterStats captures coarse outbound writer metrics.
type WriterStats struct {
	QueuedCommands    int
	SentCommands      int
	SentMessages      int
	FailedCommands    int
	MaxQueueDepth     int
	CurrentQueueDepth int
}

type outboundCommand interface {
	Kind() string
	Messages() ([]*Message, error)
}

type singleMessageCommand struct {
	message *Message
}

func (c singleMessageCommand) Kind() string {
	return fmt.Sprintf("message:%02x", c.message.messageType)
}
func (c singleMessageCommand) Messages() ([]*Message, error) {
	return []*Message{c.message}, nil
}

type queuedCommand struct {
	cmd    outboundCommand
	result chan error
}

type outboundWriter struct {
	conn       wsConn
	interval   time.Duration
	queue      chan queuedCommand
	closeCh    chan struct{}
	closeOnce  sync.Once
	mu         sync.Mutex
	closed     bool
	stats      WriterStats
	lastSendAt time.Time
}

func newOutboundWriter(conn wsConn, opts WriterOptions) *outboundWriter {
	if opts.QueueSize <= 0 {
		opts.QueueSize = DefaultWriterOptions.QueueSize
	}
	if opts.SendInterval < 0 {
		opts.SendInterval = 0
	}
	w := &outboundWriter{
		conn:     conn,
		interval: opts.SendInterval,
		queue:    make(chan queuedCommand, opts.QueueSize),
		closeCh:  make(chan struct{}),
	}
	go w.loop()
	return w
}

func (w *outboundWriter) Close() {
	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.closed = true
		close(w.closeCh)
		w.mu.Unlock()
	})
}

func (w *outboundWriter) Stats() WriterStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	stats := w.stats
	stats.CurrentQueueDepth = len(w.queue)
	return stats
}

func (w *outboundWriter) enqueue(cmd outboundCommand) error {
	result := make(chan error, 1)

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return errors.New("outbound writer closed")
	}
	w.mu.Unlock()

	qc := queuedCommand{cmd: cmd, result: result}
	select {
	case <-w.closeCh:
		return errors.New("outbound writer closed")
	case w.queue <- qc:
		w.recordQueued(len(w.queue))
	}

	select {
	case <-w.closeCh:
		return errors.New("outbound writer closed")
	case err := <-result:
		return err
	}
}

func (w *outboundWriter) loop() {
	for {
		select {
		case <-w.closeCh:
			return
		case qc := <-w.queue:
			if err := w.waitForSendWindow(); err != nil {
				qc.result <- err
				continue
			}

			msgs, err := qc.cmd.Messages()
			if err != nil {
				w.recordFailed()
				qc.result <- err
				continue
			}

			for _, m := range msgs {
				b := m.asBytes()
				if err := w.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					w.recordFailed()
					qc.result <- err
					goto next
				}
				w.recordSentMessage()
			}

			w.recordSentCommand()
			qc.result <- nil
		next:
		}
	}
}

func (w *outboundWriter) waitForSendWindow() error {
	w.mu.Lock()
	interval := w.interval
	last := w.lastSendAt
	w.mu.Unlock()

	if interval > 0 && !last.IsZero() {
		wait := interval - time.Since(last)
		if wait > 0 {
			timer := time.NewTimer(wait)
			defer timer.Stop()
			select {
			case <-w.closeCh:
				return errors.New("outbound writer closed")
			case <-timer.C:
			}
		}
	}

	w.mu.Lock()
	w.lastSendAt = time.Now()
	w.mu.Unlock()
	return nil
}

func (w *outboundWriter) recordQueued(depth int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stats.QueuedCommands++
	if depth > w.stats.MaxQueueDepth {
		w.stats.MaxQueueDepth = depth
	}
	slog.Debug("writer queued command", "depth", depth)
}

func (w *outboundWriter) recordSentMessage() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stats.SentMessages++
}

func (w *outboundWriter) recordSentCommand() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stats.SentCommands++
}

func (w *outboundWriter) recordFailed() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stats.FailedCommands++
}
