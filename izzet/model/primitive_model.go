package model

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

type PrimitiveModel struct {
	renderData []RenderData
}

func NewCube() *PrimitiveModel {
	m := &PrimitiveModel{}

	vao := initCubeVAO(50)
	renderData := RenderData{
		Name:        "primitive",
		MeshID:      0,
		Mesh:        createMeshSpec(),
		Transform:   mgl32.Ident4(),
		VAO:         vao,
		VertexCount: 48, // 3 verts per triangle * 2 triangles per face * 8 faces = 48
	}

	m.renderData = []RenderData{renderData}

	return m
}

func createMeshSpec() *modelspec.MeshSpecification {
	pbr := &modelspec.PBRMaterial{
		PBRMetallicRoughness: &modelspec.PBRMetallicRoughness{
			BaseColorTextureIndex: nil,
			BaseColorTextureName:  "",
			BaseColorFactor:       mgl32.Vec4{1, 1, 1, 1},
			MetalicFactor:         0.1,
			RoughnessFactor:       0.1,
		},
	}

	return &modelspec.MeshSpecification{
		PBRMaterial: pbr,
	}
}

func (m *PrimitiveModel) RenderData() []RenderData {
	return m.renderData
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

func initCubeVAO(length int) uint32 {
	ht := float32(length) / 2

	vertices := []float32{
		// front
		-ht, -ht, ht,
		ht, -ht, ht,
		ht, ht, ht,

		ht, ht, ht,
		-ht, ht, ht,
		-ht, -ht, ht,

		// back
		ht, ht, -ht,
		ht, -ht, -ht,
		-ht, -ht, -ht,

		-ht, -ht, -ht,
		-ht, ht, -ht,
		ht, ht, -ht,

		// right
		ht, -ht, ht,
		ht, -ht, -ht,
		ht, ht, -ht,

		ht, ht, -ht,
		ht, ht, ht,
		ht, -ht, ht,

		// left
		-ht, ht, -ht,
		-ht, -ht, -ht,
		-ht, -ht, ht,

		-ht, -ht, ht,
		-ht, ht, ht,
		-ht, ht, -ht,

		// top
		ht, ht, ht,
		ht, ht, -ht,
		-ht, ht, ht,

		-ht, ht, ht,
		ht, ht, -ht,
		-ht, ht, -ht,

		// bottom
		-ht, -ht, ht,
		ht, -ht, -ht,
		ht, -ht, ht,

		-ht, -ht, -ht,
		ht, -ht, -ht,
		-ht, -ht, ht,
	}

	// createVAOs(modelConfig *ModelConfig, meshes []*modelspec.MeshSpecification)

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	// gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, nil)
	// gl.EnableVertexAttribArray(0)

	// gl.BindVertexArray(vao)
	// iztDrawArrays(0, int32(len(vertices))/3)

	return vao
}
