package env

import (
	"github.com/dop251/goja"
	"github.com/go-go-golems/loupedeck/runtime/anim"
	"github.com/go-go-golems/loupedeck/runtime/host"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
	"github.com/go-go-golems/loupedeck/runtime/present"
	"github.com/go-go-golems/loupedeck/runtime/reactive"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

type LoupeDeckEnvironment struct {
	Reactive *reactive.Runtime
	UI       *ui.UI
	Host     *host.Runtime
	Anim     *anim.Runtime
	Present  *present.Runtime
	Metrics  *metrics.Collector
}

func Lookup(vm *goja.Runtime) (*LoupeDeckEnvironment, bool) {
	if vm == nil {
		return nil, false
	}
	value, ok := environmentsByVM.Load(vm)
	if !ok {
		return nil, false
	}
	env, ok := value.(*LoupeDeckEnvironment)
	return env, ok && env != nil
}

func Ensure(e *LoupeDeckEnvironment) *LoupeDeckEnvironment {
	if e == nil {
		e = &LoupeDeckEnvironment{}
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
	if e.Anim == nil {
		e.Anim = anim.New(e.Host)
	}
	if e.Present == nil {
		e.Present = present.New()
	}
	if e.Metrics == nil {
		e.Metrics = metrics.New()
	}
	return e
}
