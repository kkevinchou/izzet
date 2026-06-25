package server

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/event"
)

const (
	physicsCubeSize   = 1.0
	physicsSpawnAbove = 20.0
)

func (g *Server) SpawnPhysicsCube(contactPoint mgl64.Vec3) {
	position := contactPoint.Add(mgl64.Vec3{0, physicsSpawnAbove, 0})

	cube := entity.CreateCube(g.AssetManager(), physicsCubeSize)
	cube.Name = "physics_cube"
	cube.Static = false
	cube.Physics = &entity.PhysicsComponent{
		Shape:          entity.PhysicsShapeCube,
		Mass:           1,
		Restitution:    0.05,
		Friction:       0.9,
		LinearDamping:  0.05,
		AngularDamping: 0.1,
	}
	entity.SetLocalPosition(cube, position)

	g.EventsManager().EntitySpawnTopic.Write(event.EntitySpawnEvent{Entity: cube})
}
