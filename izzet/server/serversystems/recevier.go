package serversystems

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/input"
)

type ReceiverSystem struct {
	app App
}

func NewReceiverSystem(app App) *ReceiverSystem {
	return &ReceiverSystem{app: app}
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, player := range s.app.GetPlayers() {
		// for {
		// processedMessage := false
		select {

		case message := <-player.InMessageChannel:
			var inputMessage network.InputMessage
			err := json.Unmarshal(message.Body, &inputMessage)
			if err != nil {
				fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
				continue
			}
			s.app.InputBuffer().PushInput(s.app.CommandFrame(), player.ID, inputMessage.Input)
			if inputMessage.Input.KeyboardInput[input.KeyboardKeyA].Event == input.KeyboardEventDown {
				fmt.Println(s.app.CommandFrame(), "RECEIVER DOWN", player.ID)
			}
			// processedMessage = true
		default:
		}

		// if !processedMessage {
		// 	break
		// }
		// }
	}
	// for {
	// 	select {
	// 	case message := <-s.app.NetworkMessagesChannel():
	// 		if message.MessageType == network.MsgTypeGameStateUpdate {
	// 			var gameStateUpdateMessage network.GameStateUpdateMessage
	// 			err := json.Unmarshal(message.Body, &gameStateUpdateMessage)
	// 			if err != nil {
	// 				fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
	// 				continue
	// 			}
	// 			// fmt.Println(gameStateUpdateMessage)
	// 		} else if message.MessageType == network.MsgTypePlayerInput {
	// 			var inputMessage network.InputMessage
	// 			err := json.Unmarshal(message.Body, &inputMessage)
	// 			if err != nil {
	// 				fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
	// 				continue
	// 			}
	// 			fmt.Println(inputMessage)
	// 		}
	// 	default:
	// 		return
	// 	}
	// }
}
