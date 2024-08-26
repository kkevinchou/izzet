package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/internal/noise"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/kitolib/shaders"
)

const width, height int = 1024, 1024

const workGroupWidth, workGroupHeight, workGroupDepth int = 128, 128, 128

// const workGroupWidth, workGroupHeight, workGroupDepth int = 30, 30, 30

// const workGroupWidth, workGroupHeight, workGroupDepth int = 4, 4, 4

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

func (r *Renderer) setupVolumetrics(shaderManager *shaders.ShaderManager) (uint32, uint32, uint32, uint32) {
	points := noise.Worley3D(workGroupWidth, workGroupHeight, workGroupDepth)
	worleyNoiseTexture := createWorlyNoiseTexture(points)

	gl.Viewport(0, 0, int32(width), int32(height))

	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	texture := createTexture(width, height, internalTextureColorFormat, gl.RGB, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Vertices for a full-screen quad in normalized device coordinates (NDC)
	vertices := []float32{
		// Positions   // TexCoords
		-1.0, 1.0, 0.0, 1.0,
		-1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 0.0,

		-1.0, 1.0, 0.0, 1.0,
		1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
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

	return vao, worleyNoiseTexture, fbo, texture
}

func (r *Renderer) renderVolumetrics(vao, texture, fbo uint32, shaderManager *shaders.ShaderManager, assetManager *assets.AssetManager) {
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	shader := shaderManager.GetShaderProgram("worley")
	shader.Use()

	shader.SetUniformFloat("z", r.app.RuntimeConfig().NoiseZ)

	shader.SetUniformInt("tex", 0)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_3D, texture)

	shader.SetUniformInt("tex1", 1)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, assetManager.GetTexture("color_grid").ID)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func createWorlyNoiseTexture(points []float32) uint32 {
	shaderProgram := setupComputeShader()
	texture := setupComputeTexture(workGroupWidth, workGroupHeight, workGroupDepth)

	// Create the SSBO
	var ssbo uint32
	gl.GenBuffers(1, &ssbo)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, ssbo)

	// Upload the data to the SSBO
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(points)*4, gl.Ptr(points), gl.STATIC_DRAW)

	// Bind the SSBO to a binding point (0 in this case)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, ssbo)

	// Unbind the SSBO
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)

	gl.UseProgram(shaderProgram)
	gl.BindTexture(gl.TEXTURE_3D, texture)
	gl.DispatchCompute(uint32(workGroupWidth), uint32(workGroupHeight), uint32(workGroupDepth))
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	return texture
}

func setupComputeTexture(width, height, depth int) uint32 {
	var texture uint32

	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_3D, texture)

	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA32F, int32(width), int32(height), int32(depth), 0, gl.RGBA, gl.FLOAT, nil)

	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	// gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	// gl.BindImageTexture(1, texture, 0, true, 0, gl.WRITE_ONLY, gl.RGBA32F)
	gl.BindImageTexture(1, texture, 0, true, 0, gl.READ_WRITE, gl.RGBA32F)
	// gl.BindImageTexture(1, texture, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

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
