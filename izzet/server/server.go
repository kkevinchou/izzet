package server

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/izzetdata"
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
	"github.com/kkevinchou/kitolib/input"
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

	systems           []systems.System
	appMode           app.AppMode
	collisionObserver *observers.CollisionObserver

	settings *app.Settings

	players map[int]*network.Player

	newConnections chan NewConnection
	replicator     *Replicator

	commandFrame int
	inputBuffer  *inputbuffer.InputBuffer
	playerInput  map[int]input.Input
}

func NewWithFile(assetsDirectory string, filepath string) *Server {
	s := NewWithWorld(assetsDirectory, nil)
	world, err := s.serializer.ReadFromFile(filepath)
	if err != nil {
		panic(err)
	}
	s.world = world
	return s
}

func NewWithWorld(assetsDirectory string, world *world.GameWorld) *Server {
	initSeed()
	g := &Server{
		players:     map[int]*network.Player{},
		inputBuffer: inputbuffer.New(),
		playerInput: map[int]input.Input{},
	}
	g.initSettings()

	g.assetManager = assets.NewAssetManager(assetsDirectory, false)
	g.modelLibrary = modellibrary.New(false)

	start := time.Now()

	g.world = world

	fmt.Println(time.Since(start), "spatial partition done")

	g.entities = map[int]*entities.Entity{}
	dataFilePath := "izzet_data.json"
	data := izzetdata.LoadData(dataFilePath)
	g.setupAssets(g.assetManager, g.modelLibrary, data)
	g.serializer = serialization.New(g)
	g.metricsRegistry = metrics.New()
	g.collisionObserver = observers.NewCollisionObserver()

	g.newConnections = make(chan NewConnection, 100)
	g.replicator = NewReplicator(g, g.serializer)

	// THINGS TO DELETE AFTER DEBUGGING
	g.editHistory = edithistory.New()

	g.systems = append(g.systems, serversystems.NewReceiverSystem(g))
	g.systems = append(g.systems, serversystems.NewInputSystem(g))
	g.systems = append(g.systems, &systems.CameraTargetSystem{})
	g.systems = append(g.systems, serversystems.NewCharacterControllerSystem(g))
	g.systems = append(g.systems, &systems.MovementSystem{})
	g.systems = append(g.systems, systems.NewPhysicsSystem(g))
	g.systems = append(g.systems, systems.NewCollisionSystem(g, g.collisionObserver))
	g.systems = append(g.systems, systems.NewAnimationSystem(g))
	g.systems = append(g.systems, serversystems.NewSpawnerSystem(g))
	g.systems = append(g.systems, serversystems.NewEventsSystem(g, g.serializer))

	// g.setupEntities(data)
	// g.LoadWorld("cubes")
	// g.LoadWorld("multiplayer_test")

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

		select {
		case <-done:
			listener.Close()
			return
		default:
			break
		}
	}
}

func New(assetsDirectory, shaderDirectory string) *Server {
	world := world.New(map[int]*entities.Entity{})
	return NewWithWorld(assetsDirectory, world)
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

func (g *Server) setupAssets(assetManager *assets.AssetManager, modelLibrary *modellibrary.ModelLibrary, data *izzetdata.Data) {
	// docNames := []string{"demo_scene_city", "demo_scene_samurai", "alpha"}
	for docName, _ := range data.EntityAssets {
		doc := assetManager.GetDocument(docName)

		modelLibrary.RegisterDocument(doc, data)

		for _, mesh := range doc.Meshes {
			modelLibrary.RegisterMesh(doc.Name, mesh)
		}
		if len(doc.Animations) > 0 {
			modelLibrary.RegisterAnimations(docName, doc)
		}
	}
}
