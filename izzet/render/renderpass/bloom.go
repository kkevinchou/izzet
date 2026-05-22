package renderpass

import (
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/kitolib/shaders"
)

const (
	bloomMipCount         int = 6
	maxBloomTextureWidth  int = 1920
	maxBloomTextureHeight int = 1080
)

type BloomRenderPass struct {
	app renderiface.App
	sm  *shaders.ShaderManager

	downSampleWidths  []int
	downSampleHeights []int
}

func NewBloomPass(app renderiface.App, sm *shaders.ShaderManager) *BloomRenderPass {
	return &BloomRenderPass{app: app, sm: sm}
}

func (p *BloomRenderPass) Name() string {
	return "bloom_pass"
}

func (p *BloomRenderPass) Init(width, height int, ctx *context.RenderPassContext) {
	downSampleWidths, downSampleHeights := createSamplingDimensions(maxBloomTextureWidth/2, maxBloomTextureHeight/2, bloomMipCount)
	p.downSampleWidths = downSampleWidths
	p.downSampleHeights = downSampleHeights
	ctx.BloomDownSampleTextures = initSamplingTextures(downSampleWidths, downSampleHeights)
	ctx.BloomDownSampleFBO = initSamplingBuffer(ctx.BloomDownSampleTextures[0])

	upSampleWidths, upSampleHeights := createSamplingDimensions(maxBloomTextureWidth, maxBloomTextureHeight, bloomMipCount)
	ctx.BloomUpSampleTextures = initSamplingTextures(upSampleWidths, upSampleHeights)
	ctx.BloomBlendTextures = initSamplingTextures(upSampleWidths, upSampleHeights)
	ctx.BloomUpSampleFBO = initSamplingBuffer(ctx.BloomUpSampleTextures[0])
	ctx.BloomBlendFBO = initSamplingBuffer(ctx.BloomBlendTextures[0])

	ctx.BloomCompositeFBO, ctx.BloomCompositeTexture = initBloomCompositeBuffer(width, height)
	ctx.HDRTexture = ctx.MainTexture
}

func (p *BloomRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.BloomCompositeFBO)
	gl.DeleteTextures(1, &ctx.BloomCompositeTexture)
	ctx.BloomCompositeTexture = createTexture(
		width,
		height,
		rendersettings.InternalTextureColorFormatRGB,
		rendersettings.RenderFormatRGB,
		gl.FLOAT,
		gl.LINEAR,
		true,
	)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, ctx.BloomCompositeTexture, 0)
}

func (p *BloomRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
) {
	if !p.app.RuntimeConfig().Bloom {
		renderPassContext.HDRTexture = renderPassContext.MainTexture
		return
	}

	p.downSample(renderPassContext.MainTexture, renderPassContext)
	renderPassContext.BloomTexture = p.upSampleAndBlend(renderPassContext)
	renderPassContext.HDRTexture = p.composite(renderContext, renderPassContext.MainTexture, renderPassContext.BloomTexture, renderPassContext)
}

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

		texture := createTexture(
			width,
			height,
			rendersettings.InternalTextureColorFormatRGB,
			rendersettings.RenderFormatRGB,
			gl.FLOAT,
			gl.LINEAR,
			true,
		)

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

func initBloomCompositeBuffer(width, height int) (uint32, uint32) {
	fbo, textures := initFrameBuffer(
		width,
		height,
		[]int32{rendersettings.InternalTextureColorFormatRGB},
		[]uint32{rendersettings.RenderFormatRGB},
		[]uint32{gl.FLOAT},
		false,
		true,
	)
	return fbo, textures[0]
}

func (p *BloomRenderPass) downSample(srcTexture uint32, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.BloomDownSampleFBO)

	shader := p.sm.GetShaderProgram("bloom_downsample")
	shader.Use()

	for i := 0; i < len(ctx.BloomDownSampleTextures); i++ {
		width := p.downSampleWidths[i]
		height := p.downSampleHeights[i]

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, srcTexture)
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, ctx.BloomDownSampleTextures[i], 0)

		gl.BindVertexArray(rutils.GetNDCQuadVAO())
		if i == 0 {
			shader.SetUniformInt("karis", 1)
		} else {
			shader.SetUniformInt("karis", 0)
		}
		if i < int(p.app.RuntimeConfig().BloomThresholdPasses) {
			shader.SetUniformInt("bloomThresholdEnabled", 1)
		} else {
			shader.SetUniformInt("bloomThresholdEnabled", 0)
		}
		shader.SetUniformFloat("bloomThreshold", p.app.RuntimeConfig().BloomThreshold)
		rutils.IztDrawArrays(0, 6)
		srcTexture = ctx.BloomDownSampleTextures[i]
	}
}

func (p *BloomRenderPass) upSampleAndBlend(ctx *context.RenderPassContext) uint32 {
	mipsCount := len(ctx.BloomDownSampleTextures)

	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.BloomUpSampleFBO)

	upSampleSource := ctx.BloomDownSampleTextures[mipsCount-1]
	var i int
	for i = mipsCount - 1; i > 0; i-- {
		width := int32(p.downSampleWidths[i-1])
		height := int32(p.downSampleHeights[i-1])

		upSampleTarget := ctx.BloomUpSampleTextures[i]
		p.upSample(width, height, upSampleSource, upSampleTarget, ctx)

		blendTarget := ctx.BloomBlendTextures[i]
		p.blend(width, height, ctx.BloomDownSampleTextures[i-1], upSampleTarget, blendTarget, ctx)

		upSampleSource = blendTarget
	}

	blendTargetMip := ctx.BloomBlendTextures[0]
	p.upSample(int32(maxBloomTextureWidth), int32(maxBloomTextureHeight), upSampleSource, blendTargetMip, ctx)

	return blendTargetMip
}

func (p *BloomRenderPass) upSample(width, height int32, source, target uint32, ctx *context.RenderPassContext) {
	shader := p.sm.GetShaderProgram("bloom_upsample")
	shader.Use()
	shader.SetUniformFloat("upSamplingScale", p.app.RuntimeConfig().BloomUpsamplingScale)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, source)

	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.BloomUpSampleFBO)
	gl.Viewport(0, 0, width, height)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, target, 0)

	gl.BindVertexArray(rutils.GetNDCQuadVAO())
	rutils.IztDrawArrays(0, 6)
}

func (p *BloomRenderPass) blend(width, height int32, texture0, texture1, target uint32, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.BloomBlendFBO)

	shader := p.sm.GetShaderProgram("blend")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("texture0", 0)
	shader.SetUniformInt("texture1", 1)

	gl.Viewport(0, 0, width, height)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, target, 0)

	gl.BindVertexArray(rutils.GetNDCQuadVAO())
	rutils.IztDrawArrays(0, 6)
}

func (p *BloomRenderPass) composite(renderContext context.RenderContext, texture0, texture1 uint32, ctx *context.RenderPassContext) uint32 {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.BloomCompositeFBO)

	shader := p.sm.GetShaderProgram("composite")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, texture1)

	shader.SetUniformInt("scene", 0)
	shader.SetUniformInt("bloomBlur", 1)
	shader.SetUniformFloat("exposure", p.app.RuntimeConfig().Exposure)
	shader.SetUniformFloat("bloomIntensity", p.app.RuntimeConfig().BloomIntensity)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))

	gl.BindVertexArray(rutils.GetNDCQuadVAO())
	rutils.IztDrawArrays(0, 6)

	return ctx.BloomCompositeTexture
}
