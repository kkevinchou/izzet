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

func UpdateCharacterController(delta time.Duration, world GameWorld, frameInput input.Input, entity *entities.Entity) {
	keyboardInput := frameInput.KeyboardInput
	cameraRotation := frameInput.CameraRotation

	if event, ok := keyboardInput[input.KeyboardKeyG]; ok {
		if event.Event == input.KeyboardEventUp {
			entity.Physics.GravityEnabled = !entity.Physics.GravityEnabled
			entity.Physics.Velocity = mgl64.Vec3{}
		}
	}

	c := entity.CharacterControllerComponent

	var movementDir mgl64.Vec3
	speed := c.Speed
	c.ControlVector = apputils.GetControlVector(keyboardInput)

	movementDirWithoutY := calculateMovementDir(cameraRotation, c.ControlVector, false)

	if entity.Physics.GravityEnabled {
		if entity.Physics.Grounded {
			entity.Physics.Velocity = mgl64.Vec3{}
			if c.ControlVector.Y() > 0 {
				entity.Physics.Grounded = false
				entity.Physics.Velocity = entity.Physics.Velocity.Add(mgl64.Vec3{0, settings.CharacterJumpVelocity, 0})
			}
			if _, ok := keyboardInput[input.KeyboardKeyE]; ok {
				dir := cameraRotation.Rotate(mgl64.Vec3{0, 1, -5}).Normalize()
				entity.Physics.Velocity = entity.Physics.Velocity.Add(dir.Mul(50))
			}
		}
		movementDir = movementDirWithoutY

		c.WebVector = mgl64.Vec3{}
		viewVector := cameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
		if _, ok := keyboardInput[input.KeyboardKeyF]; ok {
			c.WebVector = viewVector.Mul(settings.CharacterWebSpeed)
		}

		if event, ok := keyboardInput[input.KeyboardKeyF]; ok {
			if event.Event == input.KeyboardEventUp {
				entity.Physics.Velocity = entity.Physics.Velocity.Add(viewVector.Mul(settings.CharacterWebLaunchSpeed))
			}
		}
	} else {
		movementDir = calculateMovementDir(cameraRotation, c.ControlVector, true)
		speed = c.FlySpeed
	}

	if movementDirWithoutY != apputils.ZeroVec {
		currentRotation := entities.GetLocalRotation(entity)
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

	finalMovementDir := movementDir.Mul(speed)
	finalMovementDir = finalMovementDir.Add(c.WebVector)
	finalMovementDir = finalMovementDir.Mul(float64(delta.Milliseconds()) / 1000)

	entities.SetLocalPosition(entity, entity.LocalPosition.Add(finalMovementDir))
}

func calculateMovementDir(cameraRotation mgl64.Quat, controlVector mgl64.Vec3, includeY bool) mgl64.Vec3 {
	forwardVector := cameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
	forwardVector = forwardVector.Normalize().Mul(controlVector.Z())

	rightVector := cameraRotation.Rotate(mgl64.Vec3{1, 0, 0})
	rightVector = rightVector.Normalize().Mul(controlVector.X())

	movementDir := forwardVector.Add(rightVector)
	if includeY {
		movementDir = movementDir.Add(mgl64.Vec3{0, 1, 0}.Mul(controlVector.Y()))
	} else {
		movementDir[1] = 0
	}

	if movementDir != apputils.ZeroVec {
		return movementDir.Normalize()
	}

	return movementDir
}
