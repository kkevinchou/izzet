package client

import (
	"fmt"
	"net"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/client/edithistory"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/clientsystems"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Client struct {
	gameOver      bool
	window        Window
	platform      platforms.Platform
	width, height int
	client        network.IzzetClient

	assetManager *assets.AssetManager

	camera *editorcamera.Camera

	prefabs map[int]*prefabs.Prefab

	renderSystem *render.RenderSystem
	editHistory  *edithistory.EditHistory

	relativeMouseOrigin [2]int32
	relativeMouseActive bool

	metricsRegistry *metrics.MetricsRegistry

	editorWorld *world.GameWorld
	world       *world.GameWorld

	playModeSystems   []systems.System
	editorModeSystems []systems.System
	appMode           mode.AppMode
	collisionObserver *collisionobserver.CollisionObserver
	stateBuffer       *clientsystems.StateBuffer

	runtimeConfig *runtimeconfig.RuntimeConfig

	playerID        int
	playerEntity    *entities.Entity
	playerCamera    *entities.Entity
	connection      net.Conn
	networkMessages chan network.MessageTransport
	commandFrame    int
	clientConnected bool

	commandFrameHistory *clientsystems.CommandFrameHistory
	asyncServerStarted  bool
	asyncServerDone     chan bool
	serverAddress       string

	frameInput  input.Input
	serverStats serverstats.ServerStats

	selectedEntity *entities.Entity

	project *Project

	navMesh                *navmesh.NavigationMesh
	predictionDebugLogging bool
}

func New(shaderDirectory string, config settings.Config, projectName string) *Client {
	initSeed()

	sdlPlatform, window, err := platforms.NewSDLPlatform(config.Width, config.Height, config.Fullscreen)
	if err != nil {
		panic(err)
	}

	if err := gl.Init(); err != nil {
		panic(fmt.Errorf("failed to init OpenGL %s", err))
	}
	fmt.Println("Open GL Version:", gl.GoStr(gl.GetString(gl.VERSION)))

	err = ttf.Init()
	if err != nil {
		panic(err)
	}

	// prevent load screen flashbang
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	window.Swap()

	w, h := window.GetSize()

	metricsRegistry := metrics.New()
	globals.SetClientMetricsRegistry(metricsRegistry)

	assetManager := assets.NewAssetManager(true)

	g := &Client{
		asyncServerDone: make(chan bool),
		window:          window,
		appMode:         mode.AppModeEditor,
		platform:        sdlPlatform,
		width:           w,
		height:          h,
		assetManager:    assetManager,
		serverAddress:   config.ServerAddress,
		metricsRegistry: metricsRegistry,
	}
	g.ResetWorld()
	g.initSettings()
	g.renderSystem = render.New(g, shaderDirectory, g.width, g.height)

	g.initialize()
	if projectName != "" {
		g.LoadProject(projectName)
	} else {
		g.CreateAndLoadEmptyProject()
	}

	g.setupSystems()

	return g
}

func (g *Client) Start() {
	var accumulator float64
	var renderAccumulator float64
	// var oneSecondAccumulator float64

	msPerFrame := float64(1000) / float64(settings.FPS)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	// immediate updates when swapping buffers
	err := sdl.GLSetSwapInterval(0)
	if err != nil {
		panic(err)
	}

	commandFrameCountBeforeRender := 0
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta
		renderAccumulator += delta

		currentLoopCommandFrames := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			g.platform.NewFrame()
			inputCollector := input.NewInputCollector()
			g.platform.ProcessEvents(inputCollector)
			if g.platform.ShouldStop() {
				g.Shutdown()
			}

			start := time.Now()
			g.frameInput = inputCollector.GetInput()

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

		sleepStart := time.Now()
		time.Sleep(2 * time.Millisecond)
		g.MetricsRegistry().Inc("render_sleep", float64(time.Since(sleepStart).Milliseconds()))

		if g.RuntimeConfig().LockRenderingToCommandFrameRate {
			msPerFrame = float64(settings.MSPerCommandFrame)
		} else {
			msPerFrame = float64(1000) / float64(settings.FPS)
		}

		if renderAccumulator >= msPerFrame {
			delta := time.Duration(msPerFrame) * time.Millisecond
			g.render(delta)
			// don't try to accumulate time to point where we render back to back loop iterations.
			// it's unlikely the game state has changed much from one step to the next.
			for renderAccumulator > msPerFrame {
				renderAccumulator -= msPerFrame
			}
		}
	}
}

