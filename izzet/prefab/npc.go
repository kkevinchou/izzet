package prefab

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

func CreateNPC(app App, entityType entity.EntityType) *entity.Entity {
	var modelName string
	if entityType == entity.EntityTypeVelociraptor {
		modelName = "velociraptor"
	} else if entityType == entity.EntityTypeParasaurolophus {
		modelName = "parasaurolophus"
	} else {
		panic(fmt.Sprintf("unexpected entity type %s", entityType))
	}

	handle := assets.NewSingleEntityMeshHandle(modelName)
	e := entity.CreateEmptyEntity(modelName)
	e.Kinematic = &entity.KinematicComponent{GravityEnabled: true, Speed: settings.CharacterSpeed}

	capsule := collider.NewCapsule(mgl64.Vec3{0, 3, 0}, mgl64.Vec3{0, 1, 0}, 1)
	e.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
	e.Collider.CapsuleCollider = &capsule

	e.MeshComponent = &entity.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
	e.Animation = entity.NewAnimationComponent(modelName, app.AssetManager())
	e.AttackComponent = &entity.AttackComponent{AttackRange: 4}

	entity.SetScale(e, mgl64.Vec3{0.5, 0.5, 0.5})

	e.AIComponent = &entity.AIComponent{}

	return e
}
