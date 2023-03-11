package render

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/menus"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/lib"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

type World interface {
	AssetManager() *assets.AssetManager
	Camera() *camera.Camera
	Prefabs() []*prefabs.Prefab
	Entities() []*entities.Entity
	Lights() []*entities.Entity
	GetEntityByID(id int) *entities.Entity
	BuildRelation(parent *entities.Entity, child *entities.Entity)
	RemoveParent(child *entities.Entity)

	// for panels
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
	Platform() *input.SDLPlatform

	Serializer() *serialization.Serializer
	LoadWorld()
	SaveWorld()
}

const mipsCount int = 6

type Renderer struct {
	world         World
	shaderManager *shaders.ShaderManager

	shadowMap           *ShadowMap
	imguiRenderer       *ImguiOpenGL4Renderer
	depthCubeMapTexture uint32
	depthCubeMapFBO     uint32

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

	bloomFBO      uint32
	bloomVAO      uint32
	bloomVertices []float32
	bloomTextures []uint32

	upsampleFBO         uint32
	upsampleTextures    []uint32
	blendTargetTextures []uint32

	compositeFBO     uint32
	compositeTexture uint32

	widths  []int32
	heights []int32
}

func New(world World, shaderDirectory string, width, height int) *Renderer {
	r := &Renderer{world: world}
	r.shaderManager = shaders.NewShaderManager(shaderDirectory)

	imguiIO := imgui.CurrentIO()
	imguiRenderer, err := NewImguiOpenGL4Renderer(imguiIO)
	if err != nil {
		panic(err)
	}
	r.imguiRenderer = imguiRenderer

	var data int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &data)

	// note(kevin) using exactly the max texture size sometimes causes initialization to fail.
	// so, I cap it at a fraction of the max
	settings.RuntimeMaxTextureSize = int(float32(data) * .90)

	shadowMap, err := NewShadowMap(settings.RuntimeMaxTextureSize, settings.RuntimeMaxTextureSize, settings.Far)
	if err != nil {
		panic(fmt.Sprintf("failed to create shadow map %s", err))
	}
	r.shadowMap = shadowMap
	r.depthCubeMapFBO, r.depthCubeMapTexture = lib.InitDepthCubeMap()

	r.redCircleFB, r.redCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA)
	r.greenCircleFB, r.greenCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA)
	r.blueCircleFB, r.blueCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA)
	r.yellowCircleFB, r.yellowCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024, gl.RGBA)

	compileShaders(r.shaderManager)

	renderFBO, colorTextures := r.initFrameBuffer2(width, height, gl.R11F_G11F_B10F, 2)
	r.renderFBO = renderFBO
	r.mainColorTexture = colorTextures[0]
	r.colorPickingTexture = colorTextures[1]
	r.colorPickingAttachment = gl.COLOR_ATTACHMENT1

	r.initBloom(1920, 1080)
	r.initComposite(width, height)
	return r
}

func (r *Renderer) Render(delta time.Duration, renderContext RenderContext) {
	initOpenGLRenderSettings()

	// configure camera viewer context
	position := r.world.Camera().Position
	orientation := r.world.Camera().Orientation

	viewerViewMatrix := orientation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := ViewerContext{
		Position:    position,
		Orientation: orientation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(renderContext.FovY()), renderContext.AspectRatio(), settings.Near, settings.Far),
	}

	// configure light viewer context
	modelSpaceFrustumPoints := CalculateFrustumPoints(position, orientation, settings.Near, settings.Far, renderContext.FovX(), renderContext.FovY(), renderContext.AspectRatio(), settings.ShadowMapDistanceFactor)

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
		directionalLight.LightInfo.Direction[0] = float64(panels.DBG.DirectionalLightDir[0])
		directionalLight.LightInfo.Direction[1] = float64(panels.DBG.DirectionalLightDir[1])
		directionalLight.LightInfo.Direction[2] = float64(panels.DBG.DirectionalLightDir[2])

		directionalLightX = directionalLight.LightInfo.Direction.X()
		directionalLightY = directionalLight.LightInfo.Direction.Y()
		directionalLightZ = directionalLight.LightInfo.Direction.Z()
	}

	lightOrientation := utils.Vec3ToQuat(mgl64.Vec3{directionalLightX, directionalLightY, directionalLightZ})
	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightOrientation.Mat4(), modelSpaceFrustumPoints, settings.ShadowmapZOffset)
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

	r.clearMainFrameBuffer()

	r.renderSkybox(renderContext)
	r.renderToSquareDepthMap(lightViewerContext, lightContext)
	r.renderToCubeDepthMap(lightContext)
	r.renderScene(cameraViewerContext, lightContext, renderContext)

	var finalRenderTexture uint32
	if panels.DBG.Bloom {
		r.downSample(r.mainColorTexture)
		upsampleTexture := r.upSample()
		panels.DBG.DebugTexture = upsampleTexture
		r.composite(renderContext, r.mainColorTexture, upsampleTexture)

		finalRenderTexture = r.compositeTexture
	} else {
		finalRenderTexture = r.mainColorTexture
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawTexturedQuad(&cameraViewerContext, r.shaderManager, finalRenderTexture, 1, float32(renderContext.aspectRatio), nil, false)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.renderGizmos(cameraViewerContext, renderContext)
	r.renderImgui(renderContext)
}

