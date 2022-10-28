package input

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	mouseButtonCount     = 3
	mouseButtonPrimary   = 0
	mouseButtonSecondary = 1
	mouseButtonTertiary  = 2
)

type SDLPlatform struct {
	imguiIO imgui.IO

	window     *sdl.Window
	shouldStop bool

	time uint64

	currentFrameInput Input
	lastMousePosition [2]int32
}

func NewSDLPlatform(window *sdl.Window, imguiIO imgui.IO) *SDLPlatform {
	platform := &SDLPlatform{
		window:  window,
		imguiIO: imguiIO,
	}
	platform.setKeyMapping()
	return platform
}

func (platform *SDLPlatform) PollInput() Input {
	platform.currentFrameInput = Input{
		MouseInput:        MouseInput{},
		KeyboardInput:     KeyboardInput{},
		CameraOrientation: mgl64.QuatIdent(),
		Commands:          []any{},
	}

	x, y, mouseState := sdl.GetMouseState()
	platform.imguiIO.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
		enabled := mouseState&sdl.Button(button) != 0
		platform.imguiIO.SetMouseButtonDown(i, enabled)
		platform.currentFrameInput.MouseInput.Buttons[i] = enabled
	}

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		platform.processEvent(event)
	}

	// key state is more reliable than key down events since they dont' fire for every polling cycle every frame
	keyState := sdl.GetKeyboardState()
	for k, v := range keyState {
		if v <= 0 {
			continue
		}
		key := KeyboardKey(sdl.GetScancodeName(sdl.Scancode(k)))

		if _, ok := platform.currentFrameInput.KeyboardInput[key]; !ok {
			platform.currentFrameInput.KeyboardInput[key] = KeyState{
				Key:   key,
				Event: KeyboardEventDown,
			}
		}
	}

	// if imgui.IsWindowFocusedV(imgui.FocusedFlagsAnyWindow) {
	// 	return Input{
	// 		MouseInput:    MouseInput{},
	// 		KeyboardInput: KeyboardInput{},
	// 		Commands:      platform.currentFrameInput.Commands,
	// 	}
	// }

	return platform.currentFrameInput
}

func (platform *SDLPlatform) processEvent(event sdl.Event) {
	switch event.GetType() {
	case sdl.QUIT:
		platform.currentFrameInput.Commands = append(platform.currentFrameInput.Commands, QuitCommand{})
	case sdl.MOUSEWHEEL:
		wheelEvent := event.(*sdl.MouseWheelEvent)
		var deltaX, deltaY float32
		if wheelEvent.X > 0 {
			deltaX++
		} else if wheelEvent.X < 0 {
			deltaX--
		}
		if wheelEvent.Y > 0 {
			deltaY++
		} else if wheelEvent.Y < 0 {
			deltaY--
		}
		platform.imguiIO.AddMouseWheelDelta(deltaX, deltaY)
		platform.currentFrameInput.MouseInput.MouseWheelDelta = int(deltaY)
	case sdl.MOUSEMOTION:
		motionEvent := event.(*sdl.MouseMotionEvent)
		platform.currentFrameInput.MouseInput.MouseMotionEvent.XRel += float64(motionEvent.XRel)
		platform.currentFrameInput.MouseInput.MouseMotionEvent.YRel += float64(motionEvent.YRel)
	case sdl.MOUSEBUTTONDOWN:
		buttonEvent := event.(*sdl.MouseButtonEvent)
		switch buttonEvent.Button {
		case sdl.BUTTON_RIGHT:
			platform.lastMousePosition[0] = buttonEvent.X
			platform.lastMousePosition[1] = buttonEvent.Y
			sdl.SetRelativeMouseMode(true)
		}
	case sdl.MOUSEBUTTONUP:
		buttonEvent := event.(*sdl.MouseButtonEvent)
		switch buttonEvent.Button {
		case sdl.BUTTON_RIGHT:
			sdl.SetRelativeMouseMode(false)
			sdl.GetMouseFocus().WarpMouseInWindow(platform.lastMousePosition[0], platform.lastMousePosition[1])
		}
	case sdl.TEXTINPUT:
		inputEvent := event.(*sdl.TextInputEvent)
		platform.imguiIO.AddInputCharacters(string(inputEvent.Text[:]))
	case sdl.KEYDOWN:
		keyEvent := event.(*sdl.KeyboardEvent)
		platform.imguiIO.KeyPress(int(keyEvent.Keysym.Scancode))
		platform.updateKeyModifier()
	case sdl.KEYUP:
		keyEvent := event.(*sdl.KeyboardEvent)
		platform.imguiIO.KeyRelease(int(keyEvent.Keysym.Scancode))
		platform.updateKeyModifier()

		key := KeyboardKey(sdl.GetScancodeName(keyEvent.Keysym.Scancode))
		platform.currentFrameInput.KeyboardInput[key] = KeyState{
			Key:   key,
			Event: KeyboardEventUp,
		}
	}
}

