package module_hw

import (
	"context"
	"image/color"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/loupedeck/pkg/device"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

type mockControl struct {
	brightness []int
	colors     []mockColorCall
}

type mockColorCall struct {
	button device.Button
	color  color.RGBA
}

func (m *mockControl) SetButtonColor(button device.Button, col color.RGBA) error {
	m.colors = append(m.colors, mockColorCall{button: button, color: col})
	return nil
}

func (m *mockControl) SetBrightness(brightness int) error {
	m.brightness = append(m.brightness, brightness)
	return nil
}

type testModuleSpec struct {
	env *envpkg.LoupeDeckEnvironment
}

func (s testModuleSpec) ID() string { return "test:loupedeck-hw" }
func (s testModuleSpec) RegisterRuntimeModule(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	envpkg.Store(ctx.VM, s.env)
	Register(reg)
	return ctx.AddCloser(func(context.Context) error {
		envpkg.Delete(ctx.VM)
		return nil
	})
}

func newTestRuntime(t *testing.T, env *envpkg.LoupeDeckEnvironment) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewBuilder(
		engine.WithImplicitDefaultRegistryModules(false),
		engine.WithDataOnlyDefaultRegistryModules(false),
	).WithModules(testModuleSpec{env: env}).Build()
	if err != nil {
		t.Fatalf("build runtime factory: %v", err)
	}
	rt, err := factory.NewRuntime(engine.WithStartupContext(context.Background()), engine.WithLifetimeContext(context.Background()))
	if err != nil {
		t.Fatalf("create runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	return rt
}

func runJS(t *testing.T, rt *engine.Runtime, src string) error {
	t.Helper()
	_, err := rt.Owner.Call(context.Background(), "test.run", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, err := vm.RunString(src)
		return nil, err
	})
	return err
}

func TestHardwareControlSuccess(t *testing.T) {
	control := &mockControl{}
	rt := newTestRuntime(t, &envpkg.LoupeDeckEnvironment{DeviceControl: control})

	err := runJS(t, rt, `
		const hw = require("loupedeck/hw")
		hw.setBrightness(7)
		hw.setButtonColor("Button1", { r: 1, g: 2, b: 3 })
		hw.setButtonColor("Button2", "#0a0b0c")
	`)
	if err != nil {
		t.Fatalf("run js: %v", err)
	}
	if len(control.brightness) != 1 || control.brightness[0] != 7 {
		t.Fatalf("brightness calls = %#v", control.brightness)
	}
	if len(control.colors) != 2 {
		t.Fatalf("color call count = %d", len(control.colors))
	}
	if control.colors[0].button != device.Button1 || control.colors[0].color != (color.RGBA{R: 1, G: 2, B: 3, A: 255}) {
		t.Fatalf("first color call = %#v", control.colors[0])
	}
	if control.colors[1].button != device.Button2 || control.colors[1].color != (color.RGBA{R: 10, G: 11, B: 12, A: 255}) {
		t.Fatalf("second color call = %#v", control.colors[1])
	}
}

func TestHardwareControlUnavailable(t *testing.T) {
	rt := newTestRuntime(t, &envpkg.LoupeDeckEnvironment{})
	err := runJS(t, rt, `require("loupedeck/hw").setBrightness(5)`)
	if err == nil || !strings.Contains(err.Error(), "hardware control is not available") {
		t.Fatalf("expected unavailable hardware error, got %v", err)
	}
}

func TestHardwareControlValidation(t *testing.T) {
	control := &mockControl{}
	rt := newTestRuntime(t, &envpkg.LoupeDeckEnvironment{DeviceControl: control})
	cases := []struct {
		name string
		src  string
		want string
	}{
		{name: "brightness range", src: `require("loupedeck/hw").setBrightness(11)`, want: "0..10"},
		{name: "unknown button", src: `require("loupedeck/hw").setButtonColor("Nope", {r:1,g:2,b:3})`, want: "unknown button"},
		{name: "rgb range", src: `require("loupedeck/hw").setButtonColor("Button1", {r:300,g:2,b:3})`, want: "color.r must be in range"},
		{name: "bad hex", src: `require("loupedeck/hw").setButtonColor("Button1", "#xyz")`, want: "hex color"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runJS(t, rt, tc.src)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}
