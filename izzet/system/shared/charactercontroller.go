package shared

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func UpdateCharacterController(delta time.Duration, frameInput input.Input, e *entity.Entity) {
	if e.Kinematic == nil {
		return
	}
	e.CharacterControllerComponent.ControlVector = apputils.GetControlVector(frameInput.KeyboardInput)
	updateKinematicComponent(frameInput, e)
}

func updateKinematicComponent(frameInput input.Input, e *entity.Entity) {
	keyboardInput := frameInput.KeyboardInput
	movementDir := calculateMovementDir(frameInput.CameraRotation, e.CharacterControllerComponent.ControlVector)

	e.Kinematic.MoveIntent = movementDir
	e.Kinematic.Velocity = mgl64.Vec3{}
	e.Kinematic.Jump = false

	if event, ok := keyboardInput[input.KeyboardKeyG]; ok {
		if event.Event == input.KeyboardEventUp {
			e.Kinematic.GravityEnabled = !e.Kinematic.GravityEnabled
			e.Kinematic.Velocity = mgl64.Vec3{}
			e.Kinematic.AccumulatedVelocity = mgl64.Vec3{}
			if e.Kinematic.GravityEnabled {
				e.Kinematic.Speed = settings.CharacterSpeed
			} else {
				e.Kinematic.Speed = settings.CharacterFlySpeed
			}
		}
	}

	if e.Kinematic.GravityEnabled {
		if e.Kinematic.Grounded {
			if e.CharacterControllerComponent.ControlVector.Y() > 0 {
				e.Kinematic.Jump = true
			}
		}
	}
}

func calculateMovementDir(cameraRotation mgl64.Quat, controlVector mgl64.Vec3) mgl64.Vec3 {
	forwardVector := cameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
	forwardVector = forwardVector.Normalize().Mul(controlVector.Z())

	rightVector := cameraRotation.Rotate(mgl64.Vec3{1, 0, 0})
	rightVector = rightVector.Normalize().Mul(controlVector.X())

	movementDir := forwardVector.Add(rightVector)
	movementDir = movementDir.Add(mgl64.Vec3{0, 1, 0}.Mul(controlVector.Y()))

	if movementDir != apputils.ZeroVec {
		return movementDir.Normalize()
	}

	return movementDir
}

func removeYMovement(movementDir mgl64.Vec3) mgl64.Vec3 {
	movementDirWithoutY := movementDir
	movementDirWithoutY[1] = 0
	if movementDirWithoutY != apputils.ZeroVec {
		movementDirWithoutY = movementDirWithoutY.Normalize()
	}
	return movementDirWithoutY
}
