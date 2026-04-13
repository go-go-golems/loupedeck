package device

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectAuto connects to a Loupedeck Live by automatically locating
// the first USB Loupedeck device in the system.  If you have more
// than one device and want to connect to a specific one, then use
// ConnectPath().
func ConnectAuto() (*Loupedeck, error) {
	return ConnectAutoWithOptions(DefaultWriterOptions)
}

// ConnectAutoWithOptions connects to the first available device using custom writer settings.
func ConnectAutoWithOptions(writerOptions WriterOptions) (*Loupedeck, error) {
	defaultRenderOptions := DefaultRenderOptions
	return ConnectAutoWithWriterAndRenderOptions(writerOptions, &defaultRenderOptions)
}

// ConnectAutoWithWriterAndRenderOptions connects to the first available device using
// custom writer settings and optional render scheduler settings. Passing nil for
// renderOptions disables the render scheduler so draws go directly through the writer.
func ConnectAutoWithWriterAndRenderOptions(writerOptions WriterOptions, renderOptions *RenderOptions) (*Loupedeck, error) {
	return tryConnect(ConnectSerialAuto, writerOptions, renderOptions)
}

// ConnectPath connects to a Loupedeck Live via a specified serial
// device.  If successful it returns a new Loupedeck.
func ConnectPath(serialPath string) (*Loupedeck, error) {
	return ConnectPathWithOptions(serialPath, DefaultWriterOptions)
}

// ConnectPathWithOptions connects to a specific device using custom writer settings.
func ConnectPathWithOptions(serialPath string, writerOptions WriterOptions) (*Loupedeck, error) {
	defaultRenderOptions := DefaultRenderOptions
	return ConnectPathWithWriterAndRenderOptions(serialPath, writerOptions, &defaultRenderOptions)
}

// ConnectPathWithWriterAndRenderOptions connects to a specific device using
// custom writer settings and optional render scheduler settings. Passing nil for
// renderOptions disables the render scheduler so draws go directly through the writer.
func ConnectPathWithWriterAndRenderOptions(serialPath string, writerOptions WriterOptions, renderOptions *RenderOptions) (*Loupedeck, error) {
	return tryConnect(func() (*SerialWebSockConn, error) {
		return ConnectSerialPath(serialPath)
	}, writerOptions, renderOptions)
}

type connectResult struct {
	l   *Loupedeck
	err error
}

// tryConnect helps make connections to USB devices more reliable by
// adding timeout and retry logic.
//
// Without this, 50% of the time my LoupeDeck fails to connect the
// HTTP link for the websocket.  We send the HTTP headers to request a
// websocket connection, but the LoupeDeck never returns.
//
// This is a painful workaround for that.  It uses the generic Go
// pattern for implementing a timeout (do the "real work" in a
// goroutine, feeding answers to a channel, and then add a timeout on
// select).  If the timeout triggers, then it tries a second time to
// connect.  This has a 100% success rate for me.
//
// The actual connection logic is all in doConnect(), below.
func tryConnect(open func() (*SerialWebSockConn, error), writerOptions WriterOptions, renderOptions *RenderOptions) (*Loupedeck, error) {
	c, err := open()
	if err != nil {
		return nil, err
	}

	result := make(chan connectResult, 1)
	go func(conn *SerialWebSockConn) {
		r := connectResult{}
		r.l, r.err = doConnect(conn, writerOptions, renderOptions)
		result <- r
	}(c)

	select {
	case <-time.After(2 * time.Second):
		// timeout
		slog.Info("Timeout! Trying again without timeout.")
		_ = c.Close()
		c2, err := open()
		if err != nil {
			return nil, err
		}
		return doConnect(c2, writerOptions, renderOptions)

	case result := <-result:
		return result.l, result.err
	}
}

func doConnect(c *SerialWebSockConn, writerOptions WriterOptions, renderOptions *RenderOptions) (*Loupedeck, error) {
	dialer := websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			slog.Info("Dialing...")
			return c, nil
		},
		HandshakeTimeout: 1 * time.Second,
	}

	header := http.Header{}

	slog.Info("Attempting to open websocket connection")
	conn, resp, err := dialer.Dial("ws://fake", header)

	if err != nil {
		slog.Warn("dial failed", "err", err)
		return nil, err
	}

	slog.Info("Connect successful", "resp", resp)

	profile, err := resolveProfile(c.Product)
	if err != nil {
		_ = conn.Close()
		_ = c.Close()
		return nil, err
	}

	l := &Loupedeck{
		conn:                 conn,
		serial:               c,
		writerOptions:        writerOptions,
		buttonListeners:      make(map[Button]map[uint64]ButtonFunc),
		buttonUpListeners:    make(map[Button]map[uint64]ButtonFunc),
		knobListeners:        make(map[Knob]map[uint64]KnobFunc),
		touchListeners:       make(map[TouchButton]map[uint64]TouchFunc),
		touchUpListeners:     make(map[TouchButton]map[uint64]TouchFunc),
		Vendor:               c.Vendor,
		Product:              c.Product,
		transactionCallbacks: map[byte]transactionCallback{},
		displays:             map[string]*Display{},
	}
	l.applyProfile(profile)
	l.writer = newOutboundWriter(l.conn, writerOptions)
	if renderOptions != nil {
		l.renderOptions = *renderOptions
		l.renderer = newRenderScheduler(l.writer, l.renderOptions)
	}

	slog.Info("Found Loupedeck", "vendor", l.Vendor, "product", l.Product, "model", l.Model)

	slog.Info("Sending reset.")
	data := make([]byte, 0)
	m := l.NewMessage(Reset, data)
	err = l.Send(m)
	if err != nil {
		return nil, fmt.Errorf("Unable to send: %v", err)
	}

	slog.Info("Setting default brightness.")
	data = []byte{9}
	m = l.NewMessage(SetBrightness, data)
	err = l.Send(m)
	if err != nil {
		return nil, fmt.Errorf("Unable to send: %v", err)
	}

	// Ask the device about itself.  The responses come back
	// asynchronously, so we need to provide a callback.  Since
	// `listen()` hasn't been called yet, we *have* to use
	// callbacks, blocking via 'sendAndWait' isn't going to work.
	m = l.NewMessage(Version, data)
	err = l.SendWithCallback(m, func(m *Message) {
		if len(m.data) < 3 {
			slog.Warn("Received short 'Version' response", "message_type", m.messageType, "length", len(m.data), "data", m.data)
			return
		}
		l.Version = fmt.Sprintf("%d.%d.%d", m.data[0], m.data[1], m.data[2])
		slog.Info("Received 'Version' response", "version", l.Version)
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to send: %v", err)
	}

	m = l.NewMessage(Serial, data)
	err = l.SendWithCallback(m, func(m *Message) {
		if len(m.data) == 0 {
			slog.Warn("Received empty 'Serial' response", "message_type", m.messageType)
			return
		}
		l.SerialNo = string(m.data)
		slog.Info("Received 'Serial' response", "serial", l.SerialNo)
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to send: %v", err)
	}

	return l, nil
}
