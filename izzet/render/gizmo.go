package render

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/constants"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

func (r *Renderer) drawTranslationGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
	colors := []mgl64.Vec3{mgl64.Vec3{1, 0, 0}, mgl64.Vec3{0, 0, 1}, mgl64.Vec3{0, 1, 0}}

	// in the range -1 - 1
	screenPosition := r.app.WorldToNDCPosition(*viewerContext, position)
	nearPlanePosition := r.app.NDCToWorldPosition(*viewerContext, mgl64.Vec3{screenPosition.X(), screenPosition.Y(), -float64(panels.DBG.Near)})
	renderPosition := nearPlanePosition.Sub(viewerContext.Position).Mul(settings.GizmoDistanceFactor).Add(nearPlanePosition)

	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	for i, axis := range gizmo.T.Axes {
		shader.SetUniformUInt("entityID", uint32(constants.GizmoTranslationXPickingID)+uint32(i))
		lines := [][]mgl64.Vec3{{renderPosition, renderPosition.Add(axis)}}
		color := colors[i]
		if i == gizmo.T.HoverIndex {
			color = mgl64.Vec3{1, 1, 0}
		}
		drawLines2(*viewerContext, shader, lines, settings.GizmoAxisThickness, color)
	}
}

func (r *Renderer) drawScaleGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
	shader.Use()
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	screenPosition := r.app.WorldToNDCPosition(*viewerContext, position)
	nearPlanePosition := r.app.NDCToWorldPosition(*viewerContext, mgl64.Vec3{screenPosition.X(), screenPosition.Y(), -float64(panels.DBG.Near)})
	renderPosition := nearPlanePosition.Sub(viewerContext.Position).Mul(settings.GizmoDistanceFactor).Add(nearPlanePosition)

	axisColors := map[gizmo.AxisType]mgl64.Vec3{
		gizmo.XAxis: mgl64.Vec3{1, 0, 0},
		gizmo.YAxis: mgl64.Vec3{0, 0, 1},
		gizmo.ZAxis: mgl64.Vec3{0, 1, 0},
	}
	hoverColor := mgl64.Vec3{1, 1, 0}

	cubeVAO := r.getCubeVAO(0.2)

	for _, axis := range gizmo.S.Axes {
		shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
		lines := [][]mgl64.Vec3{{renderPosition, renderPosition.Add(axis.Vector)}}
		color := axisColors[axis.Type]
		if axis.Type == gizmo.S.HoveredAxisType || gizmo.S.HoveredAxisType == gizmo.AllAxis {
			color = hoverColor
		}
		drawLines2(*viewerContext, shader, lines, settings.GizmoAxisThickness, color)

		cubePosition := renderPosition.Add(axis.Vector)

		gl.BindVertexArray(cubeVAO)
		shader.SetUniformMat4("model", mgl32.Translate3D(float32(cubePosition.X()), float32(cubePosition.Y()), float32(cubePosition.Z())))
		shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
		shader.SetUniformFloat("intensity", 10)
		iztDrawArrays(0, 36)
	}

	var cubeColor = mgl64.Vec3{1, 1, 1}
	if gizmo.S.HoveredAxisType == gizmo.AllAxis {
		cubeColor = hoverColor
	}
	gl.BindVertexArray(cubeVAO)
	shader.SetUniformMat4("model", mgl32.Translate3D(float32(renderPosition.X()), float32(renderPosition.Y()), float32(renderPosition.Z())))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(cubeColor))
	shader.SetUniformFloat("intensity", 10)
	iztDrawArrays(0, 36)
}
func (r *Renderer) drawCircleGizmo(cameraViewerContext *ViewerContext, position mgl64.Vec3, renderContext RenderContext) {
	defer resetGLRenderSettings(r.renderFBO)

	t := mgl32.Translate3D(float32(position[0]), float32(position[1]), float32(position[2]))
	s := mgl32.Scale3D(25, 25, 25)

	rotations := []mgl32.Mat4{
		mgl32.Ident4(),
		mgl32.HomogRotate3DY(90 * math.Pi / 180),
		mgl32.HomogRotate3DX(-90 * math.Pi / 180),
	}

	textures := []uint32{r.redCircleTexture, r.greenCircleTexture, r.blueCircleTexture}

	r.renderCircle()
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	for i := 0; i < 3; i++ {
		modelMatrix := t.Mul4(rotations[i]).Mul4(s)
		texture := textures[i]
		if i == gizmo.R.HoverIndex {
			texture = r.yellowCircleTexture
		}
		drawTexturedQuad(cameraViewerContext, r.shaderManager, texture, float32(renderContext.AspectRatio()), &modelMatrix, true)
	}
}

func (r *Renderer) renderCircle() {
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
