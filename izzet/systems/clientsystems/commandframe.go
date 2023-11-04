package clientsystems

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/input"
)

type EntityState struct {
	ID       int
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Velocity mgl64.Vec3
}

type CommandFrame struct {
	FrameNumber int
	FrameInput  input.Input
	PostCFState EntityState
}

type CommandFrameHistory struct {
	CommandFrames      [settings.MaxCommandFrameBufferSize]CommandFrame
	CommandFrameCount  int
	CommandFrameCursor int
}

func NewCommandFrameHistory() *CommandFrameHistory {
	return &CommandFrameHistory{CommandFrameCursor: 0}
}

func (h *CommandFrameHistory) AddCommandFrame(frameNumber int, frameInput input.Input, player *entities.Entity) {
	if h.CommandFrameCount == settings.MaxCommandFrameBufferSize {
		panic("command frame buffer size exceeded")
	}

	cf := CommandFrame{
		FrameNumber: frameNumber,
		FrameInput:  frameInput,
		PostCFState: EntityState{
			ID:       player.GetID(),
			Position: player.LocalPosition,
			Rotation: player.LocalRotation,
			Velocity: player.Physics.Velocity,
		},
	}

	h.CommandFrames[(h.CommandFrameCursor+h.CommandFrameCount)%settings.MaxCommandFrameBufferSize] = cf
	h.CommandFrameCount += 1
	if h.CommandFrames[h.CommandFrameCursor].FrameNumber > h.CommandFrames[(h.CommandFrameCursor+h.CommandFrameCount-1)%settings.MaxCommandFrameBufferSize].FrameNumber {
		fmt.Println("WAT")
	}
}

func (h *CommandFrameHistory) GetFrame(frameNumber int) (CommandFrame, error) {
	index, err := h.GetBufferIndexByFrameNumber(frameNumber)
	if err != nil {
		return CommandFrame{}, nil
	}

	return h.CommandFrames[index], nil
}

func (h *CommandFrameHistory) GetAllFramesStartingFrom(frameNumber int) ([]CommandFrame, error) {
	index, err := h.GetBufferIndexByFrameNumber(frameNumber)
	if err != nil {
		return nil, err
	}

	countDelta := frameNumber - h.CommandFrames[h.CommandFrameCursor].FrameNumber
	frameCount := h.CommandFrameCount - countDelta

	result := make([]CommandFrame, frameCount)
	for i := 0; i < frameCount; i++ {
		result[i] = h.CommandFrames[(index+i)%settings.MaxCommandFrameBufferSize]
	}

	return result, nil
}

func (h *CommandFrameHistory) GetBufferIndexByFrameNumber(frameNumber int) (int, error) {
	if h.CommandFrameCount == 0 {
		return -1, fmt.Errorf("no command frames")
	}

	startFrameNumber := h.CommandFrames[h.CommandFrameCursor].FrameNumber
	if frameNumber-startFrameNumber+1 > h.CommandFrameCount {
		return -1, fmt.Errorf("frame number %d exceeds what's currently stored. cursor: %d count: %d start frame: %d", frameNumber, h.CommandFrameCursor, h.CommandFrameCount, startFrameNumber)
	}

	if frameNumber-startFrameNumber < 0 {
		return -1, fmt.Errorf("frame number %d is too old and is no longer stored. cursor: %d count: %d start frame: %d", frameNumber, h.CommandFrameCursor, h.CommandFrameCount, startFrameNumber)
	}
	index := (h.CommandFrameCursor + frameNumber - startFrameNumber) % settings.MaxCommandFrameBufferSize
	if h.CommandFrames[index].FrameNumber != frameNumber {
		fmt.Println("Wat")
	}

	return index, nil
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
	h.CommandFrameCursor = (h.CommandFrameCursor + delta) % settings.MaxCommandFrameBufferSize
	h.CommandFrameCount -= delta

	if h.CommandFrames[h.CommandFrameCursor].FrameNumber > h.CommandFrames[(h.CommandFrameCursor+h.CommandFrameCount-1)%settings.MaxCommandFrameBufferSize].FrameNumber {
		fmt.Println("WAT")
	}

	return nil
}

func (h *CommandFrameHistory) Reset() {
	h.CommandFrameCursor = 0
	h.CommandFrameCount = 0
	for i := range h.CommandFrames {
		h.CommandFrames[i].FrameNumber = -1
	}
}
