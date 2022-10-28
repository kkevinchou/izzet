package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/model"
)

var (
	defaultXR          = mgl64.QuatRotate(mgl64.DegToRad(-90), mgl64.Vec3{1, 0, 0}).Mat4()
	defaultYR          = mgl64.QuatRotate(mgl64.DegToRad(180), mgl64.Vec3{0, 1, 0}).Mat4()
	defaultOrientation = defaultYR.Mul4(defaultXR)
	defaultScale       = mgl64.Scale3D(25, 25, 25)
)

func NewScene() *EntityImpl {
	return NewRigidBody("scene", mgl64.Ident4(), mgl64.Ident4(), types.EntityTypeScene, true)
	// return NewRigidBody("scene_giga_flat", mgl64.Ident4(), mgl64.Ident4(), types.EntityTypeScene)
}

func NewStaticRigidBody() *EntityImpl {
	return NewRigidBody("town_center", mgl64.Ident4(), mgl64.Ident4(), types.EntityTypeStaticRigidBody, false)
}

func NewDynamicRigidBody() *EntityImpl {
	return NewRigidBody("guard", mgl64.Ident4(), mgl64.Ident4(), types.EntityTypeDynamicRigidBody, true)
}

func NewRigidBody(modelName string, Scale mgl64.Mat4, Orientation mgl64.Mat4, entityType types.EntityType, useMeshCollider bool) *EntityImpl {
	transformComponent := &components.TransformComponent{
		Orientation: mgl64.QuatIdent(),
	}

	assetManager := directory.GetDirectory().AssetManager()
	modelSpec := assetManager.GetModel(modelName)

	m := model.NewModel(modelSpec)

	meshComponent := &components.MeshComponent{
		Scale:       Scale,
		Orientation: Orientation,
		Model:       m,
	}

	colliderComponent := &components.ColliderComponent{
		Contacts: map[int]bool{},
	}

	var boundingBox *collider.BoundingBox
	if useMeshCollider {
		triMesh := collider.NewTriMesh(m)
		boundingBox = collider.BoundingBoxFromModel(m)
		colliderComponent.TriMeshCollider = &triMesh
	} else {
		capsule := collider.NewCapsuleFromModel(m)
		boundingBox = collider.BoundingBoxFromCapsule(capsule)
		colliderComponent.CapsuleCollider = &capsule
	}
	colliderComponent.BoundingBoxCollider = boundingBox

	renderComponent := &components.RenderComponent{
		IsVisible: true,
	}

	physicsComponent := &components.PhysicsComponent{
		Static: true,
	}

	componentList := []components.Component{
		transformComponent,
		renderComponent,
		&components.NetworkComponent{},
		meshComponent,
		colliderComponent,
		physicsComponent,
	}

	entity := NewEntity(
		"rigidbody",
		entityType,
		components.NewComponentContainer(componentList...),
	)

	return entity
}
