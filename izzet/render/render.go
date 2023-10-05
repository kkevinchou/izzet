package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/menus"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/lib"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

type World interface {
	AssetManager() *assets.AssetManager
	ModelLibrary() *modellibrary.ModelLibrary
	Camera() *camera.Camera
	Prefabs() []*prefabs.Prefab
	Entities() []*entities.Entity
	Lights() []*entities.Entity
	GetEntityByID(id int) *entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
	NavMesh() *navmesh.NavigationMesh
	ResetNavMeshVAO()

	// for panels
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
	Platform() *input.SDLPlatform

	SetShowImguiDemo(bool)
	ShowImguiDemo() bool

	Serializer() *serialization.Serializer
	LoadWorld()
	SaveWorld()
}

const mipsCount int = 6
const MaxBloomTextureWidth int = 1920
const MaxBloomTextureHeight int = 1080

type Renderer struct {
	world         World
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

	viewerContext ViewerContext

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

	cubeVAOs     map[int]uint32
	triangleVAOs map[string]uint32

	width, height int
}

func New(world World, shaderDirectory string, width, height int) *Renderer {
	r := &Renderer{world: world, width: width, height: height}
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
	settings.RuntimeMaxTextureSize = int(float32(maxTextureSize) * .90)

	// shadowMap, err := NewShadowMap(settings.RuntimeMaxTextureSize, settings.RuntimeMaxTextureSize, float64(panels.DBG.Far))
	dimension := 8640
	shadowMap, err := NewShadowMap(dimension, dimension, float64(panels.DBG.Far))
	if err != nil {
		panic(fmt.Sprintf("failed to create shadow map %s", err))
	}
	r.shadowMap = shadowMap
	r.depthCubeMapFBO, r.depthCubeMapTexture = lib.InitDepthCubeMap()
	r.xyTextureVAO = r.init2f2fVAO()
	r.cubeVAOs = map[int]uint32{}
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
	panels.DBG.TriangleDrawCount = 0
	panels.DBG.DrawCount = 0

	// configure camera viewer context
	position := r.world.Camera().Position
	orientation := r.world.Camera().Orientation

	viewerViewMatrix := orientation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := ViewerContext{
		Position:    position,
		Orientation: orientation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(renderContext.FovY()), renderContext.AspectRatio(), float64(panels.DBG.Near), float64(panels.DBG.Far)),
	}

	lightFrustumPoints := calculateFrustumPoints(position, orientation, float64(panels.DBG.Near), float64(panels.DBG.Far), renderContext.FovX(), renderContext.FovY(), renderContext.AspectRatio(), float64(panels.DBG.ShadowFarFactor))

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

	lightOrientation := utils.Vec3ToQuat(mgl64.Vec3{directionalLightX, directionalLightY, directionalLightZ})
	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightOrientation.Mat4(), lightFrustumPoints, settings.ShadowmapZOffset)
	lightViewMatrix := mgl64.Translate3D(lightPosition.X(), lightPosition.Y(), lightPosition.Z()).Mul4(lightOrientation.Mat4()).Inv()

	lightViewerContext := ViewerContext{
		Position:          lightPosition,
		Orientation:       lightOrientation,
		InverseViewMatrix: lightViewMatrix,
		ProjectionMatrix:  lightProjectionMatrix,
	}

	lightContext := LightContext{
		// this should be the inverse of the transforms applied to the viewer context
		// if the viewer moves along -y, the universe moves along +y
		LightSpaceMatrix: lightProjectionMatrix.Mul4(lightViewMatrix),
		Lights:           r.world.Lights(),
	}

	r.viewerContext = cameraViewerContext

	r.clearMainFrameBuffer(renderContext)

	renderEntities := r.fetchRenderableEntities(position, orientation, renderContext)

	r.drawSkybox(renderContext)
	r.drawToShadowDepthMap(lightViewerContext, renderEntities)
	r.drawToCubeDepthMap(lightContext, renderEntities)
	r.drawToCameraDepthMap(cameraViewerContext, renderEntities)

	r.drawToMainColorBuffer(cameraViewerContext, lightContext, renderContext, renderEntities)
	r.drawAnnotations(cameraViewerContext, lightContext, renderContext)

	if panels.DBG.EnableSpatialPartition && panels.DBG.RenderSpatialPartition {
		drawSpatialPartition(cameraViewerContext, r.shaderManager.GetShaderProgram("flat"), mgl64.Vec3{0, 1, 0}, r.world.SpatialPartition(), 0.5)
	}

	var finalRenderTexture uint32
	if panels.DBG.Bloom {
		r.downSample(r.mainColorTexture, r.bloomTextureWidths, r.bloomTextureHeights)
		upsampleTexture := r.upSample(r.bloomTextureWidths, r.bloomTextureHeights)
		finalRenderTexture = r.composite(renderContext, r.mainColorTexture, upsampleTexture)
		if panels.SelectedComboOption == panels.ComboOptionFinalRender {
			panels.DBG.DebugTexture = finalRenderTexture
		} else if panels.SelectedComboOption == panels.ComboOptionColorPicking {
			panels.DBG.DebugTexture = r.colorPickingTexture
		} else if panels.SelectedComboOption == panels.ComboOptionHDR {
			panels.DBG.DebugTexture = r.mainColorTexture
		} else if panels.SelectedComboOption == panels.ComboOptionBloom {
			panels.DBG.DebugTexture = upsampleTexture
		} else if panels.SelectedComboOption == panels.ComboOptionShadowDepthMap {
			panels.DBG.DebugTexture = r.shadowMap.depthTexture
		} else if panels.SelectedComboOption == panels.ComboOptionCameraDepthMap {
			panels.DBG.DebugTexture = r.cameraDepthTexture
		}
	} else {
		finalRenderTexture = r.mainColorTexture
		if panels.SelectedComboOption == panels.ComboOptionFinalRender {
			panels.DBG.DebugTexture = finalRenderTexture
		} else if panels.SelectedComboOption == panels.ComboOptionColorPicking {
			panels.DBG.DebugTexture = r.colorPickingTexture
		} else if panels.SelectedComboOption == panels.ComboOptionHDR {
			panels.DBG.DebugTexture = 0
		} else if panels.SelectedComboOption == panels.ComboOptionBloom {
			panels.DBG.DebugTexture = 0
		} else if panels.SelectedComboOption == panels.ComboOptionShadowDepthMap {
			panels.DBG.DebugTexture = r.shadowMap.depthTexture
		} else if panels.SelectedComboOption == panels.ComboOptionCameraDepthMap {
			panels.DBG.DebugTexture = r.cameraDepthTexture
		}
	}

	// r.ReadAllPixels()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawTexturedQuad(&cameraViewerContext, r.shaderManager, finalRenderTexture, float32(renderContext.aspectRatio), nil, false)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.renderGizmos(cameraViewerContext, renderContext)
	r.renderImgui(renderContext)
}

