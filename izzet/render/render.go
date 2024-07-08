package render

import (
	"fmt"
	"math"
	"strings"
	"time"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	lib "github.com/kkevinchou/izzet/internal"
	"github.com/kkevinchou/izzet/internal/renderers"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/render/menus"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/render/panels/drawer"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

type GameWorld interface {
	Entities() []*entities.Entity
	Lights() []*entities.Entity
	GetEntityByID(id int) *entities.Entity
	AddEntity(entity *entities.Entity)
	SpatialPartition() *spatialpartition.SpatialPartition
}

const (
	mipsCount                  int     = 6
	MaxBloomTextureWidth       int     = 1920
	MaxBloomTextureHeight      int     = 1080
	internalTextureColorFormat int32   = gl.RGB10_A2
	uiWidthRatio               float32 = 0.2
)

type Renderer struct {
	app           renderiface.App
	world         GameWorld
	shaderManager *shaders.ShaderManager

	shadowMap           *ShadowMap
	imguiRenderer       *renderers.OpenGL3
	depthCubeMapTexture uint32
	depthCubeMapFBO     uint32

	cameraDepthMapFBO  uint32
	cameraDepthTexture uint32

	redCircleFB         uint32
	redCircleTexture    uint32
	greenCircleFB       uint32
	greenCircleTexture  uint32
	blueCircleFB        uint32
	blueCircleTexture   uint32
	yellowCircleFB      uint32
	yellowCircleTexture uint32

	cameraViewerContext ViewerContext

	renderFBO              uint32
	mainColorTexture       uint32
	colorPickingTexture    uint32
	colorPickingAttachment uint32
	imguiMainColorTexture  imgui.TextureID

	downSampleFBO      uint32
	xyTextureVAO       uint32
	downSampleTextures []uint32

	upSampleFBO         uint32
	upSampleTextures    []uint32
	blendTargetTextures []uint32

	compositeFBO          uint32
	compositeTexture      uint32
	imguiCompositeTexture imgui.TextureID

	blendFBO uint32

	bloomTextureWidths  []int
	bloomTextureHeights []int

	cubeVAOs      map[string]uint32
	batchCubeVAOs map[string]uint32
	triangleVAOs  map[string]uint32

	gameWindowHovered bool
	gameWindowWidth   int
	gameWindowHeight  int
	menuBarHeight     float32

	hoveredEntityID *int
}

func New(app renderiface.App, shaderDirectory string, width, height int) *Renderer {
	r := &Renderer{app: app, gameWindowWidth: width, gameWindowHeight: height}
	r.shaderManager = shaders.NewShaderManager(shaderDirectory)
	compileShaders(r.shaderManager)

	imguiIO := imgui.CurrentIO()
	imguiRenderer, err := renderers.NewOpenGL3(imguiIO)
	if err != nil {
		panic(err)
	}
	r.imguiRenderer = imguiRenderer

	// note(kevin) using exactly the max texture size sometimes causes initialization to fail.
	// so, I cap it at a fraction of the max
	var maxTextureSize int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxTextureSize)
	// settings.RuntimeMaxTextureSize = int(float32(maxTextureSize) * .90)
	// shadowMap, err := NewShadowMap(settings.RuntimeMaxTextureSize, settings.RuntimeMaxTextureSize, float64(panels.DBG.Far))
	dimension := 14400
	shadowMap, err := NewShadowMap(dimension, dimension, float64(r.app.RuntimeConfig().Far))
	if err != nil {
		panic(fmt.Sprintf("failed to create shadow map %s", err))
	}
	r.shadowMap = shadowMap
	r.depthCubeMapFBO, r.depthCubeMapTexture = lib.InitDepthCubeMap()
	r.xyTextureVAO = r.init2f2fVAO()
	r.cubeVAOs = map[string]uint32{}
	r.batchCubeVAOs = map[string]uint32{}
	r.triangleVAOs = map[string]uint32{}

	r.ReinitializeFrameBuffers()

	// circles for the rotation gizmo

	r.redCircleFB, r.redCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA, gl.RGBA)
	r.greenCircleFB, r.greenCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA, gl.RGBA)
	r.blueCircleFB, r.blueCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA, gl.RGBA)
	r.yellowCircleFB, r.yellowCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA, gl.RGBA)

	// bloom setup
	widths, heights := createSamplingDimensions(MaxBloomTextureWidth/2, MaxBloomTextureHeight/2, 6)
	r.bloomTextureWidths = widths
	r.bloomTextureHeights = heights
	r.downSampleTextures = initSamplingTextures(widths, heights)
	r.downSampleFBO = initSamplingBuffer(r.downSampleTextures[0])

	widths, heights = createSamplingDimensions(MaxBloomTextureWidth, MaxBloomTextureHeight, 6)
	r.upSampleTextures = initSamplingTextures(widths, heights)
	r.blendTargetTextures = initSamplingTextures(widths, heights)
	r.upSampleFBO = initSamplingBuffer(r.upSampleTextures[0])

	// the texture is only needed to properly generate the FBO
	// new textures are binded when we're in the process of blooming
	r.blendFBO, _ = r.initFBOAndTexture(width, height)

	r.initializeCircleTextures()

	return r
}

func (r *Renderer) ReinitializeFrameBuffers() {
	menuBarSize := CalculateMenuBarHeight()
	footerSize := apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled)

	windowWidth, windowHeight := r.app.WindowSize()

	width := windowWidth
	height := windowHeight - int(menuBarSize) - int(footerSize)

	if r.app.RuntimeConfig().UIEnabled {
		width = int(math.Ceil(float64(1-uiWidthRatio) * float64(windowWidth)))
	}

	r.initMainRenderFBO(width, height)
	r.initCompositeFBO(width, height)
	r.initDepthMapFBO(width, height)
}

