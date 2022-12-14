package izzet

import (
	"errors"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
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

func (g *Izzet) Render(delta time.Duration) {
	initOpenGLRenderSettings()

	// configure camera viewer context
	position := g.camera.Position
	orientation := g.camera.Orientation

	viewerViewMatrix := orientation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := ViewerContext{
		Position:    position,
		Orientation: orientation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(g.fovY), g.aspectRatio, Near, far),
	}

	// configure light viewer context
	modelSpaceFrustumPoints := CalculateFrustumPoints(position, orientation, Near, far, g.fovY, g.aspectRatio, shadowDistanceFactor)

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

	g.viewerContext = cameraViewerContext

	g.renderToDepthMap(lightViewerContext, lightContext)
	g.renderColorPicking(cameraViewerContext)
	g.renderToDisplay(cameraViewerContext, lightContext)
	g.renderCircleGizmo(&cameraViewerContext)

	g.renderGizmos(cameraViewerContext)

	g.renderImgui()
	g.window.GLSwap()
}

func (g *Izzet) renderCircleGizmo(cameraViewerContext *ViewerContext) {
	defer resetGLRenderSettings()
	r := mgl32.HomogRotate3DY(90 * math.Pi / 180)
	t := mgl32.Translate3D(0, 300, 0)
	s := mgl32.Scale3D(50, 50, 50)

	r1 := mgl32.HomogRotate3DX(-90 * math.Pi / 180)
	t1 := mgl32.Translate3D(0, 300, 0)
	s1 := mgl32.Scale3D(50, 50, 50)

	// probably only need to run this once?
	g.renderCircle()
	modelMatrix := mgl32.Translate3D(0, 300, 0).Mul4(mgl32.Scale3D(50, 50, 50))
	drawTexturedQuad(cameraViewerContext, g.shaderManager, g.redCircleTexture, 0.5, float32(g.aspectRatio), &modelMatrix, true)
	modelMatrix = t.Mul4(r).Mul4(s)
	drawTexturedQuad(cameraViewerContext, g.shaderManager, g.greenCircleTexture, 0.5, float32(g.aspectRatio), &modelMatrix, true)
	modelMatrix = t1.Mul4(r1).Mul4(s1)
	drawTexturedQuad(cameraViewerContext, g.shaderManager, g.blueCircleTexture, 0.5, float32(g.aspectRatio), &modelMatrix, true)
}

func (g *Izzet) renderImgui() {
	g.platform.NewFrame()
	imgui.NewFrame()

	imgui.BeginMainMenuBar()
	menuBarSize := imgui.WindowSize()
	if imgui.BeginMenu("File") {
		if imgui.MenuItem("New") {
		}

		if imgui.MenuItem("Open") {
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()

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

	panels.BuildExplorer(g.Entities(), g, menuBarSize)
	panels.BuildPrefabs(g.Prefabs(), g)

	// imgui.End()

	imgui.PopStyleColor()
	imgui.PopStyleVarV(10)
	var open bool
	imgui.ShowDemoWindow(&open)

	imgui.Render()
	g.imguiRenderer.Render(g.platform.DisplaySize(), g.platform.FramebufferSize(), imgui.RenderedDrawData())
}

func (g *Izzet) renderCircle() {
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	shaderManager := g.shaderManager
	var alpha float64 = 1

	gl.BindFramebuffer(gl.FRAMEBUFFER, g.redCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{1, 0, 0, alpha})

	gl.BindFramebuffer(gl.FRAMEBUFFER, g.greenCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{0, 1, 0, alpha})

	gl.BindFramebuffer(gl.FRAMEBUFFER, g.blueCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	drawCircle(shaderManager.GetShaderProgram("unit_circle"), mgl64.Vec4{0, 0, 1, alpha})
}

func (g *Izzet) renderGizmos(viewerContext ViewerContext) {
	if panels.SelectedEntity == nil {
		return
	}

	gl.Clear(gl.DEPTH_BUFFER_BIT)
	entity := g.entities[panels.SelectedEntity.ID]
	drawGizmo(&viewerContext, g.shaderManager.GetShaderProgram("flat"), entity.Position)
}

func (g *Izzet) renderToDisplay(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()
	w, h := g.window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	g.renderScene(viewerContext, lightContext, false)
}

func (g *Izzet) renderToDepthMap(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()
	g.shadowMap.Prepare()

	g.renderScene(viewerContext, lightContext, true)
}

// renderScene renders a scene from the perspective of a viewer
func (g *Izzet) renderScene(viewerContext ViewerContext, lightContext LightContext, shadowPass bool) {
	shaderManager := g.shaderManager

	for _, entity := range g.Entities() {
		modelMatrix := createModelMatrix(
			mgl64.Scale3D(1, 1, 1),
			mgl64.QuatIdent().Mat4(),
			mgl64.Translate3D(entity.Position[0], entity.Position[1], entity.Position[2]),
		)

		shader := "model_static"
		if entity.AnimationPlayer != nil {
			shader = "modelpbr"
		}

		drawModel(
			viewerContext,
			lightContext,
			g.shadowMap,
			shaderManager.GetShaderProgram(shader),
			g.assetManager,
			entity.Prefab.ModelRefs[0].Model,
			entity.AnimationPlayer,
			modelMatrix,
		)
	}

}

func createModelMatrix(scaleMatrix, rotationMatrix, translationMatrix mgl64.Mat4) mgl64.Mat4 {
	return translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
}

func drawGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
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

func (g *Izzet) initFrameBuffer(width int, height int) (uint32, uint32) {
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

func (g *Izzet) renderColorPicking(viewerContext ViewerContext) {
	defer resetGLRenderSettings()
	gl.BindFramebuffer(gl.FRAMEBUFFER, g.colorPickingFB)
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// modelMatrix := mgl32.Translate3D(0, 300, 0).Mul4(mgl32.Scale3D(50, 50, 50))
	// drawTexturedQuad(&viewerContext, g.shaderManager, g.tmpTexture, 0.5, float32(g.aspectRatio), &modelMatrix)

	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	shaderManager := g.shaderManager

	for _, entity := range g.Entities() {
		modelMatrix := createModelMatrix(
			mgl64.Scale3D(1, 1, 1),
			mgl64.QuatIdent().Mat4(),
			mgl64.Translate3D(entity.Position[0], entity.Position[1], entity.Position[2]),
		)

		shader := "color_picking"
		// if entity.AnimationPlayer != nil {
		// 	shader = "modelpbr"
		// }

		drawWIthID(
			viewerContext,
			shaderManager.GetShaderProgram(shader),
			g.assetManager,
			entity.Prefab.ModelRefs[0].Model,
			entity.AnimationPlayer,
			modelMatrix,
			entity.ID,
		)
	}

}
