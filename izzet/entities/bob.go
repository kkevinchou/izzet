package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/lib/animation"
	"github.com/kkevinchou/izzet/lib/collision/collider"
	"github.com/kkevinchou/izzet/lib/model"
)

func NewBob() *EntityImpl {
	modelName := "alpha"
	assetManager := directory.GetDirectory().AssetManager()

	transformComponent := &components.TransformComponent{
		Position:    mgl64.Vec3{0, 0, 70},
		Orientation: mgl64.QuatIdent(),
	}

	renderComponent := &components.RenderComponent{
		IsVisible: true,
	}

	modelSpec := assetManager.GetModel(modelName)
	m := model.NewModel(modelSpec)

	animationPlayer := animation.NewAnimationPlayer(m)
	animationPlayer.PlayAnimation("Idle")
	animationComponent := &components.AnimationComponent{
		Player: animationPlayer,
	}
	_ = animationComponent

	yr := mgl64.QuatRotate(mgl64.DegToRad(180), mgl64.Vec3{0, 1, 0}).Mat4()

	meshComponent := &components.MeshComponent{
		// Scale:            mgl64.Scale3D(1, 1, 1),
		// Scale: mgl64.Scale3D(15, 15, 15),
		Scale: mgl64.Scale3D(1, 1, 1),
		// Orientation: mgl64.Ident4(),
		Orientation: yr,

		Model: m,
	}

	capsule := collider.NewCapsuleFromModel(m)
	boundingBox := collider.BoundingBoxFromCapsule(capsule)

	colliderComponent := &components.ColliderComponent{
		CapsuleCollider:     &capsule,
		BoundingBoxCollider: boundingBox,
		Contacts:            map[int]bool{},
	}

	thirdPersonControllerComponent := &components.ThirdPersonControllerComponent{
		Controlled: true,
	}

	entityComponents := []components.Component{
		&components.MovementComponent{},
		&components.NetworkComponent{},
		transformComponent,
		animationComponent,
		thirdPersonControllerComponent,
		meshComponent,
		colliderComponent,
		renderComponent,
		&components.NotepadComponent{LastAction: components.ActionNone},
		components.NewInventoryComponent(),
	}

	entity := NewEntity(
		"bob",
		types.EntityTypeBob,
		components.NewComponentContainer(entityComponents...),
	)

	return entity
}
