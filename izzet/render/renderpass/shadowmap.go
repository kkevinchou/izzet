package renderpass

import (
	"errors"

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
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT,
		int32(p.dimension), int32(p.dimension), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initialize shadow map frame buffer - in the past this was due to an overly large shadow map dimension configuration"))
	}

	ctx.ShadowMapFBO = fbo
	ctx.ShadowMapTexture = texture
}

func (p *ShadowMapRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
}

func (p *ShadowMapRenderPass) Render(
	ctx context.RenderContext,
	rctx *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.ShadowMapFBO)
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
	renderGeometryWithoutColor(lightViewerContext, p.shader, p.app, ctx, rctx.ShadowCastingEntities)
}
