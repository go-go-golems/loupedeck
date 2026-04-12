package module_anim

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/pkg/runtimeowner"
	"github.com/go-go-golems/loupedeck/runtime/easing"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

const ModuleName = "loupedeck/anim"

func Register(registry *require.Registry, env *envpkg.Environment) {
	registry.RegisterNativeModule(ModuleName, func(runtime *goja.Runtime, module *goja.Object) {
		bindings, ok := runtimebridge.Lookup(runtime)
		if !ok || bindings.Owner == nil {
			panic(runtime.NewGoError(fmt.Errorf("anim module requires runtime owner bindings")))
		}
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("to", func(call goja.FunctionCall) goja.Value {
			ownerCtx := runtimeowner.OwnerContext(bindings.Owner, bindings.Context)
			get, set := numericTarget(bindings, ownerCtx, runtime, call.Argument(0))
			to := call.Argument(1).ToFloat()
			duration := time.Duration(call.Argument(2).ToInteger()) * time.Millisecond
			ease := easingFromArg(bindings, ownerCtx, runtime, call.Argument(3))
			h := env.Anim.TweenFloat64(get, set, to, duration, ease)
			return handleObject(runtime, h)
		})
		_ = exports.Set("loop", func(call goja.FunctionCall) goja.Value {
			duration := time.Duration(call.Argument(0).ToInteger()) * time.Millisecond
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("anim.loop requires a function"))
			}
			h := env.Anim.Loop(duration, func(v float64) {
				_ = bindings.Owner.Post(bindings.Context, "anim.loop.callback", func(_ context.Context, vm *goja.Runtime) {
					_, err := fn(goja.Undefined(), vm.ToValue(v))
					if err != nil {
						panic(vm.NewGoError(err))
					}
				})
			})
			return handleObject(runtime, h)
		})
		_ = exports.Set("timeline", func(goja.FunctionCall) goja.Value {
			timeline := env.Anim.Timeline()
			obj := runtime.NewObject()
			_ = obj.Set("to", func(call goja.FunctionCall) goja.Value {
				ownerCtx := runtimeowner.OwnerContext(bindings.Owner, bindings.Context)
				get, set := numericTarget(bindings, ownerCtx, runtime, call.Argument(0))
				to := call.Argument(1).ToFloat()
				duration := time.Duration(call.Argument(2).ToInteger()) * time.Millisecond
				ease := easingFromArg(bindings, ownerCtx, runtime, call.Argument(3))
				timeline.To(get, set, to, duration, ease)
				return obj
			})
			_ = obj.Set("play", func(goja.FunctionCall) goja.Value {
				return handleObject(runtime, timeline.Play())
			})
			return obj
		})
	})
}

func numericTarget(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, value goja.Value) (func() float64, func(float64)) {
	obj := value.ToObject(runtime)
	getValue, ok := goja.AssertFunction(obj.Get("get"))
	if !ok {
		panic(runtime.NewTypeError("animation target must expose get()"))
	}
	setValue, ok := goja.AssertFunction(obj.Get("set"))
	if !ok {
		panic(runtime.NewTypeError("animation target must expose set()"))
	}
	get := func() float64 {
		result, err := bindings.Owner.Call(ownerCtx, "anim.target.get", func(_ context.Context, vm *goja.Runtime) (any, error) {
			v, err := getValue(obj)
			if err != nil {
				return nil, err
			}
			return v.ToFloat(), nil
		})
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		return result.(float64)
	}
	set := func(v float64) {
		_, err := bindings.Owner.Call(ownerCtx, "anim.target.set", func(_ context.Context, vm *goja.Runtime) (any, error) {
			_, err := setValue(obj, runtime.ToValue(v))
			return nil, err
		})
		if err != nil {
			panic(runtime.NewGoError(err))
		}
	}
	return get, set
}

func easingFromArg(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, value goja.Value) easing.Func {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return easing.Linear
	}
	fn, ok := goja.AssertFunction(value)
	if !ok {
		return easing.Linear
	}
	return func(t float64) float64 {
		result, err := bindings.Owner.Call(ownerCtx, "anim.easing", func(_ context.Context, vm *goja.Runtime) (any, error) {
			result, err := fn(goja.Undefined(), vm.ToValue(t))
			if err != nil {
				return nil, err
			}
			return result.ToFloat(), nil
		})
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		return result.(float64)
	}
}

func handleObject(runtime *goja.Runtime, handle interface{ Stop() }) goja.Value {
	obj := runtime.NewObject()
	_ = obj.Set("stop", func(goja.FunctionCall) goja.Value {
		handle.Stop()
		return goja.Undefined()
	})
	return obj
}
