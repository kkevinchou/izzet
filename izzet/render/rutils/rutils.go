package rutils

import (
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/kitolib/shaders"
)

type TriangleVAO struct {
	VAO    uint32
	length int
}

var lineCache map[string][]mgl64.Vec3
var cubeCache map[string][]mgl64.Vec3
var triangleVAOCache map[string]TriangleVAO
var singleSidedQuadVAO uint32
var pickingBuffer []byte
var runtimeConfig *runtimeconfig.RuntimeConfig
var internedQuadVAOPositionUV uint32
var cubeVAOs map[string]uint32
var ndcQuadVAO uint32

func init() {
	lineCache = map[string][]mgl64.Vec3{}
	cubeCache = map[string][]mgl64.Vec3{}
	triangleVAOCache = map[string]TriangleVAO{}
	cubeVAOs = map[string]uint32{}
}

// global setter for convenience
func SetRuntimeConfig(c *runtimeconfig.RuntimeConfig) {
	runtimeConfig = c
}

func IztDrawElements(count int32) {
	runtimeConfig.TriangleDrawCount += int(count / 3)
	runtimeConfig.DrawCount += 1
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, nil)
}

func IztDrawArrays(first, count int32) {
	runtimeConfig.TriangleDrawCount += int(count / 3)
	runtimeConfig.DrawCount += 1
	gl.DrawArrays(gl.TRIANGLES, first, count)
}

func IztDrawLines(count int32) {
	runtimeConfig.DrawCount += 1
	gl.DrawArrays(gl.LINES, 0, count)
}

func DrawLineGroup(name string, shader *shaders.ShaderProgram, lines [][2]mgl64.Vec3, thickness float64, color mgl64.Vec3) {
	var vao uint32
	var length int

	if _, ok := triangleVAOCache[name]; !ok {
		var points []mgl64.Vec3
		for _, line := range lines {
			start := line[0]
			end := line[1]
			length := end.Sub(start).Len()

			if length == 0 {
				cp := cubePoints(thickness)
				for _, p := range cp {
					points = append(points, p.Add(start))
				}
			} else {
				dir := end.Sub(start).Normalize()
				q := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)

				for _, dp := range linePoints(thickness, length) {
					newEnd := q.Rotate(dp).Add(start)
					points = append(points, newEnd)
				}
			}
		}
		vao, length = generateTrisVAO(points)
		item := TriangleVAO{VAO: vao, length: length}
		triangleVAOCache[name] = item
	}

	item := triangleVAOCache[name]
	vao = item.VAO
	length = item.length

	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	shader.SetUniformFloat("intensity", 1.0)
	gl.BindVertexArray(vao)
	IztDrawArrays(0, int32(length))
}

