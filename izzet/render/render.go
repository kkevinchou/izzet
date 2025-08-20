package render

import (
	"os"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/renderers"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/menus"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/render/panels/drawer"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/renderpass"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/izzet/izzet/render/windows"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/shaders"
)

type GameWorld interface {
	Entities() []*entities.Entity
	Lights() []*entities.Entity
	GetEntityByID(id int) *entities.Entity
	AddEntity(entity *entities.Entity)
	SpatialPartition() *spatialpartition.SpatialPartition
}

const (
	mipsCount             int = 6
	MaxBloomTextureWidth  int = 1920
	MaxBloomTextureHeight int = 1080

	materialTextureWidth  int32 = 512
	materialTextureHeight int32 = 512

	uiWidthRatio float32 = 0.2
)

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

	ResizeHoverColor  imgui.Vec4 = imgui.Vec4{X: 0.4, Y: 0.4, Z: 0.4, W: 1}
	ResizeActiveColor imgui.Vec4 = imgui.Vec4{X: 0.6, Y: 0.6, Z: 0.6, W: 1}
)

type RenderSystem struct {
	app           renderiface.App
	shaderManager *shaders.ShaderManager

	imguiRenderer *renderers.OpenGL3

	cameraViewerContext context.ViewerContext

	downSampleFBO      uint32
	ndcQuadVAO         uint32
	downSampleTextures []uint32

	upSampleFBO         uint32
	upSampleTextures    []uint32
	blendTargetTextures []uint32

	compositeFBO     uint32
	compositeTexture uint32

	postProcessingFBO     uint32
	postProcessingTexture uint32

	blendFBO uint32

	bloomTextureWidths  []int
	bloomTextureHeights []int

	gameWindowHovered bool

	hoveredEntityID *int

	materialTextureMap map[types.MaterialHandle]uint32

	// list of materials whose textures need to be generated
	materialTextureQueue []types.MaterialHandle

	batchRenders []assets.Batch

	renderPasses      []renderpass.RenderPass
	renderPassContext *context.RenderPassContext

	sceneSize       [2]int
	resizeNextFrame bool
	lastResize      time.Time
}

func New(app renderiface.App, shaderDirectory string, width, height int) *RenderSystem {
	r := &RenderSystem{app: app}
	r.shaderManager = shaders.NewShaderManager(shaderDirectory)
	compileShaders(r.shaderManager)
	rutils.SetRuntimeConfig(app.RuntimeConfig())

	io := imgui.CurrentIO()
	io.SetConfigFlags(io.ConfigFlags() | imgui.ConfigFlagsDockingEnable)
	io.SetConfigDebugIsDebuggerPresent(true)

	imguiRenderer, err := renderers.NewOpenGL3(io)
	if err != nil {
		panic(err)
	}
	r.imguiRenderer = imguiRenderer
	r.ndcQuadVAO = r.init2f2fVAO()
	r.materialTextureMap = map[types.MaterialHandle]uint32{}

	r.lastResize = time.Now()
	r.initorReinitTextures(width, height, true)

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

	blendTextureFn := textureFn(width, height, []int32{rendersettings.InternalTextureColorFormatRGB}, []uint32{rendersettings.RenderFormatRGB}, []uint32{gl.FLOAT})
	r.blendFBO, _ = r.initFrameBufferNoDepth(blendTextureFn)

	cloudTexture0 := &r.app.RuntimeConfig().CloudTextures[0]
	cloudTexture0.VAO, cloudTexture0.WorleyTexture, cloudTexture0.FBO, cloudTexture0.RenderTexture = r.setupVolumetrics(r.shaderManager)

	cloudTexture1 := &r.app.RuntimeConfig().CloudTextures[1]
	cloudTexture1.VAO, cloudTexture1.WorleyTexture, cloudTexture1.FBO, cloudTexture1.RenderTexture = r.setupVolumetrics(r.shaderManager)

	r.renderPassContext = &context.RenderPassContext{}
	r.renderPasses = append(r.renderPasses, renderpass.NewCameraDepthPass(app, r.shaderManager))
	r.renderPasses = append(r.renderPasses, renderpass.NewShadowMapPass(14400, app, r.shaderManager))
	r.renderPasses = append(r.renderPasses, renderpass.NewPointLightPass(app, r.shaderManager))
	r.renderPasses = append(r.renderPasses, renderpass.NewGPass(app, r.shaderManager))
	r.renderPasses = append(r.renderPasses, renderpass.NewSSAOPass(app, r.shaderManager))
	r.renderPasses = append(r.renderPasses, renderpass.NewSSAOBlurPass(app, r.shaderManager))
	r.renderPasses = append(r.renderPasses, renderpass.NewMainPass(app, r.shaderManager))

	for _, pass := range r.renderPasses {
		pass.Init(width, height, r.renderPassContext)
	}

	return r
}

