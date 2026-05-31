package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
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

func TestRegisterScenesCommandProvider(t *testing.T) {
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	provider, ok := registry.ResolveCommandSetProvider(PackageID, "scenes")
	if !ok {
		t.Fatalf("missing command provider %s.scenes", PackageID)
	}
	if provider.DefaultMount != "loupedeck" {
		t.Fatalf("default mount = %q, want loupedeck", provider.DefaultMount)
	}
}

func TestRegisterProviderHelpSource(t *testing.T) {
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	source, ok := registry.ResolveHelpSource(PackageID, "runtime-api")
	if !ok {
		t.Fatalf("missing help source %s.runtime-api", PackageID)
	}
	if source.Root != "." || source.FS == nil {
		t.Fatalf("unexpected help source: %#v", source)
	}
	data, err := fs.ReadFile(source.FS, "topics/01-loupedeck-js-api-reference.md")
	if err != nil {
		t.Fatalf("read api reference: %v", err)
	}
	if !bytes.Contains(data, []byte("Slug: loupedeck-js-api-reference")) {
		t.Fatalf("expected embedded API reference slug, got %q", data[:min(len(data), 200)])
	}
}

func TestScenesCommandProviderBuildsRunCommand(t *testing.T) {
	provider := resolveCommandProvider(t, "scenes")
	set, err := provider.New(providerapi.CommandSetContext{
		Context:        context.Background(),
		PackageID:      PackageID,
		Name:           "scenes",
		Mount:          "loupe",
		RuntimeProfile: "main",
	})
	if err != nil {
		t.Fatalf("create command set: %v", err)
	}
	if set == nil || len(set.Commands) == 0 {
		t.Fatalf("expected commands")
	}
	if !hasTopLevelCommand(set, "run") {
		t.Fatalf("expected top-level run command in default scenes command set")
	}
}

func TestScenesCommandProviderCanDisableRunCommand(t *testing.T) {
	provider := resolveCommandProvider(t, "scenes")
	config, err := json.Marshal(map[string]any{"includeRun": false})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	set, err := provider.New(providerapi.CommandSetContext{
		Context:        context.Background(),
		PackageID:      PackageID,
		Name:           "scenes",
		Mount:          "loupe",
		RuntimeProfile: "main",
		Config:         config,
	})
	if err != nil {
		t.Fatalf("create command set: %v", err)
	}
	if set == nil {
		t.Fatalf("expected command set")
	}
	if hasTopLevelCommand(set, "run") {
		t.Fatalf("did not expect top-level run command when includeRun=false")
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

func resolveCommandProvider(t *testing.T, name string) providerapi.CommandSetProvider {
	t.Helper()
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	provider, ok := registry.ResolveCommandSetProvider(PackageID, name)
	if !ok {
		t.Fatalf("missing command provider %s.%s", PackageID, name)
	}
	return provider
}

func hasTopLevelCommand(set *providerapi.CommandSet, name string) bool {
	if set == nil {
		return false
	}
	for _, command := range set.Commands {
		if command == nil || command.Description() == nil {
			continue
		}
		desc := command.Description()
		if desc.Name == name && len(desc.Parents) == 0 {
			return true
		}
	}
	return false
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
