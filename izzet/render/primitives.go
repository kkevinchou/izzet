package render

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/app/apputils"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

var internedQuadVAOPositionUV uint32

func getInternedQuadVAOPositionUV() uint32 {
	if internedQuadVAOPositionUV == 0 {
		var internedQuadVBO uint32
		var internedQuadVertices = []float32{
			-1, -1, 0, 0.0, 0.0,
			1, -1, 0, 1.0, 0.0,
			1, 1, 0, 1.0, 1.0,
			1, 1, 0, 1.0, 1.0,
			-1, 1, 0, 0.0, 1.0,
			-1, -1, 0, 0.0, 0.0,
		}

		var backVertices []float32 = []float32{
			1, 1, 0, 1.0, 1.0,
			1, -1, 0, 1.0, 0.0,
			-1, -1, 0, 0.0, 0.0,
			-1, -1, 0, 0.0, 0.0,
			-1, 1, 0, 0.0, 1.0,
			1, 1, 0, 1.0, 1.0,
		}

		// always add the double sided vertices
		// when a draw request comes in, if doubleSided is false we only draw the first half of the vertices
		// this is wasteful for scenarios where we don't need all vertices
		internedQuadVertices = append(internedQuadVertices, backVertices...)

		// var vbo, dtqVao uint32
		apputils.GenBuffers(1, &internedQuadVBO)
		gl.GenVertexArrays(1, &internedQuadVAOPositionUV)

		gl.BindVertexArray(internedQuadVAOPositionUV)
		gl.BindBuffer(gl.ARRAY_BUFFER, internedQuadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(internedQuadVertices)*4, gl.Ptr(internedQuadVertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		gl.EnableVertexAttribArray(1)
	}

	return internedQuadVAOPositionUV
}

var internedQuadVAOPosition uint32

func getInternedQuadVAOPosition() uint32 {
	if internedQuadVAOPosition == 0 {
		var internedQuadVBO uint32
		var internedQuadVertices = []float32{
			-1, -1, 0,
			1, -1, 0,
			1, 1, 0,
			1, 1, 0,
			-1, 1, 0,
			-1, -1, 0,
		}

		var backVertices []float32 = []float32{
			1, 1, 0,
			1, -1, 0,
			-1, -1, 0,
			-1, -1, 0,
			-1, 1, 0,
			1, 1, 0,
		}

		// always add the double sided vertices
		// when a draw request comes in, if doubleSided is false we only draw the first half of the vertices
		// this is wasteful for scenarios where we don't need all vertices
		internedQuadVertices = append(internedQuadVertices, backVertices...)

		// var vbo, dtqVao uint32
		apputils.GenBuffers(1, &internedQuadVBO)
		gl.GenVertexArrays(1, &internedQuadVAOPosition)

		gl.BindVertexArray(internedQuadVAOPosition)
		gl.BindBuffer(gl.ARRAY_BUFFER, internedQuadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(internedQuadVertices)*4, gl.Ptr(internedQuadVertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)
	}

	return internedQuadVAOPosition
}

func (r *Renderer) drawTexturedQuad(viewerContext *ViewerContext, shaderManager *shaders.ShaderManager, texture uint32, aspectRatio float32, modelMatrix *mgl32.Mat4, doubleSided bool, pickingID *int) {
	vao := getInternedQuadVAOPositionUV()

	gl.BindVertexArray(vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	if modelMatrix != nil {
		shader := shaderManager.GetShaderProgram("world_space_quad")
		shader.Use()
		if pickingID != nil {
			shader.SetUniformUInt("entityID", uint32(*pickingID))
		}
		shader.SetUniformMat4("model", *modelMatrix)
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	} else {
		shader := shaderManager.GetShaderProgram("screen_space_quad")
		shader.Use()
	}

	// honestly we should clean up this quad drawing logic
	numVertices := 6
	if doubleSided {
		numVertices *= 2
	}

	r.iztDrawArrays(0, int32(numVertices))
}
