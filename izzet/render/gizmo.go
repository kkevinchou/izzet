package render

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

func (r *Renderer) drawTranslationGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
	colors := map[int]mgl64.Vec3{
		gizmo.GizmoXAxisPickingID: mgl64.Vec3{1, 0, 0},
		gizmo.GizmoYAxisPickingID: mgl64.Vec3{0, 0, 1},
		gizmo.GizmoZAxisPickingID: mgl64.Vec3{0, 1, 0},
	}

	// in the range -1 - 1
	screenPosition, behind := worldToNDCPosition(*viewerContext, position)
	if behind {
		return
	}
	nearPlanePosition := ndcToWorldPosition(*viewerContext, mgl64.Vec3{screenPosition.X(), screenPosition.Y(), -float64(r.app.RuntimeConfig().Near)})
	renderPosition := nearPlanePosition.Sub(viewerContext.Position).Mul(settings.GizmoDistanceFactor).Add(nearPlanePosition)

	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	for entityID, axis := range gizmo.TranslationGizmo.EntityIDToAxis {
		shader.SetUniformUInt("entityID", uint32(entityID))
		lines := [][]mgl64.Vec3{{renderPosition, renderPosition.Add(axis.Direction)}}
		color := colors[entityID]

		if gizmo.TranslationGizmo.HoveredEntityID == entityID {
			color = mgl64.Vec3{1, 1, 0}
		}

		r.drawLines(*viewerContext, shader, lines, settings.GizmoAxisThickness, color)
	}

	planeColor := mgl32.Vec3{189.0 / 255, 24.0 / 255, 0.0 / 255}

	// handle XZ
	color := planeColor
	if gizmo.TranslationGizmo.HoveredEntityID == gizmo.GizmoXZAxisPickingID {
		color = mgl32.Vec3{1, 1, 0}
	}
	var scaledSize float32 = 0.25
	quadModelMatrix := mgl32.Translate3D(float32(renderPosition.X())+scaledSize, float32(renderPosition.Y()), float32(renderPosition.Z())+scaledSize)
	quadModelMatrix = quadModelMatrix.Mul4(mgl32.QuatRotate(math.Pi/2, mgl32.Vec3{1, 0, 0}).Mat4())
	quadModelMatrix = quadModelMatrix.Mul4(mgl32.Scale3D(scaledSize, scaledSize, scaledSize))

	shader.SetUniformMat4("model", quadModelMatrix)
	shader.SetUniformUInt("entityID", uint32(gizmo.GizmoXZAxisPickingID))
	shader.SetUniformVec3("color", color)
	quadVAO := getInternedQuadVAOPosition()
	gl.BindVertexArray(quadVAO)
	r.iztDrawArrays(0, 12)

	// handle XY
	color = planeColor
	if gizmo.TranslationGizmo.HoveredEntityID == gizmo.GizmoXYAxisPickingID {
		color = mgl32.Vec3{1, 1, 0}
	}

	quadModelMatrix = mgl32.Translate3D(float32(renderPosition.X())+scaledSize, float32(renderPosition.Y())+scaledSize, float32(renderPosition.Z()))
	quadModelMatrix = quadModelMatrix.Mul4(mgl32.Scale3D(scaledSize, scaledSize, scaledSize))

	shader.SetUniformMat4("model", quadModelMatrix)
	shader.SetUniformUInt("entityID", uint32(gizmo.GizmoXYAxisPickingID))
	shader.SetUniformVec3("color", color)
	gl.BindVertexArray(quadVAO)
	r.iztDrawArrays(0, 12)

	// handle YZ
	color = planeColor
	if gizmo.TranslationGizmo.HoveredEntityID == gizmo.GizmoYZAxisPickingID {
		color = mgl32.Vec3{1, 1, 0}
	}

	quadModelMatrix = mgl32.Translate3D(float32(renderPosition.X()), float32(renderPosition.Y())+scaledSize, float32(renderPosition.Z())+scaledSize)
	quadModelMatrix = quadModelMatrix.Mul4(mgl32.QuatRotate(math.Pi/2, mgl32.Vec3{0, 1, 0}).Mat4())
	quadModelMatrix = quadModelMatrix.Mul4(mgl32.Scale3D(scaledSize, scaledSize, scaledSize))

	shader.SetUniformMat4("model", quadModelMatrix)
	shader.SetUniformUInt("entityID", uint32(gizmo.GizmoYZAxisPickingID))
	shader.SetUniformVec3("color", color)
	gl.BindVertexArray(quadVAO)
	r.iztDrawArrays(0, 12)
}

