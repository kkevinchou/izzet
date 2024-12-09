package platforms

import (
	"fmt"
	"math"
	"runtime"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// GLFWClientAPI identifies the render system that shall be initialized.
type GLFWClientAPI string

// This is a list of GLFWClientAPI constants.
const (
	GLFWClientAPIOpenGL2 GLFWClientAPI = "OpenGL2"
	GLFWClientAPIOpenGL3 GLFWClientAPI = "OpenGL3"
)

// GLFW implements a platform based on github.com/go-gl/glfw (v3.2).
type GLFW struct {
	imguiIO *imgui.IO

	window *glfw.Window

	keyMap map[glfw.Key]imgui.Key

	time             float64
	mouseJustPressed [3]bool
}

// NewGLFW attempts to initialize a GLFW context.
func NewGLFW(io *imgui.IO, clientAPI GLFWClientAPI) (*GLFW, error) {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize glfw: %w", err)
	}

	switch clientAPI {
	case GLFWClientAPIOpenGL2:
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
	case GLFWClientAPIOpenGL3:
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 2)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
	default:
		glfw.Terminate()
		return nil, ErrUnsupportedClientAPI
	}

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "CImGui-Go GLFW+"+string(clientAPI)+" example", nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	platform := &GLFW{
		imguiIO: io,
		window:  window,
	}
	platform.setKeyMapping()
	platform.installCallbacks()

	return platform, nil
}

// Dispose cleans up the resources.
func (platform *GLFW) Dispose() {
	platform.window.Destroy()
	glfw.Terminate()
}

// ShouldStop returns true if the window is to be closed.
func (platform *GLFW) ShouldStop() bool {
	return platform.window.ShouldClose()
}

// ProcessEvents handles all pending window events.
func (platform *GLFW) ProcessEvents() {
	glfw.PollEvents()
}

// DisplaySize returns the dimension of the display.
func (platform *GLFW) DisplaySize() [2]float32 {
	w, h := platform.window.GetSize()
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (platform *GLFW) FramebufferSize() [2]float32 {
	w, h := platform.window.GetFramebufferSize()
	return [2]float32{float32(w), float32(h)}
}

// NewFrame marks the begin of a render pass. It forwards all current state to imgui IO.
func (platform *GLFW) NewFrame() {
	// Setup display size (every frame to accommodate for window resizing)
	displaySize := platform.DisplaySize()
	platform.imguiIO.SetDisplaySize(imgui.Vec2{X: displaySize[0], Y: displaySize[1]})

	// Setup time step
	currentTime := glfw.GetTime()
	if platform.time > 0 {
		platform.imguiIO.SetDeltaTime(float32(currentTime - platform.time))
	}
	platform.time = currentTime

	// Setup inputs
	if platform.window.GetAttrib(glfw.Focused) != 0 {
		x, y := platform.window.GetCursorPos()
		platform.imguiIO.SetMousePos(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		platform.imguiIO.SetMousePos(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(platform.mouseJustPressed); i++ {
		down := platform.mouseJustPressed[i] || (platform.window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		platform.imguiIO.SetMouseButtonDown(i, down)
		platform.mouseJustPressed[i] = false
	}
}

// PostRender performs a buffer swap.
func (platform *GLFW) PostRender() {
	platform.window.SwapBuffers()
}

func (platform *GLFW) setKeyMapping() {
	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.

	platform.keyMap = map[glfw.Key]imgui.Key{
		glfw.KeyTab:       imgui.KeyTab,
		glfw.KeyLeft:      imgui.KeyLeftArrow,
		glfw.KeyRight:     imgui.KeyRightArrow,
		glfw.KeyUp:        imgui.KeyUpArrow,
		glfw.KeyDown:      imgui.KeyDownArrow,
		glfw.KeyPageUp:    imgui.KeyPageUp,
		glfw.KeyPageDown:  imgui.KeyPageDown,
		glfw.KeyHome:      imgui.KeyHome,
		glfw.KeyEnd:       imgui.KeyEnd,
		glfw.KeyInsert:    imgui.KeyInsert,
		glfw.KeyDelete:    imgui.KeyDelete,
		glfw.KeyBackspace: imgui.KeyBackspace,
		glfw.KeySpace:     imgui.KeySpace,
		glfw.KeyEnter:     imgui.KeyEnter,
		glfw.KeyEscape:    imgui.KeyEscape,
		glfw.KeyA:         imgui.KeyA,
		glfw.KeyC:         imgui.KeyC,
		glfw.KeyV:         imgui.KeyV,
		glfw.KeyX:         imgui.KeyX,
		glfw.KeyY:         imgui.KeyY,
		glfw.KeyZ:         imgui.KeyZ,

		glfw.KeyLeftControl:  imgui.ModCtrl,
		glfw.KeyRightControl: imgui.ModCtrl,
		glfw.KeyLeftAlt:      imgui.ModAlt,
		glfw.KeyRightAlt:     imgui.ModAlt,
		glfw.KeyLeftSuper:    imgui.ModSuper,
		glfw.KeyRightSuper:   imgui.ModSuper,
	}

}

func (platform *GLFW) installCallbacks() {
	platform.window.SetMouseButtonCallback(platform.mouseButtonChange)
	platform.window.SetScrollCallback(platform.mouseScrollChange)
	platform.window.SetKeyCallback(platform.keyChange)
	platform.window.SetCharCallback(platform.charChange)
}

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: mouseButtonPrimary,
	glfw.MouseButton2: mouseButtonSecondary,
	glfw.MouseButton3: mouseButtonTertiary,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	mouseButtonPrimary:   glfw.MouseButton1,
	mouseButtonSecondary: glfw.MouseButton2,
	mouseButtonTertiary:  glfw.MouseButton3,
}

func (platform *GLFW) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := glfwButtonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		platform.mouseJustPressed[buttonIndex] = true
	}
}

func (platform *GLFW) mouseScrollChange(window *glfw.Window, x, y float64) {
	platform.imguiIO.AddMouseWheelDelta(float32(x), float32(y))
}

func (platform *GLFW) keyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	imKey := imgui.Key(key)
	if mapped, ok := platform.keyMap[key]; ok {
		imKey = mapped
	}

	platform.imguiIO.AddKeyEvent(imKey, action == glfw.Press)
}

func (platform *GLFW) charChange(window *glfw.Window, char rune) {
	platform.imguiIO.AddInputCharactersUTF8(string(char))
	//platform.imguiIO.AddInputCharacters(string(char))
}

// ClipboardText returns the current clipboard text, if available.
func (platform *GLFW) ClipboardText() (string, error) {
	return platform.window.GetClipboardString(), nil
}

// SetClipboardText sets the text as the current clipboard text.
func (platform *GLFW) SetClipboardText(text string) {
	platform.window.SetClipboardString(text)
}
