package serversystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/system"
	"github.com/kkevinchou/izzet/izzet/system/shared"
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

func (s *CharacterControllerSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, camera := range world.Entities() {
		if camera.CameraComponent == nil {
			continue
		}

		target := world.GetEntityByID(camera.CameraComponent.Target)
		if target == nil || target.CharacterControllerComponent == nil {
			return
		}

		frameInput := s.app.GetPlayerInput(camera.PlayerInput.PlayerID)

		camera.SetLocalRotation(frameInput.CameraRotation)
		shared.UpdateCharacterController(delta, frameInput, target)
	}
}
