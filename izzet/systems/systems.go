package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type System interface {
	Update(time.Duration, GameWorld)
}

type GameWorld interface {
	Entities() []*entities.Entity
	GetEntityByID(int) *entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
	GetFrameInput() input.Input
	GetEvents() []events.Event
	QueueEvent(events.Event)
	ClearEventQueue()
}

type CameraSystem struct {
}

func (s *CameraSystem) Update(delta time.Duration, world GameWorld) {
	frameInput := world.GetFrameInput()

	var camera *entities.Entity
	for _, entity := range world.Entities() {
		if entity.CameraComponent != nil {
			camera = entity
			break
		}
	}

	if camera == nil {
		return
	}

	newOrientation := cameraOrientation(frameInput, camera)
	entities.SetLocalRotation(camera, newOrientation)

	targetID := camera.CameraComponent.Target
	if targetID == nil {
		return
	}

	targetEntity := world.GetEntityByID(*targetID)
	if targetEntity == nil {
		return
	}

	// swivel around target
	target := world.GetEntityByID(*camera.CameraComponent.Target)
	targetPosition := target.WorldPosition().Add(camera.CameraComponent.TargetPositionOffset)
	position := newOrientation.Rotate(mgl64.Vec3{0, 0, 100}).Add(targetPosition)

	entities.SetLocalPosition(camera, position)
}

func cameraOrientation(frameInput input.Input, camera *entities.Entity) mgl64.Quat {
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
