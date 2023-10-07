package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"

	"github.com/kkevinchou/izzet/izzet/izzetdata"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var id int

type Entity struct {
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
	Animation *AnimationComponent

	// dirty flag caching world transform
	DirtyTransformFlag   bool
	cachedWorldTransform mgl64.Mat4 // TODO: initialize to identity

	// each Entity has their own transforms and animation player
	LocalPosition mgl64.Vec3
	LocalRotation mgl64.Quat
	LocalScale    mgl64.Vec3

	MeshComponent       *MeshComponent
	InternalBoundingBox *collider.BoundingBox

	// relationships
	Parent   *Entity         `json:"-"`
	Children map[int]*Entity `json:"-"`
}

func (e *Entity) GetID() int {
	return e.ID
}

func (e *Entity) Dirty() bool {
	return e.DirtyTransformFlag
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

func InstantiateFromPrefab(prefab *prefabs.Prefab, ml *modellibrary.ModelLibrary) []*Entity {
	return CreateEntitiesFromDocument(prefab.Document, ml, prefab.IzzetData)
}

func InstantiateBaseEntity(name string, id int) *Entity {
	return &Entity{
		ID:   id,
		Name: name,

		Children: map[int]*Entity{},

		DirtyTransformFlag:   true,
		LocalPosition:        mgl64.Vec3{0, 0, 0},
		LocalRotation:        mgl64.QuatIdent(),
		LocalScale:           mgl64.Vec3{1, 1, 1},
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
	if e.InternalBoundingBox == nil {
		return nil
	}
	modelMatrix := WorldTransform(e)
	// t, r, s := utils.DecomposeF64(modelMatrix)
	// translation := mgl64.Translate3D(t.X(), t.Y(), t.Z())
	// scale := mgl64.Scale3D(s.X(), s.Y(), s.Z())

	// return e.boundingBox.Transform(translation.Mul4(r.Mat4()).Mul4(scale))
	return e.InternalBoundingBox.Transform(modelMatrix)
}

func CreateEntitiesFromDocument(document *modelspec.Document, ml *modellibrary.ModelLibrary, data *izzetdata.Data) []*Entity {
	var result []*Entity

	entityAsset := data.EntityAssets[document.Name]

	if entityAsset.SingleEntity {
		handle := modellibrary.NewGlobalHandle(document.Name)
		// entity := InstantiateEntity(document.Name)
		// entity.MeshComponent = &MeshC
		for _, scene := range document.Scenes {
			if len(scene.Nodes) > 1 {
				panic("single entity asset loading only supports a singular root entity")
			}
			for _, node := range scene.Nodes {
				entity := InstantiateEntity(document.Name)
				entity.MeshComponent = &MeshComponent{MeshHandle: handle}
				var vertices []modelspec.Vertex
				VerticesFromNode(node, document, &vertices)
				boundingBox := *collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
				entity.InternalBoundingBox = &boundingBox
				SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
				SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
				SetScale(entity, utils.Vec3F32ToF64(node.Scale))

				if len(document.Animations) > 0 {
					animations, joints := ml.GetAnimations(document.Name)
					animationPlayer := animation.NewAnimationPlayer()
					animationPlayer.Initialize(animations, joints[document.RootJoint.ID])
					entity.Animation = &AnimationComponent{RootJointID: document.RootJoint.ID, AnimationHandle: document.Name, AnimationPlayer: animationPlayer}
				}
				result = append(result, entity)
			}
		}
	} else {
		parent := InstantiateEntity(fmt.Sprintf("%s-parent", document.Name))
		result = append(result, parent)

		for _, scene := range document.Scenes {
			for _, node := range scene.Nodes {
				result = append(result, parseEntities(node, nil, document.Name, document, ml)...)
			}
		}

		var rootEntities []*Entity
		for _, e := range result {
			if e.Parent == nil {
				rootEntities = append(rootEntities, e)
			}
		}

		// only parent root entities
		for _, e := range rootEntities {
			if e.ID == parent.ID {
				continue
			}

			parent.Children[e.ID] = e
			e.Parent = parent
		}
	}

	if len(result) > 0 {
		rootEntity := result[0]
		if entityAsset.Translation != nil {
			SetLocalPosition(rootEntity, *entityAsset.Translation)
		}
		if entityAsset.Rotation != nil {
			SetLocalRotation(rootEntity, *entityAsset.Rotation)
		}
		if entityAsset.Scale != nil {
			SetScale(rootEntity, *entityAsset.Scale)
		}
	}

	return result
}

func parseEntities(node *modelspec.Node, parent *Entity, namespace string, document *modelspec.Document, ml *modellibrary.ModelLibrary) []*Entity {
	var entity *Entity

	if node.MeshID != nil {
		entity = InstantiateEntity(node.Name)
		entity.MeshComponent = &MeshComponent{MeshHandle: modellibrary.NewHandleFromMeshID(namespace, *node.MeshID)}
		var vertices []modelspec.Vertex
		VerticesFromNode(node, document, &vertices)
		boundingBox := *collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
		entity.InternalBoundingBox = &boundingBox
		SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
		SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
		SetScale(entity, utils.Vec3F32ToF64(node.Scale))

		if len(document.Animations) > 0 {
			animations, joints := ml.GetAnimations(document.Name)
			animationPlayer := animation.NewAnimationPlayer()
			animationPlayer.Initialize(animations, joints[document.RootJoint.ID])
			entity.Animation = &AnimationComponent{RootJointID: document.RootJoint.ID, AnimationHandle: document.Name, AnimationPlayer: animationPlayer}
		}
	}

	allEntities := []*Entity{}
	if entity != nil {
		allEntities = append(allEntities, entity)
	}

	for _, childNode := range node.Children {
		cs := parseEntities(childNode, entity, namespace, document, ml)
		// the first element of parseEntities is the root child node
		if entity != nil {
			if cs[0] != nil {
				cs[0].Parent = entity
				entity.Children[cs[0].ID] = cs[0]
			}
		}
		allEntities = append(allEntities, cs...)
	}

	return allEntities
}

func VerticesFromNode(node *modelspec.Node, document *modelspec.Document, out *[]modelspec.Vertex) {
	if node.MeshID != nil {
		mesh := document.Meshes[*node.MeshID]
		for _, p := range mesh.Primitives {
			*out = append(*out, p.UniqueVertices...)
		}
	}

	for _, childNode := range node.Children {
		VerticesFromNode(childNode, document, out)
	}
}
