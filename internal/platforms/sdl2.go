package platforms

import (
	"fmt"
	"math"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/kitolib/input"
	"github.com/veandco/go-sdl2/sdl"
)

type SDLPlatform struct {
	imguiIO *imgui.IO

	window     *sdl.Window
	shouldStop bool
	time       uint64
}

func NewSDLPlatform(imguiIO *imgui.IO) (*SDLPlatform, *SDLWindow, error) {
	window, err := InitSDL()
	if err != nil {
		return nil, nil, err
	}

	platform := &SDLPlatform{
		window:  window,
		imguiIO: imguiIO,
	}
	// platform.setKeyMapping()
	return platform, &SDLWindow{window: window}, nil
}

func (platform *SDLPlatform) ProcessEvents(inputCollector InputCollector) {
	// platform.currentFrameInput = Input{
	// 	WindowEvent:    WindowEvent{},
	// 	MouseInput:     MouseInput{},
	// 	KeyboardInput:  KeyboardInput{},
	// 	CameraRotation: mgl64.QuatIdent(),
	// 	Commands:       []any{},
	// }

	x, y, mouseState := sdl.GetMouseState()
	inputCollector.SetMousePosition(float64(x), float64(y))
	for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
		enabled := mouseState&sdl.Button(button) != 0
		inputCollector.SetMouseButtonDown(i, enabled)
	}

	// key state is more reliable than key down events since they dont' fire for every polling cycle every frame
	keyState := sdl.GetKeyboardState()
	for k, v := range keyState {
		if v <= 0 {
			continue
		}

		inputCollector.SetKeyState(sdl.GetScancodeName(sdl.Scancode(k)))
	}

	// return platform.currentFrameInput
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		platform.processEvent(event, inputCollector)
	}
}

