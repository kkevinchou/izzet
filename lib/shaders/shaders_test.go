package shaders_test

import (
	"runtime"
	"testing"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/lib/shaders"
	"github.com/veandco/go-sdl2/sdl"
)

func TestShader(t *testing.T) {
	runtime.LockOSThread()

	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()

	window, err := sdl.CreateWindow("hi", 0, 0, 0, 0, sdl.WINDOW_OPENGL)
	if err != nil {
		t.Error(err)
	}
	defer window.Destroy()

	window.GLCreateContext()

	gl.Init()

	shaderManager := shaders.NewShaderManager("testshaders")
	shaderManager.CompileShaderProgram("testshader", "testshader", "testshader")
	shaderProgram := shaderManager.GetShaderProgram("testshader")

	if shaderProgram == nil {
		t.Errorf("failed to create shader program %v", err.Error())
	}

	if shaderProgram.ID == 0 {
		t.Error("shader program created with id 0")
	}
}
