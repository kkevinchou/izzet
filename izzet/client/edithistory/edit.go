package edithistory

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type PositionEdit struct {
	LastPosition    mgl32.Vec3
	CurrentPosition mgl32.Vec3
	Entity          *entities.Entity
}

func NewPositionEdit(lastPosition, currentPosition mgl32.Vec3, entity *entities.Entity) *PositionEdit {
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
	LastRotation    mgl32.Quat
	CurrentRotation mgl32.Quat
	Entity          *entities.Entity
}

func NewRotationEdit(lastRotation, currentRotation mgl32.Quat, entity *entities.Entity) *RotationEdit {
	return &RotationEdit{
		LastRotation:    lastRotation,
		CurrentRotation: currentRotation,
		Entity:          entity,
	}
}

func (e *RotationEdit) Undo() {
	entities.SetLocalRotation(e.Entity, e.LastRotation)
}

func (e *RotationEdit) Redo() {
	entities.SetLocalRotation(e.Entity, e.CurrentRotation)
}

type ScaleEdit struct {
	LastScale    mgl32.Vec3
	CurrentScale mgl32.Vec3
	Entity       *entities.Entity
}

func NewScaleEdit(lastScale, currentScale mgl32.Vec3, entity *entities.Entity) *ScaleEdit {
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
