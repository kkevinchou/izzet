package serversystem

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/system"
)

type ReplicationSystem struct {
	app                   App
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

func (s *ReplicationSystem) Update(delta time.Duration, world system.GameWorld) {
	if s.app.CommandFrame()%settings.NumFramesPerGameStateUpdate != 0 {
		return
	}

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

		if entity.Kinematic != nil {
			// entityState.Velocity = entity.Kinematic.Velocity
			entityState.GravityEnabled = entity.Kinematic.GravityEnabled
		}
		if entity.Animation != nil {
			entityState.AnimationTransitions = convertAnimationTransitions(entity.Animation.AnimationTransitions)
			if len(entityState.AnimationTransitions) > 0 {
				entity.Animation.AnimationTransitions = entity.Animation.AnimationTransitions[:0]
			}
		}
		entityStates = append(entityStates, entityState)
	}

	stats := serverstats.ServerStats{
		Data: []serverstats.Stat{
			{
				Name:  "CFPS",
				Value: fmt.Sprintf("%.0f", globals.ServerRegistry().RatePerSec("command_frames", 1)),
			},
		},
	}

	for _, systemName := range s.app.SystemNames() {
		stats.Data = append(
			stats.Data,
			serverstats.Stat{
				Name:  fmt.Sprintf("%s Time", systemName),
				Value: fmt.Sprintf("%.2f", globals.ServerRegistry().AvgOver(fmt.Sprintf("%s_runtime", systemName), 1)),
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
		s.app.Logger().Info("replication", "cf", gamestateUpdateMessage.LastInputCommandFrame, "gcf", s.app.CommandFrame())
		player.Client.Send(gamestateUpdateMessage, s.app.CommandFrame())
	}
}

func convertAnimationTransitions(animationTransitions []entity.AnimationTransition) []network.AnimationTransition {
	result := make([]network.AnimationTransition, len(animationTransitions))
	for i := range len(animationTransitions) {
		result[i] = network.AnimationTransition{
			SourceState:      animationTransitions[i].SourceState,
			DestinationState: animationTransitions[i].DestinationState,
			CommandFrame:     animationTransitions[i].CommandFrame,
		}
	}
	return result
}
