package inputbuffer

import (
	"log/slog"

	"github.com/kkevinchou/izzet/internal/input"
)

const maxBufferedInput int = 10

type BufferedInput struct {
	Input             input.Input
	LocalCommandFrame int
}

type InputBuffer struct {
	playerBuffers map[int]*PlayerBuffer
	app           App
}

type PlayerBuffer struct {
	count  int
	inputs []BufferedInput
	cursor int
	// lastPulledInputCF int
}

type App interface {
	Logger() *slog.Logger
	CommandFrame() int
}

func New(app App) *InputBuffer {
	return &InputBuffer{playerBuffers: map[int]*PlayerBuffer{}, app: app}
}

func (b *InputBuffer) RegisterPlayer(playerID int) {
	b.playerBuffers[playerID] = &PlayerBuffer{inputs: make([]BufferedInput, maxBufferedInput)}
}

func (b *InputBuffer) DeregisterPlayer(playerID int) {
	delete(b.playerBuffers, playerID)
}

func (b *InputBuffer) PushInput(localCommandFrame int, playerID int, frameInput input.Input, stale bool) {
	buffer := b.playerBuffers[playerID]

	// b.app.Logger().Info("push input", "cf", localCommandFrame, "gcf", b.app.CommandFrame(), "stale", stale)
	// if buffer.count > 0 && buffer.inputs[(buffer.count-1)%maxBufferedInput].LocalCommandFrame >= localCommandFrame {
	// 	b.app.Logger().Info("drop late input", "cf", localCommandFrame, "gcf", b.app.CommandFrame())
	// 	return
	// }

	buffer.inputs[buffer.count%maxBufferedInput] = BufferedInput{LocalCommandFrame: localCommandFrame, Input: frameInput}
	buffer.count++
	if buffer.count-buffer.cursor > maxBufferedInput {
		buffer.cursor++
	}
}

func (b *InputBuffer) PullInput(playerID int, globalCommandFrame int) BufferedInput {
	buffer := b.playerBuffers[playerID]
	if buffer.count == 0 {
		// fmt.Println("no input found for player", playerID)
		return BufferedInput{}
	}

	// stale := false
	if buffer.cursor >= buffer.count && buffer.count != 0 {
		buffer.cursor = buffer.count - 1
		// prevInput := buffer.inputs[(buffer.cursor-1)%maxBufferedInput]
		// b.PushInput(prevInput.LocalCommandFrame+1, playerID, prevInput.Input, true)
		// stale = true
	}

	bufferedInput := buffer.inputs[buffer.cursor%maxBufferedInput]
	buffer.cursor++

	// if stale {
	// 	b.app.Logger().Info("read stale output", "cf", bufferedInput.LocalCommandFrame, "gcf", globalCommandFrame)
	// }

	// buffer.lastPulledInputCF = bufferedInput.LocalCommandFrame
	return bufferedInput
}
