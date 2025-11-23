package system

import (
	"time"
)

type CleanupSystem struct {
	app App
}

func NewCleanupSystem(app App) *CleanupSystem {
	return &CleanupSystem{app: app}
}

func (s *CleanupSystem) Name() string {
	return "CleanupSystem"
}

func (s *CleanupSystem) Update(delta time.Duration, world GameWorld) {
	// if s.app.PredictionDebugLogging() {
	// 	e := s.app.World().GetEntityByID(5144)
	// 	if e != nil {
	// 		fmt.Printf("\t - End Frame First Entity Position [Position: %s]\n", apputils.FormatVec(s.app.World().GetEntityByID(5142).Position()))
	// 		fmt.Printf("\t - End Frame Position [Position: %s]\n", apputils.FormatVec(e.Position()))
	// 	}
	// }
	for _, entity := range world.Entities() {
		if !entity.Static && entity.Collider != nil {
			entity.Collider.Contacts = nil
		}
	}
}
