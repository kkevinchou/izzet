package render

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/render/menus"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/izzet/lib"
	"github.com/kkevinchou/kitolib/animation"
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

const mipsCount int = 6
const MaxBloomTextureWidth int = 1920
const MaxBloomTextureHeight int = 1080

type Renderer struct {
	app           renderiface.App
	world         GameWorld
	shaderManager *shaders.ShaderManager

	shadowMap           *ShadowMap
	imguiRenderer       *ImguiOpenGL4Renderer
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

	downSampleFBO      uint32
	xyTextureVAO       uint32
	downSampleTextures []uint32

	upSampleFBO         uint32
	upSampleTextures    []uint32
	blendTargetTextures []uint32

	compositeFBO     uint32
	compositeTexture uint32

	blendFBO uint32

	bloomTextureWidths  []int
	bloomTextureHeights []int

	cubeVAOs     map[float32]uint32
	triangleVAOs map[string]uint32

	width, height int
}

func New(app renderiface.App, world GameWorld, shaderDirectory string, width, height int) *Renderer {
	r := &Renderer{app: app, world: world, width: width, height: height}
	r.shaderManager = shaders.NewShaderManager(shaderDirectory)
	compileShaders(r.shaderManager)

	imguiIO := imgui.CurrentIO()
	imguiRenderer, err := NewImguiOpenGL4Renderer(imguiIO)
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
	r.cubeVAOs = map[float32]uint32{}
	r.triangleVAOs = map[string]uint32{}

	r.initMainRenderFBO(width, height)
	r.initDepthMapFBO(width, height)

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

	r.initCompositeFBO(width, height)
	r.renderCircle()

	return r
}

func (r *Renderer) Resized(width, height int) {
	r.width, r.height = width, height
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
	renderFBO, colorTextures := r.initFrameBuffer(width, height, []int32{gl.R11F_G11F_B10F, gl.R32UI}, []uint32{gl.RGBA, gl.RED_INTEGER})
	r.renderFBO = renderFBO
	r.mainColorTexture = colorTextures[0]
	r.colorPickingTexture = colorTextures[1]
	r.colorPickingAttachment = gl.COLOR_ATTACHMENT1
}

func (r *Renderer) initCompositeFBO(width, height int) {
	r.compositeFBO, r.compositeTexture = r.initFBOAndTexture(width, height)
}

