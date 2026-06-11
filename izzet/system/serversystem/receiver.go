package serversystem

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/prefab"
	"github.com/kkevinchou/izzet/izzet/system"
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

func (s *ReceiverSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, player := range s.app.GetPlayers() {
		noMessage := false
		for !noMessage {
			select {
			case message := <-player.InMessageChannel:
				if message.MessageType == network.MsgTypePlayerInput {
					inputMessage, err := network.ExtractMessage[network.InputMessage](message)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
						continue
					}
					s.app.InputBuffer().PushInput(message.CommandFrame, player.ID, inputMessage.Input)
				} else if message.MessageType == network.MsgTypePing {
					pingMessage, err := network.ExtractMessage[network.PingMessage](message)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
						continue
					}
					player.Client.Send(pingMessage, s.app.CommandFrame())
				} else if message.MessageType == network.MsgTypeRPC {
					rpc, err := network.ExtractMessage[network.RPCMessage](message)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
						continue
					}

					if rpc.Pathfind != nil {
						s.handlePathfindRPC(rpc)
					}

					if rpc.CreateEntity != nil {
						s.handleCreateEntityRPC(world, rpc)
					}

					if rpc.RessurectRPC != nil {
						s.handleRessurectRPC(world, rpc)
					}
				}
			case <-player.DisconnectChannel:
				s.app.EventsManager().PlayerDisconnectTopic.Write(events.PlayerDisconnectEvent{PlayerID: player.ID})
			default:
				noMessage = true
			}
		}
	}
}

func (s *ReceiverSystem) handlePathfindRPC(rpc network.RPCMessage) {
	for _, e := range s.app.World().Entities() {
		if e.NavigationComponent == nil {
			continue
		}
		e.NavigationComponent.Goal = rpc.Pathfind.Goal
		e.NavigationComponent.State = entity.PathfindingStateGoalSet
	}
}

func (s *ReceiverSystem) handleRessurectRPC(world system.GameWorld, rpc network.RPCMessage) {
	e := world.GetEntityByID(rpc.RessurectRPC.ID)
	if e == nil {
		return
	}
	e.HealthComponent.Amount = 100
	e.Deadge = false
}

func (s *ReceiverSystem) handleCreateEntityRPC(world system.GameWorld, rpc network.RPCMessage) {
	e := prefab.CreateNPC(s.app, entity.EntityType(rpc.CreateEntity.EntityType))

	if rpc.CreateEntity.Patrol {
		jitterX := rand.Intn(10)
		jitterZ := rand.Intn(10)
		entity.SetLocalPosition(e, mgl64.Vec3{float64(jitterX), 20, float64(jitterZ)})

		targetDist := 20
		jitterTargetX := rand.Intn(targetDist) - 10
		jitterTargetZ := rand.Intn(targetDist) - 10
		target := mgl64.Vec3{float64(jitterTargetX), 0, float64(jitterTargetZ)}.Normalize().Mul(float64(targetDist))

		e.AIComponent.PatrolConfig = &entity.PatrolConfig{Points: []mgl64.Vec3{{float64(jitterX), 0, float64(jitterZ)}, target}}
	} else {
		e.NavigationComponent = &entity.NavigationComponent{}
	}

	spawnPoint := world.GetSpawnPoint()
	if spawnPoint != nil {
		entity.SetLocalPosition(e, spawnPoint.Position())
	}

	s.app.EventsManager().EntitySpawnTopic.Write(events.EntitySpawnEvent{Entity: e})
}
