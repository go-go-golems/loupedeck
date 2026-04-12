package module_easing

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/runtime/easing"
)

const ModuleName = "loupedeck/easing"

func Register(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, func(runtime *goja.Runtime, module *goja.Object) {
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("linear", easingFunc(runtime, easing.Linear))
		_ = exports.Set("inOutQuad", easingFunc(runtime, easing.InOutQuad))
		_ = exports.Set("inOutCubic", easingFunc(runtime, easing.InOutCubic))
		_ = exports.Set("outBack", easingFunc(runtime, easing.OutBack))
		_ = exports.Set("steps", func(call goja.FunctionCall) goja.Value {
			return runtime.ToValue(easingFunc(runtime, easing.Steps(int(call.Argument(0).ToInteger()))))
		})
	})
}

func easingFunc(runtime *goja.Runtime, fn easing.Func) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		return runtime.ToValue(fn(call.Argument(0).ToFloat()))
	}
}
