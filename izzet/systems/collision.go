package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
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
	if s.app.IsClient() {
		observer := s.app.CollisionObserver()
		observer.Clear()
		shared.ResolveCollisions(s.app, observer)
	} else {
		shared.ResolveCollisions(s.app, collisionobserver.NullCollisionExplorer)
	}
}
