package prefab

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

func CreateNPC(app App, entityType entity.EntityType) *entity.Entity {
	var modelName string
	var scale float64 = 1
	if entityType == entity.EntityTypeVelociraptor {
		modelName = "velociraptor"
		scale = 0.5
	} else if entityType == entity.EntityTypeParasaurolophus {
		modelName = "parasaurolophus"
		scale = 0.5
	} else {
		panic(fmt.Sprintf("unexpected entity type %s", entityType))
	}

	meshHandle := app.AssetManager().GetSingleEntityMeshHandle(modelName)
	e := entity.CreateEmptyEntity(modelName)
	e.Kinematic = &entity.KinematicComponent{GravityEnabled: true, Speed: 7}
	e.AimDownSightsComponent = &entity.AimDownSightsComponent{}
	e.HealthComponent = &entity.HealthComponent{Amount: 100}

	var radius float64 = settings.EntityCapsuleColliderRadius * (1 / scale)
	var length float64 = settings.EntityCapsuleColliderLength * (1 / scale)
	capsule := collider.Capsule{
		Radius: radius,
		Top:    mgl64.Vec3{0, radius + length, 0},
		Bottom: mgl64.Vec3{0, radius, 0},
	}

	e.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
	e.Collider.CapsuleCollider = &capsule

	e.MeshComponent = &entity.MeshComponent{MeshHandle: meshHandle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
	e.Animation = entity.NewAnimationComponent(app.AssetManager().GetAnimationHandle(modelName), app.AssetManager())
	e.AttackComponent = &entity.AttackComponent{AttackRange: 3}
	entity.SetScale(e, mgl64.Vec3{scale, scale, scale})

	e.AIComponent = &entity.AIComponent{}

	return e
}
