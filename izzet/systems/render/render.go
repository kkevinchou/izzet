package render

import (
	"fmt"
	_ "image/png"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/libutils"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fovx float64 = 105
	near float64 = 1
	far  float64 = 3000

	// shadow map parameters
	shadowDistanceFactor float64 = .4 // proportion of view fustrum to include in shadow cuboid
	shadowmapZOffset             = 400
)

type World interface {
	GetSingleton() *singleton.Singleton
	GetEntityByID(id int) entities.Entity
	GetPlayerEntity() entities.Entity
	MetricsRegistry() *metrics.MetricsRegistry
	QueryEntity(componentFlags int) []entities.Entity
	CommandFrame() int
	SetFocusedWindow(focusedWindow types.Window)
	GetFocusedWindow() types.Window
	GetWindowVisibility(types.Window) bool
	GetEventBroker() eventbroker.EventBroker
	SpatialPartition() *spatialpartition.SpatialPartition
	ServerStats() map[string]string
}

type Platform interface {
	NewFrame()
	DisplaySize() [2]float32
	FramebufferSize() [2]float32
}

type RenderSystem struct {
	*base.BaseSystem
	window    *sdl.Window
	world     World
	skybox    *SkyBox
	floor     *Quad
	shadowMap *ShadowMap

	width       int
	height      int
	aspectRatio float64
	fovY        float64

	imguiRenderer *ImguiOpenGL4Renderer
	platform      Platform

	entities []entities.Entity
	events   []events.Event

	timeSoFar time.Duration
}

func init() {
	err := ttf.Init()
	if err != nil {
		panic(err)
	}
}

func NewRenderSystem(world World, window *sdl.Window, platform Platform, imguiIO imgui.IO, width, height, shadowMapDimension int) *RenderSystem {
	// setting swap interval to 1 locked the framerate to 60 fps on my pc.
	// something to do with screen tearing / vsync? i dunno
	sdl.GLSetSwapInterval(1)
	gl.ClearColor(1.0, 0.5, 0.5, 0.0)
	gl.ClearDepth(1)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
	gl.Enable(gl.MULTISAMPLE)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
	gl.Disable(gl.FRAMEBUFFER_SRGB)

	aspectRatio := float64(width) / float64(height)
	shadowMap, err := NewShadowMap(shadowMapDimension, shadowMapDimension, far*shadowDistanceFactor)
	if err != nil {
		panic(fmt.Sprintf("failed to create shadow map %s", err))
	}

	imguiRenderer, err := NewImguiOpenGL4Renderer(imguiIO)
	if err != nil {
		panic(err)
	}

	renderSystem := RenderSystem{
		BaseSystem: &base.BaseSystem{},
		window:     window,
		world:      world,
		skybox:     NewSkyBox(float32(far)),
		floor:      NewQuad(quadZeroY),
		shadowMap:  shadowMap,

		width:       width,
		height:      height,
		aspectRatio: aspectRatio,
		fovY:        mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/aspectRatio)),

		platform:      platform,
		imguiRenderer: imguiRenderer,
	}

	eventBroker := world.GetEventBroker()
	eventBroker.AddObserver(&renderSystem, []events.EventType{
		events.EventTypeConsoleEnabled,
	})

	return &renderSystem
}

func (s *RenderSystem) Observe(event events.Event) {
	if event.Type() == events.EventTypeConsoleEnabled {
		s.events = append(s.events, event)
	}
}

func (s *RenderSystem) clearEvents() {
	s.events = nil
}

func (s *RenderSystem) GetCameraTransform() *components.TransformComponent {
	singleton := s.world.GetSingleton()
	if singleton.CameraID == 0 {
		return nil
	}
	camera := s.world.GetEntityByID(singleton.CameraID)
	if camera == nil {
		fmt.Printf("render syste could not find camera with entity id %d\n", singleton.CameraID)
		return nil
	}
	componentContainer := camera.GetComponentContainer()
	return componentContainer.TransformComponent
}

