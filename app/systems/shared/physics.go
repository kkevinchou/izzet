package shared

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/app/apputils"
	"github.com/kkevinchou/izzet/app/entities"
)

const (
	accelerationDueToGravity float64 = 450 // units per second
)

func PhysicsStepSingle(delta time.Duration, entity *entities.Entity) {
	PhysicsStep(delta, []*entities.Entity{entity})
}

func PhysicsStep(delta time.Duration, worldEntities []*entities.Entity) {
	for _, entity := range worldEntities {
		physicsComponent := entity.Physics
		if entity.Static || physicsComponent == nil {
			continue
		}

		if physicsComponent.GravityEnabled {
			velocityFromGravity := mgl64.Vec3{0, -accelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
			physicsComponent.Velocity = physicsComponent.Velocity.Add(velocityFromGravity)
			if physicsComponent.OrientOnVelocity {

				if physicsComponent.Velocity != apputils.ZeroVec {
					newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, mgl64.Vec3{physicsComponent.Velocity.X(), 0, physicsComponent.Velocity.Y()})
					entities.SetLocalRotation(entity, newRotation)
				}
			}
		}
		entities.SetLocalPosition(entity, entities.GetLocalPosition(entity).Add(physicsComponent.Velocity.Mul(delta.Seconds())))
	}
}
