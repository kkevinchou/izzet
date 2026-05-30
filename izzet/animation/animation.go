package animation

import (
	"bytes"
	_ "embed"
	"fmt"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
)

type GameContext struct {
	Grounded      bool
	Airborne      bool
	JumpTriggered bool
	Moving        bool
}

//go:embed player_state_machine.yaml
var playerStateMachineConfig []byte

func NewPlayerAnimationStateMachine() *iztanimation.AnimationStateMachine[GameContext] {
	return iztanimation.NewAnimationStateMachine(bytes.NewReader(playerStateMachineConfig), parseCondition)
}

// parseCondition takes in a condition name and generates the condition function
// that performs that condition check based on GameContext.
//
// The engine offers baseline conditions like "clipCompleted" which are automatically
// supported and can be referenced in config - the game level parser does not need to
// handle it.
func parseCondition(name string) iztanimation.Condition[GameContext] {
	switch name {
	case "moving":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Moving
		})
	case "notMoving":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return !ctx.Moving
		})
	case "jumpTriggered":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.JumpTriggered
		})
	case "airborne":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Airborne
		})
	case "grounded":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Grounded
		})
	default:
		panic(fmt.Sprintf("unknown animation condition %q", name))
	}
}
