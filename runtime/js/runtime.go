package js

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
)

type Runtime struct {
	*engine.Runtime
	Env *envpkg.LoupeDeckEnvironment
}

func NewRuntime(env *envpkg.LoupeDeckEnvironment) *Runtime {
	env = envpkg.Ensure(env)
	factory, err := engine.NewBuilder().
		WithRuntimeModuleRegistrars(NewRegistrar(env)).
		Build()
	if err != nil {
		panic(fmt.Errorf("build loupedeck runtime factory: %w", err))
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		panic(fmt.Errorf("create loupedeck runtime: %w", err))
	}
	return &Runtime{
		Runtime: rt,
		Env:     env,
	}
}

func (r *Runtime) RunString(ctx context.Context, src string) (goja.Value, error) {
	if r == nil || r.Runtime == nil {
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
