package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/kitolib/shaders"
)

type Platform interface {
	NewFrame()
	DisplaySize() [2]float32
	FramebufferSize() [2]float32
}

func initOpenGLRenderSettings() {
	defaultSettings()
}

func compileShaders(shaderManager *shaders.ShaderManager) {
	if err := shaderManager.CompileShaderProgram("skybox", "skybox", "skybox", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("modelpbr", "model", "pbr", ""); err != nil {
		panic(err)
	}
	// shader for rendering the depth cubemap for point shadows
	if err := shaderManager.CompileShaderProgram("modelpbr_pointshadow", "model", "pbr", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("model_debug", "model_debug", "pbr_debug", ""); err != nil {
		panic(err)
	}
	// shader for rendering the depth cubemap for point shadows
	if err := shaderManager.CompileShaderProgram("point_shadow", "point_shadow", "point_shadow", "point_shadow"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("model_debug", "model_debug", "pbr_debug", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("depthDebug", "basictexture", "depthvalue", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("flat", "flat", "flat", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("ndc", "ndc", "ndc", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("color_picking", "flat", "picking", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("basic_quad", "basic_quad", "basic_quad", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("basic_quad_world", "basic_quad_world", "basic_quad", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("unit_circle", "unit_circle", "unit_circle", ""); err != nil {
		panic(err)
	}
}

func resetGLRenderSettings() {
	defaultSettings()
	gl.BindVertexArray(0)
	gl.UseProgram(0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func defaultSettings() {
	gl.ClearColor(0.0, 0.5, 0.5, 0.0)
	gl.ClearDepth(1)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
	gl.Enable(gl.MULTISAMPLE)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
	gl.Disable(gl.FRAMEBUFFER_SRGB)
}
