package module_state

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/pkg/runtimeowner"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
)

const ModuleName = "loupedeck/state"

func Register(registry *require.Registry, env *envpkg.Environment) {
	registry.RegisterNativeModule(ModuleName, func(runtime *goja.Runtime, module *goja.Object) {
		bindings, ok := runtimebridge.Lookup(runtime)
		if !ok || bindings.Owner == nil {
			panic(runtime.NewGoError(fmt.Errorf("state module requires runtime owner bindings")))
		}
		ownerCtx := runtimeowner.OwnerContext(bindings.Owner, bindings.Context)
		exports := module.Get("exports").(*goja.Object)
		exports.Set("signal", func(call goja.FunctionCall) goja.Value {
			initial := exportValue(call.Argument(0))
			sig := reactive.NewSignal(env.Reactive, initial)
			return signalObject(bindings, ownerCtx, runtime, sig)
		})
		exports.Set("computed", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("state.computed requires a function"))
			}
			cmp := reactive.NewComputed(env.Reactive, func() any {
				result, err := bindings.Owner.Call(ownerCtx, "state.computed", func(_ context.Context, vm *goja.Runtime) (any, error) {
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
		exports.Set("batch", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("state.batch requires a function"))
			}
			env.Reactive.Batch(func() {
				_, err := bindings.Owner.Call(ownerCtx, "state.batch", func(_ context.Context, vm *goja.Runtime) (any, error) {
					_, err := fn(goja.Undefined())
					return nil, err
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
			})
			return goja.Undefined()
		})
		exports.Set("watch", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("state.watch requires a function"))
			}
			sub := env.Reactive.Watch(func() {
				_, err := bindings.Owner.Call(ownerCtx, "state.watch", func(_ context.Context, vm *goja.Runtime) (any, error) {
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
	})
}

func signalObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, sig *reactive.Signal[any]) goja.Value {
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
			result, err := bindings.Owner.Call(ownerCtx, "signal.update", func(_ context.Context, vm *goja.Runtime) (any, error) {
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
