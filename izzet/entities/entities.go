package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/model"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var id int

type Entity struct {
	ID   int
	Name string

	// dirty flag caching world transform
	dirtyTransformFlag   bool
	cachedWorldTransform mgl64.Mat4

	// each Entity has their own transforms and animation player
	localPosition mgl64.Vec3
	localRotation mgl64.Quat
	scale         mgl64.Vec3

	// // prefab info for serialization
	// Prefab            *prefabs.Prefab
	// PrefabEntityIndex int

	// model

	Model       model.RenderModel
	boundingBox *collider.BoundingBox

	// relationships
	Parent      *Entity
	Children    map[int]*Entity
	ParentJoint *modelspec.JointSpec

	// animation
	Animations      map[string]*modelspec.AnimationSpec
	AnimationPlayer *animation.AnimationPlayer

	Physics   *PhysicsComponent
	Collider  *ColliderComponent
	Movement  *MovementComponent
	Particles *ParticleGenerator
	Billboard *BillboardInfo
	IsSocket  bool
	LightInfo *LightInfo
	ImageInfo *ImageInfo
	ShapeData []*ShapeData
	Material  *MaterialComponent
}

func (e *Entity) GetID() int {
	return e.ID
}

func (e *Entity) Dirty() bool {
	return e.dirtyTransformFlag
}

func (e *Entity) NameID() string {
	return fmt.Sprintf("%s-%d", e.Name, e.ID)
}

func (e *Entity) BoundingBox() *collider.BoundingBox {
	if e.boundingBox == nil {
		return nil
	}
	modelMatrix := WorldTransform(e)
	// t, r, s := utils.DecomposeF64(modelMatrix)
	// translation := mgl64.Translate3D(t.X(), t.Y(), t.Z())
	// scale := mgl64.Scale3D(s.X(), s.Y(), s.Z())

	// return e.boundingBox.Transform(translation.Mul4(r.Mat4()).Mul4(scale))
	return e.boundingBox.Transform(modelMatrix)
}

func InstantiateFromPrefab(prefab *prefabs.Prefab) []*Entity {
	var es []*Entity
	count := 0
	for prefabEntityIndex, modelRef := range prefab.ModelRefs() {
		model := modelRef.Model
		e := InstantiateFromPrefabStaticID(id, model, prefab, prefabEntityIndex)
		e.Name = modelRef.Name
		es = append(es, e)
		id += 1
		count++
	}
	return es
}

func InstantiateFromPrefabStaticID(id int, model *model.Model, prefab *prefabs.Prefab, prefabEntityIndex int) *Entity {
	e := InstantiateBaseEntity(model.Name(), id)
	// e.Prefab = prefab
	// e.PrefabEntityIndex = prefabEntityIndex
	e.Model = model
	e.boundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(model.Vertices()))

	SetLocalPosition(e, utils.Vec3F32ToF64(model.Translation()))
	SetLocalRotation(e, utils.QuatF32ToF64(model.Rotation()))
	SetScale(e, utils.Vec3F32ToF64(model.Scale()))

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

		dirtyTransformFlag: true,
		localPosition:      mgl64.Vec3{0, 0, 0},
		localRotation:      mgl64.QuatIdent(),
		scale:              mgl64.Vec3{1, 1, 1},
	}
}

func CreateDummy(name string) *Entity {
	entity := InstantiateBaseEntity(name, id)
	id += 1
	return entity
}

func SetNextID(nextID int) {
	id = nextID
}

func BuildRelation(parent *Entity, child *Entity) {
	RemoveParent(child)
	parent.Children[child.ID] = child
	child.Parent = parent
	SetDirty(child)
}

func RemoveParent(child *Entity) {
	if parent := child.Parent; parent != nil {
		delete(parent.Children, child.ID)
		child.Parent = nil
	}
}
