package device

import (
	"encoding/binary"
	"fmt"

	"github.com/gorilla/websocket"
)

// Listen waits for events from the Loupedeck and calls
// callbacks as configured. It returns when the read loop exits.
func (l *Loupedeck) Listen() error {
	for {
		websocketMsgType, message, err := l.conn.ReadMessage()

		if err != nil {
			log.Warn().Err(err).Msg("Read error, exiting")
			return fmt.Errorf("websocket read failed: %w", err)
		}

		if len(message) == 0 {
			log.Warn().Msg("Received a 0-byte message.  Skipping")
			continue
		}

		if websocketMsgType != websocket.BinaryMessage {
			log.Warn().Int("type", websocketMsgType).Msg("Unknown websocket message type received")
		}

		m, _ := l.ParseMessage(message)
		log.Debug().Str("message", m.String()).Msg("read")

		if m.transactionID != 0 {
			if c := l.takeTransactionCallback(m.transactionID); c != nil {
				log.Debug().Uint8("transaction_id", m.transactionID).Msg("dispatching transaction callback")
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
				log.Debug().Str("button", fmt.Sprintf("%v", button)).Str("upDown", fmt.Sprintf("%v", upDown)).Msg("received uncaught button press message")
			}
		case KnobRotate:
			knob := Knob(binary.BigEndian.Uint16(message[2:]))
			value := int(message[4])
			v := value
			if value == 255 {
				v = -1
			}
			if !l.dispatchKnob(knob, v) {
				log.Debug().Str("knob", fmt.Sprintf("%v", knob)).Int("value", value).Msg("Received knob rotate message")
			}
		case Touch:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			b := touchCoordToButton(x, y)
			if !l.dispatchTouch(b, ButtonDown, x, y) {
				log.Debug().Uint16("x", x).Uint16("y", y).Uint8("id", id).Msg("Received touch message")
			}
		case TouchEnd:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			b := touchCoordToButton(x, y)
			if !l.dispatchTouch(b, ButtonUp, x, y) {
				log.Debug().Uint16("x", x).Uint16("y", y).Uint8("id", id).Msg("Received touch end message")
			}
		case TouchCT:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			log.Debug().Uint16("x", x).Uint16("y", y).Uint8("id", id).Msg("Received CT touch message (unhandled)")
		case TouchEndCT:
			x := binary.BigEndian.Uint16(message[4:])
			y := binary.BigEndian.Uint16(message[6:])
			id := message[8]
			log.Debug().Uint16("x", x).Uint16("y", y).Uint8("id", id).Msg("Received CT touch end message (unhandled)")
		default:
			log.Debug().Str("message", m.String()).Msg("received unknown message")
		}
	}
}