func (r *Renderer) initDepthMapFBO(width, height int) {
	var storedFBO int32
	gl.GetIntegerv(gl.FRAMEBUFFER_BINDING, &storedFBO)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, uint32(storedFBO))

	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT,
		int32(width), int32(height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("failed to initialize shadow map frame buffer - in the past this was due to an overly large shadow map dimension configuration")
	}

	r.cameraDepthMapFBO, r.cameraDepthTexture = depthMapFBO, texture
}

func (r *Renderer) initMainRenderFBO(width, height int) {
	renderFBO, colorTextures := r.initFrameBuffer(width, height, []int32{internalTextureColorFormat, gl.R32UI}, []uint32{gl.RGBA, gl.RED_INTEGER})
	r.renderFBO = renderFBO
	r.mainColorTexture = colorTextures[0]
	r.imguiMainColorTexture = imgui.TextureID{Data: uintptr(r.mainColorTexture)}
	r.colorPickingTexture = colorTextures[1]
	r.colorPickingAttachment = gl.COLOR_ATTACHMENT1
}

func (r *Renderer) initCompositeFBO(width, height int) {
	r.compositeFBO, r.compositeTexture = r.initFBOAndTexture(width, height)
	r.imguiCompositeTexture = imgui.TextureID{Data: uintptr(r.compositeTexture)}
}

func (r *Renderer) Render(delta time.Duration) {
	// if r.app.Minimized() || !r.app.WindowFocused() {
	// 	return
	// }

	initOpenGLRenderSettings()
	renderContext := NewRenderContext(r.gameWindowWidth, r.gameWindowHeight, float64(r.app.RuntimeConfig().FovX))
	r.app.RuntimeConfig().TriangleDrawCount = 0
	r.app.RuntimeConfig().DrawCount = 0

	// configure camera viewer context

	var position mgl64.Vec3
	var rotation mgl64.Quat = mgl64.QuatIdent()

	if r.app.AppMode() == mode.AppModeEditor {
		position = r.app.GetEditorCameraPosition()
		rotation = r.app.GetEditorCameraRotation()
	} else {
		camera := r.app.GetPlayerCamera()
		position = camera.Position()
		rotation = camera.WorldRotation()
	}

	viewerViewMatrix := rotation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := ViewerContext{
		Position: position,
		Rotation: rotation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(renderContext.FovY()), renderContext.AspectRatio(), float64(r.app.RuntimeConfig().Near), float64(r.app.RuntimeConfig().Far)),
	}

	lightFrustumPoints := calculateFrustumPoints(
		position,
		rotation,
		float64(r.app.RuntimeConfig().Near),
		float64(r.app.RuntimeConfig().Far),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		0,
		float64(r.app.RuntimeConfig().ShadowFarFactor),
	)

	// find the directional light if there is one
	lights := r.world.Lights()
	var directionalLight *entities.Entity
	for _, light := range lights {
		if light.LightInfo.Type == 0 {
			directionalLight = light
			break
		}
	}

	var directionalLightX, directionalLightY, directionalLightZ float64 = 0, -1, 0
	if directionalLight != nil {
		directionalLightX = float64(directionalLight.LightInfo.Direction3F[0])
		directionalLightY = float64(directionalLight.LightInfo.Direction3F[1])
		directionalLightZ = float64(directionalLight.LightInfo.Direction3F[2])
	}

	lightRotation := utils.Vec3ToQuat(mgl64.Vec3{directionalLightX, directionalLightY, directionalLightZ})
	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightRotation.Mat4(), lightFrustumPoints, r.app.RuntimeConfig().ShadowmapZOffset)
	lightViewMatrix := mgl64.Translate3D(lightPosition.X(), lightPosition.Y(), lightPosition.Z()).Mul4(lightRotation.Mat4()).Inv()

	lightViewerContext := ViewerContext{
		Position:          lightPosition,
		Rotation:          lightRotation,
		InverseViewMatrix: lightViewMatrix,
		ProjectionMatrix:  lightProjectionMatrix,
	}

	lightContext := LightContext{
		// this should be the inverse of the transforms applied to the viewer context
		// if the viewer moves along -y, the universe moves along +y
		LightSpaceMatrix: lightProjectionMatrix.Mul4(lightViewMatrix),
		Lights:           r.world.Lights(),
	}

	r.cameraViewerContext = cameraViewerContext

	r.clearMainFrameBuffer(renderContext)

	renderableEntities := r.fetchRenderableEntities(position, rotation, renderContext)
	shadowEntities := r.fetchShadowCastingEntities(position, rotation, renderContext)

	r.drawSkybox(renderContext)
	_ = lightViewerContext
	r.drawToShadowDepthMap(lightViewerContext, shadowEntities)
	r.drawToCubeDepthMap(lightContext, shadowEntities)
	r.drawToCameraDepthMap(cameraViewerContext, renderableEntities)

	// main color FBO
	r.drawToMainColorBuffer(cameraViewerContext, lightContext, renderContext, renderableEntities)
	r.drawAnnotations(cameraViewerContext, lightContext, renderContext)

	// clear depth for gizmo rendering
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	r.renderGizmos(cameraViewerContext, renderContext)

	// store color picking entity
	r.hoveredEntityID = r.getEntityByPixelPosition(r.app.GetFrameInput().MouseInput.Position)

	var finalRenderTexture uint32
	var imguiFinalRenderTexture imgui.TextureID
	if r.app.RuntimeConfig().Bloom {
		r.downSample(r.mainColorTexture, r.bloomTextureWidths, r.bloomTextureHeights)
		upsampleTexture := r.upSample(r.bloomTextureWidths, r.bloomTextureHeights)
		finalRenderTexture = r.composite(renderContext, r.mainColorTexture, upsampleTexture)
		imguiFinalRenderTexture = r.imguiCompositeTexture
		if panels.SelectedComboOption == panels.ComboOptionFinalRender {
			r.app.RuntimeConfig().DebugTexture = finalRenderTexture
		} else if panels.SelectedComboOption == panels.ComboOptionColorPicking {
			r.app.RuntimeConfig().DebugTexture = r.colorPickingTexture
		} else if panels.SelectedComboOption == panels.ComboOptionHDR {
			r.app.RuntimeConfig().DebugTexture = r.mainColorTexture
		} else if panels.SelectedComboOption == panels.ComboOptionBloom {
			r.app.RuntimeConfig().DebugTexture = upsampleTexture
		} else if panels.SelectedComboOption == panels.ComboOptionShadowDepthMap {
			r.app.RuntimeConfig().DebugTexture = r.shadowMap.depthTexture
		} else if panels.SelectedComboOption == panels.ComboOptionCameraDepthMap {
			r.app.RuntimeConfig().DebugTexture = r.cameraDepthTexture
		}
	} else {
		finalRenderTexture = r.mainColorTexture
		imguiFinalRenderTexture = r.imguiMainColorTexture
		if panels.SelectedComboOption == panels.ComboOptionFinalRender {
			r.app.RuntimeConfig().DebugTexture = finalRenderTexture
		} else if panels.SelectedComboOption == panels.ComboOptionColorPicking {
			r.app.RuntimeConfig().DebugTexture = r.colorPickingTexture
		} else if panels.SelectedComboOption == panels.ComboOptionHDR {
			r.app.RuntimeConfig().DebugTexture = 0
		} else if panels.SelectedComboOption == panels.ComboOptionBloom {
			r.app.RuntimeConfig().DebugTexture = 0
		} else if panels.SelectedComboOption == panels.ComboOptionShadowDepthMap {
			r.app.RuntimeConfig().DebugTexture = r.shadowMap.depthTexture
		} else if panels.SelectedComboOption == panels.ComboOptionCameraDepthMap {
			r.app.RuntimeConfig().DebugTexture = r.cameraDepthTexture
		}
	}

	// render to back buffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	// r.drawTexturedQuad(&cameraViewerContext, r.shaderManager, finalRenderTexture, float32(renderContext.aspectRatio), nil, false, nil)

	r.renderImgui(renderContext, imguiFinalRenderTexture)
}

