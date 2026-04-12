package jsmetrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/runtimebridge"
	"github.com/go-go-golems/loupedeck/runtime/metrics"
)

const BindingKeyCollector = "metricsCollector"

func Lookup(vm *goja.Runtime) (*metrics.Collector, bool) {
	bindings, ok := runtimebridge.Lookup(vm)
	if !ok || bindings.Values == nil {
		return nil, false
	}
	collector, ok := bindings.Values[BindingKeyCollector].(*metrics.Collector)
	return collector, ok && collector != nil
}

func RegisterModules(registry *require.Registry, prefix string) {
	RegisterLowLevelModuleAs(registry, moduleName(prefix, "metrics"))
	RegisterSceneModuleAs(registry, moduleName(prefix, "scene-metrics"))
}

func RegisterLowLevelModuleAs(registry *require.Registry, name string) {
	registry.RegisterNativeModule(name, func(runtime *goja.Runtime, module *goja.Object) {
		collector, ok := Lookup(runtime)
		if !ok {
			panic(runtime.NewGoError(fmt.Errorf("metrics module requires collector binding")))
		}
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("now", func(goja.FunctionCall) goja.Value {
			return runtime.ToValue(nowMillis())
		})
		_ = exports.Set("inc", func(call goja.FunctionCall) goja.Value {
			metric := call.Argument(0).String()
			delta := int64(1)
			if arg := call.Argument(1); !goja.IsUndefined(arg) && !goja.IsNull(arg) {
				delta = arg.ToInteger()
			}
			collector.Inc(metric, delta)
			return goja.Undefined()
		})
		_ = exports.Set("observeMillis", func(call goja.FunctionCall) goja.Value {
			collector.ObserveMillis(call.Argument(0).String(), call.Argument(1).ToFloat())
			return goja.Undefined()
		})
		_ = exports.Set("trace", func(call goja.FunctionCall) goja.Value {
			collector.Trace(call.Argument(0).String(), fieldsFromArg(runtime, call.Argument(1)))
			return goja.Undefined()
		})
		_ = exports.Set("time", func(call goja.FunctionCall) goja.Value {
			metric := call.Argument(0).String()
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("metrics.time requires a function"))
			}
			start := nowMillis()
			result, err := fn(goja.Undefined())
			collector.ObserveMillis(metric, nowMillis()-start)
			if err != nil {
				panic(runtime.NewGoError(err))
			}
			return result
		})
		_ = exports.Set("counted", func(call goja.FunctionCall) goja.Value {
			metric := call.Argument(0).String()
			fn, ok := goja.AssertFunction(call.Argument(1))
			if !ok {
				panic(runtime.NewTypeError("metrics.counted requires a function"))
			}
			collector.Inc(metric, 1)
			result, err := fn(goja.Undefined())
			if err != nil {
				panic(runtime.NewGoError(err))
			}
			return result
		})
	})
}

func RegisterSceneModuleAs(registry *require.Registry, name string) {
	registry.RegisterNativeModule(name, func(runtime *goja.Runtime, module *goja.Object) {
		collector, ok := Lookup(runtime)
		if !ok {
			panic(runtime.NewGoError(fmt.Errorf("scene-metrics module requires collector binding")))
		}
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("reasonCategory", func(call goja.FunctionCall) goja.Value {
			return runtime.ToValue(reasonCategory(call.Argument(0).String()))
		})
		_ = exports.Set("create", func(call goja.FunctionCall) goja.Value {
			prefix := call.Argument(0).String()
			if prefix == "" {
				prefix = "scene"
			}
			return helperObject(runtime, collector, prefix)
		})
	})
}

