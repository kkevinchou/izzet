package clientsystems

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/input"
)

type InputSystem struct {
	app App
}

func NewInputSystem(app App) *InputSystem {
	return &InputSystem{app: app}
}

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	// TODO - send inputs asynchronously
	frameInput := world.GetFrameInput()
	cameraOrientation := s.computePlayerCameraOrientation(world, frameInput)
	world.SetInputCameraOrientation(cameraOrientation)
	frameInput = world.GetFrameInput()

	inputMessage := network.InputMessage{
		Input: frameInput,
	}

	err := s.app.Client().Send(inputMessage, s.app.CommandFrame())
	if err != nil {
		fmt.Println(fmt.Errorf("failed to write input message to connection %w", err))
		return
	}
}

func (s *InputSystem) computePlayerCameraOrientation(world systems.GameWorld, frameInput input.Input) mgl64.Quat {
	var camera *entities.Entity
	for _, entity := range world.Entities() {
		if entity.CameraComponent != nil && entity.PlayerInput.PlayerID == s.app.GetPlayerID() {
			camera = entity
			break
		}
	}

	if camera == nil {
		return mgl64.QuatIdent()
	}

	newOrientation := computeCameraOrientation(frameInput, camera)
	entities.SetLocalRotation(camera, newOrientation)
	return newOrientation
}

func computeCameraOrientation(frameInput input.Input, camera *entities.Entity) mgl64.Quat {
	// camera rotations
	var xRel, yRel float64
	mouseInput := frameInput.MouseInput
	var mouseSensitivity float64 = 0.005
	if mouseInput.Buttons[1] && !mouseInput.MouseMotionEvent.IsZero() {
		xRel += -mouseInput.MouseMotionEvent.XRel * mouseSensitivity
		yRel += -mouseInput.MouseMotionEvent.YRel * mouseSensitivity
	}

	forwardVector := camera.LocalRotation.Rotate(mgl64.Vec3{0, 0, -1})
	upVector := camera.LocalRotation.Rotate(mgl64.Vec3{0, 1, 0})
	rightVector := forwardVector.Cross(upVector)

	// calculate the quaternion for the delta in rotation
	deltaRotationX := mgl64.QuatRotate(yRel, rightVector)         // pitch
	deltaRotationY := mgl64.QuatRotate(xRel, mgl64.Vec3{0, 1, 0}) // yaw
	deltaRotation := deltaRotationY.Mul(deltaRotationX)

	newOrientation := deltaRotation.Mul(camera.LocalRotation)
	// don't let the camera go upside down
	if newOrientation.Rotate(mgl64.Vec3{0, 1, 0}).Y() < 0 {
		newOrientation = camera.LocalRotation
	}

	return newOrientation
}
