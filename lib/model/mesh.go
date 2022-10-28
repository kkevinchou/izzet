package model

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/izzet/lib/modelspec"
)

type MeshChunk struct {
	vao       uint32
	textureID *uint32
	spec      *modelspec.MeshChunkSpecification
}

type Mesh struct {
	meshChunks []*MeshChunk
}

func NewMeshChunk(spec *modelspec.MeshChunkSpecification) *MeshChunk {
	mc := &MeshChunk{
		spec: spec,
	}
	if utils.IsClient() {
		mc.initializeTexture()
		mc.initializeOpenGLObjects()
	}
	return mc
}

func (m *MeshChunk) VAO() uint32 {
	return m.vao
}

func (m *MeshChunk) TextureID() *uint32 {
	return m.textureID
}

func (m *MeshChunk) Vertices() []modelspec.Vertex {
	return m.spec.Vertices
}

func (m *MeshChunk) VertexCount() int {
	return len(m.spec.Vertices)
}

func (m *MeshChunk) PBRMaterial() *modelspec.PBRMaterial {
	return m.spec.PBRMaterial
}

func (m *MeshChunk) initializeTexture() {
	if m.spec.PBRMaterial.PBRMetallicRoughness.BaseColorTextureIndex != nil {
		assetManager := directory.GetDirectory().AssetManager()
		m.textureID = &assetManager.GetTexture(m.spec.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName).ID
	}
}

func (m *MeshChunk) initializeOpenGLObjects() {
	// initialize the VAO
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	m.vao = vao

	var vertexAttributes []float32
	var jointIDsAttribute []int32
	var jointWeightsAttribute []float32

	// set up the source data for the VBOs
	for _, vertex := range m.spec.UniqueVertices {
		position := vertex.Position
		normal := vertex.Normal
		texture := vertex.Texture
		jointIDs := vertex.JointIDs
		jointWeights := vertex.JointWeights

		vertexAttributes = append(vertexAttributes,
			position.X(), position.Y(), position.Z(),
			normal.X(), normal.Y(), normal.Z(),
			texture.X(), texture.Y(),
		)

		ids, weights := FillWeights(jointIDs, jointWeights)
		for _, id := range ids {
			jointIDsAttribute = append(jointIDsAttribute, int32(id))
		}
		jointWeightsAttribute = append(jointWeightsAttribute, weights...)
	}

	totalAttributeSize := len(vertexAttributes) / len(m.spec.UniqueVertices)

	// lay out the position, normal, texture coords in a VBO
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)

	// lay out the joint IDs in a VBO
	var vboJointIDs uint32
	gl.GenBuffers(1, &vboJointIDs)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboJointIDs)
	gl.BufferData(gl.ARRAY_BUFFER, len(jointIDsAttribute)*4, gl.Ptr(jointIDsAttribute), gl.STATIC_DRAW)
	gl.VertexAttribIPointer(3, int32(settings.AnimationMaxJointWeights), gl.INT, int32(settings.AnimationMaxJointWeights)*4, nil)
	gl.EnableVertexAttribArray(3)

	// lay out the joint weights in a VBO
	var vboJointWeights uint32
	gl.GenBuffers(1, &vboJointWeights)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboJointWeights)
	gl.BufferData(gl.ARRAY_BUFFER, len(jointWeightsAttribute)*4, gl.Ptr(jointWeightsAttribute), gl.STATIC_DRAW)
	gl.VertexAttribPointer(4, int32(settings.AnimationMaxJointWeights), gl.FLOAT, false, int32(settings.AnimationMaxJointWeights)*4, nil)
	gl.EnableVertexAttribArray(4)

	// set up the EBO, each triplet of indices point to three vertices
	// that form a triangle.
	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.spec.VertexIndices)*4, gl.Ptr(m.spec.VertexIndices), gl.STATIC_DRAW)
}

func NewMesh(spec *modelspec.MeshSpecification) *Mesh {
	var meshChunks []*MeshChunk
	for _, mc := range spec.MeshChunks {
		meshChunk := NewMeshChunk(mc)
		meshChunks = append(meshChunks, meshChunk)
	}
	return &Mesh{
		meshChunks: meshChunks,
	}
}

func (m *Mesh) MeshChunks() []*MeshChunk {
	return m.meshChunks
}

func (m *Mesh) Vertices() []modelspec.Vertex {
	var vertices []modelspec.Vertex
	for _, meshChunk := range m.meshChunks {
		chunkVerts := meshChunk.Vertices()
		vertices = append(vertices, chunkVerts...)
	}
	return vertices
}
