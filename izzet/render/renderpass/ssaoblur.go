package renderpass

import (
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/kitolib/shaders"
)

type SSAOBlurRenderPass struct {
	app    renderiface.App
	shader *shaders.ShaderProgram
}

func NewSSAOBlurPass(app renderiface.App, sm *shaders.ShaderManager) *SSAOBlurRenderPass {
	return &SSAOBlurRenderPass{app: app, shader: sm.GetShaderProgram("blur")}
}

func (p *SSAOBlurRenderPass) Init(width, height int, ctx *context.RenderPassContext) {
	fbo, textures := initFrameBuffer(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT}, false, true)
	ctx.SSAOBlurFBO = fbo
	ctx.SSAOBlurTexture = textures[0]
}

func (p *SSAOBlurRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.SSAOBlurFBO)
	textures := createAndBindTextures(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT}, true)
	gl.DeleteTextures(1, &ctx.SSAOBlurTexture)
	ctx.SSAOBlurTexture = textures[0]
}

func (p *SSAOBlurRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	start := time.Now()
	defer func() {
		globals.ClientRegistry().Inc("render_ssao_blur_pass", float64(time.Since(start).Milliseconds()))
	}()
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.SSAOBlurFBO)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	p.shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.SSAOTexture)

	gl.BindVertexArray(rutils.GetNDCQuadVAO())
	rutils.IztDrawArrays(0, 6)
}
