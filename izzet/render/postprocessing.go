package render

import "github.com/go-gl/gl/v4.1-core/gl"

func (r *Renderer) postProcess(renderContext RenderContext, texture0 uint32, doPostProcessing bool) uint32 {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.postProcessingFBO)

	shader := r.shaderManager.GetShaderProgram("post_processing")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture0)

	shader.SetUniformInt("image", 0)
	shader.SetUniformInt("doPostProcessing", 1)

	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))

	gl.BindVertexArray(r.xyTextureVAO)
	r.iztDrawArrays(0, 6)

	return r.postProcessingTexture
}
