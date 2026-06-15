package clientsystem

import (
	"time"

	"github.com/kkevinchou/izzet/internal/input"
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
	e := s.app.GetPlayerEntity()
	frameInput := s.app.GetFrameInput()

	if event, ok := frameInput.KeyboardInput[input.KeyboardKeyE]; ok {
		if event.Event == input.KeyboardEventDown {
			time.Sleep(100 * time.Millisecond)
		}
	}

	shared.UpdateCharacterController(delta, frameInput, e)
	if e.AimDownSightsComponent.Fire {
		s.app.AssetManager().Play("shot")
	}
}
