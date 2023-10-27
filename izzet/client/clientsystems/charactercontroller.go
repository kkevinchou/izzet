package clientsystems

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
	var camera *entities.Entity
	for _, entity := range world.Entities() {
		// if entity.CameraComponent != nil && entity.PlayerInput != nil && entity.PlayerInput.PlayerID == s.app.GetPlayerID() {
		if entity.CameraComponent != nil {
			camera = entity
			break
		}
	}
	if camera == nil || camera.CameraComponent.Target == nil {
		return
	}

	entity := world.GetEntityByID(*camera.CameraComponent.Target)
	if entity == nil || entity.CharacterControllerComponent == nil {
		return
	}

	frameInput := world.GetFrameInput()
	shared.UpdateCharacterController(delta, world, frameInput, camera, entity)
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
		fmt.Println(s.app.CommandFrame(), "CLIENT CHARACTER MOVED", entity.WorldPosition())
	}
}
