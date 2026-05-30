package animation

import (
	"fmt"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
)

func parseCondition(name string) iztanimation.Condition[AnimationContext] {
	switch name {
	case "moving":
		return &MovingCondition{}
	case "notMoving":
		return &NotMovingCondition{}
	case "jumpTriggered":
		return &JumpTriggeredCondition{}
	case "clipCompleted":
		return &ClipCompletedCondition{}
	case "airborne":
		return &AirborneCondition{}
	case "grounded":
		return &GroundedCondition{}
	default:
		panic(fmt.Sprintf("unknown animation condition %q", name))
	}
}
