package serversystems

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
	"github.com/kkevinchou/kitolib/input"
)

const (
	jumpVelocity float64 = 200
)

type CharacterControllerSystem struct {
	app App
}

func NewCharacterControllerSystem(app App) *CharacterControllerSystem {
	return &CharacterControllerSystem{app: app}
}

func (s *CharacterControllerSystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, entity := range world.Entities() {
		if entity.PlayerInput == nil {
			continue
		}

		if entity.CameraComponent == nil {
			continue
		}

		camera := entity
		if camera.CameraComponent.Target == nil {
			return
		}

		targetEntity := world.GetEntityByID(*camera.CameraComponent.Target)
		if targetEntity == nil || targetEntity.CharacterControllerComponent == nil {
			return
		}

		frameInput := s.app.InputBuffer().PullInput(camera.PlayerInput.PlayerID)

		entities.SetLocalRotation(camera, frameInput.CameraOrientation)
		shared.UpdateCharacterController(delta, world, frameInput, camera, targetEntity)

		moved := false
		if frameInput.KeyboardInput[input.KeyboardKeyW].Event == input.KeyboardEventDown {
			moved = true
		} else if frameInput.KeyboardInput[input.KeyboardKeyA].Event == input.KeyboardEventDown {
			moved = true
		} else if frameInput.KeyboardInput[input.KeyboardKeyS].Event == input.KeyboardEventDown {
			moved = true
		} else if frameInput.KeyboardInput[input.KeyboardKeyD].Event == input.KeyboardEventDown {
			moved = true
		}
		if moved {
			fmt.Println(s.app.CommandFrame(), "SERVER CHARACTER MOVED", targetEntity.WorldPosition())
		}
	}
}
