package module_state

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/pkg/runtimebridge"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

const ModuleName = "loupedeck/state"

func Loader() require.ModuleLoader {
	return func(runtime *goja.Runtime, module *goja.Object) {
		runtimeServices, ok := runtimebridge.Lookup(runtime)
		if !ok || runtimeServices.Owner == nil {
			panic(runtime.NewGoError(fmt.Errorf("state module requires runtime services")))
		}
		env, ok := envpkg.Lookup(runtime)
		if !ok || env == nil {
			panic(runtime.NewGoError(fmt.Errorf("state module requires environment services")))
		}
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("signal", func(call goja.FunctionCall) goja.Value {
			initial := exportValue(call.Argument(0))
			sig := reactive.NewSignal(env.Reactive, initial)
			return signalObject(runtimeServices, runtime, sig)
		})
		_ = exports.Set("computed", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("state.computed requires a function"))
			}
			cmp := reactive.NewComputed(env.Reactive, func() any {
				result, err := runtimeServices.CallWithCurrentContext(runtime, "state.computed", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return exportValue(value), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result
			})
			obj := runtime.NewObject()
			_ = obj.Set("get", func(goja.FunctionCall) goja.Value {
				return runtime.ToValue(cmp.Get())
			})
			return obj
		})
		_ = exports.Set("batch", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("state.batch requires a function"))
			}
			env.Reactive.Batch(func() {
				_, err := runtimeServices.CallWithCurrentContext(runtime, "state.batch", func(_ context.Context, vm *goja.Runtime) (any, error) {
					_, err := fn(goja.Undefined())
					return nil, err
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
			})
			return goja.Undefined()
		})
		_ = exports.Set("watch", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("state.watch requires a function"))
			}
			sub := env.Reactive.Watch(func() {
				_, err := runtimeServices.CallWithCurrentContext(runtime, "state.watch", func(_ context.Context, vm *goja.Runtime) (any, error) {
					_, err := fn(goja.Undefined())
					return nil, err
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
			})
			obj := runtime.NewObject()
			_ = obj.Set("stop", func(goja.FunctionCall) goja.Value {
				sub.Stop()
				return goja.Undefined()
			})
			return obj
		})
	}
}

func Register(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, Loader())
}

func signalObject(runtimeServices runtimebridge.RuntimeServices, runtime *goja.Runtime, sig *reactive.Signal[any]) goja.Value {
	obj := runtime.NewObject()
	_ = obj.Set("get", func(goja.FunctionCall) goja.Value {
		return runtime.ToValue(sig.Get())
	})
	_ = obj.Set("set", func(call goja.FunctionCall) goja.Value {
		sig.Set(exportValue(call.Argument(0)))
		return goja.Undefined()
	})
	_ = obj.Set("update", func(call goja.FunctionCall) goja.Value {
		fn, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			panic(runtime.NewTypeError("signal.update requires a function"))
		}
		sig.Update(func(current any) any {
			result, err := runtimeServices.CallWithCurrentContext(runtime, "signal.update", func(_ context.Context, vm *goja.Runtime) (any, error) {
				value, err := fn(goja.Undefined(), vm.ToValue(current))
				if err != nil {
					return nil, err
				}
				return exportValue(value), nil
			})
			if err != nil {
				panic(runtime.NewGoError(err))
			}
			return result
		})
		return goja.Undefined()
	})
	return obj
}

func exportValue(value goja.Value) any {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	return value.Export()
}
