package client

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

func (g *Client) InstantiateEntity(documentAsset assets.DocumentAsset) *entities.Entity {
	if !documentAsset.Config.SingleEntity {
		spawnedEntities := g.createEntitiesFromDocument(documentAsset)
		for _, entity := range spawnedEntities {
			g.world.AddEntity(entity)
		}
		return spawnedEntities[0]
	}

	namespace := documentAsset.Config.Name
	document := g.AssetManager().GetDocument(namespace)
	handle := assets.NewSingleMeshHandle(namespace)
	if len(document.Scenes) != 1 {
		panic("single entity asset loading only supports a singular scene")
	}

	scene := document.Scenes[0]
	node := scene.Nodes[0]

	entity := g.createEntity(documentAsset, namespace, handle, node)
	g.world.AddEntity(entity)

	return entity
}

func (g *Client) createEntity(documentAsset assets.DocumentAsset, name string, meshHandle types.MeshHandle, node *modelspec.Node) *entities.Entity {
	document := documentAsset.Document
	config := documentAsset.Config
	entity := entities.InstantiateEntity(name)

	entity.MeshComponent = &entities.MeshComponent{
		MeshHandle:    meshHandle,
		Transform:     mgl64.Ident4(),
		Visible:       true,
		ShadowCasting: true,
	}

	var vertices []modelspec.Vertex
	entities.VerticesFromNode(node, document, &vertices)
	entity.InternalBoundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
	entities.SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
	entities.SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
	entities.SetScale(entity, utils.Vec3F32ToF64(node.Scale))

	entity.Static = config.Static
	if config.Physics {
		entity.Physics = &entities.PhysicsComponent{}
	}
	if types.ColliderType(config.ColliderType) == types.ColliderTypeMesh {
		primitives := g.assetManager.GetPrimitives(meshHandle)
		entity.Collider = &entities.ColliderComponent{ColliderGroup: types.ConvertGroupToFlag(types.ColliderGroup(config.ColliderGroup))}
		entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(entities.AssetPrimitiveToSpecPrimitive(primitives))
	}
	return entity
}

func (g *Client) createEntitiesFromDocument(documentAsset assets.DocumentAsset) []*entities.Entity {
	document := documentAsset.Document

	var spawnedEntities []*entities.Entity
	parent := entities.InstantiateEntity(fmt.Sprintf("%s-parent", document.Name))
	spawnedEntities = append(spawnedEntities, parent)

	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			spawnedEntities = append(spawnedEntities, g.createEntitiesFromNode(documentAsset, node, document.Name)...)
		}
	}

	var rootEntities []*entities.Entity
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

	return spawnedEntities
}

func (g *Client) createEntitiesFromNode(documentAsset assets.DocumentAsset, node *modelspec.Node, namespace string) []*entities.Entity {
	var entity *entities.Entity

	if node.MeshID != nil {
		meshHandle := assets.NewMeshHandle(namespace, fmt.Sprintf("%d", *node.MeshID))
		entity = g.createEntity(documentAsset, node.Name, meshHandle, node)
	}

	allEntities := []*entities.Entity{}
	if entity != nil {
		allEntities = append(allEntities, entity)
	}

	for _, childNode := range node.Children {
		cs := g.createEntitiesFromNode(documentAsset, childNode, namespace)
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
