package rpcsender

import (
	"strings"
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems/base"
)

type World interface {
	GetEventBroker() eventbroker.EventBroker
	GetPlayer() *player.Player
}

type RPCSenderSystem struct {
	*base.BaseSystem
	world  World
	events []events.Event
}

func NewRPCSenderSystem(world World) *RPCSenderSystem {
	rpcSystem := &RPCSenderSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}
	eventBroker := world.GetEventBroker()
	eventBroker.AddObserver(rpcSystem, []events.EventType{
		events.EventTypeRPC,
	})
	return rpcSystem
}

func (s *RPCSenderSystem) Observe(event events.Event) {
	if event.Type() == events.EventTypeRPC {
		s.events = append(s.events, event)
	}
}

func (s *RPCSenderSystem) clearEvents() {
	s.events = nil
}

func (s *RPCSenderSystem) Update(delta time.Duration) {
	defer s.clearEvents()

	for _, event := range s.events {
		if e, ok := event.(*events.RPCEvent); ok {
			if s.handleLocalCommand(e) {
				continue
			}
			player := s.world.GetPlayer()
			rpcMessage := knetwork.RPCMessage{Command: e.Command}
			player.Client.SendMessage(knetwork.MessageTypeRPC, rpcMessage)
		}
	}
}

func (s *RPCSenderSystem) handleLocalCommand(e *events.RPCEvent) bool {
	commandSplit := strings.Split(e.Command, " ")
	if len(commandSplit) == 2 {
		if commandSplit[0] == "collision-render" {
			if commandSplit[1] == "true" {
				settings.DebugRenderCollisionVolume = true
			} else if commandSplit[1] == "false" {
				settings.DebugRenderCollisionVolume = false
			}
			return true
		} else if commandSplit[0] == "boundingbox-render" {
			if commandSplit[1] == "true" {
				settings.DebugRenderSpatialPartition = true
			} else if commandSplit[1] == "false" {
				settings.DebugRenderSpatialPartition = false
			}
			return true
		}

	}
	return false
}

func (s *RPCSenderSystem) Name() string {
	return "RPCSystem"
}
