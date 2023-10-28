package serversystems

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/input"
)

type App interface {
	GetPlayers() map[int]*network.Player
	RegisterPlayer(playerID int, connection net.Conn) *network.Player
	InputBuffer() *inputbuffer.InputBuffer
	CommandFrame() int
	ModelLibrary() *modellibrary.ModelLibrary
	GetPlayer(playerID int) *network.Player
	GetPlayerInput(playerID int) input.Input
	SetPlayerInput(playerID int, input input.Input)
	DeregisterPlayer(playerID int)
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
			camera := world.GetEntityByID(e.PlayerCameraID)
			entity := world.GetEntityByID(e.PlayerEntityID)

			cameraMessage, err := createEntityMessage(e.PlayerID, camera)
			if err != nil {
				panic(err)
			}
			entityMessage, err := createEntityMessage(e.PlayerID, entity)
			if err != nil {
				panic(err)
			}

			for _, player := range s.app.GetPlayers() {
				player.Client.Send(cameraMessage, s.app.CommandFrame())
				player.Client.Send(entityMessage, s.app.CommandFrame())
			}
			fmt.Printf("player %d joined, camera %d, entityID %d\n", e.PlayerID, e.PlayerCameraID, e.PlayerEntityID)
		case events.PlayerDisconnectEvent:
			fmt.Printf("player %d disconnected\n", e.PlayerID)
			s.app.DeregisterPlayer(e.PlayerID)
		default:
		}
	}
	world.ClearEventQueue()
}

func createEntityMessage(playerID int, entity *entities.Entity) (network.CreateEntityMessage, error) {
	createEntityMessage := network.CreateEntityMessage{
		OwnerID: playerID,
	}

	entityBytes, err := json.Marshal(entity)
	if err != nil {
		return network.CreateEntityMessage{}, err
	}
	createEntityMessage.EntityBytes = entityBytes

	return createEntityMessage, nil
}
