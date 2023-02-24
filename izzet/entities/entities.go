package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/model"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var id int

type Entity struct {
	ID   int
	Name string

	// each Entity has their own transforms and animation player
	LocalPosition mgl64.Vec3
	LocalRotation mgl64.Quat
	Scale         mgl64.Vec3

	// shape component
	ShapeData *ShapeData

	// prefabs
	Prefab *prefabs.Prefab

	// model
	Model *model.Model

	// socket
	IsSocket bool

	// light
	LightInfo *LightInfo

	// relationships
	Parent      *Entity
	Children    map[int]*Entity
	ParentJoint *modelspec.JointSpec

	// animation
	Animations      map[string]*modelspec.AnimationSpec
	AnimationPlayer *animation.AnimationPlayer
}

func (e *Entity) WorldPosition() mgl64.Vec3 {
	m := ComputeTransformMatrix(e)
	return m.Mul4x1(mgl64.Vec4{0, 0, 0, 1}).Vec3()
}

func (e *Entity) WorldRotation() mgl64.Quat {
	m := ComputeTransformMatrix(e)
	_, r, _ := utils.DecomposeF64(m)
	return r
}

func SetNextID(nextID int) {
	id = nextID
}

func InstantiateFromPrefab(prefab *prefabs.Prefab) *Entity {
	e := InstantiateFromPrefabStaticID(id, prefab)
	id += 1
	return e
}

func InstantiateBaseEntity(name string, id int) *Entity {
	return &Entity{
		ID:   id,
		Name: fmt.Sprintf("%s-%d", name, id),

		Children: map[int]*Entity{},

		LocalPosition: mgl64.Vec3{0, 0, 0},
		LocalRotation: mgl64.QuatIdent(),
		Scale:         mgl64.Vec3{1, 1, 1},
	}
}

func InstantiateFromPrefabStaticID(id int, prefab *prefabs.Prefab) *Entity {
	e := InstantiateBaseEntity(prefab.Name, id)
	e.Prefab = prefab
	// TODO: this will break when we have prefabs supporting multiple models
	e.Model = prefab.ModelRefs[0].Model

	// animation setup
	e.Animations = prefab.ModelRefs[0].Model.Animations()
	if len(e.Animations) > 0 {
		e.AnimationPlayer = animation.NewAnimationPlayer(prefab.ModelRefs[0].Model)
	}

	return e
}

func ComputeParentAndJointTransformMatrix(entity *Entity) mgl64.Mat4 {
	parentModelMatrix := mgl64.Ident4()
	animModelMatrix := mgl64.Ident4()
	if entity.Parent != nil {
		parentModelMatrix = ComputeTransformMatrix(entity.Parent)

		parent := entity.Parent
		parentJoint := entity.ParentJoint
		if parentJoint != nil && parent != nil && parent.AnimationPlayer != nil && parent.AnimationPlayer.CurrentAnimation() != "" {
			modelSpec := parent.Model.ModelSpecification()
			animationTransforms := parent.AnimationPlayer.AnimationTransforms()
			jointTransform := animationTransforms[parentJoint.ID]
			bindTransform := modelSpec.JointMap[parentJoint.ID].FullBindTransform
			animModelMatrix = utils.Mat4F32ToF64(jointTransform).Mul4(utils.Mat4F32ToF64(bindTransform))
		}
	}

	return parentModelMatrix.Mul4(animModelMatrix)

}

// ComputeTransformMatrix computes a transform matrix that represents all
// transformations applied to it based on the following factors:
// 1. transformations from model space transformations
// 2. transformations from an animated joint that the entity is parented to
// 3. transformations from the entity's parent
func ComputeTransformMatrix(entity *Entity) mgl64.Mat4 {
	parentAndJointTransformMatrix := ComputeParentAndJointTransformMatrix(entity)

	translationMatrix := mgl64.Translate3D(entity.LocalPosition[0], entity.LocalPosition[1], entity.LocalPosition[2])
	rotationMatrix := entity.LocalRotation.Mat4()
	scaleMatrix := mgl64.Scale3D(entity.Scale.X(), entity.Scale.Y(), entity.Scale.Z())
	modelMatrix := translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)

	return parentAndJointTransformMatrix.Mul4(modelMatrix)
}
