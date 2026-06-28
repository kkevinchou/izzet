package renderpass

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/internal/shaders"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
)

type PostProcessingRenderPass struct {
	app    renderiface.App
	shader *shaders.ShaderProgram
}

func NewPostProcessingPass(app renderiface.App, sm *shaders.ShaderManager) *PostProcessingRenderPass {
	return &PostProcessingRenderPass{app: app, shader: sm.GetShaderProgram("post_processing")}
}

func (p *PostProcessingRenderPass) Name() string {
	return "post_process"
}

func (p *PostProcessingRenderPass) Init(width, height int, ctx *context.RenderPassContext) {
	fbo, textures := initFrameBuffer(
		width,
		height,
		[]int32{rendersettings.InternalTextureColorFormatRGB},
		[]uint32{rendersettings.RenderFormatRGB},
		[]uint32{gl.FLOAT},
		false,
		true,
	)
	ctx.PostProcessingFBO = fbo
	ctx.PostProcessingTexture = textures[0]
}

func (p *PostProcessingRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.PostProcessingFBO)
	textures := createAndBindTextures(
		width,
		height,
		[]int32{rendersettings.InternalTextureColorFormatRGB},
		[]uint32{rendersettings.RenderFormatRGB},
		[]uint32{gl.FLOAT},
		true,
	)
	gl.DeleteTextures(1, &ctx.PostProcessingTexture)
	ctx.PostProcessingTexture = textures[0]
}

func (p *PostProcessingRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
) {
	inputTexture := renderPassContext.HDRTexture
	if inputTexture == 0 {
		inputTexture = renderPassContext.MainTexture
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.PostProcessingFBO)

	p.shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, inputTexture)
	p.shader.SetUniformInt("image", 0)

	var kuwahara int32
	if p.app.RuntimeConfig().KuwaharaFilter {
		kuwahara = 1
	}
	p.shader.SetUniformInt("kuwahara", kuwahara)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))

	gl.BindVertexArray(rutils.GetNDCQuadVAO())
	rutils.IztDrawArrays(0, 6)
}
