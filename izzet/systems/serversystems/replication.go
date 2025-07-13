package serversystems

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type ReplicationSystem struct {
	app                   App
	accumulator           int
	destroyEntityConsumer *events.Consumer[events.DestroyEntityEvent]
}

func NewReplicationSystem(app App) *ReplicationSystem {
	eventsManager := app.EventsManager()
	return &ReplicationSystem{
		app:                   app,
		destroyEntityConsumer: events.NewConsumer(eventsManager.DestroyEntityTopic),
	}
}

func (s *ReplicationSystem) Name() string {
	return "ReplicationSystem"
}

func (s *ReplicationSystem) Update(delta time.Duration, world systems.GameWorld) {
	s.accumulator += int(delta.Milliseconds())
	if s.accumulator < settings.MSPerGameStateUpdate {
		return
	}
	s.accumulator = 0

	players := s.app.GetPlayers()

	var entityStates []network.EntityState
	for _, entity := range world.Entities() {
		if entity.CameraComponent != nil {
			continue
		}
		if entity.Static {
			continue
		}

		entityState := network.EntityState{
			EntityID: entity.ID,
			Position: entity.GetLocalPosition(),
			Rotation: entity.GetLocalRotation(),
		}
		if entity.Physics != nil {
			entityState.Velocity = entity.Physics.Velocity
			entityState.GravityEnabled = entity.Physics.GravityEnabled
		}
		if entity.Animation != nil {
			entityState.Animation = entity.Animation.AnimationPlayer.CurrentAnimation()
		}
		entityStates = append(entityStates, entityState)
	}

	stats := serverstats.ServerStats{
		Data: []serverstats.Stat{
			{
				Name:  "CFPS",
				Value: fmt.Sprintf("%.0f", s.app.MetricsRegistry().GetOneSecondSum("command_frames")),
			},
		},
	}

	for _, systemName := range s.app.SystemNames() {
		stats.Data = append(
			stats.Data,
			serverstats.Stat{
				Name:  fmt.Sprintf("%s Time", systemName),
				Value: fmt.Sprintf("%.2f", s.app.MetricsRegistry().GetOneSecondAverage(fmt.Sprintf("%s_runtime", systemName))),
			},
		)
	}

	var destroyedEntityIDs []int
	for _, e := range s.destroyEntityConsumer.ReadNewEvents() {
		destroyedEntityIDs = append(destroyedEntityIDs, e.EntityID)
	}

	gamestateUpdateMessage := network.GameStateUpdateMessage{
		EntityStates:       entityStates,
		GlobalCommandFrame: s.app.CommandFrame(),
		ServerStats:        stats,
		DestroyedEntities:  destroyedEntityIDs,
	}

	for _, player := range players {
		gamestateUpdateMessage.LastInputCommandFrame = player.LastInputLocalCommandFrame
		player.Client.Send(gamestateUpdateMessage, s.app.CommandFrame())
	}
}