func (r *Renderer) fetchRenderableEntities(cameraPosition mgl64.Vec3, orientation mgl64.Quat, renderContext RenderContext) []*entities.Entity {
	frustumPoints := calculateFrustumPoints(
		cameraPosition,
		orientation,
		float64(panels.DBG.Near),
		float64(panels.DBG.Far),
		renderContext.FovX(),
		renderContext.FovY(),
		renderContext.AspectRatio(),
		1,
	)
	frustumBoundingBox := *collider.BoundingBoxFromVertices(frustumPoints)

	var renderEntities []*entities.Entity
	if panels.DBG.EnableSpatialPartition {
		spatialPartition := r.world.SpatialPartition()
		frustumEntities := spatialPartition.QueryEntities(frustumBoundingBox)
		for _, entity := range frustumEntities {
			renderEntities = append(renderEntities, r.world.GetEntityByID(entity.GetID()))
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

		// draw bounding box
		bb := entity.BoundingBox()
		if bb != nil {
			drawAABB(
				viewerContext,
				shaderManager.GetShaderProgram("flat"),
				mgl64.Vec3{.2, 0, .7},
				bb,
				0.5,
			)
		}
	}

	// 	nm := r.world.NavMesh()

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

	// 			if panels.DBG.Bloom {
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
	// 			shader.SetUniformFloat("ambientFactor", panels.DBG.AmbientFactor)
	// 			shader.SetUniformInt("shadowMap", 31)
	// 			shader.SetUniformInt("depthCubeMap", 30)
	// 			shader.SetUniformFloat("bias", panels.DBG.PointLightBias)
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
	// 			shader.SetUniformFloat("roughness", panels.DBG.Roughness)
	// 			shader.SetUniformFloat("metallic", panels.DBG.Metallic)

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
	defer resetGLRenderSettings(r.renderFBO)

	gl.Viewport(0, 0, int32(r.width), int32(r.height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.cameraDepthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	r.renderGeometryWithoutColor(viewerContext, renderableEntities)
}

func (r *Renderer) drawToShadowDepthMap(viewerContext ViewerContext, renderableEntities []*entities.Entity) {
	defer resetGLRenderSettings(r.renderFBO)

	r.shadowMap.Prepare()

	if !panels.DBG.EnableShadowMapping {
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

		for _, renderData := range getRenderData(r.world.ModelLibrary(), entity) {
			shader.SetUniformMat4("model", m32ModelMatrix.Mul4(renderData.Transform))

			gl.BindVertexArray(renderData.GeometryVAO)
			iztDrawElements(int32(len(renderData.Primitive.VertexIndices)))
		}
	}
}

func (r *Renderer) drawToCubeDepthMap(lightContext LightContext, renderableEntities []*entities.Entity) {
	defer resetGLRenderSettings(r.renderFBO)

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

		for _, renderData := range getRenderData(r.world.ModelLibrary(), entity) {
			shader.SetUniformMat4("model", m32ModelMatrix.Mul4(renderData.Transform))

			gl.BindVertexArray(renderData.GeometryVAO)
			iztDrawElements(int32(len(renderData.Primitive.VertexIndices)))
		}
	}
}

// drawToMainColorBuffer renders a scene from the perspective of a viewer
func (r *Renderer) drawToMainColorBuffer(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext, renderableEntities []*entities.Entity) {
	defer resetGLRenderSettings(r.renderFBO)

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))

	r.renderModels(viewerContext, lightContext, renderContext, renderableEntities)

	shaderManager := r.shaderManager

	// render non-models
	for _, entity := range r.world.Entities() {
		if entity.MeshComponent != nil {
			continue
		}

		modelMatrix := entities.WorldTransform(entity)

		if len(entity.ShapeData) > 0 {
			shader := shaderManager.GetShaderProgram("flat")
			shader.Use()

			shader.SetUniformUInt("entityID", uint32(entity.ID))
			shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
			shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
			shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

			for _, shapeData := range entity.ShapeData {
				// if shapeData.Cube != nil {
				// 	var ok bool
				// 	var vao uint32
				// 	if vao, ok = r.cubeVAOs[shapeData.Cube.Length]; !ok {
				// 		vao = initCubeVAO(int(shapeData.Cube.Length))
				// 		r.cubeVAOs[shapeData.Cube.Length] = vao
				// 	}

				// 	gl.BindVertexArray(vao)
				// 	shader.SetUniformVec3("color", panels.DBG.Color)
				// 	shader.SetUniformFloat("intensity", panels.DBG.ColorIntensity)
				// 	iztDrawArrays(0, 48)
				// }

				if shapeData.Triangle != nil {
					var ok bool
					var vao uint32
					if vao, ok = r.triangleVAOs[triangleVAOKey(*shapeData.Triangle)]; !ok {
						vao = initTriangleVAO(shapeData.Triangle.V1, shapeData.Triangle.V2, shapeData.Triangle.V3)
						r.triangleVAOs[triangleVAOKey(*shapeData.Triangle)] = vao
					}

					gl.BindVertexArray(vao)
					shader.SetUniformVec3("color", mgl32.Vec3{0, 1, 1})
					shader.SetUniformFloat("intensity", panels.DBG.ColorIntensity)
					iztDrawArrays(0, 6)
				}
			}
		}

		if entity.ImageInfo != nil {
			textureName := strings.Split(entity.ImageInfo.ImageName, ".")[0]
			texture := r.world.AssetManager().GetTexture(textureName)
			if texture != nil {
				if entity.Billboard {
					shader := shaderManager.GetShaderProgram("world_space_quad")
					shader.Use()

					shader.SetUniformUInt("entityID", uint32(entity.ID))
					shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix.Mul4(r.world.Camera().Orientation.Mat4())))
					shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
					shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

					drawBillboardTexture(texture.ID, 1)
				}
			} else {
				fmt.Println("couldn't find texture", "light")
			}
		}

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
				drawLines(viewerContext, shader, lines, 0.5, color)
			}
		}

		particles := entity.Particles
		if particles != nil {
			texture := r.world.AssetManager().GetTexture("light").ID
			for _, particle := range particles.GetActiveParticles() {
				particleModelMatrix := mgl32.Translate3D(float32(particle.Position.X()), float32(particle.Position.Y()), float32(particle.Position.Z()))
				drawTexturedQuad(&viewerContext, r.shaderManager, texture, float32(renderContext.AspectRatio()), &particleModelMatrix, true)
			}
		}

		collider := entity.Collider
		if collider != nil {
			localPosition := entities.LocalPosition(entity)
			translation := mgl64.Translate3D(localPosition.X(), localPosition.Y(), localPosition.Z())
			// lots of hacky rendering stuff to get the rectangle to billboard
			center := mgl64.Vec3{localPosition.X(), 0, localPosition.Z()}
			viewerArtificialCenter := mgl64.Vec3{viewerContext.Position.X(), 0, viewerContext.Position.Z()}
			vecToViewer := viewerArtificialCenter.Sub(center).Normalize()
			billboardModelMatrix := translation.Mul4(mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, 1}, vecToViewer).Mat4())
			drawCapsuleCollider(
				viewerContext,
				lightContext,
				shaderManager.GetShaderProgram("flat"),
				mgl64.Vec3{0.5, 1, 0},
				collider.CapsuleCollider,
				billboardModelMatrix,
			)
		}
	}
}

