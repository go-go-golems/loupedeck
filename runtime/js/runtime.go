package js

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/js/module_state"
	"github.com/go-go-golems/loupedeck/runtime/js/module_ui"
)

func NewRuntime(env *envpkg.Environment) (*goja.Runtime, *envpkg.Environment) {
	env = envpkg.Ensure(env)
	vm := goja.New()
	registry := new(require.Registry)
	module_state.Register(registry, env)
	module_ui.Register(registry, env)
	registry.Enable(vm)
	return vm, env
}
