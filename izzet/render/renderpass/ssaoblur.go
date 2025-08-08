package renderpass

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
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
	ssaoBlurTextureFn := textureFn(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT})
	fbo, textures := initFrameBufferNoDepth(ssaoBlurTextureFn)
	ctx.SSAOBlurFBO = fbo
	ctx.SSAOBlurTexture = textures[0]
}

func (p *SSAOBlurRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.SSAOBlurFBO)
	ssaoBlurTextureFn := textureFn(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT})
	_, _, textures := ssaoBlurTextureFn()
	ctx.SSAOBlurTexture = textures[0]
}

func (p *SSAOBlurRenderPass) Render(ctx context.RenderContext, rctx *context.RenderPassContext, viewerContext context.ViewerContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.SSAOBlurFBO)

	gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	p.shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, rctx.SSAOTexture)

	gl.BindVertexArray(createNDCQuadVAO())
	iztDrawArrays(p.app, 0, 6)
}
