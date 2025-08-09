package renderpass

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/shaders"
)

type CameraDepthRenderPass struct {
	app    renderiface.App
	shader *shaders.ShaderProgram
}

func NewCameraDepthPass(app renderiface.App, sm *shaders.ShaderManager) *CameraDepthRenderPass {
	return &CameraDepthRenderPass{app: app, shader: sm.GetShaderProgram("modelgeo")}
}

func (p *CameraDepthRenderPass) Init(width, height int, ctx *context.RenderPassContext) {
	// CameraDepthTextureFn := textureFn(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT})
	// fbo, textures := initFrameBufferNoDepth(ssaoBlurTextureFn)
	texture := createDepthTexture(width, height)
	fbo := initDepthMapFrameBuffer(texture)
	ctx.CameraDepthFBO = fbo
	ctx.CameraDepthTexture = texture
}

func (p *CameraDepthRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.CameraDepthFBO)
	texture := createDepthTexture(width, height)
	ctx.CameraDepthTexture = texture
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, ctx.CameraDepthTexture, 0)
}

func (p *CameraDepthRenderPass) Render(ctx context.RenderContext, rctx *context.RenderPassContext, viewerContext context.ViewerContext) {
	// gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.SSAOBlurFBO)

	// gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	// gl.ClearColor(0, 0, 0, 1)
	// gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// p.shader.Use()

	// gl.ActiveTexture(gl.TEXTURE0)
	// gl.BindTexture(gl.TEXTURE_2D, ctx.SSAOTexture)

	// gl.BindVertexArray(createNDCQuadVAO())
	// iztDrawArrays(p.app, 0, 6)
	gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.CameraDepthFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	p.shader.Use()
	renderGeometryWithoutColor(viewerContext, p.shader, p.app, ctx, rctx.RenderableEntities)
}
