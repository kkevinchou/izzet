package platforms

import (
	"fmt"
	"math"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/veandco/go-sdl2/sdl"
)

type SDLPlatform struct {
	imguiIO *imgui.IO

	window     *sdl.Window
	shouldStop bool
	time       uint64

	resized bool
	keyMap  map[sdl.Scancode]imgui.Key
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

	platform.setKeyMapping()

	return platform, &SDLWindow{window: window}, nil
}

func (platform *SDLPlatform) ProcessEvents(inputCollector InputCollector) {
	x, y, mouseState := sdl.GetMouseState()
	inputCollector.SetMousePosition(float64(x), float64(y))
	for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
		enabled := mouseState&sdl.Button(button) != 0
		inputCollector.SetMouseButtonState(i, enabled)
	}

	// key state is more reliable than key down events since they dont' fire for every polling cycle every frame
	keyState := sdl.GetKeyboardState()
	for k, v := range keyState {
		if v <= 0 {
			continue
		}
		inputCollector.SetKeyStateEnabled(sdl.GetScancodeName(sdl.Scancode(k)))
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
				inputCollector.SetMouseButtonEvent(i, true)
			}
		}
	case sdl.MOUSEBUTTONUP:
		buttonEvent := event.(*sdl.MouseButtonEvent)
		for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
			if uint32(buttonEvent.Button) == button {
				inputCollector.SetMouseButtonEvent(i, false)
			}
		}
	case sdl.TEXTINPUT:
		inputEvent := event.(*sdl.TextInputEvent)
		platform.imguiIO.AddInputCharactersUTF8(string(inputEvent.Text[:]))
	case sdl.KEYDOWN:
		keyEvent := event.(*sdl.KeyboardEvent)
		platform.addKeyEvent(keyEvent, true)
		inputCollector.AddKeyEvent(sdl.GetScancodeName(keyEvent.Keysym.Scancode), true)
	case sdl.KEYUP:
		keyEvent := event.(*sdl.KeyboardEvent)
		platform.addKeyEvent(keyEvent, false)
		inputCollector.AddKeyEvent(sdl.GetScancodeName(keyEvent.Keysym.Scancode), false)
	case sdl.WINDOWEVENT:
		windowEvent := event.(*sdl.WindowEvent)
		event := windowEvent.Event
		if event == sdl.WINDOWEVENT_RESIZED {
			platform.resized = true
		}
	}
}

func (platform *SDLPlatform) NewFrame() {
	platform.resized = false

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

func (platform *SDLPlatform) setKeyMapping() {
	platform.keyMap = map[sdl.Scancode]imgui.Key{
		sdl.SCANCODE_TAB:   imgui.KeyTab,
		sdl.SCANCODE_LEFT:  imgui.KeyLeftArrow,
		sdl.SCANCODE_RIGHT: imgui.KeyRightArrow,
		sdl.SCANCODE_UP:    imgui.KeyUpArrow,
		sdl.SCANCODE_DOWN:  imgui.KeyDownArrow,

		sdl.SCANCODE_LCTRL:     imgui.ModCtrl,
		sdl.SCANCODE_RCTRL:     imgui.ModCtrl,
		sdl.SCANCODE_LALT:      imgui.ModAlt,
		sdl.SCANCODE_RALT:      imgui.ModAlt,
		sdl.SCANCODE_LSHIFT:    imgui.ModShift,
		sdl.SCANCODE_RSHIFT:    imgui.ModShift,
		sdl.SCANCODE_PAGEUP:    imgui.KeyPageUp,
		sdl.SCANCODE_PAGEDOWN:  imgui.KeyPageDown,
		sdl.SCANCODE_HOME:      imgui.KeyHome,
		sdl.SCANCODE_END:       imgui.KeyEnd,
		sdl.SCANCODE_INSERT:    imgui.KeyInsert,
		sdl.SCANCODE_DELETE:    imgui.KeyDelete,
		sdl.SCANCODE_BACKSPACE: imgui.KeyBackspace,
		sdl.SCANCODE_SPACE:     imgui.KeySpace,
		sdl.SCANCODE_RETURN:    imgui.KeyEnter,
		sdl.SCANCODE_ESCAPE:    imgui.KeyEscape,
	}

	// letters A -> Z
	for i := 0; i < 30; i++ {
		platform.keyMap[sdl.Scancode(uint32(i))] = imgui.Key(542 + i)
	}
}

// key events for imgui actions. these keys powers things like copy/paste from input text, esc to lose focus, etc
func (platform *SDLPlatform) addKeyEvent(keyEvent *sdl.KeyboardEvent, active bool) {
	scanCode := keyEvent.Keysym.Scancode
	if mapped, ok := platform.keyMap[scanCode]; ok {
		platform.imguiIO.AddKeyEvent(mapped, active)
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

func (platform *SDLPlatform) Resized() bool {
	return platform.resized
}
