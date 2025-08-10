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

func (p *CameraDepthRenderPass) Render(
	ctx context.RenderContext,
	rctx *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.CameraDepthFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	p.shader.Use()
	renderGeometryWithoutColor(p.app, p.shader, rctx.RenderableEntities, viewerContext, ctx)
}