func (r *Renderer) renderModels(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext, renderableEntities []*entities.Entity) {
	shaderManager := r.shaderManager

	shader := shaderManager.GetShaderProgram("modelpbr")
	shader.Use()

	if !panels.DBG.Bloom {
		// only tone map if we're not applying bloom, otherwise
		// we want to keep the HDR values and tone map later
		shader.SetUniformInt("applyToneMapping", 1)
	} else {
		shader.SetUniformInt("applyToneMapping", 0)
	}

	if panels.DBG.FogEnabled {
		shader.SetUniformInt("fog", 1)
	} else {
		shader.SetUniformInt("fog", 0)
	}

	var fog int32 = 0
	if panels.DBG.FogDensity != 0 {
		fog = 1
	}
	shader.SetUniformInt("fog", fog)
	shader.SetUniformInt("fogDensity", panels.DBG.FogDensity)

	shader.SetUniformInt("width", int32(r.width))
	shader.SetUniformInt("height", int32(r.height))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformFloat("ambientFactor", panels.DBG.AmbientFactor)
	shader.SetUniformInt("shadowMap", 31)
	shader.SetUniformInt("depthCubeMap", 30)
	shader.SetUniformInt("cameraDepthMap", 29)

	shader.SetUniformFloat("near", panels.DBG.Near)
	shader.SetUniformFloat("far", panels.DBG.Far)
	shader.SetUniformFloat("bias", panels.DBG.PointLightBias)
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
		if entity == nil || entity.MeshComponent == nil {
			continue
		}

		modelMatrix := entities.WorldTransform(entity)
		shader.SetUniformUInt("entityID", uint32(entity.ID))

		var animationPlayer *animation.AnimationPlayer
		if entity.Animation != nil {
			animationPlayer = entity.Animation.AnimationPlayer
		}

		drawModel(
			viewerContext,
			lightContext,
			r.shadowMap,
			shader,
			r.world.AssetManager(),
			animationPlayer,
			modelMatrix,
			r.depthCubeMapTexture,
			entity.ID,
			entity.Material,
			r.world.ModelLibrary(),
			entity,
		)

	}
}