func (r *RenderSystem) CreateMaterialTexture(handle types.MaterialHandle) {
	material := r.app.AssetManager().GetMaterial(handle)
	materialFBO, materialTexture := r.createCircleTexture(int(materialTextureWidth), int(materialTextureHeight))
	r.materialTextureMap[material.Handle] = materialTexture

	gl.BindFramebuffer(gl.FRAMEBUFFER, materialFBO)
	gl.Viewport(0, 0, materialTextureWidth, materialTextureHeight)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	shader := r.shaderManager.GetShaderProgram("material_preview")
	shader.Use()

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	pbr := material.Material.PBRMaterial.PBRMetallicRoughness

	if pbr.BaseColorTextureName != "" {
		shader.SetUniformInt("uUseAlbedoMap", 1)
		shader.SetUniformInt("uAlbedoMap", int32(pbr.BaseColorTextureCoordsIndex))

		gl.ActiveTexture(gl.TEXTURE0)
		texture := r.app.AssetManager().GetTexture(pbr.BaseColorTextureName)
		gl.BindTexture(gl.TEXTURE_2D, texture.ID)
	} else {
		shader.SetUniformInt("uUseAlbedoMap", 0)
	}
	shader.SetUniformInt("uUseMetallicMap", 0)
	shader.SetUniformInt("uUseRoughnessMap", 0)
	shader.SetUniformInt("uUseAOMap", 0)

	shader.SetUniformVec3("uAlbedo", pbr.BaseColorFactor.Vec3())
	shader.SetUniformFloat("uMetallic", pbr.MetalicFactor)
	shader.SetUniformFloat("uRoughness", pbr.RoughnessFactor)
	shader.SetUniformFloat("uAO", 1)

	shader.SetUniformVec3("uLightDir", mgl32.Vec3{0.5, 0.5, 0.5})
	shader.SetUniformVec3("uLightColor", mgl32.Vec3{5, 5, 5})

	r.iztDrawArrays(0, 6)
}

// this might be the most garbage code i've ever written
func (r *RenderSystem) initorReinitTextures(width, height int, init bool) {
	// composite FBO
	compositeTextureFn := textureFn(width, height, []int32{rendersettings.InternalTextureColorFormatRGB}, []uint32{rendersettings.RenderFormatRGB}, []uint32{gl.FLOAT})
	var compositeTextures []uint32
	if init {
		r.compositeFBO, compositeTextures = r.initFrameBufferNoDepth(compositeTextureFn)
	} else {
		gl.BindFramebuffer(gl.FRAMEBUFFER, r.compositeFBO)
		_, _, compositeTextures = compositeTextureFn()
	}
	gl.DeleteTextures(1, &r.compositeTexture)
	r.compositeTexture = compositeTextures[0]

	// post processing FBO
	postProcessingTextureFn := textureFn(width, height, []int32{rendersettings.InternalTextureColorFormatRGB}, []uint32{rendersettings.RenderFormatRGB}, []uint32{gl.FLOAT})
	var postProcessingTextures []uint32
	if init {
		r.postProcessingFBO, postProcessingTextures = r.initFrameBufferNoDepth(postProcessingTextureFn)
	} else {
		gl.BindFramebuffer(gl.FRAMEBUFFER, r.postProcessingFBO)
		_, _, postProcessingTextures = postProcessingTextureFn()
	}
	gl.DeleteTextures(1, &r.postProcessingTexture)
	r.postProcessingTexture = postProcessingTextures[0]
}

