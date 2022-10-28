package commandframe

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/lib/input"
)

type EntityState struct {
	ID          int
	Position    mgl64.Vec3
	Orientation mgl64.Quat
}

type CommandFrame struct {
	FrameNumber int
	FrameInput  input.Input
	PostCFState EntityState
}

type CommandFrameHistory struct {
	CommandFrames []CommandFrame
}

func NewCommandFrameHistory() *CommandFrameHistory {
	return &CommandFrameHistory{
		CommandFrames: []CommandFrame{},
	}
}

func (h *CommandFrameHistory) AddCommandFrame(frameNumber int, frameInput input.Input, player entities.Entity) {
	transformComponent := player.GetComponentContainer().TransformComponent

	cf := CommandFrame{
		FrameNumber: frameNumber,
		FrameInput:  frameInput,
		PostCFState: EntityState{
			ID:          player.GetID(),
			Position:    transformComponent.Position,
			Orientation: transformComponent.Orientation,
		},
	}

	h.CommandFrames = append(h.CommandFrames, cf)
}

func (h *CommandFrameHistory) GetCommandFrame(frameNumber int) *CommandFrame {
	if len(h.CommandFrames) == 0 {
		return nil
	}

	startFrameNumber := h.CommandFrames[0].FrameNumber
	if frameNumber-startFrameNumber >= len(h.CommandFrames) {
		return nil
	}
	if frameNumber-startFrameNumber < 0 {
		fmt.Printf("unexpectedly doing command frame lookup < 0, frame: %d, startFrame %d\n", frameNumber, startFrameNumber)
		return nil
	}
	return &h.CommandFrames[frameNumber-startFrameNumber]
}

func (h *CommandFrameHistory) ClearUntilFrameNumber(frameNumber int) {
	if len(h.CommandFrames) == 0 {
		return
	}

	startFrameNumber := h.CommandFrames[0].FrameNumber
	h.CommandFrames = h.CommandFrames[frameNumber-startFrameNumber:]
}

func (h *CommandFrameHistory) ClearFrames() {
	h.CommandFrames = []CommandFrame{}
}
