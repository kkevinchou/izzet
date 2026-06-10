package system

import (
	"time"

	"github.com/kkevinchou/izzet/internal/utils"
	animationparser "github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/entity"
)

type AnimationSystem struct {
	app App
}

func NewAnimationSystem(app App) *AnimationSystem {
	return &AnimationSystem{app: app}
}

func (s *AnimationSystem) Name() string {
	return "AnimationSystem"
}

func (s *AnimationSystem) Update(delta time.Duration, world GameWorld) {
	for _, e := range world.Entities() {
		if e.Animation == nil {
			continue
		}

		if e.Animation.SelectedAnimation != "" {
			animationComponent := e.Animation
			animationPlayer := e.Animation.AnimationPlayer

			if animationComponent.LoopAnimation {
				if animationPlayer.CurrentAnimation() != animationComponent.SelectedAnimation || animationPlayer.NormalizedClipProgress() >= 1 {
					animationPlayer.PlayClip(animationComponent.SelectedAnimation)
				}
				animationPlayer.Update(delta)
			} else {
				if animationPlayer.CurrentAnimation() != animationComponent.SelectedAnimation {
					animationPlayer.PlayClip(animationComponent.SelectedAnimation)
				}
				animationPlayer.SetCurrentAnimationFrame(animationComponent.SelectedAnimation, animationComponent.SelectedKeyFrame)
			}
		} else {
			if e.Kinematic != nil && ((s.app.IsClient() && s.app.GetPlayerEntity().GetID() == e.GetID()) || s.app.IsServer()) {
				var ctx animationparser.GameContext
				ctx.Grounded = e.Kinematic.Grounded
				ctx.JumpTriggered = e.Kinematic.Jump
				ctx.Moving = !utils.Vec3IsZero(e.Kinematic.MoveIntent)
				if e.AttackComponent != nil {
					ctx.Attacking = e.AttackComponent.Attacking
				}
				if e.AimDownSightsComponent != nil && e.AimDownSightsComponent.Active {
					ctx.AimDownSights = true
					ctx.AimDownSightsFire = e.AimDownSightsComponent.Fire
				}

				ctx.Dead = e.Deadge
				prevState := e.Animation.AnimationStateMachine.CurrentAnimationState()
				e.Animation.AnimationStateMachine.Update(delta, e.Animation.AnimationPlayer, ctx)
				currentState := e.Animation.AnimationStateMachine.CurrentAnimationState()

				// there's probably a better way to do this like with a callback within the animation state machine.
				// we could record more internal details.
				//
				// this doesn't handle looping
				if prevState != currentState {
					e.Animation.AnimationTransitions = append(
						e.Animation.AnimationTransitions,
						entity.AnimationTransition{
							SourceState:      prevState,
							DestinationState: currentState,
							CommandFrame:     s.app.CommandFrame(),
						},
					)
				}
			} else {
				// entities replicated to the client just need their animation player updated.
				// we rely on the game state update message to set the animation clip
				if e.Animation.AnimationPlayer.CurrentAnimation() != "" {
					e.Animation.AnimationPlayer.Update(delta)
					if e.Animation.AnimationPlayer.NormalizedClipProgress() >= 1 {
						e.Animation.AnimationPlayer.PlayClip(e.Animation.AnimationPlayer.CurrentAnimation())
					}
				}
			}
		}
	}
}
