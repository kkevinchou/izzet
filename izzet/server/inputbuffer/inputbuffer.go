package inputbuffer

import (
	"fmt"

	"github.com/kkevinchou/izzet/internal/input"
)

const maxBufferedInput int = 3

type BufferedInput struct {
	Input             input.Input
	LocalCommandFrame int
}

type InputBuffer struct {
	playerBuffers map[int]*PlayerBuffer
}

type PlayerBuffer struct {
	count                          int
	inputs                         []BufferedInput
	cursor                         int
	lastSimulatedLocalCommandFrame int
}

func New() *InputBuffer {
	return &InputBuffer{playerBuffers: map[int]*PlayerBuffer{}}
}

func (b *InputBuffer) RegisterPlayer(playerID int) {
	b.playerBuffers[playerID] = &PlayerBuffer{inputs: make([]BufferedInput, maxBufferedInput)}
}

func (b *InputBuffer) DeregisterPlayer(playerID int) {
	delete(b.playerBuffers, playerID)
}

// func (b *InputBuffer) PushInput(localCommandFrame int, playerID int, frameInput input.Input) {
// 	buffer := b.playerBuffers[playerID]
// 	buffer.inputs[buffer.count%maxBufferedInput] = BufferedInput{LocalCommandFrame: localCommandFrame, Input: frameInput}
// 	buffer.count++
// 	if buffer.count-buffer.cursor > maxBufferedInput {
// 		buffer.cursor++
// 	}
// }

// func (b *InputBuffer) PullInput(playerID int) BufferedInput {
// 	buffer := b.playerBuffers[playerID]
// 	if buffer.count == 0 {
// 		// fmt.Println("no input found for player", playerID)
// 		return BufferedInput{}
// 	}

// 	// read stale buffered input
// 	staleRead := false
// 	if buffer.cursor >= buffer.count && buffer.count != 0 {
// 		buffer.cursor = buffer.count - 1
// 		staleRead = true
// 	}

// 	bufferedInput := buffer.inputs[buffer.cursor%maxBufferedInput]
// 	if !staleRead {
// 		buffer.lastSimulatedLocalCommandFrame = bufferedInput.LocalCommandFrame
// 	} else {
// 		buffer.lastSimulatedLocalCommandFrame = buffer.lastSimulatedLocalCommandFrame + 1
// 	}
// 	buffer.cursor++

// 	return bufferedInput
// }

func (b *InputBuffer) PushInput(localCommandFrame int, playerID int, frameInput input.Input) {
	buffer := b.playerBuffers[playerID]
	buffer.inputs = append(buffer.inputs, BufferedInput{Input: frameInput, LocalCommandFrame: localCommandFrame})
	// i.inputs[playerID] = append(i.inputs[playerID], BufferedInput{Input: frameInput, LocalCommandFrame: localCommandFrame})
}

func (b *InputBuffer) PullInput(playerID int) BufferedInput {
	buffer := b.playerBuffers[playerID]
	cursor := buffer.cursor
	if cursor >= len(buffer.inputs) {
		cursor = len(buffer.inputs) - 1
	}
	if cursor == -1 {
		fmt.Println("no input found for player", playerID)
		return BufferedInput{}
	}

	bufferedInput := buffer.inputs[cursor]
	buffer.cursor += 1
	return bufferedInput
}
