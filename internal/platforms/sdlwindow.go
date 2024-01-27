package platforms

import (
	"github.com/veandco/go-sdl2/sdl"
)

type SDLWindow struct {
	window *sdl.Window
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
