package render

import (
	"fmt"
	"math"
	"time"

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

	rawContourVAOCache    uint32
	rawContourVertexCount int32

	distanceFieldVAOCache    uint32
	distanceFieldVertexCount int32

	premergeTrianglesVAOCache    uint32
	premergeTrianglesVertexCount int32

	polygonsVAOCache    uint32
	polygonsVertexCount int32

	detailedMeshVAOCache           uint32
	detailedMeshVertexCount        int32
	detailedMeshSamplesVAOCache    uint32
	detailedMeshSamplesVertexCount int32
)

func (r *Renderer) drawNavmesh(shaderManager *shaders.ShaderManager, viewerContext ViewerContext, nm *navmesh.NavigationMesh) {
	if nm.Invalidated {
		start := time.Now()
		voxelVAOCache, voxelVAOCacheVertexCount = createVoxelVAO(nm.HeightField)
		fmt.Printf("%.1f seconds to create voxel vao\n", time.Since(start).Seconds())
		start = time.Now()
		navmeshVAOCache, navmeshVAOCacheVertexCount = createCompactHeightFieldVAO(nm.CompactHeightField)
		fmt.Printf("%.1f seconds to create chf vao\n", time.Since(start).Seconds())
		start = time.Now()
		distanceFieldVAOCache, distanceFieldVertexCount = createDistanceFieldVAO(nm.CompactHeightField)
		fmt.Printf("%.1f seconds to create distance field vao\n", time.Since(start).Seconds())
		start = time.Now()
		rawContourVAOCache, rawContourVertexCount = createContourVAO(nm, false)
		fmt.Printf("%.1f seconds to create contour vao\n", time.Since(start).Seconds())
		start = time.Now()
		simplifiedContourVAOCache, simplifiedContourVertexCount = createContourVAO(nm, true)
		fmt.Printf("%.1f seconds to create simplified contour vao\n", time.Since(start).Seconds())
		start = time.Now()
		premergeTrianglesVAOCache, premergeTrianglesVertexCount = createPremergeTriangleVAO(nm)
		fmt.Printf("%.1f seconds to create premerge triangle vao\n", time.Since(start).Seconds())
		start = time.Now()
		polygonsVAOCache, polygonsVertexCount = r.createPolygonsVAO(nm)
		fmt.Printf("%.1f seconds to create polygon vao\n", time.Since(start).Seconds())
		start = time.Now()
		detailedMeshVAOCache, detailedMeshVertexCount = createDetailedMeshVAO(nm)
		detailedMeshSamplesVAOCache, detailedMeshSamplesVertexCount = createDetailedMeshSamplesVAO(nm)
		fmt.Printf("%.1f seconds to create detailed mesh vao\n", time.Since(start).Seconds())
	}

	if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionCompactHeightField {
		if navmeshVAOCacheVertexCount > 0 {
			gl.BindVertexArray(navmeshVAOCache)
			r.iztDrawElements(navmeshVAOCacheVertexCount * 36)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionDistanceField {
		if distanceFieldVertexCount > 0 {
			gl.BindVertexArray(distanceFieldVAOCache)
			r.iztDrawElements(distanceFieldVertexCount * 36)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionVoxel {
		if voxelVAOCacheVertexCount > 0 {
			gl.BindVertexArray(voxelVAOCache)
			r.iztDrawElements(voxelVAOCacheVertexCount * 36)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionRawContour {
		if rawContourVertexCount > 0 {
			r.drawContour(shaderManager, viewerContext, rawContourVAOCache, rawContourVertexCount)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionSimplifiedContour {
		if simplifiedContourVertexCount > 0 {
			r.drawContour(shaderManager, viewerContext, simplifiedContourVAOCache, simplifiedContourVertexCount)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionPremergeTriangles {
		if premergeTrianglesVertexCount > 0 {
			r.drawContour(shaderManager, viewerContext, premergeTrianglesVAOCache, premergeTrianglesVertexCount)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionPolygons {
		if polygonsVertexCount > 0 {
			r.drawContour(shaderManager, viewerContext, polygonsVAOCache, polygonsVertexCount)
		}
	} else if panels.SelectedNavmeshRenderComboOption == panels.ComboOptionDetailedMesh {
		if detailedMeshVertexCount > 0 {
			gl.BindVertexArray(detailedMeshVAOCache)
			r.iztDrawElements(detailedMeshVertexCount * 3)
			gl.BindVertexArray(detailedMeshSamplesVAOCache)
			r.iztDrawElements(detailedMeshSamplesVertexCount * 36)
		}
	} else {
		panic("WAT")
	}
}

func (r *Renderer) drawContour(shaderManager *shaders.ShaderManager, viewerContext ViewerContext, vao uint32, count int32) {
	shader := shaderManager.GetShaderProgram("line")
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	gl.BindVertexArray(vao)
	r.iztDrawLines(count)
}

func createPremergeTriangleVAO(nm *navmesh.NavigationMesh) (uint32, int32) {
	minVertex := nm.Volume.MinVertex
	iMinVertex := []int{int(minVertex.X()), int(minVertex.Y()), int(minVertex.Z())}

	var vertexAttributes []float32
	for _, tri := range nm.Mesh.PremergeTriangles {
		for i := range len(tri.Verts) {
			v0 := nm.Mesh.Vertices[tri.Verts[i]]
			v1 := nm.Mesh.Vertices[tri.Verts[(i+1)%len(tri.Verts)]]

			// v0
			vertexAttributes = append(vertexAttributes, float32(v0.X+iMinVertex[0]), float32(v0.Y+iMinVertex[1]), float32(v0.Z+iMinVertex[2]))
			vertexAttributes = append(vertexAttributes, regionIDToColor(tri.RegionID)...)

			// v1
			vertexAttributes = append(vertexAttributes, float32(v1.X+iMinVertex[0]), float32(v1.Y+iMinVertex[1]), float32(v1.Z+iMinVertex[2]))
			vertexAttributes = append(vertexAttributes, regionIDToColor(tri.RegionID)...)
		}
	}
	return createLineVAO(vertexAttributes)
}

func (r *Renderer) createPolygonsVAO(nm *navmesh.NavigationMesh) (uint32, int32) {
	minVertex := nm.Volume.MinVertex
	iMinVertex := []int{int(minVertex.X()), int(minVertex.Y()), int(minVertex.Z())}

	debugMap := r.app.RuntimeConfig().DebugBlob1IntMap

	var vertexAttributes []float32
	for id, poly := range nm.Mesh.Polygons {
		if !debugMap[id] && len(debugMap) > 0 {
			continue
		}
		for i := range len(poly.Verts) {
			v0 := nm.Mesh.Vertices[poly.Verts[i]]
			v1 := nm.Mesh.Vertices[poly.Verts[(i+1)%len(poly.Verts)]]

			// v0
			vertexAttributes = append(vertexAttributes, float32(v0.X+iMinVertex[0]), float32(v0.Y+iMinVertex[1]), float32(v0.Z+iMinVertex[2]))
			vertexAttributes = append(vertexAttributes, regionIDToColor(poly.RegionID)...)

			// v1
			vertexAttributes = append(vertexAttributes, float32(v1.X+iMinVertex[0]), float32(v1.Y+iMinVertex[1]), float32(v1.Z+iMinVertex[2]))
			vertexAttributes = append(vertexAttributes, regionIDToColor(poly.RegionID)...)
		}
	}
	return createLineVAO(vertexAttributes)
}

func createDetailedMeshVAO(nm *navmesh.NavigationMesh) (uint32, int32) {
	if nm.DetailedMesh == nil || len(nm.DetailedMesh.PolyTriangles) == 0 {
		return 0, 0
	}

	var triangles [][3]mgl32.Vec3
	var colors []float32

	for j := range len(nm.DetailedMesh.PolyTriangles) {
		for _, tri := range nm.DetailedMesh.PolyTriangles[j] {
			v0 := nm.DetailedMesh.PolyVertices[j][tri.A]
			v1 := nm.DetailedMesh.PolyVertices[j][tri.B]
			v2 := nm.DetailedMesh.PolyVertices[j][tri.C]

			triangles = append(triangles, [3]mgl32.Vec3{
				{float32(v0.X), float32(v0.Y), float32(v0.Z)},
				{float32(v1.X), float32(v1.Y), float32(v1.Z)},
				{float32(v2.X), float32(v2.Y), float32(v2.Z)},
			})
			colors = append(colors, regionIDToColor(len(triangles))...)
		}
	}

	vao := triangleAttributes(triangles, colors)

	return vao, int32(len(triangles))
}

func createDetailedMeshSamplesVAO(nm *navmesh.NavigationMesh) (uint32, int32) {
	var positions []mgl32.Vec3
	var colors []float32
	var lengths []float32

	samples := nm.DetailedMesh.Samples
	chf := nm.CompactHeightField

	for p := range len(samples) {
		for i := 0; i < len(samples[p]); i += 3 {
			position := mgl32.Vec3{
				float32(samples[p][i]) + float32(chf.BMin().X()),
				float32(samples[p][i+1]) + float32(chf.BMin().Y()),
				float32(samples[p][i+2]) + float32(chf.BMin().Z()),
			}
			positions = append(positions, position)
			colors = append(colors, regionIDToColor(p)...)
			lengths = append(lengths, 1)
		}
	}

	if len(positions) == 0 {
		return 0, 0
	}

	vao := cubeAttributes(positions, lengths, colors)

	return vao, int32(len(positions))
}

func createContourVAO(nm *navmesh.NavigationMesh, simplified bool) (uint32, int32) {
	contourSet := nm.ContourSet
	minVertex := nm.Volume.MinVertex

	var vertexAttributes []float32
	if simplified {
		for _, contour := range contourSet.Contours {
			for i, _ := range contour.Verts {
				ni := (i + 1) % len(contour.Verts)
				v0 := contour.Verts[i]
				v1 := contour.Verts[ni]

				color := regionIDToColor(contour.RegionID)

				// v0
				vertexAttributes = append(vertexAttributes, float32(float64(v0.X)+minVertex.X()), float32(float64(v0.Y)+minVertex.Y()), float32(float64(v0.Z)+minVertex.Z()))
				vertexAttributes = append(vertexAttributes, color...)

				// v1
				vertexAttributes = append(vertexAttributes, float32(float64(v1.X)+minVertex.X()), float32(float64(v1.Y)+minVertex.Y()), float32(float64(v1.Z)+minVertex.Z()))
				vertexAttributes = append(vertexAttributes, color...)
			}
		}
	} else {
		for _, contour := range contourSet.Contours {
			for i, _ := range contour.RawVerts {
				v0 := contour.RawVerts[i]
				v1 := contour.RawVerts[(i+1)%len(contour.RawVerts)]

				color := regionIDToColor(contour.RegionID)

				// v0
				vertexAttributes = append(vertexAttributes, float32(float64(v0.X)+minVertex.X()), float32(float64(v0.Y)+minVertex.Y()), float32(float64(v0.Z)+minVertex.Z()))
				vertexAttributes = append(vertexAttributes, color...)

				// v1
				vertexAttributes = append(vertexAttributes, float32(float64(v1.X)+minVertex.X()), float32(float64(v1.Y)+minVertex.Y()), float32(float64(v1.Z)+minVertex.Z()))
				vertexAttributes = append(vertexAttributes, color...)
			}
		}
	}
	return createLineVAO(vertexAttributes)
}

func createLineVAO(vertexAttributes []float32) (uint32, int32) {
	if len(vertexAttributes) == 0 {
		return 0, 0
	}
	var floatSize int32 = 4
	ptrOffset := 0
	var totalAttributeSize int32 = 6

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*int(floatSize), gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, totalAttributeSize*floatSize, gl.PtrOffset(ptrOffset))
	gl.EnableVertexAttribArray(0)

	ptrOffset += 3

	// color
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, totalAttributeSize*floatSize, gl.PtrOffset(ptrOffset*int(floatSize)))
	gl.EnableVertexAttribArray(1)

	return vao, int32(len(vertexAttributes) / int(totalAttributeSize))
}

func createDistanceFieldVAO(chf *navmesh.CompactHeightField) (uint32, int32) {
	var positions []mgl32.Vec3
	var colors []float32
	var lengths []float32

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
				colors = append(colors, distanceToColor(float32(chf.Distances[i]))...)
				lengths = append(lengths, 1)
			}
		}
	}

	if len(positions) == 0 {
		return 0, 0
	}

	vao := cubeAttributes(positions, lengths, colors)

	return vao, int32(len(positions))
}

func createCompactHeightFieldVAO(chf *navmesh.CompactHeightField) (uint32, int32) {
	var positions []mgl32.Vec3
	var colors []float32
	var lengths []float32

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
				colors = append(colors, regionIDToColor(span.RegionID())...)
				lengths = append(lengths, 1)
			}
		}
	}

	if len(positions) == 0 {
		return 0, 0
	}

	vao := cubeAttributes(positions, lengths, colors)

	return vao, int32(len(positions))
}

func debugCheck(x, z int) bool {
	for i := 0; i < len(navmesh.DBG); i += 2 {
		if navmesh.DBG[i] == x && navmesh.DBG[i+1] == z {
			return true
		}
	}

	// if x == 686 && z == 102 {
	// 	return true
	// } else if x == 696 && z == 101 {
	// 	return true
	// } else if x == 720 && z == 98 {
	// 	return true
	// } else if x == 754 && z == 94 {
	// 	return true
	// } else if x == 812 && z == 88 {
	// 	return true
	// } else if x == 688 && z == 235 {
	// 	return true
	// }
	return false
}

func createVoxelVAO(hf *navmesh.HeightField) (uint32, int32) {
	var positions []mgl32.Vec3
	var colors []float32
	var lengths []float32

	for z := range hf.Height {
		for x := range hf.Width {
			if x == 809 && z == 248 {
				position := mgl32.Vec3{
					float32(x) + float32(hf.BMin.X()),
					float32(hf.BMin.Y()),
					float32(z) + float32(hf.BMin.Z()),
				}
				positions = append(positions, position)
				colors = append(colors, 1, 0, 0)
				lengths = append(lengths, float32(hf.BMax.Y()-hf.BMin.Y()))
			} else if debugCheck(x, z) {
				position := mgl32.Vec3{
					float32(x) + float32(hf.BMin.X()),
					float32(hf.BMin.Y()),
					float32(z) + float32(hf.BMin.Z()),
				}
				positions = append(positions, position)
				colors = append(colors, 0, 1, 0)
				lengths = append(lengths, float32(hf.BMax.Y()-hf.BMin.Y()))
			} else {
				index := x + z*hf.Width
				span := hf.Spans[index]
				for span != nil {
					position := mgl32.Vec3{
						float32(x) + float32(hf.BMin.X()),
						float32(span.Min) + float32(hf.BMin.Y()),
						float32(z) + float32(hf.BMin.Z()),
					}
					positions = append(positions, position)
					color := []float32{.9, .9, .9}
					// if x == 809 || x == 796 || x == 764 || x == 617 || x == 653 || x == 808 {
					// 	color = []float32{1, 0, 0}
					// }
					// if z == 248 || z == 244 || z == 194 || z == 240 || z == 331 || z == 379 {
					// 	color = []float32{1, 0, 0}
					// }
					colors = append(colors, color...)
					lengths = append(lengths, float32(span.Max-span.Min+1))

					span = span.Next
				}
			}
		}
	}

	if len(positions) == 0 {
		return 0, 0
	}

	vao := cubeAttributes(positions, lengths, colors)
	return vao, int32(len(positions))
}

func triangleAttributes(triangles [][3]mgl32.Vec3, colors []float32) uint32 {
	var vertexAttributes []float32

	for i := range len(triangles) {
		t := &triangles[i]
		r, g, b := colors[i*3], colors[i*3+1], colors[i*3+2]

		v0 := t[0]
		v1 := t[1]
		v2 := t[2]

		normal := v1.Sub(v0).Cross(v2.Sub(v0)).Normalize()

		vertexAttributes = append(vertexAttributes, []float32{
			v0.X(), v0.Y(), v0.Z(), normal.X(), normal.Y(), normal.Z(), r, g, b,
			v1.X(), v1.Y(), v1.Z(), normal.X(), normal.Y(), normal.Z(), r, g, b,
			v2.X(), v2.Y(), v2.Z(), normal.X(), normal.Y(), normal.Z(), r, g, b,
		}...)
	}

	totalAttributeSize := 9

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// lay out the position, normal, and color in a VBO
	var vbo uint32
	apputils.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	ptrOffset := 0
	var floatSize int32 = 4

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, nil)
	gl.EnableVertexAttribArray(0)

	ptrOffset += 3

	// normal
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, gl.PtrOffset(ptrOffset*int(floatSize)))
	gl.EnableVertexAttribArray(1)

	ptrOffset += 3

	// color
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, gl.PtrOffset(ptrOffset*int(floatSize)))
	gl.EnableVertexAttribArray(2)

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

func cubeAttributes(positions []mgl32.Vec3, lengths []float32, colors []float32) uint32 {
	var vertexAttributes []float32

	for i := range len(positions) {
		position := positions[i]
		x, y, z := position.X(), position.Y(), position.Z()
		r, g, b := colors[i*3], colors[i*3+1], colors[i*3+2]

		vertexAttributes = append(vertexAttributes, []float32{
			// front
			x, y, z + 1, 0, 0, 1, r, g, b,
			1 + x, y, z + 1, 0, 0, 1, r, g, b,
			1 + x, lengths[i] + y, z + 1, 0, 0, 1, r, g, b,

			1 + x, lengths[i] + y, z + 1, 0, 0, 1, r, g, b,
			x, lengths[i] + y, z + 1, 0, 0, 1, r, g, b,
			x, y, z + 1, 0, 0, 1, r, g, b,

			// back
			1 + x, lengths[i] + y, z, 0, 0, -1, r, g, b,
			1 + x, y, z, 0, 0, -1, r, g, b,
			x, y, z, 0, 0, -1, r, g, b,

			x, y, z, 0, 0, -1, r, g, b,
			x, lengths[i] + y, z, 0, 0, -1, r, g, b,
			1 + x, lengths[i] + y, z, 0, 0, -1, r, g, b,

			// rig1
			1 + x, y, z + 1, 1, 0, 0, r, g, b,
			1 + x, y, z, 1, 0, 0, r, g, b,
			1 + x, lengths[i] + y, z, 1, 0, 0, r, g, b,

			1 + x, lengths[i] + y, z, 1, 0, 0, r, g, b,
			1 + x, lengths[i] + y, z + 1, 1, 0, 0, r, g, b,
			1 + x, y, z + 1, 1, 0, 0, r, g, b,

			// left
			x, lengths[i] + y, z, -1, 0, 0, r, g, b,
			x, y, z, -1, 0, 0, r, g, b,
			x, y, z + 1, -1, 0, 0, r, g, b,

			x, y, z + 1, -1, 0, 0, r, g, b,
			x, lengths[i] + y, z + 1, -1, 0, 0, r, g, b,
			x, lengths[i] + y, z, -1, 0, 0, r, g, b,

			// top
			1 + x, lengths[i] + y, z + 1, 0, 1, 0, r, g, b,
			1 + x, lengths[i] + y, z, 0, 1, 0, r, g, b,
			x, lengths[i] + y, z + 1, 0, 1, 0, r, g, b,

			x, lengths[i] + y, z + 1, 0, 1, 0, r, g, b,
			1 + x, lengths[i] + y, z, 0, 1, 0, r, g, b,
			x, lengths[i] + y, z, 0, 1, 0, r, g, b,

			// bottom
			x, y, z + 1, 0, -1, 0, r, g, b,
			1 + x, y, z, 0, -1, 0, r, g, b,
			1 + x, y, z + 1, 0, -1, 0, r, g, b,

			x, y, z, 0, -1, 0, r, g, b,
			1 + x, y, z, 0, -1, 0, r, g, b,
			x, y, z + 1, 0, -1, 0, r, g, b,
		}...)
	}

	totalAttributeSize := 9

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// lay out the position, normal, and color in a VBO
	var vbo uint32
	apputils.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	ptrOffset := 0
	var floatSize int32 = 4

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, nil)
	gl.EnableVertexAttribArray(0)

	ptrOffset += 3

	// normal
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, gl.PtrOffset(ptrOffset*int(floatSize)))
	gl.EnableVertexAttribArray(1)

	ptrOffset += 3

	// color
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, int32(totalAttributeSize)*floatSize, gl.PtrOffset(ptrOffset*int(floatSize)))
	gl.EnableVertexAttribArray(2)

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

func distanceToColor(distance float32) []float32 {
	c := distance / 100
	return []float32{c, c, c}
}

// regionIDToColor converts a regionID into an rgb color
func regionIDToColor(regionID int) []float32 {
	if regionID == 0 {
		return []float32{.2, .2, .2}
	}

	hue := float32(regionID) * 137.508           // 137.508 is the golden angle in degrees
	hue = float32(math.Mod(float64(hue), 360.0)) // Ensure the hue is within [0, 360)
	return HSLToRGB(hue, 1, 0.2)
}

// HSLToRGB converts an HSL color value to RGB.
func HSLToRGB(h, s, l float32) []float32 {
	c := (1.0 - float32(math.Abs(float64(2.0*l-1.0)))) * s
	x := c * (1.0 - float32(math.Abs(math.Mod(float64(h/60.0), 2.0)-1.0)))
	m := l - c/2.0

	var r, g, b float32
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return []float32{(r + m), (g + m), (b + m)}
}
