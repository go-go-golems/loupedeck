package run

import (
	"testing"

	"github.com/go-go-golems/loupedeck/pkg/device"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
)

func TestAttachDeckToEnvironmentSetsDeviceControl(t *testing.T) {
	env := envpkg.Ensure(&envpkg.LoupeDeckEnvironment{Metrics: metrics.NewWithTraceLimit(10)})
	deck := &device.Loupedeck{}

	attachDeckToEnvironment(env, deck)

	control, ok := env.DeviceControl.(*envpkg.LoupedeckDeviceControl)
	if !ok || control == nil {
		t.Fatalf("expected LoupedeckDeviceControl, got %#v", env.DeviceControl)
	}
	if control.Deck != deck {
		t.Fatal("expected DeviceControl to reference the attached deck")
	}
}
