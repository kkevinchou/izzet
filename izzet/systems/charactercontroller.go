package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type CharacterControllerSystem struct {
}

func (s *CharacterControllerSystem) Update(delta time.Duration, world GameWorld) {
	frameInput := world.GetFrameInput()
	keyboardInput := frameInput.KeyboardInput

	var camera *entities.Entity
	for _, entity := range world.Entities() {
		if entity.CameraComponent != nil {
			camera = entity
			break
		}
	}
	if camera == nil || camera.CameraComponent.Target == nil {
		return
	}

	entity := world.GetEntityByID(*camera.CameraComponent.Target)
	if entity.CharacterControllerComponent == nil {
		return
	}

	c := entity.CharacterControllerComponent

	controlVector := app.GetControlVector(keyboardInput)
	movementDir := calculateMovementDir(entities.GetLocalRotation(camera), controlVector)
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
