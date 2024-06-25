package systems

import (
	"time"

	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/izzet/app/systems/shared"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
)

type CollisionSystem struct {
	app App
}

func NewCollisionSystem(app App) *CollisionSystem {
	return &CollisionSystem{app: app}
}

func (s *CollisionSystem) Name() string {
	return "CollisionSystem"
}

func (s *CollisionSystem) Update(delta time.Duration, world GameWorld) {
	var worldEntities []*entities.Entity

	if s.app.IsClient() {
		worldEntities = []*entities.Entity{s.app.GetPlayerEntity()}
		observer := s.app.CollisionObserver()
		observer.Clear()
		shared.ResolveCollisions(s.app, worldEntities, observer)
	} else {
		worldEntities = world.Entities()
		shared.ResolveCollisions(s.app, worldEntities, collisionobserver.NullCollisionExplorer)
	}
}
