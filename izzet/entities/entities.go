package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/collision/collider"
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
	ShapeData []*ShapeData

	// prefabs
	Prefab *prefabs.Prefab

	// model
	Model       *model.Model
	boundingBox *collider.BoundingBox

	// socket
	IsSocket bool

	// light
	LightInfo *LightInfo

	// image
	ImageInfo *ImageInfo

	// misc
	Billboard *BillboardInfo

	// relationships
	Parent      *Entity
	Children    map[int]*Entity
	ParentJoint *modelspec.JointSpec

	// particles
	Particles *ParticleGenerator

	// animation
	Animations      map[string]*modelspec.AnimationSpec
	AnimationPlayer *animation.AnimationPlayer

	// physics
	Physics *PhysicsComponent

	// collision
	Collider *ColliderComponent
}

func (e *Entity) GetID() int {
	return e.ID
}

func (e *Entity) NameID() string {
	return fmt.Sprintf("%s-%d", e.Name, e.ID)
}

func (e *Entity) WorldPosition() mgl64.Vec3 {
	m := ComputeTransformMatrix(e)
	return m.Mul4x1(mgl64.Vec4{0, 0, 0, 1}).Vec3()
}

func (e *Entity) Position() mgl64.Vec3 {
	return e.WorldPosition()
}

func (e *Entity) BoundingBox() *collider.BoundingBox {
	if e.boundingBox == nil {
		return nil
	}
	modelMatrix := ComputeTransformMatrix(e)
	// t, r, s := utils.DecomposeF64(modelMatrix)
	// translation := mgl64.Translate3D(t.X(), t.Y(), t.Z())
	// scale := mgl64.Scale3D(s.X(), s.Y(), s.Z())

	// return e.boundingBox.Transform(translation.Mul4(r.Mat4()).Mul4(scale))
	return e.boundingBox.Transform(modelMatrix)
}

func (e *Entity) WorldRotation() mgl64.Quat {
	m := ComputeTransformMatrix(e)
	_, r, _ := utils.DecomposeF64(m)
	return r
}

func InstantiateFromPrefab(prefab *prefabs.Prefab) []*Entity {
	var es []*Entity
	count := 0
	for _, modelRef := range prefab.ModelRefs {
		model := modelRef.Model
		e := InstantiateFromPrefabStaticID(id, model, prefab)
		es = append(es, e)
		id += 1
		count++
	}
	return es
}

func InstantiateFromPrefabStaticID(id int, model *model.Model, prefab *prefabs.Prefab) *Entity {
	e := InstantiateBaseEntity(model.Name(), id)
	e.Prefab = prefab
	e.Model = model
	e.boundingBox = collider.BoundingBoxFromModel(e.Model)

	e.LocalPosition = utils.Vec3F32ToF64(model.Translation())
	e.LocalRotation = utils.QuatF32ToF64(model.Rotation())
	e.Scale = utils.Vec3F32ToF64(model.Scale())

	// animation setup
	e.Animations = model.Animations()
	if len(e.Animations) > 0 {
		e.AnimationPlayer = animation.NewAnimationPlayer(model)
	}

	return e
}

func InstantiateBaseEntity(name string, id int) *Entity {
	return &Entity{
		ID:   id,
		Name: name,

		Children: map[int]*Entity{},

		LocalPosition: mgl64.Vec3{0, 0, 0},
		LocalRotation: mgl64.QuatIdent(),
		Scale:         mgl64.Vec3{1, 1, 1},
	}
}

func ComputeParentAndJointTransformMatrix(entity *Entity) mgl64.Mat4 {
	parentModelMatrix := mgl64.Ident4()
	animModelMatrix := mgl64.Ident4()
	if entity.Parent != nil {
		parentModelMatrix = ComputeTransformMatrix(entity.Parent)

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

func CreateDummy(name string) *Entity {
	entity := InstantiateBaseEntity(name, id)
	id += 1
	return entity
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

func SetNextID(nextID int) {
	id = nextID
}
