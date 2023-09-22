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
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/model"
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

	navigationMesh *navmesh.NavigationMesh

	data *Data
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
	data := loadData(dataFilePath)

	g.camera = &camera.Camera{
		Position: mgl64.Vec3{0, 150, 0},
		// Orientation: mgl64.QuatIdent(),
		Orientation: mgl64.QuatRotate(mgl64.DegToRad(-20), mgl64.Vec3{1, 0, 0}),
	}

	w, h := g.window.GetSize()
	g.width, g.height = int(w), int(h)
	g.renderer = render.New(g, shaderDirectory, g.width, g.height)
	g.spatialPartition = spatialpartition.NewSpatialPartition(200, 100)

	g.entities = map[int]*entities.Entity{}
	g.prefabs = map[int]*prefabs.Prefab{}
	g.loadPrefabs(data)
	g.loadEntities()
	g.serializer = serialization.New(g)
	g.editHistory = edithistory.New()
	// g.navigationMesh = navmesh.New(g)

	return g
}

func (g *Izzet) Start() {
	var accumulator float64
	var renderAccumulator float64
	var oneSecondAccumulator float64

	msPerFrame := float64(1000) / float64(settings.FPS)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	// immediate updates when swapping buffers
	err := sdl.GLSetSwapInterval(0)
	if err != nil {
		panic(err)
	}

	var renderTime float64
	var renderTimeSamples int

	frameCount := 0
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta
		renderAccumulator += delta
		oneSecondAccumulator += delta

		currentLoopCommandFrames := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			input := g.platform.PollInput()
			g.HandleInput(input)
			g.runCommandFrame(input, time.Duration(settings.MSPerCommandFrame)*time.Millisecond)
			g.commandFrameCount++

			accumulator -= float64(settings.MSPerCommandFrame)
			currentLoopCommandFrames++
			if currentLoopCommandFrames > settings.MaxCommandFramesPerLoop {
				accumulator = 0
			}
		}

		// prevents lighting my CPU on fire
		if accumulator < float64(settings.MSPerCommandFrame)-10 {
			time.Sleep(5 * time.Millisecond)
		}

		if oneSecondAccumulator >= 1000 {
			fps := float64(frameCount) / oneSecondAccumulator * 1000
			panels.DBG.FPS = fps
			frameCount = 0

			avgRenderTime := renderTime / float64(renderTimeSamples)
			renderTime = 0
			renderTimeSamples = 0
			panels.DBG.RenderTime = avgRenderTime

			oneSecondAccumulator = 0
		}

		if renderAccumulator >= msPerFrame {
			start := time.Now()
			frameCount++
			// g.renderer.PreRenderImgui()
			// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
			renderContext := render.NewRenderContext(g.width, g.height, float64(panels.DBG.FovX))
			g.renderer.Render(time.Duration(msPerFrame)*time.Millisecond, renderContext)
			g.window.GLSwap()
			renderAccumulator -= msPerFrame

			renderTime += float64(time.Since(start).Microseconds()) / 1000
			renderTimeSamples += 1
		}
	}
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
	rand.Seed(seed)
}

func (g *Izzet) loadPrefabs(data *Data) {
	modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}

	for _, entityAsset := range data.EntityAssets {
		name := entityAsset.Name
		// multipart := entityAsset.Multipart
		var pf *prefabs.Prefab

		modelGroup := g.assetManager.GetModelGroup(name)
		models := model.CreateModelsFromModelGroup(modelGroup, modelConfig)
		// if !multipart {
		// 	models = []*model.Model{models[0]}
		// }
		pf = prefabs.CreatePrefab(name, models)
		g.prefabs[pf.ID] = pf
	}
}

func (g *Izzet) loadEntities() {
	pointLightInfo0 := &entities.LightInfo{
		Diffuse: mgl64.Vec4{1, 1, 1, 8},
		Type:    1,
	}
	pointLight0 := entities.CreateLight(pointLightInfo0)
	pointLight0.Name = "point_light"
	entities.SetLocalPosition(pointLight0, mgl64.Vec3{0, 8, 765})
	g.AddEntity(pointLight0)

	pointLightInfo1 := &entities.LightInfo{
		Diffuse: mgl64.Vec4{1, 1, 1, 8},
		Type:    1,
	}
	pointLight1 := entities.CreateLight(pointLightInfo1)
	pointLight1.Name = "point_light"
	entities.SetLocalPosition(pointLight1, mgl64.Vec3{0, 60, 0})
	g.AddEntity(pointLight1)

	lightDir := panels.DBG.DirectionalLightDir
	dirLightInfo := &entities.LightInfo{
		Diffuse:   mgl64.Vec4{1, 1, 1, 1},
		Direction: mgl64.Vec3{float64(lightDir[0]), float64(lightDir[1]), float64(lightDir[2])}.Normalize(),
	}
	directionalLight := entities.CreateLight(dirLightInfo)
	directionalLight.Name = "directional_light"
	entities.SetLocalPosition(directionalLight, mgl64.Vec3{0, 300, 0})
	// directionalLight.Particles = entities.NewParticleGenerator(100)
	g.AddEntity(directionalLight)

	pfMap := map[string]*prefabs.Prefab{}

	for _, pf := range g.Prefabs() {
		pfMap[pf.Name] = pf
	}

	dungeonPF := pfMap["demo_scene_dungeon"]
	parent := entities.CreateDummy("scene_dummy")
	g.AddEntity(parent)
	entities.SetScale(parent, mgl64.Vec3{10, 10, 10})

	for _, entity := range entities.InstantiateFromPrefab(dungeonPF) {
		entities.BuildRelation(parent, entity)
		g.AddEntity(entity)
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
