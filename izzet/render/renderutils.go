package render

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/model"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

var lineCache map[string][]mgl64.Vec3

func init() {
	lineCache = map[string][]mgl64.Vec3{}
}

func genLineKey(thickness, length float64) string {
	return fmt.Sprintf("%.3f_%.3f", thickness, length)
}

func idToPickingColor(id int) mgl32.Vec3 {
	var r float32 = float32((id&0x000000FF)>>0) / 255
	var g float32 = float32((id&0x0000FF00)>>8) / 255
	var b float32 = float32((id&0x00FF0000)>>16) / 255
	return mgl32.Vec3{r, g, b}
}

func drawModelWIthID(viewerContext ViewerContext,
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

	if animationPlayer != nil && animationPlayer.CurrentAnimation() != "" {
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
func drawTris(viewerContext ViewerContext, points []mgl64.Vec3, color mgl64.Vec3) {
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
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

// i considered using uniform blocks but the memory layout management seems like a huge pain
// https://stackoverflow.com/questions/38172696/should-i-ever-use-a-vec3-inside-of-a-uniform-buffer-or-shader-storage-buffer-o
func setupLightingUniforms(shader *shaders.ShaderProgram, lights []*entities.Entity) {
	if len(lights) > settings.MaxLightCount {
		panic(fmt.Sprintf("light count of %d exceeds max %d", len(lights), settings.MaxLightCount))
	}

	shader.SetUniformInt("lightCount", int32(len(lights)))
	for i, light := range lights {
		lightInfo := light.LightInfo
		diffuse := utils.Vec4F64ToF32(lightInfo.Diffuse)
		if lightInfo.Type == 1 {
			diffuse[3] = float32(panels.DBG.PointLightIntensity)
		}
		shader.SetUniformInt(fmt.Sprintf("lights[%d].type", i), int32(lightInfo.Type))
		shader.SetUniformVec3(fmt.Sprintf("lights[%d].dir", i), utils.Vec3F64ToF32(lightInfo.Direction))
		shader.SetUniformVec4(fmt.Sprintf("lights[%d].diffuse", i), diffuse)
		shader.SetUniformVec3(fmt.Sprintf("lights[%d].position", i), utils.Vec3F64ToF32(light.WorldPosition()))
	}
}

func drawModel(viewerContext ViewerContext,
	lightContext LightContext,
	shadowMap *ShadowMap,
	shader *shaders.ShaderProgram,
	assetManager *assets.AssetManager,
	model *model.Model,
	animationPlayer *animation.AnimationPlayer,
	modelMatrix mgl64.Mat4,
	pointLightDepthCubeMap uint32,
) {
	// TOOD(kevin): i hate this... Ideally we incorporate the model.RootTransforms to the vertex positions
	// and the animation poses so that we don't have to multiple this matrix every frame.
	m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix).Mul4(model.RootTransforms())
	_, r, _ := utils.Decompose(m32ModelMatrix)

	// TODO refactor - move common shader setup outside of draw model
	shader.Use()
	shader.SetUniformMat4("model", m32ModelMatrix)
	shader.SetUniformMat4("modelRotationMatrix", r.Mat4())
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", float32(shadowMap.ShadowDistance()))
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformInt("shadowMap", 31)
	shader.SetUniformInt("depthCubeMap", 30)

	setupLightingUniforms(shader, lightContext.Lights)

	if animationPlayer != nil && animationPlayer.CurrentAnimation() != "" {
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

	shader.SetUniformFloat("bias", panels.DBG.PointLightBias)
	shader.SetUniformFloat("far_plane", float32(settings.DepthCubeMapFar))

	gl.ActiveTexture(gl.TEXTURE30)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, pointLightDepthCubeMap)

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
			shader.SetUniformFloat("roughness", panels.DBG.Roughness)
			shader.SetUniformFloat("metallic", panels.DBG.Metallic)
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

		for _, dp := range linePoints(thickness, length) {
			newEnd := q.Rotate(dp).Add(start)
			points = append(points, newEnd)
		}
	}
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformFloat("alpha", float32(1))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	drawTris(viewerContext, points, color)
}

func cubeLines(length float64) [][]mgl64.Vec3 {
	directions := [][]float64{
		[]float64{-1, 1, 0.5},
		[]float64{-1, -1, 0.5},
		[]float64{1, -1, 0.5},
		[]float64{1, 1, 0.5},
	}

	position := mgl64.Vec3{}
	var lines [][]mgl64.Vec3
	var frontPoints []mgl64.Vec3

	// front points
	for _, direction := range directions {
		point := position.Add(mgl64.Vec3{direction[0] * length / 2, direction[1] * length / 2, direction[2] * length})
		frontPoints = append(frontPoints, point)
	}
	for i := range frontPoints {
		line := []mgl64.Vec3{frontPoints[i], frontPoints[(i+1)%len(frontPoints)]}
		lines = append(lines, line)
	}

	// back points
	var backPoints []mgl64.Vec3
	for _, point := range frontPoints {
		backPoints = append(backPoints, point.Add(mgl64.Vec3{0, 0, -length}))
	}
	for i := range backPoints {
		line := []mgl64.Vec3{backPoints[i], backPoints[(i+1)%len(backPoints)]}
		lines = append(lines, line)
	}

	// connect front and back
	for i := range frontPoints {
		line := []mgl64.Vec3{frontPoints[i], backPoints[i]}
		lines = append(lines, line)
	}

	return lines
}

// TODO: find a clean way to take 8 cube points and generate both
// a wireframe lines of the cube and the triangulated lines
func cubePoints(length float64) []mgl64.Vec3 {
	var ht float64 = length / 2
	return []mgl64.Vec3{
		// front
		{-ht, -ht, ht},
		{ht, -ht, ht},
		{ht, ht, ht},

		{ht, ht, ht},
		{-ht, ht, ht},
		{-ht, -ht, ht},

		// back
		{ht, ht, -ht},
		{ht, -ht, -ht},
		{-ht, -ht, -ht},

		{-ht, -ht, -ht},
		{-ht, ht, -ht},
		{ht, ht, -ht},

		// right
		{ht, -ht, ht},
		{ht, -ht, -ht},
		{ht, ht, -ht},

		{ht, ht, -ht},
		{ht, ht, ht},
		{ht, -ht, ht},

		// left
		{-ht, ht, -ht},
		{-ht, -ht, -ht},
		{-ht, -ht, ht},

		{-ht, -ht, ht},
		{-ht, ht, ht},
		{-ht, ht, -ht},

		// top
		{ht, ht, ht},
		{ht, ht, -ht},
		{-ht, ht, ht},

		{-ht, ht, ht},
		{ht, ht, -ht},
		{-ht, ht, -ht},

		// bottom
		{-ht, -ht, ht},
		{ht, -ht, -ht},
		{ht, -ht, ht},

		{-ht, -ht, -ht},
		{ht, -ht, -ht},
		{-ht, -ht, ht},
	}
}

func linePoints(thickness float64, length float64) []mgl64.Vec3 {
	cacheKey := genLineKey(thickness, length)
	if _, ok := lineCache[cacheKey]; ok {
		return lineCache[cacheKey]
	}

	var ht float64 = thickness / 2
	linePoints := []mgl64.Vec3{
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

	lineCache[cacheKey] = linePoints
	return linePoints
}

func drawWithNDC(shaderManager *shaders.ShaderManager) {
	// triangle
	// var vertices []float32 = []float32{
	// 	-0.5, -0.5, 1,
	// 	0.5, -0.5, 1,
	// 	0.0, 0.5, 1,
	// }

	var back float32 = 1

	// full screen
	var vertices []float32 = []float32{
		-1, 1, back,
		-1, -1, back,
		1, -1, back,

		1, -1, back,
		1, 1, back,
		-1, 1, back,
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

	shader := shaderManager.GetShaderProgram("skybox")
	shader.Use()
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawBillboardTexture(
	viewerContext *ViewerContext,
	shaderManager *shaders.ShaderManager,
	texture uint32,
	modelMatrix mgl64.Mat4,
	cameraUp mgl64.Vec3,
	cameraRight mgl64.Vec3,
) {
	topLeft := utils.Vec3F64ToF32(cameraRight.Mul(-1).Add(cameraUp))
	bottomLeft := utils.Vec3F64ToF32(cameraRight.Mul(-1).Add(cameraUp.Mul(-1)))
	topRight := utils.Vec3F64ToF32(cameraRight.Mul(1).Add(cameraUp))
	bottomRight := utils.Vec3F64ToF32(cameraRight.Mul(1).Add(cameraUp.Mul(-1)))

	var vertices []float32 = []float32{
		bottomLeft.X(), bottomLeft.Y(), bottomLeft.Z(), 0.0, 0.0,
		bottomRight.X(), bottomRight.Y(), bottomRight.Z(), 1.0, 0.0,
		topRight.X(), topRight.Y(), topRight.Z(), 1.0, 1.0,

		topRight.X(), topRight.Y(), topRight.Z(), 1.0, 1.0,
		topLeft.X(), topLeft.Y(), topLeft.Z(), 0.0, 1.0,
		bottomLeft.X(), bottomLeft.Y(), bottomLeft.Z(), 0.0, 0.0,
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

	shader := shaderManager.GetShaderProgram("basic_quad_world")
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}
func drawTexturedQuad(viewerContext *ViewerContext, shaderManager *shaders.ShaderManager, texture uint32, hudScale float32, aspectRatio float32, modelMatrix *mgl32.Mat4, doubleSided bool) {
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

func (r *Renderer) clearMainFrameBuffer() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderSkybox() {
	defer resetGLRenderSettings()
	w, h := r.world.Window().GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))

	drawWithNDC(r.shaderManager)
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
