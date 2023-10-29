package server

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

func (g *Server) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}

func (g *Server) LoadWorld(name string) bool {
	if name == "" {
		return false
	}

	filename := fmt.Sprintf("./%s.json", name)
	world, err := g.serializer.ReadFromFile(filename)
	if err != nil {
		fmt.Println("failed to load world", filename, err)
		panic(err)
	}

	g.editHistory.Clear()
	g.world.SpatialPartition().Clear()

	var maxID int
	for _, e := range world.Entities() {
		if e.ID > maxID {
			maxID = e.ID
		}
		g.entities[e.ID] = e
	}

	if len(g.entities) > 0 {
		entities.SetNextID(maxID + 1)
	}

	panels.SelectEntity(nil)
	g.SetWorld(world)
	return true
}

func (g *Server) SetWorld(world *world.GameWorld) {
	g.world = world
}

func (g *Server) ModelLibrary() *modellibrary.ModelLibrary {
	return g.modelLibrary
}

func (g *Server) GetPlayers() map[int]*network.Player {
	return g.players
}

func (g *Server) RegisterPlayer(playerID int, connection net.Conn) *network.Player {
	inMessageChannel := make(chan network.MessageTransport, 100)
	disconnectChannel := make(chan bool, 1)
	g.players[playerID] = &network.Player{
		ID: playerID, Connection: connection,
		InMessageChannel:  inMessageChannel,
		OutMessageChannel: make(chan network.MessageTransport, 100),
		DisconnectChannel: disconnectChannel,
		Client:            network.NewClient(connection),
	}

	go func(conn net.Conn, id int, ch chan network.MessageTransport, discCh chan bool) {
		for {
			decoder := json.NewDecoder(conn)
			var message network.MessageTransport
			err := decoder.Decode(&message)
			if err != nil {
				// reader := decoder.Buffered()
				// bytes, err2 := io.ReadAll(reader)
				// if err2 != nil {
				// 	fmt.Println(fmt.Errorf("error reading remaining bytes %w", err2))
				// }
				// fmt.Println(fmt.Errorf("error decoding message from player %w remaining buffered data: %s", err, string(bytes)))
				fmt.Println(fmt.Errorf("error decoding message from player %d - %w", id, err))
				if strings.Contains(err.Error(), "An existing connection was forcibly closed") {
					fmt.Println("connection closed by remote player", id)
					conn.Close()
					discCh <- true
					return
				}
				continue
			}

			if message.MessageType == network.MsgTypePlayerInput {
				ch <- message
			}
		}
	}(connection, playerID, inMessageChannel, disconnectChannel)

	return g.players[playerID]
}

func (g *Server) DeregisterPlayer(playerID int) {
	delete(g.players, playerID)
}

func (g *Server) CommandFrame() int {
	return g.commandFrame
}

func (g *Server) InputBuffer() *inputbuffer.InputBuffer {
	return g.inputBuffer
}

func (g *Server) GetPlayer(playerID int) *network.Player {
	return g.players[playerID]
}

func (g *Server) SetPlayerInput(playerID int, input input.Input) {
	g.playerInput[playerID] = input
}

func (g *Server) GetPlayerInput(playerID int) input.Input {
	return g.playerInput[playerID]
}

func (g *Server) IsServer() bool {
	return true
}

func (g *Server) IsClient() bool {
	return false
}
