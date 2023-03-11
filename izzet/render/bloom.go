package render

import (
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/panels"
)

func (r *Renderer) initBloom(maxWidth, maxHeight int) (uint32, uint32, []uint32, []uint32, []uint32) {
	var downSampleFBO uint32
	gl.GenFramebuffers(1, &downSampleFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, downSampleFBO)

	var upSampleFBO uint32
	gl.GenFramebuffers(1, &upSampleFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, upSampleFBO)

	var (
		upSampleTextures    []uint32
		downSampleTextures  []uint32
		blendTargetTextures []uint32
	)

	width, height := maxWidth, maxHeight
	for i := 0; i < mipsCount; i++ {
		// upsampling textures start one doubling earlier than the downsampling textures
		// the first step of upsampling is taking the lowest downsampling mip and upsampling it
		var upsampleTexture uint32
		gl.GenTextures(1, &upsampleTexture)
		gl.BindTexture(gl.TEXTURE_2D, upsampleTexture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		upSampleTextures = append(upSampleTextures, upsampleTexture)

		var blendTargetTexture uint32
		gl.GenTextures(1, &blendTargetTexture)
		gl.BindTexture(gl.TEXTURE_2D, blendTargetTexture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		blendTargetTextures = append(blendTargetTextures, blendTargetTexture)

		width /= 2
		height /= 2

		r.widths = append(r.widths, int32(width))
		r.heights = append(r.heights, int32(height))

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
			int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		downSampleTextures = append(downSampleTextures, texture)
	}

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, downSampleTextures[0], 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, upSampleFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, upSampleTextures[0], 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return downSampleFBO, upSampleFBO, downSampleTextures, upSampleTextures, blendTargetTextures
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
	gl.GenBuffers(1, &vbo)
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

func (r *Renderer) downSample(srcTexture uint32) {
	defer resetGLRenderSettings(r.renderFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.downSampleFBO)

	shader := r.shaderManager.GetShaderProgram("bloom_downsample")
	shader.Use()

	for i := 0; i < len(r.downSampleTextures); i++ {
		width := r.widths[i]
		height := r.heights[i]

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, srcTexture)
		gl.Viewport(0, 0, width, height)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.downSampleTextures[i], 0)

		gl.BindVertexArray(r.xyTextureVAO)
		if i == 0 {
			shader.SetUniformInt("karis", 1)
		} else {
			shader.SetUniformInt("karis", 0)
		}
		if i < int(panels.DBG.BloomThresholdPasses) {
			shader.SetUniformInt("bloomThresholdEnabled", 1)
		} else {
			shader.SetUniformInt("bloomThresholdEnabled", 0)
		}
		shader.SetUniformFloat("bloomThreshold", panels.DBG.BloomThreshold)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		srcTexture = r.downSampleTextures[i]
	}
}

// double check that the upsampling works and blends the right textures
// welp, i need to be ping ponging GG
func (r *Renderer) upSample() uint32 {
	defer resetGLRenderSettings(r.renderFBO)

	mipsCount := len(r.downSampleTextures)

	var upSampleSource uint32
	upSampleSource = r.downSampleTextures[mipsCount-1]
	var i int
	for i = mipsCount - 1; i > 0; i-- {
		blendTargetMip := r.blendTargetTextures[i]
		upSampleMip := r.upSampleTextures[i]

		gl.BindFramebuffer(gl.FRAMEBUFFER, r.upSampleFBO)

		shader := r.shaderManager.GetShaderProgram("bloom_upsample")
		shader.Use()
		shader.SetUniformFloat("upsamplingRadius", panels.DBG.BloomUpsamplingRadius)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, upSampleSource)

		gl.Viewport(0, 0, r.widths[i-1], r.heights[i-1])
		drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, upSampleMip, 0)
		gl.DrawBuffers(1, &drawBuffers[0])

		gl.BindVertexArray(r.xyTextureVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		// r.blend(r.widths[i-1], r.heights[i-1], upSampleMip, r.bloomTextures[i-1], blendTargetMip)
		r.blend(r.widths[i-1], r.heights[i-1], r.downSampleTextures[i-1], upSampleMip, blendTargetMip)
		upSampleSource = blendTargetMip
	}

	blendTargetMip := r.blendTargetTextures[0]
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.upSampleFBO)

	shader := r.shaderManager.GetShaderProgram("bloom_upsample")
	shader.Use()
	shader.SetUniformFloat("upsamplingRadius", panels.DBG.BloomUpsamplingRadius)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, upSampleSource)

	gl.Viewport(0, 0, int32(bloomTextureWidth), int32(bloomTextureHeight))
	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, blendTargetMip, 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	gl.BindVertexArray(r.xyTextureVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	return blendTargetMip
}

func (r *Renderer) blend(width, height int32, texture0, texture1, target uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.compositeFBO)

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
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func (r *Renderer) initComposite(width, height int) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R11F_G11F_B10F,
		int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	r.compositeFBO, r.compositeTexture = fbo, texture
}

func (r *Renderer) composite(renderContext RenderContext, texture0, texture1 uint32) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.compositeFBO)

	shader := r.shaderManager.GetShaderProgram("composite")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("scene", 0)
	shader.SetUniformInt("bloomBlur", 1)
	shader.SetUniformFloat("exposure", panels.DBG.Exposure)
	shader.SetUniformFloat("bloomIntensity", panels.DBG.BloomIntensity)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.compositeTexture, 0)

	gl.BindVertexArray(r.xyTextureVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
