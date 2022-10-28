package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/model"
)

func NewLootbox() *EntityImpl {
	modelName := "lootbox"
	assetManager := directory.GetDirectory().AssetManager()

	transformComponent := &components.TransformComponent{
		Position:    mgl64.Vec3{0, 50, 100},
		Orientation: mgl64.QuatIdent(),
	}

	renderComponent := &components.RenderComponent{
		IsVisible: true,
	}

	modelSpec := assetManager.GetModel(modelName)
	m := model.NewModel(modelSpec)

	yr := mgl64.QuatRotate(mgl64.DegToRad(180), mgl64.Vec3{0, 1, 0}).Mat4()
	meshComponent := &components.MeshComponent{
		Scale:       mgl64.Ident4(),
		Orientation: yr,
		Model:       m,
	}

	// triMeshCollider := collider.NewTriMesh(m)
	// boundingBox := collider.BoundingBoxFromModel(m)
	capsuleCollider := collider.NewCapsuleFromModel(m)
	boundingBox := collider.BoundingBoxFromCapsule(capsuleCollider)

	colliderComponent := &components.ColliderComponent{
		CapsuleCollider: &capsuleCollider,
		// TriMeshCollider:     &triMeshCollider,
		BoundingBoxCollider: boundingBox,
		Contacts:            map[int]bool{},
	}

	physicsComponent := &components.PhysicsComponent{
		Impulses: map[string]types.Impulse{},
	}

	entityComponents := []components.Component{
		&components.NetworkComponent{},
		transformComponent,
		meshComponent,
		colliderComponent,
		renderComponent,
		physicsComponent,
		&components.LootComponent{},
	}

	entity := NewEntity(
		"lootbox",
		types.EntityTypeLootbox,
		components.NewComponentContainer(entityComponents...),
	)

	return entity
}