func (r *Renderer) renderToSquareDepthMap(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings(r.renderFBO)
	r.shadowMap.Prepare()

	if !panels.DBG.EnableShadowMapping {
		// set the depth to be max value to prevent shadow mapping
		gl.ClearDepth(1)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		return
	}

	shaderManager := r.shaderManager
	for _, entity := range r.world.Entities() {
		modelMatrix := entities.ComputeTransformMatrix(entity)

		if entity.Prefab != nil {
			shader := shaderManager.GetShaderProgram("modelpbr")
			shader.Use()

			if entity.AnimationPlayer != nil && entity.AnimationPlayer.CurrentAnimation() != "" {
				shader.SetUniformInt("isAnimated", 1)
				animationTransforms := entity.AnimationPlayer.AnimationTransforms()
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

			model := entity.Prefab.ModelRefs[0].Model
			m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix)

			shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
			shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
			shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
			shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
			shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))

			for _, renderData := range model.RenderData() {
				ctx := model.CollectionContext()
				mesh := model.Collection().Meshes[renderData.MeshID]
				shader.SetUniformMat4("model", m32ModelMatrix.Mul4(renderData.Transform))

				gl.BindVertexArray(ctx.VAOS[renderData.MeshID])
				gl.DrawElements(gl.TRIANGLES, int32(len(mesh.Vertices)), gl.UNSIGNED_INT, nil)
			}
		}
	}
}

func (r *Renderer) renderToCubeDepthMap(lightContext LightContext) {
	defer resetGLRenderSettings(r.renderFBO)

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

	for _, entity := range r.world.Entities() {
		if entity.Prefab != nil {
			modelMatrix := entities.ComputeTransformMatrix(entity)
			model := entity.Prefab.ModelRefs[0].Model
			m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix)

			if entity.AnimationPlayer != nil && entity.AnimationPlayer.CurrentAnimation() != "" {
				shader.SetUniformInt("isAnimated", 1)
				animationTransforms := entity.AnimationPlayer.AnimationTransforms()
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

			for _, renderData := range model.RenderData() {
				ctx := model.CollectionContext()
				mesh := model.Collection().Meshes[renderData.MeshID]
				shader.SetUniformMat4("model", m32ModelMatrix.Mul4(renderData.Transform))
				// shader.SetUniformMat4("model", renderData.Transform)

				gl.BindVertexArray(ctx.VAOS[renderData.MeshID])
				gl.DrawElements(gl.TRIANGLES, int32(len(mesh.Vertices)), gl.UNSIGNED_INT, nil)
			}
		}
	}
}

