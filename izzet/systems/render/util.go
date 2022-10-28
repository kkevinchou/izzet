package render

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/lib/collision/collider"
	"github.com/kkevinchou/izzet/lib/font"
	"github.com/kkevinchou/izzet/lib/libutils"
	utils "github.com/kkevinchou/izzet/lib/libutils"
	"github.com/kkevinchou/izzet/lib/shaders"
	"github.com/kkevinchou/izzet/lib/textures"
)

const (
	defaultTexture    = "color_grid"
	healthHUDMaxWidth = 30.0
)

func drawModel(viewerContext ViewerContext, lightContext LightContext, shadowMap *ShadowMap, shader *shaders.ShaderProgram, meshComponent *components.MeshComponent, animationComponent *components.AnimationComponent, modelMatrix mgl64.Mat4, modelRotationMatrix mgl64.Mat4) {
	model := meshComponent.Model

	// TOOD(kevin): i hate this... Ideally we incorporate the model.RootTransforms to the vertex positions
	// and the animation poses so that we don't have to multiple this matrix every frame.
	m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix).Mul4(model.RootTransforms())
	_, r, _ := libutils.Decompose(m32ModelMatrix)

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

	if animationComponent != nil {
		animationTransforms := animationComponent.Player.AnimationTransforms()
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
			assetManager := directory.GetDirectory().AssetManager()
			texture := assetManager.GetTexture(defaultTexture)
			textureID = texture.ID
		}
		gl.BindTexture(gl.TEXTURE_2D, textureID)

		gl.BindVertexArray(meshChunk.VAO())
		gl.DrawElements(gl.TRIANGLES, int32(meshChunk.VertexCount()), gl.UNSIGNED_INT, nil)
	}
}