func (r *Renderer) fetchShadowCastingEntities(cameraPosition mgl64.Vec3, rotation mgl64.Quat, renderContext RenderContext) []*entities.Entity {
	frustumPoints := calculateFrustumPoints(
		cameraPosition,
		rotation,
		float64(r.app.RuntimeConfig().Near),
		float64(r.app.RuntimeConfig().Far),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		float64(r.app.RuntimeConfig().ShadowSpatialPartitionNearPlane),
		1,
	)
	frustumBoundingBox := collider.BoundingBoxFromVertices(frustumPoints)
	return r.fetchEntitiesByBoundingBox(cameraPosition, rotation, renderContext, frustumBoundingBox, entities.ShadowCasting)
}

func (r *Renderer) fetchRenderableEntities(cameraPosition mgl64.Vec3, rotation mgl64.Quat, renderContext RenderContext) []*entities.Entity {
	frustumPoints := calculateFrustumPoints(
		cameraPosition,
		rotation,
		float64(r.app.RuntimeConfig().Near),
		float64(r.app.RuntimeConfig().Far),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		0,
		1,
	)
	frustumBoundingBox := collider.BoundingBoxFromVertices(frustumPoints)
	return r.fetchEntitiesByBoundingBox(cameraPosition, rotation, renderContext, frustumBoundingBox, entities.Renderable)
}

func (r *Renderer) fetchEntitiesByBoundingBox(cameraPosition mgl64.Vec3, rotation mgl64.Quat, renderContext RenderContext, boundingBox collider.BoundingBox, filter entities.FilterFunction) []*entities.Entity {
	var renderEntities []*entities.Entity
	if r.app.RuntimeConfig().EnableSpatialPartition {
		spatialPartition := r.world.SpatialPartition()
		frustumEntities := spatialPartition.QueryEntities(boundingBox)
		for _, entity := range frustumEntities {
			e := r.world.GetEntityByID(entity.GetID())
			if !filter(e) {
				continue
			}
			renderEntities = append(renderEntities, e)
		}
	} else {
		renderEntities = r.world.Entities()
	}

	return renderEntities
}

var spanLines [][2]mgl64.Vec3

