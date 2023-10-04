package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/model"
	"github.com/kkevinchou/izzet/izzet/modellibrary"

	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var id int

type Entity struct {
	// THINGS TO REMOVE
	Model           model.RenderModel
	AnimationPlayer *animation.AnimationPlayer
	Animations      map[string]*modelspec.AnimationSpec
	RootJoint       *modelspec.JointSpec

	ID        int
	Name      string
	Billboard bool
	Physics   *PhysicsComponent
	Collider  *ColliderComponent
	Movement  *MovementComponent
	Particles *ParticleGenerator
	IsSocket  bool
	LightInfo *LightInfo
	ImageInfo *ImageInfo
	ShapeData []*ShapeData
	Material  *MaterialComponent

	// dirty flag caching world transform
	dirtyTransformFlag   bool
	cachedWorldTransform mgl64.Mat4 // TODO: initialize to identity

	// each Entity has their own transforms and animation player
	localPosition mgl64.Vec3
	localRotation mgl64.Quat
	scale         mgl64.Vec3

	MeshComponent *MeshComponent
	boundingBox   *collider.BoundingBox

	// relationships
	Parent      *Entity
	Children    map[int]*Entity
	ParentJoint *modelspec.JointSpec
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

// func (e *Entity) BoundingBox() *collider.BoundingBox {
// 	if e.boundingBox == nil {
// 		return nil
// 	}
// 	modelMatrix := WorldTransform(e)
// 	// t, r, s := utils.DecomposeF64(modelMatrix)
// 	// translation := mgl64.Translate3D(t.X(), t.Y(), t.Z())
// 	// scale := mgl64.Scale3D(s.X(), s.Y(), s.Z())

// 	// return e.boundingBox.Transform(translation.Mul4(r.Mat4()).Mul4(scale))
// 	return e.boundingBox.Transform(modelMatrix)
// }

func InstantiateFromPrefab(prefab *prefabs.Prefab) []*Entity {
	return CreateEntitiesFromDocument(prefab.Document)
}

func InstantiateBaseEntity(name string, id int) *Entity {
	return &Entity{
		ID:   id,
		Name: name,

		Children: map[int]*Entity{},

		dirtyTransformFlag:   true,
		localPosition:        mgl64.Vec3{0, 0, 0},
		localRotation:        mgl64.QuatIdent(),
		scale:                mgl64.Vec3{1, 1, 1},
		cachedWorldTransform: mgl64.Ident4(),
	}
}

func InstantiateEntity(name string) *Entity {
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

func CreateEntitiesFromDocument(document *modelspec.Document) []*Entity {
	// modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}
	var result []*Entity

	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			entity := InstantiateEntity(node.Name)

			rootNode := parseNode(node, true, mgl32.Ident4(), document.Name)
			entity.MeshComponent = &MeshComponent{Node: rootNode}

			if len(document.Animations) > 0 {
				entity.RootJoint = document.RootJoint
				entity.Animations = document.Animations
				entity.AnimationPlayer = animation.NewAnimationPlayer(entity.Animations, entity.RootJoint)
			}

			// entity.MeshComponent

			// modellibrary.GetPrimitives()
			// modellibrary.UniqueVerticesFromPrimitives()

			// entity.boundingBox =

			SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
			SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
			SetScale(entity, utils.Vec3F32ToF64(node.Scale))

			result = append(result, entity)
		}
	}

	return result
}

func parseNode(node *modelspec.Node, ignoreTransform bool, parentTransform mgl32.Mat4, namespace string) Node {
	transform := node.Transform
	if ignoreTransform {
		transform = mgl32.Ident4()
	}
	transform = parentTransform.Mul4(transform)

	eNode := Node{
		Transform: transform,
	}

	if node.MeshID != nil {
		eNode.MeshHandle = modellibrary.NewHandle(namespace, *node.MeshID)
	}

	var children []Node
	for _, childNode := range node.Children {
		children = append(children, parseNode(childNode, false, transform, namespace))
	}

	eNode.Children = children
	return eNode
}
