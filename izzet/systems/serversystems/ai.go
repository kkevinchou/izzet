package serversystems

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/systems"
)

const (
	travelThreshold = 0.5
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
	for _, e := range world.Entities() {
		aiComponent := e.AIComponent
		if aiComponent == nil {
			continue
		}

		if e.Deadge {
			continue
		}

		e.Kinematic.Velocity = mgl64.Vec3{}

		position := e.Position()

		if aiComponent.PatrolConfig != nil {
			target := aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
			if position.Sub(target).Len() < 1 {
				aiComponent.PatrolConfig.Index = (aiComponent.PatrolConfig.Index + 1) % len(aiComponent.PatrolConfig.Points)
				target = aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
			}
			dir := target.Sub(position).Normalize()
			e.Kinematic.Velocity = dir.Mul(aiComponent.Speed)
		}

		if aiComponent.RotationConfig != nil {
			r := e.GetLocalRotation()
			finalRotation := aiComponent.RotationConfig.Quat.Mul(r)
			frameRotation := utils.QInterpolate64(r, finalRotation, float64(delta.Milliseconds())/1000)
			e.SetLocalRotation(frameRotation)
		}

		if aiComponent.TargetConfig != nil {
			target := getTarget(world)
			if target != nil {
				dir := target.Position().Sub(e.Position())
				dir[1] = 0
				if dir.LenSqr() > 0 {
					dir = dir.Normalize()
					e.Kinematic.Velocity = dir.Mul(aiComponent.Speed)

					if dir != apputils.ZeroVec {
						newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
						e.SetLocalRotation(newRotation)
					}
				}
			}
		}

		if aiComponent.PathfindConfig != nil && s.app.NavMesh() != nil {
			if aiComponent.PathfindConfig.State == entity.PathfindingStateGoalSet {
				polyPath := navmesh.FindPath(s.app.NavMesh(), e.Position(), aiComponent.PathfindConfig.Goal)
				straightPath := navmesh.FindStraightPath(s.app.NavMesh().Tiles[0], e.Position(), aiComponent.PathfindConfig.Goal, polyPath)
				navmesh.PATHVERTICES = straightPath

				aiComponent.PathfindConfig.PolyPath = polyPath
				aiComponent.PathfindConfig.Path = straightPath
				aiComponent.PathfindConfig.NextTarget = 1
				aiComponent.PathfindConfig.State = entity.PathfindingStatePathing
			}

			if aiComponent.PathfindConfig.State == entity.PathfindingStatePathing {
				path := aiComponent.PathfindConfig.Path

				targetIndex := aiComponent.PathfindConfig.NextTarget
				target := aiComponent.PathfindConfig.Path[targetIndex]

				var atGoal bool
				if position.Sub(target).Len() < travelThreshold {
					position = target
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
						e.Kinematic.Velocity = dir.Mul(aiComponent.Speed)

						if dir != apputils.ZeroVec {
							newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
							e.SetLocalRotation(newRotation)
						}
					}
				} else {
					aiComponent.PathfindConfig.State = entity.PathfindingStateNoGoal
				}
			}
		}

		if aiComponent.AttackConfig != nil {
			closestDist := math.MaxFloat64
			var dirToTarget mgl64.Vec3
			var closestEntity *entity.Entity

			for _, targetEntity := range world.Entities() {
				if e.ID == targetEntity.ID {
					continue
				}

				if targetEntity.AIComponent == nil || targetEntity.AIComponent.AttackConfig == nil {
					continue
				}

				vecToTarget := targetEntity.Position().Sub(e.Position())
				if vecToTarget.Len() < closestDist {
					closestDist = vecToTarget.Len()
					closestEntity = targetEntity
					dirToTarget = vecToTarget.Normalize()
				}
			}

			if closestEntity != nil {
				if closestDist > 4 {
					aiComponent.PathfindConfig.Goal = closestEntity.Position()
					aiComponent.PathfindConfig.State = entity.PathfindingStateGoalSet
				} else {
					aiComponent.State = entity.AIStateAttack
					newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, mgl64.Vec3{dirToTarget.X(), 0, dirToTarget.Z()})
					e.SetLocalRotation(newRotation)
					aiComponent.PathfindConfig.State = entity.PathfindingStateNoGoal
				}
			}
		}
	}
}

func getTarget(world systems.GameWorld) *entity.Entity {
	var target *entity.Entity
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
