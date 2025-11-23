package server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/system"
	"github.com/kkevinchou/izzet/izzet/system/serversystems"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
)

type Server struct {
	gameOver bool

	assetManager *assets.AssetManager

	entities map[int]*entity.Entity

	world *world.GameWorld

	systems           []system.System
	appMode           types.AppMode
	collisionObserver *collisionobserver.CollisionObserver

	runtimeConfig *runtimeconfig.RuntimeConfig

	players map[int]*network.Player

	newConnections chan NewConnection

	commandFrame int
	inputBuffer  *inputbuffer.InputBuffer
	playerInput  map[int]input.Input
	eventManager *events.EventManager

	navMesh *navmesh.CompiledNavMesh

	projectName            string
	predictionDebugLogging bool
}

func NewWithFile(filepath string, projectName string) *Server {
	s := NewWithWorld(nil, projectName)
	world, err := serialization.ReadFromFile(filepath, s.assetManager)
	if err != nil {
		panic(err)
	}
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

	g.entities = map[int]*entity.Entity{}
	// g.setupAssets(data)
	g.collisionObserver = collisionobserver.NewCollisionObserver()

	g.newConnections = make(chan NewConnection, 100)

	g.systems = append(g.systems, serversystems.NewReceiverSystem(g))
	g.systems = append(g.systems, serversystems.NewInputSystem(g))
	g.systems = append(g.systems, serversystems.NewCharacterControllerSystem(g))
	g.systems = append(g.systems, serversystems.NewAISystemSystem(g))
	// g.systems = append(g.systems, system.NewPhysicsSystem(g))
	g.systems = append(g.systems, system.NewKinematicSystem(g))
	// g.systems = append(g.systems, system.NewCollisionSystem(g))
	g.systems = append(g.systems, system.NewCameraTargetSystem(g))
	g.systems = append(g.systems, serversystems.NewRulesSystem(g))
	g.systems = append(g.systems, system.NewAnimationSystem(g))
	g.systems = append(g.systems, system.NewCleanupSystem(g))
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
			globals.ServerRegistry().Inc("command_frame_nanoseconds", float64(commandFrameNanos))
			globals.ServerRegistry().Inc("command_frames", 1)
			g.world.IncrementCommandFrameCount()

			accumulator -= float64(settings.MSPerCommandFrame)
			currentLoopCommandFrames++
			if currentLoopCommandFrames > settings.MaxCommandFramesPerLoop {
				accumulator = 0
			}

			sleepTime := float64(settings.MSPerCommandFrame) - accumulator - 1
			if sleepTime >= 1 {
				time.Sleep(time.Duration(int64(sleepTime) * 1000000))
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
	world := world.New()
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
