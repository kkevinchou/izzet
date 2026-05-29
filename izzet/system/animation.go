package system

import (
	"time"

	"github.com/kkevinchou/izzet/internal/animation"
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
			animationContext := &animation.AnimationContext{
				Player:        e.Animation.AnimationPlayer,
				Grounded:      e.Kinematic.Grounded,
				JumpTriggered: e.Kinematic.Jump,
				Moving:        !apputils.IsZeroVec(e.Kinematic.MoveIntent),
				Airborne:      !e.GravityEnabled() || !e.Kinematic.Grounded,
			}
			e.Animation.AnimationStateMachine.Update(delta, s.app, world, *animationContext)
		}
	}
}