func (r *Renderer) drawAnnotations(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext) {
	shaderManager := r.shaderManager

	if r.app.RuntimeConfig().ShowSelectionBoundingBox {
		entity := r.app.SelectedEntity()
		if entity != nil {
			// draw bounding box
			if entity.HasBoundingBox() {
				r.drawAABB(
					viewerContext,
					mgl64.Vec3{.2, 0, .7},
					entity.BoundingBox(),
					0.5,
				)
			}
		}
	}

	if r.app.AppMode() == mode.AppModeEditor {
		for _, entity := range r.world.Entities() {
			lightInfo := entity.LightInfo
			if lightInfo != nil {
				if lightInfo.Type == 0 {

					direction3F := lightInfo.Direction3F
					dir := mgl64.Vec3{float64(direction3F[0]), float64(direction3F[1]), float64(direction3F[2])}.Mul(50)
					// directional light arrow
					lines := [][2]mgl64.Vec3{
						[2]mgl64.Vec3{
							entity.Position(),
							entity.Position().Add(dir),
						},
					}

					shader := shaderManager.GetShaderProgram("flat")
					color := mgl64.Vec3{252.0 / 255, 241.0 / 255, 33.0 / 255}
					shader.Use()
					shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
					shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
					shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

					r.drawLineGroup(fmt.Sprintf("%d_%v_%v", entity.ID, entity.Position(), dir), viewerContext, shader, lines, 0.5, color)
				}
			}
		}
	}

	if r.app.RuntimeConfig().EnableSpatialPartition && r.app.RuntimeConfig().RenderSpatialPartition {
		r.drawSpatialPartition(viewerContext, mgl64.Vec3{0, 1, 0}, r.world.SpatialPartition(), 0.5)
	}

	nm := r.app.NavMesh()
	if nm != nil {
		shader := shaderManager.GetShaderProgram("navmesh")
		shader.Use()
		shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

		setupLightingUniforms(shader, lightContext.Lights)
		shader.SetUniformInt("width", int32(r.gameWindowWidth))
		shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
		shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
		shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
		shader.SetUniformFloat("ambientFactor", r.app.RuntimeConfig().AmbientFactor)
		shader.SetUniformInt("shadowMap", 31)
		shader.SetUniformInt("depthCubeMap", 30)
		shader.SetUniformInt("cameraDepthMap", 29)
		if !r.app.RuntimeConfig().Bloom {
			// only tone map if we're not applying bloom, otherwise
			// we want to keep the HDR values and tone map later
			shader.SetUniformInt("applyToneMapping", 1)
		} else {
			shader.SetUniformInt("applyToneMapping", 0)
		}
		shader.SetUniformFloat("near", r.app.RuntimeConfig().Near)
		shader.SetUniformFloat("far", r.app.RuntimeConfig().Far)
		shader.SetUniformFloat("bias", r.app.RuntimeConfig().PointLightBias)
		shader.SetUniformFloat("far_plane", float32(settings.DepthCubeMapFar))
		shader.SetUniformVec3("albedo", mgl32.Vec3{1, 0, 0})

		shader.SetUniformFloat("roughness", .8)
		shader.SetUniformFloat("metallic", 0)

		r.drawNavmesh(nm)

		// draw bounding box
		volume := nm.Volume
		r.drawAABB(
			viewerContext,
			mgl64.Vec3{155.0 / 99, 180.0 / 255, 45.0 / 255},
			volume,
			0.5,
		)

		if len(nm.DebugLines) > 0 {
			shader := shaderManager.GetShaderProgram("flat")
			// color := mgl64.Vec3{252.0 / 255, 241.0 / 255, 33.0 / 255}
			color := mgl64.Vec3{1, 0, 0}
			shader.Use()
			shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
			shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
			shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
			r.drawLineGroup(fmt.Sprintf("navmesh_debuglines_%d", nm.InvalidatedTimestamp), viewerContext, shader, nm.DebugLines, 0.5, color)
		}

		nm.Invalidated = false
	}
}

func (r *Renderer) drawToCameraDepthMap(viewerContext ViewerContext, renderableEntities []*entities.Entity) {
	gl.Viewport(0, 0, int32(r.gameWindowWidth), int32(r.gameWindowHeight))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.cameraDepthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	r.renderGeometryWithoutColor(viewerContext, renderableEntities, entities.Renderable)
}

func (r *Renderer) drawToShadowDepthMap(viewerContext ViewerContext, renderableEntities []*entities.Entity) {
	r.shadowMap.Prepare()
	defer gl.CullFace(gl.BACK)

	if !r.app.RuntimeConfig().EnableShadowMapping {
		// set the depth to be max value to prevent shadow mapping
		gl.ClearDepth(1)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		return
	}

	r.renderGeometryWithoutColor(viewerContext, renderableEntities, entities.EmptyFilter)
}

func (r *Renderer) renderGeometryWithoutColor(viewerContext ViewerContext, renderableEntities []*entities.Entity, filter entities.FilterFunction) {
	shader := r.shaderManager.GetShaderProgram("modelgeo")
	shader.Use()

	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	for _, entity := range renderableEntities {
		if !filter(entity) {
			continue
		}

		if entity.Animation != nil && entity.Animation.AnimationPlayer.CurrentAnimation() != "" {
			shader.SetUniformInt("isAnimated", 1)
			animationTransforms := entity.Animation.AnimationPlayer.AnimationTransforms()
			// if animationTransforms is nil, the shader will execute reading into invalid memory
			// so, we need to explicitly guard for this
			if animationTransforms == nil {
				panic("animationTransforms not found")
			}
			for i := 0; i < len(animationTransforms); i++ {
				shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
			}
		} else {
			shader.SetUniformInt("isAnimated", 0)
		}

		modelMatrix := entities.WorldTransform(entity)
		m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix)

		primitives := r.app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
		for _, p := range primitives {
			shader.SetUniformMat4("model", m32ModelMatrix.Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform)))

			gl.BindVertexArray(p.GeometryVAO)
			r.iztDrawElements(int32(len(p.Primitive.VertexIndices)))
		}
	}
}

