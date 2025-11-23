package system

import (
	"time"

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
					var animationName = "Walk"
					if !e.Kinematic.GravityEnabled {
						animationName = "Floating"
					} else if !e.Kinematic.Grounded {
						animationName = "Falling"
					} else if !apputils.IsZeroVec(e.CharacterControllerComponent.ControlVector) {
						animationName = "Running"
					} else {
						animationName = "Idle"
					}
					animationPlayer.PlayAnimation(animationName)
				}
			} else if e.AIComponent != nil && e.Kinematic != nil {
				animationKey := entity.AnimationKeyRun
				if e.AIComponent.State == entity.AIStateAttack {
					animationKey = entity.AnimationKeyAttack
				} else if !apputils.IsZeroVec(e.Kinematic.Velocity) {
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
				if runtimeConfig.LoopAnimation {
					animationPlayer := e.Animation.AnimationPlayer
					animationPlayer.PlayAnimation(runtimeConfig.SelectedAnimation)
					e.Animation.AnimationPlayer.Update(delta)
				} else {
					e.Animation.AnimationPlayer.SetCurrentAnimationFrame(runtimeConfig.SelectedAnimation, runtimeConfig.SelectedKeyFrame)
				}
				continue
			}
		}

		e.Animation.AnimationPlayer.Update(delta)
	}
}
