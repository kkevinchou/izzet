package shared

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/input"
)

const (
	jumpVelocity float64 = 300
	webSpeed     float64 = 500
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
				entity.Physics.Velocity = entity.Physics.Velocity.Add(mgl64.Vec3{0, jumpVelocity, 0})
			}
			if _, ok := keyboardInput[input.KeyboardKeyE]; ok {
				dir := cameraRotation.Rotate(mgl64.Vec3{0, 1, -5}).Normalize()
				entity.Physics.Velocity = entity.Physics.Velocity.Add(dir.Mul(800))
			}
		}
		movementDir = movementDirWithoutY

		c.WebVector = mgl64.Vec3{}
		viewVector := cameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
		if _, ok := keyboardInput[input.KeyboardKeyF]; ok {
			c.WebVector = viewVector.Mul(webSpeed)
		}

		if event, ok := keyboardInput[input.KeyboardKeyF]; ok {
			if event.Event == input.KeyboardEventUp {
				entity.Physics.Velocity = entity.Physics.Velocity.Add(viewVector.Mul(1000))
			}
		}
	} else {
		movementDir = calculateMovementDir(cameraRotation, c.ControlVector, true)
		speed = c.FlySpeed
	}

	if movementDirWithoutY != apputils.ZeroVec {
		newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, movementDirWithoutY)
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
