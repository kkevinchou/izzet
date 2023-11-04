package clientsystems

import (
	"fmt"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/input"
)

type InputSystem struct {
	app App
	f   *os.File
}

func NewInputSystem(app App) *InputSystem {
	return &InputSystem{app: app}
}

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	frameInput := world.GetFrameInput()
	cameraRotation := s.computePlayerCameraRotation(world, frameInput)
	world.SetInputCameraRotation(cameraRotation)
	frameInput = world.GetFrameInput()

	inputMessage := network.InputMessage{
		Input: frameInput,
	}

	err := s.app.Client().Send(inputMessage, s.app.CommandFrame())
	if err != nil {
		fmt.Println(fmt.Errorf("failed to write input message to connection %w", err))
		return
	}

	// _, err = s.f.Write([]byte(fmt.Sprintf("%s - %d\n", time.Now().Format("2006-01-02 15:04:05"), s.app.CommandFrame())))
	// if err != nil {
	// 	fmt.Println("failed to write to client log")
	// }
}

func (s *InputSystem) computePlayerCameraRotation(world systems.GameWorld, frameInput input.Input) mgl64.Quat {
	camera := s.app.GetPlayerCamera()
	newRotation := computeCameraRotation(frameInput, camera)
	entities.SetLocalRotation(camera, newRotation)
	return newRotation
}

func computeCameraRotation(frameInput input.Input, camera *entities.Entity) mgl64.Quat {
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

	newRotation := deltaRotation.Mul(camera.LocalRotation)
	// don't let the camera go upside down
	if newRotation.Rotate(mgl64.Vec3{0, 1, 0}).Y() < 0 {
		newRotation = camera.LocalRotation
	}

	return newRotation
}
