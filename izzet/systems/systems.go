package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/input"
)

type App interface {
	Entities() []*entities.Entity
}

type CharacterControllerSystem struct {
}

func (s *CharacterControllerSystem) Update(delta time.Duration, app App, frameInput input.Input) {
	for _, entity := range app.Entities() {
		if entity.CharacterControllerComponent == nil {
			continue
		}

		c := entity.CharacterControllerComponent

		keyboardInput := frameInput.KeyboardInput
		if key, ok := keyboardInput[input.KeyboardKeyI]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{0, 0, -c.Speed * float64(delta.Milliseconds()) / 1000}))
		}
		if key, ok := keyboardInput[input.KeyboardKeyK]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{0, 0, c.Speed * float64(delta.Milliseconds()) / 1000}))
		}

		if key, ok := keyboardInput[input.KeyboardKeyJ]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{-c.Speed * float64(delta.Milliseconds()) / 1000, 0, 0}))
		}
		if key, ok := keyboardInput[input.KeyboardKeyL]; ok && key.Event == input.KeyboardEventDown {
			entities.SetLocalPosition(entity, entity.LocalPosition.Add(mgl64.Vec3{c.Speed * float64(delta.Milliseconds()) / 1000, 0, 0}))
		}
	}
}