func (r *Renderer) Render(delta time.Duration, renderContext RenderContext) {
	initOpenGLRenderSettings()
	r.app.RuntimeConfig().TriangleDrawCount = 0
	r.app.RuntimeConfig().DrawCount = 0

	// configure camera viewer context

	var position mgl64.Vec3
	var rotation mgl64.Quat = mgl64.QuatIdent()

	if r.app.AppMode() == app.AppModeEditor {
		position = r.app.GetEditorCameraPosition()
		rotation = r.app.GetEditorCameraRotation()
	} else {
		camera := r.app.GetPlayerCamera()
		position = camera.WorldPosition()
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
	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightRotation.Mat4(), lightFrustumPoints, settings.ShadowmapZOffset)
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

	renderEntities := r.fetchRenderableEntities(position, rotation, renderContext)
	shadowEntities := r.fetchShadowCastingEntities(position, rotation, renderContext)

	r.drawSkybox(renderContext)
	_ = lightViewerContext
	r.drawToShadowDepthMap(lightViewerContext, shadowEntities)
	r.drawToCubeDepthMap(lightContext, shadowEntities)
	r.drawToCameraDepthMap(cameraViewerContext, renderEntities)

	// main color FBO
	r.drawToMainColorBuffer(cameraViewerContext, lightContext, renderContext, renderEntities)
	r.drawAnnotations(cameraViewerContext, lightContext, renderContext)

	// clear depth for gizmo rendering
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	r.renderGizmos(cameraViewerContext, renderContext)

	var finalRenderTexture uint32
	if r.app.RuntimeConfig().Bloom {
		r.downSample(r.mainColorTexture, r.bloomTextureWidths, r.bloomTextureHeights)
		upsampleTexture := r.upSample(r.bloomTextureWidths, r.bloomTextureHeights)
		finalRenderTexture = r.composite(renderContext, r.mainColorTexture, upsampleTexture)
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
	r.drawTexturedQuad(&cameraViewerContext, r.shaderManager, finalRenderTexture, float32(renderContext.aspectRatio), nil, false, nil)

	r.renderImgui(renderContext)
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
		float64(r.app.RuntimeConfig().SPNearPlaneOffset),
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

func (r *Renderer) drawAnnotations(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext) {
	shaderManager := r.shaderManager

	// joint rendering for the selected entity
	entity := panels.SelectedEntity()
	if entity != nil {
		// draw bounding box
		if entity.HasBoundingBox() {
			r.drawAABB(
				viewerContext,
				shaderManager.GetShaderProgram("flat"),
				mgl64.Vec3{.2, 0, .7},
				entity.BoundingBox(),
				0.5,
			)
		}

	}

	if r.app.AppMode() == app.AppModeEditor {
		for _, entity := range r.world.Entities() {
			lightInfo := entity.LightInfo
			if lightInfo != nil {
				if lightInfo.Type == 0 {
					shader := shaderManager.GetShaderProgram("flat")
					color := mgl64.Vec3{252.0 / 255, 241.0 / 255, 33.0 / 255}

					direction3F := lightInfo.Direction3F
					dir := mgl64.Vec3{float64(direction3F[0]), float64(direction3F[1]), float64(direction3F[2])}.Mul(50)
					// directional light arrow
					lines := [][]mgl64.Vec3{
						[]mgl64.Vec3{
							entity.WorldPosition(),
							entity.WorldPosition().Add(dir),
						},
					}
					r.drawLines(viewerContext, shader, lines, 0.5, color)
				}
			}
		}
	}

	if r.app.RuntimeConfig().EnableSpatialPartition && r.app.RuntimeConfig().RenderSpatialPartition {
		r.drawSpatialPartition(viewerContext, r.shaderManager.GetShaderProgram("flat"), mgl64.Vec3{0, 1, 0}, r.world.SpatialPartition(), 0.5)
	}

	// modelMatrix := entities.WorldTransform(entity)
	// TODO: optimize this - can probably cache some of these computations

	// 		// draw joint
	// 		if len(panels.JointsToRender) > 0 && entity.AnimationPlayer != nil && entity.AnimationPlayer.CurrentAnimation() != "" {
	// 			jointShader := shaderManager.GetShaderProgram("flat")
	// 			color := mgl64.Vec3{0 / 255, 255.0 / 255, 85.0 / 255}

	// 			var jointLines [][]mgl64.Vec3
	// 			model := entity.Model
	// 			animationTransforms := entity.AnimationPlayer.AnimationTransforms()

	// 			for _, jid := range panels.JointsToRender {
	// 				jointTransform := animationTransforms[jid]
	// 				lines := cubeLines(15)
	// 				jt := utils.Mat4F32ToF64(jointTransform)
	// 				for _, line := range lines {
	// 					points := line
	// 					for i := 0; i < len(points); i++ {
	// 						bindTransform := model.JointMap()[jid].FullBindTransform
	// 						// The calculated joint transforms apply to joints in bind space
	// 						// 		i.e. the calculated transforms are computed as:
	// 						// 			parent3 transform * parent2 transform * parent1 transform * local joint transform * inverse bind transform * vertex
	// 						//
	// 						// so, to bring the cube into the joint's bind space (i.e. 0,0,0 is right where the joint is positioned rather than the world origin),
	// 						// we need to multiply by the full bind transform. this is composed of each parent's bind transform. however, GLTF already exports the
	// 						// inverse bind matrix which is the inverse of it. so we can just inverse the inverse (which we store as FullBindTransform)
	// 						points[i] = jt.Mul4(utils.Mat4F32ToF64(bindTransform)).Mul4x1(points[i].Vec4(1)).Vec3()
	// 						// points[i] = jt.Mul4x1(points[i].Vec4(1)).Vec3()
	// 					}
	// 				}
	// 				jointLines = append(jointLines, lines...)
	// 			}

	// 			for _, line := range jointLines {
	// 				points := line
	// 				for i := 0; i < len(points); i++ {
	// 					points[i] = modelMatrix.Mul4x1(points[i].Vec4(1)).Vec3()
	// 				}
	// 			}

	// 			drawLines(viewerContext, jointShader, jointLines, 0.5, color)
	// 		}

	// 	nm := r.app.NavMesh()

	// 	if nm != nil {
	// 		// draw bounding box
	// 		volume := nm.Volume
	// 		drawAABB(
	// 			viewerContext,
	// 			shaderManager.GetShaderProgram("flat"),
	// 			mgl64.Vec3{155.0 / 99, 180.0 / 255, 45.0 / 255},
	// 			&volume,
	// 			0.5,
	// 		)

	// 		// draw navmesh
	// 		if nm.VoxelCount() > 0 {
	// 			shader := shaderManager.GetShaderProgram("color_pbr")
	// 			shader.Use()

	// 			if r.app.RuntimeConfig().Bloom {
	// 				shader.SetUniformInt("applyToneMapping", 0)
	// 			} else {
	// 				// only tone map if we're not applying bloom, otherwise
	// 				// we want to keep the HDR values and tone map later
	// 				shader.SetUniformInt("applyToneMapping", 1)
	// 			}

	// 			shader.SetUniformMat4("model", mgl32.Ident4())
	// 			shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	// 			shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	// 			shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	// 			shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
	// 			shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	// 			shader.SetUniformFloat("ambientFactor", r.app.RuntimeConfig().AmbientFactor)
	// 			shader.SetUniformInt("shadowMap", 31)
	// 			shader.SetUniformInt("depthCubeMap", 30)
	// 			shader.SetUniformFloat("bias", r.app.RuntimeConfig().PointLightBias)
	// 			shader.SetUniformFloat("far_plane", float32(settings.DepthCubeMapFar))
	// 			shader.SetUniformInt("isAnimated", 0)
	// 			shader.SetUniformInt("hasColorOverride", 1)

	// 			// color := mgl32.Vec3{9.0 / 255, 235.0 / 255, 47.0 / 255}
	// 			color := mgl32.Vec3{3.0 / 255, 185.0 / 255, 5.0 / 255}
	// 			// color := mgl32.Vec3{200.0 / 255, 1000.0 / 255, 200.0 / 255}
	// 			shader.SetUniformVec3("albedo", color)
	// 			shader.SetUniformInt("hasPBRMaterial", 1)
	// 			shader.SetUniformFloat("ao", 1.0)
	// 			shader.SetUniformInt("hasPBRBaseColorTexture", 0)
	// 			shader.SetUniformFloat("roughness", r.app.RuntimeConfig().Roughness)
	// 			shader.SetUniformFloat("metallic", r.app.RuntimeConfig().Metallic)

	// 			setupLightingUniforms(shader, lightContext.Lights)

	// 			gl.ActiveTexture(gl.TEXTURE30)
	// 			gl.BindTexture(gl.TEXTURE_CUBE_MAP, r.depthCubeMapTexture)

	// 			gl.ActiveTexture(gl.TEXTURE31)
	// 			gl.BindTexture(gl.TEXTURE_2D, r.shadowMap.DepthTexture())

	// 			drawNavMeshTris(viewerContext, nm)
	// 		}
	// 	}
}

func (r *Renderer) drawToCameraDepthMap(viewerContext ViewerContext, renderableEntities []*entities.Entity) {
	gl.Viewport(0, 0, int32(r.width), int32(r.height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.cameraDepthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	r.renderGeometryWithoutColor(viewerContext, renderableEntities)
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

	r.renderGeometryWithoutColor(viewerContext, renderableEntities)
}

func (r *Renderer) renderGeometryWithoutColor(viewerContext ViewerContext, renderableEntities []*entities.Entity) {
	shader := r.shaderManager.GetShaderProgram("modelgeo")
	shader.Use()

	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

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

		primitives := r.app.ModelLibrary().GetPrimitives(entity.MeshComponent.MeshHandle)
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

	position := pointLight.WorldPosition()
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

		primitives := r.app.ModelLibrary().GetPrimitives(entity.MeshComponent.MeshHandle)
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
					if entity.Billboard && r.app.AppMode() == app.AppModeEditor {
						shader := shaderManager.GetShaderProgram("world_space_quad")
						shader.Use()

						position := entity.WorldPosition()
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

				lines := [][]mgl64.Vec3{
					{entity.WorldPosition().Add(rightVector), entity.WorldPosition().Add(entity.CharacterControllerComponent.WebVector)},
				}
				r.drawLines(viewerContext, shaderManager.GetShaderProgram("flat"), lines, 1, mgl64.Vec3{1, 0, 0})
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

	shader.SetUniformInt("width", int32(r.width))
	shader.SetUniformInt("height", int32(r.height))
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

		modelMatrix := entities.WorldTransform(entity)
		shader.SetUniformUInt("entityID", uint32(entity.ID))

		var animationPlayer *animation.AnimationPlayer
		if entity.Animation != nil {
			animationPlayer = entity.Animation.AnimationPlayer
		}

		r.drawModel(
			viewerContext,
			lightContext,
			r.shadowMap,
			shader,
			r.app.AssetManager(),
			animationPlayer,
			modelMatrix,
			r.depthCubeMapTexture,
			entity.ID,
			entity.Material,
			r.app.ModelLibrary(),
			entity,
		)
	}

	for _, entity := range renderableEntities {
		if entity == nil || entity.Material != nil || entity.Collider == nil {
			continue
		}

		if r.app.RuntimeConfig().RenderColliders {
			capsuleCollider := entity.Collider.CapsuleCollider
			if capsuleCollider != nil {
				shader := shaderManager.GetShaderProgram("flat")
				color := mgl64.Vec3{255.0 / 255, 147.0 / 255, 12.0 / 255}

				transform := entities.WorldTransform(entity)
				capsuleCollider := capsuleCollider.Transform(transform)

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

				r.drawLines(viewerContext, shader, lines, 0.5, color)
			}
		}
	}
}

func (r *Renderer) renderImgui(renderContext RenderContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.app.Platform().NewFrame()
	imgui.NewFrame()

	menuBarSize := menus.SetupMenuBar(r.app)
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
		imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
		imgui.PushStyleColor(imgui.StyleColorHeader, settings.HeaderColor)
		imgui.PushStyleColor(imgui.StyleColorHeaderActive, settings.HeaderColor)
		imgui.PushStyleColor(imgui.StyleColorHeaderHovered, settings.HoveredHeaderColor)
		imgui.PushStyleColor(imgui.StyleColorTitleBg, settings.TitleColor)
		imgui.PushStyleColor(imgui.StyleColorTitleBgActive, settings.TitleColor)
		imgui.PushStyleColor(imgui.StyleColorSliderGrab, settings.InActiveColorControl)
		imgui.PushStyleColor(imgui.StyleColorSliderGrabActive, settings.ActiveColorControl)
		imgui.PushStyleColor(imgui.StyleColorFrameBg, settings.InActiveColorBg)
		imgui.PushStyleColor(imgui.StyleColorFrameBgActive, settings.ActiveColorBg)
		imgui.PushStyleColor(imgui.StyleColorFrameBgHovered, settings.HoverColorBg)
		imgui.PushStyleColor(imgui.StyleColorCheckMark, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
		imgui.PushStyleColor(imgui.StyleColorButton, settings.InActiveColorControl)
		imgui.PushStyleColor(imgui.StyleColorButtonActive, settings.ActiveColorControl)
		imgui.PushStyleColor(imgui.StyleColorButtonHovered, settings.HoverColorControl)
		imgui.PushStyleColor(imgui.StyleColorTabActive, settings.ActiveColorBg)
		imgui.PushStyleColor(imgui.StyleColorTabUnfocused, settings.InActiveColorBg)
		imgui.PushStyleColor(imgui.StyleColorTabUnfocusedActive, settings.InActiveColorBg)
		imgui.PushStyleColor(imgui.StyleColorTab, settings.InActiveColorBg)
		imgui.PushStyleColor(imgui.StyleColorTabHovered, settings.HoveredHeaderColor)

		panels.BuildContentBrowser(
			r.app,
			r.world,
			renderContext,
			menuBarSize,
			r.app.Prefabs(),
		)

		panels.BuildTabsSet(
			r.app,
			r.world,
			renderContext,
			menuBarSize,
			r.app.Prefabs(),
		)

		imgui.PopStyleColorV(20)
		imgui.PopStyleVarV(7)

		if r.app.ShowImguiDemo() {
			imgui.ShowDemoWindow(nil)
		}
	}

	imgui.Render()
	r.imguiRenderer.Render(r.app.Platform().DisplaySize(), r.app.Platform().FramebufferSize(), imgui.RenderedDrawData())
}

func (r *Renderer) renderGizmos(viewerContext ViewerContext, renderContext RenderContext) {
	if panels.SelectedEntity() == nil {
		return
	}

	entity := r.world.GetEntityByID(panels.SelectedEntity().ID)
	position := entity.WorldPosition()

	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		r.drawTranslationGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat2"), position)
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
