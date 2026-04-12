package runtimebridge

import (
	"context"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/go-go-golems/loupedeck/pkg/runtimeowner"
)

// Bindings expose runtime-owned scheduling primitives and runtime-scoped values
// for modules that need async owner-thread settlement or access to host services.
type Bindings struct {
	Context context.Context
	Loop    *eventloop.EventLoop
	Owner   runtimeowner.Runner
	Values  map[string]any
}

var bindingsByVM sync.Map

func Store(vm *goja.Runtime, bindings Bindings) {
	if vm == nil {
		return
	}
	bindingsByVM.Store(vm, bindings)
}

func Lookup(vm *goja.Runtime) (Bindings, bool) {
	if vm == nil {
		return Bindings{}, false
	}
	value, ok := bindingsByVM.Load(vm)
	if !ok {
		return Bindings{}, false
	}
	bindings, ok := value.(Bindings)
	if !ok {
		return Bindings{}, false
	}
	return bindings, true
}

func Delete(vm *goja.Runtime) {
	if vm == nil {
		return
	}
	bindingsByVM.Delete(vm)
}
