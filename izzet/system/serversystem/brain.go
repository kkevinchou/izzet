package serversystem

import (
	"math"
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
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
	// only run once every 10 frames
	if s.app.CommandFrame()%10 != 0 {
		return
	}

	var players []*entity.Entity
	// var playerTarget *entity.Entity
	for _, e := range world.Entities() {
		if e.CharacterControllerComponent != nil && e.Kinematic.Grounded {
			players = append(players, e)
		}
	}

	for _, e := range world.Entities() {
		if e.AttackComponent == nil || e.NavigationComponent == nil || e.Deadge {
			continue
		}

		closestDist := math.MaxFloat64
		var closestPlayer *entity.Entity

		for _, p := range players {
			distToTarget := p.Position().Sub(e.Position()).Len()
			if distToTarget < closestDist {
				closestDist = distToTarget
				closestPlayer = p
			}
		}

		e.AttackComponent.TargetID = -1

		if closestPlayer != nil {
			e.AttackComponent.TargetID = closestPlayer.ID
			if closestDist > e.AttackComponent.AttackRange {
				e.NavigationComponent.SetGoal(closestPlayer.Position())
			} else {
				e.NavigationComponent.ClearGoal()
			}
		}
	}
}
