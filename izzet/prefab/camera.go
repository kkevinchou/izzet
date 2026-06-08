package prefab

import (
	"github.com/kkevinchou/izzet/izzet/entity"
)

func CreateCamera(playerID int) *entity.Entity {
	e := entity.CreateEmptyEntity("camera")
	e.CameraComponent = &entity.CameraComponent{CameraMode: entity.CameraModeOverShoulder}
	e.ImageComponent = entity.NewImageComponent("camera.png", 1, true)
	e.PlayerInput = &entity.PlayerInputComponent{PlayerID: playerID}
	return e
}
