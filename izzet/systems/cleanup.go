package systems

import (
	"time"
)

type CleanupSystem struct {
}

func (s *CleanupSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if !entity.Static && entity.Collider != nil {
			entity.Collider.Contacts = nil
		}

		if entity.Deadge {
			world.DeleteEntity(entity.ID)
		}
	}
}
