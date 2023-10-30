package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
)

const (
	resolveCountMax          int     = 3
	groundedThreshold        float64 = 0.85
	accelerationDueToGravity float64 = 250 // units per second
)

type PhysicsObserver interface {
	OnSpatialQuery(entityID int, count int)
	OnCollisionCheck(e1 *entities.Entity, e2 *entities.Entity)
	OnCollisionResolution(entityID int)
	OnBoundingBoxCheck(e1 *entities.Entity, e2 *entities.Entity)
	Clear()
}

type PhysicsSystem struct {
	Observer PhysicsObserver
	app      App
}

func NewPhysicsSystem(app App, physicsObserver PhysicsObserver) *PhysicsSystem {
	return &PhysicsSystem{app: app, Observer: physicsObserver}

}

func (s *PhysicsSystem) Update(delta time.Duration, world GameWorld) {
	var worldEntities []*entities.Entity
	if s.app.IsClient() {
		worldEntities = []*entities.Entity{s.app.GetPlayerEntity()}
	} else {
		worldEntities = world.Entities()
	}

	for _, entity := range worldEntities {
		physicsComponent := entity.Physics
		if entity.Static || physicsComponent == nil {
			continue
		}

		s.Observer.Clear()

		if physicsComponent.GravityEnabled {
			velocityFromGravity := mgl64.Vec3{0, -accelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
			physicsComponent.Velocity = physicsComponent.Velocity.Add(velocityFromGravity)
		}
		entities.SetLocalPosition(entity, entities.GetLocalPosition(entity).Add(physicsComponent.Velocity.Mul(delta.Seconds())))
	}

	// reset contacts - probably want to do this later
	for _, entity := range worldEntities {
		if entity.Collider == nil {
			continue
		}

		if entity.Physics != nil {
			entity.Physics.Grounded = false
		}

		if entity.Collider.Contacts != nil && entity.Physics != nil {
			for _, contact := range entity.Collider.Contacts {
				if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > groundedThreshold {
					entity.Physics.Grounded = true
					entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
				}
			}
		}

		// entity.Collider.Contacts = map[int]bool{}
		entity.Collider.Contacts = nil
	}
}
