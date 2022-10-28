package networkinput

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/playercommand/protogen/playercommand"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/lib/metrics"
	"google.golang.org/protobuf/proto"
)

type World interface {
	GetSingleton() *singleton.Singleton
	MetricsRegistry() *metrics.MetricsRegistry
	CommandFrame() int
	GetPlayer() *player.Player
	GetCamera() entities.Entity
	GetFocusedWindow() types.Window
	GetEventBroker() eventbroker.EventBroker
}

type NetworkInputSystem struct {
	*base.BaseSystem
	world    World
	entities []entities.Entity
	events   []events.Event
}

func NewNetworkInputSystem(world World) *NetworkInputSystem {
	s := &NetworkInputSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}

	eventBroker := world.GetEventBroker()
	eventBroker.AddObserver(s, []events.EventType{
		events.EventTypePlayerCommand,
	})

	return s
}

func (s *NetworkInputSystem) Observe(event events.Event) {
	if event.Type() == events.EventTypePlayerCommand {
		s.events = append(s.events, event)
	}
}

func (s *NetworkInputSystem) clearEvents() {
	s.events = nil
}

func (s *NetworkInputSystem) Update(delta time.Duration) {
	defer s.clearEvents()
	singleton := s.world.GetSingleton()

	player := s.world.GetPlayer()
	playerInput := singleton.PlayerInput[player.ID]

	commandList := playercommand.PlayerCommandList{Commands: []*playercommand.Wrapper{}}
	for _, e := range s.events {
		if cmdEvent, ok := e.(*events.PlayerCommandEvent); ok {
			commandList.Commands = append(commandList.Commands,
				cmdEvent.Command,
			)
		}
	}

	commandListBytes, err := proto.Marshal(&commandList)
	if err != nil {
		panic(err)
	}

	inputMessage := &knetwork.InputMessage{
		PlayerCommands: commandListBytes,
		CommandFrame:   singleton.CommandFrame,
		Input:          playerInput,
	}

	s.world.MetricsRegistry().Inc("newinput", 1)
	player.Client.SendMessage(knetwork.MessageTypeInput, inputMessage)
}

func (s *NetworkInputSystem) Name() string {
	return "NetworkInputSystem"
}
