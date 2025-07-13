package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/utils"
)

// WorldTransform computes a transform matrix that represents all
// transformations applied to it based on the following factors:
// 1. transformations from model space transformations
// 2. transformations from an animated joint that the entity is parented to
// 3. transformations from the entity's parent
func WorldTransform(entity *Entity) mgl64.Mat4 {
	// TODO:
	// animations can move obects around pretty regularly, we shouldn't cache world
	// transforms for entities that have animations
	if entity.Dirty() {
		parentAndJointTransformMatrix := ComputeParentAndJointTransformMatrix(entity)
		localPosition := GetLocalPosition(entity)
		translationMatrix := mgl64.Translate3D(localPosition[0], localPosition[1], localPosition[2])
		rotationMatrix := GetLocalRotation(entity).Mat4()
		scale := entity.Scale()
		scaleMatrix := mgl64.Scale3D(scale.X(), scale.Y(), scale.Z())
		modelMatrix := translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)

		entity.cachedWorldTransform = parentAndJointTransformMatrix.Mul4(modelMatrix)
		entity.DirtyTransformFlag = false
	}

	return entity.cachedWorldTransform
}

func GetLocalPosition(entity *Entity) mgl64.Vec3 {
	return entity.LocalPosition
}

func GetLocalRotation(entity *Entity) mgl64.Quat {
	return entity.LocalRotation
}

func SetLocalPosition(entity *Entity, position mgl64.Vec3) {
	SetDirty(entity)
	entity.LocalPosition = position
}

func SetLocalRotation(entity *Entity, rotation mgl64.Quat) {
	SetDirty(entity)
	entity.LocalRotation = rotation
}

func SetScale(entity *Entity, scale mgl64.Vec3) {
	SetDirty(entity)
	entity.LocalScale = scale
}

func SetDirty(entity *Entity) {
	// note - this can potentially be optimized by not setting
	// the dirty flag on children if we were already marked dirty
	for _, child := range entity.Children {
		SetDirty(child)
	}
	entity.DirtyTransformFlag = true
}

func ComputeParentAndJointTransformMatrix(entity *Entity) mgl64.Mat4 {
	if entity.Parent != nil {
		// TODO: handle animations
		return WorldTransform(entity.Parent)
	}
	return mgl64.Ident4()
}

func (e *Entity) Rotation() mgl64.Quat {
	m := WorldTransform(e)
	_, r, _ := utils.DecomposeF64(m)
	return r
}

func (e *Entity) Position() mgl64.Vec3 {
	m := WorldTransform(e)
	return m.Mul4x1(mgl64.Vec4{0, 0, 0, 1}).Vec3()
}

func (e *Entity) Scale() mgl64.Vec3 {
	// hope i don't regret this in the future
	return e.LocalScale
}
