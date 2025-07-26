package input

import "github.com/go-gl/mathgl/mgl64"

type InputPoller func() Input

type MouseWheelDirection int

type MouseMotionEvent struct {
	XRel float64
	YRel float64
}

type MouseButtonEvent string

var MouseButtonEventNone MouseButtonEvent = ""
var MouseButtonEventDown MouseButtonEvent = "DOWN"
var MouseButtonEventUp MouseButtonEvent = "UP"

func (m MouseMotionEvent) IsZero() bool {
	return m.XRel == 0 && m.YRel == 0
}

type MouseInput struct {
	Position         mgl64.Vec2
	MouseWheelDelta  int
	MouseMotionEvent MouseMotionEvent
	MouseButtonEvent [3]MouseButtonEvent
	MouseButtonState [3]bool // left, right, middle
}

type KeyboardKey string
type KeyboardEvent int

const (
	KeyboardKeyA KeyboardKey = "A"
	KeyboardKeyB KeyboardKey = "B"
	KeyboardKeyC KeyboardKey = "C"
	KeyboardKeyD KeyboardKey = "D"
	KeyboardKeyE KeyboardKey = "E"
	KeyboardKeyF KeyboardKey = "F"
	KeyboardKeyG KeyboardKey = "G"
	KeyboardKeyH KeyboardKey = "H"
	KeyboardKeyI KeyboardKey = "I"
	KeyboardKeyJ KeyboardKey = "J"
	KeyboardKeyK KeyboardKey = "K"
	KeyboardKeyL KeyboardKey = "L"
	KeyboardKeyM KeyboardKey = "M"
	KeyboardKeyN KeyboardKey = "N"
	KeyboardKeyO KeyboardKey = "O"
	KeyboardKeyP KeyboardKey = "P"
	KeyboardKeyQ KeyboardKey = "Q"
	KeyboardKeyR KeyboardKey = "R"
	KeyboardKeyS KeyboardKey = "S"
	KeyboardKeyT KeyboardKey = "T"
	KeyboardKeyU KeyboardKey = "U"
	KeyboardKeyV KeyboardKey = "V"
	KeyboardKeyW KeyboardKey = "W"
	KeyboardKeyX KeyboardKey = "X"
	KeyboardKeyY KeyboardKey = "Y"
	KeyboardKeyZ KeyboardKey = "Z"

	KeyboardKeyUp    KeyboardKey = "Up"
	KeyboardKeyDown  KeyboardKey = "Down"
	KeyboardKeyLeft  KeyboardKey = "Left"
	KeyboardKeyRight KeyboardKey = "Right"

	KeyboardKeyLShift KeyboardKey = "Left Shift"
	KeyboardKeyLCtrl  KeyboardKey = "Left Ctrl"
	KeyboardKeyLAlt   KeyboardKey = "Left Alt"
	KeyboardKeyRShift KeyboardKey = "Right Shift"
	KeyboardKeyRCtrl  KeyboardKey = "Right Ctrl"
	KeyboardKeyRAlt   KeyboardKey = "Right Alt"
	KeyboardKeySpace  KeyboardKey = "Space"
	KeyboardKeyEscape KeyboardKey = "Escape"

	KeyboardKeyTick KeyboardKey = "`"
	KeyboardKeyF1   KeyboardKey = "F1"
	KeyboardKeyF2   KeyboardKey = "F2"
	KeyboardKeyF3   KeyboardKey = "F3"
	KeyboardKeyF4   KeyboardKey = "F4"
	KeyboardKeyF5   KeyboardKey = "F5"
	KeyboardKeyF6   KeyboardKey = "F6"
	KeyboardKeyF7   KeyboardKey = "F7"
	KeyboardKeyF8   KeyboardKey = "F8"
	KeyboardKeyF9   KeyboardKey = "F9"
	KeyboardKeyF10  KeyboardKey = "F10"

	KeyboardEventUp = iota
	KeyboardEventDown
	KeyboardEventNone
)

type KeyState struct {
	Key   KeyboardKey
	Event KeyboardEvent
}

type KeyboardInput map[KeyboardKey]KeyState

type QuitCommand struct {
}

type FileDropCommand struct {
	File string
}

type WindowEvent struct {
	Resized bool
}

// Input represents the input provided by a user during a command frame
// Input should be only constructed by the input poller and should not be
// written to by any systems, only read. Input is stored in a client side
// command frame history which will copy the KeyboardInput by reference
type Input struct {
	WindowEvent    WindowEvent
	KeyboardInput  KeyboardInput
	MouseInput     MouseInput
	CameraRotation mgl64.Quat
	Commands       []any
}

// func (i Input) Copy() Input {
// 	keyboardInput := KeyboardInput{}
// 	for k, v := range i.KeyboardInput {
// 		keyboardInput[k] = v
// 	}

// 	return Input{
// 		KeyboardInput: keyboardInput,
// 		MouseInput:    i.MouseInput,
// 	}
// }

type InputCollector struct {
	MousePosition    [2]float64
	MouseButtonState [3]bool
	KeyboardInput    KeyboardInput
	MouseWheelDelta  int
	MouseMotionEvent MouseMotionEvent
	MouseButtonEvent [3]MouseButtonEvent
}

func NewInputCollector() *InputCollector {
	return &InputCollector{
		KeyboardInput: KeyboardInput{},
	}
}

func (i *InputCollector) SetMousePosition(x float64, y float64) {
	i.MousePosition[0] = x
	i.MousePosition[1] = y
}

func (i *InputCollector) SetMouseButtonEvent(index int, down bool) {
	if down {
		i.MouseButtonEvent[index] = MouseButtonEventDown
	} else {
		i.MouseButtonEvent[index] = MouseButtonEventUp
	}
}
func (i *InputCollector) SetMouseButtonState(index int, value bool) {
	i.MouseButtonState[index] = value
}

func (i *InputCollector) SetKeyStateEnabled(key string) {
	iKey := KeyboardKey(key)
	if _, ok := i.KeyboardInput[iKey]; !ok {
		i.KeyboardInput[iKey] = KeyState{
			Key:   iKey,
			Event: KeyboardEventNone,
		}
	}
}

func (i *InputCollector) AddKeyEvent(key string, down bool) {
	iKey := KeyboardKey(key)
	event := KeyboardEventDown
	if !down {
		event = KeyboardEventUp
	}
	i.KeyboardInput[iKey] = KeyState{
		Key:   iKey,
		Event: KeyboardEvent(event),
	}
}

func (i *InputCollector) AddMouseWheelDelta(x float64, y float64) {
	i.MouseWheelDelta += int(y)
}

func (i *InputCollector) AddMouseMotion(x float64, y float64) {
	i.MouseMotionEvent.XRel += x
	i.MouseMotionEvent.YRel += y
}

func (i *InputCollector) GetInput() Input {
	return Input{
		MouseInput: MouseInput{
			Position:         mgl64.Vec2{i.MousePosition[0], i.MousePosition[1]},
			MouseMotionEvent: i.MouseMotionEvent,
			MouseWheelDelta:  i.MouseWheelDelta,
			MouseButtonEvent: i.MouseButtonEvent,
			MouseButtonState: i.MouseButtonState,
		},
		KeyboardInput: i.KeyboardInput,
	}
}