func (r *Renderer) drawToCubeDepthMap(lightContext LightContext, renderableEntities []*entities.Entity) {
	// we only support cube depth maps for one point light atm
	var pointLight *entities.Entity
	for _, light := range r.world.Lights() {
		if light.LightInfo.Type == 1 {
			pointLight = light
			break
		}
	}
	if pointLight == nil {
		return
	}

	gl.Viewport(0, 0, int32(settings.DepthCubeMapWidth), int32(settings.DepthCubeMapHeight))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.depthCubeMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	position := pointLight.Position()
	shadowTransforms := computeCubeMapTransforms(position, settings.DepthCubeMapNear, settings.DepthCubeMapFar)

	shader := r.shaderManager.GetShaderProgram("point_shadow")
	shader.Use()
	for i, transform := range shadowTransforms {
		shader.SetUniformMat4(fmt.Sprintf("shadowMatrices[%d]", i), utils.Mat4F64ToF32(transform))
	}
	shader.SetUniformFloat("far_plane", float32(settings.DepthCubeMapFar))
	shader.SetUniformVec3("lightPos", utils.Vec3F64ToF32(position))

	for _, entity := range renderableEntities {
		if entity == nil || entity.MeshComponent == nil {
			continue
		}

		if entity.Animation != nil && entity.Animation.AnimationPlayer.CurrentAnimation() != "" {
			shader.SetUniformInt("isAnimated", 1)
			animationTransforms := entity.Animation.AnimationPlayer.AnimationTransforms()
			// if animationTransforms is nil, the shader will execute reading into invalid memory
			// so, we need to explicitly guard for this
			if animationTransforms == nil {
				panic("animationTransforms not found")
			}
			for i := 0; i < len(animationTransforms); i++ {
				shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
			}
		} else {
			shader.SetUniformInt("isAnimated", 0)
		}

		modelMatrix := entities.WorldTransform(entity)
		m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix)

		primitives := r.app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
		for _, p := range primitives {
			shader.SetUniformMat4("model", m32ModelMatrix.Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform)))

			gl.BindVertexArray(p.GeometryVAO)
			r.iztDrawElements(int32(len(p.Primitive.VertexIndices)))
		}
	}
}

// drawToMainColorBuffer renders a scene from the perspective of a viewer
func (r *Renderer) drawToMainColorBuffer(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext, renderableEntities []*entities.Entity) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	r.renderModels(viewerContext, lightContext, renderContext, renderableEntities)

	shaderManager := r.shaderManager

	// render non-models
	for _, entity := range r.world.Entities() {
		if entity.MeshComponent == nil {
			modelMatrix := entities.WorldTransform(entity)

			if len(entity.ShapeData) > 0 {
				shader := shaderManager.GetShaderProgram("flat")
				shader.Use()

				shader.SetUniformUInt("entityID", uint32(entity.ID))
				shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
			}

			if entity.ImageInfo != nil {
				textureName := strings.Split(entity.ImageInfo.ImageName, ".")[0]
				texture := r.app.AssetManager().GetTexture(textureName)
				if texture != nil {
					if entity.Billboard && r.app.AppMode() == mode.AppModeEditor {
						shader := shaderManager.GetShaderProgram("world_space_quad")
						shader.Use()

						position := entity.Position()
						modelMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())
						scale := entity.ImageInfo.Scale
						modelMatrix = modelMatrix.Mul4(mgl64.Scale3D(scale, scale, scale))

						shader.SetUniformUInt("entityID", uint32(entity.ID))
						shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix.Mul4(r.app.GetEditorCameraRotation().Mat4())))
						shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
						shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

						r.drawBillboardTexture(texture.ID, 1)
					}
				} else {
					fmt.Println("couldn't find texture", "light")
				}
			}
			particles := entity.Particles
			if particles != nil {
				texture := r.app.AssetManager().GetTexture("light").ID
				for _, particle := range particles.GetActiveParticles() {
					particleModelMatrix := mgl32.Translate3D(float32(particle.Position.X()), float32(particle.Position.Y()), float32(particle.Position.Z()))
					r.drawTexturedQuad(&viewerContext, r.shaderManager, texture, float32(renderContext.AspectRatio()), &particleModelMatrix, true, nil)
				}
			}
		} else if entity.CharacterControllerComponent != nil {
			v := mgl64.Vec3{}
			if entity.CharacterControllerComponent.WebVector != v {
				// r.drawAABB(
				// 	viewerContext,
				// 	shaderManager.GetShaderProgram("flat"),
				// 	mgl64.Vec3{.2, 0, .7},
				// 	entity.BoundingBox(),
				// 	0.5,
				// )

				forwardVector := viewerContext.Rotation.Rotate(mgl64.Vec3{0, 0, -1})
				upVector := viewerContext.Rotation.Rotate(mgl64.Vec3{0, 1, 0})
				// there's probably away to get the right vector directly rather than going crossing the up vector :D
				rightVector := forwardVector.Cross(upVector)

				start := entity.Position().Add(rightVector.Mul(10)).Add(mgl64.Vec3{0, 30, 0})
				lines := [][]mgl64.Vec3{
					{start, entity.Position().Add(entity.CharacterControllerComponent.WebVector)},
				}

				shader := shaderManager.GetShaderProgram("flat")
				shader.Use()
				shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

				r.drawLines(viewerContext, shader, lines, 1, mgl64.Vec3{1, 1, 1})
			}
		}
	}
}

