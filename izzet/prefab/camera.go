package prefab

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func CreateCamera(playerID int, targetEntityID int) *entity.Entity {
	e := entity.CreateEmptyEntity("camera")
	e.CameraComponent = &entity.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, settings.CameraEntityFollowVerticalOffset, 0}, Target: &targetEntityID}
	e.ImageComponent = entity.NewImageComponent("camera.png", 1, true)
	e.PlayerInput = &entity.PlayerInputComponent{PlayerID: playerID}
	return e
}