func (r *RenderSystem) ReinitializeFrameBuffers() {
	width, height := r.GameWindowSize()
	r.initorReinitTextures(width, height, false)
	for _, pass := range r.renderPasses {
		pass.Resize(width, height, r.renderPassContext)
	}
}

func (r *RenderSystem) activeCloudTexture() *runtimeconfig.CloudTexture {
	return &r.app.RuntimeConfig().CloudTextures[r.app.RuntimeConfig().ActiveCloudTextureIndex]
}

func (r *RenderSystem) Render(delta time.Duration) {
	if r.resizeNextFrame {
		r.ReinitializeFrameBuffers()
		r.resizeNextFrame = false
	}
	mr := globals.ClientRegistry()
	initOpenGLRenderSettings()
	r.app.RuntimeConfig().TriangleDrawCount = 0
	r.app.RuntimeConfig().DrawCount = 0

	start := time.Now()
	cloudTexture := r.activeCloudTexture()
	if panels.RecreateCloudTexture {
		gl.DeleteTextures(1, &cloudTexture.WorleyTexture)
		gl.DeleteTextures(1, &cloudTexture.RenderTexture)
		gl.DeleteVertexArrays(1, &cloudTexture.VAO)
		gl.DeleteFramebuffers(1, &cloudTexture.FBO)
		cloudTexture.VAO, cloudTexture.WorleyTexture, cloudTexture.FBO, cloudTexture.RenderTexture = r.setupVolumetrics(r.shaderManager)
		panels.RecreateCloudTexture = false
	}
	r.renderVolumetrics(cloudTexture.VAO, cloudTexture.WorleyTexture, cloudTexture.FBO, r.shaderManager, r.app.AssetManager())
	mr.Inc("render_volumetrics", float64(time.Since(start).Milliseconds()))

	r.createMaterialTextures()

	// get the position and rotation of either the player camera or editor camera
	var position mgl64.Vec3
	var rotation mgl64.Quat
	if r.app.AppMode() == types.AppModeEditor {
		position = r.app.GetEditorCameraPosition()
		rotation = r.app.GetEditorCameraRotation()
	} else {
		camera := r.app.GetPlayerCamera()
		position = camera.Position()
		rotation = camera.Rotation()
	}

	renderContext, cameraViewerContext, lightViewerContext, lightContext := r.createRenderingContexts(position, rotation)

	start = time.Now()
	renderableEntities := r.fetchRenderableEntities(position, rotation, renderContext)
	mr.Inc("render_query_renderable", float64(time.Since(start).Milliseconds()))

	start = time.Now()
	shadowEntities := r.fetchShadowCastingEntities(position, rotation, renderContext)
	mr.Inc("render_query_shadowcasting", float64(time.Since(start).Milliseconds()))

	renderContext.RenderableEntities = renderableEntities
	renderContext.ShadowCastingEntities = shadowEntities
	renderContext.ShadowDistance = r.app.RuntimeConfig().Far * float32(settings.ShadowMapDistanceFactor)
	renderContext.BatchRenders = r.batchRenders

	// RENDER PASSES
	for _, pass := range r.renderPasses {
		pass.Render(renderContext, r.renderPassContext, cameraViewerContext, lightContext, lightViewerContext)
	}

	// store color picking entity
	start = time.Now()
	if r.app.AppMode() == types.AppModeEditor {
		r.hoveredEntityID = r.getEntityByPixelPosition(r.renderPassContext.MainFBO, r.app.GetFrameInput().MouseInput.Position)
	}
	mr.Inc("render_colorpicking_pick", float64(time.Since(start).Milliseconds()))

	var hdrColorTexture uint32

	if r.app.RuntimeConfig().Bloom {
		start = time.Now()
		r.downSample(r.renderPassContext.MainTexture, r.bloomTextureWidths, r.bloomTextureHeights)
		upsampleTexture := r.upSampleAndBlend(r.bloomTextureWidths, r.bloomTextureHeights)
		hdrColorTexture = r.composite(renderContext, r.renderPassContext.MainTexture, upsampleTexture)
		mr.Inc("render_bloom_pass", float64(time.Since(start).Milliseconds()))

		if menus.SelectedDebugComboOption == menus.ComboOptionBloom {
			r.app.RuntimeConfig().DebugTexture = upsampleTexture
		} else if menus.SelectedDebugComboOption == menus.ComboOptionPreBloomHDR {
			r.app.RuntimeConfig().DebugTexture = r.renderPassContext.MainTexture
		}
	} else {
		hdrColorTexture = r.renderPassContext.MainTexture
	}

	start = time.Now()
	r.postProcessingTexture = r.postProcess(renderContext,
		hdrColorTexture,
	)
	mr.Inc("render_post_process", float64(time.Since(start).Milliseconds()))

	r.setDebugTexture()

	// render to back buffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	start = time.Now()
	r.renderHelper(renderContext, imgui.TextureID(r.postProcessingTexture))
	mr.Inc("render_imgui", float64(time.Since(start).Milliseconds()))
}

