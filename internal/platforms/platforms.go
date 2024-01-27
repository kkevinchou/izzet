package platforms

import "github.com/kkevinchou/kitolib/input"

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events to the input collector.
	ProcessEvents(InputCollector)
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the cimgui IO state according to user input (mouse, keyboard, ...)
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
	// ClipboardText returns the current text of the clipboard, if available.
	ClipboardText() (string, error)
	// SetClipboardText sets the text as the current text of the clipboard.
	SetClipboardText(text string)

	SetRelativeMouse(bool)
	MoveMouse(int32, int32)
}

type InputCollector interface {
	SetMousePosition(x float64, y float64)
	SetMouseButtonDown(i int, value bool)
	SetMouseButtonEvent(i int, event input.MouseButtonEvent)
	SetKeyState(key string)
	AddMouseWheelDelta(x float64, y float64)
	AddMouseMotion(x float64, y float64)
}
