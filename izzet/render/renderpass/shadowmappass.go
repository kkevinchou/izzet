package renderpass

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/shaders"
)

type ShadowMapRenderPass struct {
	app       renderiface.App
	shader    *shaders.ShaderProgram
	dimension int
}

func NewShadowMapPass(dimension int, app renderiface.App, sm *shaders.ShaderManager) *ShadowMapRenderPass {
	return &ShadowMapRenderPass{dimension: dimension, app: app, shader: sm.GetShaderProgram("cascaded_shadow_map")}
}

func (p *ShadowMapRenderPass) Name() string {
	return "shadow_pass"
}

func (p *ShadowMapRenderPass) Init(_, _ int, ctx *context.RenderPassContext) {
	fbo, texture := initShadowMapFrameBuffer(p.dimension, p.dimension)
	ctx.ShadowMapFBO = fbo
	ctx.ShadowMapTexture = texture
}

func initShadowMapFrameBuffer(width, height int) (uint32, uint32) {
	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)

	texture := createShadowMapDepthTexture(width, height)

	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, texture, 0)

	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("failed to initialize shadow map frame buffer - in the past this was due to an overly large shadow map dimension configuration")
	}

	return depthMapFBO, texture
}

func createShadowMapDepthTexture(width, height int) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, texture)

	gl.TexImage3D(gl.TEXTURE_2D_ARRAY, 0, gl.DEPTH_COMPONENT,
		int32(width), int32(height), int32(settings.NumShadowMapCascades), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// swizzle the red channel to rgba so that when we view the texture in the texture viewer
	// it is in black/white instead of red/black
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_SWIZZLE_R, gl.RED)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_SWIZZLE_G, gl.RED)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_SWIZZLE_B, gl.RED)

	return texture
}

func (p *ShadowMapRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
}

func (p *ShadowMapRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.ShadowMapFBO)
	gl.Viewport(0, 0, int32(p.dimension), int32(p.dimension))

	if !p.app.RuntimeConfig().EnableShadowMapping {
		gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, renderPassContext.ShadowMapTexture, 0)
		gl.ClearDepth(1)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		return
	}

	gl.CullFace(gl.FRONT)
	defer gl.CullFace(gl.BACK)

	p.shader.Use()

	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, renderPassContext.ShadowMapTexture, 0)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	p.shader.SetUniformInt("cascadeCount", int32(len(renderContext.ShadowMapCascades)))
	for i, cascade := range renderContext.ShadowMapCascades {
		p.shader.SetUniformMat4(fmt.Sprintf("lightSpaceMatrixArray[%d]", i), utils.Mat4F64ToF32(cascade.ViewerContext.ViewProjectionMatrix))
	}

	renderGeometryWithoutColor(p.app, p.shader, renderContext.ShadowCastingEntities, viewerContext, renderContext)
}
