package systems

import (
	"strings"
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

		// animations for remote entities is synchronized from the server
		if s.app.IsClient() {
			if s.app.AppMode() == mode.AppModePlay && s.app.GetPlayerEntity().GetID() != entity.GetID() {
				continue
			}
		}

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
			} else if entity.AIComponent.PathfindConfig.State == entities.PathfindingStateNoGoal {
				animationKey = entities.AnimationKeyIdle
			}
			animationPlayer := entity.Animation.AnimationPlayer
			animationPlayer.PlayAnimation(entity.Animation.AnimationNames[animationKey])
		} else if strings.Contains(entity.Name, "gun") {
			animationPlayer := entity.Animation.AnimationPlayer
			animationPlayer.PlayAnimation("Test")
		}

		entity.Animation.AnimationPlayer.Update(delta)
	}
}
