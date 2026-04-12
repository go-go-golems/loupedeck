package env

import (
	"github.com/go-go-golems/loupedeck/runtime/host"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

type Environment struct {
	Reactive *reactive.Runtime
	UI       *ui.UI
	Host     *host.Runtime
}

func Ensure(e *Environment) *Environment {
	if e == nil {
		e = &Environment{}
	}
	if e.Host != nil && e.UI == nil {
		e.UI = e.Host.UI
	}
	if e.UI != nil && e.Reactive == nil {
		e.Reactive = e.UI.Reactive
	}
	if e.Reactive == nil {
		e.Reactive = reactive.NewRuntime()
	}
	if e.UI == nil {
		e.UI = ui.New(e.Reactive)
	}
	if e.Host == nil {
		e.Host = host.New(e.UI)
	}
	return e
}