func (r *RenderSystem) createRenderingContexts(position mgl64.Vec3, rotation mgl64.Quat) (context.RenderContext, context.ViewerContext, context.ViewerContext, context.LightContext) {
	mr := globals.ClientRegistry()

	start := time.Now()
	var renderContext context.RenderContext
	width, height := r.GameWindowSize()
	renderContext = context.NewRenderContext(width, height, float64(r.app.RuntimeConfig().FovX))

	// configure camera viewer context

	viewerViewMatrix := rotation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := context.ViewerContext{
		Position: position,
		Rotation: rotation,

		InverseViewMatrix:                   viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		InverseViewMatrixWithoutTranslation: viewerViewMatrix.Inv(),
		ProjectionMatrix:                    mgl64.Perspective(mgl64.DegToRad(renderContext.FovY()), renderContext.AspectRatio(), float64(r.app.RuntimeConfig().Near), float64(r.app.RuntimeConfig().Far)),
	}

	lightFrustumPoints := calculateFrustumPoints(
		position,
		rotation,
		float64(r.app.RuntimeConfig().Near),
		float64(r.app.RuntimeConfig().ShadowFarDistance),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		0,
	)

	// find the directional light if there is one
	lights := r.app.World().Lights()
	var directionalLights []*entities.Entity
	var pointLights []*entities.Entity

	for _, light := range lights {
		if light.LightInfo.Type == entities.LightTypeDirection {
			directionalLights = append(directionalLights, light)
		} else if light.LightInfo.Type == entities.LightTypePoint {
			pointLights = append(pointLights, light)
		}
	}

	var directionalLightX, directionalLightY, directionalLightZ float64 = 0, -1, 0
	if len(directionalLights) > 0 {
		directionalLightX = float64(directionalLights[0].LightInfo.Direction3F[0])
		directionalLightY = float64(directionalLights[0].LightInfo.Direction3F[1])
		directionalLightZ = float64(directionalLights[0].LightInfo.Direction3F[2])
	}

	lightRotation := utils.Vec3ToQuat(mgl64.Vec3{directionalLightX, directionalLightY, directionalLightZ})
	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightRotation.Mat4(), lightFrustumPoints, r.app.RuntimeConfig().ShadowmapZOffset)
	lightViewMatrix := mgl64.Translate3D(lightPosition.X(), lightPosition.Y(), lightPosition.Z()).Mul4(lightRotation.Mat4()).Inv()

	lightViewerContext := context.ViewerContext{
		Position:          lightPosition,
		Rotation:          lightRotation,
		InverseViewMatrix: lightViewMatrix,
		ProjectionMatrix:  lightProjectionMatrix,
	}

	lightContext := context.LightContext{
		// this should be the inverse of the transforms applied to the viewer context
		// if the viewer moves along -y, the universe moves along +y
		LightSpaceMatrix: lightProjectionMatrix.Mul4(lightViewMatrix),
		Lights:           r.app.World().Lights(),
		PointLights:      pointLights,
	}

	r.cameraViewerContext = cameraViewerContext
	mr.Inc("render_context_setup", float64(time.Since(start).Milliseconds()))

	return renderContext, cameraViewerContext, lightViewerContext, lightContext
}

