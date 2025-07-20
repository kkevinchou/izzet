package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

type KinematicSystem struct {
	app App
}

func NewKinematicSystem(app App) *KinematicSystem {
	return &KinematicSystem{app: app}
}

func (s *KinematicSystem) Name() string {
	return "KinematicSystem"
}

func (s *KinematicSystem) Update(delta time.Duration, world GameWorld) {
	var ents []*entities.Entity
	if s.app.IsClient() {
		ents = []*entities.Entity{s.app.GetPlayerEntity()}
	} else {
		ents = world.Entities()
	}

	shared.KinematicStep(delta, ents, world, s.app)
}
