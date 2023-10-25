package serversystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type ClientManagementSystem struct {
	app        App
	serializer *serialization.Serializer
}

func NewClientManagementSystem(app App, serializer *serialization.Serializer) *ClientManagementSystem {
	return &ClientManagementSystem{app: app, serializer: serializer}
}

func (s *ClientManagementSystem) Update(delta time.Duration, world systems.GameWorld) {
	players := s.app.GetPlayers()
	for _, event := range world.GetEvents() {
		switch e := event.(type) {
		case events.PlayerJoinEvent:
			conn := players[e.PlayerID].Connection
			s.serializer.Write(world, conn)
		}
	}
	world.ClearEventQueue()
}
