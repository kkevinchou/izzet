package prefab

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

func createPlayer(am *assets.AssetManager) *entity.Entity {
	e := entity.InstantiateBaseEntity("player", 0)
	e.Kinematic = &entity.KinematicComponent{GravityEnabled: true, Speed: settings.CharacterSpeed}

	var radius float64 = settings.EntityCapsuleColliderRadius
	var length float64 = settings.EntityCapsuleColliderLength
	capsule := collider.Capsule{
		Radius: radius,
		Top:    mgl64.Vec3{0, radius + length, 0},
		Bottom: mgl64.Vec3{0, radius, 0},
	}

	e.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
	e.CharacterControllerComponent = &entity.CharacterControllerComponent{CameraEntityID: entity.InvalidEntityID}
	e.AimDownSightsComponent = &entity.AimDownSightsComponent{}
	e.HealthComponent = &entity.HealthComponent{Amount: 100}
	meshHandle := am.GetSingleEntityMeshHandle("mannequin_m")

	e.MeshComponent = &entity.MeshComponent{MeshHandle: meshHandle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true, InvisibleToPlayerOwner: settings.FirstPersonCamera}
	handle := am.GetAnimationHandle("mannequin_m")
	e.Animation = entity.NewAnimationComponent(am, handle, animation.StateMachineIDPlayer, entity.AnimationModeStateMachine)
	e.RenderBlend = &entity.RenderBlend{}
	entity.SetScale(e, mgl64.Vec3{1, 1, 1})

	return e
}
