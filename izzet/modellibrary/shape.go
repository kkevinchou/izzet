package modellibrary

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

const CubeMeshID int = 18

func cube() *modelspec.MeshSpecification {
	primitive := createCubePrimitive()
	return &modelspec.MeshSpecification{ID: CubeMeshID, Primitives: []*modelspec.PrimitiveSpecification{primitive}}
}

func createCubePrimitive() *modelspec.PrimitiveSpecification {
	vertFloats := cubeVertexFloatsByLength(50)

	vertexIndices := []uint32{}
	for i := 0; i < 36; i++ {
		vertexIndices = append(vertexIndices, uint32(i))
	}

	uniqueVertices := []modelspec.Vertex{}
	for i := 0; i < len(vertFloats); i += 6 {
		x := vertFloats[i]
		y := vertFloats[i+1]
		z := vertFloats[i+2]

		nx := vertFloats[i+3]
		ny := vertFloats[i+4]
		nz := vertFloats[i+5]

		uniqueVertices = append(uniqueVertices, modelspec.Vertex{
			Position:       mgl32.Vec3{x, y, z},
			Normal:         mgl32.Vec3{nx, ny, nz},
			Texture0Coords: mgl32.Vec2{},
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
		-ht, -ht, ht, 0, 0, 1,
		ht, -ht, ht, 0, 0, 1,
		ht, ht, ht, 0, 0, 1,

		ht, ht, ht, 0, 0, 1,
		-ht, ht, ht, 0, 0, 1,
		-ht, -ht, ht, 0, 0, 1,

		// back
		ht, ht, -ht, 0, 0, -1,
		ht, -ht, -ht, 0, 0, -1,
		-ht, -ht, -ht, 0, 0, -1,

		-ht, -ht, -ht, 0, 0, -1,
		-ht, ht, -ht, 0, 0, -1,
		ht, ht, -ht, 0, 0, -1,

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
}
