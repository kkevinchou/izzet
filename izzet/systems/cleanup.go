package systems

import (
	"time"
)

type CleanupSystem struct {
	app App
}

func NewCleanupSystem(app App) *CleanupSystem {
	return &CleanupSystem{app: app}
}

func (s *CleanupSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if !entity.Static && entity.Collider != nil {
			entity.Collider.Contacts = nil
		}
	}
}
