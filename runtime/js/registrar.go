package js

import (
	"context"
	"fmt"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/loupedeck/pkg/jsmetrics"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/js/module_anim"
	"github.com/go-go-golems/loupedeck/runtime/js/module_easing"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
	"github.com/go-go-golems/loupedeck/runtime/js/module_present"
	"github.com/go-go-golems/loupedeck/runtime/js/module_state"
	"github.com/go-go-golems/loupedeck/runtime/js/module_ui"
)

type Registrar struct {
	Env *envpkg.LoupeDeckEnvironment
}

func NewRegistrar(env *envpkg.LoupeDeckEnvironment) Registrar {
	return Registrar{Env: envpkg.Ensure(env)}
}

func (r Registrar) ID() string {
	return "loupedeck-runtime"
}

func (r Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	if ctx == nil {
		return fmt.Errorf("runtime module context is nil")
	}
	if reg == nil {
		return fmt.Errorf("require registry is nil")
	}
	env := envpkg.Ensure(r.Env)
	envpkg.Store(ctx.VM, env)
	ctx.SetValue("environment", env)
	if err := ctx.AddCloser(func(context.Context) error {
		envpkg.Delete(ctx.VM)
		return nil
	}); err != nil {
		return fmt.Errorf("register environment cleanup: %w", err)
	}

	module_state.Register(reg)
	module_ui.Register(reg)
	module_easing.Register(reg)
	module_anim.Register(reg)
	module_gfx.Register(reg)
	module_present.Register(reg)
	jsmetrics.RegisterModules(reg, "loupedeck")
	return nil
}
