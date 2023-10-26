package inputbuffer

import (
	"fmt"

	"github.com/kkevinchou/kitolib/input"
)

type InputBuffer struct {
	inputs map[int][]input.Input
	cursor map[int]int
}

func New() *InputBuffer {
	return &InputBuffer{
		inputs: map[int][]input.Input{},
		cursor: map[int]int{},
	}
}

func (i *InputBuffer) PushInput(commandFrame int, playerID int, frameInput input.Input) {
	i.inputs[playerID] = append(i.inputs[playerID], frameInput)
}

func (i *InputBuffer) PullInput(playerID int) input.Input {
	cursor := i.cursor[playerID]
	if cursor >= len(i.inputs[playerID]) {
		cursor = len(i.inputs[playerID]) - 1
	}
	if cursor == -1 {
		fmt.Println("no input found for player", playerID)
		return input.Input{}
	}

	input := i.inputs[playerID][cursor]
	i.cursor[playerID] += 1
	return input
}