func (s *RenderSystem) Render(delta time.Duration) {
	s.timeSoFar += delta
	defer s.clearEvents()

	transformComponent := s.GetCameraTransform()
	if transformComponent == nil {
		return
	}

	// configure camera viewer context
	viewerViewMatrix := transformComponent.Orientation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(transformComponent.Position.X(), transformComponent.Position.Y(), transformComponent.Position.Z())

	cameraViewerContext := ViewerContext{
		Position:    transformComponent.Position,
		Orientation: transformComponent.Orientation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(s.fovY), s.aspectRatio, near, far),
	}

	// configure light viewer context
	modelSpaceFrustumPoints := CalculateFrustumPoints(transformComponent.Position, transformComponent.Orientation, near, far, s.fovY, s.aspectRatio, shadowDistanceFactor)

	lightOrientation := libutils.Vec3ToQuat(mgl64.Vec3{-1, -1, -1})
	// degrees := float64(s.timeSoFar.Seconds()) * 5
	// lightOrientation = mgl64.QuatRotate(mgl64.DegToRad(degrees), mgl64.Vec3{1, 0, 0}).Mul(lightOrientation)

	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightOrientation.Mat4(), modelSpaceFrustumPoints, shadowmapZOffset)
	lightViewMatrix := mgl64.Translate3D(lightPosition.X(), lightPosition.Y(), lightPosition.Z()).Mul4(lightOrientation.Mat4()).Inv()

	lightViewerContext := ViewerContext{
		Position:          lightPosition,
		Orientation:       lightOrientation,
		InverseViewMatrix: lightViewMatrix,
		ProjectionMatrix:  lightProjectionMatrix,
	}

	lightContext := LightContext{
		DirectionalLightDir: lightOrientation.Rotate(mgl64.Vec3{0, 0, -1}),
		// this should be the inverse of the transforms applied to the viewer context
		// if the viewer moves along -y, the universe moves along +y
		LightSpaceMatrix: lightProjectionMatrix.Mul4(lightViewMatrix),
	}

	s.renderToDepthMap(lightViewerContext, lightContext)
	s.renderToDisplay(cameraViewerContext, lightContext)
	s.renderImgui()

	s.window.GLSwap()
}

func (s *RenderSystem) renderToDepthMap(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()
	s.shadowMap.Prepare()

	s.renderScene(viewerContext, lightContext, true)
}