func cubePoints(thickness float64) []mgl64.Vec3 {
	cacheKey := genCacheKey(thickness, 0)
	if _, ok := cubeCache[cacheKey]; ok {
		return cubeCache[cacheKey]
	}

	var ht float64 = thickness / 2
	points := []mgl64.Vec3{
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

	cubeCache[cacheKey] = points
	return points
}

func linePoints(thickness float64, length float64) []mgl64.Vec3 {
	cacheKey := genCacheKey(thickness, length)
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

func genCacheKey(thickness, length float64) string {
	return fmt.Sprintf("%.3f_%.3f", thickness, length)
}

func generateTrisVAO(points []mgl64.Vec3) (uint32, int) {
	var vertices []float32
	for _, point := range points {
		vertices = append(vertices, float32(point.X()), float32(point.Y()), float32(point.Z()))
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	return vao, len(vertices)
}

func DrawTexturedQuad(viewerContext *context.ViewerContext, shaderManager *shaders.ShaderManager, texture uint32, aspectRatio float32, modelMatrix *mgl32.Mat4, doubleSided bool, pickingID *int) {
	vao := getInternedQuadVAOPositionUV()

	gl.BindVertexArray(vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	if modelMatrix != nil {
		shader := shaderManager.GetShaderProgram("world_space_quad")
		shader.Use()
		if pickingID != nil {
			shader.SetUniformUInt("entityID", uint32(*pickingID))
		}
		shader.SetUniformMat4("model", *modelMatrix)
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	} else {
		shader := shaderManager.GetShaderProgram("screen_space_quad")
		shader.Use()
	}

	// honestly we should clean up this quad drawing logic
	numVertices := 6
	if doubleSided {
		numVertices *= 2
	}

	IztDrawArrays(0, int32(numVertices))
}

func getInternedQuadVAOPositionUV() uint32 {
	if internedQuadVAOPositionUV == 0 {
		var internedQuadVBO uint32
		var internedQuadVertices = []float32{
			-1, -1, 0, 0.0, 0.0,
			1, -1, 0, 1.0, 0.0,
			1, 1, 0, 1.0, 1.0,
			1, 1, 0, 1.0, 1.0,
			-1, 1, 0, 0.0, 1.0,
			-1, -1, 0, 0.0, 0.0,
		}

		var backVertices []float32 = []float32{
			1, 1, 0, 1.0, 1.0,
			1, -1, 0, 1.0, 0.0,
			-1, -1, 0, 0.0, 0.0,
			-1, -1, 0, 0.0, 0.0,
			-1, 1, 0, 0.0, 1.0,
			1, 1, 0, 1.0, 1.0,
		}

		// always add the double sided vertices
		// when a draw request comes in, if doubleSided is false we only draw the first half of the vertices
		// this is wasteful for scenarios where we don't need all vertices
		internedQuadVertices = append(internedQuadVertices, backVertices...)

		// var vbo, dtqVao uint32
		apputils.GenBuffers(1, &internedQuadVBO)
		gl.GenVertexArrays(1, &internedQuadVAOPositionUV)

		gl.BindVertexArray(internedQuadVAOPositionUV)
		gl.BindBuffer(gl.ARRAY_BUFFER, internedQuadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(internedQuadVertices)*4, gl.Ptr(internedQuadVertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		gl.EnableVertexAttribArray(1)
	}

	return internedQuadVAOPositionUV
}

var internedQuadVAOPosition uint32

func GetInternedQuadVAOPosition() uint32 {
	if internedQuadVAOPosition == 0 {
		var internedQuadVBO uint32
		var internedQuadVertices = []float32{
			-1, -1, 0,
			1, -1, 0,
			1, 1, 0,
			1, 1, 0,
			-1, 1, 0,
			-1, -1, 0,
		}

		var backVertices []float32 = []float32{
			1, 1, 0,
			1, -1, 0,
			-1, -1, 0,
			-1, -1, 0,
			-1, 1, 0,
			1, 1, 0,
		}

		// always add the double sided vertices
		// when a draw request comes in, if doubleSided is false we only draw the first half of the vertices
		// this is wasteful for scenarios where we don't need all vertices
		internedQuadVertices = append(internedQuadVertices, backVertices...)

		// var vbo, dtqVao uint32
		apputils.GenBuffers(1, &internedQuadVBO)
		gl.GenVertexArrays(1, &internedQuadVAOPosition)

		gl.BindVertexArray(internedQuadVAOPosition)
		gl.BindBuffer(gl.ARRAY_BUFFER, internedQuadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(internedQuadVertices)*4, gl.Ptr(internedQuadVertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)
	}

	return internedQuadVAOPosition
}

func DrawBillboardTexture(
	texture uint32,
	length float32,
) {
	if singleSidedQuadVAO == 0 {
		vertices := []float32{
			-1 * length, -1 * length, 0, 0.0, 0.0,
			1 * length, -1 * length, 0, 1.0, 0.0,
			1 * length, 1 * length, 0, 1.0, 1.0,
			1 * length, 1 * length, 0, 1.0, 1.0,
			-1 * length, 1 * length, 0, 0.0, 1.0,
			-1 * length, -1 * length, 0, 0.0, 0.0,
		}

		var vbo, vao uint32
		apputils.GenBuffers(1, &vbo)
		gl.GenVertexArrays(1, &vao)

		gl.BindVertexArray(vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		gl.EnableVertexAttribArray(1)

		singleSidedQuadVAO = vao
	}

	gl.BindVertexArray(singleSidedQuadVAO)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	IztDrawArrays(0, 6)
}

func DrawAABB(shader *shaders.ShaderProgram, viewerContext context.ViewerContext, color mgl64.Vec3, aabb collider.BoundingBox, thickness float64) {
	var allLines [][2]mgl64.Vec3

	d := aabb.MaxVertex.Sub(aabb.MinVertex)
	xd := d.X()
	yd := d.Y()
	zd := d.Z()

	baseHorizontalLines := [][2]mgl64.Vec3{}
	baseHorizontalLines = append(baseHorizontalLines,
		[2]mgl64.Vec3{aabb.MinVertex, aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0})},
		[2]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0}), aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd})},
		[2]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd}), aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd})},
		[2]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd}), aabb.MinVertex},
	)

	for i := 0; i < 2; i++ {
		for _, b := range baseHorizontalLines {
			allLines = append(allLines,
				[2]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i) * yd, 0}), b[1].Add(mgl64.Vec3{0, float64(i) * yd, 0})},
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
				[2]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i) * zd}), b[1].Add(mgl64.Vec3{0, 0, float64(i) * zd})},
			)
		}
	}

	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	DrawLineGroup(fmt.Sprintf("aabb_%v_%v", aabb.MinVertex, aabb.MaxVertex), shader, allLines, thickness, color)
}

