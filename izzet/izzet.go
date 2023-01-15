package izzet

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/model"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Izzet struct {
	gameOver bool
	platform *input.SDLPlatform
	window   *sdl.Window

	// render properties
	fovY          float64
	aspectRatio   float64
	shaderManager *shaders.ShaderManager
	assetManager  *assets.AssetManager
	shadowMap     *ShadowMap
	imguiRenderer *ImguiOpenGL4Renderer

	colorPickingFB      uint32
	colorPickingTexture uint32

	redCircleFB        uint32
	redCircleTexture   uint32
	greenCircleFB      uint32
	greenCircleTexture uint32
	blueCircleFB       uint32
	blueCircleTexture  uint32

	camera *Camera

	entities map[int]*entities.Entity
	prefabs  map[int]*prefabs.Prefab

	viewerContext ViewerContext
}

func New(assetsDirectory, shaderDirectory string) *Izzet {
	g := &Izzet{}
	initSeed()

	window, err := initializeOpenGL()
	if err != nil {
		panic(err)
	}

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
	g.window = window
	g.shaderManager = shaders.NewShaderManager(shaderDirectory)
	g.assetManager = assets.NewAssetManager(assetsDirectory, true)

	imguiRenderer, err := NewImguiOpenGL4Renderer(imguiIO)
	if err != nil {
		panic(err)
	}
	g.imguiRenderer = imguiRenderer

	var data int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &data)

	// note(kevin) using exactly the max texture size sometimes causes initialization to fail.
	// so, I cap it at a fraction of the max
	settings.RuntimeMaxTextureSize = int(float32(data) * .90)

	shadowMap, err := NewShadowMap(settings.RuntimeMaxTextureSize, settings.RuntimeMaxTextureSize, far*shadowDistanceFactor)
	if err != nil {
		panic(fmt.Sprintf("failed to create shadow map %s", err))
	}
	g.shadowMap = shadowMap

	w, h := g.window.GetSize()
	g.colorPickingFB, g.colorPickingTexture = g.initFrameBuffer(int(w), int(h))
	g.redCircleFB, g.redCircleTexture = g.initFrameBuffer(1024, 1024)
	g.greenCircleFB, g.greenCircleTexture = g.initFrameBuffer(1024, 1024)
	g.blueCircleFB, g.blueCircleTexture = g.initFrameBuffer(1024, 1024)

	compileShaders(g.shaderManager)

	g.aspectRatio = float64(settings.Width) / float64(settings.Height)
	g.fovY = mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/g.aspectRatio))
	g.camera = &Camera{Position: mgl64.Vec3{0, 0, 300}, Orientation: mgl64.QuatIdent()}

	g.loadPrefabs()
	g.loadEntities()

	return g
}

func (g *Izzet) Start() {
	var accumulator float64
	var renderAccumulator float64

	msPerFrame := float64(1000) / float64(settings.FPS)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	// immediate updates when swapping buffers
	err := sdl.GLSetSwapInterval(0)
	if err != nil {
		panic(err)
	}

	frameCount := 0
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta
		renderAccumulator += delta

		runCount := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			input := g.platform.PollInput()
			g.HandleInput(input)
			g.runCommandFrame(input, time.Duration(settings.MSPerCommandFrame)*time.Millisecond)

			accumulator -= float64(settings.MSPerCommandFrame)
			runCount++
		}

		// prevents lighting my CPU on fire
		if accumulator < float64(settings.MSPerCommandFrame)-10 {
			time.Sleep(5 * time.Millisecond)
		}

		if renderAccumulator >= msPerFrame {
			frameCount++
			g.Render(time.Duration(msPerFrame) * time.Millisecond)

			g.window.GLSwap()
			renderAccumulator -= msPerFrame
		}
	}
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
	rand.Seed(seed)
}

func (g *Izzet) loadPrefabs() {
	modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}

	g.prefabs = map[int]*prefabs.Prefab{}

	names := []string{"alpha", "mutant", "scene", "town_center"}

	for _, name := range names {
		spec := g.assetManager.GetModel(name)
		m := model.NewModel(spec, modelConfig)
		m.InitializeRenderingProperties(*g.assetManager)

		pf := prefabs.CreatePrefab(name, []*model.Model{m})
		g.prefabs[pf.ID] = pf
	}
}

func (g *Izzet) loadEntities() {
	g.entities = map[int]*entities.Entity{}
	for _, pf := range g.Prefabs() {
		entity := entities.InstantiateFromPrefab(pf)
		g.entities[entity.ID] = entity
	}
}
