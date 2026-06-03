package serversystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/system"
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

func (s *AISystem) Update(delta time.Duration, world system.GameWorld) {
	// for _, e := range world.Entities() {
	// 	aiComponent := e.AIComponent
	// 	if aiComponent == nil {
	// 		continue
	// 	}

	// 	if e.Deadge {
	// 		continue
	// 	}

	// 	e.Kinematic.MoveIntent = mgl64.Vec3{}

	// 	position := e.Position()

	// 	if aiComponent.PatrolConfig != nil {
	// 		target := aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
	// 		if position.Sub(target).Len() < 1 {
	// 			aiComponent.PatrolConfig.Index = (aiComponent.PatrolConfig.Index + 1) % len(aiComponent.PatrolConfig.Points)
	// 			target = aiComponent.PatrolConfig.Points[aiComponent.PatrolConfig.Index]
	// 		}
	// 		dir := target.Sub(position).Normalize()
	// 		e.Kinematic.MoveIntent = dir
	// 	}

	// 	if aiComponent.RotationConfig != nil {
	// 		r := e.GetLocalRotation()
	// 		finalRotation := aiComponent.RotationConfig.Quat.Mul(r)
	// 		frameRotation := utils.QInterpolate64(r, finalRotation, float64(delta.Milliseconds())/1000)
	// 		e.SetLocalRotation(frameRotation)
	// 	}

	// 	if aiComponent.TargetConfig != nil {
	// 		target := getTarget(world)
	// 		if target != nil {
	// 			dir := target.Position().Sub(e.Position())
	// 			dir[1] = 0
	// 			if dir.LenSqr() > 0 {
	// 				dir = dir.Normalize()
	// 				e.Kinematic.MoveIntent = dir

	// 				if dir != apputils.ZeroVec {
	// 					newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)
	// 					e.SetLocalRotation(newRotation)
	// 				}
	// 			}
	// 		}
	// 	}
	// }
}

// func getTarget(world system.GameWorld) *entity.Entity {
// 	var target *entity.Entity
// 	for _, camera := range world.Entities() {
// 		if camera.PlayerInput == nil {
// 			continue
// 		}
// 		if camera.CameraComponent.Target == nil {
// 			continue
// 		}
// 		target = world.GetEntityByID(*camera.CameraComponent.Target)
// 		break
// 	}
// 	return target
// }