func (platform *SDLPlatform) NewFrame() {
	// Setup display size (every frame to accommodate for window resizing)
	displaySize := platform.DisplaySize()
	platform.imguiIO.SetDisplaySize(imgui.Vec2{X: displaySize[0], Y: displaySize[1]})

	// Setup time step (we don't use SDL_GetTicks() because it is using millisecond resolution)
	frequency := sdl.GetPerformanceFrequency()
	currentTime := sdl.GetPerformanceCounter()
	if platform.time > 0 {
		platform.imguiIO.SetDeltaTime(float32(currentTime-platform.time) / float32(frequency))
	} else {
		const fallbackDelta = 1.0 / 60.0
		platform.imguiIO.SetDeltaTime(fallbackDelta)
	}
	platform.time = currentTime
}

// DisplaySize returns the dimension of the display.
func (platform *SDLPlatform) DisplaySize() [2]float32 {
	w, h := platform.window.GetSize()
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (platform *SDLPlatform) FramebufferSize() [2]float32 {
	w, h := platform.window.GLGetDrawableSize()
	return [2]float32{float32(w), float32(h)}
}

func (platform *SDLPlatform) updateKeyModifier() {
	modState := sdl.GetModState()
	mapModifier := func(lMask sdl.Keymod, lKey int, rMask sdl.Keymod, rKey int) (lResult int, rResult int) {
		if (modState & lMask) != 0 {
			lResult = lKey
		}
		if (modState & rMask) != 0 {
			rResult = rKey
		}
		return
	}
	platform.imguiIO.KeyShift(mapModifier(sdl.KMOD_LSHIFT, sdl.SCANCODE_LSHIFT, sdl.KMOD_RSHIFT, sdl.SCANCODE_RSHIFT))
	platform.imguiIO.KeyCtrl(mapModifier(sdl.KMOD_LCTRL, sdl.SCANCODE_LCTRL, sdl.KMOD_RCTRL, sdl.SCANCODE_RCTRL))
	platform.imguiIO.KeyAlt(mapModifier(sdl.KMOD_LALT, sdl.SCANCODE_LALT, sdl.KMOD_RALT, sdl.SCANCODE_RALT))
}

func (platform *SDLPlatform) setKeyMapping() {
	keys := map[int]int{
		imgui.KeyTab:        sdl.SCANCODE_TAB,
		imgui.KeyLeftArrow:  sdl.SCANCODE_LEFT,
		imgui.KeyRightArrow: sdl.SCANCODE_RIGHT,
		imgui.KeyUpArrow:    sdl.SCANCODE_UP,
		imgui.KeyDownArrow:  sdl.SCANCODE_DOWN,
		imgui.KeyPageUp:     sdl.SCANCODE_PAGEUP,
		imgui.KeyPageDown:   sdl.SCANCODE_PAGEDOWN,
		imgui.KeyHome:       sdl.SCANCODE_HOME,
		imgui.KeyEnd:        sdl.SCANCODE_END,
		imgui.KeyInsert:     sdl.SCANCODE_INSERT,
		imgui.KeyDelete:     sdl.SCANCODE_DELETE,
		imgui.KeyBackspace:  sdl.SCANCODE_BACKSPACE,
		imgui.KeySpace:      sdl.SCANCODE_SPACE,
		imgui.KeyEnter:      sdl.SCANCODE_RETURN,
		imgui.KeyEscape:     sdl.SCANCODE_ESCAPE,
	}

	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
	for imguiKey, nativeKey := range keys {
		platform.imguiIO.KeyMap(imguiKey, nativeKey)
	}
}

// ClipboardText returns the current clipboard text, if available.
func (platform *SDLPlatform) ClipboardText() (string, error) {
	return sdl.GetClipboardText()
}

// SetClipboardText sets the text as the current clipboard text.
func (platform *SDLPlatform) SetClipboardText(text string) {
	_ = sdl.SetClipboardText(text)
}
