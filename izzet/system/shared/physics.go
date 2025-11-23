package shared

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
)

func PhysicsStepSingle(delta time.Duration, e *entity.Entity) {
	PhysicsStep(delta, []*entity.Entity{e})
}

func PhysicsStep(delta time.Duration, worldEntities []*entity.Entity) {
	for _, e := range worldEntities {
		physicsComponent := e.Physics
		if e.Static || physicsComponent == nil {
			continue
		}

		newPosition := e.GetLocalPosition()
		newPosition = newPosition.Add(physicsComponent.Velocity.Mul(delta.Seconds()))

		entity.SetLocalPosition(e, newPosition)
	}
}
