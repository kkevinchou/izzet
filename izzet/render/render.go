package render

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
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
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	defaultTexture string = "color_grid"

	// shadow map properties
	shadowmapZOffset     float64 = 400
	fovx                 float64 = 105
	Near                 float64 = 1
	far                  float64 = 3000
	shadowDistanceFactor float64 = .4 // proportion of view fustrum to include in shadow cuboid
)

type World interface {
	AssetManager() *assets.AssetManager
	Camera() *camera.Camera
	Prefabs() []*prefabs.Prefab
	Entities() []*entities.Entity
	GetEntityByID(id int) *entities.Entity

	// for panels
	AddEntity(entity *entities.Entity)
	GetPrefabByID(id int) *prefabs.Prefab
	Window() *sdl.Window
	Platform() *input.SDLPlatform

	Serializer() *serialization.Serializer
	LoadWorld()
	SaveWorld()
}

type Renderer struct {
	world         World
	shaderManager *shaders.ShaderManager

	// render properties
	fovY        float64
	aspectRatio float64
	// shaderManager *shaders.ShaderManager
	shadowMap     *ShadowMap
	imguiRenderer *ImguiOpenGL4Renderer

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
}

func New(world World, shaderDirectory string) *Renderer {
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

	shadowMap, err := NewShadowMap(settings.RuntimeMaxTextureSize, settings.RuntimeMaxTextureSize, far*shadowDistanceFactor)
	if err != nil {
		panic(fmt.Sprintf("failed to create shadow map %s", err))
	}
	r.shadowMap = shadowMap

	w, h := r.world.Window().GetSize()
	r.colorPickingFB, r.colorPickingTexture = r.initFrameBuffer(int(w), int(h))
	r.redCircleFB, r.redCircleTexture = r.initFrameBuffer(1024, 1024)
	r.greenCircleFB, r.greenCircleTexture = r.initFrameBuffer(1024, 1024)
	r.blueCircleFB, r.blueCircleTexture = r.initFrameBuffer(1024, 1024)
	r.yellowCircleFB, r.yellowCircleTexture = r.initFrameBuffer(1024, 1024)

	compileShaders(r.shaderManager)

	r.aspectRatio = float64(settings.Width) / float64(settings.Height)
	r.fovY = mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/r.aspectRatio))

	return r
}

func (r *Renderer) Render(delta time.Duration) {
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
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(r.fovY), r.aspectRatio, Near, far),
	}

	// configure light viewer context
	modelSpaceFrustumPoints := CalculateFrustumPoints(position, orientation, Near, far, r.fovY, r.aspectRatio, shadowDistanceFactor)

	lightOrientation := utils.Vec3ToQuat(mgl64.Vec3{-1, -1, -1})
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
	_ = lightContext
	_ = lightViewerContext

	r.viewerContext = cameraViewerContext

	r.clearMainFrameBuffer()
	r.renderSkybox()
	r.renderToDepthMap(lightViewerContext, lightContext)
	r.renderColorPicking(cameraViewerContext)
	r.renderToDisplay(cameraViewerContext, lightContext)

	r.renderGizmos(cameraViewerContext)

	r.RenderImgui()
}

func (r *Renderer) RenderImgui() {
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

	// var open1 bool
	// imgui.SetNextWindowBgAlpha(0)
	// imgui.SetNextWindowPosV(imgui.Vec2{}, imgui.ConditionAlways, imgui.Vec2{})
	// imgui.SetNextWindowSizeV(imgui.Vec2{X: float32(settings.Width), Y: float32(settings.Height)}, imgui.ConditionAlways)
	// imgui.BeginV("explorer root", &open1, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize|imgui.WindowFlagsMenuBar)
	// imgui.MenuItem("test")

	panels.BuildExplorer(r.world.Entities(), r.world, menuBarSize)
	panels.BuildPrefabs(r.world.Prefabs(), r.world)

	// imgui.End()

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

func (r *Renderer) renderCircle() {
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	shaderManager := r.shaderManager
	var alpha float64 = 1

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.redCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{1, 0, 0, alpha})

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.greenCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{0, 1, 0, alpha})

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.blueCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{0, 0, 1, alpha})

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.yellowCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{1, 1, 0, alpha})
}

func calcPosition(entity *entities.Entity) mgl64.Vec3 {
	var position mgl64.Vec3
	for entity != nil {
		position = position.Add(entity.Position)
		entity = entity.Parent
	}
	return position
}

func (r *Renderer) renderGizmos(viewerContext ViewerContext) {
	if panels.SelectedEntity() == nil {
		return
	}

	gl.Clear(gl.DEPTH_BUFFER_BIT)

	entity := r.world.GetEntityByID(panels.SelectedEntity().ID)
	position := calcPosition(entity)

	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		drawTranslationGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat"), position)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
		r.drawCircleGizmo(&viewerContext, position)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
		drawScaleGizmo(&viewerContext, r.shaderManager.GetShaderProgram("flat"), position)
	}
}

func (r *Renderer) renderToDisplay(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()
	w, h := r.world.Window().GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	r.renderScene(viewerContext, lightContext, false)
}

func (r *Renderer) renderToDepthMap(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()
	r.shadowMap.Prepare()

	r.renderScene(viewerContext, lightContext, true)
}

func computeModelMatrix(entity *entities.Entity) mgl64.Mat4 {
	modelMatrix := createModelMatrix(
		mgl64.Scale3D(entity.Scale.X(), entity.Scale.Y(), entity.Scale.Z()),
		entity.Rotation.Mat4(),
		mgl64.Translate3D(entity.Position[0], entity.Position[1], entity.Position[2]),
	)
	if entity.Parent != nil {
		parentModelMatrix := computeModelMatrix(entity.Parent)
		modelMatrix = parentModelMatrix.Mul4(modelMatrix)
	}
	return modelMatrix
}

