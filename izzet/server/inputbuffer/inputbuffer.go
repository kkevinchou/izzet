package inputbuffer

import (
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/iztlog"
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

func (b *InputBuffer) PushInput(localCommandFrame int, playerID int, frameInput input.Input) {
	buffer := b.playerBuffers[playerID]

	// in this scenario, the incoming input was late and simulated based off the previous stale input
	// if the stale input matches the incoming input, then there are no issues, we guessed correctly
	// if the stale input did not match, the client will be considered to have mispredicted and will self correct
	if buffer.lastSimulatedLocalCommandFrame >= localCommandFrame {
		iztlog.ServerLogger.Info("dropped input", "cf", localCommandFrame)
		return
	}

	buffer.inputs[buffer.count%maxBufferedInput] = BufferedInput{LocalCommandFrame: localCommandFrame, Input: frameInput}
	buffer.count++
	if buffer.count-buffer.cursor > maxBufferedInput {
		buffer.cursor++
	}
}

func (b *InputBuffer) PullInput(playerID int) (BufferedInput, bool) {
	buffer := b.playerBuffers[playerID]
	if buffer.count == 0 {
		// fmt.Println("no input found for player", playerID)
		return BufferedInput{}, false
	}

	// read stale buffered input
	staleRead := false
	if buffer.cursor >= buffer.count && buffer.count != 0 {
		// push the last input into the input buffer so that we read a stale input
		prevInput := buffer.inputs[(buffer.count-1)%maxBufferedInput].Input
		b.PushInput(buffer.lastSimulatedLocalCommandFrame+1, playerID, prevInput)
		staleRead = true
	}

	bufferedInput := buffer.inputs[buffer.cursor%maxBufferedInput]
	buffer.lastSimulatedLocalCommandFrame = bufferedInput.LocalCommandFrame
	buffer.cursor++
	return bufferedInput, staleRead
}

// func (b *InputBuffer) PushInput(localCommandFrame int, playerID int, frameInput input.Input) {
// 	buffer := b.playerBuffers[playerID]
// 	buffer.inputs = append(buffer.inputs, BufferedInput{Input: frameInput, LocalCommandFrame: localCommandFrame})
// 	// i.inputs[playerID] = append(i.inputs[playerID], BufferedInput{Input: frameInput, LocalCommandFrame: localCommandFrame})
// }

// func (b *InputBuffer) PullInput(playerID int) BufferedInput {
// 	buffer := b.playerBuffers[playerID]
// 	cursor := buffer.cursor
// 	if cursor >= len(buffer.inputs) {
// 		cursor = len(buffer.inputs) - 1
// 	}
// 	if cursor == -1 {
// 		fmt.Println("no input found for player", playerID)
// 		return BufferedInput{}
// 	}

// 	bufferedInput := buffer.inputs[cursor]
// 	buffer.cursor += 1
// 	return bufferedInput
// }
