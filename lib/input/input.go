package input

import "github.com/go-gl/mathgl/mgl64"

type InputPoller func() Input

type MouseWheelDirection int

type MouseMotionEvent struct {
	XRel float64
	YRel float64
}

func (m MouseMotionEvent) IsZero() bool {
	return m.XRel == 0 && m.YRel == 0
}

type MouseInput struct {
	MouseWheelDelta  int
	MouseMotionEvent MouseMotionEvent
	Buttons          [3]bool // left, right, middle
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
	KeyboardKeySpace  KeyboardKey = "Space"
	KeyboardKeyEscape KeyboardKey = "Escape"

	KeyboardKeyTick KeyboardKey = "`"
	KeyboardKeyF1   KeyboardKey = "F1"
	KeyboardKeyF2   KeyboardKey = "F2"

	KeyboardEventUp = iota
	KeyboardEventDown
)

type KeyState struct {
	Key   KeyboardKey
	Event KeyboardEvent
}

type KeyboardInput map[KeyboardKey]KeyState

type QuitCommand struct {
}

// Input represents the input provided by a user during a command frame
// Input should be only constructed by the input poller and should not be
// written to by any systems, only read. Input is stored in a client side
// command frame history which will copy the KeyboardInput by reference
type Input struct {
	KeyboardInput     KeyboardInput
	MouseInput        MouseInput
	CameraOrientation mgl64.Quat
	Commands          []any
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