func (r *RenderSystem) setDebugTexture() {
	if menus.SelectedDebugComboOption == menus.ComboOptionFinalRender {
		r.app.RuntimeConfig().DebugTexture = r.postProcessingTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionColorPicking {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.MainColorPickingTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionShadowDepthMap {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.ShadowMapTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionCameraDepthMap {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.CameraDepthTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionVolumetric {
		cloudTexture := r.activeCloudTexture()
		r.app.RuntimeConfig().DebugTexture = cloudTexture.RenderTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionSSAO {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.SSAOTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionGBufferPosition {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.GPositionTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionGBufferNormal {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.GNormalTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionSSAOBlur {
		r.app.RuntimeConfig().DebugTexture = r.renderPassContext.SSAOBlurTexture
		r.app.RuntimeConfig().DebugAspectRatio = 0
	} else if menus.SelectedDebugComboOption == menus.ComboOptionDebug {
		// r.app.RuntimeConfig().DebugTexture = r.renderPassContext.MultiSampleDebugTexture
		// r.app.RuntimeConfig().DebugAspectRatio = 0
	}
}

func (r *RenderSystem) fetchShadowCastingEntities(cameraPosition mgl64.Vec3, rotation mgl64.Quat, renderContext context.RenderContext) []*entities.Entity {
	frustumPoints := calculateFrustumPoints(
		cameraPosition,
		rotation,
		float64(r.app.RuntimeConfig().Near),
		float64(r.app.RuntimeConfig().Far),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		float64(r.app.RuntimeConfig().ShadowSpatialPartitionNearPlane),
	)

	sp := r.app.World().SpatialPartition()
	bb := collider.BoundingBoxFromVertices(frustumPoints)

	var result []*entities.Entity
	for _, spatialEntity := range sp.QueryEntities(bb) {
		e := r.app.World().GetEntityByID(spatialEntity.GetID()) // resolve fresh by ID
		if e.MeshComponent != nil && e.MeshComponent.ShadowCasting {
			result = append(result, e)
		}
	}
	return result
}

func (r *RenderSystem) fetchRenderableEntities(cameraPosition mgl64.Vec3, rotation mgl64.Quat, renderContext context.RenderContext) []*entities.Entity {
	frustumPoints := calculateFrustumPoints(
		cameraPosition,
		rotation,
		float64(r.app.RuntimeConfig().Near),
		float64(r.app.RuntimeConfig().Far),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		0,
	)

	sp := r.app.World().SpatialPartition()
	bb := collider.BoundingBoxFromVertices(frustumPoints)

	var result []*entities.Entity
	for _, spatialEntity := range sp.QueryEntities(bb) {
		e := r.app.World().GetEntityByID(spatialEntity.GetID()) // resolve fresh by ID
		if e.MeshComponent != nil {
			result = append(result, e)
		}
	}
	return result
}

func shouldSeedDefaultLayout(io imgui.IO, dockspaceID imgui.ID) bool {
	// Simple heuristic: if ini file exists and contains [Docking][Data], skip seeding.
	// (You can get the filename via io.IniFilename(), then read it with os.ReadFile)
	// For brevity, return true only when no ini file.
	fname := io.IniFilename()
	if fname == "" {
		return true
	}
	if _, err := os.Stat(fname); err != nil { // file missing
		return true
	}
	return false
}

var viewportInitialized bool

var firstLoad bool

func init() {
	firstLoad = true
}

func (r *RenderSystem) renderViewPort(renderContext context.RenderContext) {
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPos(viewport.Pos())
	imgui.SetNextWindowSize(viewport.Size())
	imgui.SetNextWindowViewport(viewport.ID())

	flags := imgui.WindowFlagsNoDocking |
		imgui.WindowFlagsNoTitleBar |
		imgui.WindowFlagsNoCollapse |
		imgui.WindowFlagsNoResize |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoBringToFrontOnFocus |
		imgui.WindowFlagsNoNavFocus

	if r.app.RuntimeConfig().UIEnabled {
		flags |= imgui.WindowFlagsMenuBar
	}

	var colorStyles []func() = []func(){
		func() { imgui.PushStyleColorVec4(imgui.ColTitleBg, InActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTitleBgActive, InActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTitleBgCollapsed, InActiveColorBg) },

		func() { imgui.PushStyleColorVec4(imgui.ColTab, InActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTabSelected, ActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTabHovered, HoveredHeaderColor) },
		func() { imgui.PushStyleColorVec4(imgui.ColTabDimmed, InActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTabDimmedSelected, ActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTabDimmedSelectedOverline, InActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColTabSelectedOverline, InActiveColorBg) },

		// resizing grips
		func() { imgui.PushStyleColorVec4(imgui.ColResizeGripActive, ResizeActiveColor) },
		func() { imgui.PushStyleColorVec4(imgui.ColResizeGripHovered, ResizeHoverColor) },

		// misc ui elements
		func() { imgui.PushStyleColorVec4(imgui.ColSliderGrab, InActiveColorControl) },
		func() { imgui.PushStyleColorVec4(imgui.ColSliderGrabActive, ActiveColorControl) },
		func() { imgui.PushStyleColorVec4(imgui.ColButton, InActiveColorControl) },
		func() { imgui.PushStyleColorVec4(imgui.ColButtonActive, ActiveColorControl) },
		func() { imgui.PushStyleColorVec4(imgui.ColButtonHovered, HoverColorControl) },
		func() { imgui.PushStyleColorVec4(imgui.ColCheckMark, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}) },

		// collapsable headers
		func() { imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}) },
		func() { imgui.PushStyleColorVec4(imgui.ColHeader, HeaderColor) },
		func() { imgui.PushStyleColorVec4(imgui.ColHeaderActive, HeaderColor) },
		func() { imgui.PushStyleColorVec4(imgui.ColHeaderHovered, HoveredHeaderColor) },

		// sliders
		func() { imgui.PushStyleColorVec4(imgui.ColFrameBg, InActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColFrameBgActive, ActiveColorBg) },
		func() { imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, HoverColorBg) },
	}
	for _, f := range colorStyles {
		f()
	}

	imgui.BeginV("##MainDockHost", nil, flags)

	dockspaceID := imgui.IDStr("MainDockSpace")
	imgui.DockSpace(dockspaceID)

	if !viewportInitialized && shouldSeedDefaultLayout(*imgui.CurrentIO(), dockspaceID) {
		viewportInitialized = true

		var rightID, mainAfterRight imgui.ID
		var bottomID, centerID imgui.ID
		var rightBottomID, rightTopID imgui.ID

		imgui.InternalDockBuilderSplitNode(
			dockspaceID,
			imgui.DirRight,
			0.22,
			&rightID,
			&mainAfterRight,
		)

		imgui.InternalDockBuilderSplitNode(
			rightID,
			imgui.DirUp,
			0.5,
			&rightTopID,
			&rightBottomID,
		)

		imgui.InternalDockBuilderSplitNode(
			mainAfterRight,
			imgui.DirDown,
			0.28,
			&bottomID,
			&centerID,
		)

		imgui.InternalDockBuilderDockWindow("Scene", centerID)
		imgui.InternalDockBuilderDockWindow("Hierarchy", rightTopID)
		imgui.InternalDockBuilderDockWindow("WorldProps", rightTopID)
		imgui.InternalDockBuilderDockWindow("Rendering", rightTopID)
		imgui.InternalDockBuilderDockWindow("Stats", rightTopID)
		imgui.InternalDockBuilderDockWindow("Inspector", rightBottomID)

		sceneNode := imgui.InternalDockBuilderGetNode(centerID)
		sceneNode.InternalSetLocalFlags(imgui.DockNodeFlags(imgui.DockNodeFlagsNoTabBar))

		imgui.InternalDockBuilderFinish(dockspaceID)
	}

	menus.SetupMenuBar(r.app, renderContext)
	windows.RenderWindows(r.app)

	imgui.End() // host window

	r.drawScene(renderContext)

	if r.app.RuntimeConfig().UIEnabled {
		r.drawInspector()
		r.drawRightTop(renderContext)
		drawer.BuildFooter(
			r.app,
			renderContext,
			r.sceneSize[0],
			r.materialTextureMap,
		)

		if r.app.RuntimeConfig().ShowTextureViewer {
			imgui.SetNextWindowSizeV(imgui.Vec2{X: 400}, imgui.CondFirstUseEver)
			if imgui.BeginV("Texture Viewer", &r.app.RuntimeConfig().ShowTextureViewer, imgui.WindowFlagsNone) {
				if imgui.BeginCombo("##", string(menus.SelectedDebugComboOption)) {
					for _, option := range menus.DebugComboOptions {
						if imgui.SelectableBool(string(option)) {
							menus.SelectedDebugComboOption = option
						}
					}
					imgui.EndCombo()
				}

				regionSize := imgui.ContentRegionAvail()
				imageWidth := regionSize.X

				texture := imgui.TextureID(r.app.RuntimeConfig().DebugTexture)
				aspectRatio := float32(renderContext.AspectRatio())
				if r.app.RuntimeConfig().DebugAspectRatio != 0 {
					aspectRatio = float32(r.app.RuntimeConfig().DebugAspectRatio)
				}
				size := imgui.Vec2{X: imageWidth, Y: imageWidth / aspectRatio}
				if menus.SelectedDebugComboOption == menus.ComboOptionVolumetric {
					size.Y = imageWidth
				}
				// invert the Y axis since opengl vs texture coordinate systems differ
				// https://learnopengl.com/Getting-started/Textures
				imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
			}
			imgui.End()
		}
	}

	if firstLoad {
		firstLoad = false
		imgui.SetWindowFocusStr("Hierarchy")
	}

	imgui.PopStyleColorV(int32(len(colorStyles)))

	if r.app.RuntimeConfig().UIEnabled {
		if r.app.RuntimeConfig().ShowImguiDemo {
			imgui.ShowDemoWindow()
		}
	}
}

func (r *RenderSystem) drawScene(renderContext context.RenderContext) {
	r.gameWindowHovered = false
	imgui.BeginV("Scene", nil, imgui.WindowFlagsNoScrollbar)
	if imgui.IsWindowHovered() {
		r.gameWindowHovered = true
	}
	texture := imgui.TextureID(r.postProcessingTexture)

	sceneSize := imgui.ContentRegionAvail()
	nextSceneSize := [2]int{int(sceneSize.X), int(sceneSize.Y)}
	// if nextSceneSize != r.sceneSize && time.Since(r.lastResize).Milliseconds() > 25 {
	if nextSceneSize != r.sceneSize {
		r.sceneSize = nextSceneSize
		r.resizeNextFrame = true
		r.lastResize = time.Now()
	}

	size := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height())}
	imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
	imgui.End()
}

func (r *RenderSystem) drawInspector() {
	imgui.Begin("Inspector")
	panels.EntityProps(r.app.SelectedEntity(), r.app)
	imgui.End()
}

func (r *RenderSystem) drawRightTop(renderContext context.RenderContext) {
	imgui.Begin("Hierarchy")
	panels.SceneGraph(r.app)
	imgui.End()
	imgui.Begin("WorldProps")
	panels.WorldProps(r.app)
	imgui.End()
	imgui.Begin("Stats")
	panels.Stats(r.app, renderContext)
	imgui.End()
	imgui.Begin("Rendering")
	panels.Rendering(r.app)
	imgui.End()
}

func (r *RenderSystem) renderHelper(renderContext context.RenderContext, gameWindowTexture imgui.TextureID) {
	r.app.Platform().NewFrame()
	imgui.NewFrame()

	r.renderViewPort(renderContext)

	imgui.Render()
	r.imguiRenderer.Render(r.app.Platform().DisplaySize(), r.app.Platform().FramebufferSize(), imgui.CurrentDrawData())
}

func (r *RenderSystem) GameWindowHovered() bool {
	return r.gameWindowHovered
}

func (r *RenderSystem) HoveredEntityID() *int {
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

func (r *RenderSystem) QueueCreateMaterialTexture(handle types.MaterialHandle) {
	r.materialTextureQueue = append(r.materialTextureQueue, handle)
}

func (r *RenderSystem) createMaterialTextures() {
	for _, material := range r.app.AssetManager().GetMaterials() {
		if _, ok := r.materialTextureMap[material.Handle]; ok {
			continue
		}
		r.CreateMaterialTexture(material.Handle)
	}

	// queued texture creations (e.g. from a material being updated)
	for _, materialHandle := range r.materialTextureQueue {
		r.CreateMaterialTexture(materialHandle)
	}
	r.materialTextureQueue = []types.MaterialHandle{}
}
