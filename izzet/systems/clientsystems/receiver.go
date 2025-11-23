package clientsystems

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/globals"
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

func (s *ReceiverSystem) Name() string {
	return "ReceiverSystem"
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	mr := globals.ClientRegistry()

	for {
		select {
		case message := <-s.app.NetworkMessagesChannel():
			if message.MessageType == network.MsgTypeGameStateUpdate {
				var gamestateUpdateMessage network.GameStateUpdateMessage
				err := json.Unmarshal(message.Body, &gamestateUpdateMessage)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
					continue
				}

				// this is an edge case where the player has joined, and is receiving
				// a game state update but hasn't had input processed by the server yet.
				// this results in a LastInputCommandFrame of 0, which will not be found
				// in the command frame history
				if gamestateUpdateMessage.LastInputCommandFrame == 0 {
					return
				}

				s.app.SetServerStats(gamestateUpdateMessage.ServerStats)

				playerEntityID := s.app.GetPlayerEntity().GetID()
				var serverTransform network.EntityState

				for _, transform := range gamestateUpdateMessage.EntityStates {
					e := world.GetEntityByID(transform.EntityID)
					if e == nil {
						continue
					}

					if e.GetID() == playerEntityID {
						serverTransform = transform
						continue
					}

					if e.Animation != nil {
						animationPlayer := e.Animation.AnimationPlayer
						animationPlayer.PlayAnimation(transform.Animation)
					}
				}

				if len(gamestateUpdateMessage.DestroyedEntities) > 0 {
					fmt.Println("destroy", gamestateUpdateMessage.DestroyedEntities)
				}

				// entity interpolation
				sb := s.app.StateBuffer()
				sb.Push(gamestateUpdateMessage, s.app.CommandFrame())

				// prediction validation
				cfHistory := s.app.GetCommandFrameHistory()
				cf, err := cfHistory.GetFrame(gamestateUpdateMessage.LastInputCommandFrame)
				if err != nil {
					panic(err)
				}
				state := cf.PostCFState
				if apputils.Vec3ApproxEqualThreshold(state.Position, serverTransform.Position, 0.001) {
					mr.Inc("prediction_hit", 1)
					// if s.app.PredictionDebugLogging() {
					// 	fmt.Printf("\t - Predictiton Hit [Frame: %d]\n",
					// 		gamestateUpdateMessage.LastInputCommandFrame,
					// 	)
					// }
					cfHistory.ClearUntilFrameNumber(gamestateUpdateMessage.LastInputCommandFrame)
					player := s.app.GetPlayerEntity()
					player.RenderBlend.Active = false
				} else {
					mr.Inc("prediction_miss", 1)
					player := s.app.GetPlayerEntity()

					// if s.app.PredictionDebugLogging() {
					// 	fmt.Printf("\t - Predictiton Miss [Frame: %d] [Client: %s] [Server: %s]\n",
					// 		gamestateUpdateMessage.LastInputCommandFrame,
					// 		apputils.FormatVec(state.Position),
					// 		apputils.FormatVec(serverTransform.Position),
					// 	)
					// }

					player.RenderBlend.StartTime = time.Now()
					player.RenderBlend.BlendStartPosition = player.Position()
					replay(s.app, player, gamestateUpdateMessage, cfHistory, world)
				}
			} else if message.MessageType == network.MsgTypeCreateEntity {
				var createEntityMessage network.CreateEntityMessage
				err := json.Unmarshal(message.Body, &createEntityMessage)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
					continue
				}

				e, err := serialization.DeserializeEntity(createEntityMessage.EntityBytes, s.app.AssetManager())
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
					continue
				}
				world.AddEntity(e)
			} else if message.MessageType == network.MsgTypePing {
				pingMessage, err := network.ExtractMessage[network.PingMessage](message)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize ping message %w", err))
					continue
				}
				globals.ClientRegistry().Inc("ping", float64(time.Now().UnixNano()-pingMessage.UnixTime)/1000000.0)
			}
		default:
			return
		}
	}
}
