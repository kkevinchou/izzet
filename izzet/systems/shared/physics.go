package shared

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
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
			velocityFromGravity := mgl64.Vec3{0, -settings.AccelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
			physicsComponent.Velocity = physicsComponent.Velocity.Add(velocityFromGravity)
			if physicsComponent.RotateOnVelocity {

				if physicsComponent.Velocity != apputils.ZeroVec {
					newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, mgl64.Vec3{physicsComponent.Velocity.X(), 0, physicsComponent.Velocity.Y()})
					entities.SetLocalRotation(entity, newRotation)
				}
			}
		}

		aiComponent := entity.AIComponent
		newPosition := entity.GetLocalPosition()
		if aiComponent != nil {
			newPosition = newPosition.Add(aiComponent.AIVelocity.Mul(delta.Seconds()))
		}
		newPosition = newPosition.Add(physicsComponent.Velocity.Mul(delta.Seconds()))

		entities.SetLocalPosition(entity, newPosition)
	}
}
