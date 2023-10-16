package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type GameWorld interface {
	Entities() []*entities.Entity
	GetEntityByID(int) *entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}

type CharacterControllerSystem struct {
}

type CameraSystem struct {
}

func (s *CharacterControllerSystem) Update(delta time.Duration, world GameWorld, frameInput input.Input) {
	for _, entity := range world.Entities() {
		if entity.CharacterControllerComponent == nil {
			continue
		}

		c := entity.CharacterControllerComponent

		keyboardInput := frameInput.KeyboardInput
		if key, ok := keyboardInput[input.KeyboardKeyI]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{0, 0, -c.Speed * float64(delta.Milliseconds()) / 1000}))
		}
		if key, ok := keyboardInput[input.KeyboardKeyK]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{0, 0, c.Speed * float64(delta.Milliseconds()) / 1000}))
		}

		if key, ok := keyboardInput[input.KeyboardKeyJ]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{-c.Speed * float64(delta.Milliseconds()) / 1000, 0, 0}))
		}
		if key, ok := keyboardInput[input.KeyboardKeyL]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{c.Speed * float64(delta.Milliseconds()) / 1000, 0, 0}))
		}
	}
}

func (s *CameraSystem) Update(delta time.Duration, world GameWorld, frameInput input.Input) {
	var camera *entities.Entity
	for _, entity := range world.Entities() {
		if entity.CameraComponent == nil {
			continue
		}
		camera = entity
		break
	}

	if camera == nil {
		return
	}

	targetID := camera.CameraComponent.Target
	if targetID == nil {
		return
	}

	targetEntity := world.GetEntityByID(*targetID)
	if targetEntity == nil {
		return
	}

	entities.SetLocalPosition(camera, targetEntity.WorldPosition().Add(camera.CameraComponent.PositionOffset))
}
