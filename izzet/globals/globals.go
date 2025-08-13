package globals

import (
	"github.com/kkevinchou/izzet/internal/metrics"
)

var (
	clientRegistry *metrics.Registry
	serverRegistry *metrics.Registry
)

func init() {
	clientRegistry = metrics.NewRegistry(nil)
	serverRegistry = metrics.NewRegistry(nil)
}

func ClientRegistry() *metrics.Registry {
	return clientRegistry
}

func ServerRegistry() *metrics.Registry {
	return serverRegistry
}
