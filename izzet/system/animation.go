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
			if e.Kinematic == nil {
				continue
			}

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
				src, dst, transitioned := e.Animation.AnimationStateMachine.Update(delta, e.Animation.AnimationPlayer, ctx)

				if transitioned {
					e.Animation.AnimationTransitions = append(
						e.Animation.AnimationTransitions,
						entity.AnimationTransition{
							SourceState:      src,
							DestinationState: dst,
							CommandFrame:     s.app.CommandFrame(),
						},
					)
				}
			} else {
				// trigger transitions sent over from the server
				if e.Animation.ReplicationSource != "" {
					e.Animation.AnimationStateMachine.TriggerTransition(e.Animation.AnimationPlayer, e.Animation.ReplicationSource, e.Animation.ReplicationDestination)
				}
				e.Animation.AnimationPlayer.Update(delta)
			}
		}
	}
}
