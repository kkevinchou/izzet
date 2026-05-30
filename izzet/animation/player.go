package animation

import (
	"bytes"
	_ "embed"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
)

//go:embed player_state_machine.yaml
var playerStateMachineConfig []byte

type AnimationContext struct {
	Grounded      bool
	Airborne      bool
	JumpTriggered bool
	Moving        bool
	Player        *iztanimation.AnimationPlayer
}

func NewAnimationStateMachine() iztanimation.AnimationStateMachine[AnimationContext] {
	return iztanimation.NewAnimationStateMachine(bytes.NewReader(playerStateMachineConfig), parseCondition)
}
