package serversystems

import (
	"time"

	"github.com/kkevinchou/izzet/app/systems"
	"github.com/kkevinchou/izzet/app/systems/shared"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type CharacterControllerSystem struct {
	app App
}

func NewCharacterControllerSystem(app App) *CharacterControllerSystem {
	return &CharacterControllerSystem{app: app}
}

func (s *CharacterControllerSystem) Name() string {
	return "CharacterControllerSystem"
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