// renderScene renders a scene from the perspective of a viewer
func (r *Renderer) renderScene(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext) {
	defer resetGLRenderSettings(r.renderFBO)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)

	shaderManager := r.shaderManager

	for _, entity := range r.world.Entities() {
		modelMatrix := entities.ComputeTransformMatrix(entity)

		if entity.Prefab != nil {
			shaderName := "modelpbr"
			shader := shaderManager.GetShaderProgram(shaderName)
			shader.Use()
			if entity.AnimationPlayer != nil && entity.AnimationPlayer.CurrentAnimation() != "" {
				shader.SetUniformInt("isAnimated", 1)
			} else {
				shader.SetUniformInt("isAnimated", 0)
			}

			drawModel(
				viewerContext,
				lightContext,
				r.shadowMap,
				shader,
				r.world.AssetManager(),
				entity.Prefab.ModelRefs[0].Model,
				entity.AnimationPlayer,
				modelMatrix,
				r.depthCubeMapTexture,
				entity.ID,
			)

			// joint rendering for the selected entity
			selectedEntity := panels.SelectedEntity()
			if selectedEntity != nil && selectedEntity.ID == entity.ID {
				// TODO: optimize this - can probably cache some of these computations
				if len(panels.JointsToRender) > 0 && entity.AnimationPlayer != nil && entity.AnimationPlayer.CurrentAnimation() != "" {
					jointShader := shaderManager.GetShaderProgram("flat")
					color := mgl64.Vec3{0 / 255, 255.0 / 255, 85.0 / 255}

					var jointLines [][]mgl64.Vec3
					model := r.world.AssetManager().GetModel(entity.Prefab.Name)
					animationTransforms := entity.AnimationPlayer.AnimationTransforms()

					for _, jid := range panels.JointsToRender {
						jointTransform := animationTransforms[jid]
						lines := cubeLines(15)
						jt := utils.Mat4F32ToF64(jointTransform)
						for _, line := range lines {
							points := line
							for i := 0; i < len(points); i++ {
								bindTransform := model.JointMap[jid].FullBindTransform
								points[i] = jt.Mul4(utils.Mat4F32ToF64(bindTransform)).Mul4x1(points[i].Vec4(1)).Vec3()
							}
						}
						jointLines = append(jointLines, lines...)
					}

					for _, line := range jointLines {
						points := line
						for i := 0; i < len(points); i++ {
							points[i] = modelMatrix.Mul4x1(points[i].Vec4(1)).Vec3()
						}
					}

					drawLines(viewerContext, jointShader, jointLines, 0.5, color)
				}
			}
		}

		if len(entity.ShapeData) > 0 {
			shader := shaderManager.GetShaderProgram("flat")
			color := mgl64.Vec3{0 / 255, 255.0 / 255, 85.0 / 255}

			for _, shapeData := range entity.ShapeData {
				lines := cubeLines(shapeData.Cube.Length)
				for _, line := range lines {
					points := line
					for i := 0; i < len(points); i++ {
						points[i] = modelMatrix.Mul4x1(points[i].Vec4(1)).Vec3()
					}
				}

				drawLines(viewerContext, shader, lines, 0.5, color)
			}
		}

		if entity.ImageInfo != nil {
			texture := r.world.AssetManager().GetTexture("light")
			if texture != nil {
				a := mgl64.Vec4{0, 1, 0, 1}
				b := mgl64.Vec4{1, 0, 0, 1}
				cameraUp := viewerContext.Orientation.Mat4().Mul4x1(a).Vec3()
				cameraRight := viewerContext.Orientation.Mat4().Mul4x1(b).Vec3()

				if entity.Billboard != nil {
					shader := shaderManager.GetShaderProgram("basic_quad_world")
					shader.Use()
					shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
					shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
					shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
					shader.SetUniformVec3("pickingColor", idToPickingColor(entity.ID))
					shader.SetUniformInt("doColorOverride", 1)
					shader.SetUniformVec3("colorOverride", mgl32.Vec3{panels.DBG.LightColorOverride, panels.DBG.LightColorOverride, panels.DBG.LightColorOverride})

					drawBillboardTexture(texture.ID, cameraUp, cameraRight)
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

				dir := lightInfo.Direction.Normalize().Mul(50)
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
				drawTexturedQuad(&viewerContext, r.shaderManager, texture, 1, float32(renderContext.AspectRatio()), &particleModelMatrix, true)
			}
		}

		collider := entity.Collider
		if collider != nil {
			translation := mgl64.Translate3D(entity.LocalPosition.X(), entity.LocalPosition.Y(), entity.LocalPosition.Z())
			// lots of hacky rendering stuff to get the rectangle to billboard
			center := mgl64.Vec3{entity.LocalPosition.X(), 0, entity.LocalPosition.Z()}
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

func (r *Renderer) renderImgui(renderContext RenderContext) {
	defer resetGLRenderSettings(r.renderFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.world.Platform().NewFrame()
	imgui.NewFrame()

	menuBarSize := menus.SetupMenuBar(r.world)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{})
	imgui.PushStyleVarVec2(imgui.StyleVarItemInnerSpacing, imgui.Vec2{})
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{})
	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .65, Y: .79, Z: 0.30, W: 1})

	panels.BuildExplorer(r.world.Entities(), r.world, menuBarSize, renderContext)
	panels.BuildPrefabs(r.world.Prefabs(), r.world, renderContext)
	panels.BuildDebug(
		r.world,
		renderContext,
	)

	imgui.PopStyleColor()
	imgui.PopStyleVarV(10)
	var open bool
	imgui.ShowDemoWindow(&open)

	e := panels.SelectedEntity()
	if e != nil {
		if e.AnimationPlayer != nil {
			panels.BuildAnimation(r.world, e)
		}
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

func (r *Renderer) initBloom(maxWidth, maxHeight int) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var upsampleFBO uint32
	gl.GenFramebuffers(1, &upsampleFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, upsampleFBO)

	width, height := maxWidth, maxHeight
	for i := 0; i < mipsCount; i++ {
		// upsampling textures start one doubling earlier than the downsampling textures
		// the first step of upsampling is taking the lowest downsampling mip and upsampling it
		var upsampleTexture uint32
		gl.GenTextures(1, &upsampleTexture)
		gl.BindTexture(gl.TEXTURE_2D, upsampleTexture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		r.upsampleTextures = append(r.upsampleTextures, upsampleTexture)

		var blendTargetTexture uint32
		gl.GenTextures(1, &blendTargetTexture)
		gl.BindTexture(gl.TEXTURE_2D, blendTargetTexture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		r.blendTargetTextures = append(r.blendTargetTextures, blendTargetTexture)

		width /= 2
		height /= 2

		r.widths = append(r.widths, int32(width))
		r.heights = append(r.heights, int32(height))

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		r.bloomTextures = append(r.bloomTextures, texture)
	}

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.bloomTextures[0], 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, upsampleFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.upsampleTextures[0], 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	r.bloomFBO = fbo
	r.upsampleFBO = upsampleFBO

	vertices := []float32{
		-1, -1, 0.0, 0.0,
		1, -1, 1.0, 0.0,
		1, 1, 1.0, 1.0,
		1, 1, 1.0, 1.0,
		-1, 1, 0.0, 1.0,
		-1, -1, 0.0, 0.0,
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)
	r.bloomVAO = vao
	r.bloomVertices = vertices
}

func (r *Renderer) downSample(srcTexture uint32) {
	defer resetGLRenderSettings(r.renderFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.bloomFBO)

	shader := r.shaderManager.GetShaderProgram("bloom_downsample")
	shader.Use()

	for i := 0; i < len(r.bloomTextures); i++ {
		width := r.widths[i]
		height := r.heights[i]

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, srcTexture)
		gl.Viewport(0, 0, width, height)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.bloomTextures[i], 0)

		gl.BindVertexArray(r.bloomVAO)
		if i == 0 {
			shader.SetUniformInt("karis", 1)
		} else {
			shader.SetUniformInt("karis", 0)
		}
		if i < int(panels.DBG.BloomThresholdPasses) {
			shader.SetUniformInt("bloomThresholdEnabled", 1)
		} else {
			shader.SetUniformInt("bloomThresholdEnabled", 0)
		}
		shader.SetUniformFloat("bloomThreshold", panels.DBG.BloomThreshold)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		srcTexture = r.bloomTextures[i]
	}
}

// double check that the upsampling works and blends the right textures
// welp, i need to be ping ponging GG
func (r *Renderer) upSample() uint32 {
	defer resetGLRenderSettings(r.renderFBO)

	mipsCount := len(r.bloomTextures)

	var upSampleSource uint32
	upSampleSource = r.bloomTextures[mipsCount-1]
	var i int
	for i = mipsCount - 1; i > 0; i-- {
		blendTargetMip := r.blendTargetTextures[i]
		upSampleMip := r.upsampleTextures[i]

		gl.BindFramebuffer(gl.FRAMEBUFFER, r.upsampleFBO)

		shader := r.shaderManager.GetShaderProgram("bloom_upsample")
		shader.Use()
		shader.SetUniformFloat("upsamplingRadius", panels.DBG.BloomUpsamplingRadius)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, upSampleSource)

		gl.Viewport(0, 0, r.widths[i-1], r.heights[i-1])
		drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, upSampleMip, 0)
		gl.DrawBuffers(1, &drawBuffers[0])

		gl.BindVertexArray(r.bloomVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		// r.blend(r.widths[i-1], r.heights[i-1], upSampleMip, r.bloomTextures[i-1], blendTargetMip)
		r.blend(r.widths[i-1], r.heights[i-1], r.bloomTextures[i-1], upSampleMip, blendTargetMip)
		upSampleSource = blendTargetMip
	}

	return upSampleSource
}

func (r *Renderer) blend(width, height int32, texture0, texture1, target uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.compositeFBO)

	shader := r.shaderManager.GetShaderProgram("blend")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("texture0", 0)
	shader.SetUniformInt("texture1", 1)

	gl.Viewport(0, 0, width, height)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, target, 0)

	gl.BindVertexArray(r.bloomVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func (r *Renderer) initComposite(width, height int) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
		int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	r.compositeFBO, r.compositeTexture = fbo, texture
}

func (r *Renderer) composite(renderContext RenderContext, texture0, texture1 uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.compositeFBO)

	shader := r.shaderManager.GetShaderProgram("composite")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("scene", 0)
	shader.SetUniformInt("bloomBlur", 1)
	shader.SetUniformFloat("exposure", panels.DBG.Exposure)
	shader.SetUniformFloat("bloomIntensity", panels.DBG.BloomIntensity)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.compositeTexture, 0)

	gl.BindVertexArray(r.bloomVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
