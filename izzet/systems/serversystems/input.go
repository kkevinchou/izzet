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

var debugLoggingStart time.Time

func (s *InputSystem) Update(delta time.Duration, world systems.GameWorld) {
	inputBuffer := s.app.InputBuffer()
	for _, player := range s.app.GetPlayers() {
		bufferedInput := inputBuffer.PullInput(player.ID)

		if player.ID == 100001 {
			s.app.SetPredictionDebugLogging(false)
			if _, ok := bufferedInput.Input.KeyboardInput[input.KeyboardKeyA]; ok {
				if time.Since(debugLoggingStart).Seconds() > 5 {
					fmt.Println("---------------------")
					fmt.Println("---------------------")
					fmt.Println("---------------------")
				}
				debugLoggingStart = time.Now()
				s.app.SetPredictionDebugLogging(true)
				fmt.Printf("[%d] - Frame Start\n", s.app.CommandFrame())
			}
		}
		s.app.SetPlayerInput(player.ID, bufferedInput.Input)
		player := s.app.GetPlayer(player.ID)
		player.LastInputLocalCommandFrame = bufferedInput.LocalCommandFrame
	}
}
