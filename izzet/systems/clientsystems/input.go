package clientsystems

import (
	"fmt"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type InputSystem struct {
	app App
	f   *os.File
}

func NewInputSystem(app App) *InputSystem {
	return &InputSystem{app: app}
}

func (s *InputSystem) Name() string {
	return "InputSystem"
}

var predictionDebugLoggingStart time.Time

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	frameInput := s.app.GetFrameInputPtr()

	// if settings.SoloClient {
	// 	if time.Since(predictionDebugLoggingStart).Seconds() > 2 {
	// 		s.app.SetPredictionDebugLogging(false)
	// 	}

	// 	if _, ok := frameInput.KeyboardInput[input.KeyboardKeyA]; ok {
	// 		if !s.app.PredictionDebugLogging() {
	// 			fmt.Println("---------------------")
	// 			fmt.Println("---------------------")
	// 			fmt.Println("---------------------")
	// 			predictionDebugLoggingStart = time.Now()
	// 			// s.app.SetPredictionDebugLogging(true)
	// 		}
	// 	}
	// 	if s.app.PredictionDebugLogging() {
	// 		var hasInput bool
	// 		if _, ok := frameInput.KeyboardInput[input.KeyboardKeyA]; ok {
	// 			hasInput = true
	// 		}
	// 		fmt.Printf("[%d] - Frame Start [Input: %t]\n", s.app.CommandFrame(), hasInput)
	// 	}
	// }

	cameraRotation := s.computePlayerCameraRotation(world, *frameInput)
	frameInput.CameraRotation = cameraRotation

	inputMessage := network.InputMessage{
		Input: *frameInput,
	}

	err := s.app.Client().Send(inputMessage, s.app.CommandFrame())
	if err != nil {
		fmt.Println(fmt.Errorf("failed to write input message to connection %w", err))
		return
	}
}

func (s *InputSystem) computePlayerCameraRotation(world systems.GameWorld, frameInput input.Input) mgl64.Quat {
	camera := s.app.GetPlayerCamera()
	newRotation := computeCameraRotation(frameInput, camera)
	camera.SetLocalRotation(newRotation)
	return newRotation
}

func computeCameraRotation(frameInput input.Input, camera *entities.Entity) mgl64.Quat {
	// camera rotations
	var xRel, yRel float64
	mouseInput := frameInput.MouseInput
	var mouseSensitivity float64 = 0.005
	if mouseInput.MouseButtonState[1] && !mouseInput.MouseMotionEvent.IsZero() {
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
