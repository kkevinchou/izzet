package serversystem

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/utils"
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
		if e.AttackComponent.TargetID <= 0 {
			continue
		}

		target := world.GetEntityByID(e.AttackComponent.TargetID)
		vecToTarget := target.Position().Sub(e.Position())
		vecToTargetXZ := mgl64.Vec3{vecToTarget.X(), 0, vecToTarget.Z()}

		if vecToTarget.Len() <= e.AttackComponent.AttackRange && !utils.Vec3IsZero(vecToTargetXZ) {
			newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, vecToTargetXZ)
			e.AttackComponent.Attacking = true
			e.SetLocalRotation(newRotation)
		}
	}
}
