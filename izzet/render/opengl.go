package render

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/kkevinchou/kitolib/shaders"
)

type Platform interface {
	NewFrame()
	DisplaySize() [2]float32
	FramebufferSize() [2]float32
}

func initOpenGLRenderSettings() {
	// gl.ClearColor(0.0, 0.5, 0.5, 0.0)
	gl.ClearColor(1, 1, 1, 1)
	gl.ClearDepth(1)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.FRAMEBUFFER_SRGB)
}

func compileShaders(shaderManager *shaders.ShaderManager) {
	if err := shaderManager.CompileShaderProgram("skybox", "skybox", "skybox", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("modelpbr", "model", "pbr", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("navmesh", "navmesh", "navmesh", ""); err != nil {
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
	if err := shaderManager.CompileShaderProgram("unit_circle", "unit_circle", "unit_circle", ""); err != nil {
		panic(err)
	}

	// Geometry only shader - does not calculate lighting, useful for depth map calculations

	if err := shaderManager.CompileShaderProgram("modelgeo", "modelgeo", "modelgeo", ""); err != nil {
		panic(err)
	}

	// Quad rendering

	if err := shaderManager.CompileShaderProgram("screen_space_quad", "screen_space_quad", "textured_picking", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("world_space_quad", "world_space_quad", "textured_picking", ""); err != nil {
		panic(err)
	}

	// Bloom Shaders

	if err := shaderManager.CompileShaderProgram("blend", "composite", "blend", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("composite", "composite", "composite", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("bloom_downsample", "bloom", "bloom_downsample", ""); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("bloom_upsample", "bloom", "bloom_upsample", ""); err != nil {
		panic(err)
	}

	// shader for rendering the depth cubemap for point shadows
	if err := shaderManager.CompileShaderProgram("point_shadow", "point_shadow", "point_shadow", "point_shadow"); err != nil {
		panic(err)
	}
}