func (s *RenderSystem) renderToDisplay(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()

	gl.Viewport(0, 0, int32(s.width), int32(s.height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	s.renderScene(viewerContext, lightContext, false)
}

func (s *RenderSystem) renderImgui() {
	s.platform.NewFrame()
	imgui.NewFrame()

	if settings.ShowImguiDemoWindow {
		open := true
		imgui.ShowDemoWindow(&open)
	}

	s.world.SetFocusedWindow(types.WindowGame)
	if s.world.GetWindowVisibility(types.WindowDebug) {
		s.debugWindow()
	}
	if s.world.GetWindowVisibility(types.WindowConsole) {
		s.consoleWindow()
	}
	if s.world.GetWindowVisibility(types.WindowInventory) {
		s.inventoryWindow()
	}

	imgui.Render()
	s.imguiRenderer.Render(s.platform.DisplaySize(), s.platform.FramebufferSize(), imgui.RenderedDrawData())
}

// renderScene renders a scene from the perspective of a viewer
func (s *RenderSystem) renderScene(viewerContext ViewerContext, lightContext LightContext, shadowPass bool) {
	d := directory.GetDirectory()
	shaderManager := d.ShaderManager()

	// render a debug shadow map for viewing
	// drawHUDTextureToQuad(viewerContext, shaderManager.GetShaderProgram("depthDebug"), s.shadowMap.DepthTexture(), 0.4)
	// drawHUDTextureToQuad(viewerContext, shaderManager.GetShaderProgram("quadtex"), textTexture, 0.4)

	for _, entity := range s.world.QueryEntity(components.ComponentFlagRender) {
		componentContainer := entity.GetComponentContainer()
		entityPosition := componentContainer.TransformComponent.Position
		orientation := componentContainer.TransformComponent.Orientation
		translation := mgl64.Translate3D(entityPosition.X(), entityPosition.Y(), entityPosition.Z())
		renderComponent := componentContainer.RenderComponent

		if renderComponent.IsVisible {
			meshComponent := componentContainer.MeshComponent
			meshModelMatrix := createModelMatrix(
				meshComponent.Scale,
				orientation.Mat4().Mul4(meshComponent.Orientation),
				translation,
			)

			shader := "model_static"
			if componentContainer.AnimationComponent != nil {
				shader = "modelpbr"
			}

			if !shadowPass && componentContainer.HealthComponent != nil {
				center := mgl64.Vec3{componentContainer.TransformComponent.Position.X(), 0, componentContainer.TransformComponent.Position.Z()}
				viewerArtificialCenter := mgl64.Vec3{viewerContext.Position.X(), 0, viewerContext.Position.Z()}
				vecToViewer := viewerArtificialCenter.Sub(center).Normalize()
				// billboardModelMatrix := translation
				billboardModelMatrix := translation.Mul4(mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, 1}, vecToViewer).Mat4())
				// billboardModelMatrix := translation.Mul4(libutils.Vec3ToQuat(vecToViewer).Mat4())
				// billboardModelMatrix := translation.Mul4(mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, mgl64.Vec3{1, 0, 1}).Mat4())
				// fmt.Println(mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, 1}, vecToViewer))
				drawHealthHUD(
					componentContainer.HealthComponent,
					viewerContext,
					lightContext,
					shaderManager.GetShaderProgram("flat"),
					mgl64.Vec3{0.86, 0.1, 0.1},
					billboardModelMatrix,
				)
			}

			drawModel(
				viewerContext,
				lightContext,
				s.shadowMap,
				shaderManager.GetShaderProgram(shader),
				componentContainer.MeshComponent,
				componentContainer.AnimationComponent,
				meshModelMatrix,
				orientation.Mat4().Mul4(meshComponent.Orientation),
			)

			if shadowPass {
				continue
			}

			if settings.DebugRenderSpatialPartition {
				if componentContainer.ColliderComponent.BoundingBoxCollider != nil {
					bb := componentContainer.ColliderComponent.BoundingBoxCollider.Transform(componentContainer.TransformComponent.Position)
					drawAABB(
						viewerContext,
						shaderManager.GetShaderProgram("flat"),
						mgl64.Vec3{.2, 0, .7},
						bb,
						settings.DefaultLineThickness,
					)
				}
			}
		}

		if shadowPass {
			continue
		}

		// rendered objects that should not be picked up by the shadow map
		if settings.DebugRenderCollisionVolume {
			if componentContainer.ColliderComponent != nil {
				if componentContainer.ColliderComponent.CapsuleCollider != nil {
					// lots of hacky rendering stuff to get the rectangle to billboard
					center := mgl64.Vec3{componentContainer.TransformComponent.Position.X(), 0, componentContainer.TransformComponent.Position.Z()}
					viewerArtificialCenter := mgl64.Vec3{viewerContext.Position.X(), 0, viewerContext.Position.Z()}
					vecToViewer := viewerArtificialCenter.Sub(center).Normalize()
					billboardModelMatrix := translation.Mul4(mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, 1}, vecToViewer).Mat4())
					drawCapsuleCollider(
						viewerContext,
						lightContext,
						shaderManager.GetShaderProgram("flat"),
						mgl64.Vec3{0.5, 1, 0},
						componentContainer.ColliderComponent.CapsuleCollider,
						billboardModelMatrix,
					)
				} else if componentContainer.ColliderComponent.TransformedTriMeshCollider != nil {
					drawTriMeshCollider(
						viewerContext,
						lightContext,
						shaderManager.GetShaderProgram("flat"),
						mgl64.Vec3{0.5, 1, 0},
						componentContainer.ColliderComponent.TransformedTriMeshCollider,
					)
				}
			}
		}
	}

	if shadowPass {
		return
	}

	assetManager := d.AssetManager()
	drawSkyBox(
		viewerContext,
		s.skybox,
		shaderManager.GetShaderProgram("skybox"),
		assetManager.GetTexture("front"),
		assetManager.GetTexture("top"),
		assetManager.GetTexture("left"),
		assetManager.GetTexture("right"),
		assetManager.GetTexture("bottom"),
		assetManager.GetTexture("back"),
	)

	if settings.DebugRenderSpatialPartition {
		drawSpatialPartition(
			viewerContext,
			shaderManager.GetShaderProgram("flat"),
			mgl64.Vec3{0.5, 1, 0},
			s.world.SpatialPartition(),
			settings.DefaultLineThickness,
		)
	}

	// var renderText string
	// drawText(shaderManager.GetShaderProgram("quadtex"), assetManager.GetFont("robotomono-regular"), renderText, 0.8, 0)
}

func (s *RenderSystem) Update(delta time.Duration) {
}

func (s *RenderSystem) Name() string {
	return "RenderSystem"
}