func drawTriMeshCollider(viewerContext ViewerContext, lightContext LightContext, shader *shaders.ShaderProgram, color mgl64.Vec3, triMeshCollider *collider.TriMesh) {
	var vertices []float32

	for _, triangle := range triMeshCollider.Triangles {
		for _, point := range triangle.Points {
			vertices = append(vertices, float32(point.X()), float32(point.Y()), float32(point.Z()))
		}
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
	shader.SetUniformMat4("model", mgl32.Ident4())
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	shader.SetUniformFloat("alpha", float32(0.3))
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawCapsuleCollider(viewerContext ViewerContext, lightContext LightContext, shader *shaders.ShaderProgram, color mgl64.Vec3, capsuleCollider *collider.Capsule, billboardModelMatrix mgl64.Mat4) {
	radius := float32(capsuleCollider.Radius)
	top := float32(capsuleCollider.Top.Y()) + radius
	bottom := float32(capsuleCollider.Bottom.Y()) - radius

	vertices := []float32{
		-radius, bottom, 0,
		radius, bottom, 0,
		radius, top, 0,
		radius, top, 0,
		-radius, top, 0,
		-radius, bottom, 0,
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
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(billboardModelMatrix))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	shader.SetUniformFloat("alpha", float32(0.3))
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawHealthHUD(healthComponent *components.HealthComponent, viewerContext ViewerContext, lightContext LightContext, shader *shaders.ShaderProgram, color mgl64.Vec3, billboardModelMatrix mgl64.Mat4) {
	var vertices []float32

	verticalOffset := mgl64.Vec3{0, 70, 0}
	horizontalOffset := mgl64.Vec3{-healthHUDMaxWidth / 2, 0, 0}
	offsetPosition := horizontalOffset.Add(verticalOffset)
	width := healthComponent.Data.Value / 100.0 * healthHUDMaxWidth
	height := 2.0
	points := []mgl64.Vec3{
		offsetPosition.Add(mgl64.Vec3{0, 0, 0}),
		offsetPosition.Add(mgl64.Vec3{width, 0, 0}),
		offsetPosition.Add(mgl64.Vec3{width, height, 0}),
		offsetPosition.Add(mgl64.Vec3{width, height, 0}),
		offsetPosition.Add(mgl64.Vec3{0, height, 0}),
		offsetPosition.Add(mgl64.Vec3{0, 0, 0}),
	}

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
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(billboardModelMatrix))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	shader.SetUniformFloat("alpha", float32(1))
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawSkyBox(viewerContext ViewerContext, sb *SkyBox, shader *shaders.ShaderProgram, frontTexture, topTexture, leftTexture, rightTexture, bottomTexture, backTexture *textures.Texture) {
	textures := []*textures.Texture{frontTexture, topTexture, leftTexture, rightTexture, bottomTexture, backTexture}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(sb.VAO())
	shader.Use()
	shader.SetUniformInt("skyboxTexture", 0)
	shader.SetUniformMat4("model", mgl32.Ident4())
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.Orientation.Mat4().Inv()))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	for i := 0; i < 6; i++ {
		gl.BindTexture(gl.TEXTURE_2D, textures[i].ID)
		gl.DrawArrays(gl.TRIANGLES, int32(i*6), 6)
	}
}

// drawText draws text at an x,y position that represents a fractional placement (0 -> 1)
// drawText expects the glyphs within `font` to be of equal width and height
func drawText(shader *shaders.ShaderProgram, font font.Font, text string, x, y float32) {
	var vertices []float32

	// assuming the height of all glyphs are equal - may not be the case in the future
	var glyphHeight float32
	for _, glyph := range font.Glyphs {
		glyphHeight = float32(glyph.Height)
		break
	}

	// convert porportion to pixel value
	x = x * float32(settings.Width)
	y = float32(settings.Height)*(1-y) - float32(glyphHeight)

	var xOffset float32
	var yOffset float32

	textureID := font.TextureID
	for _, c := range text {
		stringChar := string(c)

		if stringChar == "\n" {
			xOffset = 0
			yOffset++
			continue
		}

		glyph := font.Glyphs[stringChar]
		if _, ok := font.Glyphs[stringChar]; !ok {
			panic(fmt.Sprintf("glyph %s not found in font", stringChar))
		}

		width := float32(glyph.Width)
		height := float32(glyph.Height)

		textureX := float32(glyph.TextureCoords.X())
		textureY := float32(glyph.TextureCoords.Y())
		widthTextureCoord := (float32(glyph.Width) / float32(font.TotalWidth))
		heightTextureCoord := (float32(glyph.Height) / float32(font.TotalHeight))

		var characterVertices []float32 = []float32{
			xOffset * width, -(yOffset * glyphHeight), -5, textureX, textureY,
			(xOffset + 1) * width, -(yOffset * glyphHeight), -5, textureX + widthTextureCoord, textureY,
			(xOffset + 1) * width, height - (yOffset * glyphHeight), -5, textureX + widthTextureCoord, heightTextureCoord,

			(xOffset + 1) * width, height - (yOffset * glyphHeight), -5, textureX + widthTextureCoord, heightTextureCoord,
			xOffset * width, height - (yOffset * glyphHeight), -5, textureX, heightTextureCoord,
			xOffset * width, -(yOffset * glyphHeight), -5, textureX, textureY,
		}

		xOffset += 1
		vertices = append(vertices, characterVertices...)
	}

	// offset based on passed in x, y position which is constant across all characters
	for i := 0; i < len(vertices); i += 5 {
		vertices[i] = vertices[i] + x
		vertices[i+1] = vertices[i+1] + y
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
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	shader.Use()
	shader.SetUniformMat4("model", mgl32.Ident4())
	shader.SetUniformMat4("view", mgl32.Ident4())
	shader.SetUniformMat4("projection", mgl32.Ortho(0, float32(settings.Width), 0, float32(settings.Height), 1, 100))

	numCharacters := len(text)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(numCharacters*6))
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
	shader.SetUniformMat4("model", mgl32.Translate3D(1.2, 0.8, -2))
	shader.SetUniformMat4("view", mgl32.Ident4())
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
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
		{-ht, ht, 0},
		{ht, ht, -length},
		{ht, ht, 0},

		{-ht, ht, -length},
		{ht, ht, -length},
		{-ht, ht, 0},
	}
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

func drawSpatialPartition(viewerContext ViewerContext, shader *shaders.ShaderProgram, color mgl64.Vec3, spatialPartition *spatialpartition.SpatialPartition, thickness float64) {
	var allLines [][]mgl64.Vec3

	d := spatialPartition.PartitionDimension * spatialPartition.PartitionCount

	var baseHorizontalLines [][]mgl64.Vec3

	// lines along z axis
	for i := 0; i < spatialPartition.PartitionCount+1; i++ {
		baseHorizontalLines = append(baseHorizontalLines,
			[]mgl64.Vec3{{float64(-d/2 + i*spatialPartition.PartitionDimension), float64(-d / 2), float64(-d / 2)}, {float64(-d/2 + i*spatialPartition.PartitionDimension), float64(-d / 2), float64(d / 2)}},
		)
	}

	// // lines along x axis
	for i := 0; i < spatialPartition.PartitionCount+1; i++ {
		baseHorizontalLines = append(baseHorizontalLines,
			[]mgl64.Vec3{{float64(-d / 2), float64(-d / 2), float64(-d/2 + i*spatialPartition.PartitionDimension)}, {float64(d / 2), float64(-d / 2), float64(-d/2 + i*spatialPartition.PartitionDimension)}},
		)
	}

	for i := 0; i < spatialPartition.PartitionCount+1; i++ {
		for _, b := range baseHorizontalLines {
			allLines = append(allLines,
				[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i * spatialPartition.PartitionDimension), 0}), b[1].Add(mgl64.Vec3{0, float64(i * spatialPartition.PartitionDimension), 0})},
			)
		}
	}

	var baseVerticalLines [][]mgl64.Vec3

	for i := 0; i < spatialPartition.PartitionCount+1; i++ {
		baseVerticalLines = append(baseVerticalLines,
			[]mgl64.Vec3{{float64(-d/2 + i*spatialPartition.PartitionDimension), float64(-d / 2), float64(-d / 2)}, {float64(-d/2 + i*spatialPartition.PartitionDimension), float64(d / 2), float64(-d / 2)}},
		)
	}

	for i := 0; i < spatialPartition.PartitionCount+1; i++ {
		for _, b := range baseVerticalLines {
			allLines = append(allLines,
				[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i * spatialPartition.PartitionDimension)}), b[1].Add(mgl64.Vec3{0, 0, float64(i * spatialPartition.PartitionDimension)})},
			)
		}
	}

	drawLines(
		viewerContext,
		shader,
		allLines,
		thickness,
		color,
	)
}

