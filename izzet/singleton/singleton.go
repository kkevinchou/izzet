package singleton

import (
	"github.com/kkevinchou/izzet/izzet/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/playercommand/protogen/playercommand"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/statebuffer"
	"github.com/kkevinchou/kitolib/input"
)

// I hate this. Find a better place to put your data, yo
type Singleton struct {
	// client fields
	PlayerID    int
	CameraID    int
	StateBuffer *statebuffer.StateBuffer

	// server fields
	InputBuffer    *inputbuffer.InputBuffer
	PlayerCommands map[int]*playercommand.PlayerCommandList

	// Common
	PlayerInput  map[int]input.Input
	CommandFrame int
}

func NewSingleton() *Singleton {
	return &Singleton{
		PlayerInput:    map[int]input.Input{},
		PlayerCommands: map[int]*playercommand.PlayerCommandList{},
		StateBuffer:    statebuffer.NewStateBuffer(settings.MaxStateBufferCommandFrames),
		InputBuffer:    inputbuffer.NewInputBuffer(settings.MaxInputBufferCommandFrames),
	}
}
