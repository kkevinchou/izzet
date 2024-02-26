package serversystems

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/systems"
)

type AISystem struct {
	app App
}

func NewAISystemSystem(app App) *AISystem {
	return &AISystem{app: app}
}

func (s *AISystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, entity := range world.Entities() {
		if entity.AIComponent == nil {
			continue
		}

		for _, camera := range world.Entities() {
			if camera.PlayerInput == nil {
				continue
			}

			if camera.CameraComponent.Target == nil {
				continue
			}

			target := world.GetEntityByID(*camera.CameraComponent.Target)

			fmt.Println("target position", target.Position())
		}

		// s.app.GetPlayers()
	}
}
