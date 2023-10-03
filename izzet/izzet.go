package izzet

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/other"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Izzet struct {
	gameOver      bool
	window        *sdl.Window
	platform      *input.SDLPlatform
	width, height int

	assetManager *assets.AssetManager
	modelLibrary *modellibrary.ModelLibrary

	camera *camera.Camera

	entities map[int]*entities.Entity
	prefabs  map[int]*prefabs.Prefab

	renderer    *render.Renderer
	serializer  *serialization.Serializer
	editHistory *edithistory.EditHistory

	commandFrameCount   int
	spatialPartition    *spatialpartition.SpatialPartition
	relativeMouseOrigin [2]int32
	relativeMouseActive bool

	navigationMesh  *navmesh.NavigationMesh
	metricsRegistry *metrics.MetricsRegistry

	data          *Data
	showImguiDemo bool
}

func New(assetsDirectory, shaderDirectory, dataFilePath string) *Izzet {
	initSeed()
	g := &Izzet{}
	window, err := initializeOpenGL()
	if err != nil {
		panic(err)
	}
	g.window = window

	// prevent load screen flashbang
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	g.window.GLSwap()

	err = ttf.Init()
	if err != nil {
		panic(err)
	}

	imgui.CreateContext(nil)
	imguiIO := imgui.CurrentIO()
	// imgui.CurrentIO().Fonts().AddFontFromFileTTF("_assets/fonts/robotomono-regular.ttf", 20)
	// imgui.CurrentIO().Fonts().AddFontFromFileTTF("_assets/fonts/helvetica.ttf", 20)
	imgui.CurrentIO().Fonts().AddFontFromFileTTF("_assets/fonts/roboto-regular.ttf", 20)
	g.platform = input.NewSDLPlatform(window, imguiIO)
	g.assetManager = assets.NewAssetManager(assetsDirectory, true)
	g.modelLibrary = modellibrary.New()
	data := loadData(dataFilePath)

	g.camera = &camera.Camera{
		Position:    mgl64.Vec3{-82, 230, 95},
		Orientation: mgl64.QuatIdent(),
	}

	start := time.Now()

	w, h := g.window.GetSize()
	g.width, g.height = int(w), int(h)
	g.renderer = render.New(g, shaderDirectory, g.width, g.height)
	g.spatialPartition = spatialpartition.NewSpatialPartition(200, 25)

	fmt.Println(time.Since(start), "spatial partition done")

	g.entities = map[int]*entities.Entity{}
	g.prefabs = map[int]*prefabs.Prefab{}
	g.setupAssets(g.assetManager, g.modelLibrary)
	g.setupPrefabs(data)
	fmt.Println(time.Since(start), "prefabs done")
	g.setupEntities()
	fmt.Println(time.Since(start), "entities done")
	g.serializer = serialization.New(g)
	g.editHistory = edithistory.New()
	// g.navigationMesh = navmesh.New(g)
	g.metricsRegistry = metrics.New()

	fmt.Println(time.Since(start), "to start up systems")

	return g
}

func (g *Izzet) Start() {
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
			g.runCommandFrame(input, time.Duration(settings.MSPerCommandFrame)*time.Millisecond)
			commandFrameNanos := time.Since(start).Nanoseconds()
			g.MetricsRegistry().Inc("command_frame_nanoseconds", float64(commandFrameNanos))
			g.commandFrameCount++
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

			panels.DBG.FPS = g.MetricsRegistry().GetOneSecondSum("fps")
			panels.DBG.CommandFrameTime = g.MetricsRegistry().GetOneSecondAverage("command_frame_nanoseconds") / 1000000
			panels.DBG.RenderTime = g.MetricsRegistry().GetOneSecondAverage("render_time")
			panels.DBG.CommandFramesPerRender = commandFrameCountBeforeRender
			commandFrameCountBeforeRender = 0

			start := time.Now()
			frameCount++
			// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
			renderContext := render.NewRenderContext(g.width, g.height, float64(panels.DBG.FovX))
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

func (g *Izzet) setupAssets(assetManager *assets.AssetManager, modelLibrary *modellibrary.ModelLibrary) {
	doc := assetManager.GetDocument("demo_scene_samurai")
	for _, mesh := range doc.Meshes {
		modelLibrary.Register(doc.Name, mesh)
	}
}

func (g *Izzet) setupPrefabs(data *Data) {
	for _, entityAsset := range data.EntityAssets {
		name := entityAsset.Name

		document := g.assetManager.GetDocument(name)
		pf := prefabs.CreatePrefab(name, document)
		g.prefabs[pf.ID] = pf
	}
}

func (g *Izzet) setupEntities() {
	pointLight := entities.CreatePointLight()
	pointLight.Movement = &entities.MovementComponent{
		PatrolConfig: &entities.PatrolConfig{Points: []mgl64.Vec3{{0, 100, 0}, {0, 300, 0}}},
		Speed:        100,
	}
	pointLight.LightInfo.PreScaledIntensity = 6
	pointLight.LightInfo.Diffuse3F = [3]float32{0.77, 0.11, 0}
	entities.SetLocalPosition(pointLight, mgl64.Vec3{0, 100, 0})
	g.AddEntity(pointLight)

	cube := entities.CreateCube(50)
	entities.SetLocalPosition(cube, mgl64.Vec3{154, 126, -22})
	g.AddEntity(cube)

	directionalLight := entities.CreateDirectionalLight()
	directionalLight.Name = "directional_light"
	directionalLight.LightInfo.PreScaledIntensity = 8
	entities.SetLocalPosition(directionalLight, mgl64.Vec3{0, 500, 0})
	g.AddEntity(directionalLight)

	doc := g.assetManager.GetDocument("demo_scene_samurai")

	parent := entities.InstantiateEntity("scene_parent")
	g.AddEntity(parent)
	entities.SetScale(parent, mgl64.Vec3{20, 20, 20})

	for _, e := range other.CreateEntitiesFromScene(doc) {
		g.AddEntity(e)
		entities.BuildRelation(parent, e)
	}
}

func initializeOpenGL() (*sdl.Window, error) {
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
	sdl.SetRelativeMouseMode(false)

	windowFlags := sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE
	if settings.Fullscreen {
		dm, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			panic(err)
		}
		settings.Width = int(dm.W)
		settings.Height = int(dm.H)
		windowFlags |= sdl.WINDOW_MAXIMIZED
		// windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
		// windowFlags |= sdl.WINDOW_FULLSCREEN
	}

	window, err := sdl.CreateWindow("IZZET GAME ENGINE", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(settings.Width), int32(settings.Height), uint32(windowFlags))
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

func (g *Izzet) mousePosToNearPlane(mouseInput input.MouseInput, width, height int) mgl64.Vec3 {
	x := mouseInput.Position.X()
	y := mouseInput.Position.Y()

	// -1 for the near plane
	ndcP := mgl64.Vec4{((x / float64(width)) - 0.5) * 2, ((y / float64(height)) - 0.5) * -2, -1, 1}
	nearPlanePos := g.renderer.ViewerContext().InverseViewMatrix.Inv().Mul4(g.renderer.ViewerContext().ProjectionMatrix.Inv()).Mul4x1(ndcP)
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}
