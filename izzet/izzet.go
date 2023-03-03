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
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Izzet struct {
	gameOver bool
	window   *sdl.Window
	platform *input.SDLPlatform

	assetManager *assets.AssetManager

	camera *camera.Camera

	entities map[int]*entities.Entity
	prefabs  map[int]*prefabs.Prefab

	renderer    *render.Renderer
	serializer  *serialization.Serializer
	editHistory *edithistory.EditHistory

	commandFrameCount int
}

func New(assetsDirectory, shaderDirectory string) *Izzet {
	initSeed()
	g := &Izzet{}
	window, err := initializeOpenGL()
	if err != nil {
		panic(err)
	}
	g.window = window

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
		Position:    mgl64.Vec3{250, 200, 300},
		Orientation: mgl64.QuatRotate(mgl64.DegToRad(90), mgl64.Vec3{0, 1, 0}).Mul(mgl64.QuatRotate(mgl64.DegToRad(-30), mgl64.Vec3{1, 0, 0})),
	}
	g.renderer = render.New(g, shaderDirectory)

	g.entities = map[int]*entities.Entity{}
	g.prefabs = map[int]*prefabs.Prefab{}
	g.loadPrefabs()
	g.loadEntities()
	g.serializer = serialization.New(g)
	g.editHistory = edithistory.New()

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
			frameCount++
			// g.renderer.PreRenderImgui()
			// todo - might have a bug here where a command frame hasn't run in this loop yet we'll call render here for imgui
			g.renderer.Render(time.Duration(msPerFrame) * time.Millisecond)
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

	names := []string{"vehicle", "alpha", "lootbox", "demo_scene_west"}

	for _, name := range names {
		var pf *prefabs.Prefab
		if name == "demo_scene_west" {
			collection := g.assetManager.GetCollection(name)
			ctx := model.CreateContext(collection)

			for i := 0; i < 1872; i++ {
				m := model.NewModelFromCollection(ctx, i, modelConfig)
				pf := prefabs.CreatePrefab(fmt.Sprintf("%s-%d", name, i), []*model.Model{m})
				g.prefabs[pf.ID] = pf
			}
		} else if name == "lootbox" {
			collection := g.assetManager.GetCollection(name)
			ctx := model.CreateContext(collection)

			for i := 0; i < 2; i++ {
				m := model.NewModelFromCollection(ctx, i, modelConfig)
				pf := prefabs.CreatePrefab(fmt.Sprintf("%s-%d", name, i), []*model.Model{m})
				g.prefabs[pf.ID] = pf
			}
		} else {
			collection := g.assetManager.GetCollection(name)
			ctx := model.CreateContext(collection)
			m := model.NewModelFromCollection(ctx, 0, modelConfig)
			pf = prefabs.CreatePrefab(name, []*model.Model{m})
			g.prefabs[pf.ID] = pf
		}
	}
}

func (g *Izzet) loadEntities() {
	// lightInfo2 := &entities.LightInfo{
	// 	Diffuse: mgl64.Vec4{1, 1, 1, 8000},
	// 	Type:    1,
	// }
	// pointLight := entities.CreateLight(lightInfo2)
	// pointLight.LocalPosition = mgl64.Vec3{0, 50, 402}
	// g.AddEntity(pointLight)

	lightInfo := &entities.LightInfo{
		Diffuse:   mgl64.Vec4{1, 1, 1, 5},
		Direction: mgl64.Vec3{float64(panels.DBG.DirectionalLightX), float64(panels.DBG.DirectionalLightY), float64(panels.DBG.DirectionalLightZ)}.Normalize(),
	}
	directionalLight := entities.CreateLight(lightInfo)
	directionalLight.LocalPosition = mgl64.Vec3{0, 200, 0}
	directionalLight.Particles = entities.NewParticleGenerator(100)
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
		} else {
			entity := entities.InstantiateFromPrefab(pf)
			g.entities[entity.ID] = entity
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

func (g *Izzet) mousePosToNearPlane(mouseInput input.MouseInput) mgl64.Vec3 {
	w, h := g.Window().GetSize()
	x := mouseInput.Position.X()
	y := mouseInput.Position.Y()

	// -1 for the near plane
	ndcP := mgl64.Vec4{((x / float64(w)) - 0.5) * 2, ((y / float64(h)) - 0.5) * -2, -1, 1}
	nearPlanePos := g.renderer.ViewerContext().InverseViewMatrix.Inv().Mul4(g.renderer.ViewerContext().ProjectionMatrix.Inv()).Mul4x1(ndcP)
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}
