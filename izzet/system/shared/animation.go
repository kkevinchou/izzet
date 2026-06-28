package shared

import (
	"time"

	"github.com/kkevinchou/izzet/internal/utils"
	animationparser "github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/entity"
)

func IsStateMachineAnimation(e *entity.Entity) bool {
	return e.Animation != nil && e.Animation.Mode == entity.AnimationModeStateMachine
}

func UpdateStateMachineAnimation(delta time.Duration, e *entity.Entity) (entity.ServerSideAnimationTransition, bool) {
	if e.Kinematic == nil {
		return entity.ServerSideAnimationTransition{}, false
	}

	ctx := animationContext(e)
	transition, transitioned := e.Animation.AnimationStateMachine.Update(delta, e.Animation.AnimationPlayer, ctx)
	return entity.ServerSideAnimationTransition{AnimationTransition: transition}, transitioned
}

func animationContext(e *entity.Entity) animationparser.GameContext {
	var ctx animationparser.GameContext
	ctx.Grounded = e.Kinematic.Grounded
	ctx.JumpTriggered = e.Kinematic.Jump
	ctx.Moving = !utils.Vec3IsZero(e.Kinematic.MoveIntent)
	ctx.Dead = e.Deadge

	if e.AttackComponent != nil {
		ctx.Attacking = e.AttackComponent.Attacking
	}
	if e.AimDownSightsComponent != nil && e.AimDownSightsComponent.Active {
		ctx.AimDownSights = true
		ctx.AimDownSightsFire = e.AimDownSightsComponent.Fire
	}

	return ctx
}
