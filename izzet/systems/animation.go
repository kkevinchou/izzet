package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/mode"
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
	for _, entity := range world.Entities() {
		if entity.Animation == nil {
			continue
		}

		var selectAnimation bool
		if s.app.IsClient() {
			if s.app.AppMode() == mode.AppModePlay && s.app.GetPlayerEntity().GetID() == entity.GetID() {
				selectAnimation = true
			} else if s.app.AppMode() == mode.AppModeEditor {
				selectAnimation = true
			}
		} else {
			selectAnimation = true
		}

		if selectAnimation {
			if entity.CharacterControllerComponent != nil {
				animationPlayer := entity.Animation.AnimationPlayer
				var animationName = "Walk"
				if !entity.Physics.GravityEnabled {
					animationName = "Floating"
				} else if !entity.Physics.Grounded {
					animationName = "Falling"
				} else if !apputils.IsZeroVec(entity.CharacterControllerComponent.ControlVector) {
					animationName = "Running"
				} else {
					animationName = "Idle"
				}
				animationPlayer.PlayAnimation(animationName)
			} else if entity.AIComponent != nil {
				animationKey := entities.AnimationKeyRun
				if entity.AIComponent.State == entities.AIStateAttack {
					animationKey = entities.AnimationKeyAttack
				} else if entity.AIComponent.PathfindConfig != nil && entity.AIComponent.PathfindConfig.State == entities.PathfindingStateNoGoal {
					animationKey = entities.AnimationKeyIdle
				}
				animationPlayer := entity.Animation.AnimationPlayer
				animationPlayer.PlayAnimation(entity.Animation.AnimationNames[animationKey])
			} else {
				if s.app.IsServer() {
					continue
				}
				runtimeConfig := s.app.RuntimeConfig()
				if runtimeConfig.LoopAnimation {
					animationPlayer := entity.Animation.AnimationPlayer
					animationPlayer.PlayAnimation(runtimeConfig.SelectedAnimation)
					entity.Animation.AnimationPlayer.Update(delta)
				} else {
					entity.Animation.AnimationPlayer.SetCurrentAnimationFrame(runtimeConfig.SelectedAnimation, runtimeConfig.SelectedKeyFrame)
				}
				continue
			}
		}

		entity.Animation.AnimationPlayer.Update(delta)
	}
}
