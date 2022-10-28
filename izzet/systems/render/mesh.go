package render

import "github.com/go-gl/gl/v4.1-core/gl"

type Mesh interface {
	GetVAO() uint32
}

type MeshImpl struct {
	vao uint32
}

func NewMesh(vertexAttributes []float32, attributeSizes []int) *MeshImpl {
	sumAttributeSize := 0
	attributeSizesUpToIndex := make([]int, len(attributeSizes))
	for i, size := range attributeSizes {
		attributeSizesUpToIndex[i] = sumAttributeSize
		sumAttributeSize += size
	}
	numAttributes := len(attributeSizes)

	vertexAttributes = []float32{
		-0.5, 0, -0.5, 0, 1, 0,
		0.5, 0, 0.5, 0, 1, 0,
		0.5, 0, -0.5, 0, 1, 0,
		0.5, 0, 0.5, 0, 1, 0,
		-0.5, 0, -0.5, 0, 1, 0,
		-0.5, 0, 0.5, 0, 1, 0,
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	for i := 0; i < numAttributes; i++ {
		gl.VertexAttribPointer(uint32(i), int32(attributeSizes[i]), gl.FLOAT, false, int32(sumAttributeSize)*4, gl.PtrOffset(attributeSizesUpToIndex[i]*4))
		gl.EnableVertexAttribArray(uint32(i))
	}

	return &MeshImpl{vao: vao}
}
