package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
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
	var ents []*entity.Entity
	if s.app.IsClient() {
		ents = []*entity.Entity{s.app.GetPlayerEntity()}
	} else {
		ents = world.Entities()
	}

	shared.KinematicStep(delta, ents, world, s.app)
}
