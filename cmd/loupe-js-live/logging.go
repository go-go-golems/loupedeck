package main

import (
	"log/slog"

	"github.com/go-go-golems/loupedeck/pkg/device"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

func registerEventLogging(env *envpkg.Environment) {
	if env == nil {
		return
	}
	for _, button := range []device.Button{
		device.Circle,
		device.Button1,
		device.Button2,
		device.Button3,
		device.Button4,
		device.Button5,
		device.Button6,
		device.Button7,
	} {
		button := button
		env.Host.OnButton(button, func(b device.Button, s device.ButtonStatus) {
			slog.Info("button event", "button", b.String(), "status", s.String())
		})
	}
	for _, touch := range []device.TouchButton{
		device.Touch1,
		device.Touch2,
		device.Touch3,
		device.Touch4,
		device.Touch5,
		device.Touch6,
		device.Touch7,
		device.Touch8,
		device.Touch9,
		device.Touch10,
		device.Touch11,
		device.Touch12,
	} {
		touch := touch
		env.Host.OnTouch(touch, func(t device.TouchButton, s device.ButtonStatus, x, y uint16) {
			slog.Info("touch event", "touch", t.String(), "status", s.String(), "x", x, "y", y)
		})
	}
	for _, knob := range []device.Knob{
		device.Knob1,
		device.Knob2,
		device.Knob3,
		device.Knob4,
		device.Knob5,
		device.Knob6,
	} {
		knob := knob
		env.Host.OnKnob(knob, func(k device.Knob, value int) {
			slog.Info("knob event", "knob", k.String(), "value", value)
		})
	}
}
