package client

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/izzetdata"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/project"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/clientsystems"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Client struct {
	gameOver      bool
	window        *sdl.Window
	platform      *input.SDLPlatform
	width, height int
	client        network.IzzetClient

	assetManager *assets.AssetManager
	modelLibrary *modellibrary.ModelLibrary

	camera *editorcamera.Camera

	prefabs map[int]*prefabs.Prefab

	renderer    *render.Renderer
	serializer  *serialization.Serializer
	editHistory *edithistory.EditHistory

	relativeMouseOrigin [2]int32
	relativeMouseActive bool

	metricsRegistry *metrics.MetricsRegistry

	showImguiDemo bool

	editorWorld *world.GameWorld
	world       *world.GameWorld

	playModeSystems   []systems.System
	editorModeSystems []systems.System
	appMode           app.AppMode
	collisionObserver *observers.CollisionObserver
	stateBuffer       *clientsystems.StateBuffer

	runtimeConfig *app.RuntimeConfig

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
	project     *project.Project
}

func New(assetsDirectory, shaderDirectory, dataFilePath string, config settings.Config, defaultWorld string) *Client {
	initSeed()

	window, err := initializeOpenGL(config)
	if err != nil {
		panic(err)
	}

	err = ttf.Init()
	if err != nil {
		panic(err)
	}

	// prevent load screen flashbang
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	window.GLSwap()

	imgui.CreateContext(nil)
	imguiIO := imgui.CurrentIO()
	imgui.CurrentIO().Fonts().AddFontFromFileTTF("_assets/fonts/roboto-regular.ttf", 20)

	w, h := window.GetSize()

	g := &Client{
		asyncServerDone: make(chan bool),
		window:          window,
		appMode:         app.AppModeEditor,
		platform:        input.NewSDLPlatform(window, imguiIO),
		width:           int(w),
		height:          int(h),
		assetManager:    assets.NewAssetManager(assetsDirectory, true),
		modelLibrary:    modellibrary.New(true),
		world:           world.New(map[int]*entities.Entity{}),
		serverAddress:   config.ServerAddress,
		project:         project.NewProject(),
	}

	g.initSettings()
	g.renderer = render.New(g, g.world, shaderDirectory, g.width, g.height)

	data := izzetdata.LoadData(dataFilePath)
	g.setupAssets(g.assetManager, g.modelLibrary, data)
	g.setupPrefabs(data)

	g.initialize()
	if defaultWorld != "" {
		g.LoadProject(defaultWorld)
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

	frameCount := 0
	commandFrameCountBeforeRender := 0
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta
		renderAccumulator += delta

		currentLoopCommandFrames := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			input := g.platform.PollInput()
			g.HandleInput(input)
			start := time.Now()
			g.frameInput = input

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

		if g.window.GetFlags()&sdl.WINDOW_MINIMIZED > 0 {
			msPerFrame = float64(1000) / float64(30)
		} else {
			msPerFrame = float64(1000) / float64(settings.FPS)
		}

		if renderAccumulator >= msPerFrame {
			g.MetricsRegistry().Inc("fps", 1)

			g.RuntimeConfig().FPS = g.MetricsRegistry().GetOneSecondSum("fps")
			g.RuntimeConfig().CommandFrameTime = g.MetricsRegistry().GetOneSecondAverage("command_frame_nanoseconds") / 1000000
			g.RuntimeConfig().RenderTime = g.MetricsRegistry().GetOneSecondAverage("render_time")
			g.RuntimeConfig().CommandFramesPerRender = commandFrameCountBeforeRender
			commandFrameCountBeforeRender = 0

			start := time.Now()
			frameCount++
			// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
			renderContext := render.NewRenderContext(g.width, g.height, float64(g.RuntimeConfig().FovX))
			g.renderer.Render(time.Duration(msPerFrame)*time.Millisecond, renderContext)
			g.window.GLSwap()
			renderTime := time.Since(start).Milliseconds()
			g.MetricsRegistry().Inc("render_time", float64(renderTime))

			// don't try to accumulate time to point where we render back to back loop iterations.
			// it's unlikely the game state has changed much from one step to the next.
			for renderAccumulator > msPerFrame {
				renderAccumulator -= msPerFrame
			}
		}
	}
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
	rand.Seed(seed)
}

func (g *Client) setupAssets(assetManager *assets.AssetManager, modelLibrary *modellibrary.ModelLibrary, data *izzetdata.Data) {
	// docNames := []string{"demo_scene_city", "demo_scene_samurai", "alpha"}
	for docName, _ := range data.EntityAssets {
		doc := assetManager.GetDocument(docName)

		if entityAsset, ok := data.EntityAssets[docName]; ok {
			if entityAsset.SingleEntity {
				modelLibrary.RegisterSingleEntityDocument(doc)
			} else {
				for _, mesh := range doc.Meshes {
					modelLibrary.RegisterMesh(doc.Name, mesh)
				}
			}
			if len(doc.Animations) > 0 {
				modelLibrary.RegisterAnimations(docName, doc)
			}
		}
	}
}

