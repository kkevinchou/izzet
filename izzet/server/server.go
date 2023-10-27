package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/serversystems"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/metrics"
)

type Server struct {
	gameOver bool

	assetManager *assets.AssetManager
	modelLibrary *modellibrary.ModelLibrary

	entities map[int]*entities.Entity

	serializer  *serialization.Serializer
	editHistory *edithistory.EditHistory

	metricsRegistry *metrics.MetricsRegistry

	world *world.GameWorld

	systems         []systems.System
	appMode         app.AppMode
	physicsObserver *observers.PhysicsObserver

	settings *app.Settings

	players map[int]network.Player

	newConnections chan NewConnection
	replicator     *Replicator

	commandFrame int
	inputBuffer  *inputbuffer.InputBuffer
}

func New(assetsDirectory, shaderDirectory, dataFilePath string) *Server {
	initSeed()
	g := &Server{
		players:     map[int]network.Player{},
		inputBuffer: inputbuffer.New(),
	}
	g.initSettings()

	g.assetManager = assets.NewAssetManager(assetsDirectory, false)
	g.modelLibrary = modellibrary.New(false)

	start := time.Now()

	g.world = world.New(map[int]*entities.Entity{})

	fmt.Println(time.Since(start), "spatial partition done")

	g.entities = map[int]*entities.Entity{}
	// data := izzetdata.LoadData(dataFilePath)
	// g.setupAssets(g.assetManager, g.modelLibrary, data)
	g.serializer = serialization.New(g, g.world)
	g.metricsRegistry = metrics.New()
	g.physicsObserver = observers.NewPhysicsObserver()

	g.newConnections = make(chan NewConnection, 100)
	g.replicator = NewReplicator(g, g.serializer)

	// THINGS TO DELETE AFTER DEBUGGING
	g.editHistory = edithistory.New()

	g.systems = append(g.systems, serversystems.NewReceiverSystem(g))
	g.systems = append(g.systems, &systems.CameraTargetSystem{})
	g.systems = append(g.systems, serversystems.NewCharacterControllerSystem(g))
	g.systems = append(g.systems, &systems.MovementSystem{})
	g.systems = append(g.systems, &systems.PhysicsSystem{Observer: g.physicsObserver})
	g.systems = append(g.systems, serversystems.NewEventsSystem(g, g.serializer))

	// g.setupEntities(data)
	// g.LoadWorld("cubes")
	g.LoadWorld("multiplayer_test")

	fmt.Println(time.Since(start), "to start up systems")

	return g
}

func (g *Server) Start() {
	g.listen()
	var accumulator float64

	// msPerFrame := float64(1000) / float64(60)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	commandFrameCountBeforeRender := 0
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
			g.world.IncrementCommandFrameCount()
			commandFrameCountBeforeRender += 1

			accumulator -= float64(settings.MSPerCommandFrame)
			currentLoopCommandFrames++
			if currentLoopCommandFrames > settings.MaxCommandFramesPerLoop {
				accumulator = 0
			}
		}
	}
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
	rand.Seed(seed)
}

func (g *Server) initSettings() {
	g.settings = &app.Settings{
		DirectionalLightDir:    [3]float32{-1, -1, -1},
		Roughness:              0.55,
		Metallic:               1.0,
		PointLightBias:         1,
		MaterialOverride:       false,
		EnableShadowMapping:    true,
		ShadowFarFactor:        1,
		SPNearPlaneOffset:      300,
		BloomIntensity:         0.04,
		Exposure:               1.0,
		AmbientFactor:          0.1,
		Bloom:                  true,
		BloomThresholdPasses:   1,
		BloomThreshold:         0.8,
		BloomUpsamplingScale:   1.0,
		Color:                  [3]float32{1, 1, 1},
		ColorIntensity:         20.0,
		RenderSpatialPartition: false,
		EnableSpatialPartition: true,
		FPS:                    0,

		Near: 1,
		Far:  3000,
		FovX: 105,

		FogStart:   200,
		FogEnd:     1000,
		FogDensity: 1,
		FogEnabled: true,

		TriangleDrawCount: 0,
		DrawCount:         0,

		NavMeshHSV:                    true,
		NavMeshRegionIDThreshold:      3000,
		NavMeshDistanceFieldThreshold: 23,
		HSVOffset:                     11,
		VoxelHighlightX:               0,
		VoxelHighlightZ:               0,
		VoxelHighlightDistanceField:   -1,
		VoxelHighlightRegionID:        -1,
	}
}

type NewConnection struct {
	PlayerID   int
	Connection net.Conn
}

func (s *Server) listen() {
	host := "0.0.0.0"
	port := "7878"
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		panic(err)
	}

	fmt.Println("listening on " + host + ":" + port)

	go func() {
		defer listener.Close()
		playerIDGenerator := 100000
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("error accepting a connection on the listener:", err.Error())
				continue
			}

			id := playerIDGenerator
			playerIDGenerator += 1
			encoder := json.NewEncoder(conn)
			err = encoder.Encode(id)
			if err != nil {
				fmt.Println("error with incoming message: %w", err)
			}

			s.newConnections <- NewConnection{PlayerID: id, Connection: conn}
		}
	}()
}
