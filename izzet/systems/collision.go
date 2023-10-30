package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

type CollisionSystem struct {
	app      App
	observer PhysicsObserver
}

func NewCollisionSystem(app App) *CollisionSystem {
	return &CollisionSystem{app: app}
}

func (s *CollisionSystem) Update(delta time.Duration, world GameWorld) {
	shared.ResolveCollisions(world, world.Entities(), observers.NewPhysicsObserver())

	var worldEntities []*entities.Entity
	if s.app.IsClient() {
		worldEntities = []*entities.Entity{s.app.GetPlayerEntity()}
	} else {
		worldEntities = world.Entities()
	}

	// reset contacts - probably want to do this later in a separate system
	for _, entity := range worldEntities {
		if entity.Static || entity.Collider == nil {
			continue
		}

		if entity.Physics != nil {
			entity.Physics.Grounded = false
		}

		if entity.Collider.Contacts != nil && entity.Physics != nil {
			for _, contact := range entity.Collider.Contacts {
				if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > shared.GroundedThreshold {
					entity.Physics.Grounded = true
					entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
				}
			}
		}

		entity.Collider.Contacts = nil
	}
}
