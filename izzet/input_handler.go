package izzet

import (
	"fmt"

	"github.com/kkevinchou/kitolib/input"
)

func (g *Izzet) Shutdown() {
	g.gameOver = true
}

func (g *Izzet) HandleInput(frameInput input.Input) {

	keyboardInput := frameInput.KeyboardInput
	if _, ok := keyboardInput[input.KeyboardKeyEscape]; ok {
		// move this into a system maybe
		g.Shutdown()
	}

	for _, cmd := range frameInput.Commands {
		if _, ok := cmd.(input.QuitCommand); ok {
			g.Shutdown()
		} else {
			panic(fmt.Sprintf("UNEXPECTED COMMAND %v", cmd))
		}
	}
}
