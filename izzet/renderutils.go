package izzet

import (
	"fmt"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/model"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

func idToPickingColor(id int) mgl32.Vec3 {
	var r float32 = float32((id&0x000000FF)>>0) / 255
	var g float32 = float32((id&0x0000FF00)>>8) / 255
	var b float32 = float32((id&0x00FF0000)>>16) / 255
	return mgl32.Vec3{r, g, b}
}

func drawWIthID(viewerContext ViewerContext,
	shader *shaders.ShaderProgram,
	assetManager *assets.AssetManager,
	model *model.Model,
	animationPlayer *animation.AnimationPlayer,
	modelMatrix mgl64.Mat4,
	id int,
) {
	// TOOD(kevin): i hate this... Ideally we incorporate the model.RootTransforms to the vertex positions
	// and the animation poses so that we don't have to multiple this matrix every frame.
	m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix).Mul4(model.RootTransforms())
	_, r, _ := utils.Decompose(m32ModelMatrix)

	shader.Use()
	shader.SetUniformMat4("model", m32ModelMatrix)
	shader.SetUniformMat4("modelRotationMatrix", r.Mat4())
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformVec3("pickingColor", idToPickingColor(id))

	if animationPlayer != nil {
		animationTransforms := animationPlayer.AnimationTransforms()
		// if animationTransforms is nil, the shader will execute reading into invalid memory
		// so, we need to explicitly guard for this
		if animationTransforms == nil {
			panic("animationTransforms not found")
		}
		for i := 0; i < len(animationTransforms); i++ {
			shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
		}
	}

	for _, meshChunk := range model.MeshChunks() {
		gl.BindVertexArray(meshChunk.VAO())
		gl.DrawElements(gl.TRIANGLES, int32(meshChunk.VertexCount()), gl.UNSIGNED_INT, nil)
	}
}

// drawTris draws a list of triangles in winding order. each triangle is defined with 3 consecutive points
func drawTris(viewerContext ViewerContext, shader *shaders.ShaderProgram, points []mgl64.Vec3, color mgl64.Vec3) {
	var vertices []float32
	for _, point := range points {
		vertices = append(vertices, float32(point.X()), float32(point.Y()), float32(point.Z()))
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
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformFloat("alpha", float32(1))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawModel(viewerContext ViewerContext,
	lightContext LightContext,
	shadowMap *ShadowMap,
	shader *shaders.ShaderProgram,
	assetManager *assets.AssetManager,
	model *model.Model,
	animationPlayer *animation.AnimationPlayer,
	modelMatrix mgl64.Mat4,
) {
	// TOOD(kevin): i hate this... Ideally we incorporate the model.RootTransforms to the vertex positions
	// and the animation poses so that we don't have to multiple this matrix every frame.
	m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix).Mul4(model.RootTransforms())
	_, r, _ := utils.Decompose(m32ModelMatrix)

	shader.Use()
	shader.SetUniformMat4("model", m32ModelMatrix)
	shader.SetUniformMat4("modelRotationMatrix", r.Mat4())
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", float32(shadowMap.ShadowDistance()))
	shader.SetUniformVec3("directionalLightDir", utils.Vec3F64ToF32(lightContext.DirectionalLightDir))
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformInt("shadowMap", 31)

	if animationPlayer != nil {
		animationTransforms := animationPlayer.AnimationTransforms()
		// if animationTransforms is nil, the shader will execute reading into invalid memory
		// so, we need to explicitly guard for this
		if animationTransforms == nil {
			panic("animationTransforms not found")
		}
		for i := 0; i < len(animationTransforms); i++ {
			shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
		}
	}

	gl.ActiveTexture(gl.TEXTURE31)
	gl.BindTexture(gl.TEXTURE_2D, shadowMap.DepthTexture())

	for _, meshChunk := range model.MeshChunks() {
		if pbr := meshChunk.PBRMaterial(); pbr != nil {
			shader.SetUniformInt("hasPBRMaterial", 1)
			shader.SetUniformVec4("pbrBaseColorFactor", pbr.PBRMetallicRoughness.BaseColorFactor)

			if pbr.PBRMetallicRoughness.BaseColorTextureIndex != nil {
				shader.SetUniformInt("hasPBRBaseColorTexture", 1)
			} else {
				shader.SetUniformInt("hasPBRBaseColorTexture", 0)
			}

			shader.SetUniformVec3("albedo", pbr.PBRMetallicRoughness.BaseColorFactor.Vec3())
			shader.SetUniformFloat("metallic", pbr.PBRMetallicRoughness.MetalicFactor)
			shader.SetUniformFloat("roughness", pbr.PBRMetallicRoughness.RoughnessFactor)
			shader.SetUniformFloat("ao", 1.0)
		} else {
			shader.SetUniformInt("hasPBRMaterial", 0)
		}

		gl.ActiveTexture(gl.TEXTURE0)
		var textureID uint32
		if meshChunk.TextureID() != nil {
			textureID = *meshChunk.TextureID()
		} else {
			texture := assetManager.GetTexture(defaultTexture)
			textureID = texture.ID
		}
		gl.BindTexture(gl.TEXTURE_2D, textureID)

		gl.BindVertexArray(meshChunk.VAO())
		gl.DrawElements(gl.TRIANGLES, int32(meshChunk.VertexCount()), gl.UNSIGNED_INT, nil)
	}
}

func toRadians(degrees float64) float64 {
	return degrees / 180 * math.Pi
}

func drawLines(viewerContext ViewerContext, shader *shaders.ShaderProgram, lines [][]mgl64.Vec3, thickness float64, color mgl64.Vec3) {
	var points []mgl64.Vec3
	for _, line := range lines {
		start := line[0]
		end := line[1]
		length := end.Sub(start).Len()

		dir := end.Sub(start).Normalize()
		q := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)

		for _, dp := range defaultPoints(thickness, length) {
			newEnd := q.Rotate(dp).Add(start)
			points = append(points, newEnd)
		}
	}
	drawTris(viewerContext, shader, points, color)
}
func defaultPoints(thickness float64, length float64) []mgl64.Vec3 {
	var ht float64 = thickness / 2
	return []mgl64.Vec3{
		// front
		{-ht, -ht, 0},
		{ht, -ht, 0},
		{ht, ht, 0},

		{ht, ht, 0},
		{-ht, ht, 0},
		{-ht, -ht, 0},

		// back
		{ht, ht, -length},
		{ht, -ht, -length},
		{-ht, -ht, -length},

		{-ht, -ht, -length},
		{-ht, ht, -length},
		{ht, ht, -length},

		// right
		{ht, -ht, 0},
		{ht, -ht, -length},
		{ht, ht, -length},

		{ht, ht, -length},
		{ht, ht, 0},
		{ht, -ht, 0},

		// left
		{-ht, ht, -length},
		{-ht, -ht, -length},
		{-ht, -ht, 0},

		{-ht, -ht, 0},
		{-ht, ht, 0},
		{-ht, ht, -length},

		// top
		{ht, ht, 0},
		{ht, ht, -length},
		{-ht, ht, 0},

		{-ht, ht, 0},
		{ht, ht, -length},
		{-ht, ht, -length},

		// bottom
		{-ht, -ht, 0},
		{ht, -ht, -length},
		{ht, -ht, 0},

		{-ht, -ht, -length},
		{ht, -ht, -length},
		{-ht, -ht, 0},
	}
}