// renderScene renders a scene from the perspective of a viewer
func (r *Renderer) renderScene(viewerContext ViewerContext, lightContext LightContext, shadowPass bool) {
	shaderManager := r.shaderManager

	for _, entity := range r.world.Entities() {
		modelMatrix := computeModelMatrix(entity)

		shader := "model_static"
		if entity.AnimationPlayer != nil && entity.AnimationPlayer.CurrentAnimation() != "" {
			shader = "modelpbr"
		}

		// if native object, do special render call
		// else:
		drawModel(
			viewerContext,
			lightContext,
			r.shadowMap,
			shaderManager.GetShaderProgram(shader),
			r.world.AssetManager(),
			entity.Prefab.ModelRefs[0].Model,
			entity.AnimationPlayer,
			modelMatrix,
		)
	}
}

func createModelMatrix(scaleMatrix, rotationMatrix, translationMatrix mgl64.Mat4) mgl64.Mat4 {
	return translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
}

func drawTranslationGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
	colors := []mgl64.Vec3{mgl64.Vec3{1, 0, 0}, mgl64.Vec3{0, 0, 1}, mgl64.Vec3{0, 1, 0}}

	for i, axis := range gizmo.T.Axes {
		lines := [][]mgl64.Vec3{
			[]mgl64.Vec3{position, position.Add(axis)},
		}
		color := colors[i]
		if i == gizmo.T.HoverIndex {
			color = mgl64.Vec3{1, 1, 0}
		}
		drawLines(*viewerContext, shader, lines, 1, color)
	}
}

func drawScaleGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
	colors := []mgl64.Vec3{mgl64.Vec3{1, 0, 0}, mgl64.Vec3{0, 0, 1}, mgl64.Vec3{0, 1, 0}}

	for i, axis := range gizmo.S.Axes {
		lines := [][]mgl64.Vec3{
			[]mgl64.Vec3{position, position.Add(axis)},
		}
		color := colors[i]
		if i == gizmo.S.HoverIndex {
			color = mgl64.Vec3{1, 1, 0}
		}
		drawLines(*viewerContext, shader, lines, 1, color)
	}
}
func (r *Renderer) drawCircleGizmo(cameraViewerContext *ViewerContext, position mgl64.Vec3) {
	defer resetGLRenderSettings()
	w, h := r.world.Window().GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))

	t := mgl32.Translate3D(float32(position[0]), float32(position[1]), float32(position[2]))
	s := mgl32.Scale3D(25, 25, 25)

	rotations := []mgl32.Mat4{
		mgl32.Ident4(),
		mgl32.HomogRotate3DY(90 * math.Pi / 180),
		mgl32.HomogRotate3DX(-90 * math.Pi / 180),
	}

	textures := []uint32{r.redCircleTexture, r.greenCircleTexture, r.blueCircleTexture}

	r.renderCircle()
	for i := 0; i < 3; i++ {
		modelMatrix := t.Mul4(rotations[i]).Mul4(s)
		texture := textures[i]
		if i == gizmo.R.HoverIndex {
			texture = r.yellowCircleTexture
		}
		drawTexturedQuad(cameraViewerContext, r.shaderManager, texture, 1, float32(r.aspectRatio), &modelMatrix, true)
	}
}

func (r *Renderer) initFrameBuffer(width int, height int) (uint32, uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, int32(width), int32(height))
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)

	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, rbo)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, texture
}

func (r *Renderer) renderColorPicking(viewerContext ViewerContext) {
	defer resetGLRenderSettings()
	w, h := r.world.Window().GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.colorPickingFB)
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	shaderManager := r.shaderManager

	for _, entity := range r.world.Entities() {
		modelMatrix := computeModelMatrix(entity)

		shader := "color_picking"
		// TODO: color picking shader for animated entities?

		drawWIthID(
			viewerContext,
			shaderManager.GetShaderProgram(shader),
			r.world.AssetManager(),
			entity.Prefab.ModelRefs[0].Model,
			entity.AnimationPlayer,
			modelMatrix,
			entity.ID,
		)
	}
}

func (r *Renderer) clearMainFrameBuffer() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderSkybox() {
	defer resetGLRenderSettings()
	w, h := r.world.Window().GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))

	drawWithNDC(r.shaderManager)

	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	shaderManager := r.shaderManager
	shaderManager.GetShaderProgram("skybox").Use()

}

func (r *Renderer) ViewerContext() ViewerContext {
	return r.viewerContext
}

func (r *Renderer) handleResize() {
	w, h := r.world.Window().GetSize()
	r.aspectRatio = float64(w) / float64(h)
	r.fovY = mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/r.aspectRatio))
}

func (r *Renderer) GetEntityByPixelPosition(pixelPosition mgl64.Vec2) *int {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.colorPickingFB)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	data := make([]byte, 4)
	_, h := r.world.Window().GetSize()
	gl.ReadPixels(int32(pixelPosition[0]), int32(h)-int32(pixelPosition[1]), 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))

	// discard the alpha channel data
	data[3] = 0

	// NOTE(kevin) actually not sure why, but this works
	// i would've expected to need to multiply by 255, but apparently it's handled somehow
	uintID := binary.LittleEndian.Uint32(data)
	if uintID == settings.EmptyColorPickingID {
		return nil
	}

	id := int(uintID)
	return &id
}
