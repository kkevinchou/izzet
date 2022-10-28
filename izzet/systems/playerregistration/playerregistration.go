package playerregistration

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/network"
)

type World interface {
	CommandFrame() int
}

type PlayerRegistrationSystem struct {
	*base.BaseSystem

	world   World
	nserver *network.Server
}

func NewPlayerRegistrationSystem(world World, host, port, connectionType string) *PlayerRegistrationSystem {
	nserver := network.NewServer(host, port, connectionType, settings.ClientIDStart)
	nserver.Start()

	return &PlayerRegistrationSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
		nserver:    nserver,
	}
}

func (s *PlayerRegistrationSystem) Update(delta time.Duration) {
	d := directory.GetDirectory()
	playerManager := d.PlayerManager()

	incomingConnections := s.nserver.PullIncomingConnections()
	for _, incomingConnection := range incomingConnections {
		fmt.Println("New player connected with id", incomingConnection.ID)

		client := network.NewClient(settings.ServerID, incomingConnection.Connection)
		client.SetCommandFrameFunction(s.world.CommandFrame)

		var playerClient types.NetworkClient = client
		playerManager.RegisterPlayer(incomingConnection.ID, playerClient)
	}
}

func (s *PlayerRegistrationSystem) Name() string {
	return "NetworkListenerSystem"
}
