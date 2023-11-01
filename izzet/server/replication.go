package server

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type App interface {
	GetPlayers() map[int]*network.Player
	CommandFrame() int
}

type Replicator struct {
	app         App
	serializer  *serialization.Serializer
	accumulator int
}

func NewReplicator(app App, serializer *serialization.Serializer) *Replicator {
	return &Replicator{app: app, serializer: serializer}
}

var count int

func (s *Replicator) Update(delta time.Duration, world systems.GameWorld) {
	s.accumulator += int(delta.Milliseconds())
	if s.accumulator < 100 {
		return
	}
	s.accumulator = 0

	players := s.app.GetPlayers()
	count += 1

	var transforms []network.Transform
	for _, entity := range world.Entities() {
		if entity.CameraComponent != nil {
			continue
		}
		if entity.Static {
			continue
		}
		t := network.Transform{
			EntityID:    entity.ID,
			Position:    entities.GetLocalPosition(entity),
			Orientation: entities.GetLocalRotation(entity),
		}
		if entity.Physics != nil {
			t.Velocity = entity.Physics.Velocity
		}
		if entity.Animation != nil {
			t.Animation = entity.Animation.AnimationPlayer.CurrentAnimation()
		}
		transforms = append(transforms, t)
	}
	gamestateUpdateMessage := network.GameStateUpdateMessage{
		Transforms: transforms,
	}

	for _, player := range players {
		gamestateUpdateMessage.LastInputCommandFrame = player.LastInputLocalCommandFrame
		player.Client.Send(gamestateUpdateMessage, s.app.CommandFrame())
	}
}