func (r *Renderer) renderModels(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext, renderableEntities []*entities.Entity) {
	shaderManager := r.shaderManager

	shader := shaderManager.GetShaderProgram("modelpbr")
	shader.Use()

	if !r.app.RuntimeConfig().Bloom {
		// only tone map if we're not applying bloom, otherwise
		// we want to keep the HDR values and tone map later
		shader.SetUniformInt("applyToneMapping", 1)
	} else {
		shader.SetUniformInt("applyToneMapping", 0)
	}

	if r.app.RuntimeConfig().FogEnabled {
		shader.SetUniformInt("fog", 1)
	} else {
		shader.SetUniformInt("fog", 0)
	}

	var fog int32 = 0
	if r.app.RuntimeConfig().FogDensity != 0 {
		fog = 1
	}
	shader.SetUniformInt("fog", fog)
	shader.SetUniformInt("fogDensity", r.app.RuntimeConfig().FogDensity)

	// TODO - this should probably be game window size?
	shader.SetUniformInt("width", int32(r.gameWindowWidth))
	shader.SetUniformInt("height", int32(r.gameWindowHeight))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformFloat("ambientFactor", r.app.RuntimeConfig().AmbientFactor)
	shader.SetUniformInt("shadowMap", 31)
	shader.SetUniformInt("depthCubeMap", 30)
	shader.SetUniformInt("cameraDepthMap", 29)

	shader.SetUniformFloat("near", r.app.RuntimeConfig().Near)
	shader.SetUniformFloat("far", r.app.RuntimeConfig().Far)
	shader.SetUniformFloat("bias", r.app.RuntimeConfig().PointLightBias)
	shader.SetUniformFloat("far_plane", float32(settings.DepthCubeMapFar))
	shader.SetUniformInt("hasColorOverride", 0)

	setupLightingUniforms(shader, lightContext.Lights)

	gl.ActiveTexture(gl.TEXTURE29)
	gl.BindTexture(gl.TEXTURE_2D, r.cameraDepthTexture)

	gl.ActiveTexture(gl.TEXTURE30)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, r.depthCubeMapTexture)

	gl.ActiveTexture(gl.TEXTURE31)
	gl.BindTexture(gl.TEXTURE_2D, r.shadowMap.DepthTexture())

	for _, entity := range renderableEntities {
		if entity == nil || entity.MeshComponent == nil || !entity.MeshComponent.Visible {
			continue
		}

		if entity.MeshComponent.InvisibleToPlayerOwner && r.app.GetPlayerEntity().GetID() == entity.GetID() {
			continue
		}

		shader.SetUniformUInt("entityID", uint32(entity.ID))

		r.drawModel(
			viewerContext,
			lightContext,
			shader,
			entity,
		)
	}

	if r.app.RuntimeConfig().ShowColliders {
		shader := shaderManager.GetShaderProgram("flat")
		shader.Use()

		for _, entity := range renderableEntities {
			if entity == nil || entity.MeshComponent == nil || entity.Collider == nil {
				continue
			}

			if entity.MeshComponent.InvisibleToPlayerOwner && r.app.GetPlayerEntity().GetID() == entity.GetID() {
				continue
			}

			modelMatrix := entities.WorldTransform(entity)

			if entity.Collider.SimplifiedTriMeshCollider != nil {
				var lines [][2]mgl64.Vec3
				for _, triangles := range entity.Collider.SimplifiedTriMeshCollider.Triangles {
					lines = append(lines, [2]mgl64.Vec3{
						triangles.Points[0],
						triangles.Points[1],
					})
					lines = append(lines, [2]mgl64.Vec3{
						triangles.Points[1],
						triangles.Points[2],
					})
					lines = append(lines, [2]mgl64.Vec3{
						triangles.Points[2],
						triangles.Points[0],
					})
				}

				if len(lines) > 0 {
					// fmt.Println(nonNilCount, nilCount)
					shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
					shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
					shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
					// r.drawLines(viewerContext, shader, lines, 0.05, mgl64.Vec3{1, 0, 1})
					// r.drawLines(viewerContext, shader, lines, 0.1, mgl64.Vec3{1, 0, 0})
					r.drawLineGroup(fmt.Sprintf("pogchamp_%d", len(lines)), viewerContext, shader, lines, 0.1, mgl64.Vec3{1, 0, 0})
				}

				var pointLines [][2]mgl64.Vec3
				for _, p := range entity.Collider.SimplifiedTriMeshCollider.DebugPoints {
					// 0 length lines
					pointLines = append(pointLines, [2]mgl64.Vec3{p, p.Add(mgl64.Vec3{0.1, 0.1, 0.1})})
				}
				if len(pointLines) > 0 {
					r.drawLineGroup(fmt.Sprintf("pogchamp_points_%d", len(pointLines)), viewerContext, shader, pointLines, 0.1, mgl64.Vec3{0, 0, 1})
				}
			}

			if entity.Collider.CapsuleCollider != nil {
				transform := entities.WorldTransform(entity)
				capsuleCollider := entity.Collider.CapsuleCollider.Transform(transform)

				top := capsuleCollider.Top
				bottom := capsuleCollider.Bottom
				radius := capsuleCollider.Radius

				var numCircleSegments int = 8
				var lines [][]mgl64.Vec3

				// -x +x vertical lines
				lines = append(lines, []mgl64.Vec3{top.Add(mgl64.Vec3{-radius, 0, 0}), bottom.Add(mgl64.Vec3{-radius, 0, 0})})
				lines = append(lines, []mgl64.Vec3{bottom.Add(mgl64.Vec3{radius, 0, 0}), top.Add(mgl64.Vec3{radius, 0, 0})})

				// -z +z vertical lines
				lines = append(lines, []mgl64.Vec3{top.Add(mgl64.Vec3{0, 0, -radius}), bottom.Add(mgl64.Vec3{0, 0, -radius})})
				lines = append(lines, []mgl64.Vec3{bottom.Add(mgl64.Vec3{0, 0, radius}), top.Add(mgl64.Vec3{0, 0, radius})})

				radiansPerSegment := 2 * math.Pi / float64(numCircleSegments)

				// top and bottom xz plane rings
				for i := 0; i < numCircleSegments; i++ {
					x1 := math.Cos(float64(i)*radiansPerSegment) * radius
					z1 := math.Sin(float64(i)*radiansPerSegment) * radius

					x2 := math.Cos(float64((i+1)%numCircleSegments)*radiansPerSegment) * radius
					z2 := math.Sin(float64((i+1)%numCircleSegments)*radiansPerSegment) * radius

					lines = append(lines, []mgl64.Vec3{top.Add(mgl64.Vec3{x1, 0, -z1}), top.Add(mgl64.Vec3{x2, 0, -z2})})
					lines = append(lines, []mgl64.Vec3{bottom.Add(mgl64.Vec3{x1, 0, -z1}), bottom.Add(mgl64.Vec3{x2, 0, -z2})})
				}

				radiansPerSegment = math.Pi / float64(numCircleSegments)

				// top and bottom xy plane rings
				for i := 0; i < numCircleSegments; i++ {
					x1 := math.Cos(float64(i)*radiansPerSegment) * radius
					y1 := math.Sin(float64(i)*radiansPerSegment) * radius

					x2 := math.Cos(float64(float64(i+1)*radiansPerSegment)) * radius
					y2 := math.Sin(float64(float64(i+1)*radiansPerSegment)) * radius

					lines = append(lines, []mgl64.Vec3{top.Add(mgl64.Vec3{x1, y1, 0}), top.Add(mgl64.Vec3{x2, y2, 0})})
					lines = append(lines, []mgl64.Vec3{bottom.Add(mgl64.Vec3{x1, -y1, 0}), bottom.Add(mgl64.Vec3{x2, -y2, 0})})
				}

				// top and bottom yz plane rings
				for i := 0; i < numCircleSegments; i++ {
					z1 := math.Cos(float64(i)*radiansPerSegment) * radius
					y1 := math.Sin(float64(i)*radiansPerSegment) * radius

					z2 := math.Cos(float64(float64(i+1)*radiansPerSegment)) * radius
					y2 := math.Sin(float64(float64(i+1)*radiansPerSegment)) * radius

					lines = append(lines, []mgl64.Vec3{top.Add(mgl64.Vec3{0, y1, z1}), top.Add(mgl64.Vec3{0, y2, z2})})
					lines = append(lines, []mgl64.Vec3{bottom.Add(mgl64.Vec3{0, -y1, z1}), bottom.Add(mgl64.Vec3{0, -y2, z2})})
				}

				shader := shaderManager.GetShaderProgram("flat")
				color := mgl64.Vec3{255.0 / 255, 147.0 / 255, 12.0 / 255}
				shader.Use()
				shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

				r.drawLines(viewerContext, shader, lines, 0.5, color)
			}
		}
	}
}

