package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

type CollisionSystem struct {
	app App
}

func NewCollisionSystem(app App) *CollisionSystem {
	return &CollisionSystem{app: app}
}

func (s *CollisionSystem) Update(delta time.Duration, world GameWorld) {
	var worldEntities []*entities.Entity

	if s.app.IsClient() {
		worldEntities = []*entities.Entity{s.app.GetPlayerEntity()}
		observer := s.app.CollisionObserver()
		observer.Clear()
		shared.ResolveCollisions(world, worldEntities, observer)
	} else {
		worldEntities = world.Entities()
		start := time.Now()
		shared.ResolveCollisions(world, worldEntities, observers.NullCollisionExplorer)

		s.app.MetricsRegistry().Inc("collision_time", float64(time.Since(start).Microseconds())/1000)
	}
}