func (r *Renderer) renderImgui(renderContext RenderContext) {
	defer resetGLRenderSettings(r.renderFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.world.Platform().NewFrame()
	imgui.NewFrame()

	menuBarSize := menus.SetupMenuBar(r.world)

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

	panels.BuildTabsSet(
		r.world,
		renderContext,
		menuBarSize,
		r.world.Prefabs(),
	)

	imgui.PopStyleColorV(20)
	imgui.PopStyleVarV(7)

	if r.world.ShowImguiDemo() {
		imgui.ShowDemoWindow(nil)
	}

	imgui.Render()
	r.imguiRenderer.Render(r.world.Platform().DisplaySize(), r.world.Platform().FramebufferSize(), imgui.RenderedDrawData())
}

func (r *Renderer) renderGizmos(viewerContext ViewerContext, renderContext RenderContext) {
	if panels.SelectedEntity() == nil {
		return
	}

	gl.Clear(gl.DEPTH_BUFFER_BIT)

	entity := r.world.GetEntityByID(panels.SelectedEntity().ID)
	position := entity.WorldPosition()

	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		drawTranslationGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat"), position)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
		r.drawCircleGizmo(&viewerContext, position, renderContext)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
		drawScaleGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat"), position)
	}
}

func triangleVAOKey(triangle entities.Triangle) string {
	return fmt.Sprintf("%v_%v_%v", triangle.V1, triangle.V2, triangle.V3)

}
