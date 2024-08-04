package serversystems

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/navmesh"
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

		position := entity.Position()

		if aiComponent.PatrolConfig != nil {
			target := aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
			if position.Sub(target).Len() < 1 {
				aiComponent.PatrolConfig.Index = (aiComponent.PatrolConfig.Index + 1) % len(aiComponent.PatrolConfig.Points)
				target = aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
			}
			movementDirection := target.Sub(position).Normalize()
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
					newPosition := position.Add(dir.Mul(aiComponent.Speed / 1000 * float64(delta.Milliseconds())))
					entities.SetLocalPosition(entity, newPosition)

					if dir != apputils.ZeroVec {
						newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
						entities.SetLocalRotation(entity, newRotation)
					}
				}
			}
		}

		if aiComponent.PathfindConfig != nil {
			if aiComponent.PathfindConfig.State == entities.PathfindingStateGoalSet {
				polyPath := navmesh.FindPath(s.app.NavMesh(), entity.Position(), aiComponent.PathfindConfig.Goal)
				straightPath := navmesh.FindStraightPath(s.app.NavMesh().Tiles[0], entity.Position(), aiComponent.PathfindConfig.Goal, polyPath)
				navmesh.PATHVERTICES = straightPath

				aiComponent.PathfindConfig.PolyPath = polyPath
				aiComponent.PathfindConfig.Path = straightPath
				aiComponent.PathfindConfig.NextTarget = 1
				aiComponent.PathfindConfig.State = entities.PathfindingStatePathing
			}

			if aiComponent.PathfindConfig.State == entities.PathfindingStatePathing {
				path := aiComponent.PathfindConfig.Path

				targetIndex := aiComponent.PathfindConfig.NextTarget
				target := aiComponent.PathfindConfig.Path[targetIndex]

				var atGoal bool
				if position.Sub(target).Len() < 1 {
					if targetIndex == len(path)-1 {
						aiComponent.PathfindConfig.Path = nil
						aiComponent.PathfindConfig.NextTarget = -1
						atGoal = true
					} else {
						targetIndex = (targetIndex + 1) % len(path)
						aiComponent.PathfindConfig.NextTarget = targetIndex
						target = aiComponent.PathfindConfig.Path[targetIndex]
					}
				}

				if !atGoal {
					vecToTarget2D := target.Sub(position)
					vecToTarget2D[1] = 0
					dir := vecToTarget2D
					if dir.LenSqr() > 0 {
						dir = dir.Normalize()
						newPosition := position.Add(dir.Mul(aiComponent.Speed / 1000 * float64(delta.Milliseconds())))
						entities.SetLocalPosition(entity, newPosition)

						if dir != apputils.ZeroVec {
							newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
							entities.SetLocalRotation(entity, newRotation)
						}
					}
				} else {
					aiComponent.PathfindConfig.State = entities.PathfindingStateNoGoal
				}
			}
		}

		if aiComponent.AttackConfig != nil {
			closestDist := math.MaxFloat64
			var dirToTarget mgl64.Vec3
			var closestEntity *entities.Entity

			for _, targetEntity := range world.Entities() {
				if entity.ID == targetEntity.ID {
					continue
				}

				if targetEntity.AIComponent == nil || targetEntity.AIComponent.AttackConfig == nil {
					continue
				}

				vecToTarget := targetEntity.Position().Sub(entity.Position())
				if vecToTarget.Len() < closestDist {
					closestDist = vecToTarget.Len()
					closestEntity = targetEntity
					dirToTarget = vecToTarget.Normalize()
				}
			}

			if closestEntity != nil {
				if closestDist > 4 {
					aiComponent.PathfindConfig.Goal = closestEntity.Position()
					aiComponent.PathfindConfig.State = entities.PathfindingStateGoalSet
				} else {
					aiComponent.State = entities.AIStateAttack
					newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, mgl64.Vec3{dirToTarget.X(), 0, dirToTarget.Z()})
					entities.SetLocalRotation(entity, newRotation)
					aiComponent.PathfindConfig.State = entities.PathfindingStateNoGoal
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
