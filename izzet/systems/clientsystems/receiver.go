package clientsystems

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
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
	mr := s.app.MetricsRegistry()

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

				// this is an edge case where the player has joined, and is receiving
				// a game state update but hasn't had input processed by the server yet.
				// this results in a LastInputCommandFrame of 0, which will not be found
				// in the command frame history
				if gameStateUpdateMessage.LastInputCommandFrame == 0 {
					return
				}

				playerEntityID := s.app.GetPlayerEntity().GetID()
				var serverConfirmedPosition mgl64.Vec3

				for _, transform := range gameStateUpdateMessage.Transforms {
					entity := world.GetEntityByID(transform.EntityID)
					if entity == nil {
						continue
					}

					if entity.GetID() == playerEntityID {
						serverConfirmedPosition = transform.Position
						continue
					}

					entities.SetLocalPosition(entity, transform.Position)
					entities.SetLocalRotation(entity, transform.Orientation)
				}

				cfHistory := s.app.GetCommandFrameHistory()
				cf, err := cfHistory.GetCommandFrame(gameStateUpdateMessage.LastInputCommandFrame)
				if err != nil {
					panic(err)
				}

				state := cf.PostCFState
				if Vec3ApproxEqualThreshold(state.Position, serverConfirmedPosition, 0.001) {
					mr.Inc("prediction_hit", 1)
					cfHistory.ClearUntilFrameNumber(gameStateUpdateMessage.LastInputCommandFrame)
				} else {
					mr.Inc("prediction_miss", 1)
					fmt.Println("PREDICTION FAILED", gameStateUpdateMessage.LastInputCommandFrame)
					cfHistory.ClearUntilFrameNumber(gameStateUpdateMessage.LastInputCommandFrame)
					// TODO - resim the frames leading up to the current command frame
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

func Vec3ApproxEqualThreshold(v1 mgl64.Vec3, v2 mgl64.Vec3, threshold float64) bool {
	return v1.ApproxFuncEqual(v2, func(a, b float64) bool {
		return math.Abs(a-b) < threshold
	})
}
