package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/kitolib/shaders"
)

// - create 3D worley points (128 x 128 x 128)
// - store worley points in a compute buffer
// - run the compute shader to create a 3d noise texture

// texture details:
//     - texture 1 - high resolution shape texture
//     - texture 2 - low resolution detail texture
//         - small details
//     - noise can be stored in R, G, B, A channels

// rendering:
//     - the fragment shader samples the 3d texture by ray marching from the view direction

func setupVolumetrics(shaderManager *shaders.ShaderManager) uint32 {
	worleyNoiseTexture := createWorlyNoiseTexture()

	width := 128
	height := 128

	fbo, texture := initFBOAndTexture(width, height)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Vertices for a full-screen quad in normalized device coordinates (NDC)
	vertices := []float32{
		-1.0, 1.0, 0.0, 3.0, // Top-left
		-1.0, -1.0, 0.0, 0.0, // Bottom-left
		1.0, 1.0, 3.0, 3.0, // Top-right
		1.0, -1.0, 3.0, 0.0, // Bottom-right
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))

	gl.BindVertexArray(0) // Unbind the VAO

	if err := shaderManager.CompileShaderProgram("worley", "worley", "worley", ""); err != nil {
		panic(err)
	}

	shader := shaderManager.GetShaderProgram("worley")
	shader.Use()

	gl.BindVertexArray(vao)
	gl.BindTexture(gl.TEXTURE_3D, worleyNoiseTexture)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	return texture
}

func createWorlyNoiseTexture() uint32 {
	const width, height, depth int = 4, 4, 4

	shaderProgram := setupComputeShader()
	texture := setupTexture(width, height, depth)

	gl.UseProgram(shaderProgram)
	// 64 work groups
	gl.DispatchCompute(uint32(width), uint32(height), uint32(depth))
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	return texture
}

func setupTexture(width, height, depth int) uint32 {
	var texture uint32

	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_3D, texture)

	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA32F, int32(width), int32(height), int32(depth), 0, gl.RGBA, gl.FLOAT, nil)

	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	// gl.BindImageTexture(0, texture, 0, true, 0, gl.WRITE_ONLY, gl.RGBA32F)
	// gl.BindImageTexture(0, texture, 0, true, 0, gl.WRITE_ONLY, gl.RGBA32F)
	gl.BindImageTexture(0, texture, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

	return texture
}

func setupComputeShader() uint32 {
	sourceStr, err := os.ReadFile(filepath.Join("shaders", "worley.compute"))
	if err != nil {
		panic(err)
	}

	compute := gl.CreateShader(gl.COMPUTE_SHADER)
	glSourceStr, free := gl.Strs(string(sourceStr) + "\x00")
	defer free()

	gl.ShaderSource(compute, 1, glSourceStr, nil)
	gl.CompileShader(compute)

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, compute)
	gl.LinkProgram(shaderProgram)

	var status int32
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		panic(fmt.Errorf("failed to link shader program:\n%s", log))
	}
	return shaderProgram
}
