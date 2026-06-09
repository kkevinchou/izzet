package serversystem

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/system"
)

type RulesSystem struct {
	app App
}

func NewMiscSystem(app App) *RulesSystem {
	return &RulesSystem{
		app: app,
	}
}

func (s *RulesSystem) Name() string {
	return "RulesSystem"
}

func (s *RulesSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, e := range world.Entities() {
		if !e.Deadge {
			continue
		}
		e.NavigationComponent.ClearGoal()
		e.Kinematic.MoveIntent = mgl64.Vec3{}

		// world.DeleteEntity(e.GetID())
		// s.app.EventsManager().DestroyEntityTopic.Write(events.DestroyEntityEvent{EntityID: e.ID})
	}
}
