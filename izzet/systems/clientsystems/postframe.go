package clientsystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/systems"
)

type PostFrameSystem struct {
	app App
}

func NewPostFrameSystem(app App) *PostFrameSystem {
	return &PostFrameSystem{app: app}
}

func (s *PostFrameSystem) Update(delta time.Duration, world systems.GameWorld) {
	history := s.app.GetCommandFrameHistory()
	history.AddCommandFrame(s.app.CommandFrame(), world.GetFrameInput(), s.app.GetPlayerEntity())
}
