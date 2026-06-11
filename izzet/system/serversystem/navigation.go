package serversystem

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/system"
)

type NavigationSystem struct {
	app App
}

func NewNavigationSystem(app App) *NavigationSystem {
	return &NavigationSystem{app: app}
}

func (s *NavigationSystem) Name() string {
	return "NavigationSystem"
}

func (s *NavigationSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, e := range world.Entities() {
		navigationComponent := e.NavigationComponent
		if navigationComponent == nil || e.Kinematic == nil {
			continue
		}

		e.Kinematic.MoveIntent = mgl64.Vec3{}

		if s.app.NavMesh() != nil {
			position := e.Position()

			if navigationComponent.PathDirty {
				polyPath := navmesh.FindPath(s.app.NavMesh(), e.Position(), navigationComponent.Goal)
				straightPath := navmesh.FindStraightPath(s.app.NavMesh().Tiles[0], e.Position(), navigationComponent.Goal, polyPath)
				navmesh.PATHVERTICES = straightPath

				if len(straightPath) < 2 {
					navigationComponent.Path = nil
					navigationComponent.NextTarget = entity.InvalidNavigationTarget
					navigationComponent.State = entity.Idle
					continue
				}

				navigationComponent.Path = straightPath
				navigationComponent.PathDirty = false
				navigationComponent.NextTarget = 1
				navigationComponent.State = entity.PathfindingStatePathing
			}

			if navigationComponent.State == entity.PathfindingStatePathing {
				path := navigationComponent.Path

				targetIndex := navigationComponent.NextTarget
				target := navigationComponent.Path[targetIndex]

				var atGoal bool
				if position.Sub(target).Len() < travelThreshold {
					position = target
					if targetIndex == len(path)-1 {
						navigationComponent.Path = nil
						navigationComponent.NextTarget = entity.InvalidNavigationTarget
						atGoal = true
					} else {
						targetIndex = (targetIndex + 1) % len(path)
						navigationComponent.NextTarget = targetIndex
						target = navigationComponent.Path[targetIndex]
					}
				}

				if !atGoal {
					vecToTarget2D := target.Sub(position)
					vecToTarget2D[1] = 0
					dir := vecToTarget2D
					if dir.LenSqr() > 0 {
						dir = dir.Normalize()
						e.Kinematic.MoveIntent = dir

						if !utils.Vec3IsZero(dir) {
							newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
							e.SetLocalRotation(newRotation)
						}
					}
				} else {
					navigationComponent.State = entity.Idle
				}
			}
		}
	}
}
