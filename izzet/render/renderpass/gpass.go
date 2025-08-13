package renderpass

import (
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/shaders"
)

const (
	gPassInternalFormat int32  = gl.RGB32F
	gPassFormat         uint32 = gl.RGB
)

type GBufferPass struct {
	app renderiface.App
	sm  *shaders.ShaderManager
}

func NewGPass(app renderiface.App, sm *shaders.ShaderManager) *GBufferPass {
	return &GBufferPass{app: app, sm: sm}
}

func (p *GBufferPass) Init(width, height int, ctx *context.RenderPassContext) {
	fbo, textures := initFrameBuffer(
		width,
		height,
		[]int32{gPassInternalFormat, gPassInternalFormat, gPassInternalFormat},
		[]uint32{gPassFormat, gPassFormat, gPassFormat},
		[]uint32{gl.FLOAT, gl.FLOAT, gl.FLOAT},
		true,
		true,
	)
	ctx.GeometryFBO = fbo
	ctx.GPositionTexture, ctx.GNormalTexture, ctx.GColorTexture = textures[0], textures[1], textures[2]
}

func (p *GBufferPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.GeometryFBO)

	textures := createAndBindTextures(
		width,
		height,
		[]int32{gPassInternalFormat, gPassInternalFormat, gPassInternalFormat},
		[]uint32{gPassFormat, gPassFormat, gPassFormat},
		[]uint32{gl.FLOAT, gl.FLOAT, gl.FLOAT},
		true,
	)

	ctx.GPositionTexture, ctx.GNormalTexture, ctx.GColorTexture = textures[0], textures[1], textures[2]
}
func (p *GBufferPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	mr := globals.ClientRegistry()
	start := time.Now()

	// bind, clear, draw
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.GeometryFBO)
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	drawModels(p.app, p.sm.GetShaderProgram("gpass"), p.sm.GetShaderProgram("gpass"), viewerContext, lightContext, renderContext, renderPassContext, renderContext.RenderableEntities)

	mr.Inc("render_gpass", float64(time.Since(start).Milliseconds()))
}
