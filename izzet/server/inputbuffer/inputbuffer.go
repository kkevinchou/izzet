package inputbuffer

import (
	"fmt"

	"github.com/kkevinchou/kitolib/input"
)

type BufferedInput struct {
	Input             input.Input
	LocalCommandFrame int
}

type InputBuffer struct {
	inputs map[int][]BufferedInput
	cursor map[int]int
}

func New() *InputBuffer {
	return &InputBuffer{
		inputs: map[int][]BufferedInput{}, // TOOD -  use a ring buffer of inputs
		cursor: map[int]int{},
	}
}

func (i *InputBuffer) PushInput(localCommandFrame int, playerID int, frameInput input.Input) {
	i.inputs[playerID] = append(i.inputs[playerID], BufferedInput{Input: frameInput, LocalCommandFrame: localCommandFrame})
}

var lastCursor int = 0

func (i *InputBuffer) PullInput(playerID int) BufferedInput {
	cursor := i.cursor[playerID]
	if cursor >= len(i.inputs[playerID]) {
		cursor = len(i.inputs[playerID]) - 1
	}
	if cursor == -1 {
		fmt.Println("no input found for player", playerID)
		return BufferedInput{}
	}

	if cursor == lastCursor {
		fmt.Printf("player %d reading same cursor\n", playerID)
	}
	lastCursor = cursor

	bufferedInput := i.inputs[playerID][cursor]
	i.cursor[playerID] += 1 // todo - should we skip incrementing if we pull an input that matches the last frame? i.e. if we detect that we have a late input?
	return bufferedInput
}
