package provider

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	"github.com/go-go-golems/loupedeck/runtime/js/module_easing"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
)

func TestRegisterProvider(t *testing.T) {
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	for _, name := range []string{module_easing.ModuleName, module_gfx.ModuleName} {
		mod, ok := registry.ResolveModule(PackageID, name)
		if !ok {
			t.Fatalf("missing module %s.%s", PackageID, name)
		}
		if mod.DefaultAs != name {
			t.Fatalf("default alias for %s = %q, want %q", name, mod.DefaultAs, name)
		}
	}
}

func TestEasingLoaderInstallsExports(t *testing.T) {
	mod := resolveModule(t, module_easing.ModuleName)
	exports := loadModule(t, mod)
	linear, ok := goja.AssertFunction(exports.Get("linear"))
	if !ok {
		t.Fatalf("linear export is not a function")
	}
	ret, err := linear(goja.Undefined(), goja.New().ToValue(0.5))
	if err != nil {
		t.Fatalf("call linear: %v", err)
	}
	if got := ret.ToFloat(); got != 0.5 {
		t.Fatalf("linear(0.5) = %v, want 0.5", got)
	}
}

func TestGfxLoaderInstallsExports(t *testing.T) {
	mod := resolveModule(t, module_gfx.ModuleName)
	exports := loadModule(t, mod)
	if _, ok := goja.AssertFunction(exports.Get("surface")); !ok {
		t.Fatalf("surface export is not a function")
	}
	if _, ok := goja.AssertFunction(exports.Get("font")); !ok {
		t.Fatalf("font export is not a function")
	}
}

func resolveModule(t *testing.T, name string) providerapi.Module {
	t.Helper()
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	mod, ok := registry.ResolveModule(PackageID, name)
	if !ok {
		t.Fatalf("missing module %s.%s", PackageID, name)
	}
	return mod
}

func loadModule(t *testing.T, mod providerapi.Module) *goja.Object {
	t.Helper()
	loader, err := mod.New(providerapi.ModuleContext{Name: mod.Name, As: mod.DefaultAs})
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}
	vm := goja.New()
	moduleObj := vm.NewObject()
	exports := vm.NewObject()
	if err := moduleObj.Set("exports", exports); err != nil {
		t.Fatalf("set exports: %v", err)
	}
	loader(vm, moduleObj)
	return exports
}
