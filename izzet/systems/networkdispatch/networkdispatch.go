package networkdispatch

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/commandframe"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/izzet/lib/metrics"
	"github.com/kkevinchou/izzet/lib/network"
)

type MessageFetcher func(world World) []*network.Message
type MessageHandler func(world World, message *network.Message)

type World interface {
	RegisterEntities([]entities.Entity)
	GetSingleton() *singleton.Singleton
	GetEventBroker() eventbroker.EventBroker
	GetCommandFrameHistory() *commandframe.CommandFrameHistory
	CommandFrame() int
	GetCamera() entities.Entity
	GetPlayerEntity() entities.Entity
	MetricsRegistry() *metrics.MetricsRegistry
	GetPlayer() *player.Player
	GetPlayerByID(id int) *player.Player
	QueryEntity(componentFlags int) []entities.Entity
	GetEntityByID(id int) entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
	SetServerStats(serverStats map[string]string)
}

type NetworkDispatchSystem struct {
	*base.BaseSystem
	world          World
	messageFetcher MessageFetcher
	messageHandler MessageHandler
}

func NewNetworkDispatchSystem(world World) *NetworkDispatchSystem {
	networkDispatchSystem := &NetworkDispatchSystem{
		BaseSystem: base.NewBaseSystem(),
		world:      world,
	}

	if utils.IsClient() {
		networkDispatchSystem.messageFetcher = clientMessageFetcher
		networkDispatchSystem.messageHandler = clientMessageHandler
	} else if utils.IsServer() {
		networkDispatchSystem.messageFetcher = connectedPlayersMessageFetcher
		networkDispatchSystem.messageHandler = serverMessageHandler
	}

	return networkDispatchSystem
}

func (s *NetworkDispatchSystem) Update(delta time.Duration) {
	var latestGameStateUpdate *network.Message
	messages := s.messageFetcher(s.world)
	for _, message := range messages {
		if message.MessageType == knetwork.MessageTypeGameStateUpdate {
			latestGameStateUpdate = message
		}
	}

	var filteredMessages []*network.Message
	for _, message := range messages {
		// only take the latest gamestate update message
		if message.MessageType == knetwork.MessageTypeGameStateUpdate && message != latestGameStateUpdate {
			continue
		}
		filteredMessages = append(filteredMessages, message)
	}

	sawInputMessage := false
	for _, message := range filteredMessages {
		if utils.IsServer() {
			if message.MessageType == knetwork.MessageTypeInput {
				sawInputMessage = true
			}
		}
		s.messageHandler(s.world, message)
	}
	_ = sawInputMessage
	// if utils.IsServer() && !sawInputMessage {
	// 	fmt.Println("MISSED AN INPUT MESSAGE ON CF", s.world.CommandFrame())
	// }
}

func (s *NetworkDispatchSystem) Name() string {
	return "NetworkDispatchSystem"
}
