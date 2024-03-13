package render

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/app/apputils"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/kitolib/shaders"
)

var (
	navmeshVAOCache map[string]uint32
	numVertices     int32
)

func (r *Renderer) drawCompactHeightField(name string, viewerContext ViewerContext, shader *shaders.ShaderProgram, chf *navmesh.CompactHeightField, distances []int) {
	if _, ok := navmeshVAOCache[name]; !ok {
		var positions []mgl32.Vec3
		var ds []int32
		var rs []int32

		for x := range chf.Width() {
			for z := range chf.Height() {
				cell := chf.Cells()[x+z*chf.Width()]
				spanIndex := cell.SpanIndex
				spanCount := cell.SpanCount

				for i := spanIndex; i < spanIndex+spanCount; i++ {
					span := chf.Spans()[i]
					position := mgl32.Vec3{
						float32(x) + float32(chf.BMin().X()),
						float32(span.Y()) + float32(chf.BMin().Y()),
						float32(z) + float32(chf.BMin().Z()),
					}
					positions = append(positions, position)
					ds = append(ds, int32(distances[i]))
					rs = append(rs, int32(0))
				}
			}
		}

		numVertices = int32(len(positions))
		vao := cubeAttributes(positions, ds, rs)
		navmeshVAOCache[name] = vao
	}

	gl.BindVertexArray(navmeshVAOCache[name])
	r.iztDrawElements(numVertices * 36)
}

func cubeAttributes(positions []mgl32.Vec3, distances []int32, regionIDs []int32) uint32 {
	var ht float32 = 1.0 / 2

	var vertexAttributes []float32

	for _, position := range positions {
		x, y, z := position.X(), position.Y(), position.Z()
		vertexAttributes = append(vertexAttributes, []float32{
			// front
			-ht + x, -ht + y, ht + z, 0, 0, -1,
			ht + x, -ht + y, ht + z, 0, 0, -1,
			ht + x, ht + y, ht + z, 0, 0, -1,

			ht + x, ht + y, ht + z, 0, 0, -1,
			-ht + x, ht + y, ht + z, 0, 0, -1,
			-ht + x, -ht + y, ht + z, 0, 0, -1,

			// back
			ht + x, ht + y, -ht + z, 0, 0, 1,
			ht + x, -ht + y, -ht + z, 0, 0, 1,
			-ht + x, -ht + y, -ht + z, 0, 0, 1,

			-ht + x, -ht + y, -ht + z, 0, 0, 1,
			-ht + x, ht + y, -ht + z, 0, 0, 1,
			ht + x, ht + y, -ht + z, 0, 0, 1,

			// right
			ht + x, -ht + y, ht + z, 1, 0, 0,
			ht + x, -ht + y, -ht + z, 1, 0, 0,
			ht + x, ht + y, -ht + z, 1, 0, 0,

			ht + x, ht + y, -ht + z, 1, 0, 0,
			ht + x, ht + y, ht + z, 1, 0, 0,
			ht + x, -ht + y, ht + z, 1, 0, 0,

			// left
			-ht + x, ht + y, -ht + z, -1, 0, 0,
			-ht + x, -ht + y, -ht + z, -1, 0, 0,
			-ht + x, -ht + y, ht + z, -1, 0, 0,

			-ht + x, -ht + y, ht + z, -1, 0, 0,
			-ht + x, ht + y, ht + z, -1, 0, 0,
			-ht + x, ht + y, -ht + z, -1, 0, 0,

			// top
			ht + x, ht + y, ht + z, 0, 1, 0,
			ht + x, ht + y, -ht + z, 0, 1, 0,
			-ht + x, ht + y, ht + z, 0, 1, 0,

			-ht + x, ht + y, ht + z, 0, 1, 0,
			ht + x, ht + y, -ht + z, 0, 1, 0,
			-ht + x, ht + y, -ht + z, 0, 1, 0,

			// bottom
			-ht + x, -ht + y, ht + z, 0, -1, 0,
			ht + x, -ht + y, -ht + z, 0, -1, 0,
			ht + x, -ht + y, ht + z, 0, -1, 0,

			-ht + x, -ht + y, -ht + z, 0, -1, 0,
			ht + x, -ht + y, -ht + z, 0, -1, 0,
			-ht + x, -ht + y, ht + z, 0, -1, 0,
		}...)
	}

	totalAttributeSize := 6

	// spanAttributes := make([]int32, len(vertexAttributes)/totalAttributeSize*2)
	spanAttributes := make([]int32, (len(distances)+len(regionIDs))*36)
	for i := 0; i < len(distances); i++ {
		for j := 0; j < 36; j++ {
			spanAttributes[i+j] = 500
			spanAttributes[i+j+1] = regionIDs[i]
		}
	}

	distRegionIndex := 0
	for i := 0; i < len(spanAttributes); i += 72 {
		for j := 0; j < 72; j += 2 {
			spanAttributes[i+j] = distances[distRegionIndex]
			spanAttributes[i+j+1] = regionIDs[distRegionIndex]
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