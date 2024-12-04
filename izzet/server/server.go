package server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/serversystems"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

type Server struct {
	gameOver bool

	assetManager *assets.AssetManager

	entities map[int]*entities.Entity

	metricsRegistry *metrics.MetricsRegistry

	world *world.GameWorld

	systems           []systems.System
	appMode           mode.AppMode
	collisionObserver *collisionobserver.CollisionObserver

	runtimeConfig *runtimeconfig.RuntimeConfig

	players map[int]*network.Player

	newConnections chan NewConnection

	commandFrame int
	inputBuffer  *inputbuffer.InputBuffer
	playerInput  map[int]input.Input
	eventManager *events.EventManager

	navMesh *navmesh.CompiledNavMesh

	projectName string
}

func NewWithFile(filepath string, projectName string) *Server {
	s := NewWithWorld(nil, projectName)
	world, err := serialization.ReadFromFile(filepath)
	if err != nil {
		panic(err)
	}
	serialization.InitDeserializedEntities(world.Entities(), s.assetManager)
	s.world = world
	return s
}

func NewWithWorld(world *world.GameWorld, projectName string) *Server {
	initSeed()
	g := &Server{
		players:      map[int]*network.Player{},
		inputBuffer:  inputbuffer.New(),
		playerInput:  map[int]input.Input{},
		eventManager: events.NewEventManager(),
		projectName:  projectName,
	}
	g.initSettings()

	g.assetManager = assets.NewAssetManager(false)

	start := time.Now()

	g.world = world

	fmt.Println(time.Since(start), "spatial partition done")

	g.entities = map[int]*entities.Entity{}
	// dataFilePath := "izzet_data.json"
	// data := izzetdata.LoadData(dataFilePath)
	// g.setupAssets(data)
	g.metricsRegistry = metrics.New()
	g.collisionObserver = collisionobserver.NewCollisionObserver()

	g.newConnections = make(chan NewConnection, 100)

	g.systems = append(g.systems, serversystems.NewReceiverSystem(g))
	g.systems = append(g.systems, serversystems.NewInputSystem(g))
	g.systems = append(g.systems, serversystems.NewCharacterControllerSystem(g))
	g.systems = append(g.systems, serversystems.NewAISystemSystem(g))
	g.systems = append(g.systems, systems.NewPhysicsSystem(g))
	g.systems = append(g.systems, systems.NewCollisionSystem(g))
	g.systems = append(g.systems, &systems.CameraTargetSystem{})
	g.systems = append(g.systems, serversystems.NewRulesSystem(g))
	g.systems = append(g.systems, systems.NewAnimationSystem(g))
	g.systems = append(g.systems, systems.NewCleanupSystem(g))
	g.systems = append(g.systems, serversystems.NewEventsSystem(g))
	g.systems = append(g.systems, serversystems.NewReplicationSystem(g))

	fmt.Println(time.Since(start), "to start up systems")

	return g
}

func (g *Server) Start(started chan bool, done chan bool) {
	listener, err := g.listen()
	if err != nil {
		panic(err)
	}

	started <- true
	var accumulator float64

	// msPerFrame := float64(1000) / float64(60)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta

		currentLoopCommandFrames := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			start := time.Now()
			g.runCommandFrame(time.Duration(settings.MSPerCommandFrame) * time.Millisecond)
			commandFrameNanos := time.Since(start).Nanoseconds()
			g.MetricsRegistry().Inc("command_frame_nanoseconds", float64(commandFrameNanos))
			g.MetricsRegistry().Inc("command_frames", 1)
			g.world.IncrementCommandFrameCount()

			accumulator -= float64(settings.MSPerCommandFrame)
			currentLoopCommandFrames++
			if currentLoopCommandFrames > settings.MaxCommandFramesPerLoop {
				accumulator = 0
			}
		}

		select {
		case <-done:
			listener.Close()
			return
		default:
			break
		}
	}
}

func New(shaderDirectory string, projectName string) *Server {
	world := world.New(map[int]*entities.Entity{})
	return NewWithWorld(world, projectName)
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
}

type NewConnection struct {
	PlayerID   int
	Connection net.Conn
}

func (s *Server) listen() (net.Listener, error) {
	host := "0.0.0.0"
	port := "7878"
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	fmt.Println("listening on " + host + ":" + port)

	go func() {
		playerIDGenerator := 100000
		for {
			conn, err := listener.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				fmt.Println("error accepting a connection on the listener:", err.Error())
				continue
			}

			id := playerIDGenerator
			playerIDGenerator += 1

			s.newConnections <- NewConnection{PlayerID: id, Connection: conn}
		}
	}()

	return listener, nil
}

func (g *Server) initSettings() {
	config := runtimeconfig.DefaultRuntimeConfig()
	g.runtimeConfig = &config
}
