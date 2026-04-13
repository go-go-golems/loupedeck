package device

import (
	//"github.com/tarm/serial"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// SerialWebSockConn implements an external dialer interface for the
// Gorilla that allows it to talk to Loupedeck's weird
// websockets-over-serial-over-USB setup.
//
// The Gorilla websockets library can use an external dialer
// interface, which means that we can use it *mostly* unmodified to
// talk to a serial device instead of a network device.  We just need
// to provide something that matches the net.Conn interface.  Here's a
// minimal implementation.
type SerialWebSockConn struct {
	Name            string
	Port            serial.Port
	Vendor, Product string
}

// Read reads bytes from the connected serial port.
func (l *SerialWebSockConn) Read(b []byte) (int, error) {
	//slog.Info("Reading", "limit_bytes", len(b))
	n, err := l.Port.Read(b)
	//slog.Info("Read", "bytes", n, "err", err, "data", fmt.Sprintf("%v", b[:n]))
	return n, err
}

// Write sends bytes to the connected serial port.
func (l *SerialWebSockConn) Write(b []byte) (int, error) {
	//slog.Info("Writing", "bytes", len(b), "message", fmt.Sprintf("%v", b))
	return l.Port.Write(b)
}

// Close closes the underlying serial port connection.
func (l *SerialWebSockConn) Close() error {
	if l == nil || l.Port == nil {
		return nil
	}
	return l.Port.Close()
}

// LocalAddr is needed for Gorilla compatibility, but doesn't actually
// make sense with serial ports.
func (l *SerialWebSockConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (l *SerialWebSockConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (l *SerialWebSockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (l *SerialWebSockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (l *SerialWebSockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// ConnectSerialAuto connects to the first compatible Loupedeck in the
// system.  To connect to a specific Loupedeck, use ConnectSerialPath.
func ConnectSerialAuto() (*SerialWebSockConn, error) {
	slog.Info("Enumerating ports")

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no serial ports found")
	}

	for _, port := range ports {
		slog.Info("Trying to open port", "port", port.Name)
		if port.IsUSB && (port.VID == "2ec2" || port.VID == "1532") {
			p, err := serial.Open(port.Name, &serial.Mode{})
			if err != nil {
				return nil, fmt.Errorf("unable to open port %q", port.Name)
			}
			conn := &SerialWebSockConn{
				Name:    port.Name,
				Port:    p,
				Vendor:  port.VID,
				Product: port.PID,
			}
			return conn, nil
		}
	}

	return nil, fmt.Errorf("no Loupedeck devices found")
}

func sameSerialPath(a, b string) bool {
	if a == b {
		return true
	}
	ac, aerr := filepath.EvalSymlinks(a)
	bc, berr := filepath.EvalSymlinks(b)
	if aerr == nil && berr == nil {
		return ac == bc
	}
	return false
}

func lookupPortDetailsForPath(serialPath string, ports []*enumerator.PortDetails) *enumerator.PortDetails {
	for _, port := range ports {
		if port == nil {
			continue
		}
		if sameSerialPath(serialPath, port.Name) {
			return port
		}
	}
	return nil
}

func lookupSerialPortMetadata(serialPath string) (string, string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", "", err
	}
	port := lookupPortDetailsForPath(serialPath, ports)
	if port == nil {
		return "", "", nil
	}
	return port.VID, port.PID, nil
}

// ConnectSerialPath connects to a specific Loupedeck, using the path
// to the USB serial device as a key.
func ConnectSerialPath(serialPath string) (*SerialWebSockConn, error) {
	vendor, product, metaErr := lookupSerialPortMetadata(serialPath)
	if metaErr != nil {
		slog.Warn("Unable to resolve serial metadata for path", "path", serialPath, "err", metaErr)
	}

	p, err := serial.Open(serialPath, &serial.Mode{})
	if err != nil {
		return nil, fmt.Errorf("unable to open serial device %q", serialPath)
	}
	conn := &SerialWebSockConn{
		Name:    serialPath,
		Port:    p,
		Vendor:  vendor,
		Product: product,
	}

	return conn, nil
}
