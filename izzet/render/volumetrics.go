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

const textureWidth, textureHeight int = 1024, 1024
const numChannels int = 4

// const cellWidth, cellHeight, cellDepth int = 10, 10, 10
// const workGroupWidth, workGroupHeight, workGroupDepth int = 512, 512, 512

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

type WorleyOctave struct {
	points                           []float32
	cellWidth, cellHeight, cellDepth float32
}

func (r *RenderSystem) setupVolumetrics(shaderManager *shaders.ShaderManager) (uint32, uint32, uint32, uint32) {
	cloudTexture := r.app.RuntimeConfig().CloudTextures[r.app.RuntimeConfig().ActiveCloudTextureIndex]
	// channel := r.app.RuntimeConfig().ActiveCloudTextureChannelIndex

	var octaves []WorleyOctave

	for i := range numChannels {
		cellWidth, cellHeight, cellDepth := cloudTexture.Channels[i].CellWidth, cloudTexture.Channels[i].CellHeight, cloudTexture.Channels[i].CellDepth
		points := noise.Worley3D(int(cellWidth), int(cellHeight), int(cellDepth))
		octave := WorleyOctave{points: points, cellWidth: float32(cellWidth), cellHeight: float32(cellHeight), cellDepth: float32(cellDepth)}
		octaves = append(octaves, octave)
	}

	worleyNoiseTexture := createWorlyNoiseTexture(
		octaves,
		cloudTexture.WorkGroupWidth,
		cloudTexture.WorkGroupHeight,
		cloudTexture.WorkGroupDepth,
	)

	gl.Viewport(0, 0, int32(textureWidth), int32(textureHeight))

	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	texture := createTexture(textureWidth, textureHeight, internalTextureColorFormatRGB, renderFormatRGB, gl.FLOAT, gl.NEAREST)
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

func (r *RenderSystem) renderVolumetrics(vao, texture, fbo uint32, shaderManager *shaders.ShaderManager, assetManager *assets.AssetManager) {
	gl.Viewport(0, 0, int32(textureWidth), int32(textureHeight))
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	shader := shaderManager.GetShaderProgram("worley")
	shader.Use()

	cloudTexture := r.app.RuntimeConfig().CloudTextures[r.app.RuntimeConfig().ActiveCloudTextureIndex]
	activeChannelIndex := r.app.RuntimeConfig().ActiveCloudTextureChannelIndex
	shader.SetUniformFloat("z", cloudTexture.Channels[activeChannelIndex].NoiseZ)
	shader.SetUniformInt("channel", int32(activeChannelIndex))

	shader.SetUniformInt("tex", 0)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_3D, texture)

	shader.SetUniformInt("tex1", 1)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, assetManager.GetTexture("color_grid").ID)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func createWorlyNoiseTexture(octaves []WorleyOctave, workGroupWidth, workGroupHeight, workGroupDepth int32) uint32 {
	shaderProgram := setupComputeShader()
	texture := setupComputeTexture(workGroupWidth, workGroupHeight, workGroupDepth)

	for i := range 4 {
		octave := octaves[i]
		// Create the SSBO
		var ssbo uint32
		gl.GenBuffers(1, &ssbo)
		gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, ssbo)

		// Upload the data to the SSBO
		gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(octave.points)*4, gl.Ptr(octave.points), gl.STATIC_DRAW)

		// Bind the SSBO to a binding point (0 in this case)
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, uint32(i), ssbo)
	}

	// Unbind the SSBO
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)

	gl.UseProgram(shaderProgram)

	wu := gl.GetUniformLocation(shaderProgram, gl.Str("widths"+"\x00"))
	hu := gl.GetUniformLocation(shaderProgram, gl.Str("heights"+"\x00"))
	du := gl.GetUniformLocation(shaderProgram, gl.Str("depths"+"\x00"))

	gl.Uniform4i(wu, int32(octaves[0].cellWidth), int32(octaves[1].cellWidth), int32(octaves[2].cellWidth), int32(octaves[3].cellWidth))
	gl.Uniform4i(hu, int32(octaves[0].cellHeight), int32(octaves[1].cellHeight), int32(octaves[2].cellHeight), int32(octaves[3].cellHeight))
	gl.Uniform4i(du, int32(octaves[0].cellDepth), int32(octaves[1].cellDepth), int32(octaves[2].cellDepth), int32(octaves[3].cellDepth))

	gl.BindTexture(gl.TEXTURE_3D, texture)
	gl.DispatchCompute(uint32(workGroupWidth), uint32(workGroupHeight), uint32(workGroupDepth))
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	return texture
}

func setupComputeTexture(width, height, depth int32) uint32 {
	var texture uint32

	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_3D, texture)

	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA32F, width, height, depth, 0, gl.RGBA, gl.FLOAT, nil)

	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	// gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	// gl.BindImageTexture(4, texture, 0, true, 0, gl.WRITE_ONLY, gl.RGBA32F)
	gl.BindImageTexture(4, texture, 0, true, 0, gl.READ_WRITE, gl.RGBA32F)
	// gl.BindImageTexture(4, texture, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

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
