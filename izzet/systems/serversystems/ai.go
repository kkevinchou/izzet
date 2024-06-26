package serversystems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/utils"
)

type AISystem struct {
	app App
}

func NewAISystemSystem(app App) *AISystem {
	return &AISystem{app: app}
}

func (s *AISystem) Name() string {
	return "AISystem"
}

func (s *AISystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, entity := range world.Entities() {
		aiComponent := entity.AIComponent
		if aiComponent == nil {
			continue
		}

		if entity.Deadge {
			continue
		}

		startPosition := entity.Position()

		if aiComponent.PatrolConfig != nil {
			target := aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
			if startPosition.Sub(target).Len() < 5 {
				aiComponent.PatrolConfig.Index = (aiComponent.PatrolConfig.Index + 1) % len(aiComponent.PatrolConfig.Points)
				target = aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
			}
			movementDirection := target.Sub(startPosition).Normalize()
			entity.Physics.Velocity = movementDirection.Mul(aiComponent.Speed)
		}

		if aiComponent.RotationConfig != nil {
			r := entities.GetLocalRotation(entity)
			finalRotation := aiComponent.RotationConfig.Quat.Mul(r)
			frameRotation := utils.QInterpolate64(r, finalRotation, float64(delta.Milliseconds())/1000)
			entities.SetLocalRotation(entity, frameRotation)
		}

		if aiComponent.TargetConfig != nil {
			target := getTarget(world)
			if target != nil {
				dir := target.Position().Sub(entity.Position())
				dir[1] = 0
				if dir.LenSqr() > 0 {
					dir = dir.Normalize()
					newPosition := startPosition.Add(dir.Mul(aiComponent.Speed / 1000 * float64(delta.Milliseconds())))
					entities.SetLocalPosition(entity, newPosition)

					if dir != apputils.ZeroVec {
						newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
						entities.SetLocalRotation(entity, newRotation)
					}
				}
			}
		}
	}
}

func getTarget(world systems.GameWorld) *entities.Entity {
	var target *entities.Entity
	for _, camera := range world.Entities() {
		if camera.PlayerInput == nil {
			continue
		}
		if camera.CameraComponent.Target == nil {
			continue
		}
		target = world.GetEntityByID(*camera.CameraComponent.Target)
		break
	}
	return target
}
