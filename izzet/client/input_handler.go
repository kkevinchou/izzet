package client

import (
	"fmt"

	"github.com/kkevinchou/kitolib/input"
)

func (g *Client) Shutdown() {
	g.gameOver = true
}

func (g *Client) HandleInput(frameInput input.Input) {
	for _, cmd := range frameInput.Commands {
		if _, ok := cmd.(input.QuitCommand); ok {
			g.Shutdown()
		} else {
			panic(fmt.Sprintf("UNEXPECTED COMMAND %v", cmd))
		}
	}
}
