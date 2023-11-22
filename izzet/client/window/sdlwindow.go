package window

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/veandco/go-sdl2/sdl"
)

type SDLWindow struct {
	window *sdl.Window
}

// TODO
// we should not be returning an *sdl.Window here, but the sdl platform code
// expects one and I'm too lazy to refactor this at the moment.
func NewSDLWindow(config settings.Config) (*SDLWindow, *sdl.Window, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, nil, fmt.Errorf("failed to init SDL %s", err)
	}

	// Enable hints for multisampling which allows opengl to use the default
	// multisampling algorithms implemented by the OpenGL rasterizer
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)

	// sdl.GLSetAttribute(sdl.GL_RED_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_GREEN_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_BLUE_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 2)

	sdl.SetRelativeMouseMode(false)

	windowFlags := sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE
	if config.Fullscreen {
		dm, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			panic(err)
		}
		config.Width = int(dm.W)
		config.Height = int(dm.H)
		// windowFlags |= sdl.WINDOW_MAXIMIZED
		windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
		// windowFlags |= sdl.WINDOW_FULLSCREEN
	}

	win, err := sdl.CreateWindow("IZZET GAME ENGINE", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(config.Width), int32(config.Height), uint32(windowFlags))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create window %s", err)
	}

	_, err = win.GLCreateContext()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create context %s", err)
	}

	return &SDLWindow{window: win}, win, nil
}

func (w *SDLWindow) Minimized() bool {
	return w.window.GetFlags()&sdl.WINDOW_MINIMIZED > 0
}

func (w *SDLWindow) GetSize() (int, int) {
	width, height := w.window.GetSize()
	return int(width), int(height)
}

func (w *SDLWindow) Swap() {
	w.window.GLSwap()
}
func (w *SDLWindow) WindowFocused() bool {
	return w.window.GetFlags()&sdl.WINDOW_INPUT_FOCUS > 0
}
