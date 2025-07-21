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

func (s *InputSystem) Name() string {
	return "InputSystem"
}

var predictionDebugLoggingStart time.Time

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	inputBuffer := s.app.InputBuffer()
	for _, player := range s.app.GetPlayers() {
		bufferedInput := inputBuffer.PullInput(player.ID)
		s.app.SetPlayerInput(player.ID, bufferedInput.Input)
		player := s.app.GetPlayer(player.ID)
		player.LastInputLocalCommandFrame = bufferedInput.LocalCommandFrame
	}
}
