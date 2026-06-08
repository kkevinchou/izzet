package prefab

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

func CreatePlayer(app App) *entity.Entity {
	var radius float64 = 0.4
	var length float64 = 1.0
	e := entity.CreateEmptyEntity("player")
	e.Kinematic = &entity.KinematicComponent{GravityEnabled: true, Speed: settings.CharacterSpeed}
	capsule := collider.Capsule{
		Radius: radius,
		Top:    mgl64.Vec3{0, radius + length, 0},
		Bottom: mgl64.Vec3{0, radius, 0},
	}
	e.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
	e.CharacterControllerComponent = &entity.CharacterControllerComponent{CameraEntityID: entity.InvalidEntityID}
	e.AimDownSightsComponent = &entity.AimDownSightsComponent{}
	handle := assets.NewSingleEntityMeshHandle("mannequin_m")

	e.MeshComponent = &entity.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true, InvisibleToPlayerOwner: settings.FirstPersonCamera}
	e.Animation = entity.NewAnimationComponent("mannequin_m", app.AssetManager())
	e.RenderBlend = &entity.RenderBlend{}
	entity.SetScale(e, mgl64.Vec3{1, 1, 1})

	return e
}
