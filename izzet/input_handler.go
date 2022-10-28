package izzet

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/lib/input"
)

func (g *Game) GameOver() {
	g.gameOver = true
}

func (g *Game) HandleInput(frameInput input.Input) {

	keyboardInput := frameInput.KeyboardInput
	if _, ok := keyboardInput[input.KeyboardKeyEscape]; ok {
		// move this into a system maybe
		g.GameOver()
	}

	if keyEvent, ok := keyboardInput[input.KeyboardKeyTick]; ok {
		if keyEvent.Event == input.KeyboardEventUp {
			if on := g.ToggleWindowVisibility(types.WindowConsole); on {
				g.GetEventBroker().Broadcast(&events.ConsoleEnabledEvent{})
			}
		}
	}

	if keyEvent, ok := keyboardInput[input.KeyboardKeyF1]; ok {
		if keyEvent.Event == input.KeyboardEventUp {
			g.ToggleWindowVisibility(types.WindowDebug)
		}
	}

	if keyEvent, ok := keyboardInput[input.KeyboardKeyI]; ok {
		if g.GetFocusedWindow() == types.WindowGame || g.GetFocusedWindow() == types.WindowInventory {
			if keyEvent.Event == input.KeyboardEventUp {
				g.ToggleWindowVisibility(types.WindowInventory)
			}
		}
	}

	for _, cmd := range frameInput.Commands {
		if _, ok := cmd.(input.QuitCommand); ok {
			g.GameOver()
		} else {
			panic(fmt.Sprintf("UNEXPECTED COMMAND %v", cmd))
		}
	}

	if g.GetFocusedWindow() == types.WindowGame {
		singleton := g.GetSingleton()
		singleton.PlayerInput[singleton.PlayerID] = frameInput
	}
}
