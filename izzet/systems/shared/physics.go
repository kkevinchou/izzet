package shared

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
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

		newPosition := entity.GetLocalPosition()
		newPosition = newPosition.Add(physicsComponent.Velocity.Mul(delta.Seconds()))

		entities.SetLocalPosition(entity, newPosition)
	}
}