func (r *Renderer) drawScaleGizmo(viewerContext *ViewerContext, shader *shaders.ShaderProgram, position mgl64.Vec3) {
	shader.Use()
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	screenPosition, behind := worldToNDCPosition(*viewerContext, position)
	if behind {
		return
	}
	nearPlanePosition := ndcToWorldPosition(*viewerContext, mgl64.Vec3{screenPosition.X(), screenPosition.Y(), -float64(r.app.RuntimeConfig().Near)})
	renderPosition := nearPlanePosition.Sub(viewerContext.Position).Mul(settings.GizmoDistanceFactor).Add(nearPlanePosition)

	colors := map[int]mgl64.Vec3{
		gizmo.GizmoXAxisPickingID:   mgl64.Vec3{1, 0, 0},
		gizmo.GizmoYAxisPickingID:   mgl64.Vec3{0, 0, 1},
		gizmo.GizmoZAxisPickingID:   mgl64.Vec3{0, 1, 0},
		gizmo.GizmoAllAxisPickingID: mgl64.Vec3{1, 1, 1},
	}
	hoverColor := mgl64.Vec3{1, 1, 0}

	cubeVAO := r.getCubeVAO(0.25)

	for entityID, axis := range gizmo.ScaleGizmo.EntityIDToAxis {
		shader.SetUniformUInt("entityID", uint32(entityID))
		shader.SetUniformMat4("model", mgl32.Ident4())

		color := colors[entityID]
		if gizmo.ScaleGizmo.HoveredEntityID == entityID || gizmo.ScaleGizmo.HoveredEntityID == gizmo.GizmoAllAxisPickingID {
			color = hoverColor
		}

		if entityID != gizmo.GizmoAllAxisPickingID {
			lines := [][]mgl64.Vec3{{renderPosition, renderPosition.Add(axis.Direction)}}
			r.drawLines(*viewerContext, shader, lines, settings.GizmoAxisThickness, color)
		}

		cubePosition := renderPosition.Add(axis.Direction)
		gl.BindVertexArray(cubeVAO)
		shader.SetUniformMat4("model", mgl32.Translate3D(float32(cubePosition.X()), float32(cubePosition.Y()), float32(cubePosition.Z())))
		shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
		shader.SetUniformFloat("intensity", 10)
		r.iztDrawArrays(0, 36)
	}
}
func (r *Renderer) drawCircleGizmo(viewerContext *ViewerContext, position mgl64.Vec3, renderContext RenderContext) {
	screenPosition, behind := worldToNDCPosition(*viewerContext, position)
	if behind {
		return
	}
	nearPlanePosition := ndcToWorldPosition(*viewerContext, mgl64.Vec3{screenPosition.X(), screenPosition.Y(), -float64(r.app.RuntimeConfig().Near)})
	renderPosition := nearPlanePosition.Sub(viewerContext.Position).Mul(settings.GizmoDistanceFactor).Add(nearPlanePosition)

	t := mgl32.Translate3D(float32(renderPosition[0]), float32(renderPosition[1]), float32(renderPosition[2]))

	rotations := []mgl32.Mat4{
		mgl32.Ident4(),
		mgl32.HomogRotate3DY(90 * math.Pi / 180),
		mgl32.HomogRotate3DX(-90 * math.Pi / 180),
	}

	pickingIDs := []int{
		gizmo.GizmoXDistancePickingID,
		gizmo.GizmoYDistancePickingID,
		gizmo.GizmoZDistancePickingID,
	}

	textures := []uint32{r.redCircleTexture, r.greenCircleTexture, r.blueCircleTexture}

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	for i := 0; i < 3; i++ {
		modelMatrix := t.Mul4(rotations[i])
		texture := textures[i]
		pickingID := pickingIDs[i]

		if pickingID == gizmo.RotationGizmo.HoveredEntityID {
			texture = r.yellowCircleTexture
		}

		r.drawTexturedQuad(viewerContext, r.shaderManager, texture, float32(renderContext.AspectRatio()), &modelMatrix, true, &pickingID)
	}
}

func (r *Renderer) drawCircle() {
	var vertices []float32 = []float32{
		-1, -1, 0,
		1, -1, 0,
		1, 1, 0,
		1, 1, 0,
		-1, 1, 0,
		-1, -1, 0,
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(vao)

	r.iztDrawArrays(0, 6)
}

// computes the near plane position for a given x y coordinate
func ndcToWorldPosition(viewerContext ViewerContext, directionVec mgl64.Vec3) mgl64.Vec3 {
	// ndcP := mgl64.Vec4{((x / float64(g.width)) - 0.5) * 2, ((y / float64(g.height)) - 0.5) * -2, -1, 1}
	nearPlanePos := viewerContext.InverseViewMatrix.Inv().Mul4(viewerContext.ProjectionMatrix.Inv()).Mul4x1(directionVec.Vec4(1))
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func worldToNDCPosition(viewerContext ViewerContext, worldPosition mgl64.Vec3) (mgl64.Vec2, bool) {
	screenPos := viewerContext.ProjectionMatrix.Mul4(viewerContext.InverseViewMatrix).Mul4x1(worldPosition.Vec4(1))
	behind := screenPos.Z() < 0
	screenPos = screenPos.Mul(1 / screenPos.W())
	return screenPos.Vec2(), behind
}
