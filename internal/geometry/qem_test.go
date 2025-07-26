package geometry_test

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/gltf"
)

func TestQ(t *testing.T) {
	config := &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL}
	doc, err := gltf.ParseGLTF("model", "../../_assets/test/stall.gltf", config)
	if err != nil {
		t.Fail()
		t.Errorf(err.Error())
	}

	geometry.SimplifyMesh(doc.Meshes[0].Primitives[0], -1)
}

func TestFlatShape(t *testing.T) {
	p := &modelspec.PrimitiveSpecification{
		VertexIndices: []uint32{
			0, 3, 4,
			0, 4, 1,
			1, 4, 5,
			1, 5, 2,
			2, 5, 6,

			3, 7, 4,
			4, 7, 8,
			4, 8, 5,
			5, 8, 9,
		},
		UniqueVertices: []modelspec.Vertex{
			modelspec.Vertex{Position: mgl32.Vec3{0, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{2, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{-0.5, 1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{0.5, 1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1.5, 1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{2.5, 1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{0, 2, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, 2, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{2, 2, 0}},
		},
	}

	geometry.SimplifyMesh(p, -1)
}

func TestPyramid(t *testing.T) {
	p := &modelspec.PrimitiveSpecification{
		VertexIndices: []uint32{
			2, 1, 3,
			2, 3, 0,
			3, 1, 0,
			2, 0, 1,
		},
		UniqueVertices: []modelspec.Vertex{
			modelspec.Vertex{Position: mgl32.Vec3{0.5, 1, -0.5}},
			modelspec.Vertex{Position: mgl32.Vec3{0, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, 0, -1}},
			modelspec.Vertex{Position: mgl32.Vec3{2, 0, 0}},
		},
	}

	geometry.SimplifyMesh(p, -1)
}

func TestBox(t *testing.T) {
	p := &modelspec.PrimitiveSpecification{
		VertexIndices: []uint32{
			// left
			4, 0, 3,
			3, 7, 4,

			// right
			5, 6, 2,
			2, 1, 5,

			// front
			4, 5, 1,
			1, 0, 4,

			// back
			7, 3, 2,
			2, 6, 7,

			// bottom
			4, 7, 6,
			6, 5, 4,

			// top
			0, 1, 8,
			1, 2, 8,
			2, 3, 8,
			3, 0, 8,
		},
		UniqueVertices: []modelspec.Vertex{
			modelspec.Vertex{Position: mgl32.Vec3{0, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, 0, -1}},
			modelspec.Vertex{Position: mgl32.Vec3{0, 0, -1}},

			modelspec.Vertex{Position: mgl32.Vec3{0, -1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, -1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, -1, -1}},
			modelspec.Vertex{Position: mgl32.Vec3{0, -1, -1}},

			modelspec.Vertex{Position: mgl32.Vec3{0.5, 0, -0.5}},
		},
	}

	geometry.SimplifyMesh(p, 1)
}

func TestTriangleMerge(t *testing.T) {
	p := &modelspec.PrimitiveSpecification{
		VertexIndices: []uint32{
			0, 1, 4,
			1, 2, 4,
			2, 0, 4,
			0, 3, 1,
			3, 2, 1,

			// back triangle
			// 0, 2, 3,
		},
		UniqueVertices: []modelspec.Vertex{
			modelspec.Vertex{Position: mgl32.Vec3{0, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{1, 0, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{0.5, 0, -1}},
			modelspec.Vertex{Position: mgl32.Vec3{1, -1, 0}},
			modelspec.Vertex{Position: mgl32.Vec3{0.5, 0, -0.5}},
		},
	}

	geometry.SimplifyMesh(p, 1)
}
