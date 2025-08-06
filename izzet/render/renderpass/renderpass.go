package renderpass

import (
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

/*

Current render passes:
	- drawToShadowDepthMap
		- empty filter
	- drawToCubeDepthMap
		- custom render method
	- drawToCameraDepthMap
	- renderGeometryWithoutColor
		- checks for batch rendering

*/

type TextureFn func() (int, int, []uint32)

// RenderPass is a single step in the frame‐render pipeline.
type RenderPass interface {
	// Init is called once at startup (or when switching pipelines)
	Init(width, height int, ctx *context.RenderPassContext) error

	// Resize is called whenever the viewport changes size
	Resize(width, height int, ctx *context.RenderPassContext)

	// Render executes the pass. It may read from
	// previous-output textures and write into its own FBO.
	Render(ctx context.RenderContext, rctx *context.RenderPassContext, viewerContext context.ViewerContext)
}

func initFrameBuffer(tf TextureFn) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	width, height, textures := tf()
	var drawBuffers []uint32

	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func initFrameBufferNoDepth(tf TextureFn) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	_, _, textures := tf()
	var drawBuffers []uint32

	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func textureFn(width int, height int, internalFormat []int32, format []uint32, xtype []uint32) func() (int, int, []uint32) {
	return func() (int, int, []uint32) {
		count := len(internalFormat)
		var textures []uint32
		for i := 0; i < count; i++ {
			texture := createTexture(width, height, internalFormat[i], format[i], xtype[i], gl.LINEAR)
			attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachment, gl.TEXTURE_2D, texture, 0)

			textures = append(textures, texture)
		}
		return width, height, textures
	}
}

func createTexture(width, height int, internalFormat int32, format uint32, xtype uint32, filtering int32) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat,
		int32(width), int32(height), 0, format, xtype, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return texture
}

func iztDrawElements(app renderiface.App, count int32) {
	app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	app.RuntimeConfig().DrawCount += 1
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, nil)
}

func iztDrawArrays(app renderiface.App, first, count int32) {
	app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	app.RuntimeConfig().DrawCount += 1
	gl.DrawArrays(gl.TRIANGLES, first, count)
}
