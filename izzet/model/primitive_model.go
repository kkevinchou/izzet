package model

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/modelspec"
)

type PrimitiveModel struct {
	vao         uint32
	geometryVAO uint32
	vertices    []modelspec.Vertex
}

func NewCube() *PrimitiveModel {
	modelConfig := &ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}
	m := &PrimitiveModel{}

	mesh := createPrimitiveSpec()
	m.vao = createVAOs(modelConfig, []*modelspec.PrimitiveSpecification{mesh})[0]
	m.geometryVAO = createGeometryVAOs(modelConfig, []*modelspec.PrimitiveSpecification{mesh})[0]

	vertices := mesh.UniqueVertices
	for _, v := range vertices {
		m.vertices = append(m.vertices, v)
	}

	return m
}

func createPrimitiveSpec() *modelspec.PrimitiveSpecification {
	vertices := cubeVertexFloatsByLength(50)

	vertexIndices := []uint32{}
	for i := 0; i < 48; i++ {
		vertexIndices = append(vertexIndices, uint32(i))
	}

	uniqueVertices := []modelspec.Vertex{}
	for i := 0; i < len(vertices); i += 6 {
		x := vertices[i]
		y := vertices[i+1]
		z := vertices[i+2]

		nx := vertices[i+3]
		ny := vertices[i+4]
		nz := vertices[i+5]

		uniqueVertices = append(uniqueVertices, modelspec.Vertex{
			Position:       mgl32.Vec3{x, y, z},
			Normal:         mgl32.Vec3{nx, ny, nz},
			Texture0Coords: mgl32.Vec2{},
			Texture1Coords: mgl32.Vec2{},
		})
	}

	return &modelspec.PrimitiveSpecification{
		VertexIndices:  vertexIndices,
		UniqueVertices: uniqueVertices,
	}
}

func (m *PrimitiveModel) RenderData() []RenderData {
	renderData := RenderData{
		Name:        "primitive",
		MeshID:      0,
		Transform:   mgl32.Ident4(),
		VAO:         m.vao,
		GeometryVAO: m.geometryVAO,
		VertexCount: 48, // 3 verts per triangle * 2 triangles per face * 8 faces = 48
	}

	return []RenderData{renderData}
}
func (m *PrimitiveModel) JointMap() map[int]*modelspec.JointSpec {
	return nil
}
func (m *PrimitiveModel) RootJoint() *modelspec.JointSpec {
	return nil
}
func (m *PrimitiveModel) Name() string {
	return "primitive"
}
func (m *PrimitiveModel) Vertices() []modelspec.Vertex {
	return m.vertices
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
