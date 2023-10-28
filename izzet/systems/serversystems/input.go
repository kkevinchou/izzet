package serversystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/systems"
)

type InputSystem struct {
	app App
}

func NewInputSystem(app App) *InputSystem {
	return &InputSystem{app: app}
}

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	inputBuffer := s.app.InputBuffer()
	for _, player := range s.app.GetPlayers() {
		input := inputBuffer.PullInput(player.ID)
		s.app.SetPlayerInput(player.ID, input)
		// player := s.app.GetPlayer(player.ID)
		// player.LastInputLocalCommandFrame =
	}
}
