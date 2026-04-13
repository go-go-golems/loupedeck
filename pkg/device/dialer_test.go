package device

import (
	"os"
	"path/filepath"
	"testing"

	"go.bug.st/serial/enumerator"
)

func TestLookupPortDetailsForPathMatchesExactName(t *testing.T) {
	ports := []*enumerator.PortDetails{{Name: "/dev/ttyACM0", VID: "2ec2", PID: "0004"}}
	port := lookupPortDetailsForPath("/dev/ttyACM0", ports)
	if port == nil {
		t.Fatalf("expected exact path match")
	}
	if port.PID != "0004" {
		t.Fatalf("unexpected PID %q", port.PID)
	}
}

func TestLookupPortDetailsForPathMatchesSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "ttyACM0")
	alias := filepath.Join(tmpDir, "by-id-device")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Symlink(target, alias); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	ports := []*enumerator.PortDetails{{Name: target, VID: "2ec2", PID: "0004"}}
	port := lookupPortDetailsForPath(alias, ports)
	if port == nil {
		t.Fatalf("expected symlink path match")
	}
	if port.PID != "0004" {
		t.Fatalf("unexpected PID %q", port.PID)
	}
}

func TestLookupPortDetailsForPathReturnsNilWhenMissing(t *testing.T) {
	ports := []*enumerator.PortDetails{{Name: "/dev/ttyACM0", VID: "2ec2", PID: "0004"}}
	if port := lookupPortDetailsForPath("/dev/ttyUSB9", ports); port != nil {
		t.Fatalf("expected no match, got %+v", port)
	}
}