func (r *Renderer) renderImgui(renderContext RenderContext, gameWindowTexture imgui.TextureID) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.app.Platform().NewFrame()
	imgui.NewFrame()
	windowWidth, windowHeight := r.app.WindowSize()

	r.gameWindowHovered = false
	menus.SetupMenuBar(r.app)
	menuBarHeight := CalculateMenuBarHeight()
	footerHeight := apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled)
	width := float32(windowWidth) + 2 // weirdly the width is always some pixels off (padding/margins maybe?)
	height := float32(windowHeight) - menuBarHeight - footerHeight

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: width, Y: height}, imgui.CondNone)
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: menuBarHeight})
	if imgui.BeginV("Final Render", nil, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoBringToFrontOnFocus) {
		regionSize := imgui.ContentRegionAvail()
		imageWidth := regionSize.X

		var gameWindowRatio float32 = 1
		if r.app.RuntimeConfig().UIEnabled {
			gameWindowRatio = 1 - uiWidthRatio
		}

		// size := imgui.Vec2{X: imageWidth * gameWindowRatio, Y: imageWidth / float32(renderContext.AspectRatio())}
		size := imgui.Vec2{X: imageWidth * gameWindowRatio, Y: regionSize.Y}
		r.gameWindowWidth = int(size.X)
		r.gameWindowHeight = int(size.Y)

		if imgui.BeginChildStrV("Game Window", size, imgui.ChildFlagsNone, imgui.WindowFlagsNoBringToFrontOnFocus) {
			imgui.ImageV(gameWindowTexture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
		}
		if imgui.BeginDragDropTarget() {
			if payload := imgui.AcceptDragDropPayload("content_browser_item"); payload != nil && payload.CData != nil {
				entityName := *(*string)(payload.CData.Data)
				entity := r.app.InstantiateEntity(entityName)
				r.app.SelectEntity(entity)
			}
			imgui.EndDragDropTarget()
			// if payload := imgui.AcceptDragDropPayloadV("content_browser_item", imgui.DragDropFlagsSourceAllowNullID); payload != nil {
			// 	fmt.Println(payload)
			// 	// data := payload.Data()
			// 	// ptr := (*string)(data)

			// 	// itemName := *ptr
			// 	// document := r.app.AssetManager().GetDocument(itemName)
			// 	// handle := assets.NewGlobalHandle(itemName)
			// 	// if len(document.Scenes) != 1 {
			// 	// 	panic("single entity asset loading only supports a singular scene")
			// 	// }

			// 	// scene := document.Scenes[0]
			// 	// node := scene.Nodes[0]

			// 	// entity := entities.InstantiateEntity(document.Name)
			// 	// entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Ident4(), Visible: true, ShadowCasting: true}
			// 	// var vertices []modelspec.Vertex
			// 	// entities.VerticesFromNode(node, document, &vertices)
			// 	// entity.InternalBoundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
			// 	// entities.SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
			// 	// entities.SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
			// 	// entities.SetScale(entity, utils.Vec3F32ToF64(node.Scale))

			// 	// r.world.AddEntity(entity)
			// 	// imgui.CloseCurrentPopup()
			// }
			// imgui.EndDragDropTarget()
		}

		if imgui.IsWindowHovered() {
			r.gameWindowHovered = true
		}

		imgui.EndChild()

		imgui.SameLine()

		if r.app.RuntimeConfig().UIEnabled {
			imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{5, 5})
			imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0)
			imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
			// imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{})
			// imgui.PushStyleVarVec2(imgui.StyleVarItemInnerSpacing, imgui.Vec2{})
			imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, 0)
			imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 0)
			imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
			imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)
			// imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{})
			imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
			imgui.PushStyleColorVec4(imgui.ColHeader, HeaderColor)
			imgui.PushStyleColorVec4(imgui.ColHeaderActive, HeaderColor)
			imgui.PushStyleColorVec4(imgui.ColHeaderHovered, HoveredHeaderColor)
			imgui.PushStyleColorVec4(imgui.ColTitleBg, TitleColor)
			imgui.PushStyleColorVec4(imgui.ColTitleBgActive, TitleColor)
			imgui.PushStyleColorVec4(imgui.ColSliderGrab, InActiveColorControl)
			imgui.PushStyleColorVec4(imgui.ColSliderGrabActive, ActiveColorControl)
			imgui.PushStyleColorVec4(imgui.ColFrameBg, InActiveColorBg)
			imgui.PushStyleColorVec4(imgui.ColFrameBgActive, ActiveColorBg)
			imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, HoverColorBg)
			imgui.PushStyleColorVec4(imgui.ColCheckMark, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
			imgui.PushStyleColorVec4(imgui.ColButton, InActiveColorControl)
			imgui.PushStyleColorVec4(imgui.ColButtonActive, ActiveColorControl)
			imgui.PushStyleColorVec4(imgui.ColButtonHovered, HoverColorControl)
			imgui.PushStyleColorVec4(imgui.ColTabActive, ActiveColorBg)
			imgui.PushStyleColorVec4(imgui.ColTabUnfocused, InActiveColorBg)
			imgui.PushStyleColorVec4(imgui.ColTabUnfocusedActive, InActiveColorBg)
			imgui.PushStyleColorVec4(imgui.ColTab, InActiveColorBg)
			imgui.PushStyleColorVec4(imgui.ColTabHovered, HoveredHeaderColor)

			panels.BuildTabsSet(
				r.app,
				renderContext,
				r.app.Prefabs(),
			)

			drawer.BuildFooter(
				r.app,
				renderContext,
				r.app.Prefabs(),
			)

			imgui.PopStyleColorV(20)
			imgui.PopStyleVarV(7)

			if r.app.ShowImguiDemo() {
				imgui.ShowDemoWindow()
			}
		}
	}
	imgui.End()
	imgui.PopStyleVarV(1)

	imgui.Render()
	r.imguiRenderer.Render(r.app.Platform().DisplaySize(), r.app.Platform().FramebufferSize(), imgui.CurrentDrawData())
}

