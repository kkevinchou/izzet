package clientsystems

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
	camera := s.app.GetPlayerCamera()

	if camera == nil || camera.CameraComponent.Target == nil {
		return
	}

	e := world.GetEntityByID(*camera.CameraComponent.Target)
	if e == nil || e.CharacterControllerComponent == nil {
		return
	}

	frameInput := s.app.GetFrameInput()
	shared.UpdateCharacterController(delta, frameInput, e)
}
