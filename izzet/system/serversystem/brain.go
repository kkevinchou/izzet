package serversystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/system"
)

type BrainSystem struct {
	app App
}

func NewBrainSystem(app App) *BrainSystem {
	return &BrainSystem{
		app: app,
	}
}

func (s *BrainSystem) Name() string {
	return "BrainSystem"
}

func (s *BrainSystem) Update(delta time.Duration, world system.GameWorld) {
	// find player within range
	// pathfind to player
	// attack
}
