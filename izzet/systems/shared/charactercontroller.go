package shared

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func UpdateCharacterController(delta time.Duration, frameInput input.Input, entity *entities.Entity) {
	if entity.Kinematic == nil {
		return
	}
	c := entity.CharacterControllerComponent
	c.ControlVector = apputils.GetControlVector(frameInput.KeyboardInput)

	movementDir := calculateMovementDir(frameInput.CameraRotation, c.ControlVector)
	updateKinematicComponent(frameInput, entity, movementDir)
}

func updateKinematicComponent(frameInput input.Input, entity *entities.Entity, movementDir mgl64.Vec3) {
	keyboardInput := frameInput.KeyboardInput

	entity.Kinematic.Velocity = mgl64.Vec3{}

	if event, ok := keyboardInput[input.KeyboardKeyG]; ok {
		if event.Event == input.KeyboardEventUp {
			entity.Kinematic.GravityEnabled = !entity.Kinematic.GravityEnabled
			entity.Kinematic.Velocity = mgl64.Vec3{}
			entity.Kinematic.AccumulatedVelocity = mgl64.Vec3{}
		}
	}

	if entity.Kinematic.GravityEnabled {
		if entity.Kinematic.Grounded {
			if entity.CharacterControllerComponent.ControlVector.Y() > 0 {
				entity.Kinematic.Grounded = false
				v := mgl64.Vec3{0, settings.CharacterJumpVelocity, 0}
				entity.Kinematic.AccumulatedVelocity = entity.Kinematic.AccumulatedVelocity.Add(v)
			}
		}
		entity.Kinematic.Velocity = entity.Kinematic.Velocity.Add(removeYMovement(movementDir).Mul(entity.CharacterControllerComponent.Speed))

		// cameraRotation := frameInput.CameraRotation
		// viewVector := cameraRotation.Rotate(mgl64.Vec3{0, 0, -1})
		// if event, ok := keyboardInput[input.KeyboardKeyF]; ok {
		// 	if event.Event == input.KeyboardEventUp {
		// 		entity.Kinematic.Velocity = entity.Kinematic.Velocity.Add(viewVector.Mul(settings.CharacterWebLaunchSpeed))
		// 	}
		// }
	} else {
		c := entity.CharacterControllerComponent
		entity.Kinematic.Velocity = entity.Kinematic.Velocity.Add(movementDir.Mul(c.FlySpeed))
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
