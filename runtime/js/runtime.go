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

func OpenRuntime(ctx context.Context, env *envpkg.LoupeDeckEnvironment, opts ...engine.Option) (*Runtime, error) {
	env = envpkg.Ensure(env)
	builder := engine.NewBuilder(opts...).
		WithModules(NewRegistrar(env)).
		// Raw Loupedeck scene commands normally install only the Loupedeck
		// runtime modules. When the Loupedeck provider is used from a generated
		// xgoja binary, additional domain modules may also have registered
		// themselves in go-go-goja's default module registry. Select the optional
		// AST analysis and filesystem module names here so scripts run through
		// `deck run` can compose Loupedeck UI modules with go-ast-analysis and
		// file-backed browser remote-control state without exposing the whole host
		// default registry.
		UseModuleMiddleware(engine.MiddlewareOnly("go-ast-analysis", "ast", "fs", "node:fs"))
	factory, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build loupedeck runtime factory: %w", err)
	}
	rt, err := factory.NewRuntime(engine.WithStartupContext(ctx), engine.WithLifetimeContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("create loupedeck runtime: %w", err)
	}
	return &Runtime{Runtime: rt, Env: env}, nil
}

func NewRuntime(env *envpkg.LoupeDeckEnvironment) *Runtime {
	rt, err := OpenRuntime(context.Background(), env)
	if err != nil {
		panic(err)
	}
	return rt
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
