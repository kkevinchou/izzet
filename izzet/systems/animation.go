package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/input"
)

type AnimationSystem struct {
}

func (s *AnimationSystem) Update(delta time.Duration, world GameWorld, frameInput input.Input) {
	for _, entity := range world.Entities() {
		if entity.Animation == nil {
			continue
		}

		if panels.LoopAnimation {
			entity.Animation.AnimationPlayer.Update(delta)
		}
	}
}
