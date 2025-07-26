package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/modelspec"
)

func (g *Server) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}

func (g *Server) GetPlayers() map[int]*network.Player {
	return g.players
}

func (g *Server) RegisterPlayer(playerID int, connection net.Conn) *network.Player {
	inMessageChannel := make(chan network.MessageTransport, 100)
	disconnectChannel := make(chan bool, 1)
	c := network.NewClient(connection)
	g.inputBuffer.RegisterPlayer(playerID)
	g.players[playerID] = &network.Player{
		ID: playerID, Connection: connection,
		InMessageChannel:  inMessageChannel,
		OutMessageChannel: make(chan network.MessageTransport, 100),
		DisconnectChannel: disconnectChannel,
		Client:            c,
	}

	go func(client network.IzzetClient, id int, ch chan network.MessageTransport, discCh chan bool) {
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
			ch <- message
		}
	}(c, playerID, inMessageChannel, disconnectChannel)

	return g.players[playerID]
}

func (g *Server) DeregisterPlayer(playerID int) {
	g.inputBuffer.DeregisterPlayer(playerID)
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

func (g *Server) World() *world.GameWorld {
	return g.world
}

func (g *Server) AssetManager() *assets.AssetManager {
	return g.assetManager
}

func (g *Server) SetNavMesh(nm *navmesh.CompiledNavMesh) {
	g.navMesh = nm
}

func (g *Server) NavMesh() *navmesh.CompiledNavMesh {
	return g.navMesh
}

func (g *Server) CopyLoadedAnimations(
	animations map[string]map[string]*modelspec.AnimationSpec,
	joints map[string]map[int]*modelspec.JointSpec,
	rootJoints map[string]int,
) {
	g.assetManager.Animations = animations
	g.assetManager.Joints = joints
	g.assetManager.RootJoints = rootJoints
}

func (g *Server) ProjectName() string {
	return g.projectName
}

func (g *Server) AppMode() mode.AppMode {
	panic("app mode should not be called in server, conslidate app mode with isClient/isServer")
}

func (g *Server) RuntimeConfig() *runtimeconfig.RuntimeConfig {
	panic("server should not be accessing runtime config")
}

func (g *Server) PredictionDebugLogging() bool {
	return g.predictionDebugLogging
}
func (g *Server) SetPredictionDebugLogging(value bool) {
	g.predictionDebugLogging = value
}
