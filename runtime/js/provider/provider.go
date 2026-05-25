package provider

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	"github.com/go-go-golems/loupedeck/runtime/js/module_easing"
	"github.com/go-go-golems/loupedeck/runtime/js/module_gfx"
)

const PackageID = "loupedeck"

func Register(registry *providerapi.Registry) error {
	return registry.Package(PackageID,
		moduleEntry(module_easing.ModuleName, "Easing functions for loupedeck animation curves.", module_easing.Loader),
		moduleEntry(module_gfx.ModuleName, "Offscreen drawing surfaces, colors, text, and font helpers.", module_gfx.Loader),
	)
}

func moduleEntry(name, description string, loader func() require.ModuleLoader) providerapi.Module {
	return providerapi.Module{
		Name:        name,
		DefaultAs:   name,
		Description: description,
		New: func(providerapi.ModuleContext) (require.ModuleLoader, error) {
			return loader(), nil
		},
	}
}
