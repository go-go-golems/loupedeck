package module_hw

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/device"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

const ModuleName = "loupedeck/hw"

var hexColorPattern = regexp.MustCompile(`^#?([0-9a-fA-F]{6})$`)

func Loader() require.ModuleLoader {
	return func(runtime *goja.Runtime, module *goja.Object) {
		env, ok := envpkg.Lookup(runtime)
		if !ok || env == nil {
			panic(runtime.NewGoError(fmt.Errorf("hw module requires environment services")))
		}
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("setBrightness", func(call goja.FunctionCall) goja.Value {
			control := requireDeviceControl(runtime, env)
			value := int(call.Argument(0).ToInteger())
			if value < 0 || value > 10 {
				panic(runtime.NewTypeError("setBrightness requires value in range 0..10"))
			}
			if err := control.SetBrightness(value); err != nil {
				panic(runtime.NewGoError(fmt.Errorf("setBrightness: %w", err)))
			}
			return goja.Undefined()
		})
		_ = exports.Set("setButtonColor", func(call goja.FunctionCall) goja.Value {
			control := requireDeviceControl(runtime, env)
			buttonName := call.Argument(0).String()
			button, err := device.ParseButton(buttonName)
			if err != nil {
				panic(runtime.NewTypeError(err.Error()))
			}
			col, err := parseColor(runtime, call.Argument(1))
			if err != nil {
				panic(runtime.NewTypeError(err.Error()))
			}
			if err := control.SetButtonColor(button, col); err != nil {
				panic(runtime.NewGoError(fmt.Errorf("setButtonColor %s: %w", buttonName, err)))
			}
			return goja.Undefined()
		})
	}
}

func Register(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, Loader())
}

func requireDeviceControl(runtime *goja.Runtime, env *envpkg.LoupeDeckEnvironment) envpkg.DeviceControl {
	if env == nil || env.DeviceControl == nil {
		panic(runtime.NewGoError(fmt.Errorf("loupedeck hardware control is not available; start with --deck-enabled and a connected device")))
	}
	return env.DeviceControl
}

func parseColor(runtime *goja.Runtime, value goja.Value) (color.RGBA, error) {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return color.RGBA{}, fmt.Errorf("setButtonColor requires a color object {r,g,b} or '#rrggbb' string")
	}
	if s, ok := value.Export().(string); ok {
		return parseHexColor(s)
	}
	obj := value.ToObject(runtime)
	r, err := uint8Prop(obj, "r")
	if err != nil {
		return color.RGBA{}, err
	}
	g, err := uint8Prop(obj, "g")
	if err != nil {
		return color.RGBA{}, err
	}
	b, err := uint8Prop(obj, "b")
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	match := hexColorPattern.FindStringSubmatch(s)
	if match == nil {
		return color.RGBA{}, fmt.Errorf("hex color must be '#rrggbb' or 'rrggbb'")
	}
	r, _ := strconv.ParseUint(match[1][0:2], 16, 8)
	g, _ := strconv.ParseUint(match[1][2:4], 16, 8)
	b, _ := strconv.ParseUint(match[1][4:6], 16, 8)
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

func uint8Prop(obj *goja.Object, name string) (uint8, error) {
	value := obj.Get(name)
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return 0, fmt.Errorf("color.%s is required", name)
	}
	switch value.Export().(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// accepted below via ToInteger()
	default:
		return 0, fmt.Errorf("color.%s must be a number in range 0..255", name)
	}
	integer := value.ToInteger()
	if integer < 0 || integer > 255 {
		return 0, fmt.Errorf("color.%s must be in range 0..255", name)
	}
	return uint8(integer), nil
}
