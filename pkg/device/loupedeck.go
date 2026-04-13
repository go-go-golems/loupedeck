/*
   Copyright 2021 Google LLC

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       https://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package loupedeck provides a Go interface for talking to a
// Loupedeck Live control surface.
//
// The Loupedeck Live with firmware 1.x appeared as a USB network
// device that we talked to via HTTP+websockets, but newer firmware
// looks like a serial device that talks a mutant version of the
// Websocket protocol.
//
// See https://github.com/foxxyz/loupedeck for Javascript code for
// talking to the Loupedeck Live; it supports more of the Loupedeck's
// functionality.
package device

import (
	"fmt"
	"image/color"
	"strings"
	"sync"
)

type transactionCallback func(m *Message)

// Loupedeck describes a Loupedeck device.
type Loupedeck struct {
	Vendor               string
	Product              string
	Model                string
	Version              string
	SerialNo             string
	serial               *SerialWebSockConn
	conn                 wsConn
	writer               *outboundWriter
	renderer             *renderScheduler
	writerOptions        WriterOptions
	renderOptions        RenderOptions
	buttonListeners      map[Button]map[uint64]ButtonFunc
	buttonUpListeners    map[Button]map[uint64]ButtonFunc
	knobListeners        map[Knob]map[uint64]KnobFunc
	touchListeners       map[TouchButton]map[uint64]TouchFunc
	touchUpListeners     map[TouchButton]map[uint64]TouchFunc
	listenerMutex        sync.RWMutex
	listenerID           uint64
	transactionID        uint8
	transactionMutex     sync.Mutex
	transactionCallbacks map[byte]transactionCallback
	displays             map[string]*Display
}

// Close closes the connection to the Loupedeck.
func (l *Loupedeck) Close() error {
	var errs []string
	if l.renderer != nil {
		l.renderer.Close()
	}
	if l.writer != nil {
		l.writer.Close()
	}
	if l.conn != nil {
		if err := l.conn.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("websocket close: %v", err))
		}
	}
	if l.serial != nil {
		if err := l.serial.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("serial close: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}
	return nil
}

// WriterStats returns a snapshot of outbound writer metrics.
func (l *Loupedeck) WriterStats() WriterStats {
	if l.writer == nil {
		return WriterStats{}
	}
	return l.writer.Stats()
}

// RenderStats returns a snapshot of render invalidation/coalescing metrics.
func (l *Loupedeck) RenderStats() RenderStats {
	if l.renderer == nil {
		return RenderStats{}
	}
	return l.renderer.Stats()
}

// SetBrightness sets the overall brightness of the Loupedeck display.
func (l *Loupedeck) SetBrightness(b int) error {
	data := make([]byte, 1)
	data[0] = byte(b)
	m := l.NewMessage(SetBrightness, data)
	return l.Send(m)
}

// SetButtonColor sets the color of a specific Button.  The
// Loupedeck Live allows the 8 buttons below the display to be set to
// specific colors, however the 'Circle' button's colors may be
// overridden to show the status of the Loupedeck Live's connection to
// the host.
func (l *Loupedeck) SetButtonColor(b Button, c color.RGBA) error {
	data := make([]byte, 4)
	data[0] = byte(b)
	data[1] = c.R
	data[2] = c.G
	data[3] = c.B
	m := l.NewMessage(SetColor, data)
	return l.Send(m)
}
