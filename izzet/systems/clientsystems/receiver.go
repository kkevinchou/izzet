package clientsystems

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/apputils"
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

func (s *ReceiverSystem) Name() string {
	return "ReceiverSystem"
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	mr := s.app.MetricsRegistry()

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
					entity := world.GetEntityByID(transform.EntityID)
					if entity == nil {
						continue
					}

					if entity.GetID() == playerEntityID {
						serverTransform = transform
						continue
					}

					if entity.Animation != nil {
						animationPlayer := entity.Animation.AnimationPlayer
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
					cfHistory.ClearUntilFrameNumber(gamestateUpdateMessage.LastInputCommandFrame)
				} else {
					mr.Inc("prediction_miss", 1)
					player := s.app.GetPlayerEntity()

					// position := entities.GetLocalPosition(player)
					// rotation := entities.GetLocalRotation(player)
					// velocity := player.Physics.Velocity
					// gravityEnabled := player.Physics.GravityEnabled

					if !player.RenderBlend.Active {
						player.RenderBlend.StartTime = time.Now()
					}
					player.RenderBlend.Active = true
					player.RenderBlend.BlendStartPosition = player.Position()
					// fmt.Println("BLEND START", time.Now())

					replay(s.app, player, gamestateUpdateMessage, cfHistory, world)

					// correctedPosition := entities.GetLocalPosition(player)
					// // correctedRotation := entities.GetLocalRotation(player)
					// // correctedVelocity := player.Physics.Velocity
					// // correctedGravityEnabled := player.Physics.GravityEnabled
					// if !player.PositionSync.Active {
					// 	player.PositionSync.Active = true
					// 	player.PositionSync.StartTime = time.Now()
					// 	player.PositionSync.Goal = correctedPosition
					// }

					// entities.SetLocalPosition(player, position)
					// entities.SetLocalRotation(player, rotation)
					// player.Physics.Velocity = velocity
					// player.Physics.GravityEnabled = gravityEnabled
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

				serialization.InitDeserializedEntity(&entity, s.app.AssetManager())
				fmt.Println("INIT ENTITY", entity.GetID(), entity.Position())
				world.AddEntity(&entity)
			} else if message.MessageType == network.MsgTypePing {
				pingMessage, err := network.ExtractMessage[network.PingMessage](message)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize ping message %w", err))
					continue
				}
				s.app.MetricsRegistry().Inc("ping", float64(time.Now().UnixNano()-pingMessage.UnixTime)/1000000.0)
			}
		default:
			return
		}
	}
}
