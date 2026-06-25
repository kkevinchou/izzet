package client

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/client/edithistory"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/system"
	"github.com/kkevinchou/izzet/izzet/system/clientsystem"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
)

type Client struct {
	gameOver bool
	window   Window
	platform platforms.Platform
	client   network.IzzetClient

	assetManager *assets.AssetManager

	camera *editorcamera.Camera

	renderSystem *render.RenderSystem
	editHistory  *edithistory.EditHistory

	capturedMouseOrigin [2]int32
	captureMouse        bool

	editorWorld *world.GameWorld
	world       *world.GameWorld

	playModeSystems   []system.System
	editorModeSystems []system.System
	appMode           types.AppMode
	collisionObserver *collisionobserver.CollisionObserver
	stateBuffer       *clientsystem.StateBuffer

	runtimeConfig *runtimeconfig.RuntimeConfig

	playerID        int
	playerEntity    *entity.Entity
	playerCamera    *entity.Entity
	connection      net.Conn
	networkMessages chan network.MessageTransport
	commandFrame    int
	clientConnected bool

	commandFrameHistory *clientsystem.CommandFrameHistory
	asyncServerStarted  bool
	asyncServerDone     chan bool
	serverAddress       string

	frameInput  input.Input
	serverStats serverstats.ServerStats

	selectedEntity *entity.Entity

	project *Project

	navMesh                *navmesh.NavigationMesh
	predictionDebugLogging bool
}

func New(shaderDirectory string, config settings.Config, logsEnabled bool) *Client {
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

	g := &Client{
		asyncServerDone: make(chan bool),
		window:          window,
		appMode:         types.AppModeEditor,
		platform:        sdlPlatform,
		serverAddress:   config.ServerAddress,
	}

	if logsEnabled {
		logHandlerOptions := &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
					attr.Value = slog.StringValue(attr.Value.Time().Local().Format("15:04:05.000"))
				}
				return attr
			},
		}
		clientLog, err := os.OpenFile(filepath.Join(settings.LogDir, "client.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		iztlog.SetClientLogger(slog.New(slog.NewJSONHandler(clientLog, logHandlerOptions)), g.CommandFrame, g.GetPlayerID)
	} else {
		iztlog.SetClientLogger(slog.New(slog.DiscardHandler), g.CommandFrame, g.GetPlayerID)
	}

	g.assetManager = assets.NewAssetManager(true, iztlog.ClientLogger)
	g.runtimeConfig = runtimeconfig.DefaultRuntimeConfig()
	g.renderSystem = render.New(g, shaderDirectory, w, h)

	g.camera = &editorcamera.Camera{
		Position: settings.EditorCameraStartPosition,
		Rotation: mgl64.QuatIdent(),
	}

	if settings.StartupProject == "" {
		g.NewProject(settings.NewProjectName)
	} else {
		g.LoadProject(settings.StartupProject)
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

	// 1 - vsync on
	// 0 - vsync off
	err := sdl.GL_SetSwapInterval(1)
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

		numSimulatedFrames := 0
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
			globals.ClientRegistry().Inc("command_frame_nanoseconds", float64(time.Since(start).Nanoseconds()))
			globals.ClientRegistry().Inc("command_frames", 1)
			g.world.IncrementCommandFrameCount()
			commandFrameCountBeforeRender += 1

			accumulator -= float64(settings.MSPerCommandFrame)
			numSimulatedFrames++
			if numSimulatedFrames >= settings.MaxCommandFramesPerLoop {
				g.Logger().Info("ran into max command frames per loop", "max", settings.MaxCommandFramesPerLoop)
				accumulator = 0
			}
		}

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

		nextCommandFrameMs := float64(settings.MSPerCommandFrame) - accumulator
		nextRenderFrameMs := msPerFrame - renderAccumulator
		minNextFrameMs := min(nextCommandFrameMs, nextRenderFrameMs)
		if minNextFrameMs > 0 {
			sleepStart := time.Now()
			time.Sleep(time.Duration(minNextFrameMs * float64(time.Millisecond)))
			globals.ClientRegistry().Inc("client_sleep_nanoseconds", float64(time.Since(sleepStart).Nanoseconds()))
		}
	}
}

func (g *Client) render(delta time.Duration) {
	globals.ClientRegistry().Inc("fps", 1)

	start := time.Now()
	// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
	g.renderSystem.Render(delta)
	swapStart := time.Now()
	g.window.Swap()
	globals.ClientRegistry().Inc("render_cpu_swap", durationMilliseconds(swapStart))
	globals.ClientRegistry().Inc("renderer_cpu_time", durationMilliseconds(start))
}

func durationMilliseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / 1000000.0
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
}

func (g *Client) setupSystems() {
	g.playModeSystems = append(g.playModeSystems, clientsystem.NewReceiverSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystem.NewInputSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystem.NewCharacterControllerSystem(g))
	g.playModeSystems = append(g.playModeSystems, system.NewKinematicSystem(g))
	g.playModeSystems = append(g.playModeSystems, system.NewCharacterOrientationSystem(g))
	g.playModeSystems = append(g.playModeSystems, system.NewCameraSystem(g))
	g.playModeSystems = append(g.playModeSystems, system.NewCombatSystem(g))
	g.playModeSystems = append(g.playModeSystems, system.NewAnimationSystem(g))
	g.playModeSystems = append(g.playModeSystems, system.NewCleanupSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystem.NewPingSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystem.NewPostFrameSystem(g))

	g.editorModeSystems = append(g.editorModeSystems, system.NewAnimationSystem(g))
}

func (g *Client) mousePosToNearPlane(mousePosition mgl64.Vec2, width, height int) mgl64.Vec3 {
	x := mousePosition.X()
	y := mousePosition.Y()

	// -1 for the near plane
	ndcP := mgl64.Vec4{((x / float64(width)) - 0.5) * 2, ((y / float64(height)) - 0.5) * -2, -1, 1}
	nearPlanePos := g.renderSystem.CameraViewerContext().ViewMatrix.Inv().Mul4(g.renderSystem.CameraViewerContext().ProjectionMatrix.Inv()).Mul4x1(ndcP)
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func (g *Client) MouseCaptured() bool {
	return g.captureMouse
}

func (g *Client) SetMouseCaptured(capture bool) {
	g.captureMouse = capture
	g.platform.SetRelativeMouse(capture)
}

type Window interface {
	Minimized() bool
	WindowFocused() bool
	GetSize() (int, int)
	Swap()
}
