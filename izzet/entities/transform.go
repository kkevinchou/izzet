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
	if entity.dirtyTransformFlag {
		parentAndJointTransformMatrix := ComputeParentAndJointTransformMatrix(entity)

		localPosition := LocalPosition(entity)
		translationMatrix := mgl64.Translate3D(localPosition[0], localPosition[1], localPosition[2])
		rotationMatrix := LocalRotation(entity).Mat4()
		scale := Scale(entity)
		scaleMatrix := mgl64.Scale3D(scale.X(), scale.Y(), scale.Z())
		modelMatrix := translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)

		entity.cachedWorldTransform = parentAndJointTransformMatrix.Mul4(modelMatrix)
		entity.dirtyTransformFlag = false
	}

	return entity.cachedWorldTransform
}

func LocalPosition(entity *Entity) mgl64.Vec3 {
	return entity.localPosition
}

func LocalRotation(entity *Entity) mgl64.Quat {
	return entity.localRotation
}

func Scale(entity *Entity) mgl64.Vec3 {
	return entity.scale
}

func SetLocalPosition(entity *Entity, position mgl64.Vec3) {
	SetDirty(entity)
	entity.localPosition = position
}

func SetLocalRotation(entity *Entity, rotation mgl64.Quat) {
	SetDirty(entity)
	entity.localRotation = rotation
}

func SetScale(entity *Entity, scale mgl64.Vec3) {
	SetDirty(entity)
	entity.scale = scale
}

func SetDirty(entity *Entity) {
	// note - this can potentially be optimized by not setting
	// the dirty flag on children if we were already marked dirty
	for _, child := range entity.Children {
		SetDirty(child)
	}
	entity.dirtyTransformFlag = true
}

func ComputeParentAndJointTransformMatrix(entity *Entity) mgl64.Mat4 {
	parentModelMatrix := mgl64.Ident4()
	animModelMatrix := mgl64.Ident4()
	if entity.Parent != nil {
		parentModelMatrix = WorldTransform(entity.Parent)

		parent := entity.Parent
		parentJoint := entity.ParentJoint
		if parentJoint != nil && parent != nil && parent.AnimationPlayer != nil && parent.AnimationPlayer.CurrentAnimation() != "" {
			animationTransforms := parent.AnimationPlayer.AnimationTransforms()
			jointTransform := animationTransforms[parentJoint.ID]
			jointMap := parent.Model.JointMap()
			bindTransform := jointMap[parentJoint.ID].FullBindTransform
			animModelMatrix = utils.Mat4F32ToF64(jointTransform).Mul4(utils.Mat4F32ToF64(bindTransform))
		}
	}

	return parentModelMatrix.Mul4(animModelMatrix)
}

func (e *Entity) WorldRotation() mgl64.Quat {
	m := WorldTransform(e)
	_, r, _ := utils.DecomposeF64(m)
	return r
}

func (e *Entity) WorldPosition() mgl64.Vec3 {
	m := WorldTransform(e)
	return m.Mul4x1(mgl64.Vec4{0, 0, 0, 1}).Vec3()
}

func (e *Entity) Position() mgl64.Vec3 {
	return e.WorldPosition()
}