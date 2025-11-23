package edithistory

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entity"
)

type PositionEdit struct {
	LastPosition    mgl64.Vec3
	CurrentPosition mgl64.Vec3
	Entity          *entity.Entity
}

func NewPositionEdit(lastPosition, currentPosition mgl64.Vec3, entity *entity.Entity) *PositionEdit {
	return &PositionEdit{
		LastPosition:    lastPosition,
		CurrentPosition: currentPosition,
		Entity:          entity,
	}
}

func (e *PositionEdit) Undo() {
	entity.SetLocalPosition(e.Entity, e.LastPosition)
}

func (e *PositionEdit) Redo() {
	entity.SetLocalPosition(e.Entity, e.CurrentPosition)
}

type RotationEdit struct {
	LastRotation    mgl64.Quat
	CurrentRotation mgl64.Quat
	Entity          *entity.Entity
}

func NewRotationEdit(lastRotation, currentRotation mgl64.Quat, entity *entity.Entity) *RotationEdit {
	return &RotationEdit{
		LastRotation:    lastRotation,
		CurrentRotation: currentRotation,
		Entity:          entity,
	}
}

func (e *RotationEdit) Undo() {
	e.Entity.SetLocalRotation(e.LastRotation)
}

func (e *RotationEdit) Redo() {
	e.Entity.SetLocalRotation(e.CurrentRotation)
}

type ScaleEdit struct {
	LastScale    mgl64.Vec3
	CurrentScale mgl64.Vec3
	Entity       *entity.Entity
}

func NewScaleEdit(lastScale, currentScale mgl64.Vec3, entity *entity.Entity) *ScaleEdit {
	return &ScaleEdit{
		LastScale:    lastScale,
		CurrentScale: currentScale,
		Entity:       entity,
	}
}

func (e *ScaleEdit) Undo() {
	entity.SetScale(e.Entity, e.LastScale)
}

func (e *ScaleEdit) Redo() {
	entity.SetScale(e.Entity, e.CurrentScale)
}