func drawTexturedQuad(viewerContext *ViewerContext, shaderManager *shaders.ShaderManager, texture uint32, hudScale float32, aspectRatio float32, modelMatrix *mgl32.Mat4, doubleSided bool) {
	// texture coords top left = 0,0 | bottom right = 1,1
	var vertices []float32 = []float32{
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
		1 * hudScale, -1 * hudScale, 0, 1.0, 0.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		-1 * hudScale, 1 * hudScale, 0, 0.0, 1.0,
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
	}

	var backVertices []float32 = []float32{
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		1 * hudScale, -1 * hudScale, 0, 1.0, 0.0,
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
		-1 * hudScale, 1 * hudScale, 0, 0.0, 1.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
	}

	if doubleSided {
		vertices = append(vertices, backVertices...)
	}

	// if we're just rendering something directly to screen without a world position
	// adjust x coord by aspect ratio
	if modelMatrix == nil {
		for i := 0; i < len(vertices); i += 5 {
			x := vertices[i]
			vertices[i] = x / aspectRatio
		}
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	if modelMatrix != nil {
		shader := shaderManager.GetShaderProgram("basic_quad_world")
		shader.Use()
		shader.SetUniformMat4("model", *modelMatrix)
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	} else {
		shader := shaderManager.GetShaderProgram("basic_quad")
		shader.Use()
	}

	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawCircle(shader *shaders.ShaderProgram, color mgl64.Vec4) {
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

	shader.Use()
	shader.SetUniformVec4("color", utils.Vec4F64To4F32(color))

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

// drawHUDTextureToQuad does a shitty perspective based rendering of a flat texture
func drawHUDTextureToQuad(viewerContext ViewerContext, shader *shaders.ShaderProgram, texture uint32, hudScale float32) {
	// texture coords top left = 0,0 | bottom right = 1,1
	var vertices []float32 = []float32{
		// front
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
		1 * hudScale, -1 * hudScale, 0, 1.0, 0.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		-1 * hudScale, 1 * hudScale, 0, 0.0, 1.0,
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	shader.Use()
	shader.SetUniformMat4("model", mgl32.Translate3D(0, 0, -2))
	shader.SetUniformMat4("view", mgl32.Ident4())
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