func GetCubeVAO(length float32, includeNormals bool) uint32 {
	hash := fmt.Sprintf("%.2f_%t", length, includeNormals)
	if _, ok := cubeVAOs[hash]; !ok {
		vao := initCubeVAO(length, includeNormals)
		cubeVAOs[hash] = vao
	}
	return cubeVAOs[hash]
}

func initCubeVAO(length float32, includeNormals bool) uint32 {
	ht := length / 2

	allVertexAttribs := []float32{
		// front
		-ht, -ht, ht, 0, 0, -1,
		ht, -ht, ht, 0, 0, -1,
		ht, ht, ht, 0, 0, -1,

		ht, ht, ht, 0, 0, -1,
		-ht, ht, ht, 0, 0, -1,
		-ht, -ht, ht, 0, 0, -1,

		// back
		ht, ht, -ht, 0, 0, 1,
		ht, -ht, -ht, 0, 0, 1,
		-ht, -ht, -ht, 0, 0, 1,

		-ht, -ht, -ht, 0, 0, 1,
		-ht, ht, -ht, 0, 0, 1,
		ht, ht, -ht, 0, 0, 1,

		// right
		ht, -ht, ht, 1, 0, 0,
		ht, -ht, -ht, 1, 0, 0,
		ht, ht, -ht, 1, 0, 0,

		ht, ht, -ht, 1, 0, 0,
		ht, ht, ht, 1, 0, 0,
		ht, -ht, ht, 1, 0, 0,

		// left
		-ht, ht, -ht, -1, 0, 0,
		-ht, -ht, -ht, -1, 0, 0,
		-ht, -ht, ht, -1, 0, 0,

		-ht, -ht, ht, -1, 0, 0,
		-ht, ht, ht, -1, 0, 0,
		-ht, ht, -ht, -1, 0, 0,

		// top
		ht, ht, ht, 0, 1, 0,
		ht, ht, -ht, 0, 1, 0,
		-ht, ht, ht, 0, 1, 0,

		-ht, ht, ht, 0, 1, 0,
		ht, ht, -ht, 0, 1, 0,
		-ht, ht, -ht, 0, 1, 0,

		// bottom
		-ht, -ht, ht, 0, -1, 0,
		ht, -ht, -ht, 0, -1, 0,
		ht, -ht, ht, 0, -1, 0,

		-ht, -ht, -ht, 0, -1, 0,
		ht, -ht, -ht, 0, -1, 0,
		-ht, -ht, ht, 0, -1, 0,
	}

	var vertices []float32

	totalVertexAttributesSize := 6
	actualVertexAttributesSize := totalVertexAttributesSize
	if !includeNormals {
		actualVertexAttributesSize -= 3
	}

	for i := 0; i < len(allVertexAttribs); i += totalVertexAttributesSize {
		for j := range actualVertexAttributesSize {
			vertices = append(vertices, allVertexAttribs[i+j])
		}
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	ptrOffset := 0
	floatSize := 4

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(actualVertexAttributesSize*floatSize), nil)
	gl.EnableVertexAttribArray(0)

	if includeNormals {
		ptrOffset += 3
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(actualVertexAttributesSize*floatSize), gl.PtrOffset(ptrOffset*floatSize))
		gl.EnableVertexAttribArray(1)
	}

	return vao
}

func GetNDCQuadVAO() uint32 {
	if ndcQuadVAO == 0 {
		vertices := []float32{
			-1, -1, 0.0, 0.0,
			1, -1, 1.0, 0.0,
			1, 1, 1.0, 1.0,
			1, 1, 1.0, 1.0,
			-1, 1, 0.0, 1.0,
			-1, -1, 0.0, 0.0,
		}

		var vbo, vao uint32
		apputils.GenBuffers(1, &vbo)
		gl.GenVertexArrays(1, &vao)

		gl.BindVertexArray(vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
		gl.EnableVertexAttribArray(1)

		ndcQuadVAO = vao
	}

	return ndcQuadVAO
}

// computes the near plane position for a given x y coordinate
func NDCToWorldPosition(viewerContext context.ViewerContext, directionVec mgl64.Vec3) mgl64.Vec3 {
	// ndcP := mgl64.Vec4{((x / float64(g.width)) - 0.5) * 2, ((y / float64(g.height)) - 0.5) * -2, -1, 1}
	nearPlanePos := viewerContext.InverseViewMatrix.Inv().Mul4(viewerContext.ProjectionMatrix.Inv()).Mul4x1(directionVec.Vec4(1))
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func TimeFunc(name string, f func()) {
	mr := globals.ClientRegistry()
	start := time.Now()
	f()
	mr.Inc(name, float64(time.Since(start).Milliseconds()))
}
