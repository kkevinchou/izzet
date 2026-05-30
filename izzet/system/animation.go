package system

import (
	"time"

	animationparser "github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/types"
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

		if s.app.IsClient() && s.app.AppMode() == types.AppModeEditor {
			runtimeConfig := s.app.RuntimeConfig()
			animationPlayer := e.Animation.AnimationPlayer

			if runtimeConfig.SelectedAnimation != "" {
				if runtimeConfig.LoopAnimation {
					if animationPlayer.CurrentAnimation() != runtimeConfig.SelectedAnimation || animationPlayer.NormalizedClipProgress() >= 1 {
						animationPlayer.PlayClip(runtimeConfig.SelectedAnimation)
					}
					animationPlayer.Update(delta)
				} else {
					if animationPlayer.CurrentAnimation() != runtimeConfig.SelectedAnimation {
						animationPlayer.PlayClip(runtimeConfig.SelectedAnimation)
					}
					e.Animation.AnimationPlayer.SetCurrentAnimationFrame(runtimeConfig.SelectedAnimation, runtimeConfig.SelectedKeyFrame)
				}
			}
		} else {
			if (s.app.IsClient() && s.app.GetPlayerEntity().GetID() == e.GetID()) || s.app.IsServer() {
				animationContext := &animationparser.AnimationContext{
					Player:        e.Animation.AnimationPlayer,
					Grounded:      e.Kinematic.Grounded,
					JumpTriggered: e.Kinematic.Jump,
					Moving:        !apputils.IsZeroVec(e.Kinematic.MoveIntent),
					Airborne:      !e.GravityEnabled() || !e.Kinematic.Grounded,
				}
				e.Animation.AnimationStateMachine.Update(delta, e.Animation.AnimationPlayer, *animationContext)
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