func drawAABB(viewerContext ViewerContext, shader *shaders.ShaderProgram, color mgl64.Vec3, aabb *collider.BoundingBox, thickness float64) {
	var allLines [][]mgl64.Vec3

	d := aabb.MaxVertex.Sub(aabb.MinVertex)
	xd := d.X()
	yd := d.Y()
	zd := d.Z()

	baseHorizontalLines := [][]mgl64.Vec3{}
	baseHorizontalLines = append(baseHorizontalLines,
		[]mgl64.Vec3{aabb.MinVertex, aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0})},
		[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0}), aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd})},
		[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd}), aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd})},
		[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd}), aabb.MinVertex},
	)

	for i := 0; i < 2; i++ {
		for _, b := range baseHorizontalLines {
			allLines = append(allLines,
				[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i) * yd, 0}), b[1].Add(mgl64.Vec3{0, float64(i) * yd, 0})},
			)
		}
	}

	baseVerticalLines := [][]mgl64.Vec3{}
	for i := 0; i < 2; i++ {
		baseVerticalLines = append(baseVerticalLines,
			[]mgl64.Vec3{aabb.MinVertex, aabb.MinVertex.Add(mgl64.Vec3{0, yd, 0})},
			[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0}), aabb.MinVertex.Add(mgl64.Vec3{xd, yd, 0})},
		)
	}

	for i := 0; i < 2; i++ {
		for _, b := range baseVerticalLines {
			allLines = append(allLines,
				[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i) * zd}), b[1].Add(mgl64.Vec3{0, 0, float64(i) * zd})},
			)
		}
	}

	drawLines(
		viewerContext,
		shader,
		allLines,
		thickness,
		color,
	)
}

func createModelMatrix(scaleMatrix, rotationMatrix, translationMatrix mgl64.Mat4) mgl64.Mat4 {
	return translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
}

func resetGLRenderSettings() {
	gl.BindVertexArray(0)
	gl.UseProgram(0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.CullFace(gl.BACK)
	gl.Enable(gl.BLEND)
}
