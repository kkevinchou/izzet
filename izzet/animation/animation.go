package animation

import (
	"bytes"
	_ "embed"
	"fmt"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
)

type GameContext struct {
	Grounded      bool
	JumpTriggered bool
	Moving        bool
	Walking       bool
	Attacking     bool
	Dead          bool
}

//go:embed player_state_machine.yaml
var playerStateMachineConfig []byte

//go:embed raptor_state_machine.yaml
var raptorStateMachineConfig []byte

func NewPlayerAnimationStateMachine() *iztanimation.AnimationStateMachine[GameContext] {
	return iztanimation.NewAnimationStateMachine(bytes.NewReader(playerStateMachineConfig), parseCondition)
}

func NewRaptorAnimationStateMachine() *iztanimation.AnimationStateMachine[GameContext] {
	return iztanimation.NewAnimationStateMachine(bytes.NewReader(raptorStateMachineConfig), parseCondition)
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
	case "stationary":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return !ctx.Moving
		})
	case "walking":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Walking
		})
	case "running":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Moving && !ctx.Walking
		})
	case "jumpTriggered":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.JumpTriggered
		})
	case "attacking":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Attacking
		})
	case "dead":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Dead
		})
	case "airborne":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return !ctx.Grounded
		})
	case "grounded":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Grounded
		})
	default:
		panic(fmt.Sprintf("unknown animation condition %q", name))
	}
}
