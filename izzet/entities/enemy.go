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

func NewEnemy() *EntityImpl {
	modelName := "mutant"
	assetManager := directory.GetDirectory().AssetManager()

	transformComponent := &components.TransformComponent{
		Position:    mgl64.Vec3{78, 78, -73},
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
		Scale:       mgl64.Ident4(),
		Orientation: yr,
		Model:       m,
	}

	// capsule := collider.NewCapsule(mgl64.Vec3{0, 18, 0}, mgl64.Vec3{0, 6, 0}, 6)
	capsule := collider.NewCapsuleFromModel(m)
	boundingBox := collider.BoundingBoxFromCapsule(capsule)

	colliderComponent := &components.ColliderComponent{
		BoundingBoxCollider: boundingBox,
		CapsuleCollider:     &capsule,
		Contacts:            map[int]bool{},
	}

	entityComponents := []components.Component{
		&components.MovementComponent{},
		&components.NetworkComponent{},
		transformComponent,
		animationComponent,
		meshComponent,
		colliderComponent,
		renderComponent,
		components.NewAIComponent(nil),
		components.NewHealthComponent(100),
		components.DefaultLootDropper(),
	}

	entity := NewEntity(
		"enemy",
		types.EntityTypeEnemy,
		components.NewComponentContainer(entityComponents...),
	)

	return entity
}
