package provider

import (
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	jsprovider "github.com/go-go-golems/loupedeck/runtime/js/provider"
)

const PackageID = jsprovider.PackageID

func Register(registry *providerapi.ProviderRegistry) error {
	return jsprovider.Register(registry)
}
