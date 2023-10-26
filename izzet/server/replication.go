package server

import (
	"encoding/json"
	"time"

	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type App interface {
	GetPlayers() map[int]network.Player
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
		transforms = append(transforms, network.Transform{EntityID: entity.ID, Position: entity.WorldPosition()})
	}
	gamestateUpdateMessage := network.GameStateUpdateMessage{Transforms: transforms}
	messageBytes, err := json.Marshal(gamestateUpdateMessage)
	if err != nil {
		return
	}

	for _, player := range players {
		conn := player.Connection
		encoder := json.NewEncoder(conn)

		message := network.Message{SenderID: 9000, CommandFrame: s.app.CommandFrame(), Timestamp: time.Now(), Body: messageBytes}
		encoder.Encode(message)
		// s.serializer.Write(world, conn)
	}
}
