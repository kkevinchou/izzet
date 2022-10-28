package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Quad struct {
	vao uint32
}

var quadZeroY = []float32{
	// position, normal, texture
	-0.5, 0, -0.5, 0, 1, 0, 0, 1,
	0.5, 0, 0.5, 0, 1, 0, 1, 0,
	0.5, 0, -0.5, 0, 1, 0, 1, 1,
	0.5, 0, 0.5, 0, 1, 0, 1, 0,
	-0.5, 0, -0.5, 0, 1, 0, 0, 1,
	-0.5, 0, 0.5, 0, 1, 0, 0, 0,
}

var quadZeroZ = []float32{
	// bottom
	-0.5, -0.5, 0, 0, 0, 1,
	0.5, -0.5, 0, 0, 0, 1,
	0.5, 0.5, 0, 0, 0, 1,
	0.5, 0.5, 0, 0, 0, 1,
	-0.5, 0.5, 0, 0, 0, 1,
	-0.5, -0.5, 0, 0, 0, 1,
}

func NewQuad(vertexAttributes []float32) *Quad {
	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)

	q := Quad{
		vao: vao,
	}
	return &q
}

func (q *Quad) GetVAO() uint32 {
	return q.vao
}
