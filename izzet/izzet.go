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
}

func New(assetsDirectory, shaderDirectory string) *Izzet {
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

	g.camera = &camera.Camera{
		Position:    mgl64.Vec3{0, 0, 0},
		Orientation: mgl64.QuatRotate(mgl64.DegToRad(90), mgl64.Vec3{0, 1, 0}).Mul(mgl64.QuatRotate(mgl64.DegToRad(-30), mgl64.Vec3{1, 0, 0})),
	}

	w, h := g.window.GetSize()
	g.width, g.height = int(w), int(h)
	g.renderer = render.New(g, shaderDirectory, g.width, g.height)

	g.entities = map[int]*entities.Entity{}
	g.prefabs = map[int]*prefabs.Prefab{}
	g.loadPrefabs()
	g.loadEntities()
	g.serializer = serialization.New(g)
	g.editHistory = edithistory.New()
	g.spatialPartition = spatialpartition.NewSpatialPartition(200, 10)

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

		for accumulator >= float64(settings.MSPerCommandFrame) {
			input := g.platform.PollInput()
			g.HandleInput(input)
			g.runCommandFrame(input, time.Duration(settings.MSPerCommandFrame)*time.Millisecond)
			g.commandFrameCount++

			accumulator -= float64(settings.MSPerCommandFrame)
		}

		// prevents lighting my CPU on fire
		if accumulator < float64(settings.MSPerCommandFrame)-10 {
			time.Sleep(5 * time.Millisecond)
		}

		if renderAccumulator >= msPerFrame {
			start := time.Now()
			frameCount++
			// g.renderer.PreRenderImgui()
			// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
			renderContext := render.NewRenderContext(g.width, g.height, settings.FovX)
			g.renderer.Render(time.Duration(msPerFrame)*time.Millisecond, renderContext)
			g.window.GLSwap()
			renderAccumulator -= msPerFrame
			panels.DBG.RenderTime = float64(time.Since(start).Microseconds()) / 1000
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

	names := []string{"vehicle", "alpha", "demo_scene_west", "demo_scene_dungeon", "broken_tree_mat", "lootbox"}

	for _, name := range names {
		var pf *prefabs.Prefab
		if name == "demo_scene_west" || name == "demo_scene_dungeon" || name == "demo_scene" || name == "lootbox" {
			collection := g.assetManager.GetCollection(name)
			ctx := model.CreateContext(collection)

			// m := model.NewModelFromCollectionAll(ctx, modelConfig)
			// pf := prefabs.CreatePrefab(name, []*model.Model{m})
			// g.prefabs[pf.ID] = pf

			models := model.NewModelsFromCollection(ctx, modelConfig)
			pf := prefabs.CreatePrefab(name, models)
			g.prefabs[pf.ID] = pf
		} else {
			collection := g.assetManager.GetCollection(name)
			ctx := model.CreateContext(collection)
			m := model.NewModelsFromCollection(ctx, modelConfig)[0]
			pf = prefabs.CreatePrefab(name, []*model.Model{m})
			g.prefabs[pf.ID] = pf
		}
	}
}

func (g *Izzet) loadEntities() {
	lightInfo2 := &entities.LightInfo{
		Diffuse: mgl64.Vec4{1, 1, 1, 8},
		Type:    1,
	}
	pointLight := entities.CreateLight(lightInfo2)
	pointLight.LocalPosition = mgl64.Vec3{0, -12, 806}
	g.AddEntity(pointLight)

	lightDir := panels.DBG.DirectionalLightDir
	lightInfo := &entities.LightInfo{
		Diffuse:   mgl64.Vec4{1, 1, 1, 5},
		Direction: mgl64.Vec3{float64(lightDir[0]), float64(lightDir[1]), float64(lightDir[2])}.Normalize(),
	}
	directionalLight := entities.CreateLight(lightInfo)
	directionalLight.LocalPosition = mgl64.Vec3{0, 100, 0}
	// directionalLight.Particles = entities.NewParticleGenerator(100)
	g.AddEntity(directionalLight)

	// cube := entities.CreateCube()
	// cube.LocalPosition = mgl64.Vec3{150, 150, 0}
	// cube.Physics = &entities.PhysicsComponent{Velocity: mgl64.Vec3{0, -10, 0}}

	// capsule := collider.NewCapsule(mgl64.Vec3{0, 10, 0}, mgl64.Vec3{0, 5, 0}, 5)
	// cube.Collider = &entities.ColliderComponent{
	// 	CapsuleCollider: &capsule,
	// }
	// g.AddEntity(cube)

	// cube2 := entities.CreateCube()
	// cube2.LocalPosition = mgl64.Vec3{150, 0, 0}
	// cube2.Physics = &entities.PhysicsComponent{}

	// capsule2 := collider.NewCapsule(mgl64.Vec3{0, 10, 0}, mgl64.Vec3{0, 5, 0}, 5)
	// cube2.Collider = &entities.ColliderComponent{
	// 	CapsuleCollider: &capsule2,
	// }
	// g.AddEntity(cube2)

	for _, pf := range g.Prefabs() {
		if pf.Name == "alpha" {
			// entity := entities.InstantiateFromPrefab(pf)
			// entity.AnimationPlayer.PlayAnimation("Cast2")
			// entity.AnimationPlayer.UpdateTo(0)
			// g.AddEntity(entity)

			// entity2 := entities.InstantiateFromPrefab(pf)
			// g.entities[entity2.ID] = entity2
			// if pf.Name == modelName {
			// 	entity2.LocalPosition = mgl64.Vec3{50, 0, 0}
			// }
		} else if pf.Name == "scene" {
			// entity := entities.InstantiateFromPrefab(pf)
			// g.AddEntity(entity)
		} else if pf.Name == "lootbox" {
			// entity := entities.InstantiateFromPrefab(pf)
			// g.entities[entity.ID] = entity
			// parent := g.GetEntityByID(0)
			// joint := parent.Model.ModelSpecification().JointMap[0]
			// entity.ParentJoint = joint
			// g.BuildRelation(parent, entity)
		} else if pf.Name == "demo_scene_dungeon" {
			parent := entities.CreateDummy("scene_dummy")
			g.AddEntity(parent)
			parent.Scale = mgl64.Vec3{10, 10, 10}

			for _, entity := range entities.InstantiateFromPrefab(pf) {
				// entity := entities.InstantiateFromPrefab(pf)
				// entity.Scale = entity.Scale.Mul(2)
				g.AddEntity(entity)
				g.BuildRelation(parent, entity)
			}
		} else if pf.Name == "vehicle" {
			// prefab := g.GetPrefabByID(pf.ID)
			// parent := entities.CreateDummy(prefab.Name)
			// g.AddEntity(parent)
			// for _, entity := range entities.InstantiateFromPrefab(prefab) {
			// 	g.AddEntity(entity)
			// 	g.BuildRelation(parent, entity)
			// }
			// parent.Scale = parent.Scale.Mul(15)
			// panels.SelectEntity(parent)
		}
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