func (platform *SDLPlatform) processEvent(event sdl.Event, inputCollector InputCollector) {
	switch event.GetType() {
	case sdl.DROPFILE:
		break
	case sdl.QUIT:
		platform.shouldStop = true
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
		inputCollector.AddMouseWheelDelta(float64(deltaX), float64(deltaY))
	case sdl.MOUSEMOTION:
		motionEvent := event.(*sdl.MouseMotionEvent)
		inputCollector.AddMouseMotion(float64(motionEvent.XRel), float64(motionEvent.YRel))
	case sdl.MOUSEBUTTONDOWN:
		buttonEvent := event.(*sdl.MouseButtonEvent)
		for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
			if uint32(buttonEvent.Button) == button {
				inputCollector.SetMouseButtonEvent(i, input.MouseButtonEventDown)
				inputCollector.SetMouseButtonDown(i, true)
			}
		}
	case sdl.MOUSEBUTTONUP:
		buttonEvent := event.(*sdl.MouseButtonEvent)
		for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
			if uint32(buttonEvent.Button) == button {
				inputCollector.SetMouseButtonEvent(i, input.MouseButtonEventUp)
				inputCollector.SetMouseButtonDown(i, false)
			}
		}
		// case sdl.TEXTINPUT:
		// 	inputEvent := event.(*sdl.TextInputEvent)
		// 	platform.imguiIO.AddInputCharacters(string(inputEvent.Text[:]))
		// case sdl.KEYDOWN:
		// 	keyEvent := event.(*sdl.KeyboardEvent)
		// 	platform.imguiIO.KeyPress(int(keyEvent.Keysym.Scancode))
		// 	platform.updateKeyModifier()
		// case sdl.KEYUP:
		// 	keyEvent := event.(*sdl.KeyboardEvent)
		// 	platform.imguiIO.KeyRelease(int(keyEvent.Keysym.Scancode))
		// 	platform.updateKeyModifier()

		// 	key := KeyboardKey(sdl.GetScancodeName(keyEvent.Keysym.Scancode))
		// 	platform.currentFrameInput.KeyboardInput[key] = KeyState{
		// 		Key:   key,
		// 		Event: KeyboardEventUp,
		// 	}
		// case sdl.WINDOWEVENT:
		// 	windowEvent := event.(*sdl.WindowEvent)
		// 	event := windowEvent.Event
		// 	if event == sdl.WINDOWEVENT_RESIZED {
		// 		platform.currentFrameInput.WindowEvent.Resized = true
		// 	}
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

	// Setup inputs
	x, y, state := sdl.GetMouseState()
	if platform.window.GetFlags()&sdl.WINDOW_INPUT_FOCUS != 0 {
		platform.imguiIO.SetMousePos(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		platform.imguiIO.SetMousePos(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	down := state&sdl.ButtonLMask() != 0
	platform.imguiIO.SetMouseButtonDown(0, down)
	down = state&sdl.ButtonRMask() != 0
	platform.imguiIO.SetMouseButtonDown(1, down)
	down = state&sdl.ButtonMMask() != 0
	platform.imguiIO.SetMouseButtonDown(2, down)
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

	mapModifier := func(lMask sdl.Keymod, rMask sdl.Keymod) bool {
		if (modState&lMask) != 0 || (modState&rMask) != 0 {
			return true
		}
		return false
	}

	platform.imguiIO.SetKeyShift(mapModifier(sdl.KMOD_LSHIFT, sdl.KMOD_RSHIFT))
	platform.imguiIO.SetKeyCtrl(mapModifier(sdl.KMOD_LCTRL, sdl.KMOD_RCTRL))
	platform.imguiIO.SetKeyAlt(mapModifier(sdl.KMOD_LALT, sdl.KMOD_RALT))
}

// func (platform *SDLPlatform) setKeyMapping() {
// 	keys := map[int]int{
// 		imgui.KeyTab:        sdl.SCANCODE_TAB,
// 		imgui.KeyLeftArrow:  sdl.SCANCODE_LEFT,
// 		imgui.KeyRightArrow: sdl.SCANCODE_RIGHT,
// 		imgui.KeyUpArrow:    sdl.SCANCODE_UP,
// 		imgui.KeyDownArrow:  sdl.SCANCODE_DOWN,
// 		imgui.KeyPageUp:     sdl.SCANCODE_PAGEUP,
// 		imgui.KeyPageDown:   sdl.SCANCODE_PAGEDOWN,
// 		imgui.KeyHome:       sdl.SCANCODE_HOME,
// 		imgui.KeyEnd:        sdl.SCANCODE_END,
// 		imgui.KeyInsert:     sdl.SCANCODE_INSERT,
// 		imgui.KeyDelete:     sdl.SCANCODE_DELETE,
// 		imgui.KeyBackspace:  sdl.SCANCODE_BACKSPACE,
// 		imgui.KeySpace:      sdl.SCANCODE_SPACE,
// 		imgui.KeyEnter:      sdl.SCANCODE_RETURN,
// 		imgui.KeyEscape:     sdl.SCANCODE_ESCAPE,
// 	}

// 	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
// 	for imguiKey, nativeKey := range keys {
// 		platform.imguiIO.KeyMap(imguiKey, nativeKey)
// 	}
// }

// ClipboardText returns the current clipboard text, if available.
func (platform *SDLPlatform) ClipboardText() (string, error) {
	return sdl.GetClipboardText()
}

// SetClipboardText sets the text as the current clipboard text.
func (platform *SDLPlatform) SetClipboardText(text string) {
	_ = sdl.SetClipboardText(text)
}

func (platform *SDLPlatform) SetRelativeMouse(value bool) {
	sdl.SetRelativeMouseMode(value)
}

func (platform *SDLPlatform) MoveMouse(x, y int32) {
	platform.window.WarpMouseInWindow(x, y)
}

func InitSDL() (*sdl.Window, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, fmt.Errorf("failed to init SDL %s", err)
	}

	// Enable hints for multisampling which allows opengl to use the default
	// multisampling algorithms implemented by the OpenGL rasterizer
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)

	// sdl.GLSetAttribute(sdl.GL_RED_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_GREEN_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_BLUE_SIZE, 10)
	// sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 2)

	sdl.SetRelativeMouseMode(false)

	windowFlags := sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE
	// if config.Fullscreen {
	// 	dm, err := sdl.GetCurrentDisplayMode(0)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	config.Width = int(dm.W)
	// 	config.Height = int(dm.H)
	// 	// windowFlags |= sdl.WINDOW_MAXIMIZED
	// 	windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
	// 	// windowFlags |= sdl.WINDOW_FULLSCREEN
	// }

	win, err := sdl.CreateWindow("IZZET GAME ENGINE", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowWidth, windowHeight, uint32(windowFlags))
	if err != nil {
		return nil, fmt.Errorf("failed to create window %s", err)
	}

	_, err = win.GLCreateContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context %s", err)
	}

	return win, nil
}

func (platform *SDLPlatform) PostRender() {
	platform.window.GLSwap()
}

func (platform *SDLPlatform) ShouldStop() bool {
	return platform.shouldStop
}
