package module_metrics

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/jsmetrics"
)

const ModuleName = "loupedeck/metrics"

func Register(registry *require.Registry) {
	jsmetrics.RegisterLowLevelModuleAs(registry, ModuleName)
}
