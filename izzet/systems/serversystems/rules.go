package serversystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type RulesSystem struct {
	app   App
	world systems.GameWorld
}

func NewRulesSystem(app App) *RulesSystem {
	return &RulesSystem{
		app: app,
	}
}

func (s *RulesSystem) Name() string {
	return "RulesSystem"
}

func (s *RulesSystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, e := range world.Entities() {
		if !e.Deadge {
			continue
		}
		s.world.DeleteEntity(e.GetID())
		s.app.EventsManager().DestroyEntityTopic.Write(events.DestroyEntityEvent{EntityID: e.ID})
	}
}
