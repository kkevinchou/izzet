package server

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	GetPlayers() map[int]*network.Player
	CommandFrame() int
	MetricsRegistry() *metrics.MetricsRegistry
}

type Replicator struct {
	app         App
	accumulator int
}

func NewReplicator(app App) *Replicator {
	return &Replicator{app: app}
}

var count int

func (s *Replicator) Update(delta time.Duration, world systems.GameWorld) {
	s.accumulator += int(delta.Milliseconds())
	if s.accumulator < settings.MSPerGameStateUpdate {
		return
	}
	s.accumulator = 0

	players := s.app.GetPlayers()
	count += 1

	var transforms []network.EntityState
	for _, entity := range world.Entities() {
		if entity.CameraComponent != nil {
			continue
		}
		if entity.Static {
			continue
		}
		t := network.EntityState{
			EntityID: entity.ID,
			Position: entities.GetLocalPosition(entity),
			Rotation: entities.GetLocalRotation(entity),
		}
		if entity.Physics != nil {
			t.Velocity = entity.Physics.Velocity
			t.GravityEnabled = entity.Physics.GravityEnabled
		}
		if entity.Animation != nil {
			t.Animation = entity.Animation.AnimationPlayer.CurrentAnimation()
		}
		transforms = append(transforms, t)
	}

	stats := serverstats.ServerStats{
		Data: []serverstats.Stat{
			{
				Name:  "CFPS",
				Value: fmt.Sprintf("%.0f", s.app.MetricsRegistry().GetOneSecondSum("command_frames")),
			},
			{
				Name:  "Collision Time",
				Value: fmt.Sprintf("%.1f", s.app.MetricsRegistry().GetOneSecondAverage("collision_time")),
			},
		},
	}

	gamestateUpdateMessage := network.GameStateUpdateMessage{
		EntityStates:       transforms,
		GlobalCommandFrame: s.app.CommandFrame(),
		ServerStats:        stats,
	}

	for _, player := range players {
		gamestateUpdateMessage.LastInputCommandFrame = player.LastInputLocalCommandFrame
		player.Client.Send(gamestateUpdateMessage, s.app.CommandFrame())
	}
}
