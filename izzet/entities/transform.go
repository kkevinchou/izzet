package entities

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/utils"
)

// WorldTransform computes a transform matrix that represents all
// transformations applied to it based on the following factors:
// 1. transformations from model space transformations
// 2. transformations from an animated joint that the entity is parented to
// 3. transformations from the entity's parent
func WorldTransform(entity *Entity) mgl32.Mat4 {
	// TODO:
	// animations can move obects around pretty regularly, we shouldn't cache world
	// transforms for entities that have animations
	if entity.Dirty() {
		parentAndJointTransformMatrix := ComputeParentAndJointTransformMatrix(entity)

		localPosition := GetLocalPosition(entity)
		translationMatrix := mgl32.Translate3D(localPosition[0], localPosition[1], localPosition[2])
		rotationMatrix := GetLocalRotation(entity).Mat4()
		scale := GetLocalScale(entity)
		scaleMatrix := mgl32.Scale3D(scale.X(), scale.Y(), scale.Z())
		modelMatrix := translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)

		entity.cachedWorldTransform = parentAndJointTransformMatrix.Mul4(modelMatrix)
		entity.DirtyTransformFlag = false
	}

	return entity.cachedWorldTransform
}

func GetLocalPosition(entity *Entity) mgl32.Vec3 {
	return entity.LocalPosition
}

func GetLocalRotation(entity *Entity) mgl32.Quat {
	return entity.LocalRotation
}

func GetLocalScale(entity *Entity) mgl32.Vec3 {
	return entity.LocalScale
}

func SetLocalPosition(entity *Entity, position mgl32.Vec3) {
	SetDirty(entity)
	entity.LocalPosition = position
}

func SetLocalRotation(entity *Entity, rotation mgl32.Quat) {
	SetDirty(entity)
	entity.LocalRotation = rotation
}

func SetScale(entity *Entity, scale mgl32.Vec3) {
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

func ComputeParentAndJointTransformMatrix(entity *Entity) mgl32.Mat4 {
	parentModelMatrix := mgl32.Ident4()
	animModelMatrix := mgl32.Ident4()
	if entity.Parent != nil {
		parentModelMatrix = WorldTransform(entity.Parent)
	}

	return parentModelMatrix.Mul4(animModelMatrix)
}

func (e *Entity) WorldRotation() mgl32.Quat {
	m := WorldTransform(e)
	_, r, _ := utils.DecomposeF64(m)
	return r
}

func (e *Entity) Position() mgl32.Vec3 {
	m := WorldTransform(e)
	return m.Mul4x1(mgl32.Vec4{0, 0, 0, 1}).Vec3()
}
