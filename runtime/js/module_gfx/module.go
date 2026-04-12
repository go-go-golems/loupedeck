package module_gfx

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/runtime/gfx"
	"golang.org/x/image/font/basicfont"
)

const ModuleName = "loupedeck/gfx"

func Register(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, func(runtime *goja.Runtime, module *goja.Object) {
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("surface", func(call goja.FunctionCall) goja.Value {
			width := int(call.Argument(0).ToInteger())
			height := int(call.Argument(1).ToInteger())
			surface := gfx.NewSurface(width, height)
			return surfaceObject(runtime, surface)
		})
	})
}

func surfaceObject(runtime *goja.Runtime, surface *gfx.Surface) goja.Value {
	obj := runtime.NewObject()
	_ = obj.Set("width", func(goja.FunctionCall) goja.Value {
		return runtime.ToValue(surface.Width())
	})
	_ = obj.Set("height", func(goja.FunctionCall) goja.Value {
		return runtime.ToValue(surface.Height())
	})
	_ = obj.Set("clear", func(call goja.FunctionCall) goja.Value {
		surface.Clear(uint8(call.Argument(0).ToInteger()))
		return goja.Undefined()
	})
	_ = obj.Set("fillRect", func(call goja.FunctionCall) goja.Value {
		surface.FillRect(
			int(call.Argument(0).ToInteger()),
			int(call.Argument(1).ToInteger()),
			int(call.Argument(2).ToInteger()),
			int(call.Argument(3).ToInteger()),
			uint8(call.Argument(4).ToInteger()),
		)
		return goja.Undefined()
	})
	_ = obj.Set("line", func(call goja.FunctionCall) goja.Value {
		surface.Line(
			int(call.Argument(0).ToInteger()),
			int(call.Argument(1).ToInteger()),
			int(call.Argument(2).ToInteger()),
			int(call.Argument(3).ToInteger()),
			uint8(call.Argument(4).ToInteger()),
		)
		return goja.Undefined()
	})
	_ = obj.Set("crosshatch", func(call goja.FunctionCall) goja.Value {
		surface.Crosshatch(
			int(call.Argument(0).ToInteger()),
			int(call.Argument(1).ToInteger()),
			int(call.Argument(2).ToInteger()),
			int(call.Argument(3).ToInteger()),
			int(call.Argument(4).ToInteger()),
			uint8(call.Argument(5).ToInteger()),
		)
		return goja.Undefined()
	})
	_ = obj.Set("text", func(call goja.FunctionCall) goja.Value {
		text := call.Argument(0).String()
		opts := textOptionsFromValue(runtime, call.Argument(1))
		surface.Text(text, opts)
		return goja.Undefined()
	})
	_ = obj.Set("compositeAdd", func(call goja.FunctionCall) goja.Value {
		other := SurfaceFromValue(call.Argument(0), runtime)
		surface.CompositeAdd(other, int(call.Argument(1).ToInteger()), int(call.Argument(2).ToInteger()))
		return goja.Undefined()
	})
	_ = obj.Set("at", func(call goja.FunctionCall) goja.Value {
		return runtime.ToValue(surface.At(int(call.Argument(0).ToInteger()), int(call.Argument(1).ToInteger())))
	})
	_ = obj.Set("__surface", surface)
	return obj
}

func SurfaceFromValue(value goja.Value, runtime *goja.Runtime) *gfx.Surface {
	obj := value.ToObject(runtime)
	exported := obj.Get("__surface").Export()
	surface, ok := exported.(*gfx.Surface)
	if !ok || surface == nil {
		panic(runtime.NewTypeError("expected gfx surface"))
	}
	return surface
}

func textOptionsFromValue(runtime *goja.Runtime, value goja.Value) gfx.TextOptions {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return gfx.TextOptions{Face: basicfont.Face7x13}
	}
	obj := value.ToObject(runtime)
	return gfx.TextOptions{
		X:          intProp(obj, "x"),
		Y:          intProp(obj, "y"),
		Width:      intProp(obj, "width"),
		Height:     intProp(obj, "height"),
		Brightness: uint8(intProp(obj, "brightness")),
		Center:     boolProp(obj, "center"),
		Face:       basicfont.Face7x13,
	}
}

func intProp(obj *goja.Object, name string) int {
	value := obj.Get(name)
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return 0
	}
	return int(value.ToInteger())
}

func boolProp(obj *goja.Object, name string) bool {
	value := obj.Get(name)
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return false
	}
	return value.ToBoolean()
}
