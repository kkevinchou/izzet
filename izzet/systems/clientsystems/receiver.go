package clientsystems

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type ReceiverSystem struct {
	app App
}

func NewReceiverSystem(app App) *ReceiverSystem {
	return &ReceiverSystem{app: app}
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	for {
		select {
		case message := <-s.app.NetworkMessagesChannel():
			if message.MessageType == network.MsgTypeGameStateUpdate {
				var gameStateUpdateMessage network.GameStateUpdateMessage
				err := json.Unmarshal(message.Body, &gameStateUpdateMessage)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
					continue
				}

				for _, transform := range gameStateUpdateMessage.Transforms {
					entity := world.GetEntityByID(transform.EntityID)
					if entity == nil {
						continue
					}

					if entity.CameraComponent != nil {
						if entity.PlayerInput.PlayerID == s.app.GetPlayerID() {
							// don't synchronize local camera position
							continue
						}
					}

					entities.SetLocalPosition(entity, transform.Position)
					entities.SetLocalRotation(entity, transform.Orientation)
				}
			} else if message.MessageType == network.MsgTypeCreateEntity {
				var createEntityMessage network.CreateEntityMessage
				err := json.Unmarshal(message.Body, &createEntityMessage)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
					continue
				}

				var entity entities.Entity
				err = json.Unmarshal(createEntityMessage.EntityBytes, &entity)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
					continue
				}

				serialization.InitDeserializedEntity(&entity, s.app.ModelLibrary(), false)
				world.AddEntity(&entity)
			}
		default:
			return
		}
	}
}
