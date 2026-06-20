package animation

import (
	"bytes"
	_ "embed"
	"fmt"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
)

type StateMachineID string

const (
	StateMachineIDUnassigned   StateMachineID = "UNASSIGNED"
	StateMachineIDPlayer       StateMachineID = "PLAYER"
	StateMachineIDVelociraptor StateMachineID = "VELOCIRAPTOR"
)

var typeToConfig map[StateMachineID][]byte = map[StateMachineID][]byte{
	StateMachineIDPlayer:       playerStateMachineConfig,
	StateMachineIDVelociraptor: raptorStateMachineConfig,
}

type GameContext struct {
	Grounded          bool
	JumpTriggered     bool
	Moving            bool
	Attacking         bool
	Dead              bool
	AimDownSights     bool
	AimDownSightsFire bool
}

//go:embed player_state_machine.yaml
var playerStateMachineConfig []byte

//go:embed raptor_state_machine.yaml
var raptorStateMachineConfig []byte

func NewStateMachine(id StateMachineID) *iztanimation.AnimationStateMachine[GameContext] {
	if cfg, ok := typeToConfig[id]; ok {
		return iztanimation.NewAnimationStateMachine(bytes.NewReader(cfg), parseCondition)
	}
	panic(fmt.Sprintf("unexpected state machine id: %s", id))
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
	case "jumpTriggered":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.JumpTriggered
		})
	case "attacking":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Attacking
		})
	case "aimDownSights":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.AimDownSights
		})
	case "notAimDownSights":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return !ctx.AimDownSights
		})
	case "aimDownSightsFire":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.AimDownSightsFire
		})
	case "dead":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return ctx.Dead
		})
	case "alive":
		return iztanimation.NewGameCondition(name, func(ctx GameContext) bool {
			return !ctx.Dead
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
