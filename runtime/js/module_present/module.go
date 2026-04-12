package module_present

import (
	"context"
	"errors"
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/pkg/runtimeowner"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

const ModuleName = "loupedeck/present"

func Register(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, func(runtime *goja.Runtime, module *goja.Object) {
		bindings, ok := runtimebridge.Lookup(runtime)
		if !ok || bindings.Owner == nil {
			panic(runtime.NewGoError(fmt.Errorf("present module requires runtime owner bindings")))
		}
		env, ok := envpkg.Lookup(runtime)
		if !ok || env == nil || env.Present == nil {
			panic(runtime.NewGoError(fmt.Errorf("present module requires environment bindings")))
		}
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("invalidate", func(call goja.FunctionCall) goja.Value {
			reason := call.Argument(0).String()
			env.Present.Invalidate(reason)
			return goja.Undefined()
		})
		_ = exports.Set("onFrame", func(call goja.FunctionCall) goja.Value {
			fn, ok := goja.AssertFunction(call.Argument(0))
			if !ok {
				panic(runtime.NewTypeError("present.onFrame requires a function"))
			}
			ownerCtx := runtimeowner.OwnerContext(bindings.Owner, bindings.Context)
			env.Present.SetRenderFunc(func(reason string) error {
				_, err := bindings.Owner.Call(ownerCtx, "present.onFrame", func(_ context.Context, vm *goja.Runtime) (any, error) {
					_, err := fn(goja.Undefined(), vm.ToValue(reason))
					return nil, err
				})
				if err != nil {
					if errors.Is(err, runtimeowner.ErrClosed) || errors.Is(err, runtimeowner.ErrScheduleRejected) || errors.Is(err, runtimeowner.ErrCanceled) {
						return nil
					}
					return err
				}
				return nil
			})
			return goja.Undefined()
		})
	})
}
