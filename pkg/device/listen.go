package device

import (
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
)

// Listen waits for events from the Loupedeck and calls
// callbacks as configured. It returns when the read loop exits.
func (l *Loupedeck) Listen() error {
	slog.Info("Listening")
	for {
		websocketMsgType, message, err := l.conn.ReadMessage()

		if err != nil {
			slog.Warn("Read error, exiting", "error", err)
			return fmt.Errorf("websocket read failed: %w", err)
		}

		if len(message) == 0 {
			slog.Warn("Received a 0-byte message.  Skipping")
			continue
		}

		if websocketMsgType != websocket.BinaryMessage {
			slog.Warn("Unknown websocket message type received", "type", websocketMsgType)
		}

		m, _ := l.ParseMessage(message)
		slog.Info("Read", "message", m.String())

		if m.transactionID != 0 {
			if c := l.takeTransactionCallback(m.transactionID); c != nil {
				slog.Info("Callback found, calling")
				c(m)
			}
			continue
		}

		//exhaustive:ignore protocol-level message handling intentionally ignores many response-only message types here.
		switch m.messageType {
		case ButtonPress:
			button := Button(binary.BigEndian.Uint16(message[2:]))
			upDown := ButtonStatus(message[4])
			if !l.dispatchButton(button, upDown) {
				slog.Info("Received uncaught button press message", "button", button, "upDown", upDown, "message", message)
			}
		case KnobRotate:
			knob := Knob(binary.BigEndian.Uint16(message[2:]))
			value := int(message[4])
			v := value
			if value == 255 {
				v = -1
			}
			if !l.dispatchKnob(knob, v) {
				slog.Debug("Received knob rotate message", "knob", knob, "value", value, "message", message)
			}
		case Touch:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			b := touchCoordToButton(x, y)
			if !l.dispatchTouch(b, ButtonDown, x, y) {
				slog.Debug("Received touch message", "x", x, "y", y, "id", id, "b", b, "message", message)
			}
		case TouchEnd:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			b := touchCoordToButton(x, y)
			if !l.dispatchTouch(b, ButtonUp, x, y) {
				slog.Debug("Received touch end message", "x", x, "y", y, "id", id, "b", b, "message", message)
			}
		case TouchCT:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			slog.Debug("Received CT touch message (unhandled)", "x", x, "y", y, "id", id, "message", message)
		case TouchEndCT:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			slog.Debug("Received CT touch end message (unhandled)", "x", x, "y", y, "id", id, "message", message)
		default:
			slog.Info("Received unknown message", "message", m.String())
		}
	}
}
