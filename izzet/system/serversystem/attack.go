package serversystem

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/system"
)

type AttackSystem struct {
	app App
}

func NewAttackSystem(app App) *AttackSystem {
	return &AttackSystem{app: app}
}

func (s *AttackSystem) Name() string {
	return "AttackSystem"
}

func (s *AttackSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, e := range world.Entities() {
		if e.AttackComponent == nil || e.Deadge {
			continue
		}
		e.AttackComponent.Attacking = false

		closestDist := math.MaxFloat64
		var dirToTarget mgl64.Vec3
		var closestEntity *entity.Entity

		for _, targetEntity := range world.Entities() {
			if e.ID == targetEntity.ID {
				continue
			}

			// only attack players
			if targetEntity.CharacterControllerComponent == nil {
				continue
			}

			vecToTarget := targetEntity.Position().Sub(e.Position())
			if vecToTarget.Len() < closestDist {
				closestDist = vecToTarget.Len()
				closestEntity = targetEntity
				dirToTarget = vecToTarget.Normalize()
			}
		}

		if closestEntity != nil && closestDist <= e.AttackComponent.AttackRange {
			newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, mgl64.Vec3{dirToTarget.X(), 0, dirToTarget.Z()})
			e.AttackComponent.Attacking = true
			e.SetLocalRotation(newRotation)
			// if e.NavigationComponent != nil {
			// 	e.NavigationComponent.State = entity.PathfindingStateNoGoal
			// }
		}
	}
}
