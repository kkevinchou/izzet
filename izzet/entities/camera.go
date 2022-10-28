package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/types"
)

const (
	maxFollowDistance     float64 = 300
	defaultFollowDistance float64 = 60
	defaultFollowY        float64 = 35
)

func NewThirdPersonCamera(positionOffset mgl64.Vec3, view mgl64.Vec2, playerID int, followTargetEntityID int) *EntityImpl {
	cameraComponent := &components.CameraComponent{
		FollowTargetEntityID: followTargetEntityID,
		FollowDistance:       defaultFollowDistance,
		MaxFollowDistance:    maxFollowDistance,
		YOffset:              defaultFollowY,
	}

	transformComponent := &components.TransformComponent{
		Orientation: mgl64.QuatIdent(),
		Position:    mgl64.Vec3{0, cameraComponent.YOffset, cameraComponent.FollowDistance},
	}

	entity := NewEntity(
		"camera",
		types.EntityTypeCamera,
		components.NewComponentContainer(
			transformComponent,
			cameraComponent,
			&components.ControlComponent{PlayerID: playerID},
		),
	)

	return entity
}
