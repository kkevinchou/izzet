package shared

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/input"
)

func UpdateCharacterController(delta time.Duration, frameInput input.Input, entity *entities.Entity) {
	c := entity.CharacterControllerComponent
	c.ControlVector = apputils.GetControlVector(frameInput.KeyboardInput)

	updateKinematicComponent(delta, frameInput, entity)

	movementDir := calculateMovementDir(frameInput.CameraRotation, c.ControlVector)
	movementDirWithoutY := removeYMovement(movementDir)

	var speed float64
	if entity.Kinematic.Grounded {
		movementDir = movementDirWithoutY
		speed = c.Speed
	} else {
		speed = c.FlySpeed
	}

	finalMovementDir := movementDir.Mul(speed)
	finalMovementDir = finalMovementDir.Add(c.WebVector)
	finalMovementDir = finalMovementDir.Mul(float64(delta.Milliseconds()) / 1000)

	entities.SetLocalPosition(entity, entity.LocalPosition.Add(finalMovementDir))
	rotateEntityToFaceMovement(entity, movementDirWithoutY)
}

func rotateEntityToFaceMovement(entity *entities.Entity, movementDirWithoutY mgl64.Vec3) {
	if movementDirWithoutY != apputils.ZeroVec {
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

		entities.SetLocalRotation(entity, newRotation)
	}
}

func updateKinematicComponent(delta time.Duration, frameInput input.Input, entity *entities.Entity) {
	keyboardInput := frameInput.KeyboardInput
	cameraRotation := frameInput.CameraRotation

	if event, ok := keyboardInput[input.KeyboardKeyG]; ok {
		if event.Event == input.KeyboardEventUp {
			entity.Kinematic.GravityEnabled = !entity.Kinematic.GravityEnabled
			entity.Kinematic.Velocity = mgl64.Vec3{}
		}
	}

	if entity.Kinematic.GravityEnabled {
		if entity.Kinematic.Grounded {
			entity.Kinematic.Velocity = mgl64.Vec3{}
			if entity.CharacterControllerComponent.ControlVector.Y() > 0 {
				entity.Kinematic.Grounded = false
				entity.Kinematic.Velocity = entity.Kinematic.Velocity.Add(mgl64.Vec3{0, settings.CharacterJumpVelocity, 0})

				dir := cameraRotation.Rotate(mgl64.Vec3{0, 1, -5}).Normalize()
				entity.Kinematic.Velocity = entity.Kinematic.Velocity.Add(dir.Mul(50))
			}
		}

		viewVector := cameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
		if event, ok := keyboardInput[input.KeyboardKeyF]; ok {
			if event.Event == input.KeyboardEventUp {
				entity.Kinematic.Velocity = entity.Kinematic.Velocity.Add(viewVector.Mul(settings.CharacterWebLaunchSpeed))
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
