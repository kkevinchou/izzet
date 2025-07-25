package edithistory

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type PositionEdit struct {
	LastPosition    mgl64.Vec3
	CurrentPosition mgl64.Vec3
	Entity          *entities.Entity
}

func NewPositionEdit(lastPosition, currentPosition mgl64.Vec3, entity *entities.Entity) *PositionEdit {
	return &PositionEdit{
		LastPosition:    lastPosition,
		CurrentPosition: currentPosition,
		Entity:          entity,
	}
}

func (e *PositionEdit) Undo() {
	entities.SetLocalPosition(e.Entity, e.LastPosition)
}

func (e *PositionEdit) Redo() {
	entities.SetLocalPosition(e.Entity, e.CurrentPosition)
}

type RotationEdit struct {
	LastRotation    mgl64.Quat
	CurrentRotation mgl64.Quat
	Entity          *entities.Entity
}

func NewRotationEdit(lastRotation, currentRotation mgl64.Quat, entity *entities.Entity) *RotationEdit {
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
	Entity       *entities.Entity
}

func NewScaleEdit(lastScale, currentScale mgl64.Vec3, entity *entities.Entity) *ScaleEdit {
	return &ScaleEdit{
		LastScale:    lastScale,
		CurrentScale: currentScale,
		Entity:       entity,
	}
}

func (e *ScaleEdit) Undo() {
	entities.SetScale(e.Entity, e.LastScale)
}

func (e *ScaleEdit) Redo() {
	entities.SetScale(e.Entity, e.CurrentScale)
}
