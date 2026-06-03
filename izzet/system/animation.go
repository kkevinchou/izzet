package system

import (
	"time"

	animationparser "github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/apputils"
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
				ctx.Moving = !apputils.IsZeroVec(e.Kinematic.MoveIntent)

				ctx.Dead = e.Deadge
				e.Animation.AnimationStateMachine.Update(delta, e.Animation.AnimationPlayer, ctx)
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
