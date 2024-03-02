package render

import (
	"errors"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/kkevinchou/izzet/app/apputils"
)

func createSamplingDimensions(startWidth int, startHeight int, count int) ([]int, []int) {
	var widths []int
	var heights []int

	width := startWidth
	height := startHeight

	for i := 0; i < count; i++ {
		widths = append(widths, width)
		heights = append(heights, height)
		width /= 2
		height /= 2
	}

	return widths, heights
}

func initSamplingTextures(widths, heights []int) []uint32 {
	var textures []uint32

	for i := 0; i < len(widths); i++ {
		width := widths[i]
		height := heights[i]

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, internalTextureColorFormat,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		textures = append(textures, texture)
	}

	return textures
}

func initSamplingBuffer(texture uint32) uint32 {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo
}

func (r *Renderer) init2f2fVAO() uint32 {
	vertices := []float32{
		-1, -1, 0.0, 0.0,
		1, -1, 1.0, 0.0,
		1, 1, 1.0, 1.0,
		1, 1, 1.0, 1.0,
		-1, 1, 0.0, 1.0,
		-1, -1, 0.0, 0.0,
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	return vao
}

func (r *Renderer) downSample(srcTexture uint32, widths, heights []int) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.downSampleFBO)

	shader := r.shaderManager.GetShaderProgram("bloom_downsample")
	shader.Use()

	for i := 0; i < len(r.downSampleTextures); i++ {
		width := widths[i]
		height := heights[i]

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, srcTexture)
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.downSampleTextures[i], 0)

		gl.BindVertexArray(r.xyTextureVAO)
		if i == 0 {
			shader.SetUniformInt("karis", 1)
		} else {
			shader.SetUniformInt("karis", 0)
		}
		if i < int(r.app.RuntimeConfig().BloomThresholdPasses) {
			shader.SetUniformInt("bloomThresholdEnabled", 1)
		} else {
			shader.SetUniformInt("bloomThresholdEnabled", 0)
		}
		shader.SetUniformFloat("bloomThreshold", r.app.RuntimeConfig().BloomThreshold)
		r.iztDrawArrays(0, 6)
		srcTexture = r.downSampleTextures[i]
	}
}

// double check that the upsampling works and blends the right textures
// welp, i need to be ping ponging GG
func (r *Renderer) upSample(widths, heights []int) uint32 {
	mipsCount := len(r.downSampleTextures)

	var upSampleSource uint32
	upSampleSource = r.downSampleTextures[mipsCount-1]
	var i int
	for i = mipsCount - 1; i > 0; i-- {
		width := int32(widths[i-1])
		height := int32(heights[i-1])

		blendTargetMip := r.blendTargetTextures[i]
		upSampleMip := r.upSampleTextures[i]

		gl.BindFramebuffer(gl.FRAMEBUFFER, r.upSampleFBO)

		shader := r.shaderManager.GetShaderProgram("bloom_upsample")
		shader.Use()
		shader.SetUniformFloat("upSamplingScale", r.app.RuntimeConfig().BloomUpsamplingScale)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, upSampleSource)

		gl.Viewport(0, 0, width, height)
		drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, upSampleMip, 0)
		gl.DrawBuffers(1, &drawBuffers[0])

		gl.BindVertexArray(r.xyTextureVAO)
		r.iztDrawArrays(0, 6)

		r.blend(width, height, r.downSampleTextures[i-1], upSampleMip, blendTargetMip)
		upSampleSource = blendTargetMip
	}

	blendTargetMip := r.blendTargetTextures[0]
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.upSampleFBO)

	shader := r.shaderManager.GetShaderProgram("bloom_upsample")
	shader.Use()
	shader.SetUniformFloat("upSamplingScale", r.app.RuntimeConfig().BloomUpsamplingScale)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, upSampleSource)

	gl.Viewport(0, 0, int32(MaxBloomTextureWidth), int32(MaxBloomTextureHeight))
	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, blendTargetMip, 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	gl.BindVertexArray(r.xyTextureVAO)
	r.iztDrawArrays(0, 6)

	return blendTargetMip
}

func (r *Renderer) blend(width, height int32, texture0, texture1, target uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.blendFBO)

	shader := r.shaderManager.GetShaderProgram("blend")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("texture0", 0)
	shader.SetUniformInt("texture1", 1)

	gl.Viewport(0, 0, width, height)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, target, 0)

	gl.BindVertexArray(r.xyTextureVAO)
	r.iztDrawArrays(0, 6)
}

func (r *Renderer) composite(renderContext RenderContext, texture0, texture1 uint32) uint32 {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.compositeFBO)

	shader := r.shaderManager.GetShaderProgram("composite")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("scene", 0)
	shader.SetUniformInt("bloomBlur", 1)
	shader.SetUniformFloat("exposure", r.app.RuntimeConfig().Exposure)
	shader.SetUniformFloat("bloomIntensity", r.app.RuntimeConfig().BloomIntensity)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.compositeTexture, 0)

	gl.BindVertexArray(r.xyTextureVAO)
	r.iztDrawArrays(0, 6)

	return r.compositeTexture
}
