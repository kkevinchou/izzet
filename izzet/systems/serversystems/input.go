package serversystems

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/input"
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

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	inputBuffer := s.app.InputBuffer()
	for _, player := range s.app.GetPlayers() {
		bufferedInput := inputBuffer.PullInput(player.ID)
		s.app.SetPlayerInput(player.ID, bufferedInput.Input)
		if bufferedInput.Input.KeyboardInput[input.KeyboardKeyJ].Event == input.KeyboardEventUp {
			fmt.Println("received J")
		}
		player := s.app.GetPlayer(player.ID)
		player.LastInputLocalCommandFrame = bufferedInput.LocalCommandFrame
	}
}
