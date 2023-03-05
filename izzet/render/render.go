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

type Renderer struct {
	world         World
	shaderManager *shaders.ShaderManager

	shadowMap           *ShadowMap
	imguiRenderer       *ImguiOpenGL4Renderer
	depthCubeMapTexture uint32
	depthCubeMapFBO     uint32

	colorPickingFB      uint32
	colorPickingTexture uint32

	redCircleFB         uint32
	redCircleTexture    uint32
	greenCircleFB       uint32
	greenCircleTexture  uint32
	blueCircleFB        uint32
	blueCircleTexture   uint32
	yellowCircleFB      uint32
	yellowCircleTexture uint32

	viewerContext ViewerContext

	drawFBO      uint32
	drawTexture0 uint32
	drawTexture1 uint32
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

	// r.newPickingTexture = r.createColorPickingAttachment(width, height)
	r.colorPickingFB, r.colorPickingTexture = r.initFrameBufferSingleColorAttachment(width, height)
	r.redCircleFB, r.redCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024)
	r.greenCircleFB, r.greenCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024)
	r.blueCircleFB, r.blueCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024)
	r.yellowCircleFB, r.yellowCircleTexture = r.initFrameBufferSingleColorAttachment(1024, 1024)

	compileShaders(r.shaderManager)

	drawFBO, colorTextures := r.initFrameBuffer(width, height, 2)
	r.drawFBO = drawFBO
	r.drawTexture0 = colorTextures[0]
	r.drawTexture1 = colorTextures[1]

	fmt.Println(colorTextures)

	panels.DBG.DebugTexture = r.drawTexture0
	return r
}

func (r *Renderer) createColorPickingAttachment(width, height int) uint32 {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, texture, 0)

	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, int32(width), int32(height))
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)

	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, rbo)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return texture
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
		directionalLight.LightInfo.Direction[0] = float64(panels.DBG.DirectionalLightX)
		directionalLight.LightInfo.Direction[1] = float64(panels.DBG.DirectionalLightY)
		directionalLight.LightInfo.Direction[2] = float64(panels.DBG.DirectionalLightZ)

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
	r.renderToColorPickingBuffer(cameraViewerContext, renderContext)
	r.renderToDisplay(cameraViewerContext, lightContext, renderContext)
	r.renderToDisplay2(cameraViewerContext, lightContext, renderContext)

	r.renderGizmos(cameraViewerContext, renderContext)
	r.renderImgui(renderContext)
}

func (r *Renderer) renderToSquareDepthMap(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()
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
	defer resetGLRenderSettings()

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

func (r *Renderer) renderToColorPickingBuffer(viewerContext ViewerContext, renderContext RenderContext) {
	defer resetGLRenderSettings()
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.colorPickingFB)
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	shaderManager := r.shaderManager

	for _, entity := range r.world.Entities() {
		modelMatrix := entities.ComputeTransformMatrix(entity)

		if entity.Prefab != nil {
			shader := "color_picking"
			// TODO: color picking shader for animated entities?

			drawModelWIthID(
				viewerContext,
				shaderManager.GetShaderProgram(shader),
				r.world.AssetManager(),
				entity.Prefab.ModelRefs[0].Model,
				entity.AnimationPlayer,
				modelMatrix,
				entity.ID,
			)
		}

		if entity.ImageInfo != nil {
			texture := r.world.AssetManager().GetTexture("light")
			if texture != nil {
				a := mgl64.Vec4{0, 1, 0, 1}
				b := mgl64.Vec4{1, 0, 0, 1}
				cameraUp := viewerContext.Orientation.Mat4().Mul4x1(a).Vec3()
				cameraRight := viewerContext.Orientation.Mat4().Mul4x1(b).Vec3()

				if entity.Billboard != nil {
					shader := shaderManager.GetShaderProgram("color_picking")
					shader.Use()
					shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
					shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
					shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
					shader.SetUniformVec3("pickingColor", idToPickingColor(entity.ID))

					drawBillboardTexture(texture.ID, cameraUp, cameraRight)
				}
			} else {
				fmt.Println("couldn't find texture", "light")
			}
		}

		if len(entity.ShapeData) > 0 {
			// color picking only supports cubes atm
			for _, shapeData := range entity.ShapeData {
				cube := shapeData.Cube
				if cube == nil {
					continue
				}

				points := cubePoints(cube.Length)

				shader := shaderManager.GetShaderProgram("color_picking")
				shader.Use()
				shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
				shader.SetUniformFloat("alpha", float32(1))
				shader.SetUniformVec3("pickingColor", idToPickingColor(entity.ID))

				drawTris(
					viewerContext,
					points,
					mgl64.Vec3{1, 0, 0},
				)
			}
		}
	}
}

func (r *Renderer) renderImgui(renderContext RenderContext) {
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

func (r *Renderer) renderToDisplay(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext) {
	defer resetGLRenderSettings()

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.renderScene(viewerContext, lightContext, renderContext)
}

func (r *Renderer) renderToDisplay2(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext) {
	defer resetGLRenderSettings()

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.drawFBO)
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.ClearColor(1, 1, 1, 1)
	r.renderScene(viewerContext, lightContext, renderContext)
}
