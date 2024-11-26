package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var id int

func InstantiateFromPrefab(prefab *prefabs.Prefab, ml *assets.AssetManager) []*Entity {
	return nil
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

func GetNextIDAndAdvance() int {
	oldID := id
	id += 1
	return oldID
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

func CreateEntitiesFromDocument(documentAsset assets.DocumentAsset, ml *assets.AssetManager) []*Entity {
	document := documentAsset.Document
	config := documentAsset.Config

	var spawnedEntities []*Entity
	parent := InstantiateEntity(fmt.Sprintf("%s-parent", document.Name))
	spawnedEntities = append(spawnedEntities, parent)

	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			spawnedEntities = append(spawnedEntities, parseEntities(node, nil, document.Name, document, ml)...)
		}
	}

	var rootEntities []*Entity
	for _, e := range spawnedEntities {
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

	for _, entity := range spawnedEntities {
		entity.Static = config.Static
		if config.Physics {
			entity.Physics = &PhysicsComponent{}
		}
		if types.ColliderType(config.ColliderType) == types.ColliderTypeMesh {
			if entity.MeshComponent == nil {
				continue
			}
			meshHandle := entity.MeshComponent.MeshHandle
			primitives := ml.GetPrimitives(meshHandle)
			entity.Collider = &ColliderComponent{ColliderGroup: types.ConvertGroupToFlag(types.ColliderGroup(config.ColliderGroup))}
			entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(MLPrimitivesTospecPrimitive(primitives))
		}
	}

	// if len(spawnedEntities) > 0 {
	// 	rootEntity := spawnedEntities[0]
	// 	if entityAsset.Translation != nil {
	// 		SetLocalPosition(rootEntity, *entityAsset.Translation)
	// 	}
	// 	if entityAsset.Rotation != nil {
	// 		SetLocalRotation(rootEntity, *entityAsset.Rotation)
	// 	}
	// 	if entityAsset.Scale != nil {
	// 		SetScale(rootEntity, *entityAsset.Scale)
	// 	}
	// }

	// return spawnedEntities
	return spawnedEntities
}

func MLPrimitivesTospecPrimitive(primitives []assets.Primitive) []*modelspec.PrimitiveSpecification {
	var result []*modelspec.PrimitiveSpecification
	for _, p := range primitives {
		result = append(result, p.Primitive)
	}
	return result
}

func parseEntities(node *modelspec.Node, parent *Entity, namespace string, document *modelspec.Document, ml *assets.AssetManager) []*Entity {
	var entity *Entity

	if node.MeshID != nil {
		entity = InstantiateEntity(node.Name)
		meshHandle := assets.NewHandleFromMeshID(namespace, *node.MeshID)
		entity.MeshComponent = &MeshComponent{MeshHandle: meshHandle, Transform: mgl64.Ident4(), Visible: true, ShadowCasting: true}
		var vertices []modelspec.Vertex
		VerticesFromNode(node, document, &vertices)
		entity.InternalBoundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
		SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
		SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
		SetScale(entity, utils.Vec3F32ToF64(node.Scale))

		if len(document.Animations) > 0 {
			entity.Animation = NewAnimationComponent(document.Name, ml)
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
