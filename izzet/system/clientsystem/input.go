package clientsystem

import (
	"fmt"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/izzet/izzet/system"
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

func (s *InputSystem) Update(delta time.Duration, world system.GameWorld) {
	frameInput := s.app.GetFrameInputPtr()

	s.attachPlayerCameraInputs(frameInput)

	s.handleSendInputToServer(frameInput)
	s.handleSetPathfindingTarget(frameInput)
	s.handleSpawnPatrolEntity(frameInput)
	s.handleSpawnEntity(frameInput)
	s.handleToggleMouseCapture(frameInput)
}

func (s *InputSystem) attachPlayerCameraInputs(frameInput *input.Input) {
	playerCamera := s.app.GetPlayerCamera()
	cameraRotation := s.computePlayerCameraRotation(playerCamera, *frameInput)
	frameInput.CameraRotation = cameraRotation
}

func (s *InputSystem) handleSendInputToServer(frameInput *input.Input) {
	inputMessage := network.InputMessage{
		Input: *frameInput,
	}

	err := s.app.Client().Send(inputMessage, s.app.CommandFrame())
	if err != nil {
		fmt.Println(fmt.Errorf("failed to write input message to connection %w", err))
	}
}

func (s *InputSystem) handleSetPathfindingTarget(frameInput *input.Input) {
	event, ok := frameInput.KeyboardInput[input.KeyboardKeyN]
	if !ok || event.Event != input.KeyboardEventUp {
		return
	}

	mousePosition := frameInput.MouseInput.Position
	width, height := s.app.SceneSize()
	ctx := s.app.CameraViewerContext()

	xNDC := (mousePosition.X()/float64(width) - 0.5) * 2

	menuBarSize := float64(render.CalculateMenuBarHeight())
	yNDC := ((float64(height)-mousePosition.Y()+menuBarSize)/float64(height) - 0.5) * 2

	nearPlanePosition := rutils.NDCToWorldPosition(ctx, mgl64.Vec3{xNDC, yNDC, -float64(s.app.RuntimeConfig().Near)})
	camera := s.app.GetPlayerCamera()
	position := camera.Position()
	point, success := s.app.IntersectRayWithEntities(position, nearPlanePosition.Sub(position).Normalize())
	if !success {
		return
	}

	rpcMessage := network.RPCMessage{
		Pathfind: &network.Pathfind{Goal: point},
	}
	s.app.Client().Send(rpcMessage, s.app.CommandFrame())
}

func (s *InputSystem) handleSpawnPatrolEntity(frameInput *input.Input) {
	event, ok := frameInput.KeyboardInput[input.KeyboardKeyJ]
	if !ok || event.Event != input.KeyboardEventUp {
		return
	}

	s.sendSpawnEntityRPC(true)
}

func (s *InputSystem) handleSpawnEntity(frameInput *input.Input) {
	event, ok := frameInput.KeyboardInput[input.KeyboardKeyK]
	if !ok || event.Event != input.KeyboardEventUp {
		return
	}

	s.sendSpawnEntityRPC(false)
}

func (s *InputSystem) sendSpawnEntityRPC(patrol bool) {
	rpcMessage := network.RPCMessage{
		CreateEntity: &network.CreateEntityRPC{
			EntityType: string(panels.SelectedCreateEntityComboOption),
			Patrol:     patrol,
		},
	}
	s.app.Client().Send(rpcMessage, s.app.CommandFrame())
}

func (s *InputSystem) handleToggleMouseCapture(frameInput *input.Input) {
	event, ok := frameInput.KeyboardInput[input.KeyboardKeyQ]
	if !ok || event.Event != input.KeyboardEventUp {
		return
	}

	mouseInput := frameInput.MouseInput
	capture := !s.app.MouseCaptured()
	s.app.SetMouseCaptured(capture)
	if capture {
		s.app.SetCapturedMouseOrigin(int32(mouseInput.Position.X()), int32(mouseInput.Position.Y()))
	}
}

func (s *InputSystem) computePlayerCameraRotation(camera *entity.Entity, frameInput input.Input) mgl64.Quat {
	if s.app.MouseCaptured() {
		newRotation := computeCameraRotation(frameInput, camera)
		camera.SetLocalRotation(newRotation)
	}
	return camera.GetLocalRotation()
}

func computeCameraRotation(frameInput input.Input, camera *entity.Entity) mgl64.Quat {
	// camera rotations
	var xRel, yRel float64
	mouseInput := frameInput.MouseInput
	var mouseSensitivity float64 = 0.005

	if !mouseInput.MouseMotionEvent.IsZero() {
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
