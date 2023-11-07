package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

type CollisionSystem struct {
	app      App
	observer *observers.CollisionObserver
}

func NewCollisionSystem(app App, observer *observers.CollisionObserver) *CollisionSystem {
	return &CollisionSystem{app: app, observer: observer}
}

func (s *CollisionSystem) Update(delta time.Duration, world GameWorld) {
	var worldEntities []*entities.Entity
	if s.app.IsClient() {
		worldEntities = []*entities.Entity{s.app.GetPlayerEntity()}
		s.observer.Clear()
		shared.ResolveCollisions(world, worldEntities, s.observer)
	} else {
		worldEntities = world.Entities()
		s.observer.Clear()
		start := time.Now()
		shared.ResolveCollisions(world, worldEntities, s.observer)

		s.app.MetricsRegistry().Inc("collision_time", float64(time.Since(start).Microseconds())/1000)
	}

}
