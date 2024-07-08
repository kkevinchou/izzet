package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

var (
	navmeshVAOCache            uint32
	navmeshVAOCacheVertexCount int32

	voxelVAOCache            uint32
	voxelVAOCacheVertexCount int32

	simplifiedContourVAOCache    uint32
	simplifiedContourVertexCount int32
)

func (r *Renderer) drawNavmesh(shaderManager *shaders.ShaderManager, viewerContext ViewerContext, nm *navmesh.NavigationMesh) {
	if nm.Invalidated {
		voxelVAOCache, voxelVAOCacheVertexCount = createVoxelVAO(nm.HeightField)
		navmeshVAOCache, navmeshVAOCacheVertexCount = createCompactHeightFieldVAO(nm.CompactHeightField, nm.BlurredDistances)
		simplifiedContourVAOCache, simplifiedContourVertexCount = createSimplifiedContourVAO(nm)
	}

	if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionCompactHeightField {
		gl.BindVertexArray(navmeshVAOCache)
		r.iztDrawElements(navmeshVAOCacheVertexCount * 36)
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionVoxel {
		gl.BindVertexArray(voxelVAOCache)
		r.iztDrawElements(voxelVAOCacheVertexCount * 36)
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionSimplifiedContour {
		gl.LineWidth(1)
		shader := shaderManager.GetShaderProgram("flat")
		// color := mgl64.Vec3{252.0 / 255, 241.0 / 255, 33.0 / 255}
		color := mgl64.Vec3{1, 0, 0}
		shader.Use()
		shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
		shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
		gl.BindVertexArray(simplifiedContourVAOCache)
		// r.iztDrawLineStrip(simplifiedContourVertexCount)
		r.iztDrawLines(simplifiedContourVertexCount)
	} else {
		panic("WAT")
	}
}

func createSimplifiedContourVAO(nm *navmesh.NavigationMesh) (uint32, int32) {
	contourSet := nm.ContourSet
	minVertex := nm.Volume.MinVertex

	// var vertices []float32
	// for _, contour := range contourSet.Contours {
	// 	for _, vert := range contour.Verts {
	// 		vertices = append(vertices, float32(vert.X+int(minVertex.X())), float32(vert.Y+int(minVertex.Y())), float32(vert.Z+int(minVertex.Z())))
	// 	}
	// }
	var vertices []float32
	for _, contour := range contourSet.Contours {
		for i, _ := range contour.Verts {
			v0 := contour.Verts[i]
			v1 := contour.Verts[(i+1)%len(contour.Verts)]
			vertices = append(vertices, float32(v0.X+int(minVertex.X())), float32(v0.Y+int(minVertex.Y())), float32(v0.Z+int(minVertex.Z())))
			vertices = append(vertices, float32(v1.X+int(minVertex.X())), float32(v1.Y+int(minVertex.Y())), float32(v1.Z+int(minVertex.Z())))
		}
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	return vao, int32(len(vertices))
}

func createCompactHeightFieldVAO(chf *navmesh.CompactHeightField, distances []int) (uint32, int32) {
	var positions []mgl32.Vec3
	var lengths []float32
	var ds []int32
	var regionIDs []int32

	for x := range chf.Width() {
		for z := range chf.Height() {
			cell := chf.Cells()[x+z*chf.Width()]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+navmesh.SpanIndex(spanCount); i++ {
				span := chf.Spans()[i]
				position := mgl32.Vec3{
					float32(x) + float32(chf.BMin().X()),
					float32(span.Y()) + float32(chf.BMin().Y()),
					float32(z) + float32(chf.BMin().Z()),
				}
				positions = append(positions, position)
				lengths = append(lengths, 1)
				ds = append(ds, int32(distances[i]))
				regionIDs = append(regionIDs, int32(span.RegionID()))
			}
		}
	}

	vao := cubeAttributes(positions, lengths, ds, regionIDs)

	return vao, int32(len(positions))
}

func createVoxelVAO(hf *navmesh.HeightField) (uint32, int32) {
	var positions []mgl32.Vec3
	var lengths []float32
	var ds []int32
	var regionIDs []int32

	for z := range hf.Height {
		for x := range hf.Width {
			index := x + z*hf.Width
			span := hf.Spans[index]
			for span != nil {
				position := mgl32.Vec3{
					float32(x) + float32(hf.BMin.X()),
					float32(span.Min) + float32(hf.BMin.Y()),
					float32(z) + float32(hf.BMin.Z()),
				}
				positions = append(positions, position)
				lengths = append(lengths, float32(span.Max-span.Min+1))
				ds = append(ds, 0)
				regionIDs = append(regionIDs, 0)

				span = span.Next
			}
		}
	}
	vao := cubeAttributes(positions, lengths, ds, regionIDs)
	return vao, int32(len(positions))
}

func cubeAttributes(positions []mgl32.Vec3, lengths []float32, distances []int32, regionRenderIDs []int32) uint32 {
	var vertexAttributes []float32

	for i, position := range positions {
		x, y, z := position.X(), position.Y(), position.Z()
		vertexAttributes = append(vertexAttributes, []float32{
			// front
			x, y, z, 0, 0, 1,
			1 + x, y, z, 0, 0, 1,
			1 + x, lengths[i] + y, z, 0, 0, 1,

			1 + x, lengths[i] + y, z, 0, 0, 1,
			x, lengths[i] + y, z, 0, 0, 1,
			x, y, z, 0, 0, 1,

			// back
			1 + x, lengths[i] + y, z - 1, 0, 0, -1,
			1 + x, y, z - 1, 0, 0, -1,
			x, y, z - 1, 0, 0, -1,

			x, y, z - 1, 0, 0, -1,
			x, lengths[i] + y, z - 1, 0, 0, -1,
			1 + x, lengths[i] + y, z - 1, 0, 0, -1,

			// rig1
			1 + x, y, z, 1, 0, 0,
			1 + x, y, z - 1, 1, 0, 0,
			1 + x, lengths[i] + y, z - 1, 1, 0, 0,

			1 + x, lengths[i] + y, z - 1, 1, 0, 0,
			1 + x, lengths[i] + y, z, 1, 0, 0,
			1 + x, y, z, 1, 0, 0,

			// left
			x, lengths[i] + y, z - 1, -1, 0, 0,
			x, y, z - 1, -1, 0, 0,
			x, y, z, -1, 0, 0,

			x, y, z, -1, 0, 0,
			x, lengths[i] + y, z, -1, 0, 0,
			x, lengths[i] + y, z - 1, -1, 0, 0,

			// top
			1 + x, lengths[i] + y, z, 0, 1, 0,
			1 + x, lengths[i] + y, z - 1, 0, 1, 0,
			x, lengths[i] + y, z, 0, 1, 0,

			x, lengths[i] + y, z, 0, 1, 0,
			1 + x, lengths[i] + y, z - 1, 0, 1, 0,
			x, lengths[i] + y, z - 1, 0, 1, 0,

			// bottom
			x, y, z, 0, -1, 0,
			1 + x, y, z - 1, 0, -1, 0,
			1 + x, y, z, 0, -1, 0,

			x, y, z - 1, 0, -1, 0,
			1 + x, y, z - 1, 0, -1, 0,
			x, y, z, 0, -1, 0,
		}...)
	}

	totalAttributeSize := 6

	// spanAttributes := make([]int32, len(vertexAttributes)/totalAttributeSize*2)
	spanAttributes := make([]int32, (len(distances)+len(regionRenderIDs))*36)
	for i := 0; i < len(distances); i++ {
		for j := 0; j < 36; j++ {
			spanAttributes[i+j] = 500
			spanAttributes[i+j+1] = regionRenderIDs[i]
		}
	}

	distRegionIndex := 0
	for i := 0; i < len(spanAttributes); i += 72 {
		for j := 0; j < 72; j += 2 {
			spanAttributes[i+j] = distances[distRegionIndex]
			spanAttributes[i+j+1] = regionRenderIDs[distRegionIndex]
		}
		distRegionIndex++
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// lay out the position and normal in a VBO
	var vbo uint32
	apputils.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	ptrOffset := 0
	var floatSize int32 = 4
	var intSize int32 = 4
	spanPtrOffset := 0
	totalSpanAttributeSize := 2

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, nil)
	gl.EnableVertexAttribArray(0)

	ptrOffset += 3

	// normal
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, gl.PtrOffset(ptrOffset*int(floatSize)))
	gl.EnableVertexAttribArray(1)

	var vboSpan uint32
	apputils.GenBuffers(1, &vboSpan)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboSpan)
	gl.BufferData(gl.ARRAY_BUFFER, len(spanAttributes)*4, gl.Ptr(spanAttributes), gl.STATIC_DRAW)

	// distance
	gl.VertexAttribIPointer(2, 1, gl.INT, int32(totalSpanAttributeSize)*intSize, nil)
	gl.EnableVertexAttribArray(2)

	spanPtrOffset += 1

	// regionID
	gl.VertexAttribIPointer(3, 1, gl.INT, int32(totalSpanAttributeSize)*intSize, gl.PtrOffset(spanPtrOffset*int(intSize)))
	gl.EnableVertexAttribArray(3)

	vertexIndices := make([]uint32, len(vertexAttributes)/totalAttributeSize)
	for i := range len(vertexIndices) {
		vertexIndices[i] = uint32(i)
	}

	// set up the EBO, each triplet of indices point to three vertices
	// that form a triangle.
	var ebo uint32
	apputils.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(vertexIndices)*4, gl.Ptr(vertexIndices), gl.STATIC_DRAW)

	return vao
}
