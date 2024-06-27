package assets

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

var nextGlobalID int

func cubeMesh(length int) *modelspec.MeshSpecification {
	primitive := createCubePrimitive(length)
	mesh := &modelspec.MeshSpecification{ID: nextGlobalID, Primitives: []*modelspec.PrimitiveSpecification{primitive}}
	nextGlobalID += 1
	return mesh
}

func createCubePrimitive(length int) *modelspec.PrimitiveSpecification {
	vertFloats := cubeVertexFloatsByLength(length)

	vertexIndices := []uint32{}
	for i := 0; i < 36; i++ {
		vertexIndices = append(vertexIndices, uint32(i))
	}

	uniqueVertices := []modelspec.Vertex{}
	for i := 0; i < len(vertFloats); i += 8 {
		x := vertFloats[i]
		y := vertFloats[i+1]
		z := vertFloats[i+2]

		nx := vertFloats[i+3]
		ny := vertFloats[i+4]
		nz := vertFloats[i+5]

		u := vertFloats[i+6]
		v := vertFloats[i+7]

		uniqueVertices = append(uniqueVertices, modelspec.Vertex{
			Position:       mgl32.Vec3{x, y, z},
			Normal:         mgl32.Vec3{nx, ny, nz},
			Texture0Coords: mgl32.Vec2{u, v},
			Texture1Coords: mgl32.Vec2{},
		})
	}

	var vertices []modelspec.Vertex
	for _, i := range vertexIndices {
		vertices = append(vertices, uniqueVertices[i])
	}

	return &modelspec.PrimitiveSpecification{
		VertexIndices:  vertexIndices,
		UniqueVertices: uniqueVertices,
		Vertices:       vertices,
	}
}

func cubeVertexFloatsByLength(length int) []float32 {
	ht := float32(length) / 2

	return []float32{
		// front
		-ht, -ht, ht, 0, 0, 1, 0, 0,
		ht, -ht, ht, 0, 0, 1, 1, 0,
		ht, ht, ht, 0, 0, 1, 1, 1,

		ht, ht, ht, 0, 0, 1, 1, 1,
		-ht, ht, ht, 0, 0, 1, 0, 1,
		-ht, -ht, ht, 0, 0, 1, 0, 0,

		// back
		ht, ht, -ht, 0, 0, -1, 0, 1,
		ht, -ht, -ht, 0, 0, -1, 0, 0,
		-ht, -ht, -ht, 0, 0, -1, 1, 0,

		-ht, -ht, -ht, 0, 0, -1, 1, 0,
		-ht, ht, -ht, 0, 0, -1, 1, 1,
		ht, ht, -ht, 0, 0, -1, 0, 1,

		// right
		ht, -ht, ht, 1, 0, 0, 0, 0,
		ht, -ht, -ht, 1, 0, 0, 1, 0,
		ht, ht, -ht, 1, 0, 0, 1, 1,

		ht, ht, -ht, 1, 0, 0, 1, 1,
		ht, ht, ht, 1, 0, 0, 0, 1,
		ht, -ht, ht, 1, 0, 0, 0, 0,

		// left
		-ht, ht, -ht, -1, 0, 0, 0, 1,
		-ht, -ht, -ht, -1, 0, 0, 0, 0,
		-ht, -ht, ht, -1, 0, 0, 1, 0,

		-ht, -ht, ht, -1, 0, 0, 1, 0,
		-ht, ht, ht, -1, 0, 0, 1, 1,
		-ht, ht, -ht, -1, 0, 0, 0, 1,

		// top
		ht, ht, ht, 0, 1, 0, 1, 0,
		ht, ht, -ht, 0, 1, 0, 1, 1,
		-ht, ht, ht, 0, 1, 0, 0, 0,

		-ht, ht, ht, 0, 1, 0, 0, 0,
		ht, ht, -ht, 0, 1, 0, 1, 1,
		-ht, ht, -ht, 0, 1, 0, 0, 1,

		// bottom
		-ht, -ht, ht, 0, -1, 0, 0, 1,
		ht, -ht, -ht, 0, -1, 0, 1, 0,
		ht, -ht, ht, 0, -1, 0, 1, 1,

		-ht, -ht, -ht, 0, -1, 0, 0, 0,
		ht, -ht, -ht, 0, -1, 0, 1, 0,
		-ht, -ht, ht, 0, -1, 0, 0, 1,
	}
}
