package env

import (
	"sync"

	"github.com/dop251/goja"
)

var environmentsByVM sync.Map

func Store(vm *goja.Runtime, env *LoupeDeckEnvironment) {
	if vm == nil || env == nil {
		return
	}
	environmentsByVM.Store(vm, env)
}

func Delete(vm *goja.Runtime) {
	if vm == nil {
		return
	}
	environmentsByVM.Delete(vm)
}
