package shared

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/input"
)

const (
	jumpVelocity float64 = 200
)

func UpdateCharacterController(delta time.Duration, world GameWorld, frameInput input.Input, entity *entities.Entity) {
	keyboardInput := frameInput.KeyboardInput
	cameraOrientation := frameInput.CameraOrientation

	c := entity.CharacterControllerComponent

	c.ControlVector = app.GetControlVector(keyboardInput)
	if entity.Physics.Grounded {
		entity.Physics.Velocity = mgl64.Vec3{}
		if c.ControlVector.Y() > 0 {
			entity.Physics.Grounded = false
			entity.Physics.Velocity = entity.Physics.Velocity.Add(mgl64.Vec3{0, jumpVelocity, 0})
		}
		if _, ok := keyboardInput[input.KeyboardKeyE]; ok {
			dir := cameraOrientation.Rotate(mgl64.Vec3{0, 1, -5}).Normalize()
			entity.Physics.Velocity = entity.Physics.Velocity.Add(dir.Mul(800))
		}
	}
	movementDir := calculateMovementDir(cameraOrientation, c.ControlVector)

	emptyVec := mgl64.Vec3{}
	if movementDir != emptyVec {
		xzMovementDir := mgl64.Vec3{movementDir.X(), 0, movementDir.Z()}
		newRotation := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, xzMovementDir)
		entities.SetLocalRotation(entity, newRotation)
	}

	entities.SetLocalPosition(entity, entity.LocalPosition.Add(movementDir.Mul(c.Speed*float64(delta.Milliseconds())/1000)))
}

func calculateMovementDir(cameraOrientation mgl64.Quat, controlVector mgl64.Vec3) mgl64.Vec3 {
	forwardVector := cameraOrientation.Rotate(mgl64.Vec3{0, 0, -1})
	forwardVector = forwardVector.Normalize().Mul(controlVector.Z())
	forwardVector[1] = 0

	rightVector := cameraOrientation.Rotate(mgl64.Vec3{1, 0, 0})
	rightVector = rightVector.Normalize().Mul(controlVector.X())
	rightVector[1] = 0

	movementDir := forwardVector.Add(rightVector)
	if movementDir.LenSqr() > 0 {
		return movementDir.Normalize()
	}

	return movementDir
}
