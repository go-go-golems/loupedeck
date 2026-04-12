package module_scene_metrics

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/loupedeck/pkg/jsmetrics"
)

const ModuleName = "loupedeck/scene-metrics"

func Register(registry *require.Registry) {
	jsmetrics.RegisterSceneModuleAs(registry, ModuleName)
}