func (r *Renderer) renderGizmos(viewerContext ViewerContext, renderContext RenderContext) {
	if r.app.SelectedEntity() == nil {
		return
	}

	entity := r.world.GetEntityByID(r.app.SelectedEntity().ID)
	position := entity.Position()

	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		r.drawTranslationGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat"), position)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
		r.drawCircleGizmo(&viewerContext, position, renderContext)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
		r.drawScaleGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat"), position)
	}
}

func triangleVAOKey(triangle entities.Triangle) string {
	return fmt.Sprintf("%v_%v_%v", triangle.V1, triangle.V2, triangle.V3)
}

func (r *Renderer) SetWorld(world *world.GameWorld) {
	r.world = world
}

func (r *Renderer) GameWindowHovered() bool {
	return r.gameWindowHovered
}

func (r *Renderer) HoveredEntityID() *int {
	return r.hoveredEntityID
}

func initOpenGLRenderSettings() {
	// gl.ClearColor(0.0, 0.5, 0.5, 0.0)
	gl.ClearColor(1, 1, 1, 1)
	gl.ClearDepth(1)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.FRAMEBUFFER_SRGB)
}

var (
	InActiveColorBg      imgui.Vec4 = imgui.Vec4{X: .1, Y: .1, Z: 0.1, W: 1}
	ActiveColorBg        imgui.Vec4 = imgui.Vec4{X: .3, Y: .3, Z: 0.3, W: 1}
	HoverColorBg         imgui.Vec4 = imgui.Vec4{X: .25, Y: .25, Z: 0.25, W: 1}
	InActiveColorControl imgui.Vec4 = imgui.Vec4{X: .4, Y: .4, Z: 0.4, W: 1}
	HoverColorControl    imgui.Vec4 = imgui.Vec4{X: .45, Y: .45, Z: 0.45, W: 1}
	ActiveColorControl   imgui.Vec4 = imgui.Vec4{X: .5, Y: .5, Z: 0.5, W: 1}
	HeaderColor          imgui.Vec4 = imgui.Vec4{X: 0.3, Y: 0.3, Z: 0.3, W: 1}
	HoveredHeaderColor   imgui.Vec4 = imgui.Vec4{X: 0.4, Y: 0.4, Z: 0.4, W: 1}
	TitleColor           imgui.Vec4 = imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1}
)
