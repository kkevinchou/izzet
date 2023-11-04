package serversystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
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

var moveCount int

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

		frameInput := s.app.GetPlayerInput(camera.PlayerInput.PlayerID)

		entities.SetLocalRotation(camera, frameInput.CameraRotation)
		shared.UpdateCharacterController(delta, world, frameInput, targetEntity)
	}
}
