package system

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type CharacterOrientationSystem struct {
	app App
}

type orientationEntity interface {
	GetLocalRotation() mgl64.Quat
	SetLocalRotation(mgl64.Quat)
}

func NewCharacterOrientationSystem(app App) *CharacterOrientationSystem {
	return &CharacterOrientationSystem{app: app}
}

func (s *CharacterOrientationSystem) Name() string {
	return "CharacterOrientationSystem"
}

func (s *CharacterOrientationSystem) Update(delta time.Duration, world GameWorld) {
	for _, e := range world.Entities() {
		if e.IsStatic() || !e.IsKinematic() {
			continue
		}

		if e.AimDownSightsComponent.Active {
			camera := world.GetEntityByID(e.CharacterControllerComponent.CameraEntityID)
			rotation := camera.LocalRotation
			forward := rotation.Rotate(mgl64.Vec3{0, 0, -1})
			forward[1] = 0 // remove vertical pitch component

			if forward.LenSqr() > 0 {
				forward = forward.Normalize()

				playerRotation := mgl64.QuatBetweenVectors(
					mgl64.Vec3{0, 0, -1},
					forward,
				)

				e.SetLocalRotation(playerRotation)
			}
		} else {
			v := e.TotalKinematicVelocity()
			vWithoutY := mgl64.Vec3{v.X(), 0, v.Z()}
			if !utils.Vec3IsZero(vWithoutY) {
				vWithoutY = vWithoutY.Normalize()
				rotateEntityToDir(e, vWithoutY)
			}
		}
	}
}

func rotateEntityToDir(entity orientationEntity, movementDirWithoutY mgl64.Vec3) {
	if !utils.Vec3IsZero(movementDirWithoutY) {
		currentRotation := entity.GetLocalRotation()
		currentViewingVector := currentRotation.Rotate(mgl64.Vec3{0, 0, -1})
		newViewingVector := movementDirWithoutY

		dot := currentViewingVector.Dot(newViewingVector)
		dot = mgl64.Clamp(dot, -1, 1)
		acuteAngle := math.Acos(dot)

		turnAnglePerFrame := (2 * math.Pi / 1000) * 2 * float64(settings.MSPerCommandFrame)

		if left := currentViewingVector.Cross(newViewingVector).Y() > 0; !left {
			turnAnglePerFrame = -turnAnglePerFrame
		}

		var newRotation mgl64.Quat

		// turning angle is less than the goal
		if math.Abs(turnAnglePerFrame) < acuteAngle {
			turningQuaternion := mgl64.QuatRotate(turnAnglePerFrame, mgl64.Vec3{0, 1, 0})
			newRotation = turningQuaternion.Mul(currentRotation)
		} else {
			// turning angle overshoots the goal, snap
			newRotation = mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, movementDirWithoutY)
		}

		entity.SetLocalRotation(newRotation)
	}
}
