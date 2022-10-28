package networkupdate

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils/entityutils"
	"github.com/kkevinchou/izzet/lib/metrics"
)

type World interface {
	RegisterEntities([]entities.Entity)
	GetEventBroker() eventbroker.EventBroker
	GetSingleton() *singleton.Singleton
	CommandFrame() int
	QueryEntity(componentFlags int) []entities.Entity
	MetricsRegistry() *metrics.MetricsRegistry
}

type NetworkUpdateSystem struct {
	*base.BaseSystem
	world         World
	elapsedFrames int
	events        []events.Event
}

func NewNetworkUpdateSystem(world World) *NetworkUpdateSystem {
	networkUpdateSystem := &NetworkUpdateSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}

	eventBroker := world.GetEventBroker()
	eventBroker.AddObserver(networkUpdateSystem, []events.EventType{
		events.EventTypeUnregisterEntity,
	})

	return networkUpdateSystem
}

func (s *NetworkUpdateSystem) Observe(event events.Event) {
	if event.Type() == events.EventTypeUnregisterEntity {
		s.events = append(s.events, event)
	}
}

func (s *NetworkUpdateSystem) Update(delta time.Duration) {
	s.elapsedFrames++
	if s.elapsedFrames < settings.CommandFramesPerServerUpdate {
		return
	}

	s.elapsedFrames %= settings.CommandFramesPerServerUpdate

	serverStats := map[string]string{
		"fps":       fmt.Sprintf("%d", int(s.world.MetricsRegistry().GetOneSecondSum("fps"))),
		"frametime": fmt.Sprintf("%d", int(s.world.MetricsRegistry().GetOneSecondAverage("frametime"))),
	}

	gameStateUpdate := &knetwork.GameStateUpdateMessage{
		Entities:    map[int]knetwork.EntitySnapshot{},
		ServerStats: serverStats,
	}

	for _, entity := range s.world.QueryEntity(components.ComponentFlagTransform | components.ComponentFlagNetwork) {
		if entity.Type() == types.EntityTypeCamera {
			continue
		}
		gameStateUpdate.Entities[entity.GetID()] = entityutils.ConstructEntitySnapshot(entity)
	}

	defer s.clearEvents()
	for _, event := range s.events {
		bytes, err := knetwork.Serialize(event)
		if err != nil {
			fmt.Println("failed to serialize event", err)
			continue
		}

		networkEvent := knetwork.Event{Type: event.Type(), Bytes: bytes}
		gameStateUpdate.Events = append(gameStateUpdate.Events, networkEvent)
	}

	d := directory.GetDirectory()
	playerManager := d.PlayerManager()

	for _, player := range playerManager.GetPlayers() {
		gameStateUpdate.LastInputCommandFrame = player.LastInputLocalCommandFrame
		gameStateUpdate.LastInputGlobalCommandFrame = player.LastInputGlobalCommandFrame
		gameStateUpdate.CurrentGlobalCommandFrame = s.world.CommandFrame()
		player.Client.SendMessage(knetwork.MessageTypeGameStateUpdate, gameStateUpdate)
	}
}

func (s *NetworkUpdateSystem) clearEvents() {
	s.events = nil
}

func (s *NetworkUpdateSystem) Name() string {
	return "NetworkUpdateSystem"
}
