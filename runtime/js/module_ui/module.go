package module_ui

import (
	"context"
	"fmt"
	"image/color"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/pkg/runtimebridge"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
	deck "github.com/go-go-golems/loupedeck/pkg/device"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

const ModuleName = "loupedeck/ui"

func Register(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, func(runtime *goja.Runtime, module *goja.Object) {
		bindings, ok := runtimebridge.Lookup(runtime)
		if !ok || bindings.Owner == nil {
			panic(runtime.NewGoError(fmt.Errorf("ui module requires runtime owner bindings")))
		}
		env, ok := envpkg.Lookup(runtime)
		if !ok || env == nil {
			panic(runtime.NewGoError(fmt.Errorf("ui module requires environment bindings")))
		}
		ownerCtx := runtimeowner.OwnerContext(bindings.Owner, bindings.Context)
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("page", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			page := env.UI.AddPage(name)
			obj := pageObject(bindings, ownerCtx, runtime, env, page)
			if fn, ok := goja.AssertFunction(call.Argument(1)); ok {
				if _, err := fn(goja.Undefined(), obj); err != nil {
					panic(runtime.NewGoError(err))
				}
			}
			return obj
		})
		_ = exports.Set("show", func(call goja.FunctionCall) goja.Value {
			if err := env.Host.Show(call.Argument(0).String()); err != nil {
				panic(runtime.NewGoError(err))
			}
			return goja.Undefined()
		})
		_ = exports.Set("onButton", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			button, err := deck.ParseButton(name)
			if err != nil {
				panic(runtime.NewTypeError(err.Error()))
			}
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("ui.onButton requires a function"))
			}
			sub := env.Host.OnButton(button, func(b deck.Button, status deck.ButtonStatus) {
				_ = bindings.Owner.Post(bindings.Context, "ui.onButton.callback", func(_ context.Context, vm *goja.Runtime) {
					event := vm.NewObject()
					_ = event.Set("name", name)
					_ = event.Set("status", status.String())
					_, err := fn(goja.Undefined(), event)
					if err != nil {
						panic(vm.NewGoError(err))
					}
				})
			})
			return subscriptionObject(runtime, sub)
		})
		_ = exports.Set("onTouch", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			touch, err := deck.ParseTouchButton(name)
			if err != nil {
				panic(runtime.NewTypeError(err.Error()))
			}
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("ui.onTouch requires a function"))
			}
			sub := env.Host.OnTouch(touch, func(_ deck.TouchButton, status deck.ButtonStatus, x, y uint16) {
				_ = bindings.Owner.Post(bindings.Context, "ui.onTouch.callback", func(_ context.Context, vm *goja.Runtime) {
					event := vm.NewObject()
					_ = event.Set("name", name)
					_ = event.Set("status", status.String())
					_ = event.Set("x", x)
					_ = event.Set("y", y)
					_, err := fn(goja.Undefined(), event)
					if err != nil {
						panic(vm.NewGoError(err))
					}
				})
			})
			return subscriptionObject(runtime, sub)
		})
		_ = exports.Set("onKnob", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			knob, err := deck.ParseKnob(name)
			if err != nil {
				panic(runtime.NewTypeError(err.Error()))
			}
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("ui.onKnob requires a function"))
			}
			sub := env.Host.OnKnob(knob, func(_ deck.Knob, value int) {
				_ = bindings.Owner.Post(bindings.Context, "ui.onKnob.callback", func(_ context.Context, vm *goja.Runtime) {
					event := vm.NewObject()
					_ = event.Set("name", name)
					_ = event.Set("value", value)
					_, err := fn(goja.Undefined(), event)
					if err != nil {
						panic(vm.NewGoError(err))
					}
				})
			})
			return subscriptionObject(runtime, sub)
		})
	})
}

func pageObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, env *envpkg.LoupeDeckEnvironment, page *ui.Page) *goja.Object {
	obj := runtime.NewObject()
	_ = obj.Set("tile", func(call goja.FunctionCall) goja.Value {
		col := int(call.Argument(0).ToInteger())
		row := int(call.Argument(1).ToInteger())
		tile := page.AddTile(col, row)
		tileObj := tileObject(bindings, ownerCtx, runtime, env, tile)
		if fn, ok := goja.AssertFunction(call.Argument(2)); ok {
			if _, err := fn(goja.Undefined(), tileObj); err != nil {
				panic(runtime.NewGoError(err))
			}
		}
		return tileObj
	})
	_ = obj.Set("display", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		display := page.AddDisplay(name)
		displayObj := displayObject(bindings, ownerCtx, runtime, env, display)
		if fn, ok := goja.AssertFunction(call.Argument(1)); ok {
			if _, err := fn(goja.Undefined(), displayObj); err != nil {
				panic(runtime.NewGoError(err))
			}
		}
		return displayObj
	})
	return obj
}

func displayObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, env *envpkg.LoupeDeckEnvironment, display *ui.Display) *goja.Object {
	obj := runtime.NewObject()
	_ = obj.Set("text", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			display.BindText(func() string {
				result, err := bindings.Owner.Call(ownerCtx, "ui.display.text", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return stringify(value), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result.(string)
			})
		} else {
			display.SetText(stringify(call.Argument(0)))
		}
		return goja.Undefined()
	})
	_ = obj.Set("icon", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			display.BindIcon(func() string {
				result, err := bindings.Owner.Call(ownerCtx, "ui.display.icon", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return stringify(value), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result.(string)
			})
		} else {
			display.SetIcon(stringify(call.Argument(0)))
		}
		return goja.Undefined()
	})
	_ = obj.Set("visible", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			display.BindVisible(func() bool {
				result, err := bindings.Owner.Call(ownerCtx, "ui.display.visible", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return value.ToBoolean(), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result.(bool)
			})
		} else {
			display.SetVisible(call.Argument(0).ToBoolean())
		}
		return goja.Undefined()
	})
	_ = obj.Set("surface", func(call goja.FunctionCall) goja.Value {
		arg := call.Argument(0)
		if goja.IsNull(arg) || goja.IsUndefined(arg) {
			display.SetSurface(nil)
		} else {
			display.SetSurface(module_gfx.SurfaceFromValue(arg, runtime))
		}
		return goja.Undefined()
	})
	_ = obj.Set("layer", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		arg := call.Argument(1)
		opts := layerOptionsFromValue(call.Argument(2), runtime)
		if goja.IsNull(arg) || goja.IsUndefined(arg) {
			display.SetLayerWithOptions(name, nil, opts)
		} else {
			display.SetLayerWithOptions(name, module_gfx.SurfaceFromValue(arg, runtime), opts)
		}
		return goja.Undefined()
	})
	_ = obj.Set("tile", func(call goja.FunctionCall) goja.Value {
		col := int(call.Argument(0).ToInteger())
		row := int(call.Argument(1).ToInteger())
		tile := display.AddTile(col, row)
		tileObj := tileObject(bindings, ownerCtx, runtime, env, tile)
		if fn, ok := goja.AssertFunction(call.Argument(2)); ok {
			if _, err := fn(goja.Undefined(), tileObj); err != nil {
				panic(runtime.NewGoError(err))
			}
		}
		return tileObj
	})
	return obj
}

func tileObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, _ *envpkg.LoupeDeckEnvironment, tile *ui.Tile) *goja.Object {
	obj := runtime.NewObject()
	_ = obj.Set("text", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			tile.BindText(func() string {
				result, err := bindings.Owner.Call(ownerCtx, "ui.tile.text", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return stringify(value), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result.(string)
			})
		} else {
			tile.SetText(stringify(call.Argument(0)))
		}
		return goja.Undefined()
	})
	_ = obj.Set("icon", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			tile.BindIcon(func() string {
				result, err := bindings.Owner.Call(ownerCtx, "ui.tile.icon", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return stringify(value), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result.(string)
			})
		} else {
			tile.SetIcon(stringify(call.Argument(0)))
		}
		return goja.Undefined()
	})
	_ = obj.Set("visible", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			tile.BindVisible(func() bool {
				result, err := bindings.Owner.Call(ownerCtx, "ui.tile.visible", func(_ context.Context, vm *goja.Runtime) (any, error) {
					value, err := fn(goja.Undefined())
					if err != nil {
						return nil, err
					}
					return value.ToBoolean(), nil
				})
				if err != nil {
					panic(runtime.NewGoError(err))
				}
				return result.(bool)
			})
		} else {
			tile.SetVisible(call.Argument(0).ToBoolean())
		}
		return goja.Undefined()
	})
	_ = obj.Set("surface", func(call goja.FunctionCall) goja.Value {
		arg := call.Argument(0)
		if goja.IsNull(arg) || goja.IsUndefined(arg) {
			tile.SetSurface(nil)
		} else {
			tile.SetSurface(module_gfx.SurfaceFromValue(arg, runtime))
		}
		return goja.Undefined()
	})
	return obj
}

func subscriptionObject(runtime *goja.Runtime, sub Subscription) goja.Value {
	obj := runtime.NewObject()
	_ = obj.Set("close", func(goja.FunctionCall) goja.Value {
		_ = sub.Close()
		return goja.Undefined()
	})
	return obj
}

type Subscription interface {
	Close() error
}

func stringify(value goja.Value) string {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	if s, ok := value.Export().(string); ok {
		return s
	}
	return fmt.Sprint(value.Export())
}

func layerOptionsFromValue(value goja.Value, runtime *goja.Runtime) ui.LayerOptions {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return ui.LayerOptions{}
	}
	obj := value.ToObject(runtime)
	rValue := obj.Get("r")
	gValue := obj.Get("g")
	bValue := obj.Get("b")
	if (rValue == nil || goja.IsUndefined(rValue) || goja.IsNull(rValue)) &&
		(gValue == nil || goja.IsUndefined(gValue) || goja.IsNull(gValue)) &&
		(bValue == nil || goja.IsUndefined(bValue) || goja.IsNull(bValue)) {
		return ui.LayerOptions{}
	}
	a := uint8(255)
	if av := obj.Get("a"); av != nil && !goja.IsUndefined(av) && !goja.IsNull(av) {
		a = uint8(av.ToInteger())
	}
	return ui.LayerOptions{Foreground: color.RGBA{
		R: uint8(rValue.ToInteger()),
		G: uint8(gValue.ToInteger()),
		B: uint8(bValue.ToInteger()),
		A: a,
	}}
}
