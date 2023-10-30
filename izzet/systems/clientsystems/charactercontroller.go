package clientsystems

import (
	"time"

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
	camera := s.app.GetPlayerCamera()

	if camera == nil || camera.CameraComponent.Target == nil {
		return
	}

	entity := world.GetEntityByID(*camera.CameraComponent.Target)
	if entity == nil || entity.CharacterControllerComponent == nil {
		return
	}

	frameInput := world.GetFrameInput()
	shared.UpdateCharacterController(delta, world, frameInput, entity)
}
