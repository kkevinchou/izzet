package clientsystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type PostFrameSystem struct {
	app App
}

func NewPostFrameSystem(app App) *PostFrameSystem {
	return &PostFrameSystem{app: app}
}

func (s *PostFrameSystem) Name() string {
	return "PostFrameSystem"
}

func (s *PostFrameSystem) Update(delta time.Duration, world systems.GameWorld) {
	sb := s.app.StateBuffer()
	if bi, ok := sb.Pull(s.app.CommandFrame()); ok {
		for _, bs := range bi.EntityStates {
			if bs.EntityID == s.app.GetPlayerEntity().ID {
				continue
			}

			entity := world.GetEntityByID(bs.EntityID)
			if entity == nil {
				continue
			}

			if entity.ID == s.app.GetPlayerCamera().ID {
				continue

			}

			if bs.Deadge {
				world.DeleteEntity(bs.EntityID)
			} else {
				entities.SetLocalPosition(entity, bs.Position)
				entity.SetLocalRotation(bs.Rotation)
			}
		}
	}

	history := s.app.GetCommandFrameHistory()
	// fmt.Printf("CLIENT ACTUAL - [%d] - %v\n", s.app.CommandFrame(), entities.GetLocalPosition(s.app.GetPlayerEntity()))
	history.AddCommandFrame(s.app.CommandFrame(), s.app.GetFrameInput(), s.app.GetPlayerEntity())
}
