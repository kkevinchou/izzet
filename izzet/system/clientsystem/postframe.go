package clientsystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/system"
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

func (s *PostFrameSystem) Update(delta time.Duration, world system.GameWorld) {
	sb := s.app.StateBuffer()
	playerEntity := s.app.GetPlayerEntity()
	if bi, ok := sb.Pull(s.app.CommandFrame()); ok {
		for _, bs := range bi.EntityStates {
			e := world.GetEntityByID(bs.EntityID)
			if e == nil {
				continue
			}

			if bs.EntityID == playerEntity.ID || bs.EntityID == playerEntity.CharacterControllerComponent.CameraEntityID {
				continue
			}

			entity.SetLocalPosition(e, bs.Position)
			e.SetLocalRotation(bs.Rotation)
			if e.Animation != nil {
				e.Animation.ReplicatedAnimationTransition = bs.AnimationTransition
			}
		}
	}

	history := s.app.GetCommandFrameHistory()
	// fmt.Printf("CLIENT ACTUAL - [%d] - %v\n", s.app.CommandFrame(), entity.GetLocalPosition(s.app.GetPlayerEntity()))
	history.AddCommandFrame(s.app.CommandFrame(), s.app.GetFrameInput(), s.app.GetPlayerEntity())
}
