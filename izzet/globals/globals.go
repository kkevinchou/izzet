package globals

import (
	"github.com/kkevinchou/izzet/internal/metrics"
)

var (
	clientRegistry *metrics.Registry
	serverRegistry *metrics.Registry
)

func init() {
	clientRegistry = metrics.NewRegistry(10, nil)
	serverRegistry = metrics.NewRegistry(10, nil)
}

func ClientRegistry() *metrics.Registry {
	return clientRegistry
}

func ServerRegistry() *metrics.Registry {
	return serverRegistry
}
