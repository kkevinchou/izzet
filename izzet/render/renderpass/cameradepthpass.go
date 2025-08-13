package renderpass

import (
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/globals"
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
	fbo, texture := initDepthOnlyFrameBuffer(width, height)
	ctx.CameraDepthFBO = fbo
	ctx.CameraDepthTexture = texture
}

func (p *CameraDepthRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.CameraDepthFBO)
	texture := createDepthTexture(width, height)
	ctx.CameraDepthTexture = texture
}

func (p *CameraDepthRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	start := time.Now()
	defer func() {
		globals.ClientRegistry().Inc("render_camera_depth_pass", float64(time.Since(start).Milliseconds()))
	}()

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.CameraDepthFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	p.shader.Use()
	renderGeometryWithoutColor(p.app, p.shader, renderContext.RenderableEntities, viewerContext, renderContext)
}
