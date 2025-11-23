package client

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/types"
)

func (g *Client) CreateEntitiesFromDocumentAsset(documentAsset assets.DocumentAsset) *entity.Entity {
	if !documentAsset.Config.SingleEntity {
		spawnedEntities := g.createEntitiesFromDocument(documentAsset)
		for _, entity := range spawnedEntities {
			g.world.AddEntity(entity)
		}
		return spawnedEntities[0]
	}

	namespace := documentAsset.Config.Name
	document := documentAsset.Document
	handle := assets.NewSingleEntityMeshHandle(namespace)
	if len(document.Scenes) != 1 {
		panic("single entity asset loading only supports a singular scene")
	}

	scene := document.Scenes[0]
	node := scene.Nodes[0]

	e := g.createEntity(documentAsset, namespace, handle, node)
	g.world.AddEntity(e)

	if len(document.Animations) > 0 {
		e.Animation = entity.NewAnimationComponent(document.Name, g.assetManager)
	}

	return e
}

func (g *Client) createEntity(documentAsset assets.DocumentAsset, name string, meshHandle types.MeshHandle, node *modelspec.Node) *entity.Entity {
	document := documentAsset.Document
	config := documentAsset.Config
	e := entity.CreateEmptyEntity(name)

	e.MeshComponent = &entity.MeshComponent{
		MeshHandle:    meshHandle,
		Transform:     mgl64.Ident4(),
		Visible:       true,
		ShadowCasting: true,
	}

	var vertices []modelspec.Vertex
	entity.VerticesFromNode(node, document, &vertices)
	entity.SetLocalPosition(e, utils.Vec3F32ToF64(node.Translation))
	e.SetLocalRotation(utils.QuatF32ToF64(node.Rotation))
	entity.SetScale(e, utils.Vec3F32ToF64(node.Scale))

	e.Static = config.Static
	if config.Physics {
		e.Physics = &entity.PhysicsComponent{}
	}

	if types.ColliderType(config.ColliderType) == types.ColliderTypeMesh {
		primitives := g.assetManager.GetPrimitives(meshHandle)
		t := collider.CreateTriMeshFromPrimitives(entity.AssetPrimitiveToSpecPrimitive(primitives))
		bb := collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
		e.Collider = entity.CreateTriMeshColliderComponent(types.ConvertGroupToFlag(types.ColliderGroup(config.ColliderGroup)), 0, *t, nil, bb)
	}
	return e
}

func (g *Client) createEntitiesFromDocument(documentAsset assets.DocumentAsset) []*entity.Entity {
	document := documentAsset.Document

	var spawnedEntities []*entity.Entity
	parent := entity.CreateEmptyEntity(fmt.Sprintf("%s-parent", document.Name))
	spawnedEntities = append(spawnedEntities, parent)

	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			spawnedEntities = append(spawnedEntities, g.createEntitiesFromNode(documentAsset, node, document.Name)...)
		}
	}

	var rootEntities []*entity.Entity
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

func (g *Client) createEntitiesFromNode(documentAsset assets.DocumentAsset, node *modelspec.Node, namespace string) []*entity.Entity {
	var e *entity.Entity

	if node.MeshID != nil {
		meshHandle := assets.NewMeshHandle(namespace, fmt.Sprintf("%d", *node.MeshID))
		e = g.createEntity(documentAsset, node.Name, meshHandle, node)
	}

	allEntities := []*entity.Entity{}
	if e != nil {
		allEntities = append(allEntities, e)
	}

	for _, childNode := range node.Children {
		cs := g.createEntitiesFromNode(documentAsset, childNode, namespace)
		// the first element of parseEntities is the root child node
		if e != nil {
			if cs[0] != nil {
				cs[0].Parent = e
				e.Children[cs[0].ID] = cs[0]
			}
		}
		allEntities = append(allEntities, cs...)
	}

	return allEntities
}
