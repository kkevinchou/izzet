package client

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/physics"
	"github.com/kkevinchou/izzet/izzet/entity"
)

const (
	physicsCubeSize   = 1.0
	physicsSpawnAbove = 20.0
)

type physicsState struct {
	world                   *physics.World
	bodyEntities            map[physics.BodyID]int
	staticBodiesInitialized bool
}

func newPhysicsState() *physicsState {
	return &physicsState{
		world:        physics.NewWorld(),
		bodyEntities: map[physics.BodyID]int{},
	}
}

func (g *Client) resetPhysics() {
	g.physics = newPhysicsState()
}

func (g *Client) SpawnPhysicsCube(contactPoint mgl64.Vec3) {
	if g.physics == nil {
		g.resetPhysics()
	}
	g.ensurePhysicsStaticBodies()

	position := contactPoint.Add(mgl64.Vec3{0, physicsSpawnAbove, 0})
	bodyOptions := physics.DefaultBodyOptions(1)
	bodyOptions.Position = position
	bodyOptions.Restitution = 0.05
	bodyOptions.Friction = 0.9
	bodyOptions.LinearDamping = 0.05
	bodyOptions.AngularDamping = 0.1

	bodyID, err := g.physics.world.CreateCubeWithOptions(physics.CubeOptions{
		BodyOptions: bodyOptions,
		Size:        mgl64.Vec3{physicsCubeSize, physicsCubeSize, physicsCubeSize},
	})
	if err != nil {
		g.Logger().Error("failed to create physics cube body", "error", err)
		return
	}

	cube := entity.CreateCube(g.AssetManager(), physicsCubeSize)
	cube.Name = "physics_cube"
	cube.Collider.TriMeshCollider = nil
	cube.Collider.SimplifiedTriMeshCollider = nil
	cube.Collider.CapsuleCollider = nil
	cube.Physics = nil
	cube.Static = false
	entity.SetLocalPosition(cube, position)
	g.world.AddEntity(cube)

	g.physics.bodyEntities[bodyID] = cube.ID
}

func (g *Client) ensurePhysicsStaticBodies() {
	if g.physics.staticBodiesInitialized {
		return
	}
	g.physics.staticBodiesInitialized = true

	startingCube := g.findPhysicsStartingCube()
	if startingCube == nil {
		g.Logger().Warn("could not find starting cube mesh for physics static body")
		return
	}

	bodyOptions := physics.DefaultBodyOptions(0)
	bodyOptions.Position = startingCube.Position()
	bodyOptions.Rotation = startingCube.Rotation()
	bodyOptions.Static = true

	_, err := g.physics.world.CreateCubeWithOptions(physics.CubeOptions{
		BodyOptions: bodyOptions,
		Size:        startingCube.Scale(),
	})
	if err != nil {
		g.Logger().Error("failed to create physics static cube body", "error", err)
	}
}

func (g *Client) findPhysicsStartingCube() *entity.Entity {
	defaultCubeHandle := g.AssetManager().DefaultCubeHandle()
	for _, e := range g.world.Entities() {
		if e == nil || e.MeshComponent == nil || !e.Static {
			continue
		}
		if e.Name != "cube" {
			continue
		}
		if e.MeshComponent.MeshHandle == defaultCubeHandle {
			return e
		}
	}
	return nil
}

func (g *Client) stepPhysics(delta time.Duration) {
	if g.physics == nil || g.physics.world == nil {
		return
	}

	g.physics.world.Step(delta)
	for bodyID, entityID := range g.physics.bodyEntities {
		body, ok := g.physics.world.Body(bodyID)
		if !ok {
			delete(g.physics.bodyEntities, bodyID)
			continue
		}

		e := g.world.GetEntityByID(entityID)
		if e == nil {
			g.physics.world.RemoveBody(bodyID)
			delete(g.physics.bodyEntities, bodyID)
			continue
		}

		transform := body.Transform()
		entity.SetLocalPosition(e, transform.Position)
		e.SetLocalRotation(transform.Rotation)
	}
}
