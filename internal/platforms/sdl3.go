package platforms

import (
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/Zyko0/go-sdl3/bin/binmix"
	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
	"github.com/Zyko0/go-sdl3/mixer"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var audioMixer *mixer.Mixer

var mouseButtonOrder = []sdl.MouseButtonFlags{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE}

type SDLPlatform struct {
	imguiIO *imgui.IO

	window     *sdl.Window
	shouldStop bool
	time       uint64

	keyMap map[sdl.Scancode]imgui.Key
}

func LoadSDL3Libraries() {
	_ = binsdl.Load()
	_ = binttf.Load()
	_ = binmix.Load()

	// return func() {
	// 	if audioMixer != nil {
	// 		audioMixer.Destroy()
	// 		audioMixer = nil
	// 	}
	// 	mixer.Quit()
	// 	ttf.Quit()
	// 	sdl.Quit()

	// 	if err := mixer.CloseLibrary(); err != nil {
	// 		panic(fmt.Errorf("close SDL3_mixer library: %w", err))
	// 	}
	// 	if err := ttf.CloseLibrary(); err != nil {
	// 		panic(fmt.Errorf("close SDL3_ttf library: %w", err))
	// 	}
	// 	sdlLib.Unload()
	// }
}

func AudioMixer() *mixer.Mixer {
	return audioMixer
}

func NewSDLPlatform(width, height int, fullscreen bool) (*SDLPlatform, *SDLWindow, error) {
	LoadSDL3Libraries()
	// defer shutdownSDL3()

	imgui.CreateContext()
	imguiIO := imgui.CurrentIO()
	imgui.CurrentIO().Fonts().AddFontFromFileTTF("_assets/fonts/roboto-regular.ttf", settings.FontSize)

	window, err := initSDL(width, height, fullscreen)
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
	mouseState, x, y := sdl.GetMouseState()
	inputCollector.SetMousePosition(float64(x), float64(y))
	for i, button := range mouseButtonOrder {
		enabled := mouseState&sdl.ButtonMask(button) != 0
		inputCollector.SetMouseButtonState(i, enabled)
	}

	// key state is more reliable than key down events since they dont' fire for every polling cycle every frame
	keyState := sdl.GetKeyboardState()
	for k, v := range keyState {
		if !v {
			continue
		}
		inputCollector.SetKeyStateEnabled(sdl.Scancode(k).Name())
	}

	var event sdl.Event
	for sdl.PollEvent(&event) {
		platform.processEvent(&event, inputCollector)
	}
}

func (platform *SDLPlatform) processEvent(event *sdl.Event, inputCollector InputCollector) {
	switch event.Type {
	case sdl.EVENT_DROP_FILE:
		break
	case sdl.EVENT_QUIT:
		platform.shouldStop = true
	case sdl.EVENT_MOUSE_WHEEL:
		wheelEvent := event.MouseWheelEvent()
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
	case sdl.EVENT_MOUSE_MOTION:
		motionEvent := event.MouseMotionEvent()
		inputCollector.AddMouseMotion(float64(motionEvent.Xrel), float64(motionEvent.Yrel))
	case sdl.EVENT_MOUSE_BUTTON_DOWN:
		buttonEvent := event.MouseButtonEvent()
		for i, button := range mouseButtonOrder {
			if buttonEvent.Button == uint8(button) {
				inputCollector.SetMouseButtonEvent(i, true)
			}
		}
	case sdl.EVENT_MOUSE_BUTTON_UP:
		buttonEvent := event.MouseButtonEvent()
		for i, button := range mouseButtonOrder {
			if buttonEvent.Button == uint8(button) {
				inputCollector.SetMouseButtonEvent(i, false)
			}
		}
	case sdl.EVENT_TEXT_INPUT:
		inputEvent := event.TextInputEvent()
		platform.imguiIO.AddInputCharactersUTF8(inputEvent.Text)
	case sdl.EVENT_KEY_DOWN:
		keyEvent := event.KeyboardEvent()
		if keyEvent.Repeat {
			return
		}
		platform.addKeyEvent(keyEvent, true)
		inputCollector.AddKeyEvent(keyEvent.Scancode.Name(), true)
	case sdl.EVENT_KEY_UP:
		keyEvent := event.KeyboardEvent()
		platform.addKeyEvent(keyEvent, false)
		inputCollector.AddKeyEvent(keyEvent.Scancode.Name(), false)
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
	state, x, y := sdl.GetMouseState()
	if platform.window.Flags()&sdl.WINDOW_INPUT_FOCUS != 0 {
		platform.imguiIO.SetMousePos(imgui.Vec2{X: x, Y: y})
	} else {
		platform.imguiIO.SetMousePos(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	down := state&sdl.ButtonMask(sdl.BUTTON_LEFT) != 0
	platform.imguiIO.SetMouseButtonDown(0, down)
	down = state&sdl.ButtonMask(sdl.BUTTON_RIGHT) != 0
	platform.imguiIO.SetMouseButtonDown(1, down)
	down = state&sdl.ButtonMask(sdl.BUTTON_MIDDLE) != 0
	platform.imguiIO.SetMouseButtonDown(2, down)
}

// DisplaySize returns the dimension of the display.
func (platform *SDLPlatform) DisplaySize() [2]float32 {
	w, h, err := platform.window.Size()
	if err != nil {
		panic(fmt.Errorf("get SDL window size: %w", err))
	}
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (platform *SDLPlatform) FramebufferSize() [2]float32 {
	w, h, err := platform.window.SizeInPixels()
	if err != nil {
		panic(fmt.Errorf("get SDL window pixel size: %w", err))
	}
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
	scanCode := keyEvent.Scancode
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
	_ = platform.window.SetRelativeMouseMode(value)
}

func (platform *SDLPlatform) MoveMouse(x, y int32) {
	platform.window.WarpMouseIn(float32(x), float32(y))
}

const fullscreenMode = sdl.WINDOW_FULLSCREEN

func (platform *SDLPlatform) Fullscreen() bool {
	return platform.window.Flags()&fullscreenMode != 0
}

func (platform *SDLPlatform) SetFullscreen(fullscreen bool) error {
	return platform.window.SetFullscreen(fullscreen)
}

func initSDL(width, height int, fullscreen bool) (*sdl.Window, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_EVENTS); err != nil {
		return nil, fmt.Errorf("failed to init SDL %s", err)
	}

	if err := mixer.Init(); err != nil {
		return nil, fmt.Errorf("failed to init SDL mixer: %w", err)
	}

	var err error
	audioMixer, err = mixer.CreateMixerDevice(sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK, &sdl.AudioSpec{
		Format:   sdl.AUDIO_S16,
		Channels: 2,
		Freq:     48000,
	})
	if err != nil {
		return nil, fmt.Errorf("create mixer audio device: %w", err)
	}

	// Enable hints for multisampling which allows opengl to use the default
	// multisampling algorithms implemented by the OpenGL rasterizer
	glAttributes := []struct {
		attr  sdl.GLAttr
		value int32
	}{
		{sdl.GL_MULTISAMPLEBUFFERS, 1},
		{sdl.GL_MULTISAMPLESAMPLES, 4},
		{sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE},
		{sdl.GL_CONTEXT_MAJOR_VERSION, 4},
		{sdl.GL_CONTEXT_MINOR_VERSION, 3},
	}
	for _, glAttribute := range glAttributes {
		if err := sdl.GL_SetAttribute(glAttribute.attr, glAttribute.value); err != nil {
			return nil, fmt.Errorf("set SDL GL attribute %d: %w", glAttribute.attr, err)
		}
	}

	windowFlags := sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE
	if fullscreen {
		windowFlags |= fullscreenMode
	} else {
		windowFlags |= sdl.WINDOW_MAXIMIZED
	}

	win, err := sdl.CreateWindow("IZZET GAME ENGINE", width, height, windowFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to create window %s", err)
	}

	if _, err = sdl.GL_CreateContext(win); err != nil {
		return nil, fmt.Errorf("failed to create context %s", err)
	}

	if err := win.SetRelativeMouseMode(false); err != nil {
		return nil, fmt.Errorf("set relative mouse mode: %w", err)
	}

	return win, nil
}

func (platform *SDLPlatform) PostRender() {
	if err := sdl.GL_SwapWindow(platform.window); err != nil {
		panic(fmt.Errorf("swap SDL GL window: %w", err))
	}
}

func (platform *SDLPlatform) ShouldStop() bool {
	return platform.shouldStop
}

type SDLWindow struct {
	window *sdl.Window
}

func (w *SDLWindow) Minimized() bool {
	return w.window.Flags()&sdl.WINDOW_MINIMIZED > 0
}

func (w *SDLWindow) GetSize() (int, int) {
	width, height, err := w.window.Size()
	if err != nil {
		panic(fmt.Errorf("get SDL window size: %w", err))
	}
	return int(width), int(height)
}

func (w *SDLWindow) Swap() {
	if err := sdl.GL_SwapWindow(w.window); err != nil {
		panic(fmt.Errorf("swap SDL GL window: %w", err))
	}
}

func (w *SDLWindow) WindowFocused() bool {
	return w.window.Flags()&sdl.WINDOW_INPUT_FOCUS > 0
}
