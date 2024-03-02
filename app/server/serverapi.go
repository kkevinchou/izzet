package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/izzet/app/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
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
	world, err := serialization.ReadFromFile(filename)
	if err != nil {
		fmt.Println("failed to load world", filename, err)
		panic(err)
	}
	serialization.InitDeserializedEntities(world.Entities(), g.modelLibrary)

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
	c := network.NewClient(connection)
	g.players[playerID] = &network.Player{
		ID: playerID, Connection: connection,
		InMessageChannel:  inMessageChannel,
		OutMessageChannel: make(chan network.MessageTransport, 100),
		DisconnectChannel: disconnectChannel,
		Client:            c,
	}

	go func(client network.IzzetClient, id int, ch chan network.MessageTransport, discCh chan bool) {
		// f, err := os.OpenFile("serverlog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		// if err != nil {
		// 	panic(err)
		// }
		// defer f.Close()

		for {
			message, err := g.players[playerID].Client.Recv()
			if err != nil {
				fmt.Println(fmt.Errorf("error decoding message from player %d - %w", id, err))
				// f.Write([]byte(fmt.Sprintf("%s - %d - FAILED TO DECODE\n", time.Now().Format("2006-01-02 15:04:05"), message.CommandFrame)))
				if strings.Contains(err.Error(), "An existing connection was forcibly closed") ||
					strings.Contains(err.Error(), "An established connection was aborted by the software in your host machine") ||
					err == io.EOF {

					if err == io.EOF {
						fmt.Println("Got EOF from remote player", id)
					}
					fmt.Println("connection closed by remote player", id)
					client.Close()
					discCh <- true
					return
				}
				continue
			}
			// _, err = f.Write([]byte(fmt.Sprintf("%s - %d - %s\n", time.Now().Format("2006-01-02 15:04:05"), message.CommandFrame, string(message.Body))))
			// if err != nil {
			// 	fmt.Println("failed to write to server log")
			// }

			ch <- message
		}
	}(c, playerID, inMessageChannel, disconnectChannel)

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

func (g *Server) GetPlayerEntity() *entities.Entity {
	panic("wat")
}

func (g *Server) GetPlayerCamera() *entities.Entity {
	panic("wat")
}

func (g *Server) SerializeWorld() []byte {
	var buffer bytes.Buffer
	serialization.Write(g.world, &buffer)
	return buffer.Bytes()
}

func (g *Server) CollisionObserver() *collisionobserver.CollisionObserver {
	return nil
}

func (g *Server) EventsManager() *events.EventManager {
	return g.eventManager
}

func (g *Server) SystemNames() []string {
	var names []string
	for _, s := range g.systems {
		names = append(names, s.Name())
	}
	return names
}