func (g *Client) render(delta time.Duration) {
	g.MetricsRegistry().Inc("fps", 1)

	start := time.Now()
	// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
	g.renderSystem.Render(delta)
	g.MetricsRegistry().Inc("render_time", float64(time.Since(start).Milliseconds()))
	start = time.Now()
	g.window.Swap()
	g.MetricsRegistry().Inc("render_swap", float64(time.Since(start).Milliseconds()))

}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
}

func (g *Client) setupSystems() {
	// input system depends on the camera system to update the camera rotation
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewReceiverSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewInputSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewCharacterControllerSystem(g))
	g.playModeSystems = append(g.playModeSystems, systems.NewKinematicSystem(g))
	// g.playModeSystems = append(g.playModeSystems, systems.NewPhysicsSystem(g))
	// g.playModeSystems = append(g.playModeSystems, systems.NewCollisionSystem(g))
	g.playModeSystems = append(g.playModeSystems, &systems.CameraTargetSystem{})
	g.playModeSystems = append(g.playModeSystems, systems.NewAnimationSystem(g))
	g.playModeSystems = append(g.playModeSystems, systems.NewCleanupSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewPingSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewPostFrameSystem(g))

	g.editorModeSystems = append(g.editorModeSystems, systems.NewAnimationSystem(g))
}

func (g *Client) setupEntities() {
	pointLight := entities.CreatePointLight()
	pointLight.AIComponent = &entities.AIComponent{
		PatrolConfig: &entities.PatrolConfig{Points: []mgl64.Vec3{{0, 100, 0}, {0, 300, 0}}},
		Speed:        100,
	}
	pointLight.LightInfo.PreScaledIntensity = 0.05
	pointLight.LightInfo.Diffuse3F = [3]float32{0.77, 0.11, 0}
	entities.SetLocalPosition(pointLight, mgl64.Vec3{0, 100, 0})
	g.world.AddEntity(pointLight)

	cube := entities.CreateCube(g.assetManager, 50)
	entities.SetLocalPosition(cube, mgl64.Vec3{0, 100, 0})
	g.world.AddEntity(cube)

	directionalLight := entities.CreateDirectionalLight()
	directionalLight.Name = "directional_light"
	directionalLight.LightInfo.PreScaledIntensity = 0.1
	entities.SetLocalPosition(directionalLight, mgl64.Vec3{0, 500, 0})
	g.world.AddEntity(directionalLight)
}

func (g *Client) mousePosToNearPlane(mousePosition mgl64.Vec2, width, height int) mgl64.Vec3 {
	x := mousePosition.X()
	y := mousePosition.Y()

	// -1 for the near plane
	ndcP := mgl64.Vec4{((x / float64(width)) - 0.5) * 2, ((y / float64(height)) - 0.5) * -2, -1, 1}
	nearPlanePos := g.renderSystem.CameraViewerContext().InverseViewMatrix.Inv().Mul4(g.renderSystem.CameraViewerContext().ProjectionMatrix.Inv()).Mul4x1(ndcP)
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func (g *Client) initSettings() {
	config := runtimeconfig.DefaultRuntimeConfig()
	g.runtimeConfig = &config
}

type Window interface {
	Minimized() bool
	WindowFocused() bool
	GetSize() (int, int)
	Swap()
}
