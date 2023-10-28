package commandframe

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/input"
)

const (
	maxCommandFrameBufferSize = 100
)

type EntityState struct {
	ID          int
	Position    mgl64.Vec3
	Orientation mgl64.Quat
	Velocity    mgl64.Vec3
}

type CommandFrame struct {
	FrameNumber int
	FrameInput  input.Input
	PostCFState EntityState
}

type CommandFrameHistory struct {
	CommandFrames      [maxCommandFrameBufferSize]CommandFrame
	CommandFrameCount  int
	CommandFrameCursor int
}

func NewCommandFrameHistory() *CommandFrameHistory {
	return &CommandFrameHistory{CommandFrameCursor: 0}
}

func (h *CommandFrameHistory) AddCommandFrame(frameNumber int, frameInput input.Input, player *entities.Entity) {
	if h.CommandFrameCount == maxCommandFrameBufferSize {
		panic("command frame buffer size exceeded")
	}

	cf := CommandFrame{
		FrameNumber: frameNumber,
		FrameInput:  frameInput,
		PostCFState: EntityState{
			ID:          player.GetID(),
			Position:    player.LocalPosition,
			Orientation: player.LocalRotation,
			Velocity:    player.Physics.Velocity,
		},
	}

	h.CommandFrames[(h.CommandFrameCursor+h.CommandFrameCount)%maxCommandFrameBufferSize] = cf
	h.CommandFrameCount += 1
}

func (h *CommandFrameHistory) GetCommandFrame(frameNumber int) (CommandFrame, error) {
	if h.CommandFrameCount == 0 {
		return CommandFrame{}, fmt.Errorf("no command frames")
	}

	startFrameNumber := h.CommandFrames[h.CommandFrameCursor].FrameNumber
	if frameNumber-startFrameNumber+1 > h.CommandFrameCount {
		return CommandFrame{}, fmt.Errorf("frame number %d exceeds what's currently stored. cursor: %d count: %d start frame: %d", frameNumber, h.CommandFrameCursor, h.CommandFrameCount, startFrameNumber)
	}

	if frameNumber-startFrameNumber < 0 {
		return CommandFrame{}, fmt.Errorf("frame number %d is too old and is no longer stored. cursor: %d count: %d start frame: %d", frameNumber, h.CommandFrameCursor, h.CommandFrameCount, startFrameNumber)
	}

	return h.CommandFrames[(h.CommandFrameCursor+frameNumber-startFrameNumber)%maxCommandFrameBufferSize], nil
}

func (h *CommandFrameHistory) ClearUntilFrameNumber(frameNumber int) error {
	if h.CommandFrameCount == 0 {
		return fmt.Errorf("got frame number %d, command frame is already empty. cursor: %d count: %d", frameNumber, h.CommandFrameCursor, h.CommandFrameCount)
	}

	startFrameNumber := h.CommandFrames[h.CommandFrameCursor].FrameNumber
	if frameNumber-startFrameNumber+1 > h.CommandFrameCount {
		return fmt.Errorf("frame number %d exceeds what's currently stored. cursor: %d count: %d start frame: %d", frameNumber, h.CommandFrameCursor, h.CommandFrameCount, startFrameNumber)
	}

	delta := frameNumber - startFrameNumber
	h.CommandFrameCursor = (h.CommandFrameCursor + delta) % maxCommandFrameBufferSize
	h.CommandFrameCount -= delta
	return nil
}

func (h *CommandFrameHistory) Reset() {
	h.CommandFrameCursor = 0
	h.CommandFrameCount = 0
}
