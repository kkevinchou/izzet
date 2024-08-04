package serversystems

import (
	"math"
	"math/rand"

	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
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
			modelName := "velociraptor"
			handle := assets.NewGlobalHandle(modelName)
			entity := entities.InstantiateEntity(modelName)
			entity.Physics = &entities.PhysicsComponent{GravityEnabled: true, RotateOnVelocity: false}
			entity.Collider = &entities.ColliderComponent{
				ColliderGroup: entities.ColliderGroupFlagPlayer,
				CollisionMask: entities.ColliderGroupFlagTerrain | entities.ColliderGroupFlagPlayer,
			}

			// primitives := s.app.AssetManager().GetPrimitives(handle)
			// verts := assets.UniqueVerticesFromPrimitives(primitives)
			// c := collider.NewCapsuleFromVertices(verts)
			c := collider.NewCapsule(mgl64.Vec3{0, 4, 0}, mgl64.Vec3{0, 2, 0}, 2)
			entity.Collider.CapsuleCollider = &c

			capsule := entity.Collider.CapsuleCollider
			entity.InternalBoundingBox = collider.BoundingBox{MinVertex: capsule.Bottom.Sub(mgl64.Vec3{c.Radius, c.Radius, c.Radius}), MaxVertex: capsule.Top.Add(mgl64.Vec3{c.Radius, c.Radius, c.Radius})}

			entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
			entity.Animation = entities.NewAnimationComponent(modelName, s.app.AssetManager())
			// entities.SetScale(entity, mgl64.Vec3{0.01, 0.01, 0.01})

			jitterX := rand.Intn(10)
			jitterZ := rand.Intn(10)
			entities.SetLocalPosition(entity, mgl64.Vec3{float64(jitterX), 60, float64(jitterZ)})

			entity.AIComponent = &entities.AIComponent{
				Speed: 25,
				// TargetConfig: &entities.TargetConfig{},
				// PatrolConfig: &entities.PatrolConfig{Points: []mgl64.Vec3{{0, 10, 0}, {100, 10, 0}}},
				PathfindConfig: &entities.PathfindConfig{},
				// AttackConfig:   &entities.AttackConfig{},
			}

			world := s.app.World()
			for _, e := range world.Entities() {
				if e.SpawnPointComponent != nil {
					entities.SetLocalPosition(entity, e.Position())
					entities.SetLocalPosition(entity, mgl64.Vec3{e.Position().X() + float64(jitterX), 60, e.Position().Z() + float64(jitterZ)})
					break
				}
			}

			s.app.EventsManager().EntitySpawnTopic.Write(events.EntitySpawnEvent{Entity: entity})
		}
	}
}
