package render

import (
	"fmt"
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

func (r *Renderer) setupVolumetrics(shaderManager *shaders.ShaderManager) uint32 {
	width := 128
	height := 128

	fbo, texture := r.initFBOAndTexture(width, height)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Vertices for a full-screen quad in normalized device coordinates (NDC)
	vertices := []float32{
		-1.0, 1.0, // Top-left
		-1.0, -1.0, // Bottom-left
		1.0, 1.0, // Top-right
		1.0, -1.0, // Bottom-right
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Enable and set up the position attribute (0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	gl.BindVertexArray(0) // Unbind the VAO

	if err := shaderManager.CompileShaderProgram("worley", "worley", "worley", ""); err != nil {
		panic(err)
	}

	shader := r.shaderManager.GetShaderProgram("worley")
	shader.Use()

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	return texture
}

func (r *Renderer) createWorlyNoiseTexture() uint32 {
	// set up shader

	const sourceStr string = `
	#version 430 core
	layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
	layout(rgba32f, binding = 0) uniform image2D imgOutput;

	void main() {
		vec4 value = vec4(0.0, 0.0, 0.0, 1.0);
		ivec2 texelCoord = ivec2(gl_GlobalInvocationID.xy);
		
		value.x = float(texelCoord.x)/(gl_NumWorkGroups.x);
		value.y = float(texelCoord.y)/(gl_NumWorkGroups.y);
		
		imageStore(imgOutput, texelCoord, value);
	}
	`

	compute := gl.CreateShader(gl.COMPUTE_SHADER)
	glSourceStr, free := gl.Strs(sourceStr + "\x00")
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

	// set up texture

	const width, height int = 512, 512

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)

	gl.BindImageTexture(0, texture, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

	gl.UseProgram(shaderProgram)
	gl.DispatchCompute(uint32(width), uint32(height), 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	return texture
}
