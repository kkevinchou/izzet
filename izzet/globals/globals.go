package globals

import (
	"github.com/kkevinchou/izzet/internal/metrics2"
)

var (
	clientRegistry *metrics2.Registry
	serverRegistry *metrics2.Registry
)

func init() {
	clientRegistry = metrics2.NewRegistry(nil)
	serverRegistry = metrics2.NewRegistry(nil)
}

func ClientRegistry() *metrics2.Registry {
	return clientRegistry
}

func ServerRegistry() *metrics2.Registry {
	return serverRegistry
}
