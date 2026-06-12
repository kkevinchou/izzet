package serversystem

import (
	"math"
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

const (
	neighborSeparationRadius   = 3.0
	neighborSeparationStrength = 0.85
	neighborSeparationEpsilon  = 0.0001
)

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
						dir = applyNeighborSeparation(e, world, dir)
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

func applyNeighborSeparation(e *entity.Entity, world system.GameWorld, desiredDir mgl64.Vec3) mgl64.Vec3 {
	spatialEntities := world.SpatialPartition().QueryEntities(e.BoundingBox())

	var entities []*entity.Entity
	for _, spatialEntity := range spatialEntities {
		entities = append(entities, world.GetEntityByID(spatialEntity.GetID()))
	}

	separation := navigationNeighborSeparation(e, entities)
	if utils.Vec3IsZero(separation) {
		return desiredDir
	}

	dir := desiredDir.Add(separation.Mul(neighborSeparationStrength))
	if utils.Vec3IsZero(dir) {
		return desiredDir
	}

	return dir.Normalize()
}

func navigationNeighborSeparation(e *entity.Entity, entities []*entity.Entity) mgl64.Vec3 {
	position := e.Position()
	var separation mgl64.Vec3
	var neighborCount int

	for _, other := range entities {
		if other == e || other.NavigationComponent == nil || other.Kinematic == nil {
			continue
		}

		offset := position.Sub(other.Position())
		offset[1] = 0
		distSqr := offset.LenSqr()
		if distSqr >= neighborSeparationRadius*neighborSeparationRadius {
			continue
		}

		var away mgl64.Vec3
		var weight float64
		if distSqr < neighborSeparationEpsilon {
			// pseudorandom separation deterministic on the entity id
			// this handles an edge case where entities overlap very closely
			away = stableSeparationDirection(e.ID, other.ID)
			weight = 1
		} else {
			dist := math.Sqrt(distSqr)
			away = offset.Mul(1 / dist)
			weight = (neighborSeparationRadius - dist) / neighborSeparationRadius
		}

		separation = separation.Add(away.Mul(weight))
		neighborCount++
	}

	if neighborCount == 0 {
		return mgl64.Vec3{}
	}
	return separation.Mul(1 / float64(neighborCount))
}

func stableSeparationDirection(selfID, otherID int) mgl64.Vec3 {
	a := selfID
	b := otherID
	sign := 1.0
	if a > b {
		a, b = b, a
		sign = -1
	}

	hash := uint32(a)*73856093 ^ uint32(b)*19349663
	angle := float64(hash&0xffff) / 0x10000 * 2 * math.Pi
	return mgl64.Vec3{math.Cos(angle) * sign, 0, math.Sin(angle) * sign}
}
