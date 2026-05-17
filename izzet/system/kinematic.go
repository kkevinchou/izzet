package system

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/system/shared"
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

	cf := s.app.CommandFrame()
	entity := world.GetEntityByID(4586)

	// if s.app.IsClient() && entity != nil {
	// 	fmt.Println()
	// }

	// if s.app.IsServer() {
	// 	player := s.app.GetPlayer(100000)
	// 	if player != nil {
	// 		cf = player.LastInputLocalCommandFrame
	// 	}
	// }

	if entity != nil {
		logger := s.app.Logger()
		if s.app.IsClient() {
			logger.Info("-", "cf", cf, "id", entity.GetID(), "position", entity.LocalPosition)
		} else {
			player := s.app.GetPlayer(100000)
			var localCF int
			if player != nil {
				localCF = player.LastInputLocalCommandFrame
			}
			logger.Info("-", "cf", localCF, "gcf", cf, "id", entity.GetID(), "position", entity.LocalPosition)
		}
	}
}
