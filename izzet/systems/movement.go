package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/utils"
)

type MovementSystem struct {
}

func (s *MovementSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if entity.Movement != nil {
			mc := entity.Movement

			if mc.PatrolConfig != nil {
				target := mc.PatrolConfig.Points[mc.PatrolConfig.Index]
				startPosition := entity.WorldPosition()
				if startPosition.Sub(target).Len() < 5 {
					mc.PatrolConfig.Index = (mc.PatrolConfig.Index + 1) % len(mc.PatrolConfig.Points)
					target = mc.PatrolConfig.Points[mc.PatrolConfig.Index]
				}
				movementDirection := target.Sub(startPosition).Normalize()
				newPosition := startPosition.Add(movementDirection.Mul(mc.Speed / 1000 * float64(delta.Milliseconds())))
				entities.SetLocalPosition(entity, newPosition)
			}

			if mc.RotationConfig != nil {
				r := entities.GetLocalRotation(entity)
				finalRotation := mc.RotationConfig.Quat.Mul(r)
				frameRotation := utils.QInterpolate64(r, finalRotation, float64(delta.Milliseconds())/1000)
				entities.SetLocalRotation(entity, frameRotation)
			}
		}
	}
}