func (g *Client) setupPrefabs(data *izzetdata.Data) {
	g.prefabs = map[int]*prefabs.Prefab{}
	for name, _ := range data.EntityAssets {
		document := g.assetManager.GetDocument(name)
		pf := prefabs.CreatePrefab(document, data)
		g.prefabs[pf.ID] = pf
	}
}

func (g *Client) setupSystems() {
	// input system depends on the camera system to update the camera rotation
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewReceiverSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewInputSystem(g))
	g.playModeSystems = append(g.playModeSystems, &systems.CameraTargetSystem{})
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewCharacterControllerSystem(g))
	g.playModeSystems = append(g.playModeSystems, systems.NewPhysicsSystem(g))
	g.playModeSystems = append(g.playModeSystems, systems.NewCollisionSystem(g, g.collisionObserver))
	g.playModeSystems = append(g.playModeSystems, systems.NewAnimationSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewPingSystem(g))
	g.playModeSystems = append(g.playModeSystems, clientsystems.NewPostFrameSystem(g))
}

func (g *Client) setupEntities(data *izzetdata.Data) {
	pointLight := entities.CreatePointLight()
	pointLight.Movement = &entities.MovementComponent{
		PatrolConfig: &entities.PatrolConfig{Points: []mgl64.Vec3{{0, 100, 0}, {0, 300, 0}}},
		Speed:        100,
	}
	pointLight.LightInfo.PreScaledIntensity = 6
	pointLight.LightInfo.Diffuse3F = [3]float32{0.77, 0.11, 0}
	entities.SetLocalPosition(pointLight, mgl64.Vec3{0, 100, 0})
	g.world.AddEntity(pointLight)

	cube := entities.CreateCube(g.modelLibrary, 50)
	entities.SetLocalPosition(cube, mgl64.Vec3{0, 100, 0})
	g.world.AddEntity(cube)

	directionalLight := entities.CreateDirectionalLight()
	directionalLight.Name = "directional_light"
	directionalLight.LightInfo.PreScaledIntensity = 8
	entities.SetLocalPosition(directionalLight, mgl64.Vec3{0, 500, 0})
	g.world.AddEntity(directionalLight)

	doc := g.assetManager.GetDocument("demo_scene_scificity")
	for _, e := range entities.CreateEntitiesFromDocument(doc, g.modelLibrary, data) {
		g.world.AddEntity(e)
	}
}

func initializeOpenGL(config settings.Config) (*sdl.Window, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, fmt.Errorf("failed to init SDL %s", err)
	}

	// Enable hints for multisampling which allows opengl to use the default
	// multisampling algorithms implemented by the OpenGL rasterizer
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)

	// sdl.GLSetAttribute(sdl.GL_RED_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_GREEN_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_BLUE_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 2)

	sdl.SetRelativeMouseMode(false)

	windowFlags := sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE
	if config.Fullscreen {
		dm, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			panic(err)
		}
		config.Width = int(dm.W)
		config.Height = int(dm.H)
		windowFlags |= sdl.WINDOW_MAXIMIZED
		// windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
		// windowFlags |= sdl.WINDOW_FULLSCREEN
	}

	window, err := sdl.CreateWindow("IZZET GAME ENGINE", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(config.Width), int32(config.Height), uint32(windowFlags))
	if err != nil {
		return nil, fmt.Errorf("failed to create window %s", err)
	}

	_, err = window.GLCreateContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context %s", err)
	}

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to init OpenGL %s", err)
	}

	fmt.Println("Open GL Version:", gl.GoStr(gl.GetString(gl.VERSION)))

	return window, nil
}

func (g *Client) mousePosToNearPlane(mouseInput input.MouseInput, width, height int) mgl64.Vec3 {
	x := mouseInput.Position.X()
	y := mouseInput.Position.Y()

	// -1 for the near plane
	ndcP := mgl64.Vec4{((x / float64(width)) - 0.5) * 2, ((y / float64(height)) - 0.5) * -2, -1, 1}
	nearPlanePos := g.renderer.CameraViewerContext().InverseViewMatrix.Inv().Mul4(g.renderer.CameraViewerContext().ProjectionMatrix.Inv()).Mul4x1(ndcP)
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func (g *Client) initSettings() {
	g.runtimeConfig = &app.RuntimeConfig{
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

		UIEnabled: true,
	}
}
