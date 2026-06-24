package client

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/physics"
	"github.com/kkevinchou/izzet/izzet/entity"
)

const (
	physicsDebugCubeSize   = 1.0
	physicsDebugSpawnAbove = 20.0
)

type physicsDebugState struct {
	world                   *physics.World
	bodyEntities            map[physics.BodyID]int
	staticBodiesInitialized bool
}

func newPhysicsDebugState() *physicsDebugState {
	return &physicsDebugState{
		world:        physics.NewWorld(),
		bodyEntities: map[physics.BodyID]int{},
	}
}

func (g *Client) resetPhysicsDebug() {
	g.physicsDebug = newPhysicsDebugState()
}

func (g *Client) SpawnPhysicsDebugCube(contactPoint mgl64.Vec3) {
	if g.physicsDebug == nil {
		g.resetPhysicsDebug()
	}
	g.ensurePhysicsDebugStaticBodies()

	position := contactPoint.Add(mgl64.Vec3{0, physicsDebugSpawnAbove, 0})
	bodyOptions := physics.DefaultBodyOptions(1)
	bodyOptions.Position = position
	bodyOptions.Restitution = 0.05
	bodyOptions.Friction = 0.9
	bodyOptions.LinearDamping = 0.05
	bodyOptions.AngularDamping = 0.2

	bodyID, err := g.physicsDebug.world.CreateCubeWithOptions(physics.CubeOptions{
		BodyOptions: bodyOptions,
		Size:        mgl64.Vec3{physicsDebugCubeSize, physicsDebugCubeSize, physicsDebugCubeSize},
	})
	if err != nil {
		g.Logger().Error("failed to create physics debug cube body", "error", err)
		return
	}

	cube := entity.CreateCube(g.AssetManager(), physicsDebugCubeSize)
	cube.Name = "physics_debug_cube"
	cube.Collider.TriMeshCollider = nil
	cube.Collider.SimplifiedTriMeshCollider = nil
	cube.Collider.CapsuleCollider = nil
	cube.Physics = nil
	cube.Static = false
	entity.SetLocalPosition(cube, position)
	g.world.AddEntity(cube)

	g.physicsDebug.bodyEntities[bodyID] = cube.ID
}

func (g *Client) ensurePhysicsDebugStaticBodies() {
	if g.physicsDebug.staticBodiesInitialized {
		return
	}
	g.physicsDebug.staticBodiesInitialized = true

	startingCube := g.findPhysicsDebugStartingCube()
	if startingCube == nil {
		g.Logger().Warn("could not find starting cube mesh for physics debug static body")
		return
	}

	bodyOptions := physics.DefaultBodyOptions(0)
	bodyOptions.Position = startingCube.Position()
	bodyOptions.Rotation = startingCube.Rotation()
	bodyOptions.Static = true

	_, err := g.physicsDebug.world.CreateCubeWithOptions(physics.CubeOptions{
		BodyOptions: bodyOptions,
		Size:        startingCube.Scale(),
	})
	if err != nil {
		g.Logger().Error("failed to create physics debug static cube body", "error", err)
	}
}

func (g *Client) findPhysicsDebugStartingCube() *entity.Entity {
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

func (g *Client) stepPhysicsDebug(delta time.Duration) {
	if g.physicsDebug == nil || g.physicsDebug.world == nil {
		return
	}

	g.physicsDebug.world.Step(delta)
	for bodyID, entityID := range g.physicsDebug.bodyEntities {
		body, ok := g.physicsDebug.world.Body(bodyID)
		if !ok {
			delete(g.physicsDebug.bodyEntities, bodyID)
			continue
		}

		e := g.world.GetEntityByID(entityID)
		if e == nil {
			g.physicsDebug.world.RemoveBody(bodyID)
			delete(g.physicsDebug.bodyEntities, bodyID)
			continue
		}

		transform := body.Transform()
		entity.SetLocalPosition(e, transform.Position)
		e.SetLocalRotation(transform.Rotation)
	}
}
