package render

import "github.com/go-gl/gl/v4.1-core/gl"

type SkyBox struct {
	vbo uint32
	vao uint32
}

func NewSkyBox(scale float32) *SkyBox {
	// var skyboxVertices []float32 = []float32{

	// }

	// texture coords top left = 0,0 | bottom right = 1,1
	var skyboxVertices []float32 = []float32{
		// front
		-0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 0.0,

		// top
		-0.5, 0.5, -0.5, 0.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 0.0,

		// left
		-0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 1.0, 0.0,
		-0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, -0.5, -0.5, 1.0, 0.0,
		-0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		// right
		0.5, 0.5, 0.5, 1.0, 1.0,
		0.5, 0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 1.0,

		// bottom
		-0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		// back
		-0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 0.0, 1.0,
		0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, 0.5, 1.0, 1.0,
	}

	for i := 0; i < len(skyboxVertices); i += 5 {
		skyboxVertices[i] *= scale
		skyboxVertices[i+1] *= scale
		skyboxVertices[i+2] *= scale
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(skyboxVertices)*4, gl.Ptr(skyboxVertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	sb := SkyBox{
		vbo: vbo,
		vao: vao,
	}
	return &sb
}

func (sb *SkyBox) VAO() uint32 {
	return sb.vao
}
