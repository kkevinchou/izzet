package system

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/entity"
)

const (
	maxBulletDistance float64 = 300
)

type CombatSystem struct {
	app App
}

type physicsSpawner interface {
	SpawnPhysicsCube(contactPoint mgl64.Vec3)
}

func NewCombatSystem(app App) *CombatSystem {
	return &CombatSystem{app: app}
}

func (s *CombatSystem) Name() string {
	return "CombatSystem"
}

func (s *CombatSystem) Update(delta time.Duration, world GameWorld) {
	for _, e := range world.Entities() {
		if e.AimDownSightsComponent == nil || !e.AimDownSightsComponent.Fire {
			continue
		}
		camera := world.GetEntityByID(e.CharacterControllerComponent.CameraEntityID)

		bulletRange := camera.LocalRotation.Rotate(mgl64.Vec3{0, 0, -1}).Normalize().Mul(maxBulletDistance)
		position := camera.Position()

		line := collider.Line{P1: position, P2: position.Add(bulletRange)}
		partitionEntities := world.SpatialPartition().EntitiesByLineSegment(line)

		var hitTargets []*entity.Entity
		for _, e := range partitionEntities {
			hitTargets = append(hitTargets, world.GetEntityByID(e.GetID()))
		}

		if hitEntityID, _, hit := collision.ClosestHit(line, hitTargets); hit {
			// if s.app.IsServer() {
			// 	if spawner, ok := s.app.(physicsSpawner); ok {
			// 		spawner.SpawnPhysicsCube(hitPoint)
			// 	}
			// }

			hitEntity := world.GetEntityByID(hitEntityID)
			if hitEntity.HealthComponent != nil {
				if s.app.IsServer() {
					hitEntity.HealthComponent.Amount -= 50
					if hitEntity.HealthComponent.Amount <= 0 {
						hitEntity.Deadge = true
					}
				} else {
					s.app.AssetManager().Play("hit-pip")
				}
			}
		}
	}
}
