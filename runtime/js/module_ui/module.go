package module_ui

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	deck "github.com/go-go-golems/loupedeck"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/pkg/runtimeowner"
	envpkg "github.com/go-go-golems/loupedeck/runtime/js/env"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
	"github.com/go-go-golems/loupedeck/runtime/ui"
)

const ModuleName = "loupedeck/ui"

var (
	buttons = map[string]deck.Button{
		"Circle":  deck.Circle,
		"Button1": deck.Button1,
		"Button2": deck.Button2,
		"Button3": deck.Button3,
		"Button4": deck.Button4,
		"Button5": deck.Button5,
		"Button6": deck.Button6,
		"Button7": deck.Button7,
	}
	touches = map[string]deck.TouchButton{
		"Touch1":  deck.Touch1,
		"Touch2":  deck.Touch2,
		"Touch3":  deck.Touch3,
		"Touch4":  deck.Touch4,
		"Touch5":  deck.Touch5,
		"Touch6":  deck.Touch6,
		"Touch7":  deck.Touch7,
		"Touch8":  deck.Touch8,
		"Touch9":  deck.Touch9,
		"Touch10": deck.Touch10,
		"Touch11": deck.Touch11,
		"Touch12": deck.Touch12,
	}
	knobs = map[string]deck.Knob{
		"Knob1": deck.Knob1,
		"Knob2": deck.Knob2,
		"Knob3": deck.Knob3,
		"Knob4": deck.Knob4,
		"Knob5": deck.Knob5,
		"Knob6": deck.Knob6,
	}
)

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
		exports.Set("page", func(call goja.FunctionCall) goja.Value {
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
		exports.Set("show", func(call goja.FunctionCall) goja.Value {
			if err := env.Host.Show(call.Argument(0).String()); err != nil {
				panic(runtime.NewGoError(err))
			}
			return goja.Undefined()
		})
		exports.Set("onButton", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			button, ok := buttons[name]
			if !ok {
				panic(runtime.NewTypeError("unknown button %q", name))
			}
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("ui.onButton requires a function"))
			}
			sub := env.Host.OnButton(button, func(b deck.Button, status deck.ButtonStatus) {
				_ = bindings.Owner.Post(bindings.Context, "ui.onButton.callback", func(_ context.Context, vm *goja.Runtime) {
					event := vm.NewObject()
					_ = event.Set("name", name)
					_ = event.Set("status", statusString(status))
					_, err := fn(goja.Undefined(), event)
					if err != nil {
						panic(vm.NewGoError(err))
					}
				})
			})
			return subscriptionObject(runtime, sub)
		})
		exports.Set("onTouch", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			touch, ok := touches[name]
			if !ok {
				panic(runtime.NewTypeError("unknown touch %q", name))
			}
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("ui.onTouch requires a function"))
			}
			sub := env.Host.OnTouch(touch, func(_ deck.TouchButton, status deck.ButtonStatus, x, y uint16) {
				_ = bindings.Owner.Post(bindings.Context, "ui.onTouch.callback", func(_ context.Context, vm *goja.Runtime) {
					event := vm.NewObject()
					_ = event.Set("name", name)
					_ = event.Set("status", statusString(status))
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
		exports.Set("onKnob", func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			knob, ok := knobs[name]
			if !ok {
				panic(runtime.NewTypeError("unknown knob %q", name))
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

func pageObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, env *envpkg.Environment, page *ui.Page) *goja.Object {
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

func displayObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, env *envpkg.Environment, display *ui.Display) *goja.Object {
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
		display.SetSurface(module_gfx.SurfaceFromValue(call.Argument(0), runtime))
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

func tileObject(bindings runtimebridge.Bindings, ownerCtx context.Context, runtime *goja.Runtime, _ *envpkg.Environment, tile *ui.Tile) *goja.Object {
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

func statusString(status deck.ButtonStatus) string {
	if status == deck.ButtonUp {
		return "up"
	}
	return "down"
}
