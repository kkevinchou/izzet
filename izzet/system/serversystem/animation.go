package serversystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/system"
	"github.com/kkevinchou/izzet/izzet/system/shared"
)

type ServerAnimationSystem struct {
	app App
}

func NewServerAnimationSystem(app App) *ServerAnimationSystem {
	return &ServerAnimationSystem{app: app}
}

func (s *ServerAnimationSystem) Name() string {
	return "ServerAnimationSystem"
}

func (s *ServerAnimationSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, e := range world.Entities() {
		if !shared.IsStateMachineAnimation(e) {
			continue
		}

		transition, transitioned := shared.UpdateStateMachineAnimation(delta, e)
		if transitioned {
			transition.GlobalCommandFrame = s.app.CommandFrame()
			e.Animation.AnimationTransitions = append(e.Animation.AnimationTransitions, transition)
		}
	}
}
