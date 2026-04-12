package js

import (
	"context"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/pkg/runtimeowner"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/js/module_anim"
	"github.com/go-go-golems/loupedeck/runtime/js/module_easing"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
	"github.com/go-go-golems/loupedeck/runtime/js/module_state"
	"github.com/go-go-golems/loupedeck/runtime/js/module_ui"
)

type Runtime struct {
	VM    *goja.Runtime
	Loop  *eventloop.EventLoop
	Owner runtimeowner.Runner
	Env   *envpkg.Environment

	runtimeCtx       context.Context
	runtimeCtxCancel context.CancelFunc
	closeOnce        sync.Once
}

func NewRuntime(env *envpkg.Environment) *Runtime {
	env = envpkg.Ensure(env)
	vm := goja.New()
	loop := eventloop.NewEventLoop()
	go loop.Start()

	registry := new(require.Registry)
	module_state.Register(registry)
	module_ui.Register(registry)
	module_easing.Register(registry)
	module_anim.Register(registry)
	module_gfx.Register(registry)
	registry.Enable(vm)

	owner := runtimeowner.NewRunner(vm, loop, runtimeowner.Options{
		Name:          "loupedeck-js-runtime",
		RecoverPanics: true,
	})
	ctx, cancel := context.WithCancel(context.Background())
	runtimebridge.Store(vm, runtimebridge.Bindings{
		Context: ctx,
		Loop:    loop,
		Owner:   owner,
		Values: map[string]any{
			envpkg.BindingKeyEnvironment: env,
		},
	})

	return &Runtime{
		VM:               vm,
		Loop:             loop,
		Owner:            owner,
		Env:              env,
		runtimeCtx:       ctx,
		runtimeCtxCancel: cancel,
	}
}

func (r *Runtime) Context() context.Context {
	if r == nil || r.runtimeCtx == nil {
		return context.Background()
	}
	return r.runtimeCtx
}

func (r *Runtime) RunString(ctx context.Context, src string) (goja.Value, error) {
	if r == nil {
		return nil, nil
	}
	result, err := r.Owner.Call(ctx, "vm.run-string", func(_ context.Context, vm *goja.Runtime) (any, error) {
		return vm.RunString(src)
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return goja.Undefined(), nil
	}
	if value, ok := result.(goja.Value); ok {
		return value, nil
	}
	return r.VM.ToValue(result), nil
}

func (r *Runtime) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	var err error
	r.closeOnce.Do(func() {
		if r.runtimeCtxCancel != nil {
			r.runtimeCtxCancel()
		}
		if r.VM != nil {
			runtimebridge.Delete(r.VM)
		}
		if r.Owner != nil {
			err = r.Owner.Shutdown(ctx)
		}
		if r.Loop != nil {
			r.Loop.Stop()
		}
	})
	return err
}
