package izzet

import (
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/menus"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/utils"
)

var (
	defaultTexture string = "color_grid"

	// shadow map properties
	shadowmapZOffset     float64 = 400
	fovx                 float64 = 105
	near                 float64 = 1
	far                  float64 = 3000
	shadowDistanceFactor float64 = .4 // proportion of view fustrum to include in shadow cuboid
)

func (g *Izzet) Render(delta time.Duration) {
	// configure camera viewer context
	position := g.camera.Position
	orientation := g.camera.Orientation

	viewerViewMatrix := orientation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := ViewerContext{
		Position:    position,
		Orientation: orientation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(g.fovY), g.aspectRatio, near, far),
	}

	// configure light viewer context
	modelSpaceFrustumPoints := CalculateFrustumPoints(position, orientation, near, far, g.fovY, g.aspectRatio, shadowDistanceFactor)

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

	g.renderToDepthMap(lightViewerContext, lightContext)
	g.renderToDisplay(cameraViewerContext, lightContext)
	g.renderImgui()

	g.window.GLSwap()
}

func (g *Izzet) renderImgui() {
	g.platform.NewFrame()
	imgui.NewFrame()

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

	menus.BuildExplorer(g.entities)
	menus.BuildPrefabs(g.entities)

	imgui.PopStyleColor()
	// imgui.PopStyleVarV(6)
	imgui.PopStyleVarV(10)
	var open bool
	imgui.ShowDemoWindow(&open)

	imgui.Render()
	g.imguiRenderer.Render(g.platform.DisplaySize(), g.platform.FramebufferSize(), imgui.RenderedDrawData())
}

func (g *Izzet) renderToDisplay(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()

	gl.Viewport(0, 0, int32(settings.Width), int32(settings.Height))
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

	for _, entity := range g.entities {
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
			entity.Model,
			entity.AnimationPlayer,
			modelMatrix,
		)
	}
}

func createModelMatrix(scaleMatrix, rotationMatrix, translationMatrix mgl64.Mat4) mgl64.Mat4 {
	return translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
}
