package serversystems

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type ReceiverSystem struct {
	app App
}

func NewReceiverSystem(app App) *ReceiverSystem {
	return &ReceiverSystem{app: app}
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, player := range s.app.GetPlayers() {
		select {
		case message := <-player.InMessageChannel:
			var inputMessage network.InputMessage
			err := json.Unmarshal(message.Body, &inputMessage)
			if err != nil {
				fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
				continue
			}
			s.app.InputBuffer().PushInput(message.CommandFrame, player.ID, inputMessage.Input)
		case <-player.DisconnectChannel:
			world.QueueEvent(events.PlayerDisconnectEvent{PlayerID: player.ID})
		default:
		}
	}
}
