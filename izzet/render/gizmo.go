package render

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/kitolib/shaders"
)

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
	axisColors := map[gizmo.AxisType]mgl64.Vec3{
		gizmo.XAxis: mgl64.Vec3{1, 0, 0},
		gizmo.YAxis: mgl64.Vec3{0, 0, 1},
		gizmo.ZAxis: mgl64.Vec3{0, 1, 0},
	}
	var cubeSize float64 = 5
	cubeLineThickness := 0.5
	hoverColor := mgl64.Vec3{1, 1, 0}

	for _, axis := range gizmo.S.Axes {
		lines := [][]mgl64.Vec3{
			[]mgl64.Vec3{position, position.Add(axis.Vector)},
		}
		color := axisColors[axis.Type]
		if axis.Type == gizmo.S.HoveredAxisType || gizmo.S.HoveredAxisType == gizmo.AllAxis {
			color = hoverColor
		}
		drawLines(*viewerContext, shader, lines, 1, color)

		cLines := cubeLines(cubeSize)
		for _, line := range cLines {
			for i := range line {
				line[i] = line[i].Add(position).Add(axis.Vector)
			}
		}
		drawLines(*viewerContext, shader, cLines, cubeLineThickness, color)
	}

	// center of scale gizmo
	cLines := cubeLines(cubeSize)
	for _, line := range cLines {
		for i := range line {
			line[i] = line[i].Add(position)
		}
	}
	var cubeColor = mgl64.Vec3{1, 1, 1}
	if gizmo.S.HoveredAxisType == gizmo.AllAxis {
		cubeColor = hoverColor
	}
	drawLines(*viewerContext, shader, cLines, cubeLineThickness, cubeColor)
}
func (r *Renderer) drawCircleGizmo(cameraViewerContext *ViewerContext, position mgl64.Vec3, renderContext RenderContext) {
	defer resetGLRenderSettings()
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))

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
		drawTexturedQuad(cameraViewerContext, r.shaderManager, texture, 1, float32(renderContext.AspectRatio()), &modelMatrix, true)
	}
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
