package izzet

import (
	"fmt"

	"github.com/kkevinchou/kitolib/input"
)

func (g *Izzet) Shutdown() {
	g.gameOver = true
}

func (g *Izzet) HandleInput(frameInput input.Input) {
	for _, cmd := range frameInput.Commands {
		if _, ok := cmd.(input.QuitCommand); ok {
			g.Shutdown()
		} else {
			panic(fmt.Sprintf("UNEXPECTED COMMAND %v", cmd))
		}
	}
}
