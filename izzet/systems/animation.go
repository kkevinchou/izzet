package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/app"
)

type AnimationSystem struct {
	app App
}

func NewAnimationSystem(app App) *AnimationSystem {
	return &AnimationSystem{app: app}
}

func (s *AnimationSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if entity.Animation == nil {
			continue
		}

		// for entities that dont belong to the current player animation selection
		// is synchronized from the gamestate update message
		if (s.app.IsClient() && s.app.GetPlayerEntity().GetID() == entity.GetID()) || s.app.IsServer() {
			if entity.CharacterControllerComponent != nil {
				animationPlayer := entity.Animation.AnimationPlayer
				currentAnimation := animationPlayer.CurrentAnimation()
				var animationName = "Walk"
				if !entity.Physics.Grounded {
					animationName = "Falling"
				} else if !app.IsZeroVec(entity.CharacterControllerComponent.ControlVector) {
					animationName = "Walk"
				} else {
					animationName = "Idle"
				}
				if currentAnimation != animationName {
					if currentAnimation == "" {
						animationPlayer.PlayAnimation(animationName)
					} else {
						animationPlayer.PlayAndBlendAnimation(animationName, 250*time.Millisecond)
					}
				}
			}
		}

		entity.Animation.AnimationPlayer.Update(delta)
	}
}
