package izzet

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/veandco/go-sdl2/sdl"
)

type Platform interface {
	NewFrame()
	DisplaySize() [2]float32
	FramebufferSize() [2]float32
}

func initOpenGLRenderSettings() {
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

func initializeOpenGL() (*sdl.Window, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, fmt.Errorf("failed to init SDL %s", err)
	}

	// Enable hints for multisampling which allows opengl to use the default
	// multisampling algorithms implemented by the OpenGL rasterizer
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)
	sdl.SetRelativeMouseMode(false)

	windowFlags := sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE
	if settings.Fullscreen {
		dm, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			panic(err)
		}
		settings.Width = int(dm.W)
		settings.Height = int(dm.H)
		windowFlags |= sdl.WINDOW_MAXIMIZED
	}
	window, err := sdl.CreateWindow("IZZET GAME ENGINE", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(settings.Width), int32(settings.Height), uint32(windowFlags))
	if err != nil {
		return nil, fmt.Errorf("failed to create window %s", err)
	}

	_, err = window.GLCreateContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context %s", err)
	}

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to init OpenGL %s", err)
	}

	fmt.Println("Open GL Version:", gl.GoStr(gl.GetString(gl.VERSION)))

	return window, nil
}

func compileShaders(shaderManager *shaders.ShaderManager) {
	if err := shaderManager.CompileShaderProgram("skybox", "skybox", "skybox"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("modelpbr", "model", "pbr"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("model_debug", "model_debug", "pbr_debug"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("model_static", "model_static", "pbr"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("depthDebug", "basictexture", "depthvalue"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("flat", "flat", "flat"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("ndc", "ndc", "ndc"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("color_picking", "model_static", "picking"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("basic_quad", "basic_quad", "basic_quad"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("basic_quad_world", "basic_quad_world", "basic_quad"); err != nil {
		panic(err)
	}
	if err := shaderManager.CompileShaderProgram("unit_circle", "unit_circle", "unit_circle"); err != nil {
		panic(err)
	}
}

func resetGLRenderSettings() {
	gl.BindVertexArray(0)
	gl.UseProgram(0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.CullFace(gl.BACK)
	gl.Enable(gl.BLEND)
}
