package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

type PhysicsSystem struct {
	app App
}

func NewPhysicsSystem(app App) *PhysicsSystem {
	return &PhysicsSystem{app: app}
}

func (s *PhysicsSystem) Name() string {
	return "PhysicsSystem"
}

func (s *PhysicsSystem) Update(delta time.Duration, world GameWorld) {
	var worldEntities []*entity.Entity
	if s.app.IsClient() {
		worldEntities = []*entity.Entity{s.app.GetPlayerEntity()}
	} else {
		worldEntities = world.Entities()
	}

	shared.PhysicsStep(delta, worldEntities)
}
