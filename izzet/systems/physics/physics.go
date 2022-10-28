package physics

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/netsync"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/utils"

	"github.com/kkevinchou/izzet/izzet/entities"
)

type World interface {
	GetSingleton() *singleton.Singleton
	GetPlayerEntity() entities.Entity
	QueryEntity(componentFlags int) []entities.Entity
}

type PhysicsSystem struct {
	*base.BaseSystem
	world World
}

func NewPhysicsSystem(world World) *PhysicsSystem {
	return &PhysicsSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}
}

func (s *PhysicsSystem) Update(delta time.Duration) {
	// physics simulation is done on the server and the results are synchronized to the client
	if utils.IsClient() {
		return
	}

	for _, entity := range s.world.QueryEntity(components.ComponentFlagPhysics | components.ComponentFlagTransform) {
		netsync.PhysicsStep(delta, entity)
	}
}

func (s *PhysicsSystem) Name() string {
	return "PhysicsSystem"
}
