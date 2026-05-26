package system

import (
	"time"

	"github.com/kkevinchou/izzet/internal/animationv2"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entity"
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

		var selectAnimation bool
		if s.app.IsClient() {
			if s.app.AppMode() == types.AppModePlay && s.app.GetPlayerEntity().GetID() == e.GetID() {
				selectAnimation = true
			} else if s.app.AppMode() == types.AppModeEditor {
				selectAnimation = true
			}
		} else {
			selectAnimation = true
		}

		if selectAnimation {
			if e.CharacterControllerComponent != nil {
				if e.Kinematic != nil {
					animationPlayer := e.Animation.AnimationPlayer
					var animationName = "Sprint_Loop"
					if !e.Kinematic.GravityEnabled {
						animationName = "Jump_Loop"
					} else if !e.Kinematic.Grounded {
						animationName = "Jump_Loop"
					} else if !apputils.IsZeroVec(e.CharacterControllerComponent.ControlVector) {
						animationName = "Sprint_Loop"
					} else {
						animationName = "Idle_Loop"
					}
					animationPlayer.PlayAnimation(animationName)
				}
			} else if e.AIComponent != nil && e.Kinematic != nil {
				animationKey := entity.AnimationKeyRun
				if e.AIComponent.State == entity.AIStateAttack {
					animationKey = entity.AnimationKeyAttack
				} else if !apputils.IsZeroVec(e.TotalKinematicVelocity()) {
					animationKey = entity.AnimationKeyRun
				} else {
					animationKey = entity.AnimationKeyIdle
				}
				animationPlayer := e.Animation.AnimationPlayer
				animationPlayer.PlayAnimation(e.Animation.AnimationNames[animationKey])
			} else {
				if s.app.IsServer() {
					continue
				}
				runtimeConfig := s.app.RuntimeConfig()
				animationPlayer := e.Animation.AnimationPlayerV2

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
						e.Animation.AnimationPlayerV2.SetCurrentAnimationFrame(runtimeConfig.SelectedAnimation, runtimeConfig.SelectedKeyFrame)
					}
				}
				if runtimeConfig.LoopAnimation {
				} else if runtimeConfig.SelectedAnimation != "" {
				}
				continue
			}
		}

		e.Animation.AnimationPlayer.Update(delta)

		animationContext := &animationv2.AnimationContext{
			Player:        e.Animation.AnimationPlayerV2,
			Grounded:      e.Kinematic.Grounded,
			JumpTriggered: e.Kinematic.Jump,
			Moving:        !apputils.IsZeroVec(e.Kinematic.MoveIntent),
			Airborne:      !e.GravityEnabled() || !e.Kinematic.Grounded,
		}
		e.Animation.AnimationStateMachine.Update(delta, s.app, world, *animationContext)
	}
}
