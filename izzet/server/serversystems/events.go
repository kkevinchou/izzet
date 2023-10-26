package serversystems

import (
	"fmt"
	"net"
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type App interface {
	GetPlayers() map[int]network.Player
	RegisterPlayer(playerID int, connection net.Conn)
}

type EventsSystem struct {
	app        App
	serializer *serialization.Serializer
}

func NewEventsSystem(app App, serializer *serialization.Serializer) *EventsSystem {
	return &EventsSystem{app: app, serializer: serializer}
}

func (s *EventsSystem) Update(delta time.Duration, world systems.GameWorld) {
	// players := s.app.GetPlayers()
	for _, event := range world.GetEvents() {
		switch e := event.(type) {
		case events.PlayerJoinEvent:
			s.app.RegisterPlayer(e.PlayerID, e.Connection)
			fmt.Println("player joined", e.PlayerID)

			// create the player entity
			// conn := players[e.PlayerID].Connection
			// s.serializer.Write(world, conn)
		}
	}
	world.ClearEventQueue()
}
