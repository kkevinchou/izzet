package serversystems

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
)

type SpawnerSystem struct {
	app App
}

func NewSpawnerSystem(app App) *SpawnerSystem {
	return &SpawnerSystem{app: app}
}

func (s *SpawnerSystem) Name() string {
	return "SpawnerSystem"
}

func (s *SpawnerSystem) Update(delta time.Duration, world systems.GameWorld) {
	// var radius float64 = 40
	// var length float64 = 80

	for _, player := range s.app.GetPlayers() {
		frameInput := s.app.GetPlayerInput(player.ID)
		if frameInput.KeyboardInput[input.KeyboardKeyJ].Event == input.KeyboardEventUp {
			handle := modellibrary.NewGlobalHandle("vampire2")
			entity := entities.InstantiateEntity("vampire2")
			entity.Physics = &entities.PhysicsComponent{GravityEnabled: true}
			entity.Collider = &entities.ColliderComponent{
				// CapsuleCollider: &collider.Capsule{
				// 	Radius: radius,
				// 	Top:    mgl64.Vec3{0, radius + length, 0},
				// 	Bottom: mgl64.Vec3{0, radius, 0},
				// },
				ColliderGroup: entities.ColliderGroupFlagPlayer,
				CollisionMask: entities.ColliderGroupFlagTerrain | entities.ColliderGroupFlagPlayer,
			}

			primitives := s.app.ModelLibrary().GetPrimitives(handle)
			verts := modellibrary.UniqueVerticesFromPrimitives(primitives)
			c := collider.NewCapsuleFromVertices(verts)
			entity.Collider.CapsuleCollider = &c

			capsule := entity.Collider.CapsuleCollider
			entity.InternalBoundingBox = collider.BoundingBox{MinVertex: capsule.Bottom.Sub(mgl64.Vec3{c.Radius, c.Radius, c.Radius}), MaxVertex: capsule.Top.Add(mgl64.Vec3{c.Radius, c.Radius, c.Radius})}

			entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
			entity.Animation = entities.NewAnimationComponent("vampire2", s.app.ModelLibrary())
			entities.SetScale(entity, mgl64.Vec3{0.25, 0.25, 0.25})
			entities.SetLocalPosition(entity, mgl64.Vec3{0, 10, 0})
			// entity.Movement = &entities.MovementComponent{PatrolConfig: &entities.PatrolConfig{Points: []mgl64.Vec3{{0, 10, 0}, {-300, 10, 0}}}, Speed: 100}
			entity.AIComponent = &entities.AIComponent{
				Speed:        100,
				TargetConfig: &entities.TargetConfig{},
			}

			s.app.EventsManager().EntitySpawnTopic.Write(events.EntitySpawnEvent{Entity: entity})
		}
	}
}