func helperObject(runtime *goja.Runtime, collector *metrics.Collector, prefix string) goja.Value {
	obj := runtime.NewObject()
	metricName := func(suffix string) string {
		if suffix == "" {
			return prefix
		}
		return prefix + "." + suffix
	}
	_ = obj.Set("prefix", func(goja.FunctionCall) goja.Value {
		return runtime.ToValue(prefix)
	})
	_ = obj.Set("reasonCategory", func(call goja.FunctionCall) goja.Value {
		return runtime.ToValue(reasonCategory(call.Argument(0).String()))
	})
	_ = obj.Set("inc", func(call goja.FunctionCall) goja.Value {
		suffix := call.Argument(0).String()
		delta := int64(1)
		if arg := call.Argument(1); !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			delta = arg.ToInteger()
		}
		collector.Inc(metricName(suffix), delta)
		return goja.Undefined()
	})
	_ = obj.Set("observeMillis", func(call goja.FunctionCall) goja.Value {
		collector.ObserveMillis(metricName(call.Argument(0).String()), call.Argument(1).ToFloat())
		return goja.Undefined()
	})
	_ = obj.Set("trace", func(call goja.FunctionCall) goja.Value {
		collector.Trace(metricName(call.Argument(0).String()), fieldsFromArg(runtime, call.Argument(1)))
		return goja.Undefined()
	})
	_ = obj.Set("time", func(call goja.FunctionCall) goja.Value {
		suffix := call.Argument(0).String()
		fn, ok := goja.AssertFunction(call.Argument(1))
		if !ok {
			panic(runtime.NewTypeError("sceneMetrics.time requires a function"))
		}
		start := nowMillis()
		result, err := fn(goja.Undefined())
		collector.ObserveMillis(metricName(suffix), nowMillis()-start)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		return result
	})
	_ = obj.Set("timeTile", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		fn, ok := goja.AssertFunction(call.Argument(1))
		if !ok {
			panic(runtime.NewTypeError("sceneMetrics.timeTile requires a function"))
		}
		start := nowMillis()
		result, err := fn(goja.Undefined())
		collector.ObserveMillis(metricName("tile."+name), nowMillis()-start)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		return result
	})
	_ = obj.Set("recordLoopTick", func(goja.FunctionCall) goja.Value {
		collector.Inc(metricName("loopTicks"), 1)
		return goja.Undefined()
	})
	_ = obj.Set("recordActivation", func(call goja.FunctionCall) goja.Value {
		reason := call.Argument(0).String()
		category := reasonCategory(reason)
		collector.Inc(metricName("activations"), 1)
		collector.Inc(metricName("activations."+category), 1)
		return goja.Undefined()
	})
	_ = obj.Set("recordRebuild", func(call goja.FunctionCall) goja.Value {
		reason := call.Argument(0).String()
		if reason == "" {
			reason = "unknown"
		}
		category := reasonCategory(reason)
		collector.Inc(metricName("renderAll.calls"), 1)
		collector.Inc(metricName("renderAll.reason."+category), 1)
		collector.Inc(metricName("renderAll.reasonExact."+reason), 1)
		arg := call.Argument(1)
		if goja.IsUndefined(arg) || goja.IsNull(arg) {
			return goja.Undefined()
		}
		fn, ok := goja.AssertFunction(arg)
		if !ok {
			panic(runtime.NewTypeError("sceneMetrics.recordRebuild second argument must be a function"))
		}
		start := nowMillis()
		result, err := fn(goja.Undefined())
		collector.ObserveMillis(metricName("renderAll"), nowMillis()-start)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		return result
	})
	return obj
}

func fieldsFromArg(runtime *goja.Runtime, value goja.Value) map[string]string {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	obj := value.ToObject(runtime)
	if obj == nil {
		return nil
	}
	keys := obj.Keys()
	if len(keys) == 0 {
		return nil
	}
	fields := make(map[string]string, len(keys))
	for _, key := range keys {
		v := obj.Get(key)
		switch {
		case goja.IsUndefined(v):
			fields[key] = "undefined"
		case goja.IsNull(v):
			fields[key] = "null"
		default:
			fields[key] = v.String()
		}
	}
	return fields
}

func reasonCategory(reason string) string {
	switch {
	case reason == "":
		return "unknown"
	case reason == "loop":
		return "loop"
	case reason == "initial":
		return "initial"
	case len(reason) > 0 && reason[0] == 'T':
		return "touch"
	case len(reason) > 0 && reason[0] == 'B':
		return "button"
	default:
		return "other"
	}
}

func moduleName(prefix, suffix string) string {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		return suffix
	}
	return prefix + "/" + suffix
}

func nowMillis() float64 {
	return float64(time.Now().UnixNano()) / 1e6
}
