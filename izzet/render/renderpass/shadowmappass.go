package renderpass

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/shaders"
)

type ShadowMapRenderPass struct {
	app       renderiface.App
	shader    *shaders.ShaderProgram
	dimension int
}

func NewShadowMapPass(dimension int, app renderiface.App, sm *shaders.ShaderManager) *ShadowMapRenderPass {
	return &ShadowMapRenderPass{dimension: dimension, app: app, shader: sm.GetShaderProgram("modelgeo")}
}

func (p *ShadowMapRenderPass) Init(_, _ int, ctx *context.RenderPassContext) {
	fbo, texture := initDepthOnlyFrameBuffer(p.dimension, p.dimension)
	ctx.ShadowMapFBO = fbo
	ctx.ShadowMapTexture = texture
}

func (p *ShadowMapRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
}

func (p *ShadowMapRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.ShadowMapFBO)
	gl.Viewport(0, 0, int32(p.dimension), int32(p.dimension))

	if !p.app.RuntimeConfig().EnableShadowMapping {
		// set the depth to be max value to prevent shadow mapping
		gl.ClearDepth(1)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		return
	}

	gl.CullFace(gl.FRONT)
	defer gl.CullFace(gl.BACK)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	p.shader.Use()
	renderGeometryWithoutColor(p.app, p.shader, renderContext.ShadowCastingEntities, lightViewerContext, renderContext)
}
